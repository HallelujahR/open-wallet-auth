package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// LegacyCredentialRepository persists imported legacy password records in PostgreSQL.
// LegacyCredentialRepository 是旧密码迁移记录的 PostgreSQL 仓储适配器。
type LegacyCredentialRepository struct {
	db *gorm.DB
}

// NewLegacyCredentialRepository creates a PostgreSQL legacy credential repository.
// NewLegacyCredentialRepository 创建旧密码迁移记录仓储。
func NewLegacyCredentialRepository(db *gorm.DB) *LegacyCredentialRepository {
	return &LegacyCredentialRepository{db: db}
}

// FindActiveByUserID returns active legacy credentials for one identity user.
// FindActiveByUserID 查询某个统一认证用户仍可用于首次登录迁移的旧密码记录。
func (r *LegacyCredentialRepository) FindActiveByUserID(ctx context.Context, userID string) ([]domainrepo.LegacyCredential, error) {
	var rows []model.LegacyCredential
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domainrepo.LegacyCredentialStatusActive).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]domainrepo.LegacyCredential, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainLegacyCredential(row))
	}
	return items, nil
}

// MarkMigrated disables one legacy credential after its password has been upgraded.
// MarkMigrated 在旧密码校验成功并升级为当前密码哈希后标记迁移完成。
func (r *LegacyCredentialRepository) MarkMigrated(ctx context.Context, id string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&model.LegacyCredential{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      domainrepo.LegacyCredentialStatusMigrated,
			"migrated_at": now,
			"updated_at":  now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainrepo.ErrNotFound
	}
	return nil
}

// toDomainLegacyCredential converts a database row into a repository DTO.
// toDomainLegacyCredential 将数据库行转换为仓储层 DTO。
func toDomainLegacyCredential(row model.LegacyCredential) domainrepo.LegacyCredential {
	return domainrepo.LegacyCredential{
		ID:           row.ID,
		UserID:       row.UserID,
		Source:       row.Source,
		HashType:     row.HashType,
		PasswordHash: row.PasswordHash,
		Salt:         row.Salt,
		Status:       row.Status,
		MigratedAt:   row.MigratedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

var _ domainrepo.LegacyCredentialRepository = (*LegacyCredentialRepository)(nil)
