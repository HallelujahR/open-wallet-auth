package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	emailusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/email"
)

// EmailHandler exposes email verification endpoints.
// EmailHandler 暴露邮箱验证码接口。
type EmailHandler struct {
	email *emailusecase.Service
}

// NewEmailHandler creates an EmailHandler bound to the email usecase service.
// NewEmailHandler 创建绑定邮箱验证用例服务的 HTTP handler。
func NewEmailHandler(email *emailusecase.Service) *EmailHandler {
	return &EmailHandler{email: email}
}

// Code creates and sends an email verification code.
// Code 创建并发送邮箱验证码。
func (h *EmailHandler) Code(c *gin.Context) {
	var req dto.EmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, emailusecase.ErrInvalidInput, "invalid request")
		return
	}
	result, err := h.email.RequestCode(c.Request.Context(), emailusecase.CodeRequest{Email: req.Email})
	if err != nil {
		writeEmailError(c, err)
		return
	}
	response.OK(c, dto.EmailCodeResponse{
		Email:     result.Email,
		ExpiresAt: result.ExpiresAt.Format(timeFormatRFC3339),
		DevCode:   result.DevCode,
	})
}

// Verify checks an email verification code.
// Verify 校验邮箱验证码。
func (h *EmailHandler) Verify(c *gin.Context) {
	var req dto.EmailVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, emailusecase.ErrInvalidInput, "invalid request")
		return
	}
	result, err := h.email.VerifyCode(c.Request.Context(), emailusecase.VerifyRequest{Email: req.Email, Code: req.Code})
	if err != nil {
		writeEmailError(c, err)
		return
	}
	response.OK(c, dto.EmailVerifyResponse{Email: result.Email, Verified: result.Verified})
}

// writeEmailError maps email usecase errors to HTTP responses.
// writeEmailError 将邮箱验证用例错误映射为 HTTP 响应。
func writeEmailError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case emailusecase.ErrInvalidCode:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		case emailusecase.ErrRateLimited:
			response.Error(c, http.StatusTooManyRequests, appErr.Code, appErr.Message)
		case emailusecase.ErrSendFailed:
			response.Error(c, http.StatusServiceUnavailable, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
