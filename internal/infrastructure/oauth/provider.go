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
type ProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
}

// HTTPProvider implements OAuth code exchange and userinfo loading with net/http.
type HTTPProvider struct {
	cfg        ProviderConfig
	httpClient *http.Client
}

// NewHTTPProvider creates a generic HTTP OAuth provider adapter.
func NewHTTPProvider(cfg ProviderConfig) *HTTPProvider {
	return &HTTPProvider{cfg: cfg, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (p *HTTPProvider) Name() string {
	return strings.ToLower(p.cfg.Name)
}

func (p *HTTPProvider) Configured() bool {
	return p.cfg.ClientID != "" && p.cfg.ClientSecret != "" && p.cfg.AuthURL != "" && p.cfg.TokenURL != "" && p.cfg.UserInfoURL != ""
}

func (p *HTTPProvider) AuthURL(state string, redirectURI string) string {
	values := url.Values{}
	values.Set("client_id", p.cfg.ClientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("response_type", "code")
	values.Set("state", state)
	if len(p.cfg.Scopes) > 0 {
		values.Set("scope", strings.Join(p.cfg.Scopes, " "))
	}
	return p.cfg.AuthURL + "?" + values.Encode()
}

func (p *HTTPProvider) FetchUser(ctx context.Context, code string, redirectURI string) (*oauthusecase.ProviderUser, error) {
	if !p.Configured() {
		return nil, errors.New("oauth provider is not configured")
	}
	accessToken, err := p.exchangeToken(ctx, code, redirectURI)
	if err != nil {
		return nil, err
	}
	return p.fetchUser(ctx, accessToken)
}

func (p *HTTPProvider) exchangeToken(ctx context.Context, code string, redirectURI string) (string, error) {
	form := url.Values{}
	form.Set("client_id", p.cfg.ClientID)
	form.Set("client_secret", p.cfg.ClientSecret)
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
	return normalizeUser(p.Name(), raw), nil
}

func normalizeUser(provider string, raw map[string]any) *oauthusecase.ProviderUser {
	switch provider {
	case "github":
		return &oauthusecase.ProviderUser{
			Subject:   stringValue(raw["id"]),
			Email:     stringValue(raw["email"]),
			Username:  stringValue(raw["login"]),
			AvatarURL: stringValue(raw["avatar_url"]),
		}
	default:
		return &oauthusecase.ProviderUser{
			Subject:   firstString(raw["sub"], raw["id"]),
			Email:     stringValue(raw["email"]),
			Username:  firstString(raw["name"], raw["login"], raw["email"]),
			AvatarURL: firstString(raw["picture"], raw["avatar_url"]),
		}
	}
}

func firstString(values ...any) string {
	for _, value := range values {
		if s := stringValue(value); s != "" {
			return s
		}
	}
	return ""
}

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

func jsonNumber(value float64) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}
