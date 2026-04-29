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
)

// ClientHandler exposes application client management endpoints.
type ClientHandler struct {
	clients *clientusecase.Service
}

// NewClientHandler creates a client management handler.
func NewClientHandler(clients *clientusecase.Service) *ClientHandler {
	return &ClientHandler{clients: clients}
}

// Create registers a new application client.
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

func writeClientError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case clientusecase.ErrClientAlreadyExists:
			response.Error(c, http.StatusConflict, appErr.Code, appErr.Message)
		default:
			response.Error(c, http.StatusBadRequest, appErr.Code, appErr.Message)
		}
		return
	}
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}

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
