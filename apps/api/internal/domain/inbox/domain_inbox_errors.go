package inbox

import "errors"

// Domain errors for the inbox package.
var (
	ErrInboxNotFound               = errors.New("inbox entry not found")
	ErrInboxNotPending             = errors.New("inbox entry is not pending")
	ErrInvalidIMEI                 = errors.New("invalid IMEI format")
	ErrInvalidFCMToken             = errors.New("invalid FCM token format")
	ErrAlreadyExists               = errors.New("inbox entry already exists for this IMEI")
	ErrAlreadyRegistered           = errors.New("device already registered")
	ErrSecretGeneration            = errors.New("failed to generate command secret")
	ErrInvalidAckAction            = errors.New("invalid acknowledge action")
	ErrInboxCannotBeAcknowledged    = errors.New("inbox entry cannot be acknowledged")
	ErrInboxCannotBeApproved       = errors.New("inbox entry cannot be approved")
	ErrInboxCannotBeRejected       = errors.New("inbox entry cannot be rejected")
)
