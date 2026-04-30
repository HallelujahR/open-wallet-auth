package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrEmailAlreadyExists  = "AUTH_EMAIL_ALREADY_EXISTS"
	ErrInvalidClient       = "CLIENT_INVALID"
	ErrInvalidCredentials  = "AUTH_INVALID_CREDENTIALS"
	ErrInvalidInput        = "AUTH_INVALID_INPUT"
	ErrInvalidCode         = "AUTH_INVALID_CODE"
	ErrInvalidRefreshToken = "AUTH_INVALID_REFRESH_TOKEN"
	ErrRateLimited         = "AUTH_RATE_LIMITED"
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

// Service orchestrates registration, login, refresh, and logout usecases.
// Service 编排注册、登录、刷新和登出业务流程，不直接处理 HTTP 或数据库细节。
type Service struct {
	users         repository.UserRepository
	clients       repository.ClientRepository
	refreshTokens repository.RefreshTokenRepository
	activity      repository.ActivityRepository
	emailCodes    repository.EmailCodeRepository
	limiter       repository.RateLimiter
	hasher        PasswordHasher
	tokenHasher   TokenHasher
	issuer        TokenIssuer
	rateLimit     bool
	loginLimit    int
	loginWindow   time.Duration
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
}

// ResetPasswordRequest is the input for resetting a password with an email code.
// ResetPasswordRequest 是使用邮箱验证码重置密码的用例输入。
type ResetPasswordRequest struct {
	Email       string
	Code        string
	NewPassword string
}

// NewService creates the auth usecase service with its required ports.
// NewService 创建认证用例服务，并通过端口注入外部依赖。
func NewService(
	users repository.UserRepository,
	clients repository.ClientRepository,
	refreshTokens repository.RefreshTokenRepository,
	activity repository.ActivityRepository,
	emailCodes repository.EmailCodeRepository,
	limiter repository.RateLimiter,
	hasher PasswordHasher,
	tokenHasher TokenHasher,
	issuer TokenIssuer,
	rateLimit bool,
	loginLimit int,
	loginWindow time.Duration,
) *Service {
	return &Service{
		users:         users,
		clients:       clients,
		refreshTokens: refreshTokens,
		activity:      activity,
		emailCodes:    emailCodes,
		limiter:       limiter,
		hasher:        hasher,
		tokenHasher:   tokenHasher,
		issuer:        issuer,
		rateLimit:     rateLimit,
		loginLimit:    loginLimit,
		loginWindow:   loginWindow,
	}
}

// Login verifies email/password credentials and issues a token pair.
// Login 校验邮箱密码并签发访问令牌与刷新令牌。
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	req.ClientID = defaultClientID(req.ClientID)
	req.Email = strings.TrimSpace(req.Email)
	if err := s.checkLoginLimit(ctx, req.ClientID, req.Email); err != nil {
		return nil, err
	}

	client, err := s.clients.FindByClientID(ctx, req.ClientID)
	if err != nil || client == nil || !client.IsActive() {
		s.recordFailedLogin(ctx, "", req.ClientID, audit.LoginMethodPassword, ErrInvalidClient, req.IP, req.UserAgent)
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}

	u, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil || u == nil || !u.IsActive() {
		userID := ""
		if u != nil {
			userID = u.ID
		}
		s.recordFailedLogin(ctx, userID, client.ClientID, audit.LoginMethodPassword, ErrInvalidCredentials, req.IP, req.UserAgent)
		return nil, domain.NewError(ErrInvalidCredentials, "invalid email or password")
	}

	if !s.hasher.Compare(u.PasswordHash, req.Password) {
		s.recordFailedLogin(ctx, u.ID, client.ClientID, audit.LoginMethodPassword, ErrInvalidCredentials, req.IP, req.UserAgent)
		return nil, domain.NewError(ErrInvalidCredentials, "invalid email or password")
	}

	pair, err := s.issuer.IssuePair(ctx, token.Claims{
		UserID:   u.ID,
		ClientID: client.ClientID,
		Audience: client.JWTAudience,
		Username: u.Username,
		Email:    u.Email,
	})
	if err != nil {
		return nil, err
	}
	if err := s.storeRefreshToken(ctx, u.ID, client.ClientID, pair.RefreshToken, req.IP, req.UserAgent); err != nil {
		return nil, err
	}

	if err := s.users.UpdateLoginInfo(ctx, u.ID); err != nil {
		return nil, err
	}
	if err := s.recordSuccessfulLogin(ctx, u.ID, client.ClientID, audit.LoginMethodPassword, req.IP, req.UserAgent); err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Token:    pair,
	}, nil
}

