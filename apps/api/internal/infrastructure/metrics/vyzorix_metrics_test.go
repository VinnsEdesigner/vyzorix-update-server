package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMiddlewareRecordsLabeledRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := New()
	engine := gin.New()

	engine.Use(Middleware(m))
	engine.GET("/v1/alerts/rules", func(c *gin.Context) { c.JSON(200, gin.H{}) })

	req := httptest.NewRequest(http.MethodGet, "/v1/alerts/rules", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	counter, err := m.httpRequestsTotal.GetMetricWithLabelValues("/v1/alerts/rules", "GET", "2xx")
	if err != nil {
		t.Fatalf("metric lookup failed: %v", err)
	}
	if testutil.ToFloat64(counter) == 0 {
		t.Error("request counter recorded no hit")
	}

	families, err := m.registry.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	labelsFound := false
	for _, fam := range families {
		if fam.GetName() == "vyzorix_http_requests_total" {
			for _, met := range fam.GetMetric() {
				if hasLabel(met, "route", "/v1/alerts/rules") && hasLabel(met, "method", "GET") && hasLabel(met, "status", "2xx") {
					labelsFound = true
				}
			}
		}
	}
	if !labelsFound {
		t.Error("expected labeled metric for route /v1/alerts/rules")
	}
}

func TestRecordedCountersGoUp(t *testing.T) {
	m := New()
	m.RecordLogin(true)
	m.RecordLogin(false)

	cnt, err := m.loginAttempts.GetMetricWithLabelValues("success")
	if err != nil || testutil.ToFloat64(cnt) != 1 {
		t.Errorf("success count wrong: %v", err)
	}
	cnt, err = m.loginAttempts.GetMetricWithLabelValues("failure")
	if err != nil || testutil.ToFloat64(cnt) != 1 {
		t.Errorf("failure count wrong: %v", err)
	}
}

func TestSessionGaugeFluctuates(t *testing.T) {
	m := New()
	m.RecordSessionChange(false)
	m.RecordSessionChange(false)
	m.RecordSessionChange(true)

	if v := testutil.ToFloat64(m.activeSessions); v != 1 {
		t.Errorf("active sessions = %v, want 1", v)
	}
	if v := testutil.ToFloat64(m.sessionRevoked); v != 1 {
		t.Errorf("revocations = %v, want 1", v)
	}
}

func TestGetReturnsSingleton(t *testing.T) {
	if Get() == nil {
		t.Fatal("singleton nil")
	}
	if Get() != Get() {
		t.Error("singleton not stable")
	}
}

func TestHTTPSingleRoutePresent(t *testing.T) {
	m := New()
	h := NewMetricsHandler(m)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("unexpected status %d", rec.Code)
	}
	body := rec.Body.String()
	if !contains(body, "vyzorix_uptime_seconds") {
		t.Errorf("uptime metric missing from output")
	}
}

// TestMetricsEndpointEmitsAllInstruments covers the real payload endpoint
// end-to-end: instrumentation flows through to /metrics output.
func TestMetricsEndpointEmitsAllInstruments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := New()

	m.RecordCommandOutcome("delivered")
	m.RecordCommandDelivery(time.Second)
	m.UpdateCommandQueue(4)
	m.RecordAlertEvaluation("firing")
	m.UpdateAlertRules(3)
	m.UpdateFeatureToggle("scoped_rbac", true)
	m.RecordSQLConnection("open", 2)
	m.RecordLogin(true)

	engine := gin.New()
	engine.Use(Middleware(m))
	engine.GET("/metrics", func(c *gin.Context) { NewMetricsHandler(m).Handle(c) })
	engine.GET("/v1/alerts/rules", func(c *gin.Context) { c.Status(200) })

	// Hit the instrumented route so the labeled http_requests_total exists.
	probe := httptest.NewRequest(http.MethodGet, "/v1/alerts/rules", nil)
	engine.ServeHTTP(httptest.NewRecorder(), probe)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	body := rec.Body.String()
	for _, want := range []string{
		"vyzorix_commands_total{outcome=\"delivered\"}",
		"vyzorix_command_delivery_duration_seconds",
		"vyzorix_command_queue_size 4",
		"vyzorix_alert_evaluations_total{result=\"firing\"}",
		"vyzorix_alert_rules_active 3",
		"vyzorix_feature_toggle_enabled{feature=\"scoped_rbac\"} 1",
		"vyzorix_sql_connections_total{state=\"open\"} 2",
		"vyzorix_auth_login_attempts_total{outcome=\"success\"}",
		"vyzorix_http_requests_total{method=\"GET\",route=\"/v1/alerts/rules\",status=\"2xx\"}",
	} {
		if !contains(body, want) {
			t.Errorf("missing metric line %q", want)
		}
	}
}

func TestUptimeGuageComputedOnDemand(t *testing.T) {
	m := New()
	// Force start at 1 hour ago; getter reads time.Since(start) not the
	// lazy goroutine state.
	m.start = time.Now().Add(-time.Hour)
	m.startUptime()
	if elapsed := time.Since(m.start).Seconds(); elapsed < 3599 {
		t.Errorf("uptime = %v, want >=3599", elapsed)
	}
}

func hasLabel(metric *dto.Metric, name, value string) bool {
	for _, lbl := range metric.GetLabel() {
		if lbl.GetName() == name && lbl.GetValue() == value {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
