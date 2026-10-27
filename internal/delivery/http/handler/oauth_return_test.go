package handler

import (
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	oauth "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/oauth"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestOAuthReturnSupportsHistoryAndHashRoutes(t *testing.T) {
	for _, tt := range []struct{ uri, prefix string }{
		{"https://app.example/callback", "https://app.example/callback#access_token="},
		{"https://app.example/#/callback", "https://app.example/#/callback?access_token="},
		{"https://app.example/#/callback?source=auth", "https://app.example/#/callback?source=auth&access_token="},
	} {
		t.Run(tt.uri, func(t *testing.T) {
			got := oauthReturnURL(&oauth.CallbackResult{ReturnURI: tt.uri, Token: &token.Pair{AccessToken: "test", ExpiresAt: time.Now()}})
			if !strings.HasPrefix(got, tt.prefix) || strings.Count(got, "#") != 1 {
				t.Fatalf("unexpected callback: %s", got)
			}
		})
	}
}

func TestOAuthReturnPreservesBusinessStateInFragment(t *testing.T) {
	state := "opaque+state/&?=#"
	got := oauthReturnURL(&oauth.CallbackResult{ReturnURI: "https://app.example/callback", ClientState: state, Token: &token.Pair{AccessToken: "test", ExpiresAt: time.Now()}})
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	_, fragment, _ := strings.Cut(got, "#")
	values, err := url.ParseQuery(fragment)
	if err != nil || values.Get("state") != state || u.RawQuery != "" {
		t.Fatal("business state must round-trip only in the fragment")
	}
}
