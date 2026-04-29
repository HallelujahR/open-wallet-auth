package client

import (
	"context"
	"errors"
	"testing"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	clientdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	"github.com/open-wallet-auth/open-wallet-auth/internal/repository"
)

func TestServiceCreateClient(t *testing.T) {
	clients := newMemoryClients()
	service := NewService(clients)

	created, err := service.Create(context.Background(), CreateRequest{
		ClientID: "example-app",
		Name:     "Example App",
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if created.JWTAudience != "example-app" {
		t.Fatalf("expected default audience, got %s", created.JWTAudience)
	}
}

func TestServiceCreateRejectsDuplicateClient(t *testing.T) {
	clients := newMemoryClients()
	clients.byID["example-app"] = &clientdomain.Client{
		ClientID: "example-app",
		Status:   clientdomain.StatusActive,
	}
	service := NewService(clients)

	_, err := service.Create(context.Background(), CreateRequest{
		ClientID: "example-app",
		Name:     "Example App",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrClientAlreadyExists {
		t.Fatalf("expected %s, got %v", ErrClientAlreadyExists, err)
	}
}

func TestServiceResolveAudience(t *testing.T) {
	clients := newMemoryClients()
	clients.byID["example-app"] = &clientdomain.Client{
		ClientID:    "example-app",
		JWTAudience: "example-audience",
		Status:      clientdomain.StatusActive,
	}
	service := NewService(clients)

	audience, err := service.ResolveAudience(context.Background(), "example-app")
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if audience != "example-audience" {
		t.Fatalf("unexpected audience: %s", audience)
	}
}

type memoryClients struct {
	byID map[string]*clientdomain.Client
}

func newMemoryClients() *memoryClients {
	return &memoryClients{byID: map[string]*clientdomain.Client{}}
}

func (m *memoryClients) FindByClientID(ctx context.Context, clientID string) (*clientdomain.Client, error) {
	client, ok := m.byID[clientID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return client, nil
}

func (m *memoryClients) Create(ctx context.Context, client *clientdomain.Client) error {
	m.byID[client.ClientID] = client
	return nil
}

func (m *memoryClients) List(ctx context.Context) ([]clientdomain.Client, error) {
	clients := make([]clientdomain.Client, 0, len(m.byID))
	for _, client := range m.byID {
		clients = append(clients, *client)
	}
	return clients, nil
}
