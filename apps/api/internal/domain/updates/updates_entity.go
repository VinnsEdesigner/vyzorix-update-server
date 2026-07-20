package updates

import (
	"time"
)

// ReleaseType represents the type of software release.
type ReleaseType string

const (
	ReleaseTypeMajor ReleaseType = "major"
	ReleaseTypeMinor ReleaseType = "minor"
	ReleaseTypePatch ReleaseType = "patch"
)

// UpdateStatus represents the status of an update push.
type UpdateStatus string

const (
	UpdateStatusPending    UpdateStatus = "pending"
	UpdateStatusInProgress UpdateStatus = "in_progress"
	UpdateStatusCompleted  UpdateStatus = "completed"
	UpdateStatusFailed     UpdateStatus = "failed"
	UpdateStatusCancelled  UpdateStatus = "cancelled"
)

// InstallType represents how an update should be installed.
type InstallType string

const (
	InstallTypeImmediate InstallType = "immediate"
	InstallTypeScheduled InstallType = "scheduled"
)

// DevicePushStatus represents the status of a device in an update push.
type DevicePushStatus string

const (
	DevicePushStatusPending      DevicePushStatus = "pending"
	DevicePushStatusSent         DevicePushStatus = "sent"
	DevicePushStatusInProgress  DevicePushStatus = "in_progress"
	DevicePushStatusAcknowledged DevicePushStatus = "acknowledged"
	DevicePushStatusCompleted   DevicePushStatus = "completed"
	DevicePushStatusFailed       DevicePushStatus = "failed"
)

// SyncStatus represents the status of a GitHub sync operation.
type SyncStatus string

const (
	SyncStatusIdle    SyncStatus = "idle"
	SyncStatusSyncing SyncStatus = "syncing"
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusError   SyncStatus = "error"
)

// UpdateVersion represents an available APK version.
type UpdateVersion struct {
	Version      string      `json:"version"`
	ReleaseType  ReleaseType `json:"releaseType"`
	APKFilename  string      `json:"apkFilename"`
	SHA256       string      `json:"sha256"`
	ReleaseNotes string      `json:"releaseNotes,omitempty"`
	ID           string      `json:"id"`
	APKSize      int64       `json:"apkSize"`
	ReleaseDate  int64       `json:"releaseDate"`
	IsLatest     bool        `json:"isLatest"`
	CreatedAt    int64       `json:"createdAt"`
	UpdatedAt    int64       `json:"updatedAt"`
}

// ReleaseDateTime returns the ReleaseDate as a time.Time.
func (v *UpdateVersion) ReleaseDateTime() time.Time {
	return time.UnixMilli(v.ReleaseDate)
}

// UpdatePush represents an update push operation.
type UpdatePush struct {
	ScheduledAt    *int64       `json:"scheduledAt,omitempty"`
	CompletedAt    *int64       `json:"completedAt,omitempty"`
	CancelledAt    *int64       `json:"cancelledAt,omitempty"`
	InstallType    InstallType  `json:"installType"`
	Status         UpdateStatus `json:"status"`
	InitiatedBy    string       `json:"initiatedBy"`
	CancelledBy    string       `json:"cancelledBy,omitempty"`
	ID             string       `json:"id"`
	VersionID      string       `json:"versionId"`
	OrganizationID string       `json:"organizationId"`
	InitiatedAt    int64        `json:"initiatedAt"`
}

// InitiatedAtTime returns the InitiatedAt as a time.Time.
func (p *UpdatePush) InitiatedAtTime() time.Time {
	return time.UnixMilli(p.InitiatedAt)
}

// IsActive returns true if the push is still active (pending or in_progress).
func (p *UpdatePush) IsActive() bool {
	return p.Status == UpdateStatusPending || p.Status == UpdateStatusInProgress
}

// CanCancel returns true if the push can be cancelled.
func (p *UpdatePush) CanCancel() bool {
	return p.Status == UpdateStatusPending || p.Status == UpdateStatusInProgress
}

// UpdatePushDevice represents a device targeted by an update push.
type UpdatePushDevice struct {
	ID             string           `json:"id"`
	PushID         string           `json:"pushId"`
	DeviceID       string           `json:"deviceId"`
	Status         DevicePushStatus `json:"status"`
	SentAt         *int64           `json:"sentAt,omitempty"`
	AcknowledgedAt *int64           `json:"acknowledgedAt,omitempty"`
	Error          string           `json:"error,omitempty"`
	RetryCount     int              `json:"retryCount"`
	CreatedAt      int64            `json:"createdAt"`
	UpdatedAt      int64            `json:"updatedAt"`
}

// IsTerminal returns true if the device push status is final.
func (d *UpdatePushDevice) IsTerminal() bool {
	return d.Status == DevicePushStatusCompleted || d.Status == DevicePushStatusFailed
}

// SyncState represents the current sync state.
type SyncState struct {
	Status        SyncStatus `json:"status"`
	LastSyncAt    *int64     `json:"lastSyncAt,omitempty"`
	NextSyncAt    *int64     `json:"nextSyncAt,omitempty"`
	Error         string     `json:"error,omitempty"`
	VersionsFound int        `json:"versionsFound,omitempty"`
}
