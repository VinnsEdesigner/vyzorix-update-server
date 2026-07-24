package updates

// PushUpdateRequest represents the request for POST /v1/updates/push.
type PushUpdateRequest struct {
	ScheduledAt    *int64   `json:"scheduledAt,omitempty"`
	Version        string   `json:"version" binding:"required"`
	InstallType    string   `json:"installType" binding:"required,oneof=immediate scheduled"`
	OrganizationID string   `json:"organizationId,omitempty"`
	DeviceIDs      []string `json:"deviceIds" binding:"required,min=1"`
}
