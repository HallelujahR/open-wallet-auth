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
// OAuthAccountRepository 是第三方账号绑定仓储端口的 PostgreSQL 适配器。
type OAuthAccountRepository struct {
	db *gorm.DB
}

// NewOAuthAccountRepository creates a PostgreSQL OAuth account repository.
// NewOAuthAccountRepository 创建 PostgreSQL OAuth 账号仓储。
func NewOAuthAccountRepository(db *gorm.DB) *OAuthAccountRepository {
	return &OAuthAccountRepository{db: db}
}

// FindByProviderSubject loads a linked OAuth account by provider identity.
// FindByProviderSubject 按服务商和第三方用户 ID 查询绑定账号。
func (r *OAuthAccountRepository) FindByProviderSubject(ctx context.Context, provider string, subject string) (*oauth.Account, error) {
	var row model.OAuthAccount
	if err := r.db.WithContext(ctx).Where("provider = ? AND provider_subject = ?", provider, subject).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainOAuthAccount(row), nil
}

// Create persists a new OAuth account link.
// Create 持久化新的第三方账号绑定。
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

// ListByUserID returns OAuth accounts linked to one identity user.
// ListByUserID 返回某个身份用户绑定的第三方账号。
func (r *OAuthAccountRepository) ListByUserID(ctx context.Context, userID string) ([]oauth.Account, error) {
	var rows []model.OAuthAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	accounts := make([]oauth.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, *toDomainOAuthAccount(row))
	}
	return accounts, nil
}

// toDomainOAuthAccount converts a database row into a domain OAuth account.
// toDomainOAuthAccount 将数据库行转换为领域 OAuth 账号实体。
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
var _ domainrepo.AdminOAuthAccountRepository = (*OAuthAccountRepository)(nil)
