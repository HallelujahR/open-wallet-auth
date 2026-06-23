package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func TestServiceListUsers(t *testing.T) {
	service := newTestService()

	result, err := service.ListUsers(context.Background(), UserListRequest{Keyword: "alice", PageSize: 200})
	if err != nil {
		t.Fatalf("ListUsers returned error: %v", err)
	}
	if result.Total != 1 || result.PageSize != 100 {
		t.Fatalf("unexpected pagination result: total=%d page_size=%d", result.Total, result.PageSize)
	}
	if result.Users[0].ID != "usr_1" {
		t.Fatalf("unexpected user id: %s", result.Users[0].ID)
	}
}

func TestServiceGetUserDetail(t *testing.T) {
	service := newTestService()

	result, err := service.GetUserDetail(context.Background(), "usr_1")
	if err != nil {
		t.Fatalf("GetUserDetail returned error: %v", err)
	}
	if result.User.ID != "usr_1" {
		t.Fatalf("unexpected user id: %s", result.User.ID)
	}
	if len(result.Clients) != 1 || len(result.Wallets) != 1 || len(result.Accounts) != 1 || len(result.Sessions) != 1 {
		t.Fatalf("expected one client, wallet, oauth account, and session")
	}
}

func TestServiceUpdateUserStatus(t *testing.T) {
	service := newTestService()

	if err := service.UpdateUserStatus(context.Background(), UpdateUserStatusRequest{UserID: "usr_1", Status: "suspended"}); err != nil {
		t.Fatalf("UpdateUserStatus returned error: %v", err)
	}
	if service.users.(*memoryAdminUsers).status != user.StatusSuspended {
		t.Fatalf("status was not updated")
	}
}

func TestServiceUpdateUserStatusRejectsInvalidStatus(t *testing.T) {
	service := newTestService()

	if err := service.UpdateUserStatus(context.Background(), UpdateUserStatusRequest{UserID: "usr_1", Status: "locked"}); err == nil {
		t.Fatalf("expected invalid status error")
	}
}

func TestServiceSetUserPassword(t *testing.T) {
	service := newTestService()

	if err := service.SetUserPassword(context.Background(), SetUserPasswordRequest{UserID: "usr_1", Password: "new-password"}); err != nil {
		t.Fatalf("SetUserPassword returned error: %v", err)
	}
	users := service.users.(*memoryAdminUsers)
	if users.passwordHash != "hash:new-password" {
		t.Fatalf("password hash was not updated: %s", users.passwordHash)
	}
	if !service.activity.(*memoryAdminActivity).recorded {
		t.Fatalf("expected security event to be recorded")
	}
}

func TestServiceSetUserPasswordRejectsShortPassword(t *testing.T) {
	service := newTestService()

	if err := service.SetUserPassword(context.Background(), SetUserPasswordRequest{UserID: "usr_1", Password: "short"}); err == nil {
		t.Fatalf("expected short password error")
	}
}

func TestServiceListLoginLogs(t *testing.T) {
	service := newTestService()

	result, err := service.ListLoginLogs(context.Background(), LoginLogListRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("ListLoginLogs returned error: %v", err)
	}
	if result.Total != 1 || result.Logs[0].UserID != "usr_1" {
		t.Fatalf("unexpected login log result")
	}
}

func TestServiceListSecurityEvents(t *testing.T) {
	service := newTestService()

	result, err := service.ListSecurityEvents(context.Background(), SecurityEventListRequest{UserID: "usr_1", EventType: "bind_email"})
	if err != nil {
		t.Fatalf("ListSecurityEvents returned error: %v", err)
	}
	if result.Total != 1 || result.Events[0].EventType != audit.SecurityEventBindEmail {
		t.Fatalf("unexpected security event result")
	}
}

func TestServiceListSessions(t *testing.T) {
	service := newTestService()

	result, err := service.ListSessions(context.Background(), SessionListRequest{UserID: "usr_1", ActiveOnly: true})
	if err != nil {
		t.Fatalf("ListSessions returned error: %v", err)
	}
	if len(result.Sessions) != 1 || result.Sessions[0].ID != "rft_1" {
		t.Fatalf("unexpected sessions result")
	}
}

func TestServiceRevokeUserSessions(t *testing.T) {
	service := newTestService()

	result, err := service.RevokeUserSessions(context.Background(), RevokeUserSessionsRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("RevokeUserSessions returned error: %v", err)
	}
	if result.Revoked != 1 {
		t.Fatalf("unexpected revoked count: %d", result.Revoked)
	}
}

func TestServiceUnbindWallet(t *testing.T) {
	service := newTestService()

	if err := service.UnbindWallet(context.Background(), UnbindRequest{UserID: "usr_1", BindingID: "wal_1"}); err != nil {
		t.Fatalf("UnbindWallet returned error: %v", err)
	}
}

func TestServiceUnbindOAuthAccount(t *testing.T) {
	service := newTestService()

	if err := service.UnbindOAuthAccount(context.Background(), UnbindRequest{UserID: "usr_1", BindingID: "oac_1"}); err != nil {
		t.Fatalf("UnbindOAuthAccount returned error: %v", err)
	}
}

