package client

import (
	"context"
	"errors"
	"strings"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	clientdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrClientAlreadyExists = "CLIENT_ALREADY_EXISTS"
	ErrClientNotFound      = "CLIENT_NOT_FOUND"
	ErrInvalidClientInput  = "CLIENT_INVALID_INPUT"
	ErrMemberNotFound      = "CLIENT_MEMBER_NOT_FOUND"
)

// Service manages application clients that can request tokens.
// Service 管理可接入认证服务并申请 token 的业务系统 client。
type Service struct {
	clients repository.ClientRepository
}

// CreateRequest is the input for creating an application client.
// CreateRequest 是创建业务系统 client 的用例输入。
type CreateRequest struct {
	ClientID            string
	Name                string
	JWTAudience         string
	AllowedOrigins      []string
	AllowedRedirectURIs []string
	WhitelistEnabled    bool
}

// UpdateAccessPolicyRequest toggles client-level login allow-list enforcement.
// UpdateAccessPolicyRequest 用于切换业务系统是否启用登录白名单。
type UpdateAccessPolicyRequest struct {
	ClientID         string
	WhitelistEnabled bool
}

// MemberRequest is the input for creating or updating an application member.
// MemberRequest 是创建或更新应用白名单成员的用例输入。
type MemberRequest struct {
	ClientID    string
	MemberID    string
	UserID      string
	Role        string
	Permissions []string
	Status      string
	Remark      string
	CreatedBy   string
}

// NewService creates the client usecase service.
// NewService 创建 client 管理用例服务。
func NewService(clients repository.ClientRepository) *Service {
	return &Service{clients: clients}
}

// Create registers a new application client.
// Create 注册一个新的业务系统 client。
func (s *Service) Create(ctx context.Context, req CreateRequest) (*clientdomain.Client, error) {
	req.ClientID = strings.TrimSpace(req.ClientID)
	req.Name = strings.TrimSpace(req.Name)
	req.JWTAudience = strings.TrimSpace(req.JWTAudience)
	if req.ClientID == "" || req.Name == "" {
		return nil, domain.NewError(ErrInvalidClientInput, "client_id and name are required")
	}
	if req.JWTAudience == "" {
		req.JWTAudience = req.ClientID
	}

	existing, err := s.clients.FindByClientID(ctx, req.ClientID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.NewError(ErrClientAlreadyExists, "client already exists")
	}

	client := &clientdomain.Client{
		ClientID:            req.ClientID,
		Name:                req.Name,
		JWTAudience:         req.JWTAudience,
		AllowedOrigins:      req.AllowedOrigins,
		AllowedRedirectURIs: req.AllowedRedirectURIs,
		WhitelistEnabled:    req.WhitelistEnabled,
		Status:              clientdomain.StatusActive,
	}
	if err := s.clients.Create(ctx, client); err != nil {
		return nil, err
	}
	return client, nil
}

// UpdateAccessPolicy changes whether a client requires explicit member access.
// UpdateAccessPolicy 修改业务系统是否要求显式成员授权后才能登录。
func (s *Service) UpdateAccessPolicy(ctx context.Context, req UpdateAccessPolicyRequest) (*clientdomain.Client, error) {
	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		return nil, domain.NewError(ErrInvalidClientInput, "client_id is required")
	}
	members, ok := s.clients.(repository.ClientMemberRepository)
	if !ok {
		return nil, domain.NewError(ErrInvalidClientInput, "client member repository is unavailable")
	}
	client, err := members.UpdateWhitelistEnabled(ctx, clientID, req.WhitelistEnabled)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(ErrClientNotFound, "client not found")
		}
		return nil, err
	}
	return client, nil
}

// ListMembers returns all allow-list members for one client.
// ListMembers 返回指定业务系统的白名单成员列表。
func (s *Service) ListMembers(ctx context.Context, clientID string) ([]clientdomain.Member, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, domain.NewError(ErrInvalidClientInput, "client_id is required")
	}
	if _, err := s.GetByClientID(ctx, clientID); err != nil {
		return nil, err
	}
	members, ok := s.clients.(repository.ClientMemberRepository)
	if !ok {
		return nil, domain.NewError(ErrInvalidClientInput, "client member repository is unavailable")
	}
	return members.ListMembers(ctx, clientID)
}

// AddMember creates or updates an allow-list member for one client.
// AddMember 为业务系统创建或更新一个白名单成员。
func (s *Service) AddMember(ctx context.Context, req MemberRequest) (*clientdomain.Member, error) {
	member, err := memberFromRequest(req)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetByClientID(ctx, member.ClientID); err != nil {
		return nil, err
	}
	members, ok := s.clients.(repository.ClientMemberRepository)
	if !ok {
		return nil, domain.NewError(ErrInvalidClientInput, "client member repository is unavailable")
	}
	if err := members.UpsertMember(ctx, member); err != nil {
		return nil, err
	}
	return member, nil
}

