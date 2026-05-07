package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/handler"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/router"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/clock"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	infrahash "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/crypto"
	infraemail "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/email"
	infrajwt "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/jwt"
	inframessage "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/message"
	infraoauth "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/oauth"
	infraphone "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/phone"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	pgrepo "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/repository"
	infraratelimit "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/ratelimit"
	infraredis "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/redis"
	infrawallet "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/wallet"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
	adminusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/admin"
	authusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
	clientusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/client"
	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
	oauthusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/oauth"
	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
	walletusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/wallet"
)

// Application owns process-level dependencies and lifecycle.
// Application 持有进程级依赖和生命周期控制。
type Application struct {
	cfg    *config.Config
	logger *zap.Logger
	server *http.Server
	sqlDB  *sql.DB
	redis  *goredis.Client
}

// New wires infrastructure adapters, usecases, and HTTP delivery.
// New 装配基础设施适配器、用例服务和 HTTP 交付层。
func New(cfg *config.Config, logger *zap.Logger) (*Application, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	if err := cfg.ValidateProduction(); err != nil {
		return nil, err
	}

	db, sqlDB, err := postgres.Open(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if cfg.Database.AutoMigrate {
		if err := db.AutoMigrate(&model.User{}, &model.Client{}, &model.UserWallet{}, &model.OAuthAccount{}, &model.WalletNonce{}, &model.RefreshToken{}, &model.LoginLog{}, &model.SecurityEvent{}, &model.UserClient{}); err != nil {
			return nil, fmt.Errorf("auto migrate database: %w", err)
		}
	}

	userRepo := pgrepo.NewUserRepository(db)
	clientRepo := pgrepo.NewClientRepository(db)
	refreshTokenRepo := pgrepo.NewRefreshTokenRepository(db)
	activityRepo := pgrepo.NewActivityRepository(db)
	walletRepo := pgrepo.NewWalletRepository(db)
	oauthAccountRepo := pgrepo.NewOAuthAccountRepository(db)
	if err := clientRepo.EnsureDefault(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure default client: %w", err)
	}

	var redisClient *goredis.Client
	if requiresRedis(cfg) && !cfg.Redis.Enabled {
		return nil, errors.New("redis is required by configured code storage or rate limiting")
	}
	if cfg.Redis.Enabled {
		redisClient, err = infraredis.Open(context.Background(), cfg.Redis)
		if err != nil {
			return nil, fmt.Errorf("open redis: %w", err)
		}
	}

	hasher := infrahash.NewBcryptHasher(0)
	tokenHasher := infrahash.NewSHA256TokenHasher()
	tokenIssuer, err := infrajwt.NewService(cfg.JWT)
	if err != nil {
		return nil, fmt.Errorf("initialize jwt service: %w", err)
	}
	phoneCodeRepo := phoneCodeRepository(cfg, redisClient)
	emailCodeRepo := emailCodeRepository(cfg, redisClient)
	limiter := rateLimiter(cfg, redisClient)
	// Wire usecases with ports only; concrete adapters stay in infrastructure.
	// 这里只做依赖装配：usecase 接收端口，具体实现留在 infrastructure 层。
	authService := authusecase.NewService(
		userRepo,
		clientRepo,
		refreshTokenRepo,
		activityRepo,
		emailCodeRepo,
		phoneCodeRepo,
		walletRepo,
		oauthAccountRepo,
		limiter,
		hasher,
		tokenHasher,
		tokenIssuer,
		cfg.Auth.RateLimitEnabled,
		cfg.Auth.LoginLimit,
		cfg.Auth.LoginWindow,
	)
	clientService := clientusecase.NewService(clientRepo)
	adminService := adminusecase.NewService(adminusecase.Dependencies{
		Users:    userRepo,
		Activity: activityRepo,
		Wallets:  walletRepo,
		Accounts: oauthAccountRepo,
		Sessions: refreshTokenRepo,
	})
	smsProvider, _ := inframessage.NewProvider(cfg.Phone.Provider)
	_, emailProvider := inframessage.NewProvider(cfg.Email.Provider)
	phoneService := phoneusecase.NewService(phoneusecase.Dependencies{
		Users:         userRepo,
		Clients:       clientRepo,
		RefreshTokens: refreshTokenRepo,
		Activity:      activityRepo,
		Codes:         phoneCodeRepo,
		Limiter:       limiter,
		Sender:        smsProvider,
		TokenHasher:   tokenHasher,
		Issuer:        tokenIssuer,
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
	emailService := emailusecase.NewService(emailusecase.Dependencies{
		Codes:         emailCodeRepo,
		Limiter:       limiter,
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
	oauthService := oauthusecase.NewService(oauthusecase.Dependencies{
		Users:         userRepo,
		Clients:       clientRepo,
		RefreshTokens: refreshTokenRepo,
		Activity:      activityRepo,
		Accounts:      oauthAccountRepo,
		States:        infraoauth.NewMemoryStateStore(),
		Providers: []oauthusecase.Provider{
			infraoauth.NewHTTPProvider(infraoauth.ProviderConfig{
				Name:         "google",
				ClientID:     cfg.OAuth.Google.ClientID,
				ClientSecret: cfg.OAuth.Google.ClientSecret,
				AuthURL:      cfg.OAuth.Google.AuthURL,
				TokenURL:     cfg.OAuth.Google.TokenURL,
				UserInfoURL:  cfg.OAuth.Google.UserInfoURL,
				Scopes:       cfg.OAuth.Google.Scopes,
				Tenants:      oauthTenants(cfg.OAuth.Google.Tenants, cfg.OAuth.Google.TenantCredentials),
			}),
			infraoauth.NewHTTPProvider(infraoauth.ProviderConfig{
				Name:         "github",
				ClientID:     cfg.OAuth.GitHub.ClientID,
				ClientSecret: cfg.OAuth.GitHub.ClientSecret,
				AuthURL:      cfg.OAuth.GitHub.AuthURL,
				TokenURL:     cfg.OAuth.GitHub.TokenURL,
				UserInfoURL:  cfg.OAuth.GitHub.UserInfoURL,
				Scopes:       cfg.OAuth.GitHub.Scopes,
				Tenants:      oauthTenants(cfg.OAuth.GitHub.Tenants, cfg.OAuth.GitHub.TenantCredentials),
			}),
		},
		TokenHasher: tokenHasher,
		Issuer:      tokenIssuer,
		StateTTL:    cfg.OAuth.StateTTL,
		Clock:       clock.SystemClock{},
	})
	walletService := walletusecase.NewService(walletusecase.Dependencies{
		Wallets:       walletRepo,
		Users:         userRepo,
		Clients:       clientRepo,
		RefreshTokens: refreshTokenRepo,
		Activity:      activityRepo,
		Verifier:      infrawallet.NewEVMVerifier(),
		Limiter:       limiter,
		TokenHasher:   tokenHasher,
		Issuer:        tokenIssuer,
		NonceTTL:      cfg.Wallet.NonceTTL,
		RateLimit:     cfg.Wallet.RateLimitEnabled,
		NonceLimit:    cfg.Wallet.NonceLimit,
		NonceWindow:   cfg.Wallet.NonceWindow,
		Clock:         clock.SystemClock{},
	})

	engine := router.New(router.Dependencies{
		Config:           cfg,
		Logger:           logger,
		Auth:             handler.NewAuthHandler(authService),
		Wallet:           handler.NewWalletHandler(walletService),
		Phone:            handler.NewPhoneHandler(phoneService),
		Email:            handler.NewEmailHandler(emailService),
		OAuth:            handler.NewOAuthHandler(oauthService),
		Client:           handler.NewClientHandler(clientService),
		Admin:            handler.NewAdminHandler(adminService),
		Token:            tokenIssuer,
		AudienceResolver: clientService,
		JWKS:             handler.NewJWKSHandler(tokenIssuer),
		AdminToken:       cfg.Management.AdminToken,
	})

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:           engine,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	return &Application{
		cfg:    cfg,
		logger: logger,
		server: server,
		sqlDB:  sqlDB,
		redis:  redisClient,
	}, nil
}

// Start runs the HTTP server until the context is cancelled or the server exits.
// Start 启动 HTTP 服务，直到上下文取消或服务退出。
func (a *Application) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("starting http server",
			zap.String("addr", a.server.Addr),
			zap.String("env", a.cfg.App.Env),
		)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
		return nil
	case err := <-errCh:
		return err
	}
}

// Shutdown gracefully stops the HTTP server and closes database resources.
// Shutdown 优雅停止 HTTP 服务，并关闭数据库资源。
func (a *Application) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down http server", zap.Duration("timeout", 10*time.Second))
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}
	var closeErr error
	if a.sqlDB != nil {
		closeErr = a.sqlDB.Close()
	}
	if a.redis != nil {
		if err := a.redis.Close(); closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

// phoneCodeRepository selects the configured phone-code storage adapter.
// phoneCodeRepository 根据配置选择手机号验证码存储适配器。
func phoneCodeRepository(cfg *config.Config, client *goredis.Client) repository.PhoneCodeRepository {
	if strings.EqualFold(cfg.Phone.CodeStore, "redis") && client != nil {
		return infraredis.NewCodeRepository(client, "owa:phone_code")
	}
	return infraphone.NewMemoryCodeRepository()
}

// emailCodeRepository selects the configured email-code storage adapter.
// emailCodeRepository 根据配置选择邮箱验证码存储适配器。
func emailCodeRepository(cfg *config.Config, client *goredis.Client) repository.EmailCodeRepository {
	if strings.EqualFold(cfg.Email.CodeStore, "redis") && client != nil {
		return infraredis.NewCodeRepository(client, "owa:email_code")
	}
	return infraemail.NewMemoryCodeRepository()
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

// rateLimiter selects Redis-backed or no-op rate limiting.
// rateLimiter 根据配置选择 Redis 限流器或空限流器。
func rateLimiter(cfg *config.Config, client *goredis.Client) repository.RateLimiter {
	if client != nil && (cfg.Phone.RateLimitEnabled || cfg.Email.RateLimitEnabled) {
		return infraredis.NewRateLimiter(client, "owa:rate")
	}
	if cfg.Phone.RateLimitEnabled || cfg.Email.RateLimitEnabled {
		return infraratelimit.NewMemoryLimiter()
	}
	return infraratelimit.NoopLimiter{}
}

// requiresRedis reports whether runtime configuration needs a Redis connection.
// requiresRedis 判断当前配置是否需要 Redis 连接。
func requiresRedis(cfg *config.Config) bool {
	return strings.EqualFold(cfg.Phone.CodeStore, "redis") ||
		strings.EqualFold(cfg.Email.CodeStore, "redis")
}
