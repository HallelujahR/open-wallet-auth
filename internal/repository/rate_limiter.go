package repository

import (
	"context"
	"time"
)

// RateLimiter controls high-risk operations such as verification-code sending.
// RateLimiter 控制验证码发送、验证码校验等高风险操作的频率。
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}
