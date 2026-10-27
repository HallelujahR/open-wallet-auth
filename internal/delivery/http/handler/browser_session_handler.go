package handler

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	authusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
)

type browserSessionService interface {
	SessionStatus(context.Context, authusecase.SessionStatusRequest) (*authusecase.SessionStatusResult, error)
	LoginWithSession(context.Context, authusecase.SessionLoginRequest) (*authusecase.SessionLoginResult, error)
	Logout(context.Context, authusecase.LogoutRequest) error
}

type browserAccount struct {
	SessionID string `json:"session_id"`
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
}

// BrowserSessions returns only identities proven by this browser's credentials.
// Raw credentials are never exposed in the response or JavaScript storage.
func (h *AuthHandler) BrowserSessions(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	accounts := map[string]browserAccount{}
	// Account cookies take precedence over the compatibility cookie.
	cookies := c.Request.Cookies()
	sort.SliceStable(cookies, func(i, j int) bool { return cookies[i].Name < cookies[j].Name })
	for _, cookie := range cookies {
		legacy := cookie.Name == authSessionCookieName
		if !legacy && (!strings.HasPrefix(cookie.Name, accountCookiePrefix) || !validSessionID(strings.TrimPrefix(cookie.Name, accountCookiePrefix))) {
			continue
		}
		user, err := h.sessions.SessionStatus(c.Request.Context(), authusecase.SessionStatusRequest{ClientID: c.Query("client_id"), SessionToken: cookie.Value})
		if err != nil {
			var appErr *domain.Error
			if errors.As(err, &appErr) && appErr.Code == authusecase.ErrInvalidRefreshToken {
				writeSessionCookie(c, cookie.Name, "", -1)
				continue
			}
			writeAuthError(c, err)
			return
		}
		id := accountSessionID(user.UserID)
		if !legacy && cookie.Name != accountCookiePrefix+id {
			continue
		}
		if _, exists := accounts[id]; exists {
			continue
		}
		accounts[id] = browserAccount{SessionID: id, ID: user.UserID, Username: user.Username, Email: user.Email, Phone: user.Phone, Avatar: user.Avatar}
		if legacy {
			writeSessionCookie(c, accountCookiePrefix+id, cookie.Value, int(authSessionCookieTTL.Seconds()))
		}
	}
	items := make([]browserAccount, 0, len(accounts))
	for _, account := range accounts {
		items = append(items, account)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].SessionID < items[j].SessionID })
	response.OK(c, gin.H{"items": items})
}

func (h *AuthHandler) selectedSession(c *gin.Context, clientID, id string) (string, bool) {
	var raw string
	if id != "" {
		if validSessionID(id) {
			raw, _ = c.Cookie(accountCookiePrefix + id)
		}
	} else {
		// Existing integrations may omit a selector only for a single account.
		count := 0
		var onlyAccount string
		for _, cookie := range c.Request.Cookies() {
			if strings.HasPrefix(cookie.Name, accountCookiePrefix) {
				count++
				onlyAccount = cookie.Value
			}
		}
		if count > 1 {
			response.Error(c, http.StatusConflict, "AUTH_ACCOUNT_SELECTION_REQUIRED", "请选择要登录的账号")
			return "", false
		}
		raw, _ = readSessionCookie(c)
		if raw == "" && count == 1 {
			raw = onlyAccount
		}
	}
	user, err := h.sessions.SessionStatus(c.Request.Context(), authusecase.SessionStatusRequest{ClientID: clientID, SessionToken: raw})
	if err != nil {
		writeAuthError(c, err)
		return "", false
	}
	if id != "" && accountSessionID(user.UserID) != id {
		response.Error(c, http.StatusUnauthorized, authusecase.ErrInvalidRefreshToken, "账号会话已变化，请重新选择")
		return "", false
	}
	return raw, true
}

// BrowserLogout revokes only sessions supplied by the current browser. Other
// devices and other accounts remain independent of a single-account sign-out.
func (h *AuthHandler) BrowserLogout(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id"`
		All       bool   `json:"all"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (!req.All && !validSessionID(req.SessionID)) {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "请选择要退出的账号")
		return
	}
	selected, _ := c.Cookie(accountCookiePrefix + req.SessionID)
	for _, cookie := range c.Request.Cookies() {
		if cookie.Name != authSessionCookieName && !strings.HasPrefix(cookie.Name, accountCookiePrefix) {
			continue
		}
		if !req.All && cookie.Name != accountCookiePrefix+req.SessionID && (selected == "" || cookie.Value != selected) {
			continue
		}
		if err := h.sessions.Logout(c.Request.Context(), authusecase.LogoutRequest{RefreshToken: cookie.Value}); err != nil {
			var appErr *domain.Error
			if !errors.As(err, &appErr) || appErr.Code != authusecase.ErrInvalidRefreshToken {
				writeAuthError(c, err)
				return
			}
		}
		writeSessionCookie(c, cookie.Name, "", -1)
	}
	c.Header("Cache-Control", "no-store")
	response.OK(c, gin.H{"logged_out": true})
}
