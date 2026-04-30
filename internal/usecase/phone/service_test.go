package phone

import (
	"context"
	"testing"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func TestServiceLoginCreatesPhoneUser(t *testing.T) {
	service := newTestService()
	if _, err := service.RequestCode(context.Background(), CodeRequest{ClientID: "default", Phone: "+8613800000000"}); err != nil {
		t.Fatalf("request code returned error: %v", err)
	}
	result, err := service.Login(context.Background(), LoginRequest{ClientID: "default", Phone: "+8613800000000", Code: "123456"})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if result.UserID == "" || result.Phone != "+8613800000000" || result.Token == nil {
		t.Fatal("expected user, phone, and token")
	}
}

func TestServiceRequestCodeRejectsRateLimitedPhone(t *testing.T) {
	service := newTestService()
	service.rateLimit = true
	service.limiter = denyLimiter{}

	_, err := service.RequestCode(context.Background(), CodeRequest{ClientID: "default", Phone: "+8613800000000"})
	if err == nil {
		t.Fatal("expected rate limit error")
	}
}

var testNow = time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)

func newTestService() *Service {
	return NewService(Dependencies{
		Users:         newMemoryUsers(),
		Clients:       defaultClients(),
		RefreshTokens: newMemoryRefreshTokens(),
		Activity:      newMemoryActivity(),
		Codes:         newMemoryCodes(),
		Sender:        noopSMS{},
		TokenHasher:   fakeTokenHasher{},
		Issuer:        fakeIssuer{},
		Enabled:       true,
		CodeTTL:       5 * time.Minute,
		DevCode:       "123456",
		ExposeDevCode: true,
		Clock:         fixedClock{},
	})
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return testNow }

type memoryUsers struct {
	byID    map[string]*user.User
	byPhone map[string]*user.User
}

func newMemoryUsers() *memoryUsers {
	return &memoryUsers{byID: map[string]*user.User{}, byPhone: map[string]*user.User{}}
}

func (m *memoryUsers) FindByID(ctx context.Context, id string) (*user.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (m *memoryUsers) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, repository.ErrNotFound
}

func (m *memoryUsers) FindByPhone(ctx context.Context, phone string) (*user.User, error) {
	u, ok := m.byPhone[phone]
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
	if u.Phone != "" {
		m.byPhone[u.Phone] = u
	}
	return nil
}

func (m *memoryUsers) UpdateLoginInfo(ctx context.Context, userID string) error {
	return nil
}

func (m *memoryUsers) UpdateEmail(ctx context.Context, userID string, email string) error {
	return nil
}

func (m *memoryUsers) UpdatePhone(ctx context.Context, userID string, phone string) error {
	u, ok := m.byID[userID]
	if !ok {
		return repository.ErrNotFound
	}
	u.Phone = phone
	m.byPhone[phone] = u
	return nil
}

func (m *memoryUsers) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	u, ok := m.byID[userID]
	if !ok {
		return repository.ErrNotFound
	}
	u.PasswordHash = passwordHash
	return nil
}

type memoryClients struct {
	byClientID map[string]*client.Client
}

func defaultClients() *memoryClients {
	return &memoryClients{byClientID: map[string]*client.Client{"default": {ClientID: "default", JWTAudience: "default", Status: client.StatusActive}}}
}

func (m *memoryClients) FindByClientID(ctx context.Context, clientID string) (*client.Client, error) {
	c, ok := m.byClientID[clientID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return c, nil
}

func (m *memoryClients) Create(ctx context.Context, c *client.Client) error { return nil }

func (m *memoryClients) List(ctx context.Context) ([]client.Client, error) { return nil, nil }

type memoryRefreshTokens struct{}

func newMemoryRefreshTokens() *memoryRefreshTokens { return &memoryRefreshTokens{} }

func (m *memoryRefreshTokens) Create(ctx context.Context, refreshToken *token.RefreshToken) error {
	return nil
}

func (m *memoryRefreshTokens) FindByHash(ctx context.Context, tokenHash string) (*token.RefreshToken, error) {
	return nil, repository.ErrNotFound
}

func (m *memoryRefreshTokens) Revoke(ctx context.Context, id string) error { return nil }

func (m *memoryRefreshTokens) Rotate(ctx context.Context, oldTokenID string, newToken *token.RefreshToken) error {
	return nil
}

func (m *memoryRefreshTokens) RevokeByUserID(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

type memoryActivity struct{}

func newMemoryActivity() *memoryActivity { return &memoryActivity{} }

func (m *memoryActivity) RecordLogin(ctx context.Context, log *audit.LoginLog) error {
	return nil
}

func (m *memoryActivity) UpsertUserClientLogin(ctx context.Context, userID string, clientID string) error {
	return nil
}

type memoryCodes struct {
	code      string
	expiresAt time.Time
}

type noopSMS struct{}

func (noopSMS) SendSMS(ctx context.Context, msg SMSMessage) error { return nil }

type denyLimiter struct{}

func (denyLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return false, nil
}

func newMemoryCodes() *memoryCodes { return &memoryCodes{} }

func (m *memoryCodes) Save(ctx context.Context, phone string, code string, expiresAt time.Time) error {
	m.code = code
	m.expiresAt = expiresAt
	return nil
}

func (m *memoryCodes) Verify(ctx context.Context, phone string, code string, now time.Time) (bool, error) {
	return m.code == code && m.expiresAt.After(now), nil
}

type fakeTokenHasher struct{}

func (fakeTokenHasher) HashToken(raw string) string { return "hash:" + raw }

type fakeIssuer struct{}

func (fakeIssuer) IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error) {
	return &token.Pair{AccessToken: "access", RefreshToken: "refresh", ExpiresAt: testNow.Add(time.Hour)}, nil
}

func (fakeIssuer) RefreshTokenTTL() time.Duration { return time.Hour }
