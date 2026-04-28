package model

import "time"

type User struct {
	ID           string     `gorm:"primaryKey;type:varchar(64)"`
	Username     string     `gorm:"type:varchar(128);not null"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex"`
	Phone        string     `gorm:"type:varchar(32);uniqueIndex"`
	PasswordHash string     `gorm:"type:varchar(255)"`
	Avatar       string     `gorm:"type:varchar(512)"`
	Status       string     `gorm:"type:varchar(32);not null;default:active"`
	LastLoginAt  *time.Time `gorm:"type:timestamptz"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;not null"`
	UpdatedAt    time.Time  `gorm:"type:timestamptz;not null"`
}

func (User) TableName() string {
	return "users"
}
