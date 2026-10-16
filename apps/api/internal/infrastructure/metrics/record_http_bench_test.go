package metrics

import (
	"testing"
	"time"
)

// BenchmarkRecordHTTPRequest measures the per-request metrics path. It runs once
// per HTTP request, so the allocation count here is a per-request cost.
//
// Before: prometheus.Labels{...} built a map and re-hashed it on every call.
// After: WithLabelValues takes the label values directly, avoiding the map.
func BenchmarkRecordHTTPRequest(b *testing.B) {
	m := New()
	d := 5 * time.Millisecond

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.RecordHTTPRequest("/v1/devices/:imei", "GET", 200, d)
	}
}
