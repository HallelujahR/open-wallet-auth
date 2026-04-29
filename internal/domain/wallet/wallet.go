package wallet

import "time"

// ChainType identifies a wallet ecosystem such as EVM.
// ChainType 标识钱包所属生态，例如 EVM。
type ChainType string

const (
	ChainTypeEVM ChainType = "evm"
)

// UserWallet links a verified wallet address to a user.
// UserWallet 记录已验证的钱包地址与用户账号之间的绑定关系。
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
// Nonce 是钱包签名登录的一次性挑战值。
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

// IsExpired reports whether the nonce can no longer be used.
// IsExpired 判断 nonce 是否已经超过可用时间。
func (n Nonce) IsExpired(now time.Time) bool {
	return !n.ExpiresAt.After(now)
}

// IsUsed reports whether the nonce has already been consumed.
// IsUsed 判断 nonce 是否已经被消费，防止签名重放。
func (n Nonce) IsUsed() bool {
	return n.UsedAt != nil
}
