package inbox

import (
	"time"
)

// InboxEntry represents a device registration request in the inbox.
// Implements the 5-state model: PENDING -> ACKNOWLEDGED -> APPROVING -> APPROVED -> REGISTERED.
type InboxEntry struct {
	AcknowledgedAt     *int64      `json:"acknowledgedAt,omitempty"`     // When device acknowledged
	ApprovingAt       *int64      `json:"approvingAt,omitempty"`        // When operator started approving
	ApprovedAt        *int64      `json:"approvedAt,omitempty"`         // When fully approved
	RejectedAt        *int64      `json:"rejectedAt,omitempty"`         // When rejected
	FCMToken          string      `json:"fcmToken"`
	FirebaseInstallID string      `json:"firebaseInstallId"`
	Model             string      `json:"model"`
	Manufacturer      string      `json:"manufacturer"`
	OSVersion         string      `json:"osVersion"`
	AppVersion        string      `json:"appVersion"`
	ID                string      `json:"id"`
	DeviceClass       string      `json:"deviceClass,omitempty"`
	Status            InboxStatus `json:"status"`
	CommandSecret     string      `json:"commandSecret,omitempty"`
	Notes             string      `json:"notes,omitempty"`
	OperatorID        string      `json:"operatorId,omitempty"`
	IMEI              string      `json:"imei"`
	DeviceName        string      `json:"deviceName,omitempty"`
	UpdatedAt         int64       `json:"updatedAt,omitempty"`
	CreatedAt         int64       `json:"createdAt"`
}

// CreatedAtTime returns the CreatedAt as a time.Time.
func (e *InboxEntry) CreatedAtTime() time.Time {
	return time.UnixMilli(e.CreatedAt)
}

// AcknowledgedAtTime returns the AcknowledgedAt as a time.Time if set.
func (e *InboxEntry) AcknowledgedAtTime() *time.Time {
	if e.AcknowledgedAt == nil {
		return nil
	}
	t := time.UnixMilli(*e.AcknowledgedAt)
	return &t
}

// ApprovingAtTime returns the ApprovingAt as a time.Time if set.
func (e *InboxEntry) ApprovingAtTime() *time.Time {
	if e.ApprovingAt == nil {
		return nil
	}
	t := time.UnixMilli(*e.ApprovingAt)
	return &t
}

// ApprovedAtTime returns the ApprovedAt as a time.Time if set.
func (e *InboxEntry) ApprovedAtTime() *time.Time {
	if e.ApprovedAt == nil {
		return nil
	}
	t := time.UnixMilli(*e.ApprovedAt)
	return &t
}

// RejectedAtTime returns the RejectedAt as a time.Time if set.
func (e *InboxEntry) RejectedAtTime() *time.Time {
	if e.RejectedAt == nil {
		return nil
	}
	t := time.UnixMilli(*e.RejectedAt)
	return &t
}

// IsPending returns true if the entry is pending.
func (e *InboxEntry) IsPending() bool {
	return e.Status == StatusPending
}

// IsAcknowledged returns true if the device has acknowledged.
func (e *InboxEntry) IsAcknowledged() bool {
	return e.Status == StatusAcknowledged
}

// IsApproving returns true if the operator is in the process of approving.
func (e *InboxEntry) IsApproving() bool {
	return e.Status == StatusApproving
}

// IsApproved returns true if the entry was approved.
func (e *InboxEntry) IsApproved() bool {
	return e.Status == StatusApproved
}

// IsRejected returns true if the entry was rejected.
func (e *InboxEntry) IsRejected() bool {
	return e.Status == StatusRejected
}

// CanBeAcknowledged returns true if the entry can be acknowledged (only pending entries).
func (e *InboxEntry) CanBeAcknowledged() bool {
	return e.Status == StatusPending
}

// CanBeApproved returns true if the entry can be approved (only acknowledged entries).
func (e *InboxEntry) CanBeApproved() bool {
	return e.Status == StatusAcknowledged
}

// CanBeRejected returns true if the entry can be rejected (pending or acknowledged).
func (e *InboxEntry) CanBeRejected() bool {
	return e.Status == StatusPending || e.Status == StatusAcknowledged
}

// RegistrationLog represents an audit log entry for registration actions.
type RegistrationLog struct {
	ID         string `json:"id"`
	DeviceID   string `json:"deviceId"`
	IMEI       string `json:"imei"`
	Action     string `json:"action"` // registered, approved, rejected, deregistered
	OperatorID string `json:"operatorId"`
	ClientIP   string `json:"clientIp"`
	UserAgent  string `json:"userAgent"`
	Details    string `json:"details,omitempty"`
	Timestamp  int64  `json:"timestamp"`
}

// TimestampTime returns the Timestamp as a time.Time.
func (l *RegistrationLog) TimestampTime() time.Time {
	return time.UnixMilli(l.Timestamp)
}
