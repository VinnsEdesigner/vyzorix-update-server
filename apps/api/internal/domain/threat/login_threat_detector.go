package threat

import "log"

// Logger defines the interface for logging threats.
type Logger interface {
	LogThreat(response *ThreatResponse, ctx *LoginContext)
}

// DefaultLogger logs threats using the standard log package.
type DefaultLogger struct{}

func (l *DefaultLogger) LogThreat(response *ThreatResponse, ctx *LoginContext) {
	log.Printf("[THREAT] type=%s severity=%s action=%s operator=%s ip=%s",
		response.Type, response.Severity, response.Action, ctx.OperatorID, ctx.IPAddress)
}

// Evaluate checks the login context against all detection rules.
func (d *Detector) Evaluate(ctx *LoginContext, logger Logger) *ThreatResponse {
	for _, rule := range d.rules {
		if rule.Condition(ctx) {
			response := &ThreatResponse{
				Type:     rule.Type,
				Severity: rule.Severity,
				Action:   rule.Action,
				Message:  "Threat detected: " + rule.Name,
				Metadata: map[string]string{
					"rule": rule.Name,
				},
			}
			if logger != nil {
				logger.LogThreat(response, ctx)
			}
			return response
		}
	}
	return nil
}

// AddRule adds a custom detection rule to the detector.
func (d *Detector) AddRule(rule *DetectionRule) {
	d.rules = append(d.rules, rule)
}

// EvaluateForBlocking checks if the login should be blocked.
func EvaluateForBlocking(ctx *LoginContext) bool {
	if ctx.FailedAttempts >= 10 {
		return true
	}
	return false
}
