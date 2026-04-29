package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config contains all runtime configuration.
// Config 汇总服务运行所需的全部配置。
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	HTTP       HTTPConfig       `mapstructure:"http"`
	Log        LogConfig        `mapstructure:"log"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Wallet     WalletConfig     `mapstructure:"wallet"`
	Phone      PhoneConfig      `mapstructure:"phone"`
	Email      EmailConfig      `mapstructure:"email"`
	OAuth      OAuthConfig      `mapstructure:"oauth"`
	Management ManagementConfig `mapstructure:"management"`
}

// AppConfig contains service identity settings.
type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

// HTTPConfig contains HTTP server settings.
type HTTPConfig struct {
	Host               string        `mapstructure:"host"`
	Port               int           `mapstructure:"port"`
	ReadHeaderTimeout  time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
	CORSAllowedOrigins []string      `mapstructure:"cors_allowed_origins"`
}

// LogConfig contains structured logging settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// DatabaseConfig contains PostgreSQL connection settings.
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	DSN             string        `mapstructure:"dsn"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig contains Redis connection settings.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig contains token signing and lifetime settings.
type JWTConfig struct {
	Issuer          string        `mapstructure:"issuer"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	PrivateKeyPath  string        `mapstructure:"private_key_path"`
	PublicKeyPath   string        `mapstructure:"public_key_path"`
	ActiveKeyID     string        `mapstructure:"active_key_id"`
}

// WalletConfig contains wallet authentication settings.
type WalletConfig struct {
	NonceTTL time.Duration `mapstructure:"nonce_ttl"`
}

// PhoneConfig contains phone verification-code login settings.
type PhoneConfig struct {
	Enabled       bool                  `mapstructure:"enabled"`
	CodeTTL       time.Duration         `mapstructure:"code_ttl"`
	DevCode       string                `mapstructure:"dev_code"`
	ExposeDevCode bool                  `mapstructure:"expose_dev_code"`
	Provider      MessageProviderConfig `mapstructure:"provider"`
}

// EmailConfig contains email verification settings.
type EmailConfig struct {
	VerificationEnabled bool                  `mapstructure:"verification_enabled"`
	CodeTTL             time.Duration         `mapstructure:"code_ttl"`
	DevCode             string                `mapstructure:"dev_code"`
	ExposeDevCode       bool                  `mapstructure:"expose_dev_code"`
	Provider            MessageProviderConfig `mapstructure:"provider"`
}

// MessageProviderConfig contains webhook-style message provider settings.
type MessageProviderConfig struct {
	Type        string            `mapstructure:"type"`
	WebhookURL  string            `mapstructure:"webhook_url"`
	BearerToken string            `mapstructure:"bearer_token"`
	Headers     map[string]string `mapstructure:"headers"`
}

// OAuthConfig contains third-party OAuth provider settings.
type OAuthConfig struct {
	StateTTL time.Duration       `mapstructure:"state_ttl"`
	Google   OAuthProviderConfig `mapstructure:"google"`
	GitHub   OAuthProviderConfig `mapstructure:"github"`
}

// OAuthProviderConfig contains one OAuth provider's credentials.
type OAuthProviderConfig struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	AuthURL      string   `mapstructure:"auth_url"`
	TokenURL     string   `mapstructure:"token_url"`
	UserInfoURL  string   `mapstructure:"user_info_url"`
	Scopes       []string `mapstructure:"scopes"`
}

// ManagementConfig contains settings for management-only APIs.
type ManagementConfig struct {
	AdminToken string `mapstructure:"admin_token"`
}

// Load reads configuration from defaults, config files, and environment variables.
// Load 按默认值、配置文件、环境变量的优先级读取运行配置。
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	setDefaults(v)

	v.SetEnvPrefix("OWA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// setDefaults defines safe local-development defaults for every config section.
// setDefaults 为每个配置段设置本地开发可用的默认值。
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "open-wallet-auth")
	v.SetDefault("app.env", "development")
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_header_timeout", "5s")
	v.SetDefault("http.read_timeout", "15s")
	v.SetDefault("http.write_timeout", "15s")
	v.SetDefault("http.idle_timeout", "60s")
	v.SetDefault("http.cors_allowed_origins", []string{"http://localhost:3000", "http://localhost:5173", "null"})
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.dsn", "postgres://open_wallet_auth:open_wallet_auth@localhost:5432/open_wallet_auth?sslmode=disable")
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "30m")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.issuer", "open-wallet-auth")
	v.SetDefault("jwt.access_token_ttl", "15m")
	v.SetDefault("jwt.refresh_token_ttl", "720h")
	v.SetDefault("jwt.private_key_path", "./configs/jwt_private.pem")
	v.SetDefault("jwt.public_key_path", "./configs/jwt_public.pem")
	v.SetDefault("jwt.active_key_id", "default")
	v.SetDefault("wallet.nonce_ttl", "5m")
	v.SetDefault("phone.enabled", true)
	v.SetDefault("phone.code_ttl", "5m")
	v.SetDefault("phone.dev_code", "123456")
	v.SetDefault("phone.expose_dev_code", true)
	v.SetDefault("phone.provider.type", "noop")
	v.SetDefault("email.verification_enabled", true)
	v.SetDefault("email.code_ttl", "15m")
	v.SetDefault("email.dev_code", "123456")
	v.SetDefault("email.expose_dev_code", true)
	v.SetDefault("email.provider.type", "noop")
	v.SetDefault("oauth.state_ttl", "10m")
	v.SetDefault("oauth.google.auth_url", "https://accounts.google.com/o/oauth2/v2/auth")
	v.SetDefault("oauth.google.token_url", "https://oauth2.googleapis.com/token")
	v.SetDefault("oauth.google.user_info_url", "https://openidconnect.googleapis.com/v1/userinfo")
	v.SetDefault("oauth.google.scopes", []string{"openid", "email", "profile"})
	v.SetDefault("oauth.github.auth_url", "https://github.com/login/oauth/authorize")
	v.SetDefault("oauth.github.token_url", "https://github.com/login/oauth/access_token")
	v.SetDefault("oauth.github.user_info_url", "https://api.github.com/user")
	v.SetDefault("oauth.github.scopes", []string{"read:user", "user:email"})
	v.SetDefault("management.admin_token", "")
}
