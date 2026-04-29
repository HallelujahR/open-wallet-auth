package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	domainuser "github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// UserRepository persists users in PostgreSQL.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a PostgreSQL user repository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domainuser.User, error) {
	var row model.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	var row model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainUser(row), nil
}

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

func (r *UserRepository) UpdateLoginInfo(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"last_login_at": now, "updated_at": now}).Error
}

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

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

var _ domainrepo.UserRepository = (*UserRepository)(nil)
