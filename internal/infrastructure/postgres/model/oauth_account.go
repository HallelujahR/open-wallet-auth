package model

import "time"

// OAuthAccount maps to the oauth_accounts table.
type OAuthAccount struct {
	ID                string    `gorm:"primaryKey;type:varchar(64)"`
	UserID            string    `gorm:"type:varchar(64);not null;index"`
	Provider          string    `gorm:"type:varchar(32);not null;uniqueIndex:idx_oauth_provider_subject"`
	ProviderSubject   string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_oauth_provider_subject"`
	ProviderEmail     string    `gorm:"type:varchar(255)"`
	ProviderUsername  string    `gorm:"type:varchar(255)"`
	ProviderAvatarURL string    `gorm:"type:varchar(512)"`
	CreatedAt         time.Time `gorm:"type:timestamptz;not null"`
	UpdatedAt         time.Time `gorm:"type:timestamptz;not null"`
}

func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}
