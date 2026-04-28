package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	authusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/auth"
)

type AuthService interface {
	Register(c gin.Context, req authusecase.RegisterRequest) (*authusecase.RegisterResult, error)
	Login(c gin.Context, req authusecase.LoginRequest) (*authusecase.LoginResult, error)
}

type AuthHandler struct {
	auth *authusecase.Service
}

func NewAuthHandler(auth *authusecase.Service) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	result, err := h.auth.Register(c.Request.Context(), authusecase.RegisterRequest{
		ClientID: req.ClientID,
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, authusecase.ErrInvalidInput, "invalid request")
		return
	}

	result, err := h.auth.Login(c.Request.Context(), authusecase.LoginRequest{
		ClientID: req.ClientID,
		Email:    req.Email,
		Password: req.Password,
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

func writeAuthError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case authusecase.ErrEmailAlreadyExists:
			response.Error(c, http.StatusConflict, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidClient:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		case authusecase.ErrInvalidCredentials:
			response.Error(c, http.StatusUnauthorized, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}

	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
