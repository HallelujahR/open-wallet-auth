package model

import (
	"time"

	"gorm.io/datatypes"
)

// Client maps to the clients table.
// Client 映射 clients 数据表。
type Client struct {
	ID                  string         `gorm:"primaryKey;type:varchar(64)"`
	ClientID            string         `gorm:"type:varchar(128);not null;uniqueIndex"`
	Name                string         `gorm:"type:varchar(128);not null"`
	JWTAudience         string         `gorm:"column:jwt_audience;type:varchar(128);not null"`
	AllowedOrigins      datatypes.JSON `gorm:"type:jsonb;not null"`
	AllowedRedirectURIs datatypes.JSON `gorm:"type:jsonb;not null"`
	Status              string         `gorm:"type:varchar(32);not null;default:active"`
	CreatedAt           time.Time      `gorm:"type:timestamptz;not null"`
	UpdatedAt           time.Time      `gorm:"type:timestamptz;not null"`
}

// TableName returns the physical table name for GORM.
// TableName 返回 GORM 使用的物理表名。
func (Client) TableName() string {
	return "clients"
}
