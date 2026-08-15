// Package metrics provides Prometheus metrics collection.
package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

var (
	metricPrefix = "vyzorix_"
)

// Counter represents a monotonically increasing counter.
type Counter struct {
	value atomic.Int64
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	c.value.Add(1)
}

// Add adds the given value to the counter.
func (c *Counter) Add(v int64) {
	c.value.Add(v)
}

// Value returns the current counter value.
func (c *Counter) Value() int64 {
	return c.value.Load()
}

// Gauge represents a gauge metric that can go up and down.
type Gauge struct {
	value atomic.Int64
}

// Set sets the gauge to the given value.
func (g *Gauge) Set(v int64) {
	g.value.Store(v)
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc() {
	g.value.Add(1)
}

// Dec decrements the gauge by 1.
func (g *Gauge) Dec() {
	g.value.Add(-1)
}

// Value returns the current gauge value.
func (g *Gauge) Value() int64 {
	return g.value.Load()
}

// Histogram represents a histogram metric.
type Histogram struct {
	buckets []int64
	bCounts []atomic.Int64
	count   atomic.Int64
	sum     atomic.Int64
}

// NewHistogram creates a new histogram with the given bucket boundaries.
func NewHistogram(buckets []int64) *Histogram {
	h := &Histogram{
		buckets: buckets,
		bCounts: make([]atomic.Int64, len(buckets)+1),
	}
	sort.Slice(h.buckets, func(i, j int) bool {
		return h.buckets[i] < h.buckets[j]
	})

	return h
}

// Observe records a value in the histogram.
func (h *Histogram) Observe(v int64) {
	h.count.Add(1)
	h.sum.Add(v)

	for i, bucket := range h.buckets {
		if v <= bucket {
			h.bCounts[i].Add(1)
			return
		}
	}

	h.bCounts[len(h.buckets)].Add(1)
}

// Metrics holds all application metrics.
type Metrics struct {
	// HTTP metrics.
	HTTPRequestsTotal    *Counter
	HTTPRequestDuration  *Histogram
	HTTPRequestsInFlight *Gauge

	// Auth metrics.
	LoginAttemptsTotal    *Counter
	LoginFailuresTotal    *Counter
	RegisterAttemptsTotal *Counter

	// Security metrics.
	CSRFFailuresTotal    *Counter
	SigningFailuresTotal *Counter
	RateLimitHitsTotal   *Counter
	AccountLockoutsTotal *Counter

	// Session metrics.
	ActiveSessions          *Gauge
	SessionRevocationsTotal *Counter

	// API metrics.
	APIClientCreationsTotal *Counter
	APIRequestsTotal        *Counter

	// Device registration metrics (Bug 46).
	DeviceRegistrationAttemptsTotal *Counter
	DeviceRegistrationSuccessTotal  *Counter
	DeviceRegistrationFailuresTotal *Counter
	InboxApprovalTotal              *Counter
	InboxRejectionTotal             *Counter
	DeviceDeregistrationsTotal      *Counter
	DeviceReRegistrationsTotal      *Counter

	// System metrics.
	UptimeSeconds *Gauge
}

// New creates a new metrics instance.
func New() *Metrics {
	return &Metrics{
		HTTPRequestsTotal:               new(Counter),
		HTTPRequestDuration:             NewHistogram([]int64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}),
		HTTPRequestsInFlight:            new(Gauge),
		LoginAttemptsTotal:              new(Counter),
		LoginFailuresTotal:              new(Counter),
		RegisterAttemptsTotal:           new(Counter),
		CSRFFailuresTotal:               new(Counter),
		SigningFailuresTotal:            new(Counter),
		RateLimitHitsTotal:              new(Counter),
		AccountLockoutsTotal:            new(Counter),
		ActiveSessions:                  new(Gauge),
		SessionRevocationsTotal:         new(Counter),
		APIClientCreationsTotal:         new(Counter),
		APIRequestsTotal:                new(Counter),
		DeviceRegistrationAttemptsTotal: new(Counter),
		DeviceRegistrationSuccessTotal:  new(Counter),
		DeviceRegistrationFailuresTotal: new(Counter),
		InboxApprovalTotal:              new(Counter),
		InboxRejectionTotal:             new(Counter),
		DeviceDeregistrationsTotal:      new(Counter),
		DeviceReRegistrationsTotal:      new(Counter),
		UptimeSeconds:                   new(Gauge),
	}
}

var globalMetrics *Metrics
var startTime = time.Now()

func init() {
	globalMetrics = New()

	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			globalMetrics.UptimeSeconds.Set(int64(time.Since(startTime).Seconds()))
		}
	}()
}

// Get returns the global metrics instance.
func Get() *Metrics {
	return globalMetrics
}

