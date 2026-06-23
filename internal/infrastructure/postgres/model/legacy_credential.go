package model

import "time"

// LegacyCredential maps to old-system password hashes imported for transparent migration.
// LegacyCredential 映射旧系统密码凭证迁移记录，用于首次登录后升级为当前 bcrypt 哈希。
type LegacyCredential struct {
	ID           string     `gorm:"primaryKey;type:varchar(64)"`
	UserID       string     `gorm:"type:varchar(64);not null;uniqueIndex:idx_legacy_user_source"`
	Source       string     `gorm:"type:varchar(128);not null;uniqueIndex:idx_legacy_user_source"`
	HashType     string     `gorm:"type:varchar(64);not null"`
	PasswordHash string     `gorm:"type:varchar(255);not null"`
	Salt         string     `gorm:"type:varchar(255);not null;default:''"`
	Status       string     `gorm:"type:varchar(32);not null;default:active"`
	MigratedAt   *time.Time `gorm:"type:timestamptz"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;not null"`
	UpdatedAt    time.Time  `gorm:"type:timestamptz;not null"`
}

// TableName returns the physical table name for GORM.
// TableName 返回 GORM 使用的物理表名。
func (LegacyCredential) TableName() string {
	return "legacy_credentials"
}
