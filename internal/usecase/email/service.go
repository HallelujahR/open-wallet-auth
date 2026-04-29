package email

import (
	"context"
	"strings"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/message"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrDisabled     = "EMAIL_VERIFICATION_DISABLED"
	ErrInvalidInput = "EMAIL_INVALID_INPUT"
	ErrInvalidCode  = "EMAIL_INVALID_CODE"
	ErrSendFailed   = "EMAIL_SEND_FAILED"
)

// Clock supplies time to keep email verification flows deterministic in tests.
type Clock interface {
	Now() time.Time
}

// Service orchestrates email verification-code sending and checking.
type Service struct {
	codes         repository.EmailCodeRepository
	sender        message.EmailProvider
	enabled       bool
	codeTTL       time.Duration
	devCode       string
	exposeDevCode bool
	clock         Clock
}

// Dependencies contains external ports required by email verification.
type Dependencies struct {
	Codes         repository.EmailCodeRepository
	Sender        message.EmailProvider
	Enabled       bool
	CodeTTL       time.Duration
	DevCode       string
	ExposeDevCode bool
	Clock         Clock
}

// CodeRequest is the input for requesting an email verification code.
type CodeRequest struct {
	Email string
}

// CodeResult describes the created email verification code.
type CodeResult struct {
	Email     string
	ExpiresAt time.Time
	DevCode   string
}

// VerifyRequest is the input for verifying an email code.
type VerifyRequest struct {
	Email string
	Code  string
}

// VerifyResult describes a successful email verification.
type VerifyResult struct {
	Email    string
	Verified bool
}

// NewService creates the email verification usecase service.
func NewService(deps Dependencies) *Service {
	return &Service{
		codes:         deps.Codes,
		sender:        deps.Sender,
		enabled:       deps.Enabled,
		codeTTL:       deps.CodeTTL,
		devCode:       deps.DevCode,
		exposeDevCode: deps.ExposeDevCode,
		clock:         deps.Clock,
	}
}

// RequestCode creates and sends a short-lived email verification code.
func (s *Service) RequestCode(ctx context.Context, req CodeRequest) (*CodeResult, error) {
	if !s.enabled {
		return nil, domain.NewError(ErrDisabled, "email verification is disabled")
	}
	email := normalizeEmail(req.Email)
	if email == "" {
		return nil, domain.NewError(ErrInvalidInput, "email is required")
	}
	code := s.devCode
	if code == "" {
		code = "123456"
	}
	expiresAt := s.clock.Now().UTC().Add(s.codeTTL)
	if err := s.codes.Save(ctx, email, code, expiresAt); err != nil {
		return nil, err
	}
	if s.sender != nil {
		if err := s.sender.SendEmail(ctx, message.EmailMessage{Email: email, Subject: "Your Open Wallet Auth verification code", Code: code}); err != nil {
			return nil, domain.WrapError(ErrSendFailed, "send email verification code failed", err)
		}
	}
	devCode := ""
	if s.exposeDevCode {
		devCode = s.devCode
	}
	return &CodeResult{Email: email, ExpiresAt: expiresAt, DevCode: devCode}, nil
}

// VerifyCode checks and consumes an email verification code.
func (s *Service) VerifyCode(ctx context.Context, req VerifyRequest) (*VerifyResult, error) {
	if !s.enabled {
		return nil, domain.NewError(ErrDisabled, "email verification is disabled")
	}
	email := normalizeEmail(req.Email)
	code := strings.TrimSpace(req.Code)
	if email == "" || code == "" {
		return nil, domain.NewError(ErrInvalidInput, "email and code are required")
	}
	ok, err := s.codes.Verify(ctx, email, code, s.clock.Now().UTC())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.NewError(ErrInvalidCode, "invalid or expired email code")
	}
	return &VerifyResult{Email: email, Verified: true}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
