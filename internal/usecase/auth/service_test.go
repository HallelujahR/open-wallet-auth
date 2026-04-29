package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func TestServiceRegisterSuccess(t *testing.T) {
	users := newMemoryUsers()
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), fakeHasher{}, fakeTokenHasher{}, fakeIssuer{})

	result, err := service.Register(context.Background(), RegisterRequest{
		ClientID: "default",
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}
	if result.UserID == "" {
		t.Fatal("expected user id")
	}
	if result.Token == nil || result.Token.AccessToken == "" {
		t.Fatal("expected token pair")
	}
}

func TestServiceRegisterRejectsExistingEmail(t *testing.T) {
	users := newMemoryUsers()
	users.byEmail["alice@example.com"] = &user.User{
		ID:     "usr_existing",
		Email:  "alice@example.com",
		Status: user.StatusActive,
	}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), fakeHasher{}, fakeTokenHasher{}, fakeIssuer{})

	_, err := service.Register(context.Background(), RegisterRequest{
		ClientID: "default",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrEmailAlreadyExists {
		t.Fatalf("expected %s, got %v", ErrEmailAlreadyExists, err)
	}
}

func TestServiceLoginRejectsInvalidPassword(t *testing.T) {
	users := newMemoryUsers()
	users.byEmail["alice@example.com"] = &user.User{
		ID:           "usr_existing",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hash:correct",
		Status:       user.StatusActive,
	}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), fakeHasher{}, fakeTokenHasher{}, fakeIssuer{})

	_, err := service.Login(context.Background(), LoginRequest{
		ClientID: "default",
		Email:    "alice@example.com",
		Password: "wrong",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrInvalidCredentials {
		t.Fatalf("expected %s, got %v", ErrInvalidCredentials, err)
	}
}

func TestServiceRefreshRotatesRefreshToken(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_existing"] = &user.User{
		ID:       "usr_existing",
		Username: "alice",
		Email:    "alice@example.com",
		Status:   user.StatusActive,
	}
	refreshTokens := newMemoryRefreshTokens()
	refreshTokens.byHash["hash:old_refresh"] = &token.RefreshToken{
		ID:        "rft_old",
		UserID:    "usr_existing",
		ClientID:  "default",
		TokenHash: "hash:old_refresh",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	refreshTokens.byID["rft_old"] = refreshTokens.byHash["hash:old_refresh"]
	activity := newMemoryActivity()
	service := NewService(users, defaultClients(), refreshTokens, activity, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{})

	result, err := service.Refresh(context.Background(), RefreshRequest{RefreshToken: "old_refresh"})
	if err != nil {
		t.Fatalf("refresh returned error: %v", err)
	}
	if result.Token == nil || result.Token.RefreshToken == "" {
		t.Fatal("expected new token pair")
	}
	if refreshTokens.byID["rft_old"].RevokedAt == nil {
		t.Fatal("expected old refresh token to be revoked")
	}
	if activity.loginCount != 1 || activity.userClientCount != 1 {
		t.Fatal("expected refresh activity to be recorded")
	}
}

type memoryUsers struct {
	byID    map[string]*user.User
	byEmail map[string]*user.User
}

func newMemoryUsers() *memoryUsers {
	return &memoryUsers{
		byID:    map[string]*user.User{},
		byEmail: map[string]*user.User{},
	}
}

func (m *memoryUsers) FindByID(ctx context.Context, id string) (*user.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (m *memoryUsers) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (m *memoryUsers) Create(ctx context.Context, u *user.User) error {
	if u.ID == "" {
		u.ID = "usr_test"
	}
	m.byID[u.ID] = u
	m.byEmail[u.Email] = u
	return nil
}

func (m *memoryUsers) UpdateLoginInfo(ctx context.Context, userID string) error {
	return nil
}

type memoryClients struct {
	byClientID map[string]*client.Client
}

type memoryRefreshTokens struct {
	byID   map[string]*token.RefreshToken
	byHash map[string]*token.RefreshToken
}

type memoryActivity struct {
	loginCount      int
	userClientCount int
}

func newMemoryActivity() *memoryActivity {
	return &memoryActivity{}
}

func (m *memoryActivity) RecordLogin(ctx context.Context, log *audit.LoginLog) error {
	m.loginCount++
	return nil
}

func (m *memoryActivity) UpsertUserClientLogin(ctx context.Context, userID string, clientID string) error {
	m.userClientCount++
	return nil
}

func newMemoryRefreshTokens() *memoryRefreshTokens {
	return &memoryRefreshTokens{
		byID:   map[string]*token.RefreshToken{},
		byHash: map[string]*token.RefreshToken{},
	}
}

func (m *memoryRefreshTokens) Create(ctx context.Context, refreshToken *token.RefreshToken) error {
	if refreshToken.ID == "" {
		refreshToken.ID = "rft_test"
	}
	m.byID[refreshToken.ID] = refreshToken
	m.byHash[refreshToken.TokenHash] = refreshToken
	return nil
}

func (m *memoryRefreshTokens) FindByHash(ctx context.Context, tokenHash string) (*token.RefreshToken, error) {
	refreshToken, ok := m.byHash[tokenHash]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return refreshToken, nil
}

func (m *memoryRefreshTokens) Revoke(ctx context.Context, id string) error {
	refreshToken, ok := m.byID[id]
	if !ok {
		return repository.ErrNotFound
	}
	now := time.Now()
	refreshToken.RevokedAt = &now
	return nil
}

func defaultClients() *memoryClients {
	return &memoryClients{
		byClientID: map[string]*client.Client{
			"default": {
				ID:          "cli_default",
				ClientID:    "default",
				JWTAudience: "default",
				Status:      client.StatusActive,
			},
		},
	}
}

func (m *memoryClients) FindByClientID(ctx context.Context, clientID string) (*client.Client, error) {
	c, ok := m.byClientID[clientID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return c, nil
}

func (m *memoryClients) Create(ctx context.Context, c *client.Client) error {
	m.byClientID[c.ClientID] = c
	return nil
}

func (m *memoryClients) List(ctx context.Context) ([]client.Client, error) {
	clients := make([]client.Client, 0, len(m.byClientID))
	for _, c := range m.byClientID {
		clients = append(clients, *c)
	}
	return clients, nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error) {
	return "hash:" + plain, nil
}

func (fakeHasher) Compare(hash string, plain string) bool {
	return hash == "hash:"+plain
}

type fakeTokenHasher struct{}

func (fakeTokenHasher) HashToken(raw string) string {
	return "hash:" + raw
}

type fakeIssuer struct{}

func (fakeIssuer) IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error) {
	return &token.Pair{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

func (fakeIssuer) RefreshTokenTTL() time.Duration {
	return time.Hour
}
