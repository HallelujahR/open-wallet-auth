package wallet

import "time"

// ChainType identifies a wallet ecosystem such as EVM.
type ChainType string

const (
	ChainTypeEVM ChainType = "evm"
)

// UserWallet links a verified wallet address to a user.
type UserWallet struct {
	ID         string
	UserID     string
	ChainType  ChainType
	Address    string
	IsPrimary  bool
	VerifiedAt time.Time
	CreatedAt  time.Time
}

// Nonce is a one-time challenge used by wallet signature login.
type Nonce struct {
	ID        string
	Address   string
	Domain    string
	ChainID   int64
	Value     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (n Nonce) IsExpired(now time.Time) bool {
	return !n.ExpiresAt.After(now)
}

func (n Nonce) IsUsed() bool {
	return n.UsedAt != nil
}
