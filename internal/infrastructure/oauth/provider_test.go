package oauth

import (
	"net/url"
	"testing"
)

// TestHTTPProviderAuthURLUsesTenantCredentials verifies host-based OAuth credential routing.
// TestHTTPProviderAuthURLUsesTenantCredentials 验证 OAuth 授权地址会按回调域名选择业务专属凭据。
func TestHTTPProviderAuthURLUsesTenantCredentials(t *testing.T) {
	provider := NewHTTPProvider(ProviderConfig{
		Name:         "github",
		ClientID:     "default-client",
		ClientSecret: "default-secret",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Tenants: map[string]ProviderTenantConfig{
			"label.example.com": {
				ClientID:     "label-client",
				ClientSecret: "label-secret",
			},
		},
	})

	authURL := provider.AuthURL("state-value", "https://auth.example.com/api/v1/oauth/github/callback", "https://label.example.com/auth/oauth/callback?provider=github")
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse auth url: %v", err)
	}
	if got := parsed.Query().Get("client_id"); got != "label-client" {
		t.Fatalf("expected tenant client_id, got %q", got)
	}
}

// TestHTTPProviderAuthURLFallsBackToDefaultCredentials verifies default OAuth credentials still work.
// TestHTTPProviderAuthURLFallsBackToDefaultCredentials 验证未命中业务域名时仍回退到默认 OAuth 凭据。
func TestHTTPProviderAuthURLFallsBackToDefaultCredentials(t *testing.T) {
	provider := NewHTTPProvider(ProviderConfig{
		Name:         "github",
		ClientID:     "default-client",
		ClientSecret: "default-secret",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Tenants: map[string]ProviderTenantConfig{
			"label.example.com": {
				ClientID:     "label-client",
				ClientSecret: "label-secret",
			},
		},
	})

	authURL := provider.AuthURL("state-value", "https://auth.example.com/api/v1/oauth/github/callback", "https://blockx.example.com/auth/oauth/callback?provider=github")
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse auth url: %v", err)
	}
	if got := parsed.Query().Get("client_id"); got != "default-client" {
		t.Fatalf("expected default client_id, got %q", got)
	}
}

// TestHTTPProviderConfiguredForRedirectRejectsUnknownTenantWithoutDefault verifies tenant-only setups fail closed.
// TestHTTPProviderConfiguredForRedirectRejectsUnknownTenantWithoutDefault 验证仅配置租户凭据时，未知业务域名会被拒绝。
func TestHTTPProviderConfiguredForRedirectRejectsUnknownTenantWithoutDefault(t *testing.T) {
	provider := NewHTTPProvider(ProviderConfig{
		Name:        "github",
		AuthURL:     "https://github.com/login/oauth/authorize",
		TokenURL:    "https://github.com/login/oauth/access_token",
		UserInfoURL: "https://api.github.com/user",
		Tenants: map[string]ProviderTenantConfig{
			"label.example.com": {
				ClientID:     "label-client",
				ClientSecret: "label-secret",
			},
		},
	})

	if !provider.ConfiguredForRedirect("https://label.example.com/auth/oauth/callback?provider=github") {
		t.Fatal("expected label tenant redirect to be configured")
	}
	if provider.ConfiguredForRedirect("https://blockx.example.com/auth/oauth/callback?provider=github") {
		t.Fatal("expected unknown tenant redirect to be rejected when no default credentials exist")
	}
}

// TestNormalizeUserReadsGoogleEmailVerification keeps trusted-email merging provider-driven.
// TestNormalizeUserReadsGoogleEmailVerification 验证 Google 的 email_verified 会进入统一用户资料。
func TestNormalizeUserReadsGoogleEmailVerification(t *testing.T) {
	profile := normalizeUser("google", map[string]any{
		"sub":            "google-subject",
		"email":          "river@example.com",
		"email_verified": true,
		"name":           "River",
	})

	if !profile.EmailVerified {
		t.Fatal("expected google email to be marked verified")
	}
	if profile.Email != "river@example.com" {
		t.Fatalf("expected email to be normalized, got %q", profile.Email)
	}
}
