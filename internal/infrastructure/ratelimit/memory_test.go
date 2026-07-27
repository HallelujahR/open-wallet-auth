package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestMemoryLimiterReset(t *testing.T) {
	limiter := NewMemoryLimiter()
	ctx := context.Background()
	key := "auth:login:default:user@example.com"

	if allowed, err := limiter.Allow(ctx, key, 1, time.Minute); err != nil || !allowed {
		t.Fatalf("first attempt should be allowed, allowed=%v err=%v", allowed, err)
	}
	if allowed, err := limiter.Allow(ctx, key, 1, time.Minute); err != nil || allowed {
		t.Fatalf("second attempt should be limited, allowed=%v err=%v", allowed, err)
	}
	if cleared, err := limiter.Reset(ctx, key); err != nil || !cleared {
		t.Fatalf("counter should be cleared, cleared=%v err=%v", cleared, err)
	}
	if allowed, err := limiter.Allow(ctx, key, 1, time.Minute); err != nil || !allowed {
		t.Fatalf("attempt after reset should be allowed, allowed=%v err=%v", allowed, err)
	}
}
