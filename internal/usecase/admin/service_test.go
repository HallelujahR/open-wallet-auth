package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
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
	if len(result.Clients) != 1 || len(result.Wallets) != 1 || len(result.Accounts) != 1 {
		t.Fatalf("expected one client, wallet, and oauth account")
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

func newTestService() *Service {
	users := &memoryAdminUsers{status: user.StatusActive}
	return NewService(Dependencies{
		Users:    users,
		Activity: memoryAdminActivity{},
		Wallets:  memoryAdminWallets{},
		Accounts: memoryAdminAccounts{},
	})
}

type memoryAdminUsers struct {
	status user.Status
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

type memoryAdminActivity struct{}

func (memoryAdminActivity) ListLoginLogs(ctx context.Context, filter repository.LoginLogFilter) ([]audit.LoginLog, int64, error) {
	return []audit.LoginLog{{ID: "log_1", UserID: "usr_1", ClientID: "default", LoginMethod: audit.LoginMethodPassword, Success: true, CreatedAt: testTime}}, 1, nil
}

func (memoryAdminActivity) ListUserClients(ctx context.Context, userID string) ([]audit.UserClient, error) {
	return []audit.UserClient{{UserID: userID, ClientID: "default", LoginCount: 2, Status: "active", FirstLoginAt: testTime, LastLoginAt: testTime}}, nil
}

type memoryAdminWallets struct{}

func (memoryAdminWallets) ListByUserID(ctx context.Context, userID string) ([]wallet.UserWallet, error) {
	return []wallet.UserWallet{{ID: "wal_1", UserID: userID, ChainType: wallet.ChainTypeEVM, Address: "0x0000000000000000000000000000000000000001", VerifiedAt: testTime, CreatedAt: testTime}}, nil
}

type memoryAdminAccounts struct{}

func (memoryAdminAccounts) ListByUserID(ctx context.Context, userID string) ([]oauth.Account, error) {
	return []oauth.Account{{ID: "oac_1", UserID: userID, Provider: "github", ProviderSubject: "10001", CreatedAt: testTime, UpdatedAt: testTime}}, nil
}

var testTime = time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
