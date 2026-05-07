package app

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/handler"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/router"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/clock"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	inframessage "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/message"
	infraoauth "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/oauth"
	infrawallet "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/wallet"
	adminusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/admin"
	authusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
	clientusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/client"
	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
	oauthusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/oauth"
	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
	walletusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/wallet"
)

// newHTTPRouter wires usecases and exposes them through HTTP handlers.
// newHTTPRouter 装配用例服务并通过 HTTP handler 暴露，具体基础设施仍由 wire_* 文件提供。
func newHTTPRouter(cfg *config.Config, logger *zap.Logger, storage *storageBundle, runtime *runtimeAdapters) *gin.Engine {
	return router.New(newRouterDependencies(cfg, logger, storage, runtime))
}

// newRouterDependencies wires usecases and HTTP handlers without starting the server.
// newRouterDependencies 装配用例服务和 HTTP handler，但不启动服务。
func newRouterDependencies(cfg *config.Config, logger *zap.Logger, storage *storageBundle, runtime *runtimeAdapters) router.Dependencies {
	authService := authusecase.NewService(
		storage.users,
		storage.clients,
		storage.refreshTokens,
		storage.activity,
		runtime.emailCodes,
		runtime.phoneCodes,
		storage.wallets,
		storage.accounts,
		runtime.limiter,
		runtime.hasher,
		runtime.tokenHash,
		runtime.issuer,
		cfg.Auth.RateLimitEnabled,
		cfg.Auth.LoginLimit,
		cfg.Auth.LoginWindow,
	)
	clientService := clientusecase.NewService(storage.clients)
	adminService := adminusecase.NewService(adminusecase.Dependencies{
		Users:    storage.users,
		Activity: storage.activity,
		Wallets:  storage.wallets,
		Accounts: storage.accounts,
		Sessions: storage.refreshTokens,
	})

	return router.Dependencies{
		Config:           cfg,
		Logger:           logger,
		Auth:             handler.NewAuthHandler(authService),
		Wallet:           handler.NewWalletHandler(newWalletService(cfg, storage, runtime)),
		Phone:            handler.NewPhoneHandler(newPhoneService(cfg, storage, runtime)),
		Email:            handler.NewEmailHandler(newEmailService(cfg, runtime)),
		OAuth:            handler.NewOAuthHandler(newOAuthService(cfg, storage, runtime)),
		Client:           handler.NewClientHandler(clientService),
		Admin:            handler.NewAdminHandler(adminService),
		Token:            runtime.issuer,
		AudienceResolver: clientService,
		JWKS:             handler.NewJWKSHandler(runtime.issuer),
		AdminToken:       cfg.Management.AdminToken,
	}
}

// newPhoneService creates the phone login usecase.
// newPhoneService 创建手机号验证码登录用例。
func newPhoneService(cfg *config.Config, storage *storageBundle, runtime *runtimeAdapters) *phoneusecase.Service {
	smsProvider, _ := inframessage.NewProvider(cfg.Phone.Provider)
	return phoneusecase.NewService(phoneusecase.Dependencies{
		Users:         storage.users,
		Clients:       storage.clients,
		RefreshTokens: storage.refreshTokens,
		Activity:      storage.activity,
		Codes:         runtime.phoneCodes,
		Limiter:       runtime.limiter,
		Sender:        smsProvider,
		TokenHasher:   runtime.tokenHash,
		Issuer:        runtime.issuer,
		Enabled:       cfg.Phone.Enabled,
		CodeTTL:       cfg.Phone.CodeTTL,
		RateLimit:     cfg.Phone.RateLimitEnabled,
		SendLimit:     cfg.Phone.SendLimit,
		SendWindow:    cfg.Phone.SendWindow,
		VerifyLimit:   cfg.Phone.VerifyLimit,
		VerifyWindow:  cfg.Phone.VerifyWindow,
		DevCode:       cfg.Phone.DevCode,
		ExposeDevCode: cfg.Phone.ExposeDevCode,
		Clock:         clock.SystemClock{},
	})
}

