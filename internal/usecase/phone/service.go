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

// Service orchestrates phone verification-code login.
type Service struct {
	users         repository.UserRepository
	clients       repository.ClientRepository
	refreshTokens repository.RefreshTokenRepository
	activity      repository.ActivityRepository
	codes         repository.PhoneCodeRepository
	tokenHasher   TokenHasher
	issuer        TokenIssuer
	codeTTL       time.Duration
	devCode       string
	clock         Clock
}

// Dependencies contains external ports required by phone login.
type Dependencies struct {
	Users         repository.UserRepository
	Clients       repository.ClientRepository
	RefreshTokens repository.RefreshTokenRepository
	Activity      repository.ActivityRepository
	Codes         repository.PhoneCodeRepository
	TokenHasher   TokenHasher
	Issuer        TokenIssuer
	CodeTTL       time.Duration
	DevCode       string
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
		tokenHasher:   deps.TokenHasher,
		issuer:        deps.Issuer,
		codeTTL:       deps.CodeTTL,
		devCode:       deps.DevCode,
		clock:         deps.Clock,
	}
}

// RequestCode creates a short-lived verification code for phone login.
func (s *Service) RequestCode(ctx context.Context, req CodeRequest) (*CodeResult, error) {
	clientID := defaultClientID(req.ClientID)
	phone := normalizePhone(req.Phone)
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
	return &CodeResult{Phone: phone, ExpiresAt: expiresAt, DevCode: s.devCode}, nil
}

// Login verifies the phone code, creates the user if needed, and issues tokens.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	clientID := defaultClientID(req.ClientID)
	phone := normalizePhone(req.Phone)
	code := strings.TrimSpace(req.Code)
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
