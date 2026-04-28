package wallet

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	walletdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrInvalidWalletAddress = "WALLET_INVALID_ADDRESS"
)

type Clock interface {
	Now() time.Time
}

type Service struct {
	wallets repository.WalletRepository
	ttl     time.Duration
	clock   Clock
}

type NonceRequest struct {
	Address string
	Domain  string
	ChainID int64
}

type NonceResult struct {
	Nonce     string
	ExpiresAt time.Time
}

func NewService(wallets repository.WalletRepository, ttl time.Duration, clock Clock) *Service {
	return &Service{
		wallets: wallets,
		ttl:     ttl,
		clock:   clock,
	}
}

func (s *Service) CreateNonce(ctx context.Context, req NonceRequest) (*NonceResult, error) {
	address := strings.TrimSpace(req.Address)
	if address == "" {
		return nil, domain.NewError(ErrInvalidWalletAddress, "wallet address is required")
	}

	value, err := randomNonce()
	if err != nil {
		return nil, err
	}

	now := s.clock.Now().UTC()
	nonce := &walletdomain.Nonce{
		Address:   strings.ToLower(address),
		Domain:    req.Domain,
		ChainID:   req.ChainID,
		Value:     value,
		ExpiresAt: now.Add(s.ttl),
		CreatedAt: now,
	}
	if err := s.wallets.CreateNonce(ctx, nonce); err != nil {
		return nil, err
	}

	return &NonceResult{
		Nonce:     value,
		ExpiresAt: nonce.ExpiresAt,
	}, nil
}

func randomNonce() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
