package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	oauthusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/oauth"
)

// ProviderConfig contains OAuth provider endpoints and credentials.
// ProviderConfig 保存 OAuth 服务商端点和凭证配置。
type ProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
	Tenants      map[string]ProviderTenantConfig
}

// ProviderTenantConfig overrides OAuth credentials for one redirect host.
// ProviderTenantConfig 按 redirect_uri 域名覆盖 OAuth 凭据，保证多业务系统共用认证中台时仍可使用独立 OAuth App。
type ProviderTenantConfig struct {
	ClientID     string
	ClientSecret string
}

// HTTPProvider implements OAuth code exchange and userinfo loading with net/http.
// HTTPProvider 使用标准 HTTP 实现 OAuth 授权码交换和用户信息获取。
type HTTPProvider struct {
	cfg        ProviderConfig
	httpClient *http.Client
}

// NewHTTPProvider creates a generic HTTP OAuth provider adapter.
// NewHTTPProvider 创建通用 OAuth HTTP 适配器。
func NewHTTPProvider(cfg ProviderConfig) *HTTPProvider {
	return &HTTPProvider{cfg: cfg, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

// Name returns the normalized provider name used by the usecase registry.
// Name 返回用例层 provider 注册表使用的归一化服务商名称。
func (p *HTTPProvider) Name() string {
	return strings.ToLower(p.cfg.Name)
}

// Configured reports whether all required provider settings are present.
// Configured 判断该 OAuth 服务商是否已经具备完整配置。
func (p *HTTPProvider) Configured() bool {
	if p.configComplete(p.cfg.ClientID, p.cfg.ClientSecret) {
		return true
	}
	for _, tenant := range p.cfg.Tenants {
		if p.configComplete(tenant.ClientID, tenant.ClientSecret) {
			return true
		}
	}
	return false
}

// ConfiguredForRedirect reports whether redirect_uri can resolve a complete credential pair.
// ConfiguredForRedirect 判断指定回调地址是否能匹配到完整 OAuth 凭据。
func (p *HTTPProvider) ConfiguredForRedirect(redirectURI string) bool {
	cfg := p.configForRedirect(redirectURI)
	return p.configComplete(cfg.ClientID, cfg.ClientSecret)
}

// AuthURL builds the provider authorization URL for browser redirection.
// AuthURL 构造浏览器需要跳转的第三方授权地址，凭据可按业务 return_uri 域名选择。
func (p *HTTPProvider) AuthURL(state string, redirectURI string, credentialURI string) string {
	cfg := p.configForRedirect(credentialURI)
	values := url.Values{}
	values.Set("client_id", cfg.ClientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("response_type", "code")
	values.Set("state", state)
	if len(p.cfg.Scopes) > 0 {
		values.Set("scope", strings.Join(p.cfg.Scopes, " "))
	}
	return p.cfg.AuthURL + "?" + values.Encode()
}

// FetchUser exchanges the code and returns a normalized provider profile.
// FetchUser 用授权码换取访问令牌，并按业务域名选择对应 OAuth 凭据。
func (p *HTTPProvider) FetchUser(ctx context.Context, code string, redirectURI string, credentialURI string) (*oauthusecase.ProviderUser, error) {
	cfg := p.configForRedirect(credentialURI)
	if !p.configComplete(cfg.ClientID, cfg.ClientSecret) {
		return nil, errors.New("oauth provider is not configured")
	}
	accessToken, err := p.exchangeToken(ctx, cfg, code, redirectURI)
	if err != nil {
		return nil, err
	}
	return p.fetchUser(ctx, accessToken)
}

// exchangeToken calls the provider token endpoint with the authorization code.
// exchangeToken 调用服务商 token endpoint，用授权码换取 access_token。
func (p *HTTPProvider) exchangeToken(ctx context.Context, cfg ProviderConfig, code string, redirectURI string) (string, error) {
	form := url.Values{}
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("oauth token endpoint returned non-2xx status")
	}
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", errors.New("oauth token response missing access_token")
	}
	return payload.AccessToken, nil
}

// configComplete checks shared provider endpoints plus one credential pair.
// configComplete 校验公共 OAuth 端点和某一组 client 凭据是否完整。
func (p *HTTPProvider) configComplete(clientID string, clientSecret string) bool {
	return clientID != "" && clientSecret != "" && p.cfg.AuthURL != "" && p.cfg.TokenURL != "" && p.cfg.UserInfoURL != ""
}

// configForRedirect returns tenant credentials matched by redirect_uri host.
// configForRedirect 根据 redirect_uri 的 host 选择租户凭据，未命中时回退到默认凭据。
func (p *HTTPProvider) configForRedirect(redirectURI string) ProviderConfig {
	host := redirectHost(redirectURI)
	if host != "" {
		if tenant, ok := p.cfg.Tenants[host]; ok && tenant.ClientID != "" && tenant.ClientSecret != "" {
			cfg := p.cfg
			cfg.ClientID = tenant.ClientID
			cfg.ClientSecret = tenant.ClientSecret
			return cfg
		}
	}
	return p.cfg
}

// redirectHost extracts a normalized host from redirect_uri.
// redirectHost 从 redirect_uri 中提取归一化域名。
func redirectHost(redirectURI string) string {
	parsed, err := url.Parse(strings.TrimSpace(redirectURI))
	if err != nil {
		return ""
	}
	return strings.ToLower(parsed.Hostname())
}

// fetchUser loads raw userinfo and converts it into the usecase provider model.
// fetchUser 拉取原始用户信息，并转换为用例层统一用户模型。
func (p *HTTPProvider) fetchUser(ctx context.Context, accessToken string) (*oauthusecase.ProviderUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("oauth userinfo endpoint returned non-2xx status")
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	profile := normalizeUser(p.Name(), raw)
	if p.Name() == "github" && profile.Email == "" {
		if email, err := p.fetchGitHubPrimaryEmail(ctx, accessToken); err == nil {
			profile.Email = email
			profile.EmailVerified = email != ""
		}
	}
	return profile, nil
}

// normalizeUser maps provider-specific JSON into a normalized ProviderUser.
// normalizeUser 将不同服务商的 JSON 字段映射为统一 ProviderUser。
func normalizeUser(provider string, raw map[string]any) *oauthusecase.ProviderUser {
	switch provider {
	case "github":
		return &oauthusecase.ProviderUser{
			Subject:       stringValue(raw["id"]),
			Email:         stringValue(raw["email"]),
			EmailVerified: stringValue(raw["email"]) != "",
			Username:      stringValue(raw["login"]),
			AvatarURL:     stringValue(raw["avatar_url"]),
		}
	default:
		return &oauthusecase.ProviderUser{
			Subject:       firstString(raw["sub"], raw["id"]),
			Email:         stringValue(raw["email"]),
			EmailVerified: boolValue(raw["email_verified"]),
			Username:      firstString(raw["name"], raw["login"], raw["email"]),
			AvatarURL:     firstString(raw["picture"], raw["avatar_url"]),
		}
	}
}

// fetchGitHubPrimaryEmail loads the verified primary email from GitHub's email API.
// fetchGitHubPrimaryEmail 从 GitHub 邮箱列表接口读取已验证的主邮箱。
func (p *HTTPProvider) fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("github email endpoint returned non-2xx status")
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, email := range emails {
		if email.Primary && email.Verified && strings.TrimSpace(email.Email) != "" {
			return strings.TrimSpace(email.Email), nil
		}
	}
	return "", nil
}

// firstString returns the first non-empty string-like value.
// firstString 返回第一个非空的字符串型值。
func firstString(values ...any) string {
	for _, value := range values {
		if s := stringValue(value); s != "" {
			return s
		}
	}
	return ""
}

// stringValue converts simple JSON values to stable string identifiers.
// stringValue 将简单 JSON 值转换为稳定字符串标识。
func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strings.TrimRight(strings.TrimRight(jsonNumber(v), "0"), ".")
	default:
		return ""
	}
}

// boolValue converts simple JSON boolean values.
// boolValue 转换 JSON 布尔值。
func boolValue(value any) bool {
	v, ok := value.(bool)
	return ok && v
}

// jsonNumber renders a JSON number without scientific notation surprises.
// jsonNumber 将 JSON 数字渲染为字符串，避免科学计数法带来的标识变化。
func jsonNumber(value float64) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}
