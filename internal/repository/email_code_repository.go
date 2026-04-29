package repository

import (
	"context"
	"time"
)

// EmailCodeRepository stores short-lived email verification codes.
type EmailCodeRepository interface {
	Save(ctx context.Context, email string, code string, expiresAt time.Time) error
	Verify(ctx context.Context, email string, code string, now time.Time) (bool, error)
}
