package ratelimit

import (
	"context"
	"time"
)

// NoopLimiter allows every operation.
// NoopLimiter 放行所有操作，适合未启用限流的本地开发场景。
type NoopLimiter struct{}

// Allow always returns true.
// Allow 始终返回允许。
func (NoopLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return true, nil
}
