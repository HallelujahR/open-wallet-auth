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

// Service orchestrates registration, login, refresh, and logout usecases.
// Service 编排注册、登录、刷新和登出业务流程，不直接处理 HTTP 或数据库细节。
type Service struct {
	users         repository.UserRepository
	clients       repository.ClientRepository
	refreshTokens repository.RefreshTokenRepository
	activity      repository.ActivityRepository
	emailCodes    repository.EmailCodeRepository
	phoneCodes    repository.PhoneCodeRepository
	wallets       repository.WalletRepository
	accounts      repository.OAuthAccountRepository
	limiter       repository.RateLimiter
	hasher        PasswordHasher
	tokenHasher   TokenHasher
	issuer        TokenIssuer
	rateLimit     bool
	loginLimit    int
	loginWindow   time.Duration
}

// NewService creates the auth usecase service with its required ports.
// NewService 创建认证用例服务，并通过端口注入外部依赖。
func NewService(
	users repository.UserRepository,
	clients repository.ClientRepository,
	refreshTokens repository.RefreshTokenRepository,
	activity repository.ActivityRepository,
	emailCodes repository.EmailCodeRepository,
	phoneCodes repository.PhoneCodeRepository,
	wallets repository.WalletRepository,
	accounts repository.OAuthAccountRepository,
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
		phoneCodes:    phoneCodes,
		wallets:       wallets,
		accounts:      accounts,
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

// GetProfile returns the current user's persisted profile and bindings.
// GetProfile 返回当前用户的持久化身份资料和绑定方式。
func (s *Service) GetProfile(ctx context.Context, userID string) (*ProfileResult, error) {
	methods, err := s.loginMethodSummary(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	return &ProfileResult{
		User:         *methods.user,
		Wallets:      methods.wallets,
		Accounts:     methods.accounts,
		LoginMethods: methods.names(),
	}, nil
}

// UpdateProfile updates display-only profile fields for the current user.
// UpdateProfile 更新当前用户的展示型身份资料字段。
func (s *Service) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*ProfileResult, error) {
	userID := strings.TrimSpace(req.UserID)
	username := strings.TrimSpace(req.Username)
	avatar := strings.TrimSpace(req.Avatar)
	if userID == "" || username == "" {
		return nil, domain.NewError(ErrInvalidInput, "user id and username are required")
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(ErrInvalidCredentials, "authenticated user is unavailable")
		}
		return nil, err
	}
	if err := s.users.UpdateProfile(ctx, userID, username, avatar); err != nil {
		return nil, err
	}
	return s.GetProfile(ctx, userID)
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
	s.recordSecurityEvent(ctx, audit.SecurityEvent{
		UserID:     userID,
		EventType:  audit.SecurityEventChangePassword,
		TargetType: "user",
		TargetID:   userID,
		IP:         req.IP,
		UserAgent:  req.UserAgent,
		Success:    true,
	})
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
	s.recordSecurityEvent(ctx, audit.SecurityEvent{
		UserID:     u.ID,
		EventType:  audit.SecurityEventResetPassword,
		TargetType: "email",
		TargetID:   email,
		IP:         req.IP,
		UserAgent:  req.UserAgent,
		Success:    true,
	})
	return nil
}

