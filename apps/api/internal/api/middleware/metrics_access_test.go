package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func metricsEngine(cfg MetricsAccessConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/metrics", MetricsAccess(cfg), func(c *gin.Context) {
		c.String(http.StatusOK, "vyzorix_uptime_seconds 1\n")
	})
	return r
}

func scrape(r *gin.Engine, remoteAddr, auth string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = remoteAddr
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	r.ServeHTTP(w, req)
	return w
}

// TestMetricsAccess_LoopbackAllowedByDefault proves a local scraper (Prometheus on
// the same host) works without a token.
func TestMetricsAccess_LoopbackAllowedByDefault(t *testing.T) {
	r := metricsEngine(MetricsAccessConfig{AllowLoopback: true})

	if w := scrape(r, "127.0.0.1:5000", ""); w.Code != http.StatusOK {
		t.Fatalf("loopback scrape should be allowed, got %d", w.Code)
	}
}

// TestMetricsAccess_PublicDeniedByDefault is the core fix: an internet-facing
// scrape with no token is refused rather than served.
func TestMetricsAccess_PublicDeniedByDefault(t *testing.T) {
	r := metricsEngine(MetricsAccessConfig{AllowLoopback: true})

	w := scrape(r, "203.0.113.9:40000", "")
	if w.Code != http.StatusForbidden {
		t.Fatalf("public scrape with no token must be 403, got %d", w.Code)
	}
	if w.Body.Len() > 0 && w.Body.String()[0] != '{' {
		t.Fatalf("denied response must not carry metrics, got %q", w.Body.String())
	}
}

// TestMetricsAccess_TokenRequiredAndValidated proves the scrape token is enforced.
func TestMetricsAccess_TokenRequiredAndValidated(t *testing.T) {
	r := metricsEngine(MetricsAccessConfig{AllowLoopback: true, Token: "s3cr3t-scrape-token"})

	if w := scrape(r, "203.0.113.9:40000", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("missing token must be 401, got %d", w.Code)
	}
	if w := scrape(r, "203.0.113.9:40000", "Bearer wrong"); w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token must be 401, got %d", w.Code)
	}
	if w := scrape(r, "203.0.113.9:40000", "Bearer s3cr3t-scrape-token"); w.Code != http.StatusOK {
		t.Fatalf("correct token must be 200, got %d", w.Code)
	}
	// Scheme is case-insensitive per RFC 7235.
	if w := scrape(r, "203.0.113.9:40000", "bearer s3cr3t-scrape-token"); w.Code != http.StatusOK {
		t.Fatalf("lowercase bearer must be accepted, got %d", w.Code)
	}
}

// TestMetricsAccess_CIDRAllowlist proves an external Prometheus inside a known
// network is allowed even without a token.
func TestMetricsAccess_CIDRAllowlist(t *testing.T) {
	r := metricsEngine(MetricsAccessConfig{
		AllowLoopback: true,
		AllowedCIDRs:  []string{"10.0.0.0/8", "192.168.1.0/24"},
	})

	if w := scrape(r, "10.1.2.3:5000", ""); w.Code != http.StatusOK {
		t.Fatalf("peer inside the allowlist should be allowed, got %d", w.Code)
	}
	if w := scrape(r, "192.168.1.42:5000", ""); w.Code != http.StatusOK {
		t.Fatalf("peer inside the /24 should be allowed, got %d", w.Code)
	}
	if w := scrape(r, "198.51.100.7:5000", ""); w.Code != http.StatusForbidden {
		t.Fatalf("peer outside the allowlist must be 403, got %d", w.Code)
	}
}

// TestMetricsAccess_IgnoresForwardedForHeader is the regression guard for the
// bypass found while testing: the guard must key off the transport peer, not
// c.ClientIP(), because gin derives ClientIP from X-Forwarded-For, which any
// caller can set. A spoofed loopback header must NOT grant access.
func TestMetricsAccess_IgnoresForwardedForHeader(t *testing.T) {
	r := metricsEngine(MetricsAccessConfig{AllowLoopback: true, Token: "scrape-token"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "203.0.113.9:40000" // a real, remote peer
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	req.Header.Set("X-Real-IP", "127.0.0.1")
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatal("a spoofed X-Forwarded-For: 127.0.0.1 must not bypass the metrics guard")
	}
}

// TestMetricsAccess_IgnoresForwardedForForCIDR proves the allowlist is not
// steerable by a forwarded header either.
func TestMetricsAccess_IgnoresForwardedForForCIDR(t *testing.T) {
	r := metricsEngine(MetricsAccessConfig{AllowedCIDRs: []string{"10.0.0.0/8"}})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "203.0.113.9:40000"
	req.Header.Set("X-Forwarded-For", "10.1.2.3")
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatal("a spoofed X-Forwarded-For must not satisfy the CIDR allowlist")
	}
}

// TestMetricsAccess_BareIPInAllowlist proves a bare address is treated as a host prefix.
func TestMetricsAccess_BareIPInAllowlist(t *testing.T) {
	r := metricsEngine(MetricsAccessConfig{AllowedCIDRs: []string{"203.0.113.9"}})

	if w := scrape(r, "203.0.113.9:5000", ""); w.Code != http.StatusOK {
		t.Fatalf("exact host should be allowed, got %d", w.Code)
	}
	if w := scrape(r, "203.0.113.10:5000", ""); w.Code != http.StatusForbidden {
		t.Fatalf("different host must be 403, got %d", w.Code)
	}
}
