package repository

import (
	"context"
	"strings"
	"time"
)

// RateLimiter controls high-risk operations such as verification-code sending.
// RateLimiter 控制验证码发送、验证码校验等高风险操作的频率。
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// RateLimitResetter removes one rate-limit counter.
// RateLimitResetter 删除指定限流计数，供受控的管理端运维操作使用。
type RateLimitResetter interface {
	Reset(ctx context.Context, key string) (bool, error)
}

// PasswordLoginRateLimitKey builds the shared key used by login checks and admin resets.
// PasswordLoginRateLimitKey 生成密码登录检查与管理端清除操作共同使用的限流键。
func PasswordLoginRateLimitKey(clientID string, email string) string {
	return "auth:login:" + strings.TrimSpace(clientID) + ":" + strings.ToLower(strings.TrimSpace(email))
}
