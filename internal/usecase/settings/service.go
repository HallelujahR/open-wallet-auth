package settings

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

const runtimeSettingsKey = "runtime.providers"

// Service manages runtime-editable provider settings.
// Service 管理运行期可编辑的服务商配置。
type Service struct {
	store    repository.SettingsRepository
	defaults Snapshot
	readonly ReadonlySettings
}

// Snapshot is the full provider settings payload used by management APIs.
// Snapshot 是管理接口读写的完整服务商配置快照。
type Snapshot struct {
	HTTP  HTTPSettings  `json:"http"`
	Phone PhoneSettings `json:"phone"`
	Email EmailSettings `json:"email"`
	OAuth OAuthSettings `json:"oauth"`
}

// HTTPSettings contains browser-facing HTTP settings that can change at runtime.
// HTTPSettings 保存可运行期调整的浏览器访问配置。
type HTTPSettings struct {
	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
}

// PhoneSettings contains phone login and SMS provider settings.
// PhoneSettings 保存手机号登录开关和短信服务商配置。
type PhoneSettings struct {
	Enabled  bool                    `json:"enabled"`
	Provider MessageProviderSettings `json:"provider"`
}

// EmailSettings contains email verification and email provider settings.
// EmailSettings 保存邮箱验证开关和邮件服务商配置。
type EmailSettings struct {
	VerificationEnabled bool                    `json:"verification_enabled"`
	Provider            MessageProviderSettings `json:"provider"`
}

// MessageProviderSettings contains webhook, SMTP, and Aliyun SMS settings.
// MessageProviderSettings 保存 Webhook、SMTP 和阿里云短信配置。
type MessageProviderSettings struct {
	Type      string            `json:"type"`
	Webhook   WebhookSettings   `json:"webhook"`
	SMTP      SMTPSettings      `json:"smtp"`
	AliyunSMS AliyunSMSSettings `json:"aliyun_sms"`
	Headers   map[string]string `json:"headers"`
}

// WebhookSettings contains a generic HTTP message gateway config.
// WebhookSettings 保存通用 HTTP 消息网关配置。
type WebhookSettings struct {
	URL         string `json:"url"`
	BearerToken string `json:"bearer_token,omitempty"`
}

// SMTPSettings contains SMTP email provider config.
// SMTPSettings 保存 SMTP 邮件服务商配置。
type SMTPSettings struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	From     string `json:"from"`
}

// AliyunSMSSettings contains Aliyun SMS provider config.
// AliyunSMSSettings 保存阿里云短信服务商配置。
type AliyunSMSSettings struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret,omitempty"`
	SignName        string `json:"sign_name"`
	TemplateCode    string `json:"template_code"`
	RegionID        string `json:"region_id"`
	Endpoint        string `json:"endpoint"`
}

// OAuthSettings contains supported OAuth provider settings.
// OAuthSettings 保存当前支持的 OAuth 服务商配置。
type OAuthSettings struct {
	Google OAuthProviderSettings `json:"google"`
	GitHub OAuthProviderSettings `json:"github"`
}

// OAuthProviderSettings contains one OAuth provider's endpoints and credentials.
// OAuthProviderSettings 保存单个 OAuth 服务商的端点和凭据。
type OAuthProviderSettings struct {
	ClientID          string                        `json:"client_id"`
	ClientSecret      string                        `json:"client_secret,omitempty"`
	AuthURL           string                        `json:"auth_url"`
	TokenURL          string                        `json:"token_url"`
	UserInfoURL       string                        `json:"user_info_url"`
	Scopes            []string                      `json:"scopes"`
	TenantCredentials []OAuthProviderTenantSettings `json:"tenant_credentials"`
}

