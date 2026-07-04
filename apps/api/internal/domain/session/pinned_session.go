package session

import "time"

// PinnedSession represents a session pinned to a specific device/IP for trust.
type PinnedSession struct {
	ID           string
	SessionID    string
	OperatorID   string
	DeviceFinger string
	IPAddress    string
	PinnedAt     time.Time
	LastVerified time.Time
	IsTrusted    bool
}

// VerifyPinnedSession checks if the session matches the pinned device/IP.
func (p *PinnedSession) VerifyPinnedSession(currentIP, deviceFinger string) bool {
	if !p.IsTrusted {
		return false
	}
	// Allow some IP variation for mobile users (same /24 subnet)
	if p.IPAddress != currentIP {
		// Could implement subnet matching here
		return false
	}
	return p.DeviceFinger == deviceFinger || p.DeviceFinger == ""
}

// UpdateVerification updates the last verified timestamp.
func (p *PinnedSession) UpdateVerification() {
	p.LastVerified = time.Now()
}
