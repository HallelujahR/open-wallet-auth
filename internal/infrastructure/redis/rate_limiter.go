package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// RateLimiter implements fixed-window rate limiting with Redis counters.
// RateLimiter 使用 Redis 计数器实现固定窗口限流。
type RateLimiter struct {
	client *goredis.Client
	prefix string
}

// NewRateLimiter creates a Redis-backed rate limiter.
// NewRateLimiter 创建 Redis 限流器。
func NewRateLimiter(client *goredis.Client, prefix string) *RateLimiter {
	return &RateLimiter{client: client, prefix: prefix}
}

// Allow increments the key counter and reports whether it stays within limit.
// Allow 递增 key 的计数，并返回是否仍在限制范围内。
func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 || window <= 0 {
		return true, nil
	}
	redisKey := r.prefix + ":" + key
	count, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := r.client.Expire(ctx, redisKey, window).Err(); err != nil {
			return false, err
		}
	}
	return count <= int64(limit), nil
}

// Reset removes one Redis-backed rate-limit counter.
// Reset 删除一个 Redis 限流计数。
func (r *RateLimiter) Reset(ctx context.Context, key string) (bool, error) {
	deleted, err := r.client.Del(ctx, r.prefix+":"+key).Result()
	if err != nil {
		return false, err
	}
	return deleted > 0, nil
}
