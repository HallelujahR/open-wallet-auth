package repository

import (
	"context"
	"encoding/json"
)

// SettingsRepository persists editable management settings.
// SettingsRepository 持久化管理后台可编辑配置。
type SettingsRepository interface {
	Get(ctx context.Context, key string) (json.RawMessage, error)
	Upsert(ctx context.Context, key string, value json.RawMessage) error
}
