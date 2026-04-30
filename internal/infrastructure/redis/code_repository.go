package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// CodeRepository stores verification codes in Redis with key expiration.
// CodeRepository 使用 Redis 保存验证码，并依赖 key 过期时间自动失效。
type CodeRepository struct {
	client *goredis.Client
	prefix string
}

// NewCodeRepository creates a Redis-backed verification-code repository.
// NewCodeRepository 创建 Redis 验证码仓储。
func NewCodeRepository(client *goredis.Client, prefix string) *CodeRepository {
	return &CodeRepository{client: client, prefix: prefix}
}

// Save stores a verification code until its expiration time.
// Save 保存验证码，并设置到期自动删除时间。
func (r *CodeRepository) Save(ctx context.Context, subject string, code string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}
	return r.client.Set(ctx, r.key(subject), code, ttl).Err()
}

// Verify checks and consumes a matching verification code.
// Verify 校验并消费匹配的验证码，避免重复使用。
func (r *CodeRepository) Verify(ctx context.Context, subject string, code string, now time.Time) (bool, error) {
	key := r.key(subject)
	stored, err := r.client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if stored != code {
		return false, nil
	}
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return false, err
	}
	return true, nil
}

// key returns the Redis key for one verification-code subject.
// key 返回某个验证码主体对应的 Redis key。
func (r *CodeRepository) key(subject string) string {
	return r.prefix + ":" + subject
}
