package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// ActivityRepository persists login logs and user-client activity in PostgreSQL.
// ActivityRepository 是登录审计和用户-client 活动仓储的 PostgreSQL 适配器。
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository creates a PostgreSQL activity repository.
// NewActivityRepository 创建 PostgreSQL 活动仓储。
func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// RecordLogin stores one login audit event.
// RecordLogin 记录一次登录审计事件。
func (r *ActivityRepository) RecordLogin(ctx context.Context, log *audit.LoginLog) error {
	now := time.Now().UTC()
	if log.ID == "" {
		log.ID = "log_" + uuid.NewString()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}

	row := model.LoginLog{
		ID:          log.ID,
		UserID:      log.UserID,
		ClientID:    log.ClientID,
		LoginMethod: string(log.LoginMethod),
		IP:          log.IP,
		UserAgent:   log.UserAgent,
		Success:     log.Success,
		FailureCode: log.FailureCode,
		CreatedAt:   log.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

// UpsertUserClientLogin maintains the user-client relationship and counters.
// UpsertUserClientLogin 维护用户与业务系统的登录关系和次数统计。
func (r *ActivityRepository) UpsertUserClientLogin(ctx context.Context, userID string, clientID string) error {
	now := time.Now().UTC()
	row := model.UserClient{
		ID:           "ucl_" + uuid.NewString(),
		UserID:       userID,
		ClientID:     clientID,
		FirstLoginAt: now,
		LastLoginAt:  now,
		LoginCount:   1,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "client_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"last_login_at": now,
				"login_count":   gorm.Expr("user_clients.login_count + 1"),
				"updated_at":    now,
			}),
		}).
		Create(&row).Error
}

var _ domainrepo.ActivityRepository = (*ActivityRepository)(nil)
