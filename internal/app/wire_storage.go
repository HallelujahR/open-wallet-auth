package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/model"
	pgrepo "github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/repository"
)

// storageBundle groups persistence adapters used by application usecases.
// storageBundle 汇总应用用例需要的持久化适配器，避免 app.New 直接暴露装配细节。
type storageBundle struct {
	sqlDB         *sql.DB
	users         *pgrepo.UserRepository
	clients       *pgrepo.ClientRepository
	refreshTokens *pgrepo.RefreshTokenRepository
	activity      *pgrepo.ActivityRepository
	wallets       *pgrepo.WalletRepository
	accounts      *pgrepo.OAuthAccountRepository
	settings      *pgrepo.SettingsRepository
}

// newStorage opens PostgreSQL, runs optional auto migration, and creates repositories.
// newStorage 打开 PostgreSQL、按配置执行自动迁移，并创建各类仓储适配器。
func newStorage(ctx context.Context, cfg *config.Config) (*storageBundle, error) {
	db, sqlDB, err := postgres.Open(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if cfg.Database.AutoMigrate {
		if err := db.AutoMigrate(&model.User{}, &model.Client{}, &model.UserWallet{}, &model.OAuthAccount{}, &model.WalletNonce{}, &model.RefreshToken{}, &model.LoginLog{}, &model.SecurityEvent{}, &model.UserClient{}, &model.SystemSetting{}); err != nil {
			return nil, fmt.Errorf("auto migrate database: %w", err)
		}
	}

	clients := pgrepo.NewClientRepository(db)
	if err := clients.EnsureDefault(ctx); err != nil {
		return nil, fmt.Errorf("ensure default client: %w", err)
	}

	return &storageBundle{
		sqlDB:         sqlDB,
		users:         pgrepo.NewUserRepository(db),
		clients:       clients,
		refreshTokens: pgrepo.NewRefreshTokenRepository(db),
		activity:      pgrepo.NewActivityRepository(db),
		wallets:       pgrepo.NewWalletRepository(db),
		accounts:      pgrepo.NewOAuthAccountRepository(db),
		settings:      pgrepo.NewSettingsRepository(db),
	}, nil
}
