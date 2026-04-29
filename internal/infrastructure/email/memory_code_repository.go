package email

import (
	"context"
	"sync"
	"time"
)

type emailCode struct {
	code      string
	expiresAt time.Time
}

// MemoryCodeRepository stores email verification codes in process memory.
type MemoryCodeRepository struct {
	mu    sync.Mutex
	codes map[string]emailCode
}

// NewMemoryCodeRepository creates an in-memory email code repository for local demos.
func NewMemoryCodeRepository() *MemoryCodeRepository {
	return &MemoryCodeRepository{codes: map[string]emailCode{}}
}

func (r *MemoryCodeRepository) Save(ctx context.Context, email string, code string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes[email] = emailCode{code: code, expiresAt: expiresAt}
	return nil
}

func (r *MemoryCodeRepository) Verify(ctx context.Context, email string, code string, now time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.codes[email]
	if !ok || stored.code != code || !stored.expiresAt.After(now) {
		return false, nil
	}
	delete(r.codes, email)
	return true, nil
}
