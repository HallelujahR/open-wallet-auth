package auth

import (
	"context"
	"strings"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	oauthdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	walletdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
)

// loginMethodSummary loads current login bindings for safety checks.
// loginMethodSummary 加载当前登录方式绑定，用于解绑前的安全校验。
func (s *Service) loginMethodSummary(ctx context.Context, userID string) (*loginMethodSummary, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil || !u.IsActive() {
		return nil, domain.NewError(ErrInvalidCredentials, "authenticated user is unavailable")
	}
	var wallets []walletdomain.UserWallet
	if s.wallets != nil {
		wallets, err = s.wallets.ListByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
	}
	var accounts []oauthdomain.Account
	if s.accounts != nil {
		accounts, err = s.accounts.ListByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
	}
	return &loginMethodSummary{user: u, wallets: wallets, accounts: accounts}, nil
}

type loginMethodSummary struct {
	user     *user.User
	wallets  []walletdomain.UserWallet
	accounts []oauthdomain.Account
}

// total counts independent ways the user can still authenticate.
// total 统计用户仍可用于登录的独立方式数量。
func (s loginMethodSummary) total() int {
	total := 0
	if s.user.Email != "" {
		total++
	}
	if s.user.Phone != "" {
		total++
	}
	total += len(s.wallets)
	total += len(s.accounts)
	return total
}

// names returns stable login-method names for profile responses.
// names 返回稳定的登录方式名称，用于资料接口响应。
func (s loginMethodSummary) names() []string {
	names := []string{}
	if s.user.Email != "" {
		names = append(names, "email")
	}
	if s.user.Phone != "" {
		names = append(names, "phone")
	}
	if len(s.wallets) > 0 {
		names = append(names, "wallet")
	}
	if len(s.accounts) > 0 {
		names = append(names, "oauth")
	}
	return names
}

// normalizeUnbindInput validates current-user unbinding input.
// normalizeUnbindInput 校验当前用户解绑请求输入。
func normalizeUnbindInput(req UnbindRequest) (string, string, error) {
	userID := strings.TrimSpace(req.UserID)
	bindingID := strings.TrimSpace(req.BindingID)
	if userID == "" || bindingID == "" {
		return "", "", domain.NewError(ErrInvalidInput, "user id and binding id are required")
	}
	return userID, bindingID, nil
}

// walletBindingExists reports whether a wallet binding belongs to the user.
// walletBindingExists 判断钱包绑定是否属于当前用户。
func walletBindingExists(wallets []walletdomain.UserWallet, walletID string) bool {
	for _, wallet := range wallets {
		if wallet.ID == walletID {
			return true
		}
	}
	return false
}

// oauthBindingExists reports whether an OAuth binding belongs to the user.
// oauthBindingExists 判断 OAuth 绑定是否属于当前用户。
func oauthBindingExists(accounts []oauthdomain.Account, accountID string) bool {
	for _, account := range accounts {
		if account.ID == accountID {
			return true
		}
	}
	return false
}
