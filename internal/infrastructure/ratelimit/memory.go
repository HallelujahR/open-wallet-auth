package ratelimit

import (
	"context"
	"sync"
	"time"
)

// MemoryLimiter implements fixed-window rate limiting in process memory.
// MemoryLimiter 使用进程内存实现固定窗口限流，适合本地开发和单实例部署。
type MemoryLimiter struct {
	mu      sync.Mutex
	entries map[string]memoryEntry
}

type memoryEntry struct {
	count     int
	expiresAt time.Time
}

// NewMemoryLimiter creates an in-memory rate limiter.
// NewMemoryLimiter 创建内存限流器。
func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{entries: map[string]memoryEntry{}}
}

// Allow increments the key counter and reports whether it stays within limit.
// Allow 递增 key 的计数，并返回是否仍在限制范围内。
func (l *MemoryLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 || window <= 0 {
		return true, nil
	}
	now := time.Now().UTC()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := l.entries[key]
	if entry.expiresAt.IsZero() || !entry.expiresAt.After(now) {
		entry = memoryEntry{expiresAt: now.Add(window)}
	}
	entry.count++
	l.entries[key] = entry
	return entry.count <= limit, nil
}
