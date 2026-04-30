package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
)

// OAuthAccountRepository defines persistence operations for third-party identities.
type OAuthAccountRepository interface {
	FindByProviderSubject(ctx context.Context, provider string, subject string) (*oauth.Account, error)
	ListByUserID(ctx context.Context, userID string) ([]oauth.Account, error)
	Create(ctx context.Context, account *oauth.Account) error
	DeleteByID(ctx context.Context, userID string, accountID string) error
}
