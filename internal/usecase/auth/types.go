package auth

import (
	"context"
	"time"

	oauthdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	walletdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
)

// PasswordHasher hashes and verifies user passwords.
// PasswordHasher 是密码哈希端口，具体算法由 infrastructure 注入。
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash string, plain string) bool
}

// TokenIssuer issues access and refresh tokens.
// TokenIssuer 是 token 签发端口，用例层不直接依赖 JWT 实现。
type TokenIssuer interface {
	IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error)
	RefreshTokenTTL() time.Duration
}

// TokenHasher hashes opaque refresh tokens before persistence.
// TokenHasher 在刷新令牌入库前做单向哈希，避免明文落库。
type TokenHasher interface {
	HashToken(raw string) string
}

// LoginRequest is the input for password login.
// LoginRequest 是邮箱密码登录的用例输入。
type LoginRequest struct {
	ClientID  string
	Email     string
	Password  string
	IP        string
	UserAgent string
}

// LoginResult is returned after a successful password login.
// LoginResult 是邮箱密码登录成功后的用例输出。
type LoginResult struct {
	UserID   string
	Username string
	Email    string
	Token    *token.Pair
}

// RegisterRequest is the input for email/password registration.
// RegisterRequest 是邮箱密码注册的用例输入。
type RegisterRequest struct {
	ClientID  string
	Username  string
	Email     string
	Password  string
	IP        string
	UserAgent string
}

// RegisterResult is returned after successful registration.
// RegisterResult 是注册成功后的用例输出。
type RegisterResult struct {
	UserID   string
	Username string
	Email    string
	Token    *token.Pair
}

// RefreshRequest is the input for refresh token rotation.
// RefreshRequest 是刷新令牌轮换的用例输入。
type RefreshRequest struct {
	RefreshToken string
	IP           string
	UserAgent    string
}

// RefreshResult is returned after successful refresh token rotation.
// RefreshResult 是刷新令牌轮换成功后的用例输出。
type RefreshResult struct {
	UserID   string
	Username string
	Email    string
	Token    *token.Pair
}

// LogoutRequest is the input for refresh token revocation.
// LogoutRequest 是登出时吊销刷新令牌的用例输入。
type LogoutRequest struct {
	RefreshToken string
}

// ChangePasswordRequest is the input for an authenticated password change.
// ChangePasswordRequest 是已登录用户修改密码的用例输入。
type ChangePasswordRequest struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
	IP              string
	UserAgent       string
}

// ResetPasswordRequest is the input for resetting a password with an email code.
// ResetPasswordRequest 是使用邮箱验证码重置密码的用例输入。
type ResetPasswordRequest struct {
	Email       string
	Code        string
	NewPassword string
	IP          string
	UserAgent   string
}

// BindEmailRequest is the input for binding an email to the current user.
// BindEmailRequest 是当前用户绑定邮箱的用例输入。
type BindEmailRequest struct {
	UserID    string
	Email     string
	Code      string
	IP        string
	UserAgent string
}

// BindPhoneRequest is the input for binding a phone number to the current user.
// BindPhoneRequest 是当前用户绑定手机号的用例输入。
type BindPhoneRequest struct {
	UserID    string
	Phone     string
	Code      string
	IP        string
	UserAgent string
}

// BindContactResult describes the bound contact value.
// BindContactResult 描述绑定后的联系方式。
type BindContactResult struct {
	UserID string
	Value  string
}

// UnbindRequest is the input for removing a login-method binding from current user.
// UnbindRequest 是当前用户移除某个登录方式绑定的用例输入。
type UnbindRequest struct {
	UserID    string
	BindingID string
	IP        string
	UserAgent string
}

// ProfileResult aggregates current identity profile and login-method bindings.
// ProfileResult 聚合当前身份资料和登录方式绑定。
type ProfileResult struct {
	User         user.User
	Wallets      []walletdomain.UserWallet
	Accounts     []oauthdomain.Account
	LoginMethods []string
}

// UpdateProfileRequest is the input for updating display-only profile fields.
// UpdateProfileRequest 是更新展示型身份资料字段的用例输入。
type UpdateProfileRequest struct {
	UserID   string
	Username string
	Avatar   string
}
