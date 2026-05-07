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
	settingsusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/settings"
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
	settingsService := settingsusecase.NewService(storage.settings, defaultSettingsSnapshot(cfg))
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
		CORSOrigins:      settingsService,
		Auth:             handler.NewAuthHandler(authService),
		Wallet:           handler.NewWalletHandler(newWalletService(cfg, storage, runtime)),
		Phone:            handler.NewPhoneHandler(newPhoneService(cfg, storage, runtime, settingsService)),
		Email:            handler.NewEmailHandler(newEmailService(cfg, runtime, settingsService)),
		OAuth:            handler.NewOAuthHandler(newOAuthService(cfg, storage, runtime, settingsService)),
		Client:           handler.NewClientHandler(clientService),
		Admin:            handler.NewAdminHandler(adminService),
		Settings:         handler.NewSettingsHandler(settingsService),
		Token:            runtime.issuer,
		AudienceResolver: clientService,
		JWKS:             handler.NewJWKSHandler(runtime.issuer),
		AdminToken:       cfg.Management.AdminToken,
	}
}

// newPhoneService creates the phone login usecase.
// newPhoneService 创建手机号验证码登录用例。
func newPhoneService(cfg *config.Config, storage *storageBundle, runtime *runtimeAdapters, settingsService *settingsusecase.Service) *phoneusecase.Service {
	smsProvider := inframessage.NewDynamicSMSProvider(settingsService, cfg.Phone.Provider)
	return phoneusecase.NewService(phoneusecase.Dependencies{
		Users:         storage.users,
		Clients:       storage.clients,
		RefreshTokens: storage.refreshTokens,
		Activity:      storage.activity,
		Codes:         runtime.phoneCodes,
		Limiter:       runtime.limiter,
		Sender:        smsProvider,
		EnabledReader: settingsService,
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
func newEmailService(cfg *config.Config, runtime *runtimeAdapters, settingsService *settingsusecase.Service) *emailusecase.Service {
	emailProvider := inframessage.NewDynamicEmailProvider(settingsService, cfg.Email.Provider)
	return emailusecase.NewService(emailusecase.Dependencies{
		Codes:         runtime.emailCodes,
		Limiter:       runtime.limiter,
		Sender:        emailProvider,
		EnabledReader: settingsService,
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
func newOAuthService(cfg *config.Config, storage *storageBundle, runtime *runtimeAdapters, settingsService *settingsusecase.Service) *oauthusecase.Service {
	return oauthusecase.NewService(oauthusecase.Dependencies{
		Users:         storage.users,
		Clients:       storage.clients,
		RefreshTokens: storage.refreshTokens,
		Activity:      storage.activity,
		Accounts:      storage.accounts,
		States:        infraoauth.NewMemoryStateStore(),
		Providers: []oauthusecase.Provider{
			infraoauth.NewDynamicProvider("google", oauthProviderConfig("google", cfg.OAuth.Google), settingsService),
			infraoauth.NewDynamicProvider("github", oauthProviderConfig("github", cfg.OAuth.GitHub), settingsService),
		},
		TokenHasher: runtime.tokenHash,
		Issuer:      runtime.issuer,
		StateTTL:    cfg.OAuth.StateTTL,
		Clock:       clock.SystemClock{},
	})
}

// defaultSettingsSnapshot converts boot-time config into editable provider settings.
// defaultSettingsSnapshot 将启动配置转换为管理后台可编辑的服务商配置默认值。
func defaultSettingsSnapshot(cfg *config.Config) settingsusecase.Snapshot {
	return settingsusecase.Snapshot{
		HTTP: settingsusecase.HTTPSettings{
			CORSAllowedOrigins: cfg.HTTP.CORSAllowedOrigins,
		},
		Phone: settingsusecase.PhoneSettings{
			Enabled:  cfg.Phone.Enabled,
			Provider: messageSettings(cfg.Phone.Provider),
		},
		Email: settingsusecase.EmailSettings{
			VerificationEnabled: cfg.Email.VerificationEnabled,
			Provider:            messageSettings(cfg.Email.Provider),
		},
		OAuth: settingsusecase.OAuthSettings{
			Google: oauthSettings(cfg.OAuth.Google),
			GitHub: oauthSettings(cfg.OAuth.GitHub),
		},
	}
}

// messageSettings converts infrastructure message config into settings usecase config.
// messageSettings 将 infrastructure 消息配置转换为 settings 用例配置。
func messageSettings(cfg config.MessageProviderConfig) settingsusecase.MessageProviderSettings {
	return settingsusecase.MessageProviderSettings{
		Type: cfg.Type,
		Webhook: settingsusecase.WebhookSettings{
			URL:         cfg.Webhook.URL,
			BearerToken: cfg.Webhook.BearerToken,
		},
		SMTP: settingsusecase.SMTPSettings{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			Username: cfg.SMTP.Username,
			Password: cfg.SMTP.Password,
			From:     cfg.SMTP.From,
		},
		AliyunSMS: settingsusecase.AliyunSMSSettings{
			AccessKeyID:     cfg.AliyunSMS.AccessKeyID,
			AccessKeySecret: cfg.AliyunSMS.AccessKeySecret,
			SignName:        cfg.AliyunSMS.SignName,
			TemplateCode:    cfg.AliyunSMS.TemplateCode,
			RegionID:        cfg.AliyunSMS.RegionID,
			Endpoint:        cfg.AliyunSMS.Endpoint,
		},
		Headers: cfg.Headers,
	}
}

// oauthSettings converts infrastructure OAuth config into settings usecase config.
// oauthSettings 将 infrastructure OAuth 配置转换为 settings 用例配置。
func oauthSettings(cfg config.OAuthProviderConfig) settingsusecase.OAuthProviderSettings {
	tenants := make([]settingsusecase.OAuthProviderTenantSettings, 0, len(cfg.TenantCredentials)+len(cfg.Tenants))
	for host, tenant := range cfg.Tenants {
		tenants = append(tenants, settingsusecase.OAuthProviderTenantSettings{
			Host:         host,
			ClientID:     tenant.ClientID,
			ClientSecret: tenant.ClientSecret,
		})
	}
	for _, tenant := range cfg.TenantCredentials {
		tenants = append(tenants, settingsusecase.OAuthProviderTenantSettings{
			Host:         tenant.Host,
			ClientID:     tenant.ClientID,
			ClientSecret: tenant.ClientSecret,
		})
	}
	return settingsusecase.OAuthProviderSettings{
		ClientID:          cfg.ClientID,
		ClientSecret:      cfg.ClientSecret,
		AuthURL:           cfg.AuthURL,
		TokenURL:          cfg.TokenURL,
		UserInfoURL:       cfg.UserInfoURL,
		Scopes:            cfg.Scopes,
		TenantCredentials: tenants,
	}
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
