package auth

import (
	"context"
	"strings"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
)

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

// recordSecurityEvent writes sensitive-operation audit data without interrupting business flow.
// recordSecurityEvent 以尽力而为方式记录敏感操作审计，不影响主业务流程。
func (s *Service) recordSecurityEvent(ctx context.Context, event audit.SecurityEvent) {
	if s.activity == nil {
		return
	}
	_ = s.activity.RecordSecurityEvent(ctx, &event)
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