// OAuthProviderTenantSettings overrides OAuth credentials for one host.
// OAuthProviderTenantSettings 按业务域名覆盖 OAuth 凭据。
type OAuthProviderTenantSettings struct {
	Host         string `json:"host"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// SecretStatus tells the UI whether a secret is configured without exposing it.
// SecretStatus 告诉前端密钥是否已配置，但不返回密钥明文。
type SecretStatus struct {
	Configured bool   `json:"configured"`
	Masked     string `json:"masked"`
}

// ReadonlySettings contains startup-level settings shown but not editable in UI.
// ReadonlySettings 保存只展示不允许页面修改的启动级配置。
type ReadonlySettings struct {
	App      ReadonlyAppSettings      `json:"app"`
	HTTP     ReadonlyHTTPSettings     `json:"http"`
	Database ReadonlyDatabaseSettings `json:"database"`
	Redis    ReadonlyRedisSettings    `json:"redis"`
	JWT      ReadonlyJWTSettings      `json:"jwt"`
}

// ReadonlyAppSettings contains service identity.
// ReadonlyAppSettings 保存服务身份信息。
type ReadonlyAppSettings struct {
	Name string `json:"name"`
	Env  string `json:"env"`
}

// ReadonlyHTTPSettings contains startup HTTP listener settings.
// ReadonlyHTTPSettings 保存启动级 HTTP 监听配置。
type ReadonlyHTTPSettings struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// ReadonlyDatabaseSettings contains database connection settings.
// ReadonlyDatabaseSettings 保存数据库连接配置。
type ReadonlyDatabaseSettings struct {
	Driver      string `json:"driver"`
	DSN         string `json:"dsn"`
	AutoMigrate bool   `json:"auto_migrate"`
}

// ReadonlyRedisSettings contains Redis connection settings.
// ReadonlyRedisSettings 保存 Redis 连接配置。
type ReadonlyRedisSettings struct {
	Enabled  bool   `json:"enabled"`
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// ReadonlyJWTSettings contains JWT key and lifetime settings.
// ReadonlyJWTSettings 保存 JWT 密钥路径和生命周期配置。
type ReadonlyJWTSettings struct {
	Issuer          string `json:"issuer"`
	AccessTokenTTL  string `json:"access_token_ttl"`
	RefreshTokenTTL string `json:"refresh_token_ttl"`
	PrivateKeyPath  string `json:"private_key_path"`
	PublicKeyPath   string `json:"public_key_path"`
	ActiveKeyID     string `json:"active_key_id"`
}

// PublicSnapshot is a redacted settings snapshot returned to the admin console.
// PublicSnapshot 是返回管理后台的脱敏配置快照。
type PublicSnapshot struct {
	Settings Snapshot                `json:"settings"`
	Secrets  map[string]SecretStatus `json:"secrets"`
	Readonly ReadonlySettings        `json:"readonly"`
}

// NewService creates a settings service with immutable config defaults.
// NewService 创建系统配置服务，并注入启动配置作为默认值。
func NewService(store repository.SettingsRepository, defaults Snapshot, readonly ReadonlySettings) *Service {
	return &Service{store: store, defaults: defaults, readonly: readonly}
}

// Get returns merged settings, using database overrides when present.
// Get 返回合并后的配置；数据库配置存在时覆盖启动默认值。
func (s *Service) Get(ctx context.Context) (Snapshot, error) {
	current := s.defaults
	raw, err := s.store.Get(ctx, runtimeSettingsKey)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return current, nil
		}
		return Snapshot{}, err
	}
	if len(raw) == 0 {
		return current, nil
	}
	if err := json.Unmarshal(raw, &current); err != nil {
		return Snapshot{}, err
	}
	return current, nil
}

// Public returns a redacted settings snapshot for the admin UI.
// Public 返回管理后台使用的脱敏配置快照。
func (s *Service) Public(ctx context.Context) (*PublicSnapshot, error) {
	current, err := s.Get(ctx)
	if err != nil {
		return nil, err
	}
	redacted, secrets := redact(current)
	return &PublicSnapshot{Settings: redacted, Secrets: secrets, Readonly: redactReadonly(s.readonly)}, nil
}

// PhoneEnabled returns the current phone-login switch.
// PhoneEnabled 返回当前手机号登录开关。
func (s *Service) PhoneEnabled(ctx context.Context) (bool, error) {
	current, err := s.Get(ctx)
	if err != nil {
		return false, err
	}
	return current.Phone.Enabled, nil
}

// EmailVerificationEnabled returns the current email-verification switch.
// EmailVerificationEnabled 返回当前邮箱验证开关。
func (s *Service) EmailVerificationEnabled(ctx context.Context) (bool, error) {
	current, err := s.Get(ctx)
	if err != nil {
		return false, err
	}
	return current.Email.VerificationEnabled, nil
}

// CORSAllowedOrigins returns current browser origins allowed to call the auth service.
// CORSAllowedOrigins 返回当前允许调用认证服务的浏览器来源。
func (s *Service) CORSAllowedOrigins(ctx context.Context) ([]string, error) {
	current, err := s.Get(ctx)
	if err != nil {
		return nil, err
	}
	return current.HTTP.CORSAllowedOrigins, nil
}

// Update validates and persists editable provider settings.
// Update 校验并持久化管理后台提交的服务商配置。
func (s *Service) Update(ctx context.Context, next Snapshot) (*PublicSnapshot, error) {
	current, err := s.Get(ctx)
	if err != nil {
		return nil, err
	}
	merged := mergeSecrets(current, next)
	normalize(&merged)
	raw, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}
	if err := s.store.Upsert(ctx, runtimeSettingsKey, raw); err != nil {
		return nil, err
	}
	return s.Public(ctx)
}

// normalize trims common string fields before persistence.
// normalize 在持久化前清理常见字符串字段。
func normalize(s *Snapshot) {
	s.HTTP.CORSAllowedOrigins = normalizeOrigins(s.HTTP.CORSAllowedOrigins)
	s.Phone.Provider.Type = strings.ToLower(strings.TrimSpace(s.Phone.Provider.Type))
	s.Email.Provider.Type = strings.ToLower(strings.TrimSpace(s.Email.Provider.Type))
	normalizeOAuth(&s.OAuth.Google)
	normalizeOAuth(&s.OAuth.GitHub)
}

// normalizeOrigins trims empty origins and removes duplicates.
// normalizeOrigins 清理空来源并去重。
func normalizeOrigins(origins []string) []string {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		normalized = append(normalized, origin)
	}
	return normalized
}

// normalizeOAuth trims OAuth provider and tenant fields.
// normalizeOAuth 清理 OAuth provider 和租户字段。
func normalizeOAuth(p *OAuthProviderSettings) {
	p.ClientID = strings.TrimSpace(p.ClientID)
	p.AuthURL = strings.TrimSpace(p.AuthURL)
	p.TokenURL = strings.TrimSpace(p.TokenURL)
	p.UserInfoURL = strings.TrimSpace(p.UserInfoURL)
	for i := range p.TenantCredentials {
		p.TenantCredentials[i].Host = strings.ToLower(strings.TrimSpace(p.TenantCredentials[i].Host))
		p.TenantCredentials[i].ClientID = strings.TrimSpace(p.TenantCredentials[i].ClientID)
	}
}

// mergeSecrets keeps existing secrets when the UI submits an empty secret field.
// mergeSecrets 在前端提交空密钥时保留已有密钥，避免脱敏表单误清空凭据。
func mergeSecrets(current Snapshot, next Snapshot) Snapshot {
	if next.Phone.Provider.Webhook.BearerToken == "" {
		next.Phone.Provider.Webhook.BearerToken = current.Phone.Provider.Webhook.BearerToken
	}
	if next.Phone.Provider.AliyunSMS.AccessKeySecret == "" {
		next.Phone.Provider.AliyunSMS.AccessKeySecret = current.Phone.Provider.AliyunSMS.AccessKeySecret
	}
	if next.Email.Provider.Webhook.BearerToken == "" {
		next.Email.Provider.Webhook.BearerToken = current.Email.Provider.Webhook.BearerToken
	}
	if next.Email.Provider.SMTP.Password == "" {
		next.Email.Provider.SMTP.Password = current.Email.Provider.SMTP.Password
	}
	mergeOAuthSecrets(&current.OAuth.Google, &next.OAuth.Google)
	mergeOAuthSecrets(&current.OAuth.GitHub, &next.OAuth.GitHub)
	return next
}

// mergeOAuthSecrets preserves default and tenant OAuth secrets.
// mergeOAuthSecrets 保留默认和租户级 OAuth 密钥。
func mergeOAuthSecrets(current *OAuthProviderSettings, next *OAuthProviderSettings) {
	if next.ClientSecret == "" {
		next.ClientSecret = current.ClientSecret
	}
	byHost := map[string]string{}
	for _, tenant := range current.TenantCredentials {
		byHost[strings.ToLower(strings.TrimSpace(tenant.Host))] = tenant.ClientSecret
	}
	for i := range next.TenantCredentials {
		if next.TenantCredentials[i].ClientSecret == "" {
			next.TenantCredentials[i].ClientSecret = byHost[strings.ToLower(strings.TrimSpace(next.TenantCredentials[i].Host))]
		}
	}
}

// redact removes secret values and records their configured state.
// redact 移除密钥明文，并记录密钥是否已配置。
func redact(snapshot Snapshot) (Snapshot, map[string]SecretStatus) {
	secrets := map[string]SecretStatus{}
	record := func(key string, value string) {
		secrets[key] = SecretStatus{Configured: value != "", Masked: mask(value)}
	}
	record("phone.provider.webhook.bearer_token", snapshot.Phone.Provider.Webhook.BearerToken)
	snapshot.Phone.Provider.Webhook.BearerToken = ""
	record("phone.provider.aliyun_sms.access_key_secret", snapshot.Phone.Provider.AliyunSMS.AccessKeySecret)
	snapshot.Phone.Provider.AliyunSMS.AccessKeySecret = ""
	record("email.provider.webhook.bearer_token", snapshot.Email.Provider.Webhook.BearerToken)
	snapshot.Email.Provider.Webhook.BearerToken = ""
	record("email.provider.smtp.password", snapshot.Email.Provider.SMTP.Password)
	snapshot.Email.Provider.SMTP.Password = ""
	redactOAuth("oauth.google", &snapshot.OAuth.Google, secrets)
	redactOAuth("oauth.github", &snapshot.OAuth.GitHub, secrets)
	return snapshot, secrets
}

// redactOAuth removes OAuth secrets from one provider.
// redactOAuth 移除单个 OAuth 服务商的密钥明文。
func redactOAuth(prefix string, provider *OAuthProviderSettings, secrets map[string]SecretStatus) {
	secrets[prefix+".client_secret"] = SecretStatus{Configured: provider.ClientSecret != "", Masked: mask(provider.ClientSecret)}
	provider.ClientSecret = ""
	for i := range provider.TenantCredentials {
		key := prefix + ".tenant_credentials." + provider.TenantCredentials[i].Host + ".client_secret"
		secrets[key] = SecretStatus{Configured: provider.TenantCredentials[i].ClientSecret != "", Masked: mask(provider.TenantCredentials[i].ClientSecret)}
		provider.TenantCredentials[i].ClientSecret = ""
	}
}

// mask returns a short stable hint for configured secrets.
// mask 返回密钥的短脱敏提示。
func mask(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 6 {
		return "******"
	}
	return value[:3] + "******" + value[len(value)-3:]
}

// redactReadonly masks sensitive startup settings.
// redactReadonly 对启动级敏感配置做脱敏。
func redactReadonly(readonly ReadonlySettings) ReadonlySettings {
	readonly.Database.DSN = maskDSN(readonly.Database.DSN)
	readonly.Redis.Password = mask(readonly.Redis.Password)
	return readonly
}

// maskDSN keeps database type and host context while hiding credentials.
// maskDSN 保留数据库类型和主机上下文，同时隐藏账号密码。
func maskDSN(value string) string {
	if value == "" {
		return ""
	}
	at := strings.LastIndex(value, "@")
	scheme := strings.Index(value, "://")
	if at > 0 && scheme > 0 && scheme+3 < at {
		return value[:scheme+3] + "***:***" + value[at:]
	}
	return mask(value)
}
