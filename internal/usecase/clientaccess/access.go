package clientaccess

import (
	"context"
	"errors"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	// ErrAccessDenied is returned when a client requires allow-list access and the user is not an active member.
	// ErrAccessDenied 表示业务系统已启用白名单，但当前用户没有启用状态的访问授权。
	ErrAccessDenied = "CLIENT_ACCESS_DENIED"
)

// Authorize checks client-level allow-list access and returns the matched member when enabled.
// Authorize 在业务系统启用白名单时检查用户准入授权，未启用时保持旧登录逻辑。
func Authorize(ctx context.Context, clients repository.ClientRepository, authClient *client.Client, userID string) (*client.Member, error) {
	if authClient == nil || !authClient.WhitelistEnabled {
		return nil, nil
	}
	members, ok := clients.(repository.ClientAccessReader)
	if !ok {
		return nil, domain.NewError(ErrAccessDenied, "client access allow-list is unavailable")
	}
	member, err := members.FindActiveMember(ctx, authClient.ClientID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(ErrAccessDenied, "暂无该应用访问权限，请联系管理员")
		}
		return nil, err
	}
	return member, nil
}

// ApplyClaims enriches token claims with app-scoped role and permissions.
// ApplyClaims 将应用成员授权中的角色和权限写入当前业务系统 token。
func ApplyClaims(claims token.Claims, member *client.Member) token.Claims {
	if member == nil {
		return claims
	}
	if member.Role != "" {
		claims.Roles = []string{member.Role}
	}
	claims.Permissions = append([]string(nil), member.Permissions...)
	return claims
}
