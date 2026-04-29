package email

import (
	"context"
	"strings"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrDisabled     = "EMAIL_VERIFICATION_DISABLED"
	ErrInvalidInput = "EMAIL_INVALID_INPUT"
	ErrInvalidCode  = "EMAIL_INVALID_CODE"
	ErrSendFailed   = "EMAIL_SEND_FAILED"
)

// Clock supplies time to keep email verification flows deterministic in tests.
// Clock 抽象时间来源，便于测试邮箱验证码过期逻辑。
type Clock interface {
	Now() time.Time
}

// EmailMessage describes the verification email requested by the email usecase.
// EmailMessage 描述邮箱验证用例请求发送的邮件内容。
type EmailMessage struct {
	Email   string
	Subject string
	Code    string
}

// EmailProvider sends verification emails through an external adapter.
// EmailProvider 是邮件发送端口，具体服务商实现应放在 infrastructure 层。
type EmailProvider interface {
	SendEmail(ctx context.Context, msg EmailMessage) error
}

// Service orchestrates email verification-code sending and checking.
// Service 只负责邮箱验证码业务规则，邮件网关和服务商细节由 infrastructure 适配。
type Service struct {
	codes         repository.EmailCodeRepository
	sender        EmailProvider
	enabled       bool
	codeTTL       time.Duration
	devCode       string
	exposeDevCode bool
	clock         Clock
}

// Dependencies contains external ports required by email verification.
// Dependencies 汇总邮箱验证用例需要的验证码仓储、邮件发送和时间端口。
type Dependencies struct {
	Codes         repository.EmailCodeRepository
	Sender        EmailProvider
	Enabled       bool
	CodeTTL       time.Duration
	DevCode       string
	ExposeDevCode bool
	Clock         Clock
}

// CodeRequest is the input for requesting an email verification code.
// CodeRequest 是申请邮箱验证码的用例输入。
type CodeRequest struct {
	Email string
}

// CodeResult describes the created email verification code.
// CodeResult 描述已创建邮箱验证码的过期信息。
type CodeResult struct {
	Email     string
	ExpiresAt time.Time
	DevCode   string
}

// VerifyRequest is the input for verifying an email code.
// VerifyRequest 是校验邮箱验证码的用例输入。
type VerifyRequest struct {
	Email string
	Code  string
}

// VerifyResult describes a successful email verification.
// VerifyResult 描述邮箱验证码校验成功的结果。
type VerifyResult struct {
	Email    string
	Verified bool
}

// NewService creates the email verification usecase service.
// NewService 创建邮箱验证用例服务，并注入外部端口。
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
// RequestCode 通过 EmailProvider 端口发送验证码，支持 noop/webhook/自定义服务商实现。
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
		if err := s.sender.SendEmail(ctx, EmailMessage{Email: email, Subject: "Your Open Wallet Auth verification code", Code: code}); err != nil {
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
// VerifyCode 校验成功后消费验证码，避免同一验证码重复使用。
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

// normalizeEmail trims spaces and lowercases emails for consistent verification.
// normalizeEmail 去除空白并转小写，保证邮箱验证码存取口径一致。
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
