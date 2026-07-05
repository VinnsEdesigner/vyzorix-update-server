package updates

import "time"

// Pagination represents pagination information.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// PaginationResponse represents pagination in responses.
type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// VersionResponse represents a version in API responses.
type VersionResponse struct {
	Version      string `json:"version"`
	ReleaseType  string `json:"releaseType"`
	Status       string `json:"status"`
	APKFilename  string `json:"apkFilename"`
	SHA256       string `json:"sha256"`
	ReleaseNotes string `json:"releaseNotes,omitempty"`
	APKSize      int64  `json:"apkSize"`
	ReleasedAt   int64  `json:"releasedAt"`
	IsLatest     bool   `json:"isLatest"`
}

// ListVersionsResponse represents the response for GET /v1/updates/versions.
type ListVersionsResponse struct {
	Versions   []VersionResponse `json:"versions"`
	Pagination Pagination        `json:"pagination"`
}

// ChangelogEntry represents a changelog entry.
type ChangelogEntry struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Type    string `json:"type"`
	Notes   string `json:"notes"`
}

// GetChangelogResponse represents the response for GET /v1/updates/changelog.
type GetChangelogResponse struct {
	Changelog []ChangelogEntry `json:"changelog"`
}

// SyncStatusInfo represents sync status information.
type SyncStatusInfo struct {
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	LastSyncAt int64  `json:"lastSyncAt,omitempty"`
	NextSyncAt int64  `json:"nextSyncAt,omitempty"`
}

// LatestVersionInfo represents the latest version information.
type LatestVersionInfo struct {
	Version     string `json:"version"`
	ReleaseType string `json:"releaseType"`
	APKFilename string `json:"apkFilename"`
	SHA256      string `json:"sha256"`
	ReleasedAt  int64  `json:"releasedAt"`
	APKSize     int64  `json:"apkSize"`
}

// DeviceStatusInfo represents device status in GetStatus response.
type DeviceStatusInfo struct {
	CurrentVersion string `json:"currentVersion,omitempty"`
	NeedsUpdate    bool   `json:"needsUpdate"`
}

// GetStatusResponse represents the response for GET /v1/updates/status.
type GetStatusResponse struct {
	Latest LatestVersionInfo `json:"latest,omitempty"`
	Device *DeviceStatusInfo `json:"device,omitempty"`
	Sync   SyncStatusInfo    `json:"sync"`
}

// PushDeviceCounts represents device counts in push responses.
type PushDeviceCounts struct {
	Total        int `json:"total"`
	Pending      int `json:"pending"`
	Sent         int `json:"sent"`
	Acknowledged int `json:"acknowledged"`
	Failed       int `json:"failed"`
}

// PushUpdateResponse represents the response for POST /v1/updates/push.
type PushUpdateResponse struct {
	PushID        string           `json:"pushId"`
	Version       string           `json:"version"`
	InstallType   string           `json:"installType"`
	ScheduledAt   *int64           `json:"scheduledAt,omitempty"`
	InitiatedBy   string           `json:"initiatedBy"`
	Status        string           `json:"status"`
	DeviceIDs     []string         `json:"deviceIds"`
	FailedDevices []FailedDevice   `json:"failedDevices,omitempty"`
	Devices       PushDeviceCounts `json:"devices"`
	InitiatedAt   int64            `json:"initiatedAt"`
}

// FailedDevice represents a device that failed to be added to a push.
type FailedDevice struct {
	DeviceID string `json:"deviceId"`
	Reason   string `json:"reason"`
}

// HistoryDeviceCounts represents device counts in history.
type HistoryDeviceCounts struct {
	Pending      int `json:"pending,omitempty"`
	Sent         int `json:"sent,omitempty"`
	Acknowledged int `json:"acknowledged,omitempty"`
	Failed       int `json:"failed,omitempty"`
}

// PushHistoryEntry represents a push entry in history.
type PushHistoryEntry struct {
	CompletedAt *int64              `json:"completedAt,omitempty"`
	CancelledAt *int64              `json:"cancelledAt,omitempty"`
	ScheduledAt *int64              `json:"scheduledAt,omitempty"`
	ID          string              `json:"id"`
	Version     string              `json:"version"`
	InstallType string              `json:"installType"`
	Status      string              `json:"status"`
	InitiatedBy string              `json:"initiatedBy"`
	Devices     HistoryDeviceCounts `json:"devices"`
	DeviceCount int                 `json:"deviceCount"`
	InitiatedAt int64               `json:"initiatedAt"`
}

// CancelPushResponse represents the response for POST /v1/updates/history/:id/cancel.
type CancelPushResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	CancelledBy string `json:"cancelledBy"`
	CancelledAt int64  `json:"cancelledAt"`
}

// ListHistoryResponse represents the response for GET /v1/updates/history.
type ListHistoryResponse struct {
	Pushes     []PushHistoryEntry `json:"pushes"`
	Pagination PaginationResponse `json:"pagination"`
}

// PushDetailDevice represents a device in push detail response.
type PushDetailDevice struct {
	ID             string `json:"id"`
	DeviceID       string `json:"deviceId"`
	DeviceName     string `json:"deviceName,omitempty"`
	Status         string `json:"status"`
	SentAt         *int64 `json:"sentAt,omitempty"`
	AcknowledgedAt *int64 `json:"acknowledgedAt,omitempty"`
	Error          string `json:"error,omitempty"`
}

// PushDetailResponse represents the response for GET /v1/updates/history/:pushId.
type PushDetailResponse struct {
	ScheduledAt *int64             `json:"scheduledAt,omitempty"`
	CompletedAt *int64             `json:"completedAt,omitempty"`
	CancelledAt *int64             `json:"cancelledAt,omitempty"`
	ID          string             `json:"id"`
	Version     string             `json:"version"`
	InstallType string             `json:"installType"`
	Status      string             `json:"status"`
	InitiatedBy string             `json:"initiatedBy"`
	Devices     []PushDetailDevice `json:"devices"`
	InitiatedAt int64              `json:"initiatedAt"`
}

// ExportResponse represents the response for GET /v1/updates/export.
type ExportResponse struct {
	Format     string            `json:"format"`
	Versions   []VersionResponse `json:"versions"`
	Changelog  []ChangelogEntry  `json:"changelog"`
	ExportedAt int64             `json:"exportedAt"`
}

// SyncResponse represents the response for POST /v1/updates/sync.
type SyncResponse struct {
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
	StartedAt     int64  `json:"startedAt"`
	VersionsFound int    `json:"versionsFound,omitempty"`
}

// GetSyncStatusResponse represents the response for GET /v1/updates/sync/status.
type GetSyncStatusResponse struct {
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
	LastSyncAt    int64  `json:"lastSyncAt,omitempty"`
	NextSyncAt    int64  `json:"nextSyncAt,omitempty"`
	VersionsFound int    `json:"versionsFound,omitempty"`
}

// NewTimestamp creates a new timestamp in milliseconds.
func NewTimestamp(t time.Time) int64 {
	return t.UnixMilli()
}

// PtrToInt64 returns a pointer to an int64.
func PtrToInt64(v int64) *int64 {
	return &v
}
