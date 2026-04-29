package model

import "time"

// LoginLog maps to the login_logs table.
type LoginLog struct {
	ID          string    `gorm:"primaryKey;type:varchar(64)"`
	UserID      string    `gorm:"type:varchar(64);index"`
	ClientID    string    `gorm:"type:varchar(128);index"`
	LoginMethod string    `gorm:"type:varchar(32);not null"`
	IP          string    `gorm:"type:varchar(64)"`
	UserAgent   string    `gorm:"type:text"`
	Success     bool      `gorm:"not null"`
	FailureCode string    `gorm:"type:varchar(128)"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}
