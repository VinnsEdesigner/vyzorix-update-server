package threat

import (
	"log"
	"time"
)

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

type Detector struct {
	rules []*DetectionRule
}

func NewDetector() *Detector {
	return &Detector{rules: DefaultRules()}
}

func NewDetectorWithRules(rules []*DetectionRule) *Detector {
	return &Detector{rules: rules}
}

func (d *Detector) Evaluate(ctx *LoginContext) (*ThreatResponse, error) {
	for _, rule := range d.rules {
		if rule.Condition(ctx) {
			return &ThreatResponse{
				Type:     rule.Type,
				Severity: rule.Severity,
				Action:   rule.Action,
				Message:  "Threat detected: " + rule.Name,
				Metadata: map[string]string{
					"rule": rule.Name,
					"time": time.Now().Format(time.RFC3339),
				},
			}, nil
		}
	}
	return nil, nil
}

func (d *Detector) AddRule(rule *DetectionRule) {
	d.rules = append(d.rules, rule)
}

type Logger interface {
	LogThreat(response *ThreatResponse, ctx *LoginContext)
}

type DefaultLogger struct{}

func (l *DefaultLogger) LogThreat(response *ThreatResponse, ctx *LoginContext) {
	log.Printf("[THREAT DETECTED] type=%s severity=%s action=%s operator=%s ip=%s",
		response.Type, response.Severity, response.Action, ctx.OperatorID, ctx.IPAddress)
}
