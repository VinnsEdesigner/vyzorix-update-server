package device

// DeviceResponse is the safe JSON representation returned to clients.
type DeviceResponse struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Status          DeviceStatus `json:"status"`
	LastSeen        int64        `json:"lastSeen"`
	FirmwareVersion string       `json:"firmwareVersion"`
	IPAddress       string       `json:"ipAddress,omitempty"`
	Tags            []string     `json:"tags,omitempty"`
}
