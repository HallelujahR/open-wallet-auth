package clientaccess

import (
	"testing"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
)

func TestLegacyCallbackCompatibilityAndStateBoundValidation(t *testing.T) {
	c := &client.Client{Status: client.StatusActive}
	if ValidateReturnURI(c, "https://legacy.example/callback", false) != nil {
		t.Fatal("legacy callback protocol changed")
	}
	if ValidateReturnURI(c, "https://legacy.example/callback", true) == nil {
		t.Fatal("state-bound callback requires registration")
	}
	for _, raw := range []string{"javascript:alert(1)", "https://user:password@app.example/callback", "//app.example/callback"} {
		if ValidateReturnURI(c, raw, false) == nil {
			t.Fatal("unsafe legacy callback accepted")
		}
	}
	c.AllowedRedirectURIs = []string{"https://app.example/callback"}
	for _, bound := range []bool{false, true} {
		if ValidateReturnURI(c, "https://app.example/callback", bound) != nil || ValidateReturnURI(c, "https://app.example/callback?other=1", bound) == nil {
			t.Fatal("registered callback must match exactly")
		}
	}
	c.Status = client.StatusDisabled
	if ValidateReturnURI(c, "https://app.example/callback", false) == nil {
		t.Fatal("disabled client accepted")
	}
}