// Register creates a password user and immediately issues a token pair.
// Register 创建邮箱密码用户，并在注册成功后直接签发 token。
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	req.ClientID = defaultClientID(req.ClientID)
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)

	if req.Email == "" || req.Password == "" {
		return nil, domain.NewError(ErrInvalidInput, "email and password are required")
	}
	if req.Username == "" {
		req.Username = strings.Split(req.Email, "@")[0]
	}

	client, err := s.clients.FindByClientID(ctx, req.ClientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}

	existing, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.NewError(ErrEmailAlreadyExists, "email already exists")
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Status:       user.StatusActive,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	pair, err := s.issuer.IssuePair(ctx, token.Claims{
		UserID:   u.ID,
		ClientID: client.ClientID,
		Audience: client.JWTAudience,
		Username: u.Username,
		Email:    u.Email,
	})
	if err != nil {
		return nil, err
	}
	if err := s.storeRefreshToken(ctx, u.ID, client.ClientID, pair.RefreshToken, req.IP, req.UserAgent); err != nil {
		return nil, err
	}
	if err := s.recordSuccessfulLogin(ctx, u.ID, client.ClientID, audit.LoginMethodPassword, req.IP, req.UserAgent); err != nil {
		return nil, err
	}

	return &RegisterResult{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Token:    pair,
	}, nil
}

