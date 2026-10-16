package subscription

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
)

func originChecker(allowed ...string) func(*http.Request) bool {
	return allowedOrigins(config.Config{AllowedOrigins: allowed})
}

func requestWithOrigin(origin string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}

// TestAllowedOrigins_WildcardSubdomainNoSuffixBypass is the regression guard for
// the HasSuffix bug: "*.example.com" must NOT match a lookalike registrable domain
// that merely ends with the same characters.
func TestAllowedOrigins_WildcardSubdomainNoSuffixBypass(t *testing.T) {
	allow := originChecker("*.example.com")

	// Genuine subdomains and the bare domain.
	for _, ok := range []string{
		"https://app.example.com",
		"https://a.b.example.com",
		"https://example.com",
	} {
		if !allow(requestWithOrigin(ok)) {
			t.Errorf("%s should be allowed for *.example.com", ok)
		}
	}

	// Lookalikes that the old suffix check wrongly accepted.
	for _, bad := range []string{
		"https://evil-example.com",
		"https://notexample.com",
		"https://example.com.evil.com",
		"https://example.com.attacker.net",
	} {
		if allow(requestWithOrigin(bad)) {
			t.Errorf("%s must NOT be allowed for *.example.com", bad)
		}
	}
}

func TestAllowedOrigins_ExactMatch(t *testing.T) {
	allow := originChecker("https://dash.example.com")

	if !allow(requestWithOrigin("https://dash.example.com")) {
		t.Error("exact origin should be allowed")
	}
	if allow(requestWithOrigin("https://other.example.com")) {
		t.Error("a different host must not be allowed")
	}
}

func TestAllowedOrigins_EmptyOriginAllowed(t *testing.T) {
	allow := originChecker("https://dash.example.com")
	if !allow(requestWithOrigin("")) {
		t.Error("a request with no Origin header should be allowed")
	}
}

func TestAllowedOrigins_MalformedRejected(t *testing.T) {
	allow := originChecker("*.example.com")
	for _, bad := range []string{"not a url", "://nohost", "example.com"} {
		if allow(requestWithOrigin(bad)) {
			t.Errorf("malformed origin %q must be rejected", bad)
		}
	}
}

func TestAllowedOrigins_WildcardAll(t *testing.T) {
	allow := originChecker("*")
	if !allow(requestWithOrigin("https://anything.example")) {
		t.Error("'*' should allow any origin")
	}
}

func TestAllowedOrigins_TrailingDotDomain(t *testing.T) {
	allow := originChecker("*.example.com.")
	if !allow(requestWithOrigin("https://app.example.com")) {
		t.Error("a trailing dot on the configured domain should still match")
	}
}
