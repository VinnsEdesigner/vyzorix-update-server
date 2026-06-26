package updates

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Details interface{} `json:"details,omitempty"`
	Message string      `json:"message"`
	Code    string      `json:"error"`
}

// ServiceError represents a service-level error.
type ServiceError struct {
	Details interface{}
	Code    string
	Message string
	Status  int
}

func (e *ServiceError) Error() string {
	return e.Message
}

// Common service errors.
var (
	ErrVersionNotFound       = &ServiceError{Code: "version_not_found", Message: "APK version not found", Status: 400}
	ErrPushNotFound          = &ServiceError{Code: "push_not_found", Message: "Update push not found", Status: 404}
	ErrPushNotCancellable    = &ServiceError{Code: "push_not_cancellable", Message: "Push cannot be cancelled", Status: 400}
	ErrSyncAlreadyInProgress = &ServiceError{Code: "sync_already_in_progress", Message: "Sync already in progress", Status: 409}
	ErrBadRequest            = &ServiceError{Code: "bad_request", Message: "Invalid request", Status: 400}
	ErrInternalServer        = &ServiceError{Code: "internal_error", Message: "Internal server error", Status: 500}
)

// IsTerminalError returns true if the error is a terminal/service error.
func IsTerminalError(err error) bool {
	_, ok := err.(*ServiceError)
	return ok
}

// AsServiceError converts an error to a ServiceError if possible.
func AsServiceError(err error) *ServiceError {
	if se, ok := err.(*ServiceError); ok {
		return se
	}
	return nil
}

// ToErrorResponse converts a ServiceError to an ErrorResponse.
func (e *ServiceError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}
}
