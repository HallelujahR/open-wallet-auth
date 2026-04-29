package phone

import (
	"context"
	"sync"
	"time"
)

type phoneCode struct {
	code      string
	expiresAt time.Time
}

// MemoryCodeRepository stores phone verification codes in process memory.
type MemoryCodeRepository struct {
	mu    sync.Mutex
	codes map[string]phoneCode
}

// NewMemoryCodeRepository creates an in-memory phone code repository for local demos.
func NewMemoryCodeRepository() *MemoryCodeRepository {
	return &MemoryCodeRepository{codes: map[string]phoneCode{}}
}

func (r *MemoryCodeRepository) Save(ctx context.Context, phone string, code string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes[phone] = phoneCode{code: code, expiresAt: expiresAt}
	return nil
}

func (r *MemoryCodeRepository) Verify(ctx context.Context, phone string, code string, now time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.codes[phone]
	if !ok || stored.code != code || !stored.expiresAt.After(now) {
		return false, nil
	}
	delete(r.codes, phone)
	return true, nil
}
