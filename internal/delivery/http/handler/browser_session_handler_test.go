package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
	auth "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
)

type fakeBrowserSessions struct {
	users    map[string]string
	revoked  map[string]bool
	selected string
}

func (s *fakeBrowserSessions) SessionStatus(_ context.Context, req auth.SessionStatusRequest) (*auth.SessionStatusResult, error) {
	id := s.users[req.SessionToken]
	if id == "" || s.revoked[req.SessionToken] {
		return nil, domain.NewError(auth.ErrInvalidRefreshToken, "expired")
	}
	return &auth.SessionStatusResult{UserID: id, Username: id, Email: id + "@example.test"}, nil
}
func (s *fakeBrowserSessions) LoginWithSession(ctx context.Context, req auth.SessionLoginRequest) (*auth.SessionLoginResult, error) {
	user, err := s.SessionStatus(ctx, auth.SessionStatusRequest{SessionToken: req.SessionToken})
	if err != nil {
		return nil, err
	}
	s.selected = user.UserID
	return &auth.SessionLoginResult{UserID: user.UserID, Username: user.Username, Token: &token.Pair{AccessToken: "access-" + user.UserID, RefreshToken: req.SessionToken, ExpiresAt: time.Now().Add(time.Hour)}}, nil
}
func (s *fakeBrowserSessions) Logout(_ context.Context, req auth.LogoutRequest) error {
	s.revoked[req.RefreshToken] = true
	return nil
}
func sessionTestServer() (*gin.Engine, *fakeBrowserSessions) {
	gin.SetMode(gin.TestMode)
	service := &fakeBrowserSessions{users: map[string]string{"token-a": "alice", "token-b": "bob"}, revoked: map[string]bool{}}
	h := &AuthHandler{sessions: service}
	r := gin.New()
	r.GET("/sessions", h.BrowserSessions)
	r.POST("/login", h.SessionLogin)
	r.POST("/logout", h.BrowserLogout)
	return r, service
}
func browserRequest(r *gin.Engine, method, path, body string, cookies map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func accountCookies() map[string]string {
	return map[string]string{accountCookiePrefix + accountSessionID("alice"): "token-a", accountCookiePrefix + accountSessionID("bob"): "token-b", authSessionCookieName: "token-b"}
}
func TestBrowserListsBothAccountsWithoutCredentials(t *testing.T) {
	r, _ := sessionTestServer()
	w := browserRequest(r, "GET", "/sessions?client_id=app", "", accountCookies())
	var payload struct {
		Data struct {
			Items []browserAccount `json:"items"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	if w.Code != 200 || len(payload.Data.Items) != 2 {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "token-a") || strings.Contains(w.Body.String(), "token-b") {
		t.Fatal("raw credentials exposed")
	}
}
func TestSelectingAliceIgnoresBobsLastLogin(t *testing.T) {
	r, s := sessionTestServer()
	w := browserRequest(r, "POST", "/login", `{"client_id":"app","session_id":"`+accountSessionID("alice")+`"}`, accountCookies())
	if w.Code != 200 || s.selected != "alice" {
		t.Fatalf("selected %q: %s", s.selected, w.Body.String())
	}
}
func TestMultiAccountLoginRequiresExplicitSelection(t *testing.T) {
	r, s := sessionTestServer()
	w := browserRequest(r, "POST", "/login", `{"client_id":"app"}`, accountCookies())
	if w.Code != 409 || s.selected != "" {
		t.Fatalf("ambiguous identity accepted: %s", w.Body.String())
	}
}

func TestSingleAccountLoginWorksWithoutCompatibilityCookie(t *testing.T) {
	r, s := sessionTestServer()
	w := browserRequest(r, "POST", "/login", `{"client_id":"app"}`, map[string]string{
		accountCookiePrefix + accountSessionID("bob"): "token-b",
	})
	if w.Code != 200 || s.selected != "bob" {
		t.Fatalf("remaining account not selected: %s", w.Body.String())
	}
}
func TestExpiredMissingAndForgedSelectionNeverFallsBack(t *testing.T) {
	for _, scenario := range []string{"expired", "missing", "forged"} {
		t.Run(scenario, func(t *testing.T) {
			r, s := sessionTestServer()
			cookies := accountCookies()
			name := accountCookiePrefix + accountSessionID("alice")
			switch scenario {
			case "expired":
				s.revoked["token-a"] = true
			case "missing":
				delete(cookies, name)
			case "forged":
				cookies[name] = "token-b"
			}
			w := browserRequest(r, "POST", "/login", `{"client_id":"app","session_id":"`+accountSessionID("alice")+`"}`, cookies)
			if w.Code != 401 || s.selected != "" {
				t.Fatalf("unexpected fallback: %s", w.Body.String())
			}
		})
	}
}
func TestLegacySessionMigratesAndExpiredAccountIsRemoved(t *testing.T) {
	r, s := sessionTestServer()
	s.revoked["token-a"] = true
	w := browserRequest(r, "GET", "/sessions?client_id=app", "", map[string]string{authSessionCookieName: "token-b", accountCookiePrefix + accountSessionID("alice"): "token-a"})
	migrated, cleared := false, false
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == accountCookiePrefix+accountSessionID("bob") && cookie.Value == "token-b" && cookie.HttpOnly {
			migrated = true
		}
		if cookie.Name == accountCookiePrefix+accountSessionID("alice") && cookie.MaxAge < 0 {
			cleared = true
		}
	}
	if !migrated || !cleared {
		t.Fatalf("migration=%v cleanup=%v", migrated, cleared)
	}
}
func TestLogoutOnePreservesOtherAccount(t *testing.T) {
	r, s := sessionTestServer()
	w := browserRequest(r, "POST", "/logout", `{"session_id":"`+accountSessionID("alice")+`"}`, accountCookies())
	if w.Code != 200 || !s.revoked["token-a"] || s.revoked["token-b"] {
		t.Fatalf("incorrect revocation: %+v", s.revoked)
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == authSessionCookieName || c.Name == accountCookiePrefix+accountSessionID("bob") {
			t.Fatal("unrelated cookie cleared")
		}
	}
}
func TestLogoutAllRevokesAllBrowserAccounts(t *testing.T) {
	r, s := sessionTestServer()
	w := browserRequest(r, "POST", "/logout", `{"all":true}`, accountCookies())
	if w.Code != 200 || !s.revoked["token-a"] || !s.revoked["token-b"] {
		t.Fatalf("incorrect revocation: %+v", s.revoked)
	}
}
func TestIndependentLoginCookiesPreserveOtherAccounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "https://auth.example.test/login", nil)
	setSessionCookie(c, "bob", "token-b")
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == accountCookiePrefix+accountSessionID("alice") {
			t.Fatal("alice overwritten")
		}
		if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
			t.Fatal("unsafe cookie attributes")
		}
	}
}
