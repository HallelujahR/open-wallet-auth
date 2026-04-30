package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
)

// Open creates a Redis client and verifies connectivity with PING.
// Open 创建 Redis 客户端，并通过 PING 验证连接可用。
func Open(ctx context.Context, cfg config.RedisConfig) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
