package dto

// PhoneCodeRequest is the HTTP request body for creating a phone login code.
type PhoneCodeRequest struct {
	ClientID string `json:"client_id"`
	Phone    string `json:"phone" binding:"required"`
}

// PhoneCodeResponse is returned after creating a phone login code.
type PhoneCodeResponse struct {
	Phone     string `json:"phone"`
	ExpiresAt string `json:"expires_at"`
	DevCode   string `json:"dev_code,omitempty"`
}

// PhoneLoginRequest is the HTTP request body for phone-code login.
type PhoneLoginRequest struct {
	ClientID string `json:"client_id"`
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// PhoneAuthUser is the user payload returned by phone endpoints.
type PhoneAuthUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

// PhoneAuthResponse is returned after a successful phone login.
type PhoneAuthResponse struct {
	User  PhoneAuthUser `json:"user"`
	Token TokenPair     `json:"token"`
}
