package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrEmailAlreadyExists = "AUTH_EMAIL_ALREADY_EXISTS"
	ErrInvalidClient      = "CLIENT_INVALID"
	ErrInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	ErrInvalidInput       = "AUTH_INVALID_INPUT"
)

type PasswordHasher interface {
	Hash(plain string) (string, error)
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
	UserID   string
	Username string
	Email    string
	Token    *token.Pair
}

type RegisterRequest struct {
	ClientID string
	Username string
	Email    string
	Password string
}

type RegisterResult struct {
	UserID   string
	Username string
	Email    string
	Token    *token.Pair
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
	req.ClientID = defaultClientID(req.ClientID)
	req.Email = strings.TrimSpace(req.Email)

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
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Token:    pair,
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	req.ClientID = defaultClientID(req.ClientID)
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)

	if req.Email == "" || req.Password == "" {
		return nil, domain.NewError(ErrInvalidInput, "email and password are required")
	}
	if req.Username == "" {
		req.Username = strings.Split(req.Email, "@")[0]
	}

	client, err := s.clients.FindByClientID(ctx, req.ClientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}

	existing, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.NewError(ErrEmailAlreadyExists, "email already exists")
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Status:       user.StatusActive,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
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

	return &RegisterResult{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Token:    pair,
	}, nil
}

func defaultClientID(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "default"
	}
	return clientID
}
