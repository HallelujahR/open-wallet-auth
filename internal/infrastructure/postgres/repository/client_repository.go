package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// ClientRepository persists clients in PostgreSQL.
// ClientRepository 是业务系统 client 仓储端口的 PostgreSQL 适配器。
type ClientRepository struct {
	db *gorm.DB
}

// NewClientRepository creates a PostgreSQL client repository.
// NewClientRepository 创建 PostgreSQL client 仓储。
func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

// FindByClientID loads an application client by public client id.
// FindByClientID 按公开 client_id 查询业务系统 client。
func (r *ClientRepository) FindByClientID(ctx context.Context, clientID string) (*client.Client, error) {
	var row model.Client
	if err := r.db.WithContext(ctx).Where("client_id = ?", clientID).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainClient(row), nil
}

// Create persists a new application client and JSON configuration fields.
// Create 持久化新的业务系统 client，并写入 JSON 配置字段。
func (r *ClientRepository) Create(ctx context.Context, c *client.Client) error {
	now := time.Now().UTC()
	if c.ID == "" {
		c.ID = "cli_" + uuid.NewString()
	}
	if c.JWTAudience == "" {
		c.JWTAudience = c.ClientID
	}
	if c.Status == "" {
		c.Status = client.StatusActive
	}
	c.CreatedAt = now
	c.UpdatedAt = now

	origins, err := json.Marshal(c.AllowedOrigins)
	if err != nil {
		return err
	}
	redirectURIs, err := json.Marshal(c.AllowedRedirectURIs)
	if err != nil {
		return err
	}

	row := model.Client{
		ID:                  c.ID,
		ClientID:            c.ClientID,
		Name:                c.Name,
		JWTAudience:         c.JWTAudience,
		AllowedOrigins:      datatypes.JSON(origins),
		AllowedRedirectURIs: datatypes.JSON(redirectURIs),
		Status:              string(c.Status),
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

// List returns all clients ordered by newest first.
// List 按创建时间倒序返回所有业务系统 client。
func (r *ClientRepository) List(ctx context.Context) ([]client.Client, error) {
	var rows []model.Client
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}

	clients := make([]client.Client, 0, len(rows))
	for _, row := range rows {
		clients = append(clients, *toDomainClient(row))
	}
	return clients, nil
}

// EnsureDefault creates a default client for local development and first boot.
// EnsureDefault 在首次启动或本地开发时创建默认 client。
func (r *ClientRepository) EnsureDefault(ctx context.Context) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Client{}).Where("client_id = ?", "default").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	row := model.Client{
		ID:                  "cli_" + uuid.NewString(),
		ClientID:            "default",
		Name:                "Default Application",
		JWTAudience:         "default",
		AllowedOrigins:      []byte(`[]`),
		AllowedRedirectURIs: []byte(`[]`),
		Status:              string(client.StatusActive),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

// toDomainClient converts a database row into the domain client entity.
// toDomainClient 将数据库行转换为领域 client 实体。
func toDomainClient(row model.Client) *client.Client {
	return &client.Client{
		ID:                  row.ID,
		ClientID:            row.ClientID,
		Name:                row.Name,
		JWTAudience:         row.JWTAudience,
		AllowedOrigins:      jsonStringSlice(row.AllowedOrigins),
		AllowedRedirectURIs: jsonStringSlice(row.AllowedRedirectURIs),
		Status:              client.Status(row.Status),
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}

// jsonStringSlice decodes a JSON column into a string slice.
// jsonStringSlice 将 JSON 字段解析为字符串切片。
func jsonStringSlice(raw []byte) []string {
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	return values
}

var _ domainrepo.ClientRepository = (*ClientRepository)(nil)
