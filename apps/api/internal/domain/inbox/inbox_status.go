package inbox

// InboxStatus represents the status of an inbox entry.
type InboxStatus string

const (
	StatusPending  InboxStatus = "pending"
	StatusApproved InboxStatus = "approved"
	StatusRejected InboxStatus = "rejected"
)

// AckAction represents the action to take on an inbox entry.
type AckAction string

const (
	AckActionApprove AckAction = "approve"
	AckActionReject  AckAction = "reject"
)

// IsValid checks if the status is a valid inbox status.
func (s InboxStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusRejected:
		return true
	default:
		return false
	}
}

// IsValid checks if the action is a valid acknowledge action.
func (a AckAction) IsValid() bool {
	switch a {
	case AckActionApprove, AckActionReject:
		return true
	default:
		return false
	}
}
