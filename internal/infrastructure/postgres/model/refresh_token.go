package model

import "time"

// RefreshToken maps to the refresh_tokens table.
// RefreshToken 映射 refresh_tokens 数据表。
type RefreshToken struct {
	ID         string     `gorm:"primaryKey;type:varchar(64)"`
	UserID     string     `gorm:"type:varchar(64);not null;index"`
	ClientID   string     `gorm:"type:varchar(128);not null;index"`
	TokenHash  string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	IP         string     `gorm:"type:varchar(64)"`
	UserAgent  string     `gorm:"type:text"`
	ExpiresAt  time.Time  `gorm:"type:timestamptz;not null;index"`
	RevokedAt  *time.Time `gorm:"type:timestamptz"`
	LastUsedAt *time.Time `gorm:"type:timestamptz"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;not null"`
}

// TableName returns the physical table name for GORM.
// TableName 返回 GORM 使用的物理表名。
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
