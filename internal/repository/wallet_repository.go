package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
)

type WalletRepository interface {
	FindByAddress(ctx context.Context, chainType wallet.ChainType, address string) (*wallet.UserWallet, error)
	CreateWallet(ctx context.Context, w *wallet.UserWallet) error
	CreateNonce(ctx context.Context, nonce *wallet.Nonce) error
	FindNonce(ctx context.Context, address string, nonce string) (*wallet.Nonce, error)
	MarkNonceUsed(ctx context.Context, nonceID string) error
}