// UpdateMember changes one allow-list member's role, permissions, status, or remark.
// UpdateMember 更新单个白名单成员的角色、权限、状态或备注。
func (s *Service) UpdateMember(ctx context.Context, req MemberRequest) error {
	clientID := strings.TrimSpace(req.ClientID)
	memberID := strings.TrimSpace(req.MemberID)
	role := strings.TrimSpace(req.Role)
	status := clientdomain.MemberStatus(strings.TrimSpace(req.Status))
	if clientID == "" || memberID == "" {
		return domain.NewError(ErrInvalidClientInput, "client_id and member_id are required")
	}
	if role == "" {
		role = "member"
	}
	if status == "" {
		status = clientdomain.MemberStatusActive
	}
	if status != clientdomain.MemberStatusActive && status != clientdomain.MemberStatusDisabled {
		return domain.NewError(ErrInvalidClientInput, "invalid member status")
	}
	member := &clientdomain.Member{
		ClientID:    clientID,
		ID:          memberID,
		Role:        role,
		Permissions: normalizePermissions(req.Permissions),
		Status:      status,
		Remark:      strings.TrimSpace(req.Remark),
	}
	members, ok := s.clients.(repository.ClientMemberRepository)
	if !ok {
		return domain.NewError(ErrInvalidClientInput, "client member repository is unavailable")
	}
	if err := members.UpdateMember(ctx, member); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.NewError(ErrMemberNotFound, "client member not found")
		}
		return err
	}
	return nil
}

// DeleteMember removes one user from a client allow-list.
// DeleteMember 从业务系统白名单中移除一个用户。
func (s *Service) DeleteMember(ctx context.Context, clientID string, memberID string) error {
	clientID = strings.TrimSpace(clientID)
	memberID = strings.TrimSpace(memberID)
	if clientID == "" || memberID == "" {
		return domain.NewError(ErrInvalidClientInput, "client_id and member_id are required")
	}
	members, ok := s.clients.(repository.ClientMemberRepository)
	if !ok {
		return domain.NewError(ErrInvalidClientInput, "client member repository is unavailable")
	}
	if err := members.DeleteMember(ctx, clientID, memberID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.NewError(ErrMemberNotFound, "client member not found")
		}
		return err
	}
	return nil
}

// List returns all configured application clients.
// List 返回当前已配置的所有业务系统 client。
func (s *Service) List(ctx context.Context) ([]clientdomain.Client, error) {
	return s.clients.List(ctx)
}

// GetByClientID returns one configured application client.
// GetByClientID 按 client_id 返回一个已配置的业务系统。
func (s *Service) GetByClientID(ctx context.Context, clientID string) (*clientdomain.Client, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		clientID = "default"
	}
	client, err := s.clients.FindByClientID(ctx, clientID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(ErrClientNotFound, "client not found")
		}
		return nil, err
	}
	return client, nil
}

// memberFromRequest normalizes member fields shared by create and update.
// memberFromRequest 统一清洗白名单成员创建和更新请求。
func memberFromRequest(req MemberRequest) (*clientdomain.Member, error) {
	clientID := strings.TrimSpace(req.ClientID)
	userID := strings.TrimSpace(req.UserID)
	role := strings.TrimSpace(req.Role)
	status := clientdomain.MemberStatus(strings.TrimSpace(req.Status))
	if clientID == "" || userID == "" {
		return nil, domain.NewError(ErrInvalidClientInput, "client_id and user_id are required")
	}
	if role == "" {
		role = "member"
	}
	if status == "" {
		status = clientdomain.MemberStatusActive
	}
	if status != clientdomain.MemberStatusActive && status != clientdomain.MemberStatusDisabled {
		return nil, domain.NewError(ErrInvalidClientInput, "invalid member status")
	}
	return &clientdomain.Member{
		ClientID:    clientID,
		UserID:      userID,
		Role:        role,
		Permissions: normalizePermissions(req.Permissions),
		Status:      status,
		Remark:      strings.TrimSpace(req.Remark),
		CreatedBy:   strings.TrimSpace(req.CreatedBy),
	}, nil
}

// normalizePermissions trims and de-duplicates permission keys.
// normalizePermissions 清洗权限标识，避免空值和重复值写入 token。
func normalizePermissions(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	permissions := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		permissions = append(permissions, value)
	}
	return permissions
}
