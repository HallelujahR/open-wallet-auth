package oauth

import (
	"context"
	"strings"

	oauthusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/oauth"
	settingsusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/settings"
)

// SettingsProvider supplies the latest runtime provider settings.
// SettingsProvider 提供最新的运行期 OAuth 配置。
type SettingsProvider interface {
	Get(ctx context.Context) (settingsusecase.Snapshot, error)
}

// DynamicProvider resolves OAuth credentials from editable settings on each operation.
// DynamicProvider 每次 OAuth 操作前读取最新可视化配置。
type DynamicProvider struct {
	name     string
	base     ProviderConfig
	settings SettingsProvider
}

// NewDynamicProvider creates a runtime-configurable OAuth provider.
// NewDynamicProvider 创建可运行期配置的 OAuth provider。
func NewDynamicProvider(name string, base ProviderConfig, settings SettingsProvider) *DynamicProvider {
	base.Name = name
	return &DynamicProvider{name: strings.ToLower(name), base: base, settings: settings}
}

// Name returns the provider name.
// Name 返回服务商名称。
func (p *DynamicProvider) Name() string {
	return p.name
}

// Configured reports whether the current provider settings are complete.
// Configured 判断当前服务商配置是否完整。
func (p *DynamicProvider) Configured() bool {
	return p.provider(context.Background()).Configured()
}

// ConfiguredForRedirect checks tenant/default credentials for redirect_uri.
// ConfiguredForRedirect 检查指定回调地址是否具备可用凭据。
func (p *DynamicProvider) ConfiguredForRedirect(redirectURI string) bool {
	return p.provider(context.Background()).ConfiguredForRedirect(redirectURI)
}

// AuthURL builds the authorization URL from latest settings.
// AuthURL 使用最新配置构造授权跳转地址。
func (p *DynamicProvider) AuthURL(state string, redirectURI string, credentialURI string) string {
	return p.provider(context.Background()).AuthURL(state, redirectURI, credentialURI)
}

// FetchUser exchanges code and loads userinfo from latest settings.
// FetchUser 使用最新配置交换授权码并拉取用户信息。
func (p *DynamicProvider) FetchUser(ctx context.Context, code string, redirectURI string, credentialURI string) (*oauthusecase.ProviderUser, error) {
	return p.provider(ctx).FetchUser(ctx, code, redirectURI, credentialURI)
}

// provider creates an HTTP provider from the latest settings snapshot.
// provider 根据最新配置快照创建 HTTP provider。
func (p *DynamicProvider) provider(ctx context.Context) *HTTPProvider {
	cfg := p.base
	if p.settings != nil {
		if snapshot, err := p.settings.Get(ctx); err == nil {
			switch p.name {
			case "google":
				cfg = oauthConfig("google", snapshot.OAuth.Google)
			case "github":
				cfg = oauthConfig("github", snapshot.OAuth.GitHub)
			}
		}
	}
	return NewHTTPProvider(cfg)
}

// oauthConfig converts management settings into provider config.
// oauthConfig 将管理配置转换为 OAuth provider 配置。
func oauthConfig(name string, input settingsusecase.OAuthProviderSettings) ProviderConfig {
	tenants := make(map[string]ProviderTenantConfig, len(input.TenantCredentials))
	for _, tenant := range input.TenantCredentials {
		host := strings.ToLower(strings.TrimSpace(tenant.Host))
		if host == "" {
			continue
		}
		tenants[host] = ProviderTenantConfig{ClientID: tenant.ClientID, ClientSecret: tenant.ClientSecret}
	}
	return ProviderConfig{
		Name:         name,
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		AuthURL:      input.AuthURL,
		TokenURL:     input.TokenURL,
		UserInfoURL:  input.UserInfoURL,
		Scopes:       input.Scopes,
		Tenants:      tenants,
	}
}
