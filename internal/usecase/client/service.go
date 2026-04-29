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
	ErrInvalidClientInput  = "CLIENT_INVALID_INPUT"
)

// Service manages application clients that can request tokens.
type Service struct {
	clients repository.ClientRepository
}

// CreateRequest is the input for creating an application client.
type CreateRequest struct {
	ClientID            string
	Name                string
	JWTAudience         string
	AllowedOrigins      []string
	AllowedRedirectURIs []string
}

// NewService creates the client usecase service.
func NewService(clients repository.ClientRepository) *Service {
	return &Service{clients: clients}
}

// Create registers a new application client.
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
func (s *Service) List(ctx context.Context) ([]clientdomain.Client, error) {
	return s.clients.List(ctx)
}
