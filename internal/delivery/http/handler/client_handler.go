package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/dto"
	"github.com/open-wallet-auth/open-wallet-auth/internal/delivery/http/response"
	"github.com/open-wallet-auth/open-wallet-auth/internal/domain"
	clientdomain "github.com/open-wallet-auth/open-wallet-auth/internal/domain/client"
	clientusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/client"
	settingsusecase "github.com/open-wallet-auth/open-wallet-auth/internal/usecase/settings"
)

// ClientHandler exposes application client management endpoints.
// ClientHandler 暴露业务系统 client 管理接口。
type ClientHandler struct {
	clients  *clientusecase.Service
	settings *settingsusecase.Service
}

// NewClientHandler creates a client management handler.
// NewClientHandler 创建 client 管理 HTTP handler。
func NewClientHandler(clients *clientusecase.Service, settings *settingsusecase.Service) *ClientHandler {
	return &ClientHandler{clients: clients, settings: settings}
}

// Create registers a new application client.
// Create 注册新的业务系统 client。
func (h *ClientHandler) Create(c *gin.Context) {
	var req dto.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, clientusecase.ErrInvalidClientInput, "invalid request")
		return
	}

	client, err := h.clients.Create(c.Request.Context(), clientusecase.CreateRequest{
		ClientID:            req.ClientID,
		Name:                req.Name,
		JWTAudience:         req.JWTAudience,
		AllowedOrigins:      req.AllowedOrigins,
		AllowedRedirectURIs: req.AllowedRedirectURIs,
	})
	if err != nil {
		writeClientError(c, err)
		return
	}

	response.OK(c, toClientResponse(*client))
}

// List returns all configured application clients.
// List 返回所有已配置的业务系统 client。
func (h *ClientHandler) List(c *gin.Context) {
	clients, err := h.clients.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	data := make([]dto.ClientResponse, 0, len(clients))
	for _, client := range clients {
		data = append(data, toClientResponse(client))
	}
	response.OK(c, data)
}

// LoginConfig returns public login-page configuration for one application.
// LoginConfig 返回某个业务系统的统一登录页公开配置。
func (h *ClientHandler) LoginConfig(c *gin.Context) {
	client, err := h.clients.GetByClientID(c.Request.Context(), c.Query("client_id"))
	if err != nil {
		writeClientError(c, err)
		return
	}
	login, err := h.settings.LoginSettings(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	response.OK(c, dto.LoginConfigResponse{
		Client: dto.PublicClientResponse{
			ClientID: client.ClientID,
			Name:     client.Name,
		},
		Login: dto.LoginPageSettingsResponse{
			BrandName:      login.BrandName,
			BrandMark:      login.BrandMark,
			Subtitle:       login.Subtitle,
			EnableRegister: login.EnableRegister,
			EnablePhone:    login.EnablePhone,
			EnableGitHub:   login.EnableGitHub,
			EnableGoogle:   login.EnableGoogle,
			EnableWallet:   login.EnableWallet,
		},
	})
}

// writeClientError maps client usecase errors to HTTP responses.
// writeClientError 将 client 用例错误映射为 HTTP 响应。
func writeClientError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case clientusecase.ErrClientAlreadyExists:
			response.Error(c, http.StatusConflict, appErr.Code, appErr.Message)
		case clientusecase.ErrClientNotFound:
			response.Error(c, http.StatusNotFound, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}

// toClientResponse converts a domain client to its HTTP DTO.
// toClientResponse 将领域 client 转换为 HTTP 响应 DTO。
func toClientResponse(client clientdomain.Client) dto.ClientResponse {
	return dto.ClientResponse{
		ID:                  client.ID,
		ClientID:            client.ClientID,
		Name:                client.Name,
		JWTAudience:         client.JWTAudience,
		AllowedOrigins:      client.AllowedOrigins,
		AllowedRedirectURIs: client.AllowedRedirectURIs,
		Status:              string(client.Status),
		CreatedAt:           client.CreatedAt.Format(timeFormatRFC3339),
	}
}
