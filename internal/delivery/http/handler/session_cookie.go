package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	authSessionCookieName = "owa_session"
	authSessionCookiePath = "/"
	authSessionCookieTTL  = 30 * 24 * time.Hour
)

// setSessionCookie stores the central auth session in an HttpOnly cookie.
// setSessionCookie 将中台登录会话写入 HttpOnly Cookie，业务前端无法直接读取刷新令牌。
func setSessionCookie(c *gin.Context, refreshToken string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     authSessionCookieName,
		Value:    refreshToken,
		Path:     authSessionCookiePath,
		MaxAge:   int(authSessionCookieTTL.Seconds()),
		HttpOnly: true,
		Secure:   requestIsSecure(c),
		SameSite: http.SameSiteLaxMode,
	})
}

// readSessionCookie returns the central auth session token from the request.
// readSessionCookie 从请求中读取中台登录会话 token。
func readSessionCookie(c *gin.Context) (string, bool) {
	value, err := c.Cookie(authSessionCookieName)
	if err != nil || strings.TrimSpace(value) == "" {
		return "", false
	}
	return value, true
}

// clearSessionCookie removes the central auth session from the browser.
// clearSessionCookie 清除浏览器中的中台登录会话。
func clearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     authSessionCookieName,
		Value:    "",
		Path:     authSessionCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   requestIsSecure(c),
		SameSite: http.SameSiteLaxMode,
	})
}

// requestIsSecure respects reverse-proxy HTTPS headers when deciding cookie security.
// requestIsSecure 兼容反向代理场景，根据 HTTPS 请求头判断是否写 Secure Cookie。
func requestIsSecure(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	return strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
}
