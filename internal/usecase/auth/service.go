package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrEmailAlreadyExists  = "AUTH_EMAIL_ALREADY_EXISTS"
	ErrInvalidClient       = "CLIENT_INVALID"
	ErrInvalidCredentials  = "AUTH_INVALID_CREDENTIALS"
	ErrInvalidInput        = "AUTH_INVALID_INPUT"
	ErrInvalidRefreshToken = "AUTH_INVALID_REFRESH_TOKEN"
)

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash string, plain string) bool
}

type TokenIssuer interface {
	IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error)
	RefreshTokenTTL() time.Duration
}

type TokenHasher interface {
	HashToken(raw string) string
}

type Service struct {
	users         repository.UserRepository
	clients       repository.ClientRepository
	refreshTokens repository.RefreshTokenRepository
	hasher        PasswordHasher
	tokenHasher   TokenHasher
	issuer        TokenIssuer
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

type RefreshRequest struct {
	RefreshToken string
}

type RefreshResult struct {
	UserID   string
	Username string
	Email    string
	Token    *token.Pair
}

type LogoutRequest struct {
	RefreshToken string
}

func NewService(
	users repository.UserRepository,
	clients repository.ClientRepository,
	refreshTokens repository.RefreshTokenRepository,
	hasher PasswordHasher,
	tokenHasher TokenHasher,
	issuer TokenIssuer,
) *Service {
	return &Service{
		users:         users,
		clients:       clients,
		refreshTokens: refreshTokens,
		hasher:        hasher,
		tokenHasher:   tokenHasher,
		issuer:        issuer,
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
	if err := s.storeRefreshToken(ctx, u.ID, client.ClientID, pair.RefreshToken); err != nil {
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
	if err := s.storeRefreshToken(ctx, u.ID, client.ClientID, pair.RefreshToken); err != nil {
		return nil, err
	}

	return &RegisterResult{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Token:    pair,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*RefreshResult, error) {
	raw := strings.TrimSpace(req.RefreshToken)
	if raw == "" {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	refreshToken, err := s.refreshTokens.FindByHash(ctx, s.tokenHasher.HashToken(raw))
	if err != nil || refreshToken == nil {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}
	now := time.Now().UTC()
	if refreshToken.IsRevoked() || refreshToken.IsExpired(now) {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	u, err := s.users.FindByID(ctx, refreshToken.UserID)
	if err != nil || u == nil || !u.IsActive() {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	client, err := s.clients.FindByClientID(ctx, refreshToken.ClientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	if err := s.refreshTokens.Revoke(ctx, refreshToken.ID); err != nil {
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
	if err := s.storeRefreshToken(ctx, u.ID, client.ClientID, pair.RefreshToken); err != nil {
		return nil, err
	}

	return &RefreshResult{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Token:    pair,
	}, nil
}

func (s *Service) Logout(ctx context.Context, req LogoutRequest) error {
	raw := strings.TrimSpace(req.RefreshToken)
	if raw == "" {
		return domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	refreshToken, err := s.refreshTokens.FindByHash(ctx, s.tokenHasher.HashToken(raw))
	if err != nil || refreshToken == nil {
		return domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}
	if refreshToken.IsRevoked() {
		return nil
	}
	return s.refreshTokens.Revoke(ctx, refreshToken.ID)
}

func (s *Service) storeRefreshToken(ctx context.Context, userID string, clientID string, raw string) error {
	return s.refreshTokens.Create(ctx, &token.RefreshToken{
		UserID:    userID,
		ClientID:  clientID,
		TokenHash: s.tokenHasher.HashToken(raw),
		ExpiresAt: time.Now().UTC().Add(s.issuer.RefreshTokenTTL()),
	})
}

func defaultClientID(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "default"
	}
	return clientID
}
