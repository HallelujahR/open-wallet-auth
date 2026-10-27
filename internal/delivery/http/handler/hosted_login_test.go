package handler

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	clientusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/client"
)

type hostedClients struct{ app *client.Client }

func (r hostedClients) FindByClientID(context.Context, string) (*client.Client, error) {
	return r.app, nil
}
func (r hostedClients) Create(context.Context, *client.Client) error  { return nil }
func (r hostedClients) List(context.Context) ([]client.Client, error) { return nil, nil }

func TestHostedLoginRetainsDeployedCallbackAndStateGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewClientHandler(clientusecase.NewService(hostedClients{&client.Client{Status: client.StatusActive, AllowedRedirectURIs: []string{"https://app.example/callback"}}}), nil)
	for _, sample := range []struct {
		uri, state, extra string
		allowed           bool
	}{
		{"https://app.example/callback", "opaque-state", "", true},
		{"https://app.example/callback", "", "", true},
		{"https://evil.example/callback", "opaque-state", "", false},
		{"https://app.example/callback", "opaque-state", "&return_uri=https://evil.example", false},
		{"https://app.example/callback", strings.Repeat("s", 513), "", false},
		{"", "opaque-state", "", false},
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		q := url.Values{"client_id": {"synthetic"}, "return_uri": {sample.uri}, "state": {sample.state}}
		c.Request = httptest.NewRequest("GET", "/login?"+q.Encode()+sample.extra, nil)
		if got := h.ValidateHostedLogin(c); got != sample.allowed {
			t.Fatalf("unexpected validation for allowed=%v", sample.allowed)
		}
		if !sample.allowed && w.Code != 400 {
			t.Fatalf("invalid callback returned %d", w.Code)
		}
	}
}
