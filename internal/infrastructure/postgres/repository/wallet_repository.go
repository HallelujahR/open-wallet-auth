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
// WalletRepository 是钱包绑定和一次性 nonce 仓储端口的 PostgreSQL 适配器。
type WalletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a PostgreSQL wallet repository.
// NewWalletRepository 创建 PostgreSQL 钱包仓储。
func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// FindByAddress loads an existing wallet binding by chain and address.
// FindByAddress 按链类型和地址查询已有钱包绑定。
func (r *WalletRepository) FindByAddress(ctx context.Context, chainType walletdomain.ChainType, address string) (*walletdomain.UserWallet, error) {
	var row model.UserWallet
	if err := r.db.WithContext(ctx).Where("chain_type = ? AND address = ?", string(chainType), address).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainWallet(row), nil
}

// CreateWallet persists a verified wallet binding.
// CreateWallet 持久化已验证的钱包绑定关系。
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

// CreateNonce persists a one-time wallet login challenge.
// CreateNonce 持久化一次性钱包登录挑战值。
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

// FindNonce loads a wallet challenge by address and nonce value.
// FindNonce 按地址和 nonce 值查询钱包登录挑战。
func (r *WalletRepository) FindNonce(ctx context.Context, address string, nonce string) (*walletdomain.Nonce, error) {
	var row model.WalletNonce
	if err := r.db.WithContext(ctx).Where("address = ? AND nonce = ?", address, nonce).First(&row).Error; err != nil {
		return nil, mapGormError(err)
	}
	return toDomainNonce(row), nil
}

// MarkNonceUsed consumes a nonce exactly once.
// MarkNonceUsed 原子地消费 nonce，防止重复签名登录。
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

// ListByUserID returns all wallet bindings for one identity user.
// ListByUserID 返回某个身份用户的全部钱包绑定。
func (r *WalletRepository) ListByUserID(ctx context.Context, userID string) ([]walletdomain.UserWallet, error) {
	var rows []model.UserWallet
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	wallets := make([]walletdomain.UserWallet, 0, len(rows))
	for _, row := range rows {
		wallets = append(wallets, *toDomainWallet(row))
	}
	return wallets, nil
}

// toDomainWallet converts a wallet row into the domain wallet entity.
// toDomainWallet 将钱包数据库行转换为领域钱包实体。
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

// toDomainNonce converts a nonce row into the domain nonce entity.
// toDomainNonce 将 nonce 数据库行转换为领域 nonce 实体。
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
var _ domainrepo.AdminWalletRepository = (*WalletRepository)(nil)
