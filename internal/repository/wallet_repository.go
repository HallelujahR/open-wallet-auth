package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
)

// WalletRepository defines persistence operations for wallets and login nonces.
type WalletRepository interface {
	FindByAddress(ctx context.Context, chainType wallet.ChainType, address string) (*wallet.UserWallet, error)
	ListByUserID(ctx context.Context, userID string) ([]wallet.UserWallet, error)
	CreateWallet(ctx context.Context, w *wallet.UserWallet) error
	DeleteByID(ctx context.Context, userID string, walletID string) error
	CreateNonce(ctx context.Context, nonce *wallet.Nonce) error
	FindNonce(ctx context.Context, address string, nonce string) (*wallet.Nonce, error)
	MarkNonceUsed(ctx context.Context, nonceID string) error
}
