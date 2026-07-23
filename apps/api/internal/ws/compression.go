package hub

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
)

// CompressionConfig holds configuration for message compression.
type CompressionConfig struct {
	// Threshold is the minimum message size in bytes to trigger compression (default 1024)
	Threshold int
	// Level is the gzip compression level (default gzip.DefaultCompression)
	Level int
	// EnableCompression enables compression (default true)
	EnableCompression bool
}

// DefaultCompressionConfig returns the default compression configuration.
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		Threshold:         1024,
		Level:             gzip.DefaultCompression,
		EnableCompression: true,
	}
}

// CompressionMetrics holds compression metrics.
type CompressionMetrics struct {
	TotalCompressed   int64   `json:"totalCompressed"`
	TotalUncompressed int64   `json:"totalUncompressed"`
	TotalBypassed     int64   `json:"totalBypassed"`
	CompressionRatio  float64 `json:"compressionRatio"`
	OriginalBytes     int64   `json:"originalBytes"`
	CompressedBytes   int64   `json:"compressedBytes"`
}

// Compression provides GZIP compression for WebSocket messages.
type Compression struct {
	log     *slog.Logger
	config  *CompressionConfig
	metrics CompressionMetrics
	mu      sync.RWMutex
	
	gzipPool sync.Pool
}

// NewCompression creates a new Compression handler.
func NewCompression(log *slog.Logger, cfg *CompressionConfig) *Compression {
	if cfg == nil {
		cfg = DefaultCompressionConfig()
	}

	c := &Compression{
		log:    log,
		config: cfg,
	}

	
	c.gzipPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	return c
}

// CompressMessage compresses a message if it exceeds the threshold.
// Returns the compressed message and a boolean indicating if compression was applied.

func (c *Compression) CompressMessage(data []byte) ([]byte, bool, error) {
	if !c.config.EnableCompression || len(data) < c.config.Threshold {
		c.incrementBypassed()
		return data, false, nil
	}

	
	bufInterface := c.gzipPool.Get()
	var buf *bytes.Buffer
	if bufInterface != nil {
		var ok bool
		buf, ok = bufInterface.(*bytes.Buffer)
		if !ok {
			buf = new(bytes.Buffer)
		}
	} else {
		buf = new(bytes.Buffer)
	}
	buf.Reset()
	defer func() {
		buf.Reset()
		c.gzipPool.Put(buf)
	}()

	// Create a new gzip writer - lightweight allocation
	writer, err := gzip.NewWriterLevel(buf, c.config.Level)
	if err != nil {
		c.log.Warn("failed to create gzip writer", "err", err)
		c.incrementBypassed()
		return data, false, nil
	}

	if _, err := writer.Write(data); err != nil {
		c.log.Warn("failed to write data to gzip", "err", err)
		c.incrementBypassed()
		return data, false, nil
	}

	if err := writer.Close(); err != nil {
		c.log.Warn("failed to close gzip writer", "err", err)
		c.incrementBypassed()
		return data, false, nil
	}

	compressed := buf.Bytes()

	// Only use compressed if it's actually smaller
	if len(compressed) >= len(data) {
		c.incrementBypassed()
		return data, false, nil
	}

	c.incrementCompressed(len(data), len(compressed))

	// Calculate compression ratio for G4 verification
	return compressed, true, nil
}

// RecordCompression records compression statistics for monitoring.
func (c *Compression) RecordCompression(originalSize, compressedSize int) {
	// Metrics are already recorded in incrementCompressed
}

// DecompressMessage decompresses a GZIP compressed message.
func (c *Compression) DecompressMessage(data []byte) ([]byte, error) {
	// Check for gzip magic bytes
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		return data, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer func() { _ = reader.Close() }()

	return io.ReadAll(reader)
}

// CompressJSON compresses a JSON message if it exceeds the threshold.
func (c *Compression) CompressJSON(v interface{}) ([]byte, bool, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, false, err
	}

	return c.CompressMessage(data)
}

// ShouldCompress returns true if a message of the given size should be compressed.
func (c *Compression) ShouldCompress(size int) bool {
	return c.config.EnableCompression && size >= c.config.Threshold
}

// GetMetrics returns a copy of the current metrics.
func (c *Compression) GetMetrics() CompressionMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.metrics
}

func (c *Compression) incrementCompressed(original, compressed int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.TotalCompressed++
	c.metrics.OriginalBytes += int64(original)
	c.metrics.CompressedBytes += int64(compressed)

	if c.metrics.OriginalBytes > 0 {
		c.metrics.CompressionRatio = float64(c.metrics.CompressedBytes) / float64(c.metrics.OriginalBytes)
	}
}

func (c *Compression) incrementBypassed() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.TotalBypassed++
}

// IsCompressed checks if data is gzip compressed by looking for magic bytes.
func IsCompressed(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// CompressedFrame represents a WebSocket frame with compression metadata.
type CompressedFrame struct {
	Type         string `json:"type"`
	Data         []byte `json:"data"`
	OriginalSize int    `json:"originalSize"`
	Compressed   bool   `json:"compressed"`
}

// WrapCompressedFrame creates a compressed frame wrapper.
func WrapCompressedFrame(frameType string, data []byte, cfg *CompressionConfig) (*CompressedFrame, error) {
	comp := NewCompression(nil, cfg)

	compressed, didCompress, err := comp.CompressMessage(data)
	if err != nil {
		return nil, err
	}

	return &CompressedFrame{
		Type:         frameType,
		Compressed:   didCompress,
		OriginalSize: len(data),
		Data:         compressed,
	}, nil
}
