package oauth

import (
	"context"
	"sync"
	"time"

	oauthusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/oauth"
)

// stateEntry is an in-memory OAuth state value with expiration.
// stateEntry 是内存中的 OAuth state 记录和过期时间。
type stateEntry struct {
	value     oauthusecase.StateValue
	expiresAt time.Time
}

// MemoryStateStore stores OAuth state in process memory for local deployments.
// MemoryStateStore 将 OAuth state 保存在进程内存中，适合本地部署和 demo。
type MemoryStateStore struct {
	mu     sync.Mutex
	states map[string]stateEntry
}

// NewMemoryStateStore creates an in-memory OAuth state store.
// NewMemoryStateStore 创建内存 OAuth state 仓储。
func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{states: map[string]stateEntry{}}
}

// Save stores a state value until the callback deadline.
// Save 保存 OAuth state 及其回调有效期。
func (s *MemoryStateStore) Save(ctx context.Context, state string, value oauthusecase.StateValue, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = stateEntry{value: value, expiresAt: expiresAt}
	return nil
}

// Take returns and consumes a valid state value.
// Take 返回并消费有效的 OAuth state。
func (s *MemoryStateStore) Take(ctx context.Context, state string, now time.Time) (*oauthusecase.StateValue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.states[state]
	if !ok || !entry.expiresAt.After(now) {
		return nil, nil
	}
	delete(s.states, state)
	return &entry.value, nil
}