// newEmailService creates the email verification usecase.
// newEmailService 创建邮箱验证码用例。
func newEmailService(cfg *config.Config, runtime *runtimeAdapters) *emailusecase.Service {
	_, emailProvider := inframessage.NewProvider(cfg.Email.Provider)
	return emailusecase.NewService(emailusecase.Dependencies{
		Codes:         runtime.emailCodes,
		Limiter:       runtime.limiter,
		Sender:        emailProvider,
		Enabled:       cfg.Email.VerificationEnabled,
		CodeTTL:       cfg.Email.CodeTTL,
		RateLimit:     cfg.Email.RateLimitEnabled,
		SendLimit:     cfg.Email.SendLimit,
		SendWindow:    cfg.Email.SendWindow,
		VerifyLimit:   cfg.Email.VerifyLimit,
		VerifyWindow:  cfg.Email.VerifyWindow,
		DevCode:       cfg.Email.DevCode,
		ExposeDevCode: cfg.Email.ExposeDevCode,
		Clock:         clock.SystemClock{},
	})
}

// newOAuthService creates OAuth providers and the OAuth login usecase.
// newOAuthService 创建 OAuth provider 适配器和 OAuth 登录用例。
func newOAuthService(cfg *config.Config, storage *storageBundle, runtime *runtimeAdapters) *oauthusecase.Service {
	return oauthusecase.NewService(oauthusecase.Dependencies{
		Users:         storage.users,
		Clients:       storage.clients,
		RefreshTokens: storage.refreshTokens,
		Activity:      storage.activity,
		Accounts:      storage.accounts,
		States:        infraoauth.NewMemoryStateStore(),
		Providers: []oauthusecase.Provider{
			infraoauth.NewHTTPProvider(oauthProviderConfig("google", cfg.OAuth.Google)),
			infraoauth.NewHTTPProvider(oauthProviderConfig("github", cfg.OAuth.GitHub)),
		},
		TokenHasher: runtime.tokenHash,
		Issuer:      runtime.issuer,
		StateTTL:    cfg.OAuth.StateTTL,
		Clock:       clock.SystemClock{},
	})
}

// newWalletService creates the EVM wallet login usecase.
// newWalletService 创建 EVM 钱包登录用例。
func newWalletService(cfg *config.Config, storage *storageBundle, runtime *runtimeAdapters) *walletusecase.Service {
	return walletusecase.NewService(walletusecase.Dependencies{
		Wallets:       storage.wallets,
		Users:         storage.users,
		Clients:       storage.clients,
		RefreshTokens: storage.refreshTokens,
		Activity:      storage.activity,
		Verifier:      infrawallet.NewEVMVerifier(),
		Limiter:       runtime.limiter,
		TokenHasher:   runtime.tokenHash,
		Issuer:        runtime.issuer,
		NonceTTL:      cfg.Wallet.NonceTTL,
		RateLimit:     cfg.Wallet.RateLimitEnabled,
		NonceLimit:    cfg.Wallet.NonceLimit,
		NonceWindow:   cfg.Wallet.NonceWindow,
		Clock:         clock.SystemClock{},
	})
}

// oauthProviderConfig converts runtime config into an OAuth provider adapter config.
// oauthProviderConfig 将运行配置转换为 OAuth provider 适配器配置。
func oauthProviderConfig(name string, cfg config.OAuthProviderConfig) infraoauth.ProviderConfig {
	return infraoauth.ProviderConfig{
		Name:         name,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		AuthURL:      cfg.AuthURL,
		TokenURL:     cfg.TokenURL,
		UserInfoURL:  cfg.UserInfoURL,
		Scopes:       cfg.Scopes,
		Tenants:      oauthTenants(cfg.Tenants, cfg.TenantCredentials),
	}
}

// oauthTenants maps config tenant credentials into the OAuth provider adapter model.
// oauthTenants 将配置层租户凭据转换为 OAuth provider 适配器可用的结构。
func oauthTenants(source map[string]config.OAuthProviderTenantConfig, entries []config.OAuthProviderTenantConfig) map[string]infraoauth.ProviderTenantConfig {
	if len(source) == 0 && len(entries) == 0 {
		return nil
	}
	tenants := make(map[string]infraoauth.ProviderTenantConfig, len(source)+len(entries))
	for host, tenant := range source {
		tenants[strings.ToLower(strings.TrimSpace(host))] = infraoauth.ProviderTenantConfig{
			ClientID:     tenant.ClientID,
			ClientSecret: tenant.ClientSecret,
		}
	}
	for _, tenant := range entries {
		host := strings.ToLower(strings.TrimSpace(tenant.Host))
		if host == "" {
			continue
		}
		tenants[host] = infraoauth.ProviderTenantConfig{
			ClientID:     tenant.ClientID,
			ClientSecret: tenant.ClientSecret,
		}
	}
	return tenants
}
