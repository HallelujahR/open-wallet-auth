package email

import (
	"context"
	"sync"
	"time"
)

// emailCode is an in-memory verification-code record.
// emailCode 是内存中的邮箱验证码记录。
type emailCode struct {
	code      string
	expiresAt time.Time
}

// MemoryCodeRepository stores email verification codes in process memory.
// MemoryCodeRepository 将邮箱验证码保存在进程内存中，适合 demo 和本地开发。
type MemoryCodeRepository struct {
	mu    sync.Mutex
	codes map[string]emailCode
}

// NewMemoryCodeRepository creates an in-memory email code repository for local demos.
// NewMemoryCodeRepository 创建本地 demo 使用的内存邮箱验证码仓储。
func NewMemoryCodeRepository() *MemoryCodeRepository {
	return &MemoryCodeRepository{codes: map[string]emailCode{}}
}

// Save stores or replaces an email verification code until its expiration time.
// Save 保存或覆盖邮箱验证码，并记录过期时间。
func (r *MemoryCodeRepository) Save(ctx context.Context, email string, code string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes[email] = emailCode{code: code, expiresAt: expiresAt}
	return nil
}

// Verify checks and consumes a matching, unexpired email code.
// Verify 校验并消费未过期的邮箱验证码。
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
