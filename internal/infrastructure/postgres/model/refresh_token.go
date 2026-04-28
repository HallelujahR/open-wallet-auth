package model

import "time"

type RefreshToken struct {
	ID        string     `gorm:"primaryKey;type:varchar(64)"`
	UserID    string     `gorm:"type:varchar(64);not null;index"`
	ClientID  string     `gorm:"type:varchar(128);not null;index"`
	TokenHash string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	ExpiresAt time.Time  `gorm:"type:timestamptz;not null;index"`
	RevokedAt *time.Time `gorm:"type:timestamptz"`
	CreatedAt time.Time  `gorm:"type:timestamptz;not null"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
