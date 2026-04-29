package domain

import "fmt"

// Error is a domain-level error with a stable machine-readable code.
// Error 是领域层统一错误，向上暴露稳定的机器可读错误码。
type Error struct {
	Code    string
	Message string
	Err     error
}

// Error returns the user-facing error message and preserves wrapped details for logs.
// Error 返回面向调用方的错误信息，同时在包装错误时保留底层错误细节。
func (e *Error) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// Unwrap exposes the underlying cause for errors.Is/errors.As matching.
// Unwrap 暴露底层错误，便于 errors.Is/errors.As 做错误判断。
func (e *Error) Unwrap() error {
	return e.Err
}

// NewError creates a domain error without wrapping an underlying cause.
// NewError 创建不包装底层原因的领域错误。
func NewError(code string, message string) *Error {
	return &Error{Code: code, Message: message}
}

// WrapError creates a domain error that preserves the underlying cause.
// WrapError 创建带底层原因的领域错误，适合保留基础设施错误上下文。
func WrapError(code string, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}
