package phone

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
	ErrInvalidClient = "CLIENT_INVALID"
	ErrInvalidInput  = "PHONE_INVALID_INPUT"
	ErrInvalidCode   = "PHONE_INVALID_CODE"
	ErrDisabled      = "PHONE_LOGIN_DISABLED"
	ErrSendFailed    = "PHONE_SEND_FAILED"
)

// Clock supplies time to keep phone-code flows deterministic in tests.
type Clock interface {
	Now() time.Time
}

// TokenIssuer issues access and refresh tokens for phone login.
type TokenIssuer interface {
	IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error)
	RefreshTokenTTL() time.Duration
}

// TokenHasher hashes opaque refresh tokens before persistence.
type TokenHasher interface {
	HashToken(raw string) string
}

// SMSMessage describes the verification SMS content requested by the phone usecase.
// SMSMessage 描述手机号登录用例请求发送的短信验证码内容。
type SMSMessage struct {
	Phone string
	Code  string
}

// SMSProvider sends verification SMS messages through an external adapter.
// SMSProvider 是短信发送端口，具体云厂商实现应放在 infrastructure 层。
type SMSProvider interface {
	SendSMS(ctx context.Context, msg SMSMessage) error
}

// Service orchestrates phone verification-code login.
// Service 只编排验证码、用户、token，不直接依赖任何短信云厂商 SDK。
type Service struct {
	users         repository.UserRepository
	clients       repository.ClientRepository
	refreshTokens repository.RefreshTokenRepository
	activity      repository.ActivityRepository
	codes         repository.PhoneCodeRepository
	sender        SMSProvider
	tokenHasher   TokenHasher
	issuer        TokenIssuer
	enabled       bool
	codeTTL       time.Duration
	devCode       string
	exposeDevCode bool
	clock         Clock
}

// Dependencies contains external ports required by phone login.
type Dependencies struct {
	Users         repository.UserRepository
	Clients       repository.ClientRepository
	RefreshTokens repository.RefreshTokenRepository
	Activity      repository.ActivityRepository
	Codes         repository.PhoneCodeRepository
	Sender        SMSProvider
	TokenHasher   TokenHasher
	Issuer        TokenIssuer
	Enabled       bool
	CodeTTL       time.Duration
	DevCode       string
	ExposeDevCode bool
	Clock         Clock
}

// CodeRequest is the input for requesting a phone verification code.
type CodeRequest struct {
	ClientID string
	Phone    string
}

// CodeResult describes the created phone verification code.
type CodeResult struct {
	Phone     string
	ExpiresAt time.Time
	DevCode   string
}

// LoginRequest is the input for phone-code login.
type LoginRequest struct {
	ClientID  string
	Phone     string
	Code      string
	IP        string
	UserAgent string
}

// LoginResult is returned after a successful phone-code login.
type LoginResult struct {
	UserID   string
	Username string
	Phone    string
	Token    *token.Pair
}

// NewService creates the phone usecase service.
func NewService(deps Dependencies) *Service {
	return &Service{
		users:         deps.Users,
		clients:       deps.Clients,
		refreshTokens: deps.RefreshTokens,
		activity:      deps.Activity,
		codes:         deps.Codes,
		sender:        deps.Sender,
		tokenHasher:   deps.TokenHasher,
		issuer:        deps.Issuer,
		enabled:       deps.Enabled,
		codeTTL:       deps.CodeTTL,
		devCode:       deps.DevCode,
		exposeDevCode: deps.ExposeDevCode,
		clock:         deps.Clock,
	}
}

// RequestCode creates a short-lived verification code for phone login.
// RequestCode 通过 SMSProvider 端口发送验证码，真实供应商由 infrastructure 注入。
func (s *Service) RequestCode(ctx context.Context, req CodeRequest) (*CodeResult, error) {
	clientID := defaultClientID(req.ClientID)
	phone := normalizePhone(req.Phone)
	if !s.enabled {
		return nil, domain.NewError(ErrDisabled, "phone login is disabled")
	}
	if phone == "" {
		return nil, domain.NewError(ErrInvalidInput, "phone is required")
	}
	client, err := s.clients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}

	code := s.devCode
	if code == "" {
		code = "123456"
	}
	expiresAt := s.clock.Now().UTC().Add(s.codeTTL)
	if err := s.codes.Save(ctx, phone, code, expiresAt); err != nil {
		return nil, err
	}
	if s.sender != nil {
		if err := s.sender.SendSMS(ctx, SMSMessage{Phone: phone, Code: code}); err != nil {
			return nil, domain.WrapError(ErrSendFailed, "send phone verification code failed", err)
		}
	}
	devCode := ""
	if s.exposeDevCode {
		devCode = s.devCode
	}
	return &CodeResult{Phone: phone, ExpiresAt: expiresAt, DevCode: devCode}, nil
}

// Login verifies the phone code, creates the user if needed, and issues tokens.
// Login 消费验证码后统一走本服务 JWT/refresh-token 签发流程。
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	clientID := defaultClientID(req.ClientID)
	phone := normalizePhone(req.Phone)
	code := strings.TrimSpace(req.Code)
	if !s.enabled {
		return nil, domain.NewError(ErrDisabled, "phone login is disabled")
	}
	if phone == "" || code == "" {
		return nil, domain.NewError(ErrInvalidInput, "phone and code are required")
	}
	client, err := s.clients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}
	ok, err := s.codes.Verify(ctx, phone, code, s.clock.Now().UTC())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.NewError(ErrInvalidCode, "invalid or expired phone code")
	}

	u, err := s.users.FindByPhone(ctx, phone)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if u == nil {
		u = &user.User{
			Username: "phone_" + safePhoneSuffix(phone),
			Phone:    phone,
			Status:   user.StatusActive,
		}
		if err := s.users.Create(ctx, u); err != nil {
			return nil, err
		}
	}
	if !u.IsActive() {
		return nil, domain.NewError(ErrInvalidCode, "phone user is unavailable")
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
	if err := s.refreshTokens.Create(ctx, &token.RefreshToken{
		UserID:    u.ID,
		ClientID:  client.ClientID,
		TokenHash: s.tokenHasher.HashToken(pair.RefreshToken),
		ExpiresAt: s.clock.Now().UTC().Add(s.issuer.RefreshTokenTTL()),
	}); err != nil {
		return nil, err
	}
	if err := s.users.UpdateLoginInfo(ctx, u.ID); err != nil {
		return nil, err
	}
	if s.activity != nil {
		if err := s.activity.RecordLogin(ctx, &audit.LoginLog{
			UserID:      u.ID,
			ClientID:    client.ClientID,
			LoginMethod: audit.LoginMethodPhone,
			IP:          req.IP,
			UserAgent:   req.UserAgent,
			Success:     true,
		}); err != nil {
			return nil, err
		}
		if err := s.activity.UpsertUserClientLogin(ctx, u.ID, client.ClientID); err != nil {
			return nil, err
		}
	}
	return &LoginResult{UserID: u.ID, Username: u.Username, Phone: u.Phone, Token: pair}, nil
}

func normalizePhone(phone string) string {
	return strings.ReplaceAll(strings.TrimSpace(phone), " ", "")
}

func safePhoneSuffix(phone string) string {
	phone = strings.TrimLeft(phone, "+")
	if len(phone) <= 4 {
		return phone
	}
	return phone[len(phone)-4:]
}

func defaultClientID(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "default"
	}
	return clientID
}
