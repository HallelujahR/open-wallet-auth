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
// Clock 抽象时间来源，便于测试钱包 nonce 过期逻辑。
type Clock interface {
	Now() time.Time
}

// AddressVerifier validates and normalizes wallet addresses.
// AddressVerifier 是钱包地址校验端口，用例层不绑定具体链 SDK。
type AddressVerifier interface {
	NormalizeAddress(address string) (string, error)
	VerifyMessage(address string, message string, signature string) (bool, error)
}

// TokenIssuer issues access and refresh tokens for wallet login.
// TokenIssuer 是钱包登录的 token 签发端口。
type TokenIssuer interface {
	IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error)
	RefreshTokenTTL() time.Duration
}

// TokenHasher hashes opaque refresh tokens before persistence.
// TokenHasher 在刷新令牌入库前做单向哈希。
type TokenHasher interface {
	HashToken(raw string) string
}

// Service orchestrates wallet nonce and verification usecases.
// Service 编排钱包 nonce、签名校验、用户绑定与 token 签发流程。
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
// Dependencies 汇总钱包用例需要的仓储、签名校验和 token 端口。
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
// NonceRequest 是创建钱包登录挑战值的用例输入。
type NonceRequest struct {
	Address string
	Domain  string
	ChainID int64
}

// NonceResult contains a wallet login nonce and its expiration time.
// NonceResult 返回前端需要签名的消息、nonce 和过期时间。
type NonceResult struct {
	Nonce     string
	Message   string
	ExpiresAt time.Time
}

// VerifyRequest is the input for wallet signature login.
// VerifyRequest 是钱包签名登录的用例输入。
type VerifyRequest struct {
	ClientID  string
	Address   string
	Nonce     string
	Signature string
	IP        string
	UserAgent string
}

// VerifyResult is returned after a successful wallet signature login.
// VerifyResult 是钱包签名登录成功后的用例输出。
type VerifyResult struct {
	UserID   string
	Username string
	Email    string
	Wallets  []string
	Token    *token.Pair
}

// NewService creates the wallet usecase service.
// NewService 创建钱包登录用例服务，并注入所有外部端口。
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
// CreateNonce 创建短期有效的钱包签名挑战消息。
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
// VerifySignature 校验钱包签名，必要时创建用户并签发 token。
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
// BuildSIWEMessage 构造前端钱包必须签名的 SIWE 兼容消息。
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

// messageURI normalizes a domain into a URI value for the SIWE message.
// messageURI 将域名归一化为 SIWE 消息中的 URI 字段。
func messageURI(domain string) string {
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return domain
	}
	return "https://" + domain
}

// createWalletUser creates a local user profile for a first-time wallet address.
// createWalletUser 为首次使用的钱包地址创建本地用户资料。
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

// storeRefreshToken hashes and persists a wallet-login refresh token.
// storeRefreshToken 将钱包登录刷新令牌哈希后写入仓储。
func (s *Service) storeRefreshToken(ctx context.Context, userID string, clientID string, raw string) error {
	return s.refreshTokens.Create(ctx, &token.RefreshToken{
		UserID:    userID,
		ClientID:  clientID,
		TokenHash: s.tokenHasher.HashToken(raw),
		ExpiresAt: s.clock.Now().UTC().Add(s.issuer.RefreshTokenTTL()),
	})
}

// recordSuccessfulLogin writes wallet login audit data.
// recordSuccessfulLogin 记录钱包登录审计日志和用户-client 关系。
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

// normalizeAddress delegates address validation to the injected verifier port.
// normalizeAddress 通过注入的钱包校验端口完成地址格式化和校验。
func (s *Service) normalizeAddress(address string) (string, error) {
	if s.verifier == nil {
		return "", errors.New("wallet verifier is required")
	}
	return s.verifier.NormalizeAddress(strings.TrimSpace(address))
}

// randomNonce creates a cryptographically random URL-safe nonce.
// randomNonce 创建密码学安全、URL 友好的随机 nonce。
func randomNonce() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// defaultClientID normalizes an empty client id to the built-in default client.
// defaultClientID 将空 client_id 归一化为内置 default 业务系统。
func defaultClientID(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "default"
	}
	return clientID
}
