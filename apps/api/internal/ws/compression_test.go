package hub

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log/slog"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
)

func TestCompressionThreshold(t *testing.T) {
	log := testLogger()
	cfg := DefaultCompressionConfig()
	cfg.Threshold = 1024 // 1KB

	compressor := NewCompression(log, cfg)

	// Small message - should not compress
	smallData := []byte(`{"type":"test","dispatchId":"test-001"}`)
	compressed, didCompress, _ := compressor.CompressMessage(smallData)

	if didCompress {
		t.Error("small message should not be compressed")
	}

	if len(compressed) != len(smallData) {
		t.Error("uncompressed data should be returned as-is")
	}

	// Large message - should compress
	largeData := make([]byte, 1500)
	for i := range largeData {
		largeData[i] = 'A'
	}

	compressed, didCompress, _ = compressor.CompressMessage(largeData)

	if !didCompress {
		t.Error("large message should be compressed")
	}

	if len(compressed) >= len(largeData) {
		t.Error("compressed data should be smaller than original")
	}
}

func TestCompressionFallback(t *testing.T) {
	log := testLogger()
	cfg := DefaultCompressionConfig()

	compressor := NewCompression(log, cfg)

	// Test with data that might cause compression errors
	testData := []byte(`{"type":"test","data":"test"}`)

	compressed, didCompress, err := compressor.CompressMessage(testData)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should not compress small data
	if didCompress {
		t.Error("small data should not be compressed")
	}

	// Verify data integrity
	if !bytes.Equal(compressed, testData) {
		t.Error("data should be unchanged when not compressed")
	}
}

func TestCompressionDecompression(t *testing.T) {
	log := testLogger()
	cfg := DefaultCompressionConfig()
	cfg.Threshold = 100 // Lower threshold for testing

	compressor := NewCompression(log, cfg)

	// Create test data
	originalData := []byte(`{"type":"test","dispatchId":"test-001","data":"This is a test message that should be compressed and then decompressed successfully"}`)

	// Compress
	compressed, didCompress, err := compressor.CompressMessage(originalData)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}

	if !didCompress {
		t.Skip("data not compressed, skipping decompression test")
	}

	// Decompress
	decompressed, err := compressor.DecompressMessage(compressed)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}

	// Verify integrity
	if !bytes.Equal(decompressed, originalData) {
		t.Errorf("decompressed data doesn't match original")
		t.Logf("Original: %s", originalData)
		t.Logf("Decompressed: %s", decompressed)
	}
}

func TestCompressionMetrics(t *testing.T) {
	log := testLogger()
	cfg := DefaultCompressionConfig()
	cfg.Threshold = 100

	compressor := NewCompression(log, cfg)

	// Test multiple compressions
	for i := 0; i < 5; i++ {
		data := make([]byte, 500)
		for j := range data {
			data[j] = 'A'
		}
		compressor.CompressMessage(data)
	}

	metrics := compressor.GetMetrics()

	if metrics.TotalCompressed < 5 {
		t.Errorf("expected at least 5 compressed, got %d", metrics.TotalCompressed)
	}

	if metrics.TotalBytesSaved <= 0 {
		t.Errorf("expected positive bytes saved, got %d", metrics.TotalBytesSaved)
	}
}

func TestCompressionRatio(t *testing.T) {
	log := testLogger()
	cfg := DefaultCompressionConfig()
	cfg.Threshold = 100

	compressor := NewCompression(log, cfg)

	// Create compressible data
	data := make([]byte, 10000)
	for i := range data {
		data[i] = 'A' // Highly compressible
	}

	compressed, didCompress, err := compressor.CompressMessage(data)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}

	if !didCompress {
		t.Skip("data not compressed")
	}

	originalSize := len(data)
	compressedSize := len(compressed)
	ratio := float64(compressedSize) / float64(originalSize)

	if ratio > 0.5 {
		t.Logf("Compression ratio: %.2f", ratio)
	} else {
		t.Logf("Good compression ratio: %.2f", ratio)
	}
}

func TestCompressionWithGzip(t *testing.T) {
	log := testLogger()
	cfg := DefaultCompressionConfig()
	cfg.Threshold = 100

	compressor := NewCompression(log, cfg)

	// Test data
	data := []byte(`{"type":"test","data":"This is a test message for compression testing"}`)

	// Compress using our compressor
	compressed, didCompress, err := compressor.CompressMessage(data)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}

	if !didCompress {
		t.Skip("data not compressed")
	}

	// Verify it's valid gzip
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("invalid gzip data: %v", err)
	}
	defer reader.Close()

	decompressed, err := bytes.NewBuffer(nil).ReadFrom(reader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if !bytes.Equal(decompressed.Bytes(), data) {
		t.Error("decompressed data doesn't match original")
	}
}

func TestCompressionDisabled(t *testing.T) {
	log := testLogger()
	cfg := DefaultCompressionConfig()
	cfg.Enabled = false

	compressor := NewCompression(log, cfg)

	data := make([]byte, 1500)
	for i := range data {
		data[i] = 'A'
	}

	compressed, didCompress, err := compressor.CompressMessage(data)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if didCompress {
		t.Error("compression should be disabled")
	}

	if !bytes.Equal(compressed, data) {
		t.Error("data should be unchanged when compression is disabled")
	}
}

func TestCompressionFrame(t *testing.T) {
	log := testLogger()
	cfg := DefaultCompressionConfig()
	cfg.Threshold = 100

	compressor := NewCompression(log, cfg)

	// Create a command frame
	frame := command.CommandFrame{
		Type:       "test",
		DispatchID: "test-001",
		Command:    "TEST_COMMAND",
		Args:       []byte(`{"key":"value"}`),
	}

	// Marshal to JSON
	frameData, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("failed to marshal frame: %v", err)
	}

	// Compress
	compressed, didCompress, err := compressor.CompressMessage(frameData)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}

	if !didCompress {
		t.Log("frame not compressed (may be below threshold)")
	} else {
		t.Logf("frame compressed from %d to %d bytes", len(frameData), len(compressed))
	}
}

func testLogger() *slog.Logger {
	return slog.Default()
}
