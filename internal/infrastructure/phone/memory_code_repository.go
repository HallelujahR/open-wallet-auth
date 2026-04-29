package phone

import (
	"context"
	"sync"
	"time"
)

// phoneCode is an in-memory verification-code record.
// phoneCode 是内存中的手机号验证码记录。
type phoneCode struct {
	code      string
	expiresAt time.Time
}

// MemoryCodeRepository stores phone verification codes in process memory.
// MemoryCodeRepository 将手机号验证码保存在进程内存中，适合 demo 和本地开发。
type MemoryCodeRepository struct {
	mu    sync.Mutex
	codes map[string]phoneCode
}

// NewMemoryCodeRepository creates an in-memory phone code repository for local demos.
// NewMemoryCodeRepository 创建本地 demo 使用的内存手机号验证码仓储。
func NewMemoryCodeRepository() *MemoryCodeRepository {
	return &MemoryCodeRepository{codes: map[string]phoneCode{}}
}

// Save stores or replaces a phone verification code until its expiration time.
// Save 保存或覆盖手机号验证码，并记录过期时间。
func (r *MemoryCodeRepository) Save(ctx context.Context, phone string, code string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes[phone] = phoneCode{code: code, expiresAt: expiresAt}
	return nil
}

// Verify checks and consumes a matching, unexpired phone code.
// Verify 校验并消费未过期的手机号验证码。
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
