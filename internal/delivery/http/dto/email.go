package dto

// EmailCodeRequest is the HTTP request body for creating an email verification code.
type EmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// EmailCodeResponse is returned after creating an email verification code.
type EmailCodeResponse struct {
	Email     string `json:"email"`
	ExpiresAt string `json:"expires_at"`
	DevCode   string `json:"dev_code,omitempty"`
}

// EmailVerifyRequest is the HTTP request body for verifying an email code.
type EmailVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

// EmailVerifyResponse is returned after verifying an email address.
type EmailVerifyResponse struct {
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}
