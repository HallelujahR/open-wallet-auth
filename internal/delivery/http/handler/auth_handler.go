package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	authusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
)

// AuthHandler exposes authentication usecases as HTTP endpoints.
// AuthHandler 将认证用例暴露为 HTTP 接口。
type AuthHandler struct {
	auth *authusecase.Service
}

// NewAuthHandler creates an AuthHandler bound to the auth usecase service.
// NewAuthHandler 创建绑定认证用例服务的 HTTP handler。
func NewAuthHandler(auth *authusecase.Service) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register handles email/password registration and returns a token pair.
// Register 处理邮箱密码注册请求，并返回 token 组合。
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	result, err := h.auth.Register(c.Request.Context(), authusecase.RegisterRequest{
		ClientID:  req.ClientID,
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	setSessionCookie(c, result.Token.RefreshToken)
	response.OK(c, dto.AuthResponse{
		User: dto.AuthUser{
			ID:       result.UserID,
			Username: result.Username,
			Email:    result.Email,
		},
		Token: dto.TokenPair{
			AccessToken:  result.Token.AccessToken,
			RefreshToken: result.Token.RefreshToken,
			ExpiresAt:    result.Token.ExpiresAt.Format(timeFormatRFC3339),
			TokenType:    "Bearer",
		},
	})
}

// Login handles email/password authentication and returns a token pair.
// Login 处理邮箱密码登录请求，并返回 token 组合。
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	result, err := h.auth.Login(c.Request.Context(), authusecase.LoginRequest{
		ClientID:  req.ClientID,
		Email:     req.Email,
		Password:  req.Password,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	setSessionCookie(c, result.Token.RefreshToken)
	response.OK(c, dto.AuthResponse{
		User: dto.AuthUser{
			ID:       result.UserID,
			Username: result.Username,
			Email:    result.Email,
		},
		Token: dto.TokenPair{
			AccessToken:  result.Token.AccessToken,
			RefreshToken: result.Token.RefreshToken,
			ExpiresAt:    result.Token.ExpiresAt.Format(timeFormatRFC3339),
			TokenType:    "Bearer",
		},
	})
}

// Refresh rotates a refresh token and returns a new token pair.
// Refresh 轮换刷新令牌，并返回新的 token 组合。
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	result, err := h.auth.Refresh(c.Request.Context(), authusecase.RefreshRequest{
		RefreshToken: req.RefreshToken,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	setSessionCookie(c, result.Token.RefreshToken)
	response.OK(c, dto.AuthResponse{
		User: dto.AuthUser{
			ID:       result.UserID,
			Username: result.Username,
			Email:    result.Email,
		},
		Token: dto.TokenPair{
			AccessToken:  result.Token.AccessToken,
			RefreshToken: result.Token.RefreshToken,
			ExpiresAt:    result.Token.ExpiresAt.Format(timeFormatRFC3339),
			TokenType:    "Bearer",
		},
	})
}

// Logout revokes a refresh token.
// Logout 吊销刷新令牌。
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	if err := h.auth.Logout(c.Request.Context(), authusecase.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		writeAuthError(c, err)
		return
	}

	clearSessionCookie(c)
	response.OK(c, gin.H{"logged_out": true})
}

// Session reports the current central auth session user if the browser is signed in.
// Session 返回当前浏览器中台会话对应的用户，用于统一登录页展示“一键登录”。
func (h *AuthHandler) Session(c *gin.Context) {
	sessionToken, ok := readSessionCookie(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, authusecase.ErrInvalidRefreshToken, "not signed in")
		return
	}

	result, err := h.auth.SessionStatus(c.Request.Context(), authusecase.SessionStatusRequest{
		ClientID:     c.Query("client_id"),
		SessionToken: sessionToken,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, dto.AuthUser{
		ID:       result.UserID,
		Username: result.Username,
		Email:    result.Email,
	})
}

// SessionLogin exchanges the central auth session for a token pair for one client.
// SessionLogin 使用中台会话为某个业务系统签发 token，实现跨系统一键登录。
func (h *AuthHandler) SessionLogin(c *gin.Context) {
	var req dto.SessionLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}
	sessionToken, ok := readSessionCookie(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, authusecase.ErrInvalidRefreshToken, "not signed in")
		return
	}

	result, err := h.auth.LoginWithSession(c.Request.Context(), authusecase.SessionLoginRequest{
		ClientID:     req.ClientID,
		SessionToken: sessionToken,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	setSessionCookie(c, result.Token.RefreshToken)
	response.OK(c, dto.AuthResponse{
		User: dto.AuthUser{
			ID:       result.UserID,
			Username: result.Username,
			Email:    result.Email,
		},
		Token: dto.TokenPair{
			AccessToken:  result.Token.AccessToken,
			RefreshToken: result.Token.RefreshToken,
			ExpiresAt:    result.Token.ExpiresAt.Format(timeFormatRFC3339),
			TokenType:    "Bearer",
		},
	})
}
