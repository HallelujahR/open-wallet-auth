package auth

const (
	ErrEmailAlreadyExists  = "AUTH_EMAIL_ALREADY_EXISTS"
	ErrEmailAlreadyBound   = "AUTH_EMAIL_ALREADY_BOUND"
	ErrPhoneAlreadyBound   = "AUTH_PHONE_ALREADY_BOUND"
	ErrInvalidClient       = "CLIENT_INVALID"
	ErrInvalidCode         = "AUTH_INVALID_CODE"
	ErrInvalidCredentials  = "AUTH_INVALID_CREDENTIALS"
	ErrInvalidInput        = "AUTH_INVALID_INPUT"
	ErrInvalidRefreshToken = "AUTH_INVALID_REFRESH_TOKEN"
	ErrBindingNotFound     = "AUTH_BINDING_NOT_FOUND"
	ErrLastLoginMethod     = "AUTH_LAST_LOGIN_METHOD"
	ErrRateLimited         = "AUTH_RATE_LIMITED"
)
