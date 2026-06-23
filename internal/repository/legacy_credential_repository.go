package repository

import (
	"context"
	"time"
)

const (
	LegacyCredentialStatusActive   = "active"
	LegacyCredentialStatusMigrated = "migrated"
	LegacyCredentialStatusDisabled = "disabled"
)

// LegacyCredential contains an imported old-system password verifier.
// LegacyCredential 表示从旧系统导入的密码校验材料。
type LegacyCredential struct {
	ID           string
	UserID       string
	Source       string
	HashType     string
	PasswordHash string
	Salt         string
	Status       string
	MigratedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// LegacyCredentialRepository reads and updates imported password migration records.
// LegacyCredentialRepository 负责读取和更新旧密码迁移记录。
type LegacyCredentialRepository interface {
	FindActiveByUserID(ctx context.Context, userID string) ([]LegacyCredential, error)
	MarkMigrated(ctx context.Context, id string) error
}
