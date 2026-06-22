package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
		WhitelistEnabled:    c.WhitelistEnabled,
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
		WhitelistEnabled:    false,
		Status:              string(client.StatusActive),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

// UpdateWhitelistEnabled changes whether one client enforces allow-list login.
// UpdateWhitelistEnabled 修改业务系统是否启用登录白名单。
func (r *ClientRepository) UpdateWhitelistEnabled(ctx context.Context, clientID string, enabled bool) (*client.Client, error) {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&model.Client{}).
		Where("client_id = ?", clientID).
		Updates(map[string]any{"whitelist_enabled": enabled, "updated_at": now})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domainrepo.ErrNotFound
	}
	return r.FindByClientID(ctx, clientID)
}

// ListMembers returns allow-list members for one client with identity summaries.
// ListMembers 返回某个业务系统的白名单成员，并带上身份摘要字段。
func (r *ClientRepository) ListMembers(ctx context.Context, clientID string) ([]client.Member, error) {
	type memberRow struct {
		model.ClientMember
		Username string
		Email    string
		Phone    string
	}

	var rows []memberRow
	err := r.db.WithContext(ctx).
		Table("client_members").
		Select("client_members.*, users.username, users.email, users.phone").
		Joins("LEFT JOIN users ON users.id = client_members.user_id").
		Where("client_members.client_id = ?", clientID).
		Order("client_members.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	members := make([]client.Member, 0, len(rows))
	for _, row := range rows {
		member := toDomainClientMember(row.ClientMember)
		member.Username = row.Username
		member.Email = row.Email
		member.Phone = row.Phone
		members = append(members, *member)
	}
	return members, nil
}

// FindActiveMember returns an active allow-list member for token issuance.
// FindActiveMember 查询可用于签发 token 的启用白名单成员。
func (r *ClientRepository) FindActiveMember(ctx context.Context, clientID string, userID string) (*client.Member, error) {
	var row model.ClientMember
	if err := r.db.WithContext(ctx).
		Where("client_id = ? AND user_id = ? AND status = ?", clientID, userID, string(client.MemberStatusActive)).
		First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainClientMember(row), nil
}

// UpsertMember adds a user to a client allow-list or updates an existing member.
// UpsertMember 将用户加入应用白名单；若已存在则更新角色、权限和状态。
func (r *ClientRepository) UpsertMember(ctx context.Context, member *client.Member) error {
	now := time.Now().UTC()
	if member.ID == "" {
		member.ID = "clm_" + uuid.NewString()
	}
	if member.Role == "" {
		member.Role = "member"
	}
	if member.Status == "" {
		member.Status = client.MemberStatusActive
	}
	member.CreatedAt = now
	member.UpdatedAt = now

	row, err := toClientMemberRow(member)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "client_id"}, {Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"role":        row.Role,
				"permissions": row.Permissions,
				"status":      row.Status,
				"remark":      row.Remark,
				"updated_at":  now,
			}),
		}).
		Create(&row).Error
}

// UpdateMember updates role, permissions, status, and remark for one member.
// UpdateMember 更新应用成员的角色、权限、状态和备注。
func (r *ClientRepository) UpdateMember(ctx context.Context, member *client.Member) error {
	permissions, err := json.Marshal(member.Permissions)
	if err != nil {
		return err
	}
	result := r.db.WithContext(ctx).
		Model(&model.ClientMember{}).
		Where("id = ? AND client_id = ?", member.ID, member.ClientID).
		Updates(map[string]any{
			"role":        member.Role,
			"permissions": datatypes.JSON(permissions),
			"status":      string(member.Status),
			"remark":      member.Remark,
			"updated_at":  time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainrepo.ErrNotFound
	}
	return nil
}

// DeleteMember removes a user from a client allow-list.
// DeleteMember 从应用白名单中移除用户授权。
func (r *ClientRepository) DeleteMember(ctx context.Context, clientID string, memberID string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND client_id = ?", memberID, clientID).Delete(&model.ClientMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainrepo.ErrNotFound
	}
	return nil
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
		WhitelistEnabled:    row.WhitelistEnabled,
		Status:              client.Status(row.Status),
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}

// toClientMemberRow converts a domain allow-list member into a database row.
// toClientMemberRow 将领域白名单成员转换为数据库行。
func toClientMemberRow(member *client.Member) (model.ClientMember, error) {
	permissions, err := json.Marshal(member.Permissions)
	if err != nil {
		return model.ClientMember{}, err
	}
	return model.ClientMember{
		ID:          member.ID,
		ClientID:    member.ClientID,
		UserID:      member.UserID,
		Role:        member.Role,
		Permissions: datatypes.JSON(permissions),
		Status:      string(member.Status),
		Remark:      member.Remark,
		CreatedBy:   member.CreatedBy,
		CreatedAt:   member.CreatedAt,
		UpdatedAt:   member.UpdatedAt,
	}, nil
}

// toDomainClientMember converts a database row into a domain allow-list member.
// toDomainClientMember 将数据库行转换为领域白名单成员。
func toDomainClientMember(row model.ClientMember) *client.Member {
	return &client.Member{
		ID:          row.ID,
		ClientID:    row.ClientID,
		UserID:      row.UserID,
		Role:        row.Role,
		Permissions: jsonStringSlice(row.Permissions),
		Status:      client.MemberStatus(row.Status),
		Remark:      row.Remark,
		CreatedBy:   row.CreatedBy,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
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
var _ domainrepo.ClientMemberRepository = (*ClientRepository)(nil)
