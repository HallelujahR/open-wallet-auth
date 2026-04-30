package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateProductionAllowsDevelopmentDefaults confirms local demo defaults stay frictionless.
// TestValidateProductionAllowsDevelopmentDefaults 确认本地 Demo 默认配置不会被生产校验拦截。
func TestValidateProductionAllowsDevelopmentDefaults(t *testing.T) {
	cfg := Config{
		App: AppConfig{Env: "development"},
		Database: DatabaseConfig{
			AutoMigrate: true,
		},
		Phone: PhoneConfig{
			Enabled:       true,
			ExposeDevCode: true,
			Provider:      MessageProviderConfig{Type: "noop"},
		},
		Email: EmailConfig{
			VerificationEnabled: true,
			ExposeDevCode:       true,
			Provider:            MessageProviderConfig{Type: "noop"},
		},
		HTTP: HTTPConfig{
			CORSAllowedOrigins: []string{"null"},
		},
	}

	if err := cfg.ValidateProduction(); err != nil {
		t.Fatalf("development config should pass production validator: %v", err)
	}
}

// TestValidateProductionRejectsUnsafeSettings verifies production startup fails fast on risky settings.
// TestValidateProductionRejectsUnsafeSettings 验证生产环境会快速拒绝高风险配置。
func TestValidateProductionRejectsUnsafeSettings(t *testing.T) {
	cfg := Config{
		App:        AppConfig{Env: "production"},
		Management: ManagementConfig{AdminToken: "short"},
		Database:   DatabaseConfig{AutoMigrate: true},
		Phone: PhoneConfig{
			Enabled:       true,
			ExposeDevCode: true,
			Provider:      MessageProviderConfig{Type: "noop"},
		},
		Email: EmailConfig{
			VerificationEnabled: true,
			ExposeDevCode:       true,
			Provider:            MessageProviderConfig{Type: "noop"},
		},
		HTTP: HTTPConfig{
			CORSAllowedOrigins: []string{"*"},
		},
		JWT: JWTConfig{
			PrivateKeyPath: filepath.Join(t.TempDir(), "missing-private.pem"),
			PublicKeyPath:  filepath.Join(t.TempDir(), "missing-public.pem"),
		},
	}

	err := cfg.ValidateProduction()
	if err == nil {
		t.Fatal("production config should reject unsafe settings")
	}
	for _, want := range []string{
		"database.auto_migrate",
		"management.admin_token",
		"expose_dev_code",
		"phone.provider.type",
		"email.provider.type",
		"cors_allowed_origins",
		"private_key_path",
		"public_key_path",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q should mention %q", err.Error(), want)
		}
	}
}

// TestValidateProductionAcceptsHardenedSettings verifies a hardened production config can boot.
// TestValidateProductionAcceptsHardenedSettings 验证加固后的生产配置可以通过启动检查。
func TestValidateProductionAcceptsHardenedSettings(t *testing.T) {
	privateKeyPath := writeTempFile(t, "private.pem")
	publicKeyPath := writeTempFile(t, "public.pem")
	cfg := Config{
		App:        AppConfig{Env: "production"},
		Management: ManagementConfig{AdminToken: "0123456789abcdef0123456789abcdef"},
		Database:   DatabaseConfig{AutoMigrate: false},
		Phone: PhoneConfig{
			Enabled:       true,
			ExposeDevCode: false,
			Provider:      MessageProviderConfig{Type: "webhook", Webhook: WebhookConfig{URL: "https://sms.example.com/send"}},
		},
		Email: EmailConfig{
			VerificationEnabled: true,
			ExposeDevCode:       false,
			Provider:            MessageProviderConfig{Type: "smtp", SMTP: SMTPConfig{Host: "smtp.example.com", Port: 587, From: "noreply@example.com"}},
		},
		HTTP: HTTPConfig{
			CORSAllowedOrigins: []string{"https://console.example.com"},
		},
		JWT: JWTConfig{
			PrivateKeyPath: privateKeyPath,
			PublicKeyPath:  publicKeyPath,
		},
	}

	if err := cfg.ValidateProduction(); err != nil {
		t.Fatalf("hardened production config should pass: %v", err)
	}
}

// writeTempFile creates a small key placeholder for config validation tests.
// writeTempFile 为配置校验测试创建一个临时密钥占位文件。
func writeTempFile(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("test-key"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}
