package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// OAuthAccountRepository persists third-party OAuth account links.
type OAuthAccountRepository struct {
	db *gorm.DB
}

// NewOAuthAccountRepository creates a PostgreSQL OAuth account repository.
func NewOAuthAccountRepository(db *gorm.DB) *OAuthAccountRepository {
	return &OAuthAccountRepository{db: db}
}

func (r *OAuthAccountRepository) FindByProviderSubject(ctx context.Context, provider string, subject string) (*oauth.Account, error) {
	var row model.OAuthAccount
	if err := r.db.WithContext(ctx).Where("provider = ? AND provider_subject = ?", provider, subject).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainOAuthAccount(row), nil
}

func (r *OAuthAccountRepository) Create(ctx context.Context, account *oauth.Account) error {
	now := time.Now().UTC()
	if account.ID == "" {
		account.ID = "oac_" + uuid.NewString()
	}
	account.CreatedAt = now
	account.UpdatedAt = now

	row := model.OAuthAccount{
		ID:                account.ID,
		UserID:            account.UserID,
		Provider:          account.Provider,
		ProviderSubject:   account.ProviderSubject,
		ProviderEmail:     account.ProviderEmail,
		ProviderUsername:  account.ProviderUsername,
		ProviderAvatarURL: account.ProviderAvatarURL,
		CreatedAt:         account.CreatedAt,
		UpdatedAt:         account.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func toDomainOAuthAccount(row model.OAuthAccount) *oauth.Account {
	return &oauth.Account{
		ID:                row.ID,
		UserID:            row.UserID,
		Provider:          row.Provider,
		ProviderSubject:   row.ProviderSubject,
		ProviderEmail:     row.ProviderEmail,
		ProviderUsername:  row.ProviderUsername,
		ProviderAvatarURL: row.ProviderAvatarURL,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

var _ domainrepo.OAuthAccountRepository = (*OAuthAccountRepository)(nil)
