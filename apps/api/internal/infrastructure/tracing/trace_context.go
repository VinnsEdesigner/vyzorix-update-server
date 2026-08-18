
// Package tracing provides request tracing capabilities.
package tracing

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// TraceIDHeader is the HTTP header containing the trace ID.
	TraceIDHeader = "X-Trace-ID"

	// TraceIDLength is the number of random bytes in a trace ID.
	TraceIDLength = 16

	// ContextKeyTraceID is the key used in gin.Context for trace ID.
	ContextKeyTraceID = "trace_id"

	// ContextKeyRequestStart is the key for request start time.
	ContextKeyRequestStart = "request_start"
)

// Pool for reusing trace ID buffers to reduce allocations.
var bufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, TraceIDLength)
		return &b
	},
}

// GenerateTraceID creates a new unique trace ID.
// Format: 32 hex characters (16 bytes of randomness).
// Example: "a1b2c3d4e5f6789012345678abcdef01"
func GenerateTraceID() string {
	bufferPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufferPtr)

	_, err := rand.Read(*bufferPtr)
	if err != nil {
		// Fallback to a pseudo-random ID if crypto/rand fails
		return fallbackTraceID()
	}

	return hex.EncodeToString(*bufferPtr)
}

// fallbackTraceID generates a trace ID using alternative entropy sources when
// crypto/rand is unavailable. This is a last-resort path (crypto/rand.Read
// failing is essentially impossible on supported platforms) but it must still
// produce a unique, well-formed ID. We mix a high-resolution timestamp with a
// process-wide atomic counter and hash the result with FNV so the output is
// 32 hex chars and ValidateTraceID accepts it.
//
//nolint:gosec // G115: the narrowing int->byte conversions are intentional byte extraction (low 8 bits per shift).
func fallbackTraceID() string {
	fallbackCounter.Add(1)
	var b [TraceIDLength]byte
	nano := time.Now().UnixNano()
	counter := fallbackCounter.Load()
	for i := 0; i < 8; i++ {
		b[i] = byte(nano >> (uint(i) * 8))
	}
	for i := 0; i < 8; i++ {
		b[8+i] = byte(counter >> (uint(i) * 8))
	}
	// Diffuse so the bytes look random, not obviously time-structured.
	h := fnv.New32a()
	_, _ = h.Write(b[:])
	seed := h.Sum32()
	for i := range b {
		b[i] = byte(uint32(b[i]*31) ^ (seed >> (uint(i%4) * 8)))
	}
	return hex.EncodeToString(b[:])
}

// fallbackCounter guarantees uniqueness across rapid successive fallback calls.
var fallbackCounter atomic.Uint64

// ValidateTraceID checks if a trace ID has the expected format.
// Returns true if the trace ID is valid.
func ValidateTraceID(traceID string) bool {
	if len(traceID) != TraceIDLength*2 {
		return false
	}
	// Check if all characters are valid hex
	for _, c := range traceID {
		if !isHexChar(c) {
			return false
		}
	}
	return true
}

// isHexChar returns true if c is a hexadecimal character.
func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

// ExtractTraceID extracts a valid trace ID from the provided string.
// If the input is invalid or empty, generates a new trace ID.
func ExtractOrGenerate(input string) string {
	input = strings.TrimSpace(input)
	if ValidateTraceID(input) {
		return input
	}
	return GenerateTraceID()
}

// TraceIDWithPrefix creates a trace ID with a prefix for easier log scanning.
// Format: "PREFIX:traceid"
// Example: "req:a1b2c3d4..."
func TraceIDWithPrefix(prefix, traceID string) string {
	return fmt.Sprintf("%s:%s", prefix, traceID)
}

// ParsePrefixedTraceID extracts the trace ID from a prefixed trace ID.
// Returns the trace ID and prefix.
func ParsePrefixedTraceID(prefixed string) (prefix, traceID string, ok bool) {
	parts := strings.SplitN(prefixed, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	prefix, traceID = parts[0], parts[1]
	if !ValidateTraceID(traceID) {
		return "", "", false
	}
	return prefix, traceID, true
}

// TraceContext holds trace information for a request.
type TraceContext struct {
	TraceID  string
	SpanID   string
	ParentID string
}

// NewTraceContext creates a new trace context with generated IDs.
func NewTraceContext() *TraceContext {
	return &TraceContext{
		TraceID: GenerateTraceID(),
		SpanID:  GenerateTraceID()[:16], // Short span ID
	}
}

// NewTraceContextWithParent creates a trace context with a parent span ID.
func NewTraceContextWithParent(parentID string) *TraceContext {
	ctx := NewTraceContext()
	ctx.ParentID = parentID
	return ctx
}

// TraceIDParts represents the components of a distributed trace ID.
type TraceIDParts struct {
	OrgID   string
	UserID  string
	TraceID string
	SpanID  string
}

// ParseOrgTraceID extracts org-specific parts from a trace ID.
// Useful for multi-tenant tracing.
func ParseOrgTraceID(traceID, orgID, userID string) TraceIDParts {
	return TraceIDParts{
		OrgID:   orgID,
		UserID:  userID,
		TraceID: traceID,
		SpanID:  traceID[:16],
	}
}

// FormatForLogging formats the trace context for structured logging.
func (t *TraceContext) FormatForLogging() map[string]string {
	parts := map[string]string{
		"trace_id": t.TraceID,
		"span_id":  t.SpanID,
	}
	if t.ParentID != "" {
		parts["parent_id"] = t.ParentID
	}
	return parts
}