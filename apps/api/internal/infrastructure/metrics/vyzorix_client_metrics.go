// Package metrics wires the domain metrics to prometheus/client_golang
// collectors. Collectors are registered on a private Registry per Metrics
// instance so tests and the singleton do not fight over the global default.
package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metric names uses the vyzor_ prefix; namespace is fixed at registration.
const namespace = "vyzorix"

// Metrics owns the prometheus collectors for all instrumented aspects.
type Metrics struct {
	registry *prometheus.Registry

	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	loginAttempts    *prometheus.CounterVec
	loginFailures    *prometheus.CounterVec

	securityFailures *prometheus.CounterVec

	activeSessions prometheus.Gauge
	sessionRevoked prometheus.Counter

	apiClientCreations prometheus.Counter

	deviceRegAttempts *prometheus.CounterVec

	commandsTotal    *prometheus.CounterVec
	commandDelivery  prometheus.Histogram
	commandQueueSize prometheus.Gauge

	alertRulesActive    prometheus.Gauge
	alertEvaluationTotal *prometheus.CounterVec

	sqlConnections *prometheus.GaugeVec

	featureToggles *prometheus.GaugeVec

	notificationDeliveries *prometheus.CounterVec
	serviceAccountTokens  *prometheus.CounterVec

	uptime prometheus.Gauge
	start  time.Time

	// Uptime goroutine starts once per instance via initOnce.
	initOnce sync.Once
}


// counterVec short-cuts named CounterVec construction because funlen will
// complain otherwise.
func counterVec(name, help string, labels ...string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      name,
		Help:      help,
	}, labels)
}

// New creates a Metrics instance registering collectors on a fresh registry.
func New() *Metrics {
	return registerAll(newMetrics())
}

// newMetrics builds the collector set (kept short; split from registerAll).
func newMetrics() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		start:    time.Now(),
	}

	m.httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "HTTP requests by route, method, and status.",
	}, []string{"route", "method", "status"})

	m.httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request duration by route, method, and status.",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"route", "method", "status"})

	m.httpRequestsInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "http_requests_in_flight",
		Help:      "HTTP requests currently being served.",
	})

	m.loginAttempts = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "login_attempts_total",
		Help:      "Login attempts by outcome.",
	}, []string{"outcome"})

	m.loginFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "login_failures_total",
		Help:      "Login failures by reason.",
	}, []string{"reason"})


	m.securityFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "security_failures_total",
		Help:      "Security failures by class (csrf, signing, rate_limit, lockout).",
	}, []string{"class"})

	m.activeSessions = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "active_sessions",
		Help:      "Currently active user sessions.",
	})

	m.sessionRevoked = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "session_revocations_total",
		Help:      "Session revocations.",
	})

	m.apiClientCreations = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "api_client_creations_total",
		Help:      "API client creations.",
	})

	m.deviceRegAttempts = counterVec("device_registration_total", "Device registration outcome counts.", "outcome")

	m.uptime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "uptime_seconds",
		Help:      "Process uptime in seconds.",
	})

	m.commandsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "commands_total",
		Help:      "Command lifecycle events by outcome.",
	}, []string{"outcome"})

	m.commandDelivery = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "command_delivery_duration_seconds",
		Help:      "Time from command creation to delivery confirmation.",
		Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
	})

	m.commandQueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "command_queue_size",
		Help:      "Pending command outbox queue depth.",
	})

	m.alertRulesActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "alert_rules_active",
		Help:      "Number of enabled alert rules.",
	})

	m.alertEvaluationTotal = counterVec("alert_evaluations_total", "Alert evaluations by transition.", "result")
	m.sqlConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sql_connections_total",
		Help:      "Database connection counts by state.",
	}, []string{"state"})

	m.featureToggles = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "feature_toggle_enabled",
		Help:      "Feature toggle states: 1 = on, 0 = off.",
	}, []string{"feature"})

	m.notificationDeliveries = counterVec("notification_deliveries_total", "Notification deliveries by channel and outcome.", "channel", "outcome")
	m.serviceAccountTokens = counterVec("service_account_tokens_total", "Service account token lifecycle events by action.", "action")

	m.loginAttempts = counterVec("auth_login_attempts_total", "Auth login attempts by outcome.", "outcome")

	collectors := []prometheus.Collector{
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestsInFlight,
		m.loginAttempts,
		m.loginFailures,
		
		m.securityFailures,
		m.activeSessions,
		m.sessionRevoked,
		m.apiClientCreations,
		m.deviceRegAttempts,
		m.uptime,
		m.commandsTotal,
		m.commandDelivery,
		m.commandQueueSize,
		m.alertRulesActive,
		m.alertEvaluationTotal,
		m.sqlConnections,
		m.featureToggles,
		m.notificationDeliveries,
		m.serviceAccountTokens,
	}
	for _, c := range collectors {
		m.registry.MustRegister(c)
	}

	m.startUptime()

	return m
}

// registerAll finalizes the collector registration.
func registerAll(m *Metrics) *Metrics {
	// Entry-point hook.
	return m
}

// Registry exposes the private registry for the handler and tests.
func (m *Metrics) Registry() *prometheus.Registry { return m.registry }

