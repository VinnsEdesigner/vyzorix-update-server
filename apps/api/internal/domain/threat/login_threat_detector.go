package threat

import (
	"log/slog"
	"strings"
)

// Logger defines the interface for logging threats.
type Logger interface {
	LogThreat(response *ThreatResponse, ctx *LoginContext)
}

// DefaultLogger logs threats using structured logging for security events.
type DefaultLogger struct {
	logger *slog.Logger
}

// NewDefaultLogger creates a logger with the default slog logger.
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		logger: slog.Default(),
	}
}

// NewDefaultLoggerWith creates a logger with a custom slog.Logger.
func NewDefaultLoggerWith(logger *slog.Logger) *DefaultLogger {
	return &DefaultLogger{
		logger: logger,
	}
}

func (l *DefaultLogger) LogThreat(response *ThreatResponse, ctx *LoginContext) {
	// Use structured logging for security events - avoid logging sensitive data
	l.logger.Warn("threat_detected",
		"threat_type", response.Type,
		"severity", response.Severity,
		"action", response.Action,
		"rule", response.Metadata["rule"],
		"operator_id_hash", hashIdentifier(ctx.OperatorID),
		"ip_address", sanitizeIP(ctx.IPAddress),
	)
}

// hashIdentifier creates a one-way hash of an identifier for logging without exposing it.
func hashIdentifier(id string) string {
	if id == "" {
		return "<empty>"
	}
	// Use a simple hash with unsigned arithmetic to avoid overflow
	var h uint64
	for _, c := range id {
		h = h*31 + uint64(c)
	}
	// Convert hash to two letters A-Z using modulo
	// Use uint64 to prevent negative values from overflow
	return string(rune('A'+rune(h%26))) + string(rune('A'+rune((h/26)%26)))
}

// sanitizeIP masks IP addresses for logging (last octet for IPv4, last half for IPv6).
func sanitizeIP(ip string) string {
	if ip == "" {
		return "<empty>"
	}
	// For privacy, show only first 3 octets of IPv4
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return parts[0] + "." + parts[1] + "." + parts[2] + ".x"
	}
	// For IPv6, find the colon separator and take first half
	colonIdx := strings.Index(ip, ":")
	if colonIdx > 0 {
		// This looks like IPv6 - show up to the first colon group
		return ip[:colonIdx] + ":...x"
	}
	return ip
}

// Evaluate checks the login context against all detection rules and returns
// the first matched threat. For multiple threats, use EvaluateAll.
func (d *Detector) Evaluate(ctx *LoginContext, logger Logger) *ThreatResponse {
	if ctx == nil {
		return nil
	}
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
	if rule == nil {
		return
	}
	d.rules = append(d.rules, rule)
}

// EvaluateForBlocking checks if the login should be blocked based on failed attempts.
func EvaluateForBlocking(ctx *LoginContext) bool {
	// Lowered from 10 to 5 for earlier blocking
	if ctx == nil {
		return false
	}
	return ctx.FailedAttempts >= 5
}
