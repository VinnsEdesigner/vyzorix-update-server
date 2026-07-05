package inbox

import (
	"errors"
	"net/http"
)

// Service errors for the inbox package.
var (
	ErrInboxNotFound            = errors.New("inbox entry not found")
	ErrInboxNotPending          = errors.New("inbox entry is not pending")
	ErrInboxNotAcknowledged     = errors.New("inbox entry is not acknowledged, device must acknowledge first")
	ErrInboxNotApproved         = errors.New("inbox entry is not approved, cannot resend notification")
	ErrInboxCannotBeAcknowledged = errors.New("inbox entry cannot be acknowledged, only pending entries can")
	ErrInboxCannotBeApproved    = errors.New("inbox entry cannot be approved, device must acknowledge first")
	ErrInboxCannotBeRejected    = errors.New("inbox entry cannot be rejected in current state")
	ErrInvalidIMEI              = errors.New("invalid IMEI format")
	ErrInvalidFCMToken          = errors.New("invalid FCM token format")
	ErrInvalidFirebaseInstallID = errors.New("invalid Firebase install ID format")
	ErrAlreadyExists            = errors.New("device already exists in inbox")
	ErrAlreadyRegistered        = errors.New("device already registered as confirmed device")
	ErrDeviceAlreadyExists      = errors.New("device already registered, use re-registration flow")
	ErrSecretGeneration         = errors.New("failed to generate command secret")
	ErrInvalidAckAction         = errors.New("invalid acknowledge action")
	ErrInvalidOperatorAction    = errors.New("invalid operator action")
	ErrFCMNotification          = errors.New("failed to send FCM notification")
	ErrUnauthorized             = errors.New("operator not authorized to perform this action")
)

// ServiceError represents an error with code and HTTP status.
type ServiceError struct {
	Details interface{} `json:"details,omitempty"`
	Code    string      `json:"error"`
	Message string      `json:"message"`
	Status  int         `json:"-"`
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

// ToServiceError converts domain errors to ServiceError with appropriate HTTP status.
func ToServiceError(err error) *ServiceError {
	switch {
	case errors.Is(err, ErrInboxNotFound):
		return &ServiceError{Code: "not_found", Message: err.Error(), Status: http.StatusNotFound}
	case errors.Is(err, ErrInboxNotPending):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrInboxNotAcknowledged):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrInboxNotApproved):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrInboxCannotBeAcknowledged):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrInboxCannotBeApproved):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrInboxCannotBeRejected):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrInvalidIMEI):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrInvalidFCMToken):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrInvalidFirebaseInstallID):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrAlreadyExists):
		return &ServiceError{Code: "conflict", Message: err.Error(), Status: http.StatusConflict}
	case errors.Is(err, ErrAlreadyRegistered), errors.Is(err, ErrDeviceAlreadyExists):
		return &ServiceError{Code: "conflict", Message: err.Error(), Status: http.StatusConflict}
	case errors.Is(err, ErrSecretGeneration):
		return &ServiceError{Code: "internal_error", Message: err.Error(), Status: http.StatusInternalServerError}
	case errors.Is(err, ErrInvalidAckAction):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrInvalidOperatorAction):
		return &ServiceError{Code: "bad_request", Message: err.Error(), Status: http.StatusBadRequest}
	case errors.Is(err, ErrFCMNotification):
		return &ServiceError{Code: "internal_error", Message: err.Error(), Status: http.StatusInternalServerError}
	case errors.Is(err, ErrUnauthorized):
		return &ServiceError{Code: "forbidden", Message: err.Error(), Status: http.StatusForbidden}
	default:
		return &ServiceError{Code: "internal_error", Message: "an unexpected error occurred", Status: http.StatusInternalServerError}
	}
}
