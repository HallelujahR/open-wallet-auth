package clientaccess

import (
	"net/url"

	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
)

const ErrInvalidRedirectURI = "CLIENT_REDIRECT_DENIED"

// ValidateReturnURI preserves the deployed legacy protocol for clients without
// a callback list. State-bound flows always require an exact registered callback.
func ValidateReturnURI(c *client.Client, raw string, bound bool) error {
	denied := func() error { return domain.NewError(ErrInvalidRedirectURI, "callback address is not allowed") }
	if c == nil || !c.IsActive() {
		return denied()
	}
	if raw == "" && !bound {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return denied()
	}
	if len(c.AllowedRedirectURIs) == 0 && !bound {
		return nil
	}
	if !c.AllowsReturnURI(raw) {
		return denied()
	}
	return nil
}
