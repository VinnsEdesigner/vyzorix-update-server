package tracing

import (
	"strings"
	"testing"
)

func TestGenerateTraceID(t *testing.T) {
	id1 := GenerateTraceID()
	id2 := GenerateTraceID()

	if len(id1) != TraceIDLength*2 {
		t.Errorf("TraceID length = %d, want %d", len(id1), TraceIDLength*2)
	}

	if id1 == id2 {
		t.Error("GenerateTraceID should generate unique IDs")
	}

	for _, c := range id1 {
		if !isHexChar(c) {
			t.Errorf("GenerateTraceID contains non-hex char: %c", c)
		}
	}
}

func TestValidateTraceID(t *testing.T) {
	validID := GenerateTraceID()

	tests := []struct {
		traceID string
		valid   bool
	}{
		{validID, true},
		{strings.ToLower(validID), true},
		{"aabbccddeeff00112233445566778899", true},
		{"not-hex-characters-here-12345678", false},
		{"too-short", false},
		{"toolongtoobigtoolongtoobigtoolongtoobig", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.traceID, func(t *testing.T) {
			if ValidateTraceID(tt.traceID) != tt.valid {
				t.Errorf("ValidateTraceID(%q) = %v, want %v", tt.traceID, !tt.valid, tt.valid)
			}
		})
	}
}

func TestExtractOrGenerate(t *testing.T) {
	validID := GenerateTraceID()

	result := ExtractOrGenerate(validID)
	if result != validID {
		t.Errorf("ExtractOrGenerate(%q) = %q, want same", validID, result)
	}

	invalidID := "not-valid"
	result = ExtractOrGenerate(invalidID)
	if result == invalidID {
		t.Error("ExtractOrGenerate should generate new ID for invalid input")
	}
	if !ValidateTraceID(result) {
		t.Error("ExtractOrGenerate should return valid trace ID")
	}
}

func TestTraceIDWithPrefix(t *testing.T) {
	traceID := "abc123def456"
	result := TraceIDWithPrefix("req", traceID)

	expected := "req:abc123def456"
	if result != expected {
		t.Errorf("TraceIDWithPrefix = %q, want %q", result, expected)
	}
}

func TestParsePrefixedTraceID(t *testing.T) {
	traceID := GenerateTraceID()
	prefixed := "org:" + traceID

	prefix, tid, ok := ParsePrefixedTraceID(prefixed)
	if !ok {
		t.Fatal("ParsePrefixedTraceID failed")
	}
	if prefix != "org" {
		t.Errorf("prefix = %q, want 'org:acme'", prefix)
	}
	if tid != traceID {
		t.Errorf("traceID = %q, want %q", tid, traceID)
	}
}

func TestParsePrefixedTraceIDInvalid(t *testing.T) {
	_, _, ok := ParsePrefixedTraceID("no-colon-here")
	if ok {
		t.Error("ParsePrefixedTraceID should reject input without colon")
	}
}

func TestNewTraceContext(t *testing.T) {
	ctx := NewTraceContext()

	if ctx.TraceID == "" {
		t.Error("TraceID should be set")
	}
	if len(ctx.SpanID) != 16 {
		t.Errorf("SpanID length = %d, want 16", len(ctx.SpanID))
	}
	if ctx.ParentID != "" {
		t.Error("ParentID should be empty for new context")
	}
}

func TestNewTraceContextWithParent(t *testing.T) {
	parentID := "parent-span-id"
	ctx := NewTraceContextWithParent(parentID)

	if ctx.ParentID != parentID {
		t.Errorf("ParentID = %q, want %q", ctx.ParentID, parentID)
	}
}

func TestTraceContextFormatForLogging(t *testing.T) {
	ctx := &TraceContext{
		TraceID:  "trace-abc",
		SpanID:   "span-123",
		ParentID: "parent-456",
	}

	parts := ctx.FormatForLogging()

	if parts["trace_id"] != "trace-abc" {
		t.Errorf("trace_id = %q, want 'trace-abc'", parts["trace_id"])
	}
	if parts["span_id"] != "span-123" {
		t.Errorf("span_id = %q, want 'span-123'", parts["span_id"])
	}
	if parts["parent_id"] != "parent-456" {
		t.Errorf("parent_id = %q, want 'parent-456'", parts["parent_id"])
	}
}

func TestTraceIDConstants(t *testing.T) {
	if TraceIDHeader != "X-Trace-ID" {
		t.Errorf("TraceIDHeader = %q, want 'X-Trace-ID'", TraceIDHeader)
	}
	if ContextKeyTraceID != "trace_id" {
		t.Errorf("ContextKeyTraceID = %q, want 'trace_id'", ContextKeyTraceID)
	}
	if TraceIDLength != 16 {
		t.Errorf("TraceIDLength = %d, want 16", TraceIDLength)
	}
}

// TestFallbackTraceIDUnique ensures the last-resort fallback still produces
// distinct, well-formed IDs so it can never collide on rapid successive calls.
func TestFallbackTraceIDUnique(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		id := fallbackTraceID()
		if !ValidateTraceID(id) {
			t.Fatalf("fallbackTraceID() = %q is not a valid trace ID", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("fallbackTraceID() produced duplicate %q after %d calls", id, i)
		}
		seen[id] = struct{}{}
	}
}
