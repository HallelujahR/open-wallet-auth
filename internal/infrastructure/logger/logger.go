package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

// New creates a configured zap logger.
// New 根据配置创建 zap 结构化日志器。
func New(cfg config.LogConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.Set(strings.ToLower(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	if cfg.Format == "console" {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(level)
	}

	return zapCfg.Build()
}

// Error returns a zap error field.
// Error 返回统一的 zap 错误字段。
func Error(err error) zap.Field {
	return zap.Error(err)
}
