package model

import (
	"encoding/json"
	"time"
)

// SystemSetting maps editable management settings to PostgreSQL.
// SystemSetting 映射管理后台可编辑配置，敏感字段由应用层负责脱敏展示。
type SystemSetting struct {
	Key       string          `gorm:"primaryKey;type:varchar(128)"`
	Value     json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt time.Time       `gorm:"type:timestamptz;not null"`
	UpdatedAt time.Time       `gorm:"type:timestamptz;not null"`
}

// TableName returns the physical table name for GORM.
// TableName 返回 GORM 使用的物理表名。
func (SystemSetting) TableName() string {
	return "system_settings"
}
