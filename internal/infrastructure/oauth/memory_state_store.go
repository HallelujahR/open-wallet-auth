package oauth

import (
	"context"
	"sync"
	"time"

	oauthusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/oauth"
)

type stateEntry struct {
	value     oauthusecase.StateValue
	expiresAt time.Time
}

// MemoryStateStore stores OAuth state in process memory for local deployments.
type MemoryStateStore struct {
	mu     sync.Mutex
	states map[string]stateEntry
}

// NewMemoryStateStore creates an in-memory OAuth state store.
func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{states: map[string]stateEntry{}}
}

func (s *MemoryStateStore) Save(ctx context.Context, state string, value oauthusecase.StateValue, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = stateEntry{value: value, expiresAt: expiresAt}
	return nil
}

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