// BindEmail verifies an email code and binds the email to the current user.
// BindEmail 校验邮箱验证码，并把邮箱绑定到当前用户。
func (s *Service) BindEmail(ctx context.Context, req BindEmailRequest) (*BindContactResult, error) {
	if s.emailCodes == nil {
		return nil, domain.NewError(ErrInvalidInput, "email binding is not configured")
	}
	userID := strings.TrimSpace(req.UserID)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)
	if userID == "" || email == "" || code == "" {
		return nil, domain.NewError(ErrInvalidInput, "user id, email, and code are required")
	}
	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil || !u.IsActive() {
		return nil, domain.NewError(ErrInvalidCredentials, "authenticated user is unavailable")
	}
	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		if existing.ID == userID {
			s.recordSecurityEvent(ctx, audit.SecurityEvent{
				UserID:     userID,
				EventType:  audit.SecurityEventBindEmail,
				TargetType: "email",
				TargetID:   email,
				IP:         req.IP,
				UserAgent:  req.UserAgent,
				Success:    true,
			})
			return &BindContactResult{UserID: userID, Value: email}, nil
		}
		return nil, domain.NewError(ErrEmailAlreadyBound, "email is already bound to another account")
	}
	if u.Email != "" && !strings.EqualFold(u.Email, email) {
		return nil, domain.NewError(ErrEmailAlreadyBound, "current account already has an email")
	}
	ok, err := s.emailCodes.Verify(ctx, email, code, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.NewError(ErrInvalidCode, "invalid or expired email code")
	}
	if err := s.users.UpdateEmail(ctx, userID, email); err != nil {
		return nil, err
	}
	s.recordSecurityEvent(ctx, audit.SecurityEvent{
		UserID:     userID,
		EventType:  audit.SecurityEventBindEmail,
		TargetType: "email",
		TargetID:   email,
		IP:         req.IP,
		UserAgent:  req.UserAgent,
		Success:    true,
	})
	return &BindContactResult{UserID: userID, Value: email}, nil
}

// BindPhone verifies a phone code and binds the phone number to the current user.
// BindPhone 校验手机号验证码，并把手机号绑定到当前用户。
func (s *Service) BindPhone(ctx context.Context, req BindPhoneRequest) (*BindContactResult, error) {
	if s.phoneCodes == nil {
		return nil, domain.NewError(ErrInvalidInput, "phone binding is not configured")
	}
	userID := strings.TrimSpace(req.UserID)
	phone := strings.ReplaceAll(strings.TrimSpace(req.Phone), " ", "")
	code := strings.TrimSpace(req.Code)
	if userID == "" || phone == "" || code == "" {
		return nil, domain.NewError(ErrInvalidInput, "user id, phone, and code are required")
	}
	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil || !u.IsActive() {
		return nil, domain.NewError(ErrInvalidCredentials, "authenticated user is unavailable")
	}
	existing, err := s.users.FindByPhone(ctx, phone)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		if existing.ID == userID {
			s.recordSecurityEvent(ctx, audit.SecurityEvent{
				UserID:     userID,
				EventType:  audit.SecurityEventBindPhone,
				TargetType: "phone",
				TargetID:   phone,
				IP:         req.IP,
				UserAgent:  req.UserAgent,
				Success:    true,
			})
			return &BindContactResult{UserID: userID, Value: phone}, nil
		}
		return nil, domain.NewError(ErrPhoneAlreadyBound, "phone is already bound to another account")
	}
	if u.Phone != "" && u.Phone != phone {
		return nil, domain.NewError(ErrPhoneAlreadyBound, "current account already has a phone")
	}
	ok, err := s.phoneCodes.Verify(ctx, phone, code, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.NewError(ErrInvalidCode, "invalid or expired phone code")
	}
	if err := s.users.UpdatePhone(ctx, userID, phone); err != nil {
		return nil, err
	}
	s.recordSecurityEvent(ctx, audit.SecurityEvent{
		UserID:     userID,
		EventType:  audit.SecurityEventBindPhone,
		TargetType: "phone",
		TargetID:   phone,
		IP:         req.IP,
		UserAgent:  req.UserAgent,
		Success:    true,
	})
	return &BindContactResult{UserID: userID, Value: phone}, nil
}

// UnbindEmail removes the current user's email when another login method remains.
// UnbindEmail 在仍保留其他登录方式时解绑当前用户邮箱。
func (s *Service) UnbindEmail(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domain.NewError(ErrInvalidInput, "user id is required")
	}
	methods, err := s.loginMethodSummary(ctx, userID)
	if err != nil {
		return err
	}
	if methods.user.Email == "" {
		return nil
	}
	if methods.total() <= 1 {
		return domain.NewError(ErrLastLoginMethod, "at least one login method must remain")
	}
	if err := s.users.UpdateEmail(ctx, userID, ""); err != nil {
		return err
	}
	s.recordSecurityEvent(ctx, audit.SecurityEvent{
		UserID:     userID,
		EventType:  audit.SecurityEventUnbindEmail,
		TargetType: "email",
		TargetID:   methods.user.Email,
		Success:    true,
	})
	return nil
}

