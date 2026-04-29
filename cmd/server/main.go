package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/app"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/logger"
)

// main loads configuration, builds the application, and handles graceful shutdown.
// main 加载配置、装配应用，并处理优雅关闭。
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Log)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = log.Sync()
	}()

	application, err := app.New(cfg, log)
	if err != nil {
		log.Fatal("failed to initialize application", logger.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := application.Start(ctx); err != nil {
		log.Fatal("application stopped with error", logger.Error(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Error("application shutdown failed", logger.Error(err))
	}
}
