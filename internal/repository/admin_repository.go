package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
)

// UserListFilter contains pagination and search filters for identity management.
// UserListFilter 描述身份用户列表的分页和搜索条件。
type UserListFilter struct {
	Keyword  string
	Status   user.Status
	Page     int
	PageSize int
}

// LoginLogFilter contains pagination and ownership filters for login logs.
// LoginLogFilter 描述登录日志查询的分页和归属过滤条件。
type LoginLogFilter struct {
	UserID   string
	ClientID string
	Page     int
	PageSize int
}

// AdminUserRepository defines user operations required by identity management.
// AdminUserRepository 定义身份管理接口需要的用户仓储能力。
type AdminUserRepository interface {
	FindByID(ctx context.Context, id string) (*user.User, error)
	List(ctx context.Context, filter UserListFilter) ([]user.User, int64, error)
	UpdateStatus(ctx context.Context, userID string, status user.Status) error
}

// AdminActivityRepository defines audit operations required by identity management.
// AdminActivityRepository 定义身份管理接口需要的登录审计查询能力。
type AdminActivityRepository interface {
	ListLoginLogs(ctx context.Context, filter LoginLogFilter) ([]audit.LoginLog, int64, error)
	ListUserClients(ctx context.Context, userID string) ([]audit.UserClient, error)
}

// AdminWalletRepository defines wallet queries required by identity management.
// AdminWalletRepository 定义身份管理接口需要的钱包查询能力。
type AdminWalletRepository interface {
	ListByUserID(ctx context.Context, userID string) ([]wallet.UserWallet, error)
}

// AdminOAuthAccountRepository defines OAuth account queries required by identity management.
// AdminOAuthAccountRepository 定义身份管理接口需要的第三方账号查询能力。
type AdminOAuthAccountRepository interface {
	ListByUserID(ctx context.Context, userID string) ([]oauth.Account, error)
}
