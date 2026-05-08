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
		Status:              clientdomain.StatusActive,
	}
	if err := s.clients.Create(ctx, client); err != nil {
		return nil, err
	}
	return client, nil
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
