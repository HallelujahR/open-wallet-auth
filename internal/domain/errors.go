package domain

import "fmt"

// Error is a domain-level error with a stable machine-readable code.
type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// NewError creates a domain error without wrapping an underlying cause.
func NewError(code string, message string) *Error {
	return &Error{Code: code, Message: message}
}

// WrapError creates a domain error that preserves the underlying cause.
func WrapError(code string, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}
