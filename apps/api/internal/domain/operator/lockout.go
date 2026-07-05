package operator

import "time"

// LockoutReason represents why an operator was locked out.
type LockoutReason string

const (
	LockoutReasonFailedAttempts  LockoutReason = "failed_attempts"
	LockoutReasonAdminAction     LockoutReason = "admin_action"
	LockoutReasonSecurityEvent   LockoutReason = "security_event"
	LockoutReasonPasswordExpired LockoutReason = "password_expired"
)

// LockoutState represents the lockout status of an operator.
type LockoutState struct {
	LockedAt  *time.Time
	ExpiresAt *time.Time
	Reason    LockoutReason
	LockedBy  string
	Attempts  int
	IsLocked  bool
}

// IsActive returns true if the lockout is currently enforced.
func (s *LockoutState) IsActive() bool {
	if !s.IsLocked {
		return false
	}
	if s.ExpiresAt != nil && time.Now().After(*s.ExpiresAt) {
		return false
	}
	return true
}

// NewLockoutState creates a new lockout state.
func NewLockoutState(reason LockoutReason, lockedBy string, duration time.Duration) *LockoutState {
	state := &LockoutState{
		IsLocked: true,
		Reason:   reason,
		LockedBy: lockedBy,
		Attempts: 0,
	}
	now := time.Now()
	state.LockedAt = &now
	if duration > 0 {
		expires := now.Add(duration)
		state.ExpiresAt = &expires
	}
	return state
}
