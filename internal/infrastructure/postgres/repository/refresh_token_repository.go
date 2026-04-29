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

// RefreshTokenRepository persists refresh tokens in PostgreSQL.
// RefreshTokenRepository 是刷新令牌仓储端口的 PostgreSQL 适配器。
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a PostgreSQL refresh token repository.
// NewRefreshTokenRepository 创建 PostgreSQL 刷新令牌仓储。
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create persists a hashed refresh token record.
// Create 持久化已哈希的刷新令牌记录。
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

// FindByHash loads a refresh token by its one-way hash.
// FindByHash 按单向哈希查询刷新令牌。
func (r *RefreshTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*token.RefreshToken, error) {
	var row model.RefreshToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainRefreshToken(row), nil
}

// Revoke marks a refresh token as revoked without deleting history.
// Revoke 标记刷新令牌为已吊销，同时保留审计历史。
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now).Error
}

// toDomainRefreshToken converts a database row into a domain refresh token.
// toDomainRefreshToken 将数据库行转换为领域刷新令牌实体。
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
