package model

import "time"

// SecurityEvent maps to the security_events table.
// SecurityEvent 映射 security_events 数据表。
type SecurityEvent struct {
	ID          string    `gorm:"primaryKey;type:varchar(64)"`
	UserID      string    `gorm:"type:varchar(64);index"`
	EventType   string    `gorm:"type:varchar(64);not null;index"`
	TargetType  string    `gorm:"type:varchar(64)"`
	TargetID    string    `gorm:"type:varchar(255)"`
	IP          string    `gorm:"type:varchar(64)"`
	UserAgent   string    `gorm:"type:text"`
	Success     bool      `gorm:"not null"`
	FailureCode string    `gorm:"type:varchar(128)"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null"`
}

// TableName returns the physical table name for GORM.
// TableName 返回 GORM 使用的物理表名。
func (SecurityEvent) TableName() string {
	return "security_events"
}
