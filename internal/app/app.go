package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/handler"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/router"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	infrahash "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/crypto"
	infrajwt "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/jwt"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	pgrepo "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/repository"
	authusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
)

// Application owns process-level dependencies and lifecycle.
type Application struct {
	cfg    *config.Config
	logger *zap.Logger
	server *http.Server
	sqlDB  *sql.DB
}

// New wires infrastructure adapters, usecases, and HTTP delivery.
func New(cfg *config.Config, logger *zap.Logger) (*Application, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	db, sqlDB, err := postgres.Open(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if cfg.Database.AutoMigrate {
		if err := db.AutoMigrate(&model.User{}, &model.Client{}, &model.RefreshToken{}); err != nil {
			return nil, fmt.Errorf("auto migrate database: %w", err)
		}
	}

	userRepo := pgrepo.NewUserRepository(db)
	clientRepo := pgrepo.NewClientRepository(db)
	refreshTokenRepo := pgrepo.NewRefreshTokenRepository(db)
	if err := clientRepo.EnsureDefault(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure default client: %w", err)
	}

	hasher := infrahash.NewBcryptHasher(0)
	tokenHasher := infrahash.NewSHA256TokenHasher()
	tokenIssuer, err := infrajwt.NewService(cfg.JWT)
	if err != nil {
		return nil, fmt.Errorf("initialize jwt service: %w", err)
	}
	authService := authusecase.NewService(userRepo, clientRepo, refreshTokenRepo, hasher, tokenHasher, tokenIssuer)

	engine := router.New(router.Dependencies{
		Config: cfg,
		Logger: logger,
		Auth:   handler.NewAuthHandler(authService),
		Token:  tokenIssuer,
		JWKS:   handler.NewJWKSHandler(tokenIssuer),
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
	}, nil
}

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

func (a *Application) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down http server", zap.Duration("timeout", 10*time.Second))
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}
	if a.sqlDB != nil {
		return a.sqlDB.Close()
	}
	return nil
}
