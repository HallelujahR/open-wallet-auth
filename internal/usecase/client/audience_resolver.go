package client

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
)

// ResolveAudience returns the JWT audience for an active client.
// ResolveAudience 返回可用 client 的 JWT audience，供中间件按业务系统校验 token。
func (s *Service) ResolveAudience(ctx context.Context, clientID string) (string, error) {
	client, err := s.clients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || !client.IsActive() {
		return "", domain.NewError(ErrInvalidClientInput, "invalid client")
	}
	return client.JWTAudience, nil
}
