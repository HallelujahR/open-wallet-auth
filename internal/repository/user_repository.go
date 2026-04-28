package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	Create(ctx context.Context, u *user.User) error
	UpdateLoginInfo(ctx context.Context, userID string) error
}
