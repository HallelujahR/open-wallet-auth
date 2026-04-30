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
	ErrInvalidRefreshToken = "AUTH_INVALID_REFRESH_TOKEN"
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
	hasher        PasswordHasher
	tokenHasher   TokenHasher
	issuer        TokenIssuer
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

// NewService creates the auth usecase service with its required ports.
// NewService 创建认证用例服务，并通过端口注入外部依赖。
func NewService(
	users repository.UserRepository,
	clients repository.ClientRepository,
	refreshTokens repository.RefreshTokenRepository,
	activity repository.ActivityRepository,
	hasher PasswordHasher,
	tokenHasher TokenHasher,
	issuer TokenIssuer,
) *Service {
	return &Service{
		users:         users,
		clients:       clients,
		refreshTokens: refreshTokens,
		activity:      activity,
		hasher:        hasher,
		tokenHasher:   tokenHasher,
		issuer:        issuer,
	}
}

// Login verifies email/password credentials and issues a token pair.
// Login 校验邮箱密码并签发访问令牌与刷新令牌。
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	req.ClientID = defaultClientID(req.ClientID)
	req.Email = strings.TrimSpace(req.Email)

	client, err := s.clients.FindByClientID(ctx, req.ClientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}

	u, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil || u == nil || !u.IsActive() {
		return nil, domain.NewError(ErrInvalidCredentials, "invalid email or password")
	}

	if !s.hasher.Compare(u.PasswordHash, req.Password) {
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

	if err := s.refreshTokens.Revoke(ctx, refreshToken.ID); err != nil {
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

// defaultClientID normalizes an empty client id to the built-in default client.
// defaultClientID 将空 client_id 归一化为内置 default 业务系统。
func defaultClientID(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "default"
	}
	return clientID
}
