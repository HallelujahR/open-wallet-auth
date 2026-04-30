package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
)

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshToken *token.RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*token.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	Rotate(ctx context.Context, oldTokenID string, newToken *token.RefreshToken) error
}

// RefreshTokenListFilter contains filters for refresh-token session queries.
// RefreshTokenListFilter 描述刷新令牌会话查询条件。
type RefreshTokenListFilter struct {
	UserID     string
	ClientID   string
	ActiveOnly bool
}

// RefreshTokenAdminRepository defines management operations for token sessions.
// RefreshTokenAdminRepository 定义 token 会话管理需要的仓储能力。
type RefreshTokenAdminRepository interface {
	List(ctx context.Context, filter RefreshTokenListFilter) ([]token.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID string) (int64, error)
	RevokeByUserAndClient(ctx context.Context, userID string, clientID string) (int64, error)
}
