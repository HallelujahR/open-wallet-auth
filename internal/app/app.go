package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/router"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

type Application struct {
	cfg    *config.Config
	logger *zap.Logger
	server *http.Server
}

func New(cfg *config.Config, logger *zap.Logger) (*Application, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	engine := router.New(router.Dependencies{
		Config: cfg,
		Logger: logger,
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
	return a.server.Shutdown(ctx)
}
