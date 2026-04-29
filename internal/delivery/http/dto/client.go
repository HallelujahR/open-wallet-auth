package dto

// CreateClientRequest is the HTTP request body for creating an application client.
type CreateClientRequest struct {
	ClientID            string   `json:"client_id" binding:"required"`
	Name                string   `json:"name" binding:"required"`
	JWTAudience         string   `json:"jwt_audience"`
	AllowedOrigins      []string `json:"allowed_origins"`
	AllowedRedirectURIs []string `json:"allowed_redirect_uris"`
}

// ClientResponse is the HTTP representation of an application client.
type ClientResponse struct {
	ID                  string   `json:"id"`
	ClientID            string   `json:"client_id"`
	Name                string   `json:"name"`
	JWTAudience         string   `json:"jwt_audience"`
	AllowedOrigins      []string `json:"allowed_origins"`
	AllowedRedirectURIs []string `json:"allowed_redirect_uris"`
	Status              string   `json:"status"`
	CreatedAt           string   `json:"created_at"`
}
