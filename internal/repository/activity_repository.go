package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/audit"
)

// ActivityRepository records login events and user-client activity.
type ActivityRepository interface {
	RecordLogin(ctx context.Context, log *audit.LoginLog) error
	UpsertUserClientLogin(ctx context.Context, userID string, clientID string) error
}