func newTestService() *Service {
	users := &memoryAdminUsers{status: user.StatusActive}
	activity := &memoryAdminActivity{}
	return NewService(Dependencies{
		Users:    users,
		Activity: activity,
		Wallets:  memoryAdminWallets{},
		Accounts: memoryAdminAccounts{},
		Sessions: memoryAdminSessions{},
		Hasher:   fakeAdminHasher{},
	})
}

type memoryAdminUsers struct {
	status       user.Status
	passwordHash string
}

func (m *memoryAdminUsers) FindByID(ctx context.Context, id string) (*user.User, error) {
	if id != "usr_1" {
		return nil, repository.ErrNotFound
	}
	return &user.User{ID: "usr_1", Username: "alice", Email: "alice@example.com", Status: m.status, CreatedAt: testTime, UpdatedAt: testTime}, nil
}

func (m *memoryAdminUsers) List(ctx context.Context, filter repository.UserListFilter) ([]user.User, int64, error) {
	if filter.Status != "" && filter.Status != m.status {
		return nil, 0, nil
	}
	return []user.User{{ID: "usr_1", Username: "alice", Email: "alice@example.com", Status: m.status, CreatedAt: testTime, UpdatedAt: testTime}}, 1, nil
}

func (m *memoryAdminUsers) UpdateStatus(ctx context.Context, userID string, status user.Status) error {
	if userID != "usr_1" {
		return repository.ErrNotFound
	}
	if status == "" {
		return errors.New("empty status")
	}
	m.status = status
	return nil
}

func (m *memoryAdminUsers) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	if userID != "usr_1" {
		return repository.ErrNotFound
	}
	m.passwordHash = passwordHash
	return nil
}

type fakeAdminHasher struct{}

func (fakeAdminHasher) Hash(plain string) (string, error) {
	return "hash:" + plain, nil
}

type memoryAdminActivity struct {
	recorded bool
}

func (*memoryAdminActivity) ListLoginLogs(ctx context.Context, filter repository.LoginLogFilter) ([]audit.LoginLog, int64, error) {
	return []audit.LoginLog{{ID: "log_1", UserID: "usr_1", ClientID: "default", LoginMethod: audit.LoginMethodPassword, Success: true, CreatedAt: testTime}}, 1, nil
}

func (*memoryAdminActivity) ListSecurityEvents(ctx context.Context, filter repository.SecurityEventFilter) ([]audit.SecurityEvent, int64, error) {
	return []audit.SecurityEvent{{ID: "sec_1", UserID: "usr_1", EventType: audit.SecurityEventBindEmail, TargetType: "email", TargetID: "alice@example.com", Success: true, CreatedAt: testTime}}, 1, nil
}

func (*memoryAdminActivity) ListUserClients(ctx context.Context, userID string) ([]audit.UserClient, error) {
	return []audit.UserClient{{UserID: userID, ClientID: "default", LoginCount: 2, Status: "active", FirstLoginAt: testTime, LastLoginAt: testTime}}, nil
}

func (m *memoryAdminActivity) RecordSecurityEvent(ctx context.Context, event *audit.SecurityEvent) error {
	m.recorded = event != nil && event.EventType == audit.SecurityEventAdminSetPassword
	return nil
}

type memoryAdminWallets struct{}

func (memoryAdminWallets) ListByUserID(ctx context.Context, userID string) ([]wallet.UserWallet, error) {
	return []wallet.UserWallet{{ID: "wal_1", UserID: userID, ChainType: wallet.ChainTypeEVM, Address: "0x0000000000000000000000000000000000000001", VerifiedAt: testTime, CreatedAt: testTime}}, nil
}

func (memoryAdminWallets) DeleteByID(ctx context.Context, userID string, walletID string) error {
	if userID != "usr_1" || walletID != "wal_1" {
		return repository.ErrNotFound
	}
	return nil
}

type memoryAdminAccounts struct{}

func (memoryAdminAccounts) ListByUserID(ctx context.Context, userID string) ([]oauth.Account, error) {
	return []oauth.Account{{ID: "oac_1", UserID: userID, Provider: "github", ProviderSubject: "10001", CreatedAt: testTime, UpdatedAt: testTime}}, nil
}

func (memoryAdminAccounts) DeleteByID(ctx context.Context, userID string, accountID string) error {
	if userID != "usr_1" || accountID != "oac_1" {
		return repository.ErrNotFound
	}
	return nil
}

type memoryAdminSessions struct{}

func (memoryAdminSessions) List(ctx context.Context, filter repository.RefreshTokenListFilter) ([]token.RefreshToken, error) {
	return []token.RefreshToken{{ID: "rft_1", UserID: "usr_1", ClientID: "default", ExpiresAt: testTime.Add(time.Hour), CreatedAt: testTime}}, nil
}

func (memoryAdminSessions) Revoke(ctx context.Context, id string) error {
	return nil
}

func (memoryAdminSessions) RevokeByUserID(ctx context.Context, userID string) (int64, error) {
	return 1, nil
}

func (memoryAdminSessions) RevokeByUserAndClient(ctx context.Context, userID string, clientID string) (int64, error) {
	return 1, nil
}

var testTime = time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
