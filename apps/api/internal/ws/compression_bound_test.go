package hub

import (
	"bytes"
	"compress/gzip"
	"testing"
)

// TestDecompressMessage_RejectsBomb proves a gzip payload that expands past the cap
// is rejected rather than materialised in memory.
func TestDecompressMessage_RejectsBomb(t *testing.T) {
	c := NewCompression(testLogger(), DefaultCompressionConfig())

	// ~20 MiB of zeros compresses to a few KB; it must not be expanded.
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(make([]byte, MaxDecompressedSize+1024)); err != nil {
		t.Fatalf("compress: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// The compressed form is small (bomb-shaped).
	if buf.Len() > MaxDecompressedSize/100 {
		t.Fatalf("expected a small compressed payload, got %d bytes", buf.Len())
	}

	if _, err := c.DecompressMessage(buf.Bytes()); err == nil {
		t.Fatal("a payload exceeding the cap must be rejected")
	}
}

// TestDecompressMessage_AllowsNormalTraffic proves the cap does not break ordinary
// messages.
func TestDecompressMessage_AllowsNormalTraffic(t *testing.T) {
	c := NewCompression(testLogger(), DefaultCompressionConfig())
	payload := bytes.Repeat([]byte(`{"type":"telemetry","riskScore":42}`), 100) // > 1KB threshold

	compressed, did, err := c.CompressMessage(payload)
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if !did {
		t.Skip("payload below compression threshold")
	}

	out, err := c.DecompressMessage(compressed)
	if err != nil {
		t.Fatalf("normal payload should decompress: %v", err)
	}
	if !bytes.Equal(out, payload) {
		t.Fatalf("round-trip mismatch: got %q", out)
	}
}

// TestDecompressMessage_NonGzipPassthrough proves uncompressed input is returned
// unchanged.
func TestDecompressMessage_NonGzipPassthrough(t *testing.T) {
	c := NewCompression(testLogger(), DefaultCompressionConfig())
	raw := []byte("not gzip")

	out, err := c.DecompressMessage(raw)
	if err != nil {
		t.Fatalf("passthrough should not error: %v", err)
	}
	if !bytes.Equal(out, raw) {
		t.Fatal("non-gzip input must be returned unchanged")
	}
}
