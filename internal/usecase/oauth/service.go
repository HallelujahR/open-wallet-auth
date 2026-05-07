package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
	oauthdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/oauth"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/user"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const (
	ErrInvalidClient   = "CLIENT_INVALID"
	ErrInvalidProvider = "OAUTH_INVALID_PROVIDER"
	ErrInvalidState    = "OAUTH_INVALID_STATE"
	ErrOAuthBound      = "OAUTH_ALREADY_BOUND"
	ErrProviderFailed  = "OAUTH_PROVIDER_FAILED"
)

// Clock supplies time to keep OAuth state flows deterministic in tests.
// Clock 抽象时间来源，便于测试 OAuth state 过期逻辑。
type Clock interface {
	Now() time.Time
}

// ProviderUser is the normalized profile returned by an OAuth provider.
// ProviderUser 是第三方 OAuth 用户资料的统一表示，避免 usecase 依赖 Google/GitHub 原始响应。
type ProviderUser struct {
	Subject       string
	Email         string
	EmailVerified bool
	Username      string
	AvatarURL     string
}

// Provider hides third-party OAuth implementation details from usecases.
// Provider 是 OAuth 服务商端口，具体 HTTP token/userinfo 交换实现放在 infrastructure 层。
type Provider interface {
	Name() string
	AuthURL(state string, redirectURI string) string
	FetchUser(ctx context.Context, code string, redirectURI string) (*ProviderUser, error)
	Configured() bool
	ConfiguredForRedirect(redirectURI string) bool
}

// StateStore persists short-lived OAuth state between start and callback.
// StateStore 保存 OAuth start/callback 之间的短期 state，防止回调伪造和跨 client 混用。
type StateStore interface {
	Save(ctx context.Context, state string, value StateValue, expiresAt time.Time) error
	Take(ctx context.Context, state string, now time.Time) (*StateValue, error)
}

// TokenIssuer issues access and refresh tokens for OAuth login.
// TokenIssuer 是 OAuth 登录的 token 签发端口。
type TokenIssuer interface {
	IssuePair(ctx context.Context, claims token.Claims) (*token.Pair, error)
	RefreshTokenTTL() time.Duration
}

// TokenHasher hashes opaque refresh tokens before persistence.
// TokenHasher 在刷新令牌入库前做单向哈希。
type TokenHasher interface {
	HashToken(raw string) string
}

// StateValue records trusted values from the OAuth start request.
// StateValue 保存 OAuth 发起阶段已校验的可信参数。
type StateValue struct {
	ClientID    string
	RedirectURI string
	BindUserID  string
}

// Service orchestrates OAuth start and callback login.
// Service 编排 OAuth 授权发起、回调、账号归并和 token 签发流程。
type Service struct {
	users         repository.UserRepository
	clients       repository.ClientRepository
	refreshTokens repository.RefreshTokenRepository
	activity      repository.ActivityRepository
	accounts      repository.OAuthAccountRepository
	states        StateStore
	providers     map[string]Provider
	tokenHasher   TokenHasher
	issuer        TokenIssuer
	stateTTL      time.Duration
	clock         Clock
}

// Dependencies contains external ports required by OAuth login.
// Dependencies 汇总 OAuth 用例需要的仓储、state、provider 和 token 端口。
type Dependencies struct {
	Users         repository.UserRepository
	Clients       repository.ClientRepository
	RefreshTokens repository.RefreshTokenRepository
	Activity      repository.ActivityRepository
	Accounts      repository.OAuthAccountRepository
	States        StateStore
	Providers     []Provider
	TokenHasher   TokenHasher
	Issuer        TokenIssuer
	StateTTL      time.Duration
	Clock         Clock
}

// StartRequest is the input for creating an OAuth authorization URL.
// StartRequest 是创建第三方授权地址的用例输入。
type StartRequest struct {
	Provider    string
	ClientID    string
	RedirectURI string
	BindUserID  string
}

// StartResult contains the provider redirect URL.
// StartResult 返回前端需要跳转的第三方授权地址。
type StartResult struct {
	Provider string
	AuthURL  string
	State    string
}

