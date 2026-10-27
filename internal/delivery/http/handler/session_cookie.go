package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

const (
	authSessionCookieName = "owa_session"
	authSessionCookiePath = "/"
	authSessionCookieTTL  = 30 * 24 * time.Hour
	accountCookiePrefix   = "owa_account_"
)

// Account cookies are independent so concurrent logins in separate tabs cannot
// overwrite another identity. Only HttpOnly credentials are stored in the browser;
// every selector is checked against the persisted, revocable server session.
func accountSessionID(userID string) string {
	sum := sha256.Sum256([]byte(userID))
	return hex.EncodeToString(sum[:16])
}
func writeSessionCookie(c *gin.Context, name, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{Name: name, Value: value, Path: authSessionCookiePath,
		MaxAge: maxAge, HttpOnly: true, Secure: requestIsSecure(c), SameSite: http.SameSiteLaxMode})
}
func setSessionCookie(c *gin.Context, userID, refreshToken string) {
	// The session-list endpoint migrates a validated legacy session before a new login.
	writeSessionCookie(c, accountCookiePrefix+accountSessionID(userID), refreshToken, int(authSessionCookieTTL.Seconds()))
	writeSessionCookie(c, authSessionCookieName, refreshToken, int(authSessionCookieTTL.Seconds()))
}
func readSessionCookie(c *gin.Context) (string, bool) {
	value, err := c.Cookie(authSessionCookieName)
	return value, err == nil && strings.TrimSpace(value) != ""
}
func clearTokenCookies(c *gin.Context, raw string) {
	for _, cookie := range c.Request.Cookies() {
		if (cookie.Name == authSessionCookieName || strings.HasPrefix(cookie.Name, accountCookiePrefix)) && cookie.Value == raw {
			writeSessionCookie(c, cookie.Name, "", -1)
		}
	}
}
func validSessionID(id string) bool {
	decoded, err := hex.DecodeString(id)
	return err == nil && len(decoded) == 16
}
func requestIsSecure(c *gin.Context) bool {
	return c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
}
