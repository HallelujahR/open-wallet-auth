package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	goredis "github.com/redis/go-redis/v9"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	infrahash "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/crypto"
	infraemail "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/email"
	infrajwt "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/jwt"
	infraphone "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/phone"
	infraratelimit "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/ratelimit"
	infraredis "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/redis"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// runtimeAdapters groups non-database adapters shared across usecases.
// runtimeAdapters 汇总 JWT、哈希、Redis、验证码存储和限流等运行时适配器。
type runtimeAdapters struct {
	redis      *goredis.Client
	hasher     *infrahash.BcryptHasher
	tokenHash  infrahash.SHA256TokenHasher
	issuer     *infrajwt.Service
	phoneCodes repository.PhoneCodeRepository
	emailCodes repository.EmailCodeRepository
	limiter    repository.RateLimiter
}

// newRuntimeAdapters creates infrastructure adapters driven by runtime configuration.
// newRuntimeAdapters 根据运行配置创建基础设施适配器，并校验 Redis/JWT 等依赖。
func newRuntimeAdapters(ctx context.Context, cfg *config.Config) (*runtimeAdapters, error) {
	redisClient, err := openRedisIfNeeded(ctx, cfg)
	if err != nil {
		return nil, err
	}

	issuer, err := infrajwt.NewService(cfg.JWT)
	if err != nil {
		return nil, fmt.Errorf("initialize jwt service: %w", err)
	}

	return &runtimeAdapters{
		redis:      redisClient,
		hasher:     infrahash.NewBcryptHasher(0),
		tokenHash:  infrahash.NewSHA256TokenHasher(),
		issuer:     issuer,
		phoneCodes: phoneCodeRepository(cfg, redisClient),
		emailCodes: emailCodeRepository(cfg, redisClient),
		limiter:    rateLimiter(cfg, redisClient),
	}, nil
}

// openRedisIfNeeded opens Redis only when configuration enables it.
// openRedisIfNeeded 仅在配置启用时打开 Redis，并在必须使用 Redis 时做启动保护。
func openRedisIfNeeded(ctx context.Context, cfg *config.Config) (*goredis.Client, error) {
	if requiresRedis(cfg) && !cfg.Redis.Enabled {
		return nil, errors.New("redis is required by configured code storage or rate limiting")
	}
	if !cfg.Redis.Enabled {
		return nil, nil
	}
	client, err := infraredis.Open(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("open redis: %w", err)
	}
	return client, nil
}

// phoneCodeRepository selects the configured phone-code storage adapter.
// phoneCodeRepository 根据配置选择手机号验证码存储适配器。
func phoneCodeRepository(cfg *config.Config, client *goredis.Client) repository.PhoneCodeRepository {
	if strings.EqualFold(cfg.Phone.CodeStore, "redis") && client != nil {
		return infraredis.NewCodeRepository(client, "owa:phone_code")
	}
	return infraphone.NewMemoryCodeRepository()
}

// emailCodeRepository selects the configured email-code storage adapter.
// emailCodeRepository 根据配置选择邮箱验证码存储适配器。
func emailCodeRepository(cfg *config.Config, client *goredis.Client) repository.EmailCodeRepository {
	if strings.EqualFold(cfg.Email.CodeStore, "redis") && client != nil {
		return infraredis.NewCodeRepository(client, "owa:email_code")
	}
	return infraemail.NewMemoryCodeRepository()
}

// rateLimiter selects Redis-backed, memory-backed, or no-op rate limiting.
// rateLimiter 根据配置选择 Redis、内存或空限流器。
func rateLimiter(cfg *config.Config, client *goredis.Client) repository.RateLimiter {
	if client != nil && rateLimitEnabled(cfg) {
		return infraredis.NewRateLimiter(client, "owa:rate")
	}
	if rateLimitEnabled(cfg) {
		return infraratelimit.NewMemoryLimiter()
	}
	return infraratelimit.NoopLimiter{}
}

// rateLimitEnabled reports whether any usecase needs a limiter.
// rateLimitEnabled 判断是否有任一用例需要限流器。
func rateLimitEnabled(cfg *config.Config) bool {
	return cfg.Phone.RateLimitEnabled ||
		cfg.Email.RateLimitEnabled ||
		cfg.Auth.RateLimitEnabled ||
		cfg.Wallet.RateLimitEnabled
}

// requiresRedis reports whether runtime configuration needs a Redis connection.
// requiresRedis 判断当前配置是否需要 Redis 连接。
func requiresRedis(cfg *config.Config) bool {
	return strings.EqualFold(cfg.Phone.CodeStore, "redis") ||
		strings.EqualFold(cfg.Email.CodeStore, "redis")
}
