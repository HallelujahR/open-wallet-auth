package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
)

// ClientRepository defines persistence operations for application clients.
type ClientRepository interface {
	FindByClientID(ctx context.Context, clientID string) (*client.Client, error)
}