// Collect returns all metrics in Prometheus text format.
func (m *Metrics) Collect() map[string]any {
	return map[string]any{
		"http_requests_total":                m.HTTPRequestsTotal.Value(),
		"http_request_duration_count":        m.HTTPRequestDuration.count.Load(),
		"http_requests_in_flight":            m.HTTPRequestsInFlight.Value(),
		"login_attempts_total":               m.LoginAttemptsTotal.Value(),
		"login_failures_total":               m.LoginFailuresTotal.Value(),
		"register_attempts_total":            m.RegisterAttemptsTotal.Value(),
		"csrf_failures_total":                m.CSRFFailuresTotal.Value(),
		"signing_failures_total":             m.SigningFailuresTotal.Value(),
		"rate_limit_hits_total":              m.RateLimitHitsTotal.Value(),
		"account_lockouts_total":             m.AccountLockoutsTotal.Value(),
		"active_sessions":                    m.ActiveSessions.Value(),
		"session_revocations_total":          m.SessionRevocationsTotal.Value(),
		"api_client_creations_total":         m.APIClientCreationsTotal.Value(),
		"api_requests_total":                 m.APIRequestsTotal.Value(),
		"device_registration_attempts_total": m.DeviceRegistrationAttemptsTotal.Value(),
		"device_registration_success_total":  m.DeviceRegistrationSuccessTotal.Value(),
		"device_registration_failures_total": m.DeviceRegistrationFailuresTotal.Value(),
		"inbox_approval_total":               m.InboxApprovalTotal.Value(),
		"inbox_rejection_total":              m.InboxRejectionTotal.Value(),
		"device_deregistrations_total":       m.DeviceDeregistrationsTotal.Value(),
		"device_re_registrations_total":      m.DeviceReRegistrationsTotal.Value(),
		"uptime_seconds":                     m.UptimeSeconds.Value(),
	}
}

// PrometheusOutput returns metrics in Prometheus text exposition format.
func (m *Metrics) PrometheusOutput() string {
	var sb strings.Builder

	metrics := m.Collect()
	keys := make([]string, 0, len(metrics))

	for k := range metrics {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(&sb, "# HELP %s%s\n", metricPrefix, k)
		fmt.Fprintf(&sb, "# TYPE %s%s gauge\n", metricPrefix, k)
		fmt.Fprintf(&sb, "%s%s %v\n\n", metricPrefix, k, metrics[k])
	}

	return sb.String()
}

// RecordHTTPRequest records an HTTP request.
func (m *Metrics) RecordHTTPRequest(duration time.Duration, statusCode int) {
	m.HTTPRequestsTotal.Inc()
	m.HTTPRequestDuration.Observe(duration.Milliseconds())
}

// RecordLogin records a login attempt.
func (m *Metrics) RecordLogin(success bool) {
	m.LoginAttemptsTotal.Inc()

	if !success {
		m.LoginFailuresTotal.Inc()
	}
}

// RecordRegister records a registration attempt.
func (m *Metrics) RecordRegister() {
	m.RegisterAttemptsTotal.Inc()
}

// RecordSecurityFailure records a security-related failure.
func (m *Metrics) RecordSecurityFailure(failureType string) {
	switch failureType {
	case "csrf":
		m.CSRFFailuresTotal.Inc()
	case "signing":
		m.SigningFailuresTotal.Inc()
	case "rate_limit":
		m.RateLimitHitsTotal.Inc()
	case "lockout":
		m.AccountLockoutsTotal.Inc()
	}
}

// RecordSessionChange records a session creation or revocation.
func (m *Metrics) RecordSessionChange(revoked bool) {
	if revoked {
		m.SessionRevocationsTotal.Inc()
		m.ActiveSessions.Dec()
	} else {
		m.ActiveSessions.Inc()
	}
}

// RecordDeviceRegistrationAttempt records a device registration attempt (Bug 46).
func (m *Metrics) RecordDeviceRegistrationAttempt() {
	m.DeviceRegistrationAttemptsTotal.Inc()
}

// RecordDeviceRegistrationSuccess records a successful device registration.
func (m *Metrics) RecordDeviceRegistrationSuccess() {
	m.DeviceRegistrationSuccessTotal.Inc()
}

// RecordDeviceRegistrationFailure records a failed device registration.
func (m *Metrics) RecordDeviceRegistrationFailure() {
	m.DeviceRegistrationFailuresTotal.Inc()
}

// RecordInboxApproval records an inbox approval.
func (m *Metrics) RecordInboxApproval() {
	m.InboxApprovalTotal.Inc()
}

// RecordInboxRejection records an inbox rejection.
func (m *Metrics) RecordInboxRejection() {
	m.InboxRejectionTotal.Inc()
}

// RecordDeviceDeregistration records a device deregistration.
func (m *Metrics) RecordDeviceDeregistration() {
	m.DeviceDeregistrationsTotal.Inc()
}

// RecordDeviceReRegistration records a device re-registration (after deregistration).
func (m *Metrics) RecordDeviceReRegistration() {
	m.DeviceReRegistrationsTotal.Inc()
}
