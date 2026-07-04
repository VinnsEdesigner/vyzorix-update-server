package threat

import "time"

// ThreatType categorizes the type of threat detected.
type ThreatType string

const (
	ThreatImpossibleTravel ThreatType = "impossible_travel"
	ThreatNewGeography     ThreatType = "new_geography"
	ThreatBruteForce       ThreatType = "brute_force"
	ThreatSuspiciousDevice ThreatType = "suspicious_device"
)

// Severity indicates the severity level of a threat.
type Severity string

const (
	SeverityLow     Severity = "low"
	SeverityMedium  Severity = "medium"
	SeverityHigh    Severity = "high"
	SeverityCritical Severity = "critical"
)

// ThreatAction defines what action to take when a threat is detected.
type ThreatAction string

const (
	ActionAllow         ThreatAction = "allow"
	ActionAlert         ThreatAction = "alert"
	ActionMFAChallenge  ThreatAction = "mfa_challenge"
	ActionBlock         ThreatAction = "block"
)

// LoginContext contains data about a login attempt for threat evaluation.
type LoginContext struct {
	OperatorID    string
	IPAddress    string
	Location      string
	UserAgent    string
	Timestamp    time.Time
	LastLogin    *LastLoginInfo
	FailedAttempts int
	DeviceFinger string
}

// LastLoginInfo stores information about the previous login.
type LastLoginInfo struct {
	Timestamp  time.Time
	IPAddress  string
	Location    string
	DeviceFinger string
}

// DetectionRule defines a rule for detecting threats.
type DetectionRule struct {
	Name      string
	Type      ThreatType
	Severity  Severity
	Action    ThreatAction
	Condition func(ctx *LoginContext) bool
}

// ThreatResponse contains the result of threat evaluation.
type ThreatResponse struct {
	Type     ThreatType
	Severity Severity
	Action   ThreatAction
	Message  string
	Metadata map[string]string
}

// Detector evaluates login attempts for threats.
type Detector struct {
	rules []*DetectionRule
}

// NewDetector creates a detector with default rules.
func NewDetector() *Detector {
	return &Detector{
		rules: DefaultRules(),
	}
}

// DefaultRules returns the default threat detection rules.
func DefaultRules() []*DetectionRule {
	return []*DetectionRule{
		{
			Name:     "impossible_travel",
			Type:     ThreatImpossibleTravel,
			Severity: SeverityHigh,
			Action:   ActionMFAChallenge,
			Condition: func(ctx *LoginContext) bool {
				if ctx.LastLogin == nil {
					return false
				}
				timeSinceLastLogin := ctx.Timestamp.Sub(ctx.LastLogin.Timestamp)
				if timeSinceLastLogin < 30*time.Minute && ctx.IPAddress != ctx.LastLogin.IPAddress {
					return true
				}
				return false
			},
		},
		{
			Name:     "new_geography",
			Type:     ThreatNewGeography,
			Severity: SeverityMedium,
			Action:   ActionAlert,
			Condition: func(ctx *LoginContext) bool {
				if ctx.LastLogin == nil {
					return false
				}
				return ctx.Location != "" && ctx.Location != ctx.LastLogin.Location
			},
		},
		{
			Name:     "failed_attempts",
			Type:     ThreatBruteForce,
			Severity: SeverityHigh,
			Action:   ActionBlock,
			Condition: func(ctx *LoginContext) bool {
				return ctx.FailedAttempts >= 10
			},
		},
	}
}
