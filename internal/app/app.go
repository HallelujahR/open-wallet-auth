package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
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

	storage, err := newStorage(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	runtime, err := newRuntimeAdapters(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	engine := newHTTPRouter(cfg, logger, storage, runtime)

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
		sqlDB:  storage.sqlDB,
		redis:  runtime.redis,
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
