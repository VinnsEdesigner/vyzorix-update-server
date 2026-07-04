package device

// DeviceStatus represents the current status of a device.
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusPending DeviceStatus = "pending"
	DeviceStatusError   DeviceStatus = "error"
)

// IsOnline returns true if the device is online.
func (s DeviceStatus) IsOnline() bool {
	return s == DeviceStatusOnline
}

// IsOffline returns true if the device is offline.
func (s DeviceStatus) IsOffline() bool {
	return s == DeviceStatusOffline
}
