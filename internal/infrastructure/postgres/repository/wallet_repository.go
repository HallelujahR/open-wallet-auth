package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	walletdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	domainrepo "github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

// WalletRepository persists wallet bindings and one-time login nonces.
type WalletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a PostgreSQL wallet repository.
func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) FindByAddress(ctx context.Context, chainType walletdomain.ChainType, address string) (*walletdomain.UserWallet, error) {
	var row model.UserWallet
	if err := r.db.WithContext(ctx).Where("chain_type = ? AND address = ?", string(chainType), address).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainWallet(row), nil
}

func (r *WalletRepository) CreateWallet(ctx context.Context, w *walletdomain.UserWallet) error {
	now := time.Now().UTC()
	if w.ID == "" {
		w.ID = "wal_" + uuid.NewString()
	}
	if w.VerifiedAt.IsZero() {
		w.VerifiedAt = now
	}
	w.CreatedAt = now

	row := model.UserWallet{
		ID:         w.ID,
		UserID:     w.UserID,
		ChainType:  string(w.ChainType),
		Address:    w.Address,
		IsPrimary:  w.IsPrimary,
		VerifiedAt: w.VerifiedAt,
		CreatedAt:  w.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *WalletRepository) CreateNonce(ctx context.Context, nonce *walletdomain.Nonce) error {
	if nonce.ID == "" {
		nonce.ID = "wno_" + uuid.NewString()
	}
	row := model.WalletNonce{
		ID:        nonce.ID,
		Address:   nonce.Address,
		Domain:    nonce.Domain,
		ChainID:   nonce.ChainID,
		Nonce:     nonce.Value,
		ExpiresAt: nonce.ExpiresAt,
		UsedAt:    nonce.UsedAt,
		CreatedAt: nonce.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *WalletRepository) FindNonce(ctx context.Context, address string, nonce string) (*walletdomain.Nonce, error) {
	var row model.WalletNonce
	if err := r.db.WithContext(ctx).Where("address = ? AND nonce = ?", address, nonce).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainNonce(row), nil
}

func (r *WalletRepository) MarkNonceUsed(ctx context.Context, nonceID string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&model.WalletNonce{}).
		Where("id = ? AND used_at IS NULL", nonceID).
		Update("used_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainrepo.ErrNotFound
	}
	return nil
}

func toDomainWallet(row model.UserWallet) *walletdomain.UserWallet {
	return &walletdomain.UserWallet{
		ID:         row.ID,
		UserID:     row.UserID,
		ChainType:  walletdomain.ChainType(row.ChainType),
		Address:    row.Address,
		IsPrimary:  row.IsPrimary,
		VerifiedAt: row.VerifiedAt,
		CreatedAt:  row.CreatedAt,
	}
}

func toDomainNonce(row model.WalletNonce) *walletdomain.Nonce {
	return &walletdomain.Nonce{
		ID:        row.ID,
		Address:   row.Address,
		Domain:    row.Domain,
		ChainID:   row.ChainID,
		Value:     row.Nonce,
		ExpiresAt: row.ExpiresAt,
		UsedAt:    row.UsedAt,
		CreatedAt: row.CreatedAt,
	}
}

var _ domainrepo.WalletRepository = (*WalletRepository)(nil)