// CallbackRequest is the input for completing OAuth login.
// CallbackRequest 是完成 OAuth 回调登录的用例输入。
type CallbackRequest struct {
	Provider  string
	Code      string
	State     string
	IP        string
	UserAgent string
}

// CallbackResult is returned after a successful OAuth callback.
// CallbackResult 是 OAuth 回调登录成功后的用例输出。
type CallbackResult struct {
	UserID   string
	Username string
	Email    string
	Token    *token.Pair
}

// NewService creates the OAuth usecase service.
// NewService 创建 OAuth 用例服务，并按 provider name 建立查找表。
func NewService(deps Dependencies) *Service {
	providers := make(map[string]Provider, len(deps.Providers))
	for _, provider := range deps.Providers {
		if provider != nil {
			providers[provider.Name()] = provider
		}
	}
	return &Service{
		users:         deps.Users,
		clients:       deps.Clients,
		refreshTokens: deps.RefreshTokens,
		activity:      deps.Activity,
		accounts:      deps.Accounts,
		states:        deps.States,
		providers:     providers,
		tokenHasher:   deps.TokenHasher,
		issuer:        deps.Issuer,
		stateTTL:      deps.StateTTL,
		clock:         deps.Clock,
	}
}

// Start validates the client and returns a provider authorization URL.
// Start 只生成授权地址，不直接接触第三方 HTTP 细节。
func (s *Service) Start(ctx context.Context, req StartRequest) (*StartResult, error) {
	provider, err := s.provider(req.Provider)
	if err != nil {
		return nil, err
	}
	if !provider.Configured() {
		return nil, domain.NewError(ErrProviderFailed, "oauth provider is not configured")
	}
	clientID := defaultClientID(req.ClientID)
	client, err := s.clients.FindByClientID(ctx, clientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}
	redirectURI := strings.TrimSpace(req.RedirectURI)
	if redirectURI == "" {
		return nil, domain.NewError(ErrInvalidState, "redirect_uri is required")
	}
	if !provider.ConfiguredForRedirect(redirectURI) {
		return nil, domain.NewError(ErrProviderFailed, "oauth provider is not configured for redirect_uri")
	}
	state, err := randomState()
	if err != nil {
		return nil, err
	}
	if err := s.states.Save(ctx, state, StateValue{ClientID: client.ClientID, RedirectURI: redirectURI, BindUserID: strings.TrimSpace(req.BindUserID)}, s.clock.Now().UTC().Add(s.stateTTL)); err != nil {
		return nil, err
	}
	return &StartResult{Provider: provider.Name(), AuthURL: provider.AuthURL(state, redirectURI), State: state}, nil
}

// Callback exchanges an OAuth code for a provider user, links the account, and issues tokens.
// Callback 完成第三方身份到本地用户的映射，并统一签发本服务 JWT。
func (s *Service) Callback(ctx context.Context, req CallbackRequest) (*CallbackResult, error) {
	provider, err := s.provider(req.Provider)
	if err != nil {
		return nil, err
	}
	state, err := s.states.Take(ctx, strings.TrimSpace(req.State), s.clock.Now().UTC())
	if err != nil || state == nil {
		return nil, domain.NewError(ErrInvalidState, "invalid oauth state")
	}
	client, err := s.clients.FindByClientID(ctx, state.ClientID)
	if err != nil || client == nil || !client.IsActive() {
		return nil, domain.NewError(ErrInvalidClient, "invalid client")
	}
	profile, err := provider.FetchUser(ctx, strings.TrimSpace(req.Code), state.RedirectURI)
	if err != nil || profile == nil || profile.Subject == "" {
		return nil, domain.WrapError(ErrProviderFailed, "oauth provider request failed", err)
	}

	account, err := s.accounts.FindByProviderSubject(ctx, provider.Name(), profile.Subject)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	u, err := s.resolveUser(ctx, provider.Name(), profile, account, state.BindUserID)
	if err != nil {
		return nil, err
	}

	pair, err := s.issuer.IssuePair(ctx, token.Claims{
		UserID:   u.ID,
		ClientID: client.ClientID,
		Audience: client.JWTAudience,
		Username: u.Username,
		Email:    u.Email,
	})
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokens.Create(ctx, &token.RefreshToken{
		UserID:    u.ID,
		ClientID:  client.ClientID,
		TokenHash: s.tokenHasher.HashToken(pair.RefreshToken),
		IP:        req.IP,
		UserAgent: req.UserAgent,
		ExpiresAt: s.clock.Now().UTC().Add(s.issuer.RefreshTokenTTL()),
	}); err != nil {
		return nil, err
	}
	if err := s.users.UpdateLoginInfo(ctx, u.ID); err != nil {
		return nil, err
	}
	if s.activity != nil {
		if err := s.activity.RecordLogin(ctx, &audit.LoginLog{
			UserID:      u.ID,
			ClientID:    client.ClientID,
			LoginMethod: audit.LoginMethodOAuth,
			IP:          req.IP,
			UserAgent:   req.UserAgent,
			Success:     true,
		}); err != nil {
			return nil, err
		}
		if err := s.activity.UpsertUserClientLogin(ctx, u.ID, client.ClientID); err != nil {
			return nil, err
		}
	}
	return &CallbackResult{UserID: u.ID, Username: u.Username, Email: u.Email, Token: pair}, nil
}

