package domain

import "fmt"

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

func NewError(code string, message string) *Error {
	return &Error{Code: code, Message: message}
}

func WrapError(code string, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}
