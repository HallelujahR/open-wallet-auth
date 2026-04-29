package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config contains all runtime configuration.
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	HTTP       HTTPConfig       `mapstructure:"http"`
	Log        LogConfig        `mapstructure:"log"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Management ManagementConfig `mapstructure:"management"`
}

// AppConfig contains service identity settings.
type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

// HTTPConfig contains HTTP server settings.
type HTTPConfig struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
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

// ManagementConfig contains settings for management-only APIs.
type ManagementConfig struct {
	AdminToken string `mapstructure:"admin_token"`
}

// Load reads configuration from defaults, config files, and environment variables.
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

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "open-wallet-auth")
	v.SetDefault("app.env", "development")
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_header_timeout", "5s")
	v.SetDefault("http.read_timeout", "15s")
	v.SetDefault("http.write_timeout", "15s")
	v.SetDefault("http.idle_timeout", "60s")
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
	v.SetDefault("management.admin_token", "")
}
