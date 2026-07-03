package inbox

import (
	"time"
)

// InboxEntry represents a device registration request in the inbox.
type InboxEntry struct {
	ID                string     `json:"id"`
	IMEI              string     `json:"imei"`
	Model             string     `json:"model"`
	Manufacturer      string     `json:"manufacturer"`
	OSVersion         string     `json:"osVersion"`
	AppVersion        string     `json:"appVersion"`
	FCMToken          string     `json:"fcmToken"`
	FirebaseInstallID string     `json:"firebaseInstallId"`
	Status            InboxStatus `json:"status"`
	CommandSecret     string     `json:"commandSecret,omitempty"`
	Notes             string     `json:"notes,omitempty"`
	OperatorID        string     `json:"operatorId,omitempty"`
	CreatedAt         int64      `json:"createdAt"`
	ApprovedAt        *int64     `json:"approvedAt,omitempty"`
	RejectedAt        *int64     `json:"rejectedAt,omitempty"`
}

// CreatedAtTime returns the CreatedAt as a time.Time.
func (e *InboxEntry) CreatedAtTime() time.Time {
	return time.UnixMilli(e.CreatedAt)
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

// IsPending returns true if the entry is pending approval.
func (e *InboxEntry) IsPending() bool {
	return e.Status == StatusPending
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

// RegistrationLog represents an audit log entry for registration actions.
type RegistrationLog struct {
	ID           string    `json:"id"`
	DeviceID     string    `json:"deviceId"`
	IMEI         string    `json:"imei"`
	Action       string    `json:"action"` // registered, approved, rejected, deregistered
	OperatorID   string    `json:"operatorId"`
	ClientIP     string    `json:"clientIp"`
	UserAgent    string    `json:"userAgent"`
	Details      string    `json:"details,omitempty"`
	Timestamp    int64     `json:"timestamp"`
}

// TimestampTime returns the Timestamp as a time.Time.
func (l *RegistrationLog) TimestampTime() time.Time {
	return time.UnixMilli(l.Timestamp)
}
