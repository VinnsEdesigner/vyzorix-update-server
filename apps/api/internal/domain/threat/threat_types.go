package threat

import (
	"sort"
	"time"
)

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
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// severityRank returns numeric rank for severity comparison (higher = more severe).
func severityRank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

// ThreatAction defines what action to take when a threat is detected.
type ThreatAction string

const (
	ActionAllow        ThreatAction = "allow"
	ActionAlert        ThreatAction = "alert"
	ActionMFAChallenge ThreatAction = "mfa_challenge"
	ActionBlock        ThreatAction = "block"
)

// LoginContext contains data about a login attempt for threat evaluation.
// All fields should be validated and sanitized before use.
type LoginContext struct { //
	OperatorID     string
	IPAddress     string
	Location      string
	UserAgent     string
	DeviceFinger  string
	FailedAttempts int
	Timestamp     time.Time
	LastLogin     *LastLoginInfo
}

// LastLoginInfo stores information about the previous login.
type LastLoginInfo struct {
	Timestamp    time.Time
	IPAddress   string
	Location    string
	DeviceFinger string
}

// DetectionRule defines a rule for detecting threats.
type DetectionRule struct {
	Name      string
	Condition func(ctx *LoginContext) bool
	Type      ThreatType
	Severity  Severity
	Action    ThreatAction
}

// ThreatResponse contains the result of threat evaluation.
type ThreatResponse struct {
	Metadata map[string]string
	Type     ThreatType
	Severity Severity
	Action   ThreatAction
	Message  string
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
			Name:     "failed_attempts",
			Type:     ThreatBruteForce,
			Severity: SeverityHigh,
			Action:   ActionBlock,
			Condition: func(ctx *LoginContext) bool {
				return ctx.FailedAttempts >= 5 // Lowered threshold for earlier blocking.
			},
		},
		{
			Name:     "impossible_travel",
			Type:     ThreatImpossibleTravel,
			Severity: SeverityCritical,
			Action:   ActionBlock,
			Condition: func(ctx *LoginContext) bool {
				if ctx.LastLogin == nil {
					return false
				}
				// Check for impossible travel: same operator from different IP within short time.
				timeSinceLastLogin := ctx.Timestamp.Sub(ctx.LastLogin.Timestamp)
				// If less than 30 minutes between logins from different locations, it's suspicious.
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
			Action:   ActionMFAChallenge,
			Condition: func(ctx *LoginContext) bool {
				if ctx.LastLogin == nil {
					return false
				}
				// Only flag if location is significantly different.
				return ctx.Location != "" && ctx.Location != ctx.LastLogin.Location
			},
		},
	}
}

// EvaluateAll checks the login context against all detection rules and returns.
// all matched threats, sorted by severity (most severe first).
func (d *Detector) EvaluateAll(ctx *LoginContext) []*ThreatResponse {
	var threats []*ThreatResponse

	for _, rule := range d.rules {
		if rule.Condition(ctx) {
			threats = append(threats, &ThreatResponse{
				Type:     rule.Type,
				Severity: rule.Severity,
				Action:   rule.Action,
				Message:  "Threat detected: " + rule.Name,
				Metadata: map[string]string{
					"rule": rule.Name,
				},
			})
		}
	}

	// Sort by severity descending.
	sort.Slice(threats, func(i, j int) bool {
		return severityRank(threats[i].Severity) > severityRank(threats[j].Severity)
	})

	return threats
}

// GetBlockingAction returns the most severe blocking action if any threat requires blocking.
func (d *Detector) GetBlockingAction(ctx *LoginContext) *ThreatResponse {
	threats := d.EvaluateAll(ctx)
	for _, t := range threats {
		if t.Action == ActionBlock {
			return t
		}
	}
	return nil
}
