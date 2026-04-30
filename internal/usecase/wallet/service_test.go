package wallet

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
	walletdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func TestServiceCreateNonceReturnsMessage(t *testing.T) {
	service := newTestService()

	result, err := service.CreateNonce(context.Background(), NonceRequest{
		Address: "0x0000000000000000000000000000000000000001",
		Domain:  "example.com",
		ChainID: 1,
	})
	if err != nil {
		t.Fatalf("create nonce returned error: %v", err)
	}
	if result.Nonce == "" || result.Message == "" {
		t.Fatal("expected nonce and message")
	}
	if result.ExpiresAt.Sub(testNow) != 5*time.Minute {
		t.Fatal("expected configured nonce ttl")
	}
}

func TestServiceCreateNonceRejectsRateLimitedAddress(t *testing.T) {
	service := newTestService()
	service.limiter = denyLimiter{}
	service.rateLimit = true
	service.nonceLimit = 1
	service.nonceWindow = time.Minute

	_, err := service.CreateNonce(context.Background(), NonceRequest{
		Address: "0x0000000000000000000000000000000000000001",
		Domain:  "example.com",
		ChainID: 1,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrRateLimited {
		t.Fatalf("expected %s, got %v", ErrRateLimited, err)
	}
}

func TestServiceVerifySignatureCreatesWalletUser(t *testing.T) {
	service := newTestService()
	result, err := service.CreateNonce(context.Background(), NonceRequest{
		Address: "0x0000000000000000000000000000000000000001",
		Domain:  "example.com",
		ChainID: 1,
	})
	if err != nil {
		t.Fatalf("create nonce returned error: %v", err)
	}

	login, err := service.VerifySignature(context.Background(), VerifyRequest{
		ClientID:  "default",
		Address:   "0x0000000000000000000000000000000000000001",
		Nonce:     result.Nonce,
		Signature: "valid",
	})
	if err != nil {
		t.Fatalf("verify returned error: %v", err)
	}
	if login.UserID == "" || login.Token == nil || len(login.Wallets) != 1 {
		t.Fatal("expected user, token, and wallet claim")
	}
	if service.wallets.(*memoryWallets).nonces[result.Nonce].UsedAt == nil {
		t.Fatal("expected nonce to be marked used")
	}
	if service.activity.(*memoryActivity).loginCount != 1 {
		t.Fatal("expected wallet login activity")
	}
}

func TestServiceVerifySignatureRejectsReusedNonce(t *testing.T) {
	service := newTestService()
	result, err := service.CreateNonce(context.Background(), NonceRequest{
		Address: "0x0000000000000000000000000000000000000001",
		Domain:  "example.com",
		ChainID: 1,
	})
	if err != nil {
		t.Fatalf("create nonce returned error: %v", err)
	}
	req := VerifyRequest{
		ClientID:  "default",
		Address:   "0x0000000000000000000000000000000000000001",
		Nonce:     result.Nonce,
		Signature: "valid",
	}
	if _, err := service.VerifySignature(context.Background(), req); err != nil {
		t.Fatalf("first verify returned error: %v", err)
	}
	_, err = service.VerifySignature(context.Background(), req)
	if err == nil {
		t.Fatal("expected reused nonce error")
	}
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrInvalidNonce {
		t.Fatalf("expected %s, got %v", ErrInvalidNonce, err)
	}
}

var testNow = time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)

func newTestService() *Service {
	return NewService(Dependencies{
		Wallets:       newMemoryWallets(),
		Users:         newMemoryUsers(),
		Clients:       defaultClients(),
		RefreshTokens: newMemoryRefreshTokens(),
		Activity:      newMemoryActivity(),
		Verifier:      fakeVerifier{},
		TokenHasher:   fakeTokenHasher{},
		Issuer:        fakeIssuer{},
		NonceTTL:      5 * time.Minute,
		Clock:         fixedClock{},
	})
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return testNow }

type fakeVerifier struct{}

func (fakeVerifier) NormalizeAddress(address string) (string, error) {
	if address == "" {
		return "", errors.New("empty address")
	}
	return "0x0000000000000000000000000000000000000001", nil
}

func (fakeVerifier) VerifyMessage(address string, message string, signature string) (bool, error) {
	return signature == "valid", nil
}

type denyLimiter struct{}

func (denyLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return false, nil
}

type memoryWallets struct {
	byAddress map[string]*walletdomain.UserWallet
	nonces    map[string]*walletdomain.Nonce
}

func newMemoryWallets() *memoryWallets {
	return &memoryWallets{
		byAddress: map[string]*walletdomain.UserWallet{},
		nonces:    map[string]*walletdomain.Nonce{},
	}
}

func (m *memoryWallets) FindByAddress(ctx context.Context, chainType walletdomain.ChainType, address string) (*walletdomain.UserWallet, error) {
	w, ok := m.byAddress[string(chainType)+":"+address]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return w, nil
}

func (m *memoryWallets) CreateWallet(ctx context.Context, w *walletdomain.UserWallet) error {
	if w.ID == "" {
		w.ID = "wal_test"
	}
	m.byAddress[string(w.ChainType)+":"+w.Address] = w
	return nil
}

func (m *memoryWallets) CreateNonce(ctx context.Context, nonce *walletdomain.Nonce) error {
	if nonce.ID == "" {
		nonce.ID = "wno_test"
	}
	m.nonces[nonce.Value] = nonce
	return nil
}

func (m *memoryWallets) FindNonce(ctx context.Context, address string, nonce string) (*walletdomain.Nonce, error) {
	n, ok := m.nonces[nonce]
	if !ok || n.Address != address {
		return nil, repository.ErrNotFound
	}
	return n, nil
}

func (m *memoryWallets) MarkNonceUsed(ctx context.Context, nonceID string) error {
	for _, nonce := range m.nonces {
		if nonce.ID == nonceID && nonce.UsedAt == nil {
			now := testNow
			nonce.UsedAt = &now
			return nil
		}
	}
	return repository.ErrNotFound
}

type memoryUsers struct {
	byID    map[string]*user.User
	byPhone map[string]*user.User
}

func newMemoryUsers() *memoryUsers {
	return &memoryUsers{
		byID:    map[string]*user.User{},
		byPhone: map[string]*user.User{},
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

func (m *memoryClients) Create(ctx context.Context, c *client.Client) error {
	m.byClientID[c.ClientID] = c
	return nil
}

func (m *memoryClients) List(ctx context.Context) ([]client.Client, error) {
	return nil, nil
}

type memoryRefreshTokens struct{}

func newMemoryRefreshTokens() *memoryRefreshTokens { return &memoryRefreshTokens{} }

func (m *memoryRefreshTokens) Create(ctx context.Context, refreshToken *token.RefreshToken) error {
	return nil
}

func (m *memoryRefreshTokens) FindByHash(ctx context.Context, tokenHash string) (*token.RefreshToken, error) {
	return nil, repository.ErrNotFound
}

func (m *memoryRefreshTokens) Revoke(ctx context.Context, id string) error {
	return nil
}

type memoryActivity struct {
	loginCount int
}

func newMemoryActivity() *memoryActivity { return &memoryActivity{} }

func (m *memoryActivity) RecordLogin(ctx context.Context, log *audit.LoginLog) error {
	m.loginCount++
	return nil
}

func (m *memoryActivity) UpsertUserClientLogin(ctx context.Context, userID string, clientID string) error {
	return nil
}

type fakeTokenHasher struct{}

func (fakeTokenHasher) HashToken(raw string) string { return "hash:" + raw }

type fakeIssuer struct{}

func (fakeIssuer) IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error) {
	return &token.Pair{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    testNow.Add(time.Hour),
	}, nil
}

func (fakeIssuer) RefreshTokenTTL() time.Duration { return time.Hour }
