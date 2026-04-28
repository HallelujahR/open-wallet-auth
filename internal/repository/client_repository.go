package repository

import (
	"context"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
)

type ClientRepository interface {
	FindByClientID(ctx context.Context, clientID string) (*client.Client, error)
}
