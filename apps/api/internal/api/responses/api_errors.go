package responses

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Common error codes - standardized across the API.
const (
	ErrCodeBadRequest    = "bad_request"
	ErrCodeUnauthorized  = "unauthorized"
	ErrCodeForbidden     = "forbidden"
	ErrCodeNotFound      = "not_found"
	ErrCodeConflict      = "conflict"
	ErrCodeInternalError = "internal_error"
	ErrCodeRateLimit     = "rate_limit_exceeded"
	ErrCodeInvalidInput  = "invalid_input"
)

// ErrorMessages maps error codes to safe user messages.
var ErrorMessages = map[string]string{
	ErrCodeBadRequest:    "The request was invalid",
	ErrCodeUnauthorized:  "Authentication required",
	ErrCodeForbidden:     "Access denied",
	ErrCodeNotFound:      "Resource not found",
	ErrCodeConflict:      "Resource already exists",
	ErrCodeInternalError: "An unexpected error occurred",
	ErrCodeRateLimit:     "Too many requests, please try again later",
	ErrCodeInvalidInput:  "Invalid input provided",
}

// APIError represents a structured API error response.
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// RespondWithError sends a standardized error response.
// Logs the actual error internally but returns only safe info to client.
func RespondWithError(c *gin.Context, statusCode int, code string, err error) {
	if err != nil {
		// Log actual error internally - not exposed to client.
		slog.Warn("api_error",
			slog.String("path", c.Request.URL.Path),
			slog.String("error", err.Error()),
		)
	}

	message := ErrorMessages[code]
	if message == "" {
		message = "An error occurred"
	}

	c.JSON(statusCode, APIError{
		Error:   code,
		Message: message,
	})
}

// RespondWithMessage sends a standardized error with a custom (safe) message.
func RespondWithMessage(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, APIError{
		Error:   code,
		Message: message,
	})
}

// BadRequest sends 400 Bad Request.
func BadRequest(c *gin.Context, message string) {
	RespondWithMessage(c, http.StatusBadRequest, ErrCodeBadRequest, message)
}

// Unauthorized sends 401 Unauthorized.
func Unauthorized(c *gin.Context) {
	RespondWithMessage(c, http.StatusUnauthorized, ErrCodeUnauthorized, "Authentication required")
}

// Forbidden sends 403 Forbidden.
func Forbidden(c *gin.Context) {
	RespondWithMessage(c, http.StatusForbidden, ErrCodeForbidden, "Access denied")
}

// NotFound sends 404 Not Found.
func NotFound(c *gin.Context) {
	RespondWithMessage(c, http.StatusNotFound, ErrCodeNotFound, "Resource not found")
}

// Conflict sends 409 Conflict.
func Conflict(c *gin.Context) {
	RespondWithMessage(c, http.StatusConflict, ErrCodeConflict, "Resource already exists")
}

// TooManyRequests sends 429 Too Many Requests.
func TooManyRequests(c *gin.Context) {
	RespondWithMessage(c, http.StatusTooManyRequests, ErrCodeRateLimit, "Too many requests, please try again later")
}

// InternalError sends 500 Internal Server Error.
func InternalError(c *gin.Context) {
	RespondWithMessage(c, http.StatusInternalServerError, ErrCodeInternalError, "An unexpected error occurred")
}

// GetErrorCode returns a safe error code string based on HTTP status.
func GetErrorCode(status int) string {
	switch {
	case status >= 500:
		return ErrCodeInternalError
	case status == http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case status == http.StatusForbidden:
		return ErrCodeForbidden
	case status == http.StatusNotFound:
		return ErrCodeNotFound
	case status == http.StatusBadRequest:
		return ErrCodeBadRequest
	case status == http.StatusTooManyRequests:
		return ErrCodeRateLimit
	case status >= 400:
		return ErrCodeBadRequest
	default:
		return ErrCodeInternalError
	}
}

// GetErrorMessage returns a safe, user-friendly error message.
func GetErrorMessage(status int) string {
	switch {
	case status >= 500:
		return ErrorMessages[ErrCodeInternalError]
	case status == http.StatusUnauthorized:
		return ErrorMessages[ErrCodeUnauthorized]
	case status == http.StatusForbidden:
		return ErrorMessages[ErrCodeForbidden]
	case status == http.StatusNotFound:
		return ErrorMessages[ErrCodeNotFound]
	case status == http.StatusBadRequest:
		return ErrorMessages[ErrCodeBadRequest]
	case status == http.StatusTooManyRequests:
		return ErrorMessages[ErrCodeRateLimit]
	case status >= 400:
		return "Request could not be completed"
	default:
		return "An error occurred"
	}
}