// resolveUser links an OAuth profile according to explicit binding rules.
// resolveUser 按显式绑定规则处理第三方账号和本地用户关系。
func (s *Service) resolveUser(ctx context.Context, provider string, profile *ProviderUser, account *oauthdomain.Account, bindUserID string) (*user.User, error) {
	if account != nil {
		if bindUserID != "" && account.UserID != bindUserID {
			return nil, domain.NewError(ErrOAuthBound, "oauth account is already bound to another account")
		}
		return s.users.FindByID(ctx, account.UserID)
	}

	u, err := s.resolveOAuthTargetUser(ctx, provider, profile, bindUserID)
	if err != nil {
		return nil, err
	}
	if !u.IsActive() {
		return nil, domain.NewError(ErrProviderFailed, "oauth user is unavailable")
	}
	if err := s.accounts.Create(ctx, &oauthdomain.Account{
		UserID:            u.ID,
		Provider:          provider,
		ProviderSubject:   profile.Subject,
		ProviderEmail:     profile.Email,
		ProviderUsername:  profile.Username,
		ProviderAvatarURL: profile.AvatarURL,
	}); err != nil {
		return nil, err
	}
	return u, nil
}

// resolveOAuthTargetUser returns the user that should receive a new OAuth binding.
// resolveOAuthTargetUser 返回新 OAuth 绑定应该归属的用户。
func (s *Service) resolveOAuthTargetUser(ctx context.Context, provider string, profile *ProviderUser, bindUserID string) (*user.User, error) {
	if bindUserID != "" {
		u, err := s.users.FindByID(ctx, bindUserID)
		if err != nil || u == nil || !u.IsActive() {
			return nil, domain.NewError(ErrProviderFailed, "oauth binding user is unavailable")
		}
		return u, nil
	}
	if profile.Email != "" {
		existing, err := s.users.FindByEmail(ctx, profile.Email)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		if existing != nil {
			if !profile.EmailVerified {
				return nil, domain.NewError(ErrOAuthBound, "oauth email is already used by another account; login first and bind explicitly")
			}
			if !existing.IsActive() {
				return nil, domain.NewError(ErrProviderFailed, "oauth user is unavailable")
			}
			return existing, nil
		}
	}
	username := strings.TrimSpace(profile.Username)
	if username == "" {
		username = provider + "_" + profile.Subject
	}
	u := &user.User{Username: username, Email: profile.Email, Avatar: profile.AvatarURL, Status: user.StatusActive}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// provider returns a configured provider adapter by normalized provider name.
// provider 根据归一化后的服务商名称返回对应适配器。
func (s *Service) provider(name string) (Provider, error) {
	provider, ok := s.providers[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return nil, domain.NewError(ErrInvalidProvider, "invalid oauth provider")
	}
	return provider, nil
}

// randomState creates a cryptographically random OAuth state value.
// randomState 创建密码学安全的 OAuth state，用于防止回调伪造。
func randomState() (string, error) {
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
