package threat

import "time"

// ThreatType defines the type of threat detected.
type ThreatType string

const (
	ThreatImpossibleTravel   ThreatType = "impossible_travel"
	ThreatNewGeography       ThreatType = "new_geography"
	ThreatSuspiciousDevice   ThreatType = "suspicious_device"
	ThreatBruteForce        ThreatType = "brute_force"
	ThreatCredentialStuffing ThreatType = "credential_stuffing"
)

// ThreatSeverity defines the severity level.
type ThreatSeverity string

const (
	SeverityLow      ThreatSeverity = "low"
	SeverityMedium   ThreatSeverity = "medium"
	SeverityHigh     ThreatSeverity = "high"
	SeverityCritical ThreatSeverity = "critical"
)

// ThreatResponse defines the action to take for a detected threat.
type ThreatResponse struct {
	Type     ThreatType
	Severity ThreatSeverity
	Action   ThreatAction
	Message  string
	Metadata map[string]string
}

// ThreatAction defines what action to take when a threat is detected.
type ThreatAction string

const (
	ActionAllow        ThreatAction = "allow"
	ActionAlert        ThreatAction = "alert"
	ActionMFAChallenge ThreatAction = "mfa_challenge"
	ActionBlock        ThreatAction = "block"
	ActionLockout      ThreatAction = "lockout"
)

// LoginContext contains context about a login attempt for threat evaluation.
type LoginContext struct {
	OperatorID     string
	IPAddress     string
	UserAgent     string
	Fingerprint   string
	Location      string
	Timestamp     time.Time
	LastLogin     *LastLoginInfo
	KnownDevices  []string
	FailedAttempts int
}

// LastLoginInfo contains information about the last successful login.
type LastLoginInfo struct {
	Timestamp time.Time
	IPAddress string
	Location  string
}

// Evaluator evaluates login attempts for threats.
type Evaluator interface {
	Evaluate(ctx *LoginContext) (*ThreatResponse, error)
}

// DetectionRule defines a single threat detection rule.
type DetectionRule struct {
	Name      string
	Type      ThreatType
	Severity  ThreatSeverity
	Action    ThreatAction
	Condition func(*LoginContext) bool
}
