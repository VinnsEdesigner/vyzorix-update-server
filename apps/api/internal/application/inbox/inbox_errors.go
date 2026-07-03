package inbox

import "errors"

// Service errors for the inbox package.
var (
	ErrInboxNotFound      = errors.New("inbox entry not found")
	ErrInboxNotPending    = errors.New("inbox entry is not pending, cannot acknowledge")
	ErrInvalidIMEI        = errors.New("invalid IMEI format")
	ErrInvalidFCMToken    = errors.New("invalid FCM token format")
	ErrAlreadyExists      = errors.New("device already exists in inbox")
	ErrAlreadyRegistered  = errors.New("device already registered")
	ErrSecretGeneration   = errors.New("failed to generate command secret")
	ErrInvalidAckAction   = errors.New("invalid acknowledge action, must be 'approve' or 'reject'")
	ErrFCMNotification    = errors.New("failed to send FCM notification")
)

// ServiceError represents an error with code and HTTP status.
type ServiceError struct {
	Code    string      `json:"error"`
	Message string      `json:"message"`
	Status  int         `json:"-"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *ServiceError) Error() string {
	return e.Message
}

// ToErrorResponse converts to ErrorResponse.
func (e *ServiceError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}
}

// AsServiceError converts an error to ServiceError.
func AsServiceError(err error) *ServiceError {
	var se *ServiceError
	if errors.As(err, &se) {
		return se
	}
	return nil
}
