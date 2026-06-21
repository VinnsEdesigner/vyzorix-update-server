// Package errors provides GraphQL-specific error handling.
package errors

import (
	"errors"
	"fmt"
)

// Standard GraphQL error codes
const (
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeBadRequest       = "BAD_REQUEST"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeValidation       = "VALIDATION_ERROR"
	CodeRateLimited      = "RATE_LIMITED"
	CodeAlreadyExists    = "ALREADY_EXISTS"
)

// Error represents a GraphQL-specific error with code and message.
type Error struct {
	Code      string
	Message   string
	Path      []interface{}
	Locations []Location
}

// Location represents a location in the GraphQL query.
type Location struct {
	Line   int
	Column int
}

// Implement the Error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new GraphQL error.
func New(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WithPath adds path information to the error.
func (e *Error) WithPath(path ...interface{}) *Error {
	e.Path = path
	return e
}

// WithLocation adds location information to the error.
func (e *Error) WithLocation(line, column int) *Error {
	e.Locations = []Location{{Line: line, Column: column}}
	return e
}

// Common errors
var (
	ErrUnauthorized = New(CodeUnauthorized, "authentication required")
	ErrForbidden    = New(CodeForbidden, "access denied")
	ErrNotFound     = New(CodeNotFound, "resource not found")
	ErrBadRequest   = New(CodeBadRequest, "invalid request")
	ErrInternal     = New(CodeInternalError, "internal server error")
)

// Wrap wraps a standard error into a GraphQL error with code.
func Wrap(err error, code string) error {
	if err == nil {
		return nil
	}
	var gqlErr *Error
	if errors.As(err, &gqlErr) {
		return err
	}
	return &Error{
		Code:    code,
		Message: err.Error(),
	}
}

// Is checks if the error matches the given code.
func Is(err error, code string) bool {
	var gqlErr *Error
	if errors.As(err, &gqlErr) {
		return gqlErr.Code == code
	}
	return false
}

// Helper functions for creating typed errors
func Unauthorized(format string, args ...interface{}) *Error {
	return New(CodeUnauthorized, fmt.Sprintf(format, args...))
}

func Forbidden(format string, args ...interface{}) *Error {
	return New(CodeForbidden, fmt.Sprintf(format, args...))
}

func NotFound(format string, args ...interface{}) *Error {
	return New(CodeNotFound, fmt.Sprintf(format, args...))
}

func BadRequest(format string, args ...interface{}) *Error {
	return New(CodeBadRequest, fmt.Sprintf(format, args...))
}

func Internal(format string, args ...interface{}) *Error {
	return New(CodeInternalError, fmt.Sprintf(format, args...))
}

func Validation(format string, args ...interface{}) *Error {
	return New(CodeValidation, fmt.Sprintf(format, args...))
}