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
		ID:         refreshToken.ID,
		UserID:     refreshToken.UserID,
		ClientID:   refreshToken.ClientID,
		TokenHash:  refreshToken.TokenHash,
		IP:         refreshToken.IP,
		UserAgent:  refreshToken.UserAgent,
		ExpiresAt:  refreshToken.ExpiresAt,
		RevokedAt:  refreshToken.RevokedAt,
		LastUsedAt: refreshToken.LastUsedAt,
		CreatedAt:  refreshToken.CreatedAt,
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

// List returns refresh-token sessions for management APIs.
// List 为管理接口返回刷新令牌会话列表。
func (r *RefreshTokenRepository) List(ctx context.Context, filter domainrepo.RefreshTokenListFilter) ([]token.RefreshToken, error) {
	query := r.db.WithContext(ctx).Model(&model.RefreshToken{})
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.ClientID != "" {
		query = query.Where("client_id = ?", filter.ClientID)
	}
	if filter.ActiveOnly {
		query = query.Where("revoked_at IS NULL AND expires_at > ?", time.Now().UTC())
	}
	var rows []model.RefreshToken
	if err := query.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	tokens := make([]token.RefreshToken, 0, len(rows))
	for _, row := range rows {
		tokens = append(tokens, *toDomainRefreshToken(row))
	}
	return tokens, nil
}

// RevokeByUserID revokes every active refresh token owned by a user.
// RevokeByUserID 吊销某个用户的全部有效刷新令牌。
func (r *RefreshTokenRepository) RevokeByUserID(ctx context.Context, userID string) (int64, error) {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now)
	return result.RowsAffected, result.Error
}

// RevokeByUserAndClient revokes active refresh tokens for one user and client.
// RevokeByUserAndClient 吊销某个用户在指定业务系统下的有效刷新令牌。
func (r *RefreshTokenRepository) RevokeByUserAndClient(ctx context.Context, userID string, clientID string) (int64, error) {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND client_id = ? AND revoked_at IS NULL", userID, clientID).
		Update("revoked_at", now)
	return result.RowsAffected, result.Error
}

// toDomainRefreshToken converts a database row into a domain refresh token.
// toDomainRefreshToken 将数据库行转换为领域刷新令牌实体。
func toDomainRefreshToken(row model.RefreshToken) *token.RefreshToken {
	return &token.RefreshToken{
		ID:         row.ID,
		UserID:     row.UserID,
		ClientID:   row.ClientID,
		TokenHash:  row.TokenHash,
		IP:         row.IP,
		UserAgent:  row.UserAgent,
		ExpiresAt:  row.ExpiresAt,
		RevokedAt:  row.RevokedAt,
		LastUsedAt: row.LastUsedAt,
		CreatedAt:  row.CreatedAt,
	}
}

var _ domainrepo.RefreshTokenRepository = (*RefreshTokenRepository)(nil)
var _ domainrepo.RefreshTokenAdminRepository = (*RefreshTokenRepository)(nil)
