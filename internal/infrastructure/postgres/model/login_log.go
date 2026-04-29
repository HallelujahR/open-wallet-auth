package model

import "time"

// LoginLog maps to the login_logs table.
// LoginLog 映射 login_logs 数据表。
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

// TableName returns the physical table name for GORM.
// TableName 返回 GORM 使用的物理表名。
func (LoginLog) TableName() string {
	return "login_logs"
}
