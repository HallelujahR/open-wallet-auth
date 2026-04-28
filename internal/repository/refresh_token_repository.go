package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshToken *token.RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*token.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
}
