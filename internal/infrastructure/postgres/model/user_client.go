package model

import "time"

// UserClient maps to the user_clients table.
// UserClient 映射 user_clients 数据表。
type UserClient struct {
	ID           string    `gorm:"primaryKey;type:varchar(64)"`
	UserID       string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_user_client"`
	ClientID     string    `gorm:"type:varchar(128);not null;uniqueIndex:idx_user_client"`
	FirstLoginAt time.Time `gorm:"type:timestamptz;not null"`
	LastLoginAt  time.Time `gorm:"type:timestamptz;not null"`
	LoginCount   int64     `gorm:"not null;default:1"`
	Status       string    `gorm:"type:varchar(32);not null;default:active"`
	CreatedAt    time.Time `gorm:"type:timestamptz;not null"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;not null"`
}

// TableName returns the physical table name for GORM.
// TableName 返回 GORM 使用的物理表名。
func (UserClient) TableName() string {
	return "user_clients"
}
