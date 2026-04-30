package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrInvalidInput = "ADMIN_INVALID_INPUT"
	ErrUserNotFound = "ADMIN_USER_NOT_FOUND"
)

// Service orchestrates identity-management queries and commands.
// Service 编排统一身份管理的查询和操作，不处理具体 HTTP 或数据库细节。
type Service struct {
	users    repository.AdminUserRepository
	activity repository.AdminActivityRepository
	wallets  repository.AdminWalletRepository
	accounts repository.AdminOAuthAccountRepository
}

// Dependencies contains external ports required by identity management.
// Dependencies 汇总身份管理用例需要的仓储端口。
type Dependencies struct {
	Users    repository.AdminUserRepository
	Activity repository.AdminActivityRepository
	Wallets  repository.AdminWalletRepository
	Accounts repository.AdminOAuthAccountRepository
}

// UserListRequest is the input for listing identity users.
// UserListRequest 是查询身份用户列表的用例输入。
type UserListRequest struct {
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

// UserListResult contains paginated identity users.
// UserListResult 返回分页后的身份用户列表。
type UserListResult struct {
	Users    []user.User
	Total    int64
	Page     int
	PageSize int
}

// UserDetailResult aggregates one identity user's login and binding data.
// UserDetailResult 聚合单个身份用户的登录系统和绑定信息。
type UserDetailResult struct {
	User     user.User
	Clients  []audit.UserClient
	Wallets  []wallet.UserWallet
	Accounts []oauth.Account
}

// UpdateUserStatusRequest is the input for enabling or disabling an identity.
// UpdateUserStatusRequest 是启用或禁用身份用户的用例输入。
type UpdateUserStatusRequest struct {
	UserID string
	Status string
}

// LoginLogListRequest is the input for listing login audit logs.
// LoginLogListRequest 是查询登录审计日志的用例输入。
type LoginLogListRequest struct {
	UserID   string
	ClientID string
	Page     int
	PageSize int
}

// LoginLogListResult contains paginated login audit logs.
// LoginLogListResult 返回分页后的登录审计日志。
type LoginLogListResult struct {
	Logs     []audit.LoginLog
	Total    int64
	Page     int
	PageSize int
}

// NewService creates the identity-management usecase service.
// NewService 创建身份管理用例服务，并注入外部端口。
func NewService(deps Dependencies) *Service {
	return &Service{users: deps.Users, activity: deps.Activity, wallets: deps.Wallets, accounts: deps.Accounts}
}

// ListUsers returns identity users with pagination and simple search.
// ListUsers 按分页和关键字查询统一身份用户。
func (s *Service) ListUsers(ctx context.Context, req UserListRequest) (*UserListResult, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	status := user.Status(strings.TrimSpace(req.Status))
	if status != "" && !validStatus(status) {
		return nil, domain.NewError(ErrInvalidInput, "invalid user status")
	}
	users, total, err := s.users.List(ctx, repository.UserListFilter{
		Keyword:  strings.TrimSpace(req.Keyword),
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	return &UserListResult{Users: users, Total: total, Page: page, PageSize: pageSize}, nil
}

// GetUserDetail returns one identity user with login clients and linked accounts.
// GetUserDetail 查询单个身份用户，并返回登录系统和账号绑定信息。
func (s *Service) GetUserDetail(ctx context.Context, userID string) (*UserDetailResult, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domain.NewError(ErrInvalidInput, "user_id is required")
	}
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		return nil, domain.NewError(ErrUserNotFound, "user not found")
	}
	clients, err := s.activity.ListUserClients(ctx, userID)
	if err != nil {
		return nil, err
	}
	wallets, err := s.wallets.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	accounts, err := s.accounts.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &UserDetailResult{User: *u, Clients: clients, Wallets: wallets, Accounts: accounts}, nil
}

// UpdateUserStatus changes whether an identity can authenticate.
// UpdateUserStatus 修改身份用户是否允许继续登录。
func (s *Service) UpdateUserStatus(ctx context.Context, req UpdateUserStatusRequest) error {
	userID := strings.TrimSpace(req.UserID)
	status := user.Status(strings.TrimSpace(req.Status))
	if userID == "" || !validStatus(status) {
		return domain.NewError(ErrInvalidInput, "user_id and valid status are required")
	}
	if err := s.users.UpdateStatus(ctx, userID, status); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.NewError(ErrUserNotFound, "user not found")
		}
		return err
	}
	return nil
}

// ListLoginLogs returns login audit logs with pagination.
// ListLoginLogs 按分页查询登录审计日志。
func (s *Service) ListLoginLogs(ctx context.Context, req LoginLogListRequest) (*LoginLogListResult, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	logs, total, err := s.activity.ListLoginLogs(ctx, repository.LoginLogFilter{
		UserID:   strings.TrimSpace(req.UserID),
		ClientID: strings.TrimSpace(req.ClientID),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	return &LoginLogListResult{Logs: logs, Total: total, Page: page, PageSize: pageSize}, nil
}

// normalizePage returns bounded pagination values for management lists.
// normalizePage 返回受限制的分页参数，避免管理接口一次性查询过多数据。
func normalizePage(page int, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// validStatus reports whether a user status is accepted by management APIs.
// validStatus 判断管理接口是否接受该用户状态。
func validStatus(status user.Status) bool {
	return status == user.StatusActive || status == user.StatusSuspended || status == user.StatusDeleted
}
