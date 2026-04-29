package wallet

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	walletdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/wallet"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrInvalidWalletAddress = "WALLET_INVALID_ADDRESS"
	ErrInvalidClient        = "CLIENT_INVALID"
	ErrInvalidNonce         = "WALLET_INVALID_NONCE"
	ErrInvalidSignature     = "WALLET_INVALID_SIGNATURE"
)

// Clock supplies time to keep wallet flows deterministic in tests.
type Clock interface {
	Now() time.Time
}

// AddressVerifier validates and normalizes wallet addresses.
type AddressVerifier interface {
	NormalizeAddress(address string) (string, error)
	VerifyMessage(address string, message string, signature string) (bool, error)
}

// TokenIssuer issues access and refresh tokens for wallet login.
type TokenIssuer interface {
	IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error)
	RefreshTokenTTL() time.Duration
}

// TokenHasher hashes opaque refresh tokens before persistence.
type TokenHasher interface {
	HashToken(raw string) string
}

// Service orchestrates wallet nonce and verification usecases.
type Service struct {
	wallets       repository.WalletRepository
	users         repository.UserRepository
	clients       repository.ClientRepository
	refreshTokens repository.RefreshTokenRepository
	activity      repository.ActivityRepository
	verifier      AddressVerifier
	tokenHasher   TokenHasher
	issuer        TokenIssuer
	ttl           time.Duration
	clock         Clock
}

// Dependencies contains external ports required by the wallet usecase.
type Dependencies struct {
	Wallets       repository.WalletRepository
	Users         repository.UserRepository
	Clients       repository.ClientRepository
	RefreshTokens repository.RefreshTokenRepository
	Activity      repository.ActivityRepository
	Verifier      AddressVerifier
	TokenHasher   TokenHasher
	Issuer        TokenIssuer
	NonceTTL      time.Duration
	Clock         Clock
}

// NonceRequest is the input for creating a wallet login nonce.
type NonceRequest struct {
	Address string
	Domain  string
	ChainID int64
}

// NonceResult contains a wallet login nonce and its expiration time.
type NonceResult struct {
	Nonce     string
	Message   string
	ExpiresAt time.Time
}

// VerifyRequest is the input for wallet signature login.
type VerifyRequest struct {
	ClientID  string
	Address   string
	Nonce     string
	Signature string
	IP        string
	UserAgent string
}

// VerifyResult is returned after a successful wallet signature login.
type VerifyResult struct {
	UserID   string
	Username string
	Email    string
	Wallets  []string
	Token    *token.Pair
}

// NewService creates the wallet usecase service.
func NewService(deps Dependencies) *Service {
	return &Service{
		wallets:       deps.Wallets,
		users:         deps.Users,
		clients:       deps.Clients,
		refreshTokens: deps.RefreshTokens,
		activity:      deps.Activity,
		verifier:      deps.Verifier,
		tokenHasher:   deps.TokenHasher,
		issuer:        deps.Issuer,
		ttl:           deps.NonceTTL,
		clock:         deps.Clock,
	}
}

// CreateNonce creates a short-lived challenge message for wallet login.
func (s *Service) CreateNonce(ctx context.Context, req NonceRequest) (*NonceResult, error) {
	address, err := s.normalizeAddress(req.Address)
	if err != nil {
		return nil, domain.NewError(ErrInvalidWalletAddress, "wallet address is required")
	}
	domainName := strings.TrimSpace(req.Domain)
	if domainName == "" {
		return nil, domain.NewError(ErrInvalidNonce, "domain is required")
	}
	if req.ChainID <= 0 {
		req.ChainID = 1
	}

	value, err := randomNonce()
	if err != nil {
		return nil, err
	}

	now := s.clock.Now().UTC()
	nonce := &walletdomain.Nonce{
		Address:   address,
		Domain:    domainName,
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
		Message:   BuildSIWEMessage(nonce),
		ExpiresAt: nonce.ExpiresAt,
	}, nil
}