// RecordHTTPRequest records an HTTP request with route, method, status labels.
func (m *Metrics) RecordHTTPRequest(route, method string, status int, duration time.Duration) {
	statusStr := "2xx"
	switch {
	case status >= 500:
		statusStr = "5xx"
	case status >= 400:
		statusStr = "4xx"
	case status >= 300:
		statusStr = "3xx"
	}
	labels := prometheus.Labels{"route": route, "method": method, "status": statusStr}
	m.httpRequestsTotal.With(labels).Inc()
	m.httpRequestDuration.With(labels).Observe(duration.Seconds())
}

// RecordLogin records a login attempt with its outcome.
func (m *Metrics) RecordLogin(success bool) {
	outcome := "success"
	if !success {
		outcome = "failure"
	}
	m.loginAttempts.WithLabelValues(outcome).Inc()
}

// RecordLoginFailure records a failed login with a reason label (e.g., lockout).
func (m *Metrics) RecordLoginFailure(reason string) {
	m.loginFailures.WithLabelValues(reason).Inc()
}

// RecordDeviceRegistrationAttempt records a registration lifecycle event.
func (m *Metrics) RecordDeviceRegistrationAttempt() {
	m.deviceRegAttempts.WithLabelValues("attempt").Inc()
}

// RecordDeviceRegistrationSuccess records a successful device registration.
func (m *Metrics) RecordDeviceRegistrationSuccess() {
	m.deviceRegAttempts.WithLabelValues("success").Inc()
}

// RecordDeviceRegistrationFailure records a failed registration or inbox rejection.
func (s *Metrics) RecordRegisterFailure() {
	s.deviceRegAttempts.WithLabelValues("failure").Inc()
}

// RecordDeviceDeregistration records a deregistration.
func (m *Metrics) RecordDeviceDeregistration() {
	m.deviceRegAttempts.WithLabelValues("deregistered").Inc()
}

// RecordDeviceReRegistration records a post-deregister registration.
func (m *Metrics) RecordDeviceReRegistration() {
	m.deviceRegAttempts.WithLabelValues("re_registered").Inc()
}

// RecordDeviceRegistrationFailure records a failed registration (inbox path).
func (m *Metrics) RecordDeviceRegistrationFailure() {
	m.deviceRegAttempts.WithLabelValues("failure").Inc()
}

// RecordSecurityFailure records a security failure by class.
func (m *Metrics) RecordSecurityFailure(class string) {
	m.securityFailures.WithLabelValues(class).Inc()
}

// RecordSessionChange tracks active sessions and revocations.
func (m *Metrics) RecordSessionChange(revoked bool) {
	if revoked {
		m.activeSessions.Dec()
		m.sessionRevoked.Inc()
		return
	}
	m.activeSessions.Inc()
}

// RecordAPIClientCreation counts API client creations.
func (m *Metrics) RecordAPIClientCreation() {
	m.apiClientCreations.Inc()
}

// Uptime goroutine refreshes the gauge every minute.
func (m *Metrics) startUptime() {
	m.initOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				m.uptime.Set(time.Since(m.start).Seconds())
			}
		}()
	})
}

// RecordCommandOutcome counts a delivered / retried / failed command.
func (m *Metrics) RecordCommandOutcome(outcome string) {
	m.commandsTotal.WithLabelValues(outcome).Inc()
}

// RecordCommandDelivery measures time from creation to delivery.
func (m *Metrics) RecordCommandDelivery(elapsed time.Duration) {
	m.commandDelivery.Observe(elapsed.Seconds())
}

// UpdateCommandQueue stores the current outbox queue depth.
func (m *Metrics) UpdateCommandQueue(pending int) {
	m.commandQueueSize.Set(float64(pending))
}

// RecordAlertEvaluation counts a rule evaluation outcome.
func (m *Metrics) RecordAlertEvaluation(result string) {
	m.alertEvaluationTotal.WithLabelValues(result).Inc()
}

// UpdateAlertRules stores the active rule count.
func (m *Metrics) UpdateAlertRules(count int) {
	m.alertRulesActive.Set(float64(count))
}

// RecordSQLConnection sets database connection counts by state: open, idle, in_use.
func (m *Metrics) RecordSQLConnection(state string, count int) {
	m.sqlConnections.WithLabelValues(state).Set(float64(count))
}

// UpdateFeatureToggle reports toggle state: enabled -> 1, disabled -> 0.
func (m *Metrics) UpdateFeatureToggle(name string, enabled bool) {
	v := 0.0
	if enabled {
		v = 1
	}
	m.featureToggles.WithLabelValues(name).Set(v)
}

// RecordNotificationDelivery counts a delivery attempt by channel and outcome.
func (m *Metrics) RecordNotificationDelivery(channel, outcome string) {
	m.notificationDeliveries.WithLabelValues(channel, outcome).Inc()
}

// RecordServiceAccountToken counts service account token lifecycle events.
func (m *Metrics) RecordServiceAccountToken(serviceID, action string) {
	m.serviceAccountTokens.WithLabelValues(action).Inc()
}

var singletonOnce sync.Once
var singleton *Metrics

// Get returns the global singleton instance, initialized lazily on first call.
func Get() *Metrics {
singletonOnce.Do(func() { singleton = New() })
return singleton
}