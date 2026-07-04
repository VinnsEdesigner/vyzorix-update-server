// Package handler provides HTTP handlers for GraphQL.
package handler

import (
	"errors"
	"time"

	gqlerrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/errors"
)

// ResponsePresenter handles standardized GraphQL response formatting.
type ResponsePresenter struct{}

// NewResponsePresenter creates a new response presenter.
func NewResponsePresenter() *ResponsePresenter {
	return &ResponsePresenter{}
}

// Success returns a successful response.
func (p *ResponsePresenter) Success(data interface{}) Response {
	return Response{
		Data:   data,
		Errors: nil,
	}
}

// Error returns an error response.
func (p *ResponsePresenter) Error(err error) Response {
	if err == nil {
		return Response{
			Data:   nil,
			Errors: nil,
		}
	}

	var gqlErr *gqlerrors.Error
	if errors.As(err, &gqlErr) {
		// Use the GraphQL error as-is
	} else {
		gqlErr = gqlerrors.Internal("%s", err.Error())
	}

	return Response{
		Data: nil,
		Errors: []ErrorDetail{
			{
				Message: gqlErr.Message,
				Code:    gqlErr.Code,
			},
		},
	}
}

// ErrorList returns an error response with multiple errors.
func (p *ResponsePresenter) ErrorList(errors []error) Response {
	details := make([]ErrorDetail, 0, len(errors))

	for _, err := range errors {
		if gqlErr, ok := err.(*gqlerrors.Error); ok {
			details = append(details, ErrorDetail{
				Message: gqlErr.Message,
				Code:    gqlErr.Code,
			})
		} else {
			details = append(details, ErrorDetail{
				Message: err.Error(),
				Code:    gqlerrors.CodeInternalError,
			})
		}
	}

	return Response{
		Data:   nil,
		Errors: details,
	}
}

// BadRequest creates a bad request error response.
func (p *ResponsePresenter) BadRequest(message string) Response {
	return p.Error(gqlerrors.BadRequest("%s", message))
}

// NotFound creates a not found error response.
func (p *ResponsePresenter) NotFound(message string) Response {
	return p.Error(gqlerrors.NotFound("%s", message))
}

// Unauthorized creates an unauthorized error response.
func (p *ResponsePresenter) Unauthorized() Response {
	return p.Error(gqlerrors.ErrUnauthorized)
}

// Forbidden creates a forbidden error response.
func (p *ResponsePresenter) Forbidden(message string) Response {
	return p.Error(gqlerrors.Forbidden("%s", message))
}

// Internal creates an internal error response.
func (p *ResponsePresenter) Internal(message string) Response {
	return p.Error(gqlerrors.Internal("%s", message))
}

// ValidationError creates a validation error response.
func (p *ResponsePresenter) ValidationError(message string) Response {
	return p.Error(gqlerrors.Validation("%s", message))
}

// ErrorDetail represents a single error in a response.
type ErrorDetail struct {
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Locations  []ErrorLocation        `json:"locations,omitempty"`
}

// ErrorLocation represents a location in the GraphQL query.
type ErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// RequestLog represents a request log entry.
type RequestLog struct {
	Timestamp  time.Time              `json:"timestamp"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	Operation  string                 `json:"operation,omitempty"`
	OperatorID string                 `json:"operatorId,omitempty"`
	ClientIP   string                 `json:"clientIp,omitempty"`
	UserAgent  string                 `json:"userAgent,omitempty"`
	Duration   time.Duration          `json:"duration,omitempty"`
	StatusCode int                    `json:"statusCode"`
	ErrorCount int                    `json:"errorCount"`
}

// Logger interface for request logging.
type Logger interface {
	LogRequest(log RequestLog)
}
