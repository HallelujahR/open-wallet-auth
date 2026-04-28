package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func TestServiceRegisterSuccess(t *testing.T) {
	users := newMemoryUsers()
	service := NewService(users, defaultClients(), fakeHasher{}, fakeIssuer{})

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
	service := NewService(users, defaultClients(), fakeHasher{}, fakeIssuer{})

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
	service := NewService(users, defaultClients(), fakeHasher{}, fakeIssuer{})

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

type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error) {
	return "hash:" + plain, nil
}

func (fakeHasher) Compare(hash string, plain string) bool {
	return hash == "hash:"+plain
}

type fakeIssuer struct{}

func (fakeIssuer) IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error) {
	return &token.Pair{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}
