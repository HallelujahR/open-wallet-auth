package model

import "time"

// UserWallet maps to the user_wallets table.
type UserWallet struct {
	ID         string    `gorm:"primaryKey;type:varchar(64)"`
	UserID     string    `gorm:"type:varchar(64);not null;index"`
	ChainType  string    `gorm:"type:varchar(32);not null;uniqueIndex:idx_wallet_chain_address"`
	Address    string    `gorm:"type:varchar(128);not null;uniqueIndex:idx_wallet_chain_address"`
	IsPrimary  bool      `gorm:"not null;default:false"`
	VerifiedAt time.Time `gorm:"type:timestamptz;not null"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null"`
}

func (UserWallet) TableName() string {
	return "user_wallets"
}

// WalletNonce maps to the wallet_nonces table.
type WalletNonce struct {
	ID        string     `gorm:"primaryKey;type:varchar(64)"`
	Address   string     `gorm:"type:varchar(128);not null;index"`
	Domain    string     `gorm:"type:varchar(255);not null"`
	ChainID   int64      `gorm:"not null"`
	Nonce     string     `gorm:"type:varchar(128);not null;uniqueIndex"`
	ExpiresAt time.Time  `gorm:"type:timestamptz;not null"`
	UsedAt    *time.Time `gorm:"type:timestamptz"`
	CreatedAt time.Time  `gorm:"type:timestamptz;not null"`
}

func (WalletNonce) TableName() string {
	return "wallet_nonces"
}
