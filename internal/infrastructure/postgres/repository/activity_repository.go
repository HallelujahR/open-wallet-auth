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

// RecordSecurityEvent stores one security-sensitive account event.
// RecordSecurityEvent 记录一次账号安全敏感操作事件。
func (r *ActivityRepository) RecordSecurityEvent(ctx context.Context, event *audit.SecurityEvent) error {
	now := time.Now().UTC()
	if event.ID == "" {
		event.ID = "sec_" + uuid.NewString()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}

	row := model.SecurityEvent{
		ID:          event.ID,
		UserID:      event.UserID,
		EventType:   string(event.EventType),
		TargetType:  event.TargetType,
		TargetID:    event.TargetID,
		IP:          event.IP,
		UserAgent:   event.UserAgent,
		Success:     event.Success,
		FailureCode: event.FailureCode,
		CreatedAt:   event.CreatedAt,
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

// ListLoginLogs returns paginated login audit events for management APIs.
// ListLoginLogs 为管理接口返回分页后的登录审计事件。
func (r *ActivityRepository) ListLoginLogs(ctx context.Context, filter domainrepo.LoginLogFilter) ([]audit.LoginLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.LoginLog{})
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.ClientID != "" {
		query = query.Where("client_id = ?", filter.ClientID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	var rows []model.LoginLog
	if err := query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	logs := make([]audit.LoginLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, toDomainLoginLog(row))
	}
	return logs, total, nil
}

// ListSecurityEvents returns paginated sensitive-operation audit events.
// ListSecurityEvents 返回分页后的敏感操作审计事件。
func (r *ActivityRepository) ListSecurityEvents(ctx context.Context, filter domainrepo.SecurityEventFilter) ([]audit.SecurityEvent, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.SecurityEvent{})
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	var rows []model.SecurityEvent
	if err := query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	events := make([]audit.SecurityEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, toDomainSecurityEvent(row))
	}
	return events, total, nil
}

// ListUserClients returns systems that a user has authenticated into.
// ListUserClients 返回该用户登录过的业务系统关系。
func (r *ActivityRepository) ListUserClients(ctx context.Context, userID string) ([]audit.UserClient, error) {
	var rows []model.UserClient
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("last_login_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	clients := make([]audit.UserClient, 0, len(rows))
	for _, row := range rows {
		clients = append(clients, toDomainUserClient(row))
	}
	return clients, nil
}

// toDomainLoginLog converts a login log row into the domain audit entity.
// toDomainLoginLog 将登录日志数据库行转换为领域审计实体。
func toDomainLoginLog(row model.LoginLog) audit.LoginLog {
	return audit.LoginLog{
		ID:          row.ID,
		UserID:      row.UserID,
		ClientID:    row.ClientID,
		LoginMethod: audit.LoginMethod(row.LoginMethod),
		IP:          row.IP,
		UserAgent:   row.UserAgent,
		Success:     row.Success,
		FailureCode: row.FailureCode,
		CreatedAt:   row.CreatedAt,
	}
}

// toDomainSecurityEvent converts a security event row into the domain audit entity.
// toDomainSecurityEvent 将安全事件数据库行转换为领域审计实体。
func toDomainSecurityEvent(row model.SecurityEvent) audit.SecurityEvent {
	return audit.SecurityEvent{
		ID:          row.ID,
		UserID:      row.UserID,
		EventType:   audit.SecurityEventType(row.EventType),
		TargetType:  row.TargetType,
		TargetID:    row.TargetID,
		IP:          row.IP,
		UserAgent:   row.UserAgent,
		Success:     row.Success,
		FailureCode: row.FailureCode,
		CreatedAt:   row.CreatedAt,
	}
}

// toDomainUserClient converts a user-client row into the domain audit entity.
// toDomainUserClient 将用户-client 数据库行转换为领域审计实体。
func toDomainUserClient(row model.UserClient) audit.UserClient {
	return audit.UserClient{
		ID:           row.ID,
		UserID:       row.UserID,
		ClientID:     row.ClientID,
		FirstLoginAt: row.FirstLoginAt,
		LastLoginAt:  row.LastLoginAt,
		LoginCount:   row.LoginCount,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

var _ domainrepo.ActivityRepository = (*ActivityRepository)(nil)
var _ domainrepo.AdminActivityRepository = (*ActivityRepository)(nil)
