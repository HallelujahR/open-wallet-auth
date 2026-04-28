package auth

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrInvalidClient      = "CLIENT_INVALID"
	ErrInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
)

type PasswordHasher interface {
	Compare(hash string, plain string) bool
}

type TokenIssuer interface {
	IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error)
}

type Service struct {
	users   repository.UserRepository
	clients repository.ClientRepository
	hasher  PasswordHasher
	issuer  TokenIssuer
}

type LoginRequest struct {
	ClientID string
	Email    string
	Password string
}

type LoginResult struct {
	UserID string
	Token  *token.Pair
}

func NewService(
	users repository.UserRepository,
	clients repository.ClientRepository,
	hasher PasswordHasher,
	issuer TokenIssuer,
) *Service {
	return &Service{
		users:   users,
		clients: clients,
		hasher:  hasher,
		issuer:  issuer,
	}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	client, err := s.clients.FindByClientID(ctx, req.ClientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}

	u, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil || u == nil || !u.IsActive() {
		return nil, domain.NewError(ErrInvalidCredentials, "invalid email or password")
	}

	if !s.hasher.Compare(u.PasswordHash, req.Password) {
		return nil, domain.NewError(ErrInvalidCredentials, "invalid email or password")
	}

	pair, err := s.issuer.IssuePair(ctx, token.Claims{
		UserID:   u.ID,
		ClientID: client.ClientID,
		Audience: client.JWTAudience,
		Username: u.Username,
		Email:    u.Email,
	})
	if err != nil {
		return nil, err
	}

	if err := s.users.UpdateLoginInfo(ctx, u.ID); err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID: u.ID,
		Token:  pair,
	}, nil
}
