package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	domainuser "github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// UserRepository persists users in PostgreSQL.
// UserRepository 是用户仓储端口的 PostgreSQL 适配器。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a PostgreSQL user repository.
// NewUserRepository 创建 PostgreSQL 用户仓储。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID loads a user by internal user id.
// FindByID 按内部用户 ID 查询用户。
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domainuser.User, error) {
	var row model.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainUser(row), nil
}

// FindByEmail loads a user by normalized email.
// FindByEmail 按归一化邮箱查询用户。
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	var row model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainUser(row), nil
}

// FindByPhone loads a user by phone number.
// FindByPhone 按手机号查询用户。
func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domainuser.User, error) {
	var row model.User
	if err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainUser(row), nil
}

// Create persists a new user and fills default identity fields.
// Create 持久化新用户，并补齐默认 ID、状态和时间字段。
func (r *UserRepository) Create(ctx context.Context, u *domainuser.User) error {
	now := time.Now().UTC()
	if u.ID == "" {
		u.ID = "usr_" + uuid.NewString()
	}
	if u.Status == "" {
		u.Status = domainuser.StatusActive
	}
	u.CreatedAt = now
	u.UpdatedAt = now

	row := model.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        stringPtrOrNil(u.Email),
		Phone:        stringPtrOrNil(u.Phone),
		PasswordHash: u.PasswordHash,
		Avatar:       u.Avatar,
		Status:       string(u.Status),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

// UpdateLoginInfo stores the latest successful login time.
// UpdateLoginInfo 更新用户最近一次成功登录时间。
func (r *UserRepository) UpdateLoginInfo(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"last_login_at": now, "updated_at": now}).Error
}

// UpdatePassword stores a new password hash for an existing identity user.
// UpdatePassword 为已存在的身份用户保存新的密码哈希。
func (r *UserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"password_hash": passwordHash, "updated_at": time.Now().UTC()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainrepo.ErrNotFound
	}
	return nil
}

// List returns paginated identity users for management APIs.
// List 为管理接口返回分页后的身份用户列表。
func (r *UserRepository) List(ctx context.Context, filter domainrepo.UserListFilter) ([]domainuser.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.User{})
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("id LIKE ? OR username LIKE ? OR email LIKE ? OR phone LIKE ?", like, like, like, like)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	var rows []model.User
	if err := query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	users := make([]domainuser.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *toDomainUser(row))
	}
	return users, total, nil
}

// UpdateStatus changes the identity user's lifecycle status.
// UpdateStatus 修改身份用户的生命周期状态。
func (r *UserRepository) UpdateStatus(ctx context.Context, userID string, status domainuser.Status) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"status": string(status), "updated_at": time.Now().UTC()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainrepo.ErrNotFound
	}
	return nil
}

// toDomainUser converts a database row into the domain user entity.
// toDomainUser 将数据库行转换为领域用户实体。
func toDomainUser(row model.User) *domainuser.User {
	return &domainuser.User{
		ID:           row.ID,
		Username:     row.Username,
		Email:        stringValue(row.Email),
		Phone:        stringValue(row.Phone),
		PasswordHash: row.PasswordHash,
		Avatar:       row.Avatar,
		Status:       domainuser.Status(row.Status),
		LastLoginAt:  row.LastLoginAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

// stringPtrOrNil converts an empty string to nil for nullable columns.
// stringPtrOrNil 将空字符串转换为 nil，以适配可空数据库字段。
func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// stringValue converts a nullable database string into a domain string.
// stringValue 将可空数据库字符串转换为领域层字符串。
func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// normalizePage returns bounded pagination values for repository queries.
// normalizePage 返回受限制的分页参数，避免数据库查询过大。
func normalizePage(page int, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

var _ domainrepo.UserRepository = (*UserRepository)(nil)
var _ domainrepo.AdminUserRepository = (*UserRepository)(nil)
