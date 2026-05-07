package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// SettingsRepository stores runtime-editable system settings in PostgreSQL.
// SettingsRepository 将运行期可编辑系统配置保存到 PostgreSQL。
type SettingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository creates a PostgreSQL settings repository.
// NewSettingsRepository 创建 PostgreSQL 系统配置仓储。
func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get loads one JSON setting by key.
// Get 按 key 读取一项 JSON 配置。
func (r *SettingsRepository) Get(ctx context.Context, key string) (json.RawMessage, error) {
	var row model.SystemSetting
	err := r.db.WithContext(ctx).First(&row, "key = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainrepo.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.Value, nil
}

// Upsert creates or replaces one JSON setting.
// Upsert 创建或替换一项 JSON 配置。
func (r *SettingsRepository) Upsert(ctx context.Context, key string, value json.RawMessage) error {
	now := time.Now().UTC()
	row := model.SystemSetting{Key: key, Value: value, CreatedAt: now, UpdatedAt: now}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{"value": value, "updated_at": now}),
	}).Create(&row).Error
}
