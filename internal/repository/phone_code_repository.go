package repository

import (
	"context"
	"time"
)

// PhoneCodeRepository stores short-lived SMS-style verification codes.
type PhoneCodeRepository interface {
	Save(ctx context.Context, phone string, code string, expiresAt time.Time) error
	Verify(ctx context.Context, phone string, code string, now time.Time) (bool, error)
}
