package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	"github.com/open-wallet-auth/open-wallet-auth/internal/usecase/clientaccess"
	phoneusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/phone"
)

// PhoneHandler exposes phone-code authentication endpoints.
// PhoneHandler 暴露手机号验证码认证接口。
type PhoneHandler struct {
	phone *phoneusecase.Service
}

// NewPhoneHandler creates a PhoneHandler bound to the phone usecase service.
// NewPhoneHandler 创建绑定手机号登录用例服务的 HTTP handler。
func NewPhoneHandler(phone *phoneusecase.Service) *PhoneHandler {
	return &PhoneHandler{phone: phone}
}

// Code creates a short-lived phone verification code.
// Code 创建短期有效的手机号验证码。
func (h *PhoneHandler) Code(c *gin.Context) {
	var req dto.PhoneCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, phoneusecase.ErrInvalidInput, "invalid request")
		return
	}
	result, err := h.phone.RequestCode(c.Request.Context(), phoneusecase.CodeRequest{
		ClientID: req.ClientID,
		Phone:    req.Phone,
	})
	if err != nil {
		writePhoneError(c, err)
		return
	}
	response.OK(c, dto.PhoneCodeResponse{
		Phone:     result.Phone,
		ExpiresAt: result.ExpiresAt.Format(timeFormatRFC3339),
		DevCode:   result.DevCode,
	})
}

// Login verifies a phone code and returns a token pair.
// Login 校验手机号验证码并返回 token 组合。
func (h *PhoneHandler) Login(c *gin.Context) {
	var req dto.PhoneLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, phoneusecase.ErrInvalidInput, "invalid request")
		return
	}
	result, err := h.phone.Login(c.Request.Context(), phoneusecase.LoginRequest{
		ClientID:  req.ClientID,
		Phone:     req.Phone,
		Code:      req.Code,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		writePhoneError(c, err)
		return
	}
	setSessionCookie(c, result.Token.RefreshToken)
	response.OK(c, dto.PhoneAuthResponse{
		User: dto.PhoneAuthUser{
			ID:       result.UserID,
			Username: result.Username,
			Phone:    result.Phone,
		},
		Token: dto.TokenPair{
			AccessToken:  result.Token.AccessToken,
			RefreshToken: result.Token.RefreshToken,
			ExpiresAt:    result.Token.ExpiresAt.Format(timeFormatRFC3339),
			TokenType:    "Bearer",
		},
	})
}

// writePhoneError maps phone usecase errors to HTTP responses.
// writePhoneError 将手机号登录用例错误映射为 HTTP 响应。
func writePhoneError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case phoneusecase.ErrInvalidCode:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		case clientaccess.ErrAccessDenied:
			response.Error(c, http.StatusForbidden, appErr.Code, appErr.Message)
		case phoneusecase.ErrRateLimited:
			response.Error(c, http.StatusTooManyRequests, appErr.Code, appErr.Message)
		case phoneusecase.ErrSendFailed:
			response.Error(c, http.StatusServiceUnavailable, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
