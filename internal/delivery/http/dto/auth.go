package dto

// RegisterRequest is the HTTP request body for registration.
type RegisterRequest struct {
	ClientID string `json:"client_id"`
	Username string `json:"username"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the HTTP request body for password login.
type LoginRequest struct {
	ClientID string `json:"client_id"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest is the HTTP request body for token rotation.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest is the HTTP request body for logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthUser is the user payload returned by auth endpoints.
type AuthUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// TokenPair is the token payload returned by auth endpoints.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

// AuthResponse is the successful response payload for auth endpoints.
type AuthResponse struct {
	User  AuthUser  `json:"user"`
	Token TokenPair `json:"token"`
}
