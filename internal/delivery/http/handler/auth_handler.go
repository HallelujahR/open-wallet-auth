package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/contextkey"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain/token"
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

	response.OK(c, gin.H{"logged_out": true})
}

// ChangePassword updates the authenticated user's password.
// ChangePassword 修改当前登录用户的密码。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	if err := h.auth.ChangePassword(c.Request.Context(), authusecase.ChangePasswordRequest{
		UserID:          authClaims.UserID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, gin.H{"password_changed": true})
}

// ResetPassword resets a password with a verified email code.
// ResetPassword 使用邮箱验证码重置用户密码。
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	if err := h.auth.ResetPassword(c.Request.Context(), authusecase.ResetPasswordRequest{
		Email:       req.Email,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	}); err != nil {
		writeAuthError(c, err)
		return
	}

	response.OK(c, gin.H{"password_reset": true})
}

// Me returns the authenticated user claims injected by auth middleware.
// Me 返回认证中间件注入的当前用户 claims。
func (h *AuthHandler) Me(c *gin.Context) {
	authClaims, ok := currentAuthClaims(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid authorization token")
		return
	}

	response.OK(c, dto.AuthUser{
		ID:       authClaims.UserID,
		Username: authClaims.Username,
		Email:    authClaims.Email,
	})
}

// currentAuthClaims reads token claims from Gin context.
// currentAuthClaims 从 Gin 上下文读取认证 claims。
func currentAuthClaims(c *gin.Context) (*token.Claims, bool) {
	claims, ok := c.Get(contextkey.AuthClaims)
	if !ok {
		return nil, false
	}

	authClaims, ok := claims.(*token.Claims)
	if !ok {
		return nil, false
	}
	return authClaims, true
}

// writeAuthError maps usecase/domain errors to stable HTTP responses.
// writeAuthError 将用例/领域错误映射为稳定的 HTTP 响应。
func writeAuthError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case authusecase.ErrEmailAlreadyExists:
			response.Error(c, http.StatusConflict, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidClient:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidCode:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidCredentials:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidRefreshToken:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		case authusecase.ErrRateLimited:
			response.Error(c, http.StatusTooManyRequests, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}

	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
