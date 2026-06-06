package security

import (
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

func TestOriginValidator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		origins    []string
		testOrigin string
		want       bool
	}{
		// Empty origin tests (non-browser clients)
		{
			name:       "empty origin allowed",
			origins:    []string{"https://example.com"},
			testOrigin: "",
			want:       true,
		},
		{
			name:       "empty origin with wildcard",
			origins:    []string{"*"},
			testOrigin: "",
			want:       true,
		},

		// Wildcard tests
		{
			name:       "wildcard allows any origin",
			origins:    []string{"*"},
			testOrigin: "https://evil.com",
			want:       true,
		},

		// Direct match tests
		{
			name:       "exact match allowed",
			origins:    []string{"https://example.com"},
			testOrigin: "https://example.com",
			want:       true,
		},
		{
			name:       "non-matching origin rejected",
			origins:    []string{"https://example.com"},
			testOrigin: "https://evil.com",
			want:       false,
		},

		// Case-insensitive tests
		{
			name:       "case insensitive match",
			origins:    []string{"https://Example.com"},
			testOrigin: "https://example.com",
			want:       true,
		},
		{
			name:       "uppercase origin matches lowercase config",
			origins:    []string{"https://example.com"},
			testOrigin: "HTTPS://EXAMPLE.COM",
			want:       true,
		},

		// Scheme tests
		{
			name:       "http rejected in production",
			origins:    []string{"https://example.com"},
			testOrigin: "http://example.com",
			want:       false,
		},
		{
			name:       "ws rejected",
			origins:    []string{"https://example.com"},
			testOrigin: "ws://example.com",
			want:       false,
		},

		// Multiple origins
		{
			name:       "first origin matches",
			origins:    []string{"https://a.com", "https://b.com"},
			testOrigin: "https://a.com",
			want:       true,
		},
		{
			name:       "second origin matches",
			origins:    []string{"https://a.com", "https://b.com"},
			testOrigin: "https://b.com",
			want:       true,
		},
		{
			name:       "neither origin matches",
			origins:    []string{"https://a.com", "https://b.com"},
			testOrigin: "https://c.com",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewOriginValidator(tt.origins)
			if got := v.Validate(tt.testOrigin); got != tt.want {
				t.Errorf("Validate(%q) = %v, want %v", tt.testOrigin, got, tt.want)
			}
		})
	}
}

func TestOriginValidator_CheckOrigin(t *testing.T) {
	v := NewOriginValidator([]string{"https://example.com", "https://app.example.com"})
	checkOrigin := v.CheckOrigin()

	tests := []struct {
		name       string
		origin     string
		wantResult bool
	}{
		{
			name:       "allowed origin",
			origin:     "https://example.com",
			wantResult: true,
		},
		{
			name:       "second allowed origin",
			origin:     "https://app.example.com",
			wantResult: true,
		},
		{
			name:       "disallowed origin",
			origin:     "https://evil.com",
			wantResult: false,
		},
		{
			name:       "empty origin",
			origin:     "",
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Header: http.Header{},
			}
			req.Header.Set("Origin", tt.origin)

			if got := checkOrigin(req); got != tt.wantResult {
				t.Errorf("CheckOrigin() with Origin %q = %v, want %v", tt.origin, got, tt.wantResult)
			}
		})
	}
}

func TestOriginValidator_CheckOriginWithLogging(t *testing.T) {
	// Create a logger
	logger := slog.Default()

	v := NewOriginValidator([]string{"https://example.com"})
	v.SetLogger(logger)

	// Should not panic with logger set
	checkOrigin := v.CheckOrigin()

	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("Origin", "https://evil.com")

	// Should reject without panic
	if got := checkOrigin(req); got != false {
		t.Errorf("CheckOrigin() should reject evil.com, got %v", got)
	}
}

func TestOriginValidator_SetLogger(t *testing.T) {
	v := NewOriginValidator([]string{"https://example.com"})

	// Should not panic
	v.SetLogger(slog.Default())

	// Should work with nil logger too
	v.SetLogger(nil)

	checkOrigin := v.CheckOrigin()
	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("Origin", "https://example.com")

	if got := checkOrigin(req); got != true {
		t.Errorf("CheckOrigin() should allow example.com, got %v", got)
	}
}