// VerifySignature validates a wallet signature, creates the user if needed, and issues tokens.
func (s *Service) VerifySignature(ctx context.Context, req VerifyRequest) (*VerifyResult, error) {
	clientID := defaultClientID(req.ClientID)
	address, err := s.normalizeAddress(req.Address)
	if err != nil {
		return nil, domain.NewError(ErrInvalidWalletAddress, "invalid wallet address")
	}
	nonceValue := strings.TrimSpace(req.Nonce)
	signature := strings.TrimSpace(req.Signature)
	if nonceValue == "" || signature == "" {
		return nil, domain.NewError(ErrInvalidSignature, "nonce and signature are required")
	}

	client, err := s.clients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}

	nonce, err := s.wallets.FindNonce(ctx, address, nonceValue)
	if err != nil || nonce == nil || nonce.IsUsed() || nonce.IsExpired(s.clock.Now().UTC()) {
		return nil, domain.NewError(ErrInvalidNonce, "invalid or expired nonce")
	}

	message := BuildSIWEMessage(nonce)
	ok, err := s.verifier.VerifyMessage(address, message, signature)
	if err != nil || !ok {
		return nil, domain.NewError(ErrInvalidSignature, "invalid wallet signature")
	}

	wallet, err := s.wallets.FindByAddress(ctx, walletdomain.ChainTypeEVM, address)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	var u *user.User
	if wallet != nil {
		u, err = s.users.FindByID(ctx, wallet.UserID)
		if err != nil || u == nil || !u.IsActive() {
			return nil, domain.NewError(ErrInvalidSignature, "wallet user is unavailable")
		}
	} else {
		u, err = s.createWalletUser(ctx, address)
		if err != nil {
			return nil, err
		}
		wallet = &walletdomain.UserWallet{
			UserID:     u.ID,
			ChainType:  walletdomain.ChainTypeEVM,
			Address:    address,
			IsPrimary:  true,
			VerifiedAt: s.clock.Now().UTC(),
		}
		if err := s.wallets.CreateWallet(ctx, wallet); err != nil {
			return nil, err
		}
	}

	if err := s.wallets.MarkNonceUsed(ctx, nonce.ID); err != nil {
		return nil, err
	}

	pair, err := s.issuer.IssuePair(ctx, token.Claims{
		UserID:   u.ID,
		ClientID: client.ClientID,
		Audience: client.JWTAudience,
		Username: u.Username,
		Email:    u.Email,
		Wallets:  []string{address},
	})
	if err != nil {
		return nil, err
	}
	if err := s.storeRefreshToken(ctx, u.ID, client.ClientID, pair.RefreshToken); err != nil {
		return nil, err
	}
	if err := s.users.UpdateLoginInfo(ctx, u.ID); err != nil {
		return nil, err
	}
	if err := s.recordSuccessfulLogin(ctx, u.ID, client.ClientID, req.IP, req.UserAgent); err != nil {
		return nil, err
	}

	return &VerifyResult{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Wallets:  []string{address},
		Token:    pair,
	}, nil
}

// BuildSIWEMessage returns the exact message clients must ask wallets to sign.
func BuildSIWEMessage(nonce *walletdomain.Nonce) string {
	return fmt.Sprintf("%s wants you to sign in with your Ethereum account:\n%s\n\nSign in to Open Wallet Auth.\n\nURI: %s\nVersion: 1\nChain ID: %d\nNonce: %s\nIssued At: %s",
		nonce.Domain,
		nonce.Address,
		messageURI(nonce.Domain),
		nonce.ChainID,
		nonce.Value,
		nonce.CreatedAt.UTC().Format(time.RFC3339),
	)
}

func messageURI(domain string) string {
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return domain
	}
	return "https://" + domain
}

func (s *Service) createWalletUser(ctx context.Context, address string) (*user.User, error) {
	shortAddress := address
	if len(address) > 12 {
		shortAddress = address[:6] + address[len(address)-4:]
	}
	u := &user.User{
		Username: "wallet_" + strings.ToLower(shortAddress),
		Status:   user.StatusActive,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) storeRefreshToken(ctx context.Context, userID string, clientID string, raw string) error {
	return s.refreshTokens.Create(ctx, &token.RefreshToken{
		UserID:    userID,
		ClientID:  clientID,
		TokenHash: s.tokenHasher.HashToken(raw),
		ExpiresAt: s.clock.Now().UTC().Add(s.issuer.RefreshTokenTTL()),
	})
}

func (s *Service) recordSuccessfulLogin(ctx context.Context, userID string, clientID string, ip string, userAgent string) error {
	if s.activity == nil {
		return nil
	}
	if err := s.activity.RecordLogin(ctx, &audit.LoginLog{
		UserID:      userID,
		ClientID:    clientID,
		LoginMethod: audit.LoginMethodWallet,
		IP:          ip,
		UserAgent:   userAgent,
		Success:     true,
	}); err != nil {
		return err
	}
	return s.activity.UpsertUserClientLogin(ctx, userID, clientID)
}

func (s *Service) normalizeAddress(address string) (string, error) {
	if s.verifier == nil {
		return "", errors.New("wallet verifier is required")
	}
	return s.verifier.NormalizeAddress(strings.TrimSpace(address))
}

func randomNonce() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func defaultClientID(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "default"
	}
	return clientID
}