// Refresh rotates a valid refresh token and returns a fresh token pair.
// Refresh 轮换有效刷新令牌，旧令牌吊销后签发新的 token 组合。
func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*RefreshResult, error) {
	raw := strings.TrimSpace(req.RefreshToken)
	if raw == "" {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	refreshToken, err := s.refreshTokens.FindByHash(ctx, s.tokenHasher.HashToken(raw))
	if err != nil || refreshToken == nil {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}
	now := time.Now().UTC()
	if refreshToken.IsRevoked() || refreshToken.IsExpired(now) {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	u, err := s.users.FindByID(ctx, refreshToken.UserID)
	if err != nil || u == nil || !u.IsActive() {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	client, err := s.clients.FindByClientID(ctx, refreshToken.ClientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	pair, err := s.issuer.IssuePair(ctx, token.Claims{
		UserID:   u.ID,
		ClientID: client.ClientID,
		Audience: client.JWTAudience,
		Username: u.Username,
		Email:    u.Email,
	})
	if err != nil {
		return nil, err
	}
	if err := s.rotateRefreshToken(ctx, refreshToken.ID, u.ID, client.ClientID, pair.RefreshToken, req.IP, req.UserAgent); err != nil {
		return nil, err
	}
	if err := s.recordSuccessfulLogin(ctx, u.ID, client.ClientID, audit.LoginMethodRefresh, req.IP, req.UserAgent); err != nil {
		return nil, err
	}

	return &RefreshResult{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Token:    pair,
	}, nil
}

// Logout revokes a refresh token so it can no longer be rotated.
// Logout 吊销刷新令牌，使其无法继续换取新的 token。
func (s *Service) Logout(ctx context.Context, req LogoutRequest) error {
	raw := strings.TrimSpace(req.RefreshToken)
	if raw == "" {
		return domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}

	refreshToken, err := s.refreshTokens.FindByHash(ctx, s.tokenHasher.HashToken(raw))
	if err != nil || refreshToken == nil {
		return domain.NewError(ErrInvalidRefreshToken, "invalid refresh token")
	}
	if refreshToken.IsRevoked() {
		return nil
	}
	return s.refreshTokens.Revoke(ctx, refreshToken.ID)
}

// ChangePassword verifies the current password and stores a new password hash.
// ChangePassword 校验当前密码后保存新的密码哈希。
func (s *Service) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	userID := strings.TrimSpace(req.UserID)
	if userID == "" || strings.TrimSpace(req.CurrentPassword) == "" || len(req.NewPassword) < 8 {
		return domain.NewError(ErrInvalidInput, "user id, current password, and a new password with at least 8 characters are required")
	}

	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil || !u.IsActive() || u.PasswordHash == "" {
		return domain.NewError(ErrInvalidCredentials, "invalid current password")
	}
	if !s.hasher.Compare(u.PasswordHash, req.CurrentPassword) {
		return domain.NewError(ErrInvalidCredentials, "invalid current password")
	}
	if s.hasher.Compare(u.PasswordHash, req.NewPassword) {
		return domain.NewError(ErrInvalidInput, "new password must be different")
	}

	hash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	return nil
}

// ResetPassword verifies an email code and replaces the user's password hash.
// ResetPassword 校验邮箱验证码后重置用户密码哈希。
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if s.emailCodes == nil {
		return domain.NewError(ErrInvalidInput, "password reset is not configured")
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)
	if email == "" || code == "" || len(req.NewPassword) < 8 {
		return domain.NewError(ErrInvalidInput, "email, code, and a new password with at least 8 characters are required")
	}

	ok, err := s.emailCodes.Verify(ctx, email, code, time.Now().UTC())
	if err != nil {
		return err
	}
	if !ok {
		return domain.NewError(ErrInvalidCode, "invalid or expired email code")
	}

	u, err := s.users.FindByEmail(ctx, email)
	if err != nil || u == nil || !u.IsActive() {
		return domain.NewError(ErrInvalidCredentials, "invalid email user")
	}
	if s.hasher.Compare(u.PasswordHash, req.NewPassword) {
		return domain.NewError(ErrInvalidInput, "new password must be different")
	}

	hash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, u.ID, hash); err != nil {
		return err
	}
	if _, err := s.refreshTokens.RevokeByUserID(ctx, u.ID); err != nil {
		return err
	}
	return nil
}

// checkLoginLimit verifies password-login rate limits.
// checkLoginLimit 校验邮箱密码登录是否超过频率限制。
func (s *Service) checkLoginLimit(ctx context.Context, clientID string, email string) error {
	if !s.rateLimit || s.limiter == nil || email == "" {
		return nil
	}
	ok, err := s.limiter.Allow(ctx, "auth:login:"+clientID+":"+strings.ToLower(email), s.loginLimit, s.loginWindow)
	if err != nil {
		return err
	}
	if !ok {
		return domain.NewError(ErrRateLimited, "too many login attempts")
	}
	return nil
}

// storeRefreshToken hashes and persists the opaque refresh token.
// storeRefreshToken 将刷新令牌哈希后落库，避免保存明文 token。
func (s *Service) storeRefreshToken(ctx context.Context, userID string, clientID string, raw string, ip string, userAgent string) error {
	return s.refreshTokens.Create(ctx, &token.RefreshToken{
		UserID:    userID,
		ClientID:  clientID,
		TokenHash: s.tokenHasher.HashToken(raw),
		IP:        ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().UTC().Add(s.issuer.RefreshTokenTTL()),
	})
}

// rotateRefreshToken revokes the old token and persists the replacement atomically.
// rotateRefreshToken 原子化吊销旧刷新令牌并保存替换令牌。
func (s *Service) rotateRefreshToken(ctx context.Context, oldTokenID string, userID string, clientID string, raw string, ip string, userAgent string) error {
	return s.refreshTokens.Rotate(ctx, oldTokenID, &token.RefreshToken{
		UserID:    userID,
		ClientID:  clientID,
		TokenHash: s.tokenHasher.HashToken(raw),
		IP:        ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().UTC().Add(s.issuer.RefreshTokenTTL()),
	})
}

// recordSuccessfulLogin writes audit data and the user-client activity relation.
// recordSuccessfulLogin 记录登录审计日志，并维护用户与业务系统的最近登录关系。
func (s *Service) recordSuccessfulLogin(ctx context.Context, userID string, clientID string, method audit.LoginMethod, ip string, userAgent string) error {
	if s.activity == nil {
		return nil
	}
	if err := s.activity.RecordLogin(ctx, &audit.LoginLog{
		UserID:      userID,
		ClientID:    clientID,
		LoginMethod: method,
		IP:          ip,
		UserAgent:   userAgent,
		Success:     true,
	}); err != nil {
		return err
	}
	return s.activity.UpsertUserClientLogin(ctx, userID, clientID)
}

// recordFailedLogin writes a best-effort audit event without changing auth results.
// recordFailedLogin 以尽力而为方式记录失败登录审计，不改变认证接口返回结果。
func (s *Service) recordFailedLogin(ctx context.Context, userID string, clientID string, method audit.LoginMethod, failureCode string, ip string, userAgent string) {
	if s.activity == nil {
		return
	}
	_ = s.activity.RecordLogin(ctx, &audit.LoginLog{
		UserID:      userID,
		ClientID:    clientID,
		LoginMethod: method,
		IP:          ip,
		UserAgent:   userAgent,
		Success:     false,
		FailureCode: failureCode,
	})
}

// defaultClientID normalizes an empty client id to the built-in default client.
// defaultClientID 将空 client_id 归一化为内置 default 业务系统。
func defaultClientID(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "default"
	}
	return clientID
}