// UnbindPhone removes the current user's phone when another login method remains.
// UnbindPhone 在仍保留其他登录方式时解绑当前用户手机号。
func (s *Service) UnbindPhone(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domain.NewError(ErrInvalidInput, "user id is required")
	}
	methods, err := s.loginMethodSummary(ctx, userID)
	if err != nil {
		return err
	}
	if methods.user.Phone == "" {
		return nil
	}
	if methods.total() <= 1 {
		return domain.NewError(ErrLastLoginMethod, "at least one login method must remain")
	}
	if err := s.users.UpdatePhone(ctx, userID, ""); err != nil {
		return err
	}
	s.recordSecurityEvent(ctx, audit.SecurityEvent{
		UserID:     userID,
		EventType:  audit.SecurityEventUnbindPhone,
		TargetType: "phone",
		TargetID:   methods.user.Phone,
		Success:    true,
	})
	return nil
}

// UnbindWallet removes one wallet binding owned by the current user.
// UnbindWallet 解绑当前用户拥有的一个钱包。
func (s *Service) UnbindWallet(ctx context.Context, req UnbindRequest) error {
	if s.wallets == nil {
		return domain.NewError(ErrInvalidInput, "wallet unbinding is not configured")
	}
	userID, bindingID, err := normalizeUnbindInput(req)
	if err != nil {
		return err
	}
	methods, err := s.loginMethodSummary(ctx, userID)
	if err != nil {
		return err
	}
	if !walletBindingExists(methods.wallets, bindingID) {
		return domain.NewError(ErrBindingNotFound, "wallet binding not found")
	}
	if methods.total() <= 1 && len(methods.wallets) > 0 {
		return domain.NewError(ErrLastLoginMethod, "at least one login method must remain")
	}
	if err := s.wallets.DeleteByID(ctx, userID, bindingID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.NewError(ErrBindingNotFound, "wallet binding not found")
		}
		return err
	}
	s.recordSecurityEvent(ctx, audit.SecurityEvent{
		UserID:     userID,
		EventType:  audit.SecurityEventUnbindWallet,
		TargetType: "wallet",
		TargetID:   bindingID,
		IP:         req.IP,
		UserAgent:  req.UserAgent,
		Success:    true,
	})
	return nil
}

// UnbindOAuthAccount removes one OAuth binding owned by the current user.
// UnbindOAuthAccount 解绑当前用户拥有的一个 OAuth 账号。
func (s *Service) UnbindOAuthAccount(ctx context.Context, req UnbindRequest) error {
	if s.accounts == nil {
		return domain.NewError(ErrInvalidInput, "oauth unbinding is not configured")
	}
	userID, bindingID, err := normalizeUnbindInput(req)
	if err != nil {
		return err
	}
	methods, err := s.loginMethodSummary(ctx, userID)
	if err != nil {
		return err
	}
	if !oauthBindingExists(methods.accounts, bindingID) {
		return domain.NewError(ErrBindingNotFound, "oauth account binding not found")
	}
	if methods.total() <= 1 && len(methods.accounts) > 0 {
		return domain.NewError(ErrLastLoginMethod, "at least one login method must remain")
	}
	if err := s.accounts.DeleteByID(ctx, userID, bindingID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.NewError(ErrBindingNotFound, "oauth account binding not found")
		}
		return err
	}
	s.recordSecurityEvent(ctx, audit.SecurityEvent{
		UserID:     userID,
		EventType:  audit.SecurityEventUnbindOAuth,
		TargetType: "oauth_account",
		TargetID:   bindingID,
		IP:         req.IP,
		UserAgent:  req.UserAgent,
		Success:    true,
	})
	return nil
}
