package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	oauthdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	walletdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func TestServiceRegisterSuccess(t *testing.T) {
	users := newMemoryUsers()
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

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
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

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
	activity := newMemoryActivity()
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), activity, nil, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

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
	if activity.failedCount != 1 || activity.failureCode != ErrInvalidCredentials {
		t.Fatal("expected failed login audit to be recorded")
	}
}

func TestServiceLoginRejectsRateLimitedEmail(t *testing.T) {
	users := newMemoryUsers()
	users.byEmail["alice@example.com"] = &user.User{
		ID:           "usr_existing",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hash:correct",
		Status:       user.StatusActive,
	}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, nil, nil, nil, denyLimiter{}, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, true, 1, time.Minute)

	_, err := service.Login(context.Background(), LoginRequest{
		ClientID: "default",
		Email:    "alice@example.com",
		Password: "correct",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrRateLimited {
		t.Fatalf("expected %s, got %v", ErrRateLimited, err)
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
	service := NewService(users, defaultClients(), refreshTokens, activity, nil, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

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
	if refreshTokens.byHash["hash:refresh"] == nil {
		t.Fatal("expected new refresh token to be stored")
	}
	if activity.loginCount != 1 || activity.userClientCount != 1 {
		t.Fatal("expected refresh activity to be recorded")
	}
}

func TestServiceChangePasswordUpdatesHash(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_existing"] = &user.User{
		ID:           "usr_existing",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hash:old-password",
		Status:       user.StatusActive,
	}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	err := service.ChangePassword(context.Background(), ChangePasswordRequest{
		UserID:          "usr_existing",
		CurrentPassword: "old-password",
		NewPassword:     "new-password",
	})
	if err != nil {
		t.Fatalf("change password returned error: %v", err)
	}
	if users.byID["usr_existing"].PasswordHash != "hash:new-password" {
		t.Fatal("expected password hash to be updated")
	}
}

func TestServiceChangePasswordRejectsInvalidCurrentPassword(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_existing"] = &user.User{
		ID:           "usr_existing",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hash:old-password",
		Status:       user.StatusActive,
	}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	err := service.ChangePassword(context.Background(), ChangePasswordRequest{
		UserID:          "usr_existing",
		CurrentPassword: "wrong-password",
		NewPassword:     "new-password",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrInvalidCredentials {
		t.Fatalf("expected %s, got %v", ErrInvalidCredentials, err)
	}
}

func TestServiceResetPasswordUpdatesHashWithEmailCode(t *testing.T) {
	users := newMemoryUsers()
	existing := &user.User{
		ID:           "usr_existing",
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hash:old-password",
		Status:       user.StatusActive,
	}
	users.byEmail["alice@example.com"] = existing
	users.byID["usr_existing"] = existing
	codes := newMemoryEmailCodes()
	if err := codes.Save(context.Background(), "alice@example.com", "123456", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("save code returned error: %v", err)
	}
	refreshTokens := newMemoryRefreshTokens()
	refreshTokens.byHash["hash:active_refresh"] = &token.RefreshToken{
		ID:        "rft_active",
		UserID:    "usr_existing",
		ClientID:  "default",
		TokenHash: "hash:active_refresh",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	refreshTokens.byID["rft_active"] = refreshTokens.byHash["hash:active_refresh"]
	service := NewService(users, defaultClients(), refreshTokens, newMemoryActivity(), codes, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	err := service.ResetPassword(context.Background(), ResetPasswordRequest{
		Email:       "alice@example.com",
		Code:        "123456",
		NewPassword: "new-password",
	})
	if err != nil {
		t.Fatalf("reset password returned error: %v", err)
	}
	if users.byEmail["alice@example.com"].PasswordHash != "hash:new-password" {
		t.Fatal("expected password hash to be updated")
	}
	if refreshTokens.byID["rft_active"].RevokedAt == nil {
		t.Fatal("expected existing refresh-token sessions to be revoked")
	}
}

func TestServiceResetPasswordRejectsInvalidCode(t *testing.T) {
	users := newMemoryUsers()
	codes := newMemoryEmailCodes()
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), codes, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	err := service.ResetPassword(context.Background(), ResetPasswordRequest{
		Email:       "alice@example.com",
		Code:        "wrong",
		NewPassword: "new-password",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrInvalidCode {
		t.Fatalf("expected %s, got %v", ErrInvalidCode, err)
	}
}

func TestServiceBindEmailAttachesUnusedEmail(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_wallet"] = &user.User{ID: "usr_wallet", Username: "wallet_user", Status: user.StatusActive}
	codes := newMemoryEmailCodes()
	if err := codes.Save(context.Background(), "alice@example.com", "123456", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("save code returned error: %v", err)
	}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), codes, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	result, err := service.BindEmail(context.Background(), BindEmailRequest{UserID: "usr_wallet", Email: "Alice@Example.com", Code: "123456"})
	if err != nil {
		t.Fatalf("bind email returned error: %v", err)
	}
	if result.Value != "alice@example.com" || users.byID["usr_wallet"].Email != "alice@example.com" {
		t.Fatal("expected normalized email to be bound")
	}
}

func TestServiceBindEmailRejectsEmailOwnedByAnotherUser(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_wallet"] = &user.User{ID: "usr_wallet", Username: "wallet_user", Status: user.StatusActive}
	users.byEmail["alice@example.com"] = &user.User{ID: "usr_email", Email: "alice@example.com", Status: user.StatusActive}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), newMemoryEmailCodes(), nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	_, err := service.BindEmail(context.Background(), BindEmailRequest{UserID: "usr_wallet", Email: "alice@example.com", Code: "123456"})
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrEmailAlreadyBound {
		t.Fatalf("expected %s, got %v", ErrEmailAlreadyBound, err)
	}
}

func TestServiceBindPhoneAttachesUnusedPhone(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_wallet"] = &user.User{ID: "usr_wallet", Username: "wallet_user", Status: user.StatusActive}
	codes := newMemoryPhoneCodes()
	if err := codes.Save(context.Background(), "+8613800000000", "123456", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("save code returned error: %v", err)
	}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, codes, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	result, err := service.BindPhone(context.Background(), BindPhoneRequest{UserID: "usr_wallet", Phone: "+8613800000000", Code: "123456"})
	if err != nil {
		t.Fatalf("bind phone returned error: %v", err)
	}
	if result.Value != "+8613800000000" || users.byID["usr_wallet"].Phone != "+8613800000000" {
		t.Fatal("expected phone to be bound")
	}
}

func TestServiceUnbindWalletKeepsAnotherLoginMethod(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_1"] = &user.User{ID: "usr_1", Email: "alice@example.com", Status: user.StatusActive}
	wallets := newMemoryWallets()
	wallets.items["wal_1"] = walletdomain.UserWallet{ID: "wal_1", UserID: "usr_1", Address: "0x1", ChainType: walletdomain.ChainTypeEVM}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, nil, wallets, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	if err := service.UnbindWallet(context.Background(), UnbindRequest{UserID: "usr_1", BindingID: "wal_1"}); err != nil {
		t.Fatalf("unbind wallet returned error: %v", err)
	}
	if _, ok := wallets.items["wal_1"]; ok {
		t.Fatal("expected wallet to be deleted")
	}
}

func TestServiceUnbindWalletRejectsLastLoginMethod(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_1"] = &user.User{ID: "usr_1", Status: user.StatusActive}
	wallets := newMemoryWallets()
	wallets.items["wal_1"] = walletdomain.UserWallet{ID: "wal_1", UserID: "usr_1", Address: "0x1", ChainType: walletdomain.ChainTypeEVM}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, nil, wallets, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	err := service.UnbindWallet(context.Background(), UnbindRequest{UserID: "usr_1", BindingID: "wal_1"})
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrLastLoginMethod {
		t.Fatalf("expected %s, got %v", ErrLastLoginMethod, err)
	}
}

func TestServiceUnbindOAuthKeepsAnotherLoginMethod(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_1"] = &user.User{ID: "usr_1", Phone: "+8613800000000", Status: user.StatusActive}
	accounts := newMemoryOAuthAccounts()
	accounts.items["oac_1"] = oauthdomain.Account{ID: "oac_1", UserID: "usr_1", Provider: "github", ProviderSubject: "10001"}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, nil, nil, accounts, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	if err := service.UnbindOAuthAccount(context.Background(), UnbindRequest{UserID: "usr_1", BindingID: "oac_1"}); err != nil {
		t.Fatalf("unbind oauth returned error: %v", err)
	}
	if _, ok := accounts.items["oac_1"]; ok {
		t.Fatal("expected oauth account to be deleted")
	}
}

func TestServiceUnbindEmailRejectsLastLoginMethod(t *testing.T) {
	users := newMemoryUsers()
	users.byID["usr_1"] = &user.User{ID: "usr_1", Email: "alice@example.com", Status: user.StatusActive}
	service := NewService(users, defaultClients(), newMemoryRefreshTokens(), newMemoryActivity(), nil, nil, nil, nil, nil, fakeHasher{}, fakeTokenHasher{}, fakeIssuer{}, false, 0, 0)

	err := service.UnbindEmail(context.Background(), "usr_1")
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrLastLoginMethod {
		t.Fatalf("expected %s, got %v", ErrLastLoginMethod, err)
	}
}

type memoryUsers struct {
	byID    map[string]*user.User
	byEmail map[string]*user.User
	byPhone map[string]*user.User
}

func newMemoryUsers() *memoryUsers {
	return &memoryUsers{
		byID:    map[string]*user.User{},
		byEmail: map[string]*user.User{},
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
	u, ok := m.byEmail[email]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
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
	if u.Email != "" {
		m.byEmail[u.Email] = u
	}
	if u.Phone != "" {
		m.byPhone[u.Phone] = u
	}
	return nil
}

func (m *memoryUsers) UpdateLoginInfo(ctx context.Context, userID string) error {
	return nil
}

func (m *memoryUsers) UpdateEmail(ctx context.Context, userID string, email string) error {
	u, ok := m.byID[userID]
	if !ok {
		return repository.ErrNotFound
	}
	if u.Email != "" {
		delete(m.byEmail, u.Email)
	}
	u.Email = email
	if email != "" {
		m.byEmail[email] = u
	}
	return nil
}

func (m *memoryUsers) UpdatePhone(ctx context.Context, userID string, phone string) error {
	u, ok := m.byID[userID]
	if !ok {
		return repository.ErrNotFound
	}
	if u.Phone != "" {
		delete(m.byPhone, u.Phone)
	}
	u.Phone = phone
	if phone != "" {
		m.byPhone[phone] = u
	}
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

type memoryRefreshTokens struct {
	byID   map[string]*token.RefreshToken
	byHash map[string]*token.RefreshToken
}

type memoryActivity struct {
	loginCount      int
	userClientCount int
	failedCount     int
	failureCode     string
}

type memoryEmailCodes struct {
	codes map[string]memoryEmailCode
}

type memoryEmailCode struct {
	code      string
	expiresAt time.Time
}

type memoryWallets struct {
	items map[string]walletdomain.UserWallet
}

func newMemoryWallets() *memoryWallets {
	return &memoryWallets{items: map[string]walletdomain.UserWallet{}}
}

func (m *memoryWallets) FindByAddress(ctx context.Context, chainType walletdomain.ChainType, address string) (*walletdomain.UserWallet, error) {
	for _, item := range m.items {
		if item.ChainType == chainType && item.Address == address {
			return &item, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memoryWallets) ListByUserID(ctx context.Context, userID string) ([]walletdomain.UserWallet, error) {
	items := []walletdomain.UserWallet{}
	for _, item := range m.items {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (m *memoryWallets) CreateWallet(ctx context.Context, w *walletdomain.UserWallet) error {
	if w.ID == "" {
		w.ID = "wal_test"
	}
	m.items[w.ID] = *w
	return nil
}

func (m *memoryWallets) DeleteByID(ctx context.Context, userID string, walletID string) error {
	item, ok := m.items[walletID]
	if !ok || item.UserID != userID {
		return repository.ErrNotFound
	}
	delete(m.items, walletID)
	return nil
}

func (m *memoryWallets) CreateNonce(ctx context.Context, nonce *walletdomain.Nonce) error {
	return nil
}

func (m *memoryWallets) FindNonce(ctx context.Context, address string, nonce string) (*walletdomain.Nonce, error) {
	return nil, repository.ErrNotFound
}

func (m *memoryWallets) MarkNonceUsed(ctx context.Context, nonceID string) error {
	return nil
}

type memoryOAuthAccounts struct {
	items map[string]oauthdomain.Account
}

func newMemoryOAuthAccounts() *memoryOAuthAccounts {
	return &memoryOAuthAccounts{items: map[string]oauthdomain.Account{}}
}

func (m *memoryOAuthAccounts) FindByProviderSubject(ctx context.Context, provider string, subject string) (*oauthdomain.Account, error) {
	for _, item := range m.items {
		if item.Provider == provider && item.ProviderSubject == subject {
			return &item, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memoryOAuthAccounts) ListByUserID(ctx context.Context, userID string) ([]oauthdomain.Account, error) {
	items := []oauthdomain.Account{}
	for _, item := range m.items {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (m *memoryOAuthAccounts) Create(ctx context.Context, account *oauthdomain.Account) error {
	if account.ID == "" {
		account.ID = "oac_test"
	}
	m.items[account.ID] = *account
	return nil
}

func (m *memoryOAuthAccounts) DeleteByID(ctx context.Context, userID string, accountID string) error {
	item, ok := m.items[accountID]
	if !ok || item.UserID != userID {
		return repository.ErrNotFound
	}
	delete(m.items, accountID)
	return nil
}

type memoryPhoneCodes struct {
	codes map[string]memoryEmailCode
}

func newMemoryPhoneCodes() *memoryPhoneCodes {
	return &memoryPhoneCodes{codes: map[string]memoryEmailCode{}}
}

func (m *memoryPhoneCodes) Save(ctx context.Context, phone string, code string, expiresAt time.Time) error {
	m.codes[phone] = memoryEmailCode{code: code, expiresAt: expiresAt}
	return nil
}

func (m *memoryPhoneCodes) Verify(ctx context.Context, phone string, code string, now time.Time) (bool, error) {
	stored, ok := m.codes[phone]
	if !ok || stored.code != code || !stored.expiresAt.After(now) {
		return false, nil
	}
	delete(m.codes, phone)
	return true, nil
}

func newMemoryEmailCodes() *memoryEmailCodes {
	return &memoryEmailCodes{codes: map[string]memoryEmailCode{}}
}

func (m *memoryEmailCodes) Save(ctx context.Context, email string, code string, expiresAt time.Time) error {
	m.codes[email] = memoryEmailCode{code: code, expiresAt: expiresAt}
	return nil
}

func (m *memoryEmailCodes) Verify(ctx context.Context, email string, code string, now time.Time) (bool, error) {
	stored, ok := m.codes[email]
	if !ok || stored.code != code || !stored.expiresAt.After(now) {
		return false, nil
	}
	delete(m.codes, email)
	return true, nil
}

func newMemoryActivity() *memoryActivity {
	return &memoryActivity{}
}

func (m *memoryActivity) RecordLogin(ctx context.Context, log *audit.LoginLog) error {
	if log.Success {
		m.loginCount++
		return nil
	}
	m.failedCount++
	m.failureCode = log.FailureCode
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

func (m *memoryRefreshTokens) Rotate(ctx context.Context, oldTokenID string, newToken *token.RefreshToken) error {
	if err := m.Revoke(ctx, oldTokenID); err != nil {
		return err
	}
	return m.Create(ctx, newToken)
}

func (m *memoryRefreshTokens) RevokeByUserID(ctx context.Context, userID string) (int64, error) {
	now := time.Now()
	var count int64
	for _, refreshToken := range m.byID {
		if refreshToken.UserID == userID && refreshToken.RevokedAt == nil {
			refreshToken.RevokedAt = &now
			count++
		}
	}
	return count, nil
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

type denyLimiter struct{}

func (denyLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return false, nil
}
