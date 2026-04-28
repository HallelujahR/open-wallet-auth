package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, refreshToken *token.RefreshToken) error {
	now := time.Now().UTC()
	if refreshToken.ID == "" {
		refreshToken.ID = "rft_" + uuid.NewString()
	}
	if refreshToken.CreatedAt.IsZero() {
		refreshToken.CreatedAt = now
	}

	row := model.RefreshToken{
		ID:        refreshToken.ID,
		UserID:    refreshToken.UserID,
		ClientID:  refreshToken.ClientID,
		TokenHash: refreshToken.TokenHash,
		ExpiresAt: refreshToken.ExpiresAt,
		RevokedAt: refreshToken.RevokedAt,
		CreatedAt: refreshToken.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *RefreshTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*token.RefreshToken, error) {
	var row model.RefreshToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainRefreshToken(row), nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now).Error
}

func toDomainRefreshToken(row model.RefreshToken) *token.RefreshToken {
	return &token.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		ClientID:  row.ClientID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		RevokedAt: row.RevokedAt,
		CreatedAt: row.CreatedAt,
	}
}

var _ domainrepo.RefreshTokenRepository = (*RefreshTokenRepository)(nil)
