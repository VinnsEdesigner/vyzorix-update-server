
package errors

import (
	stderrors "errors"
	"fmt"
)

// ValidationError is the structured validation failure returned by request
// validators. It carries field-level details that the error/response layers
// surface to clients as the `details` of a VALIDATION_FAILED ServerError.
type ValidationError struct {
	Details []ValidationDetail
}

// NewValidationError wraps a set of field-level validation details into a
// ValidationError. Callers pass nil/empty to denote "no details" — though in
// practice validators only construct this when there are details.
func NewValidationError(details []ValidationDetail) *ValidationError {
	return &ValidationError{Details: details}
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e == nil || len(e.Details) == 0 {
		return "validation failed"
	}
	if len(e.Details) == 1 {
		d := e.Details[0]
		return fmt.Sprintf("validation failed: %s: %s", d.Field, d.Message)
	}
	return fmt.Sprintf("validation failed: %d field error(s)", len(e.Details))
}

// ValidationDetailsOf extracts field-level validation details from an error,
// returning the details and true when the error is (or wraps) a
// *ValidationError. Used by the error middleware to render structured 400s.
func ValidationDetailsOf(err error) ([]ValidationDetail, bool) {
	var ve *ValidationError
	if !stderrors.As(err, &ve) {
		return nil, false
	}
	if ve == nil {
		return nil, true
	}
	return ve.Details, true
}

// AsServerError extracts a *ServerError from an error (direct or wrapped),
// returning nil when the error is not a ServerError. Used by the error
// middleware to render a ServerError's own code/message/details/docs-url.
func AsServerError(err error) *ServerError {
	var se *ServerError
	if !stderrors.As(err, &se) {
		return nil
	}
	return se
}