func TestOriginValidator_AllowedOrigins(t *testing.T) {
	origins := []string{"https://a.com", "https://b.com", "https://C.com"}
	v := NewOriginValidator(origins)

	got := v.AllowedOrigins()

	// Should have at least the normalized versions of each origin
	if len(got) < 3 {
		t.Errorf("AllowedOrigins() returned %d items, want at least 3", len(got))
	}

	// Check that all original origins are present
	for _, origin := range origins {
		found := false
		for _, g := range got {
			if strings.EqualFold(g, origin) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AllowedOrigins() missing origin %q", origin)
		}
	}
}

func TestOriginValidator_IsWildcardAllowed(t *testing.T) {
	tests := []struct {
		name      string
		origins   []string
		wantWild  bool
	}{
		{
			name:      "wildcard present",
			origins:   []string{"*"},
			wantWild:  true,
		},
		{
			name:      "wildcard with others",
			origins:   []string{"https://example.com", "*"},
			wantWild:  true,
		},
		{
			name:      "no wildcard",
			origins:   []string{"https://example.com"},
			wantWild:  false,
		},
		{
			name:      "empty list",
			origins:   []string{},
			wantWild:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewOriginValidator(tt.origins)
			if got := v.IsWildcardAllowed(); got != tt.wantWild {
				t.Errorf("IsWildcardAllowed() = %v, want %v", got, tt.wantWild)
			}
		})
	}
}

func TestOriginValidator_ValidateWithDetails(t *testing.T) {
	tests := []struct {
		name          string
		origins       []string
		testOrigin    string
		wantValid     bool
		wantReason    string
	}{
		{
			name:       "empty origin details",
			origins:    []string{"https://example.com"},
			testOrigin: "",
			wantValid:  true,
			wantReason: "empty origin (non-browser client)",
		},
		{
			name:       "wildcard reason",
			origins:    []string{"*"},
			testOrigin: "https://anything.com",
			wantValid:  true,
			wantReason: "wildcard origin allowed",
		},
		{
			name:       "direct match reason",
			origins:    []string{"https://example.com"},
			testOrigin: "https://example.com",
			wantValid:  true,
			wantReason: "direct match",
		},
		{
			name:       "non-secure scheme",
			origins:    []string{"https://example.com"},
			testOrigin: "http://example.com",
			wantValid:  false,
			wantReason: "non-secure scheme rejected",
		},
		{
			name:       "not in allowed list",
			origins:    []string{"https://example.com"},
			testOrigin: "https://other.com",
			wantValid:  false,
			wantReason: "origin not in allowed list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewOriginValidator(tt.origins)
			valid, reason := v.ValidateWithDetails(tt.testOrigin)

			if valid != tt.wantValid {
				t.Errorf("ValidateWithDetails(%q) valid = %v, want %v", tt.testOrigin, valid, tt.wantValid)
			}
			if reason != tt.wantReason {
				t.Errorf("ValidateWithDetails(%q) reason = %q, want %q", tt.testOrigin, reason, tt.wantReason)
			}
		})
	}
}

func TestOriginValidator_Trimming(t *testing.T) {
	tests := []struct {
		name       string
		origins    []string
		testOrigin string
		want       bool
	}{
		{
			name:       "trim spaces from config",
			origins:    []string{"  https://example.com  "},
			testOrigin: "https://example.com",
			want:       true,
		},
		{
			name:       "trim tabs from config",
			origins:    []string{"\thttps://example.com\t"},
			testOrigin: "https://example.com",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewOriginValidator(tt.origins)
			if got := v.Validate(tt.testOrigin); got != tt.want {
				t.Errorf("Validate(%q) = %v, want %v", tt.testOrigin, got, tt.want)
			}
		})
	}
}

func TestOriginValidator_CheckOriginWithoutLogging(t *testing.T) {
	v := NewOriginValidator([]string{"https://example.com"})
	checkOrigin := v.CheckOriginWithoutLogging()

	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("Origin", "https://evil.com")

	if got := checkOrigin(req); got != false {
		t.Errorf("CheckOriginWithoutLogging() should reject evil.com")
	}
}

// Benchmark tests
func BenchmarkOriginValidator_Validate(b *testing.B) {
	v := NewOriginValidator([]string{"https://example.com", "https://app.example.com", "https://dashboard.example.com"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate("https://example.com")
	}
}

func BenchmarkOriginValidator_CheckOrigin(b *testing.B) {
	v := NewOriginValidator([]string{"https://example.com"})
	checkOrigin := v.CheckOrigin()

	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("Origin", "https://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkOrigin(req)
	}
}
