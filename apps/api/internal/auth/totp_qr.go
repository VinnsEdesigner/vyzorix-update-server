// Package security provides QR code generation for TOTP enrollment using enterprise-grade.
// QR code encoding. This implementation uses the skip2/go-qrcode library which provides.
// full QR code generation with Reed-Solomon error correction, all encoding modes (byte,.
// alphanumeric, numeric), and proper mask patterns compliant with ISO/IEC 18004.
package auth

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/skip2/go-qrcode"
)

// QRCode represents a properly encoded QR code image with full error correction.
// Uses QR Code model 2 which supports all data types and error correction levels.
type QRCode struct {
	// Content is the original data encoded in the QR code.
	Content string
	// Level is the error correction level used.
	Level QRLevel
	// Size is the width/height in pixels.
	Size int
	// Pixels contains the QR code bitmap for rendering.
	// Each pixel is true for black, false for white.
	Pixels [][]bool
	// pngData is the pre-encoded PNG data for efficient transport.
	pngData []byte
}

// QRLevel represents QR code error correction levels.
// Higher levels provide more redundancy but require more space.
type QRLevel int

const (
	// LevelLow provides ~7% error recovery (recommended for small codes).
	LevelLow QRLevel = iota
	// LevelMedium provides ~15% error recovery (default, recommended).
	LevelMedium
	// LevelQuartile provides ~25% error recovery (good for outdoor/industrial).
	LevelQuartile
	// LevelHigh provides ~30% error recovery (best for dusty/damaged surfaces).
	LevelHigh
)

// String returns the string representation of the QR level.
func (l QRLevel) String() string {
	switch l {
	case LevelLow:
		return "L"
	case LevelMedium:
		return "M"
	case LevelQuartile:
		return "Q"
	case LevelHigh:
		return "H"
	default:
		return "M"
	}
}

// toQRCodeLevel converts our QRLevel to the library's QRLevel.
func (l QRLevel) toQRCodeLevel() qrcode.RecoveryLevel {
	switch l {
	case LevelLow:
		return qrcode.Low
	case LevelMedium:
		return qrcode.Medium
	case LevelQuartile:
		return qrcode.High
	case LevelHigh:
		return qrcode.Highest
	default:
		return qrcode.Medium
	}
}

// NewQRCode creates a QR code with the given size in pixels.
// The size should be a multiple of 4 for optimal alignment.
func NewQRCode(size int) *QRCode {
	// Ensure minimum size for scannable QR code (at least 21 modules).
	if size < 21 {
		size = 21
	}
	// Ensure size is even for proper pixel alignment.
	if size%2 != 0 {
		size++
	}
	pixels := make([][]bool, size)
	for i := range pixels {
		pixels[i] = make([]bool, size)
	}
	return &QRCode{
		Size:   size,
		Pixels: pixels,
	}
}

// Encode generates a proper QR code from the given data using byte encoding mode.
// This implements the full QR code specification with:.
// - Byte encoding mode (8-bit characters).
// - Reed-Solomon error correction at the specified level.
// - Automatic version selection based on data length.
// - Proper mask pattern selection for optimal scannability.
// - Format and version information per ISO/IEC 18004.
func (q *QRCode) Encode(data string) error {
	return q.EncodeWithLevel(data, LevelMedium)
}

// EncodeWithLevel generates a QR code with a specific error correction level.
// Higher levels (H > Q > M > L) provide more error recovery but require more space.
func (q *QRCode) EncodeWithLevel(data string, level QRLevel) error {
	if data == "" {
		return fmt.Errorf("cannot encode empty data")
	}

	// Update content and level.
	q.Content = data
	q.Level = level

	// Generate PNG using the enterprise-grade library.
	// This handles all encoding: data analysis, version selection, error correction,.
	// structure final message, module placement, mask pattern selection.
	pngData, err := qrcode.Encode(data, level.toQRCodeLevel(), q.Size)
	if err != nil {
		return fmt.Errorf("QR code encoding failed: %w", err)
	}
	q.pngData = pngData

	// Decode the PNG back to pixels for any custom rendering needs.
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return fmt.Errorf("failed to decode generated QR code: %w", err)
	}

	// Copy pixels to our struct for custom rendering.
	bounds := img.Bounds()
	if len(q.Pixels) != bounds.Dy() || len(q.Pixels[0]) != bounds.Dx() {
		// Resize pixel array if needed.
		q.Pixels = make([][]bool, bounds.Dy())
		for i := range q.Pixels {
			q.Pixels[i] = make([]bool, bounds.Dx())
		}
		q.Size = bounds.Dy()
	}

	// Convert image to pixel array (black = true, white = false).
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get the pixel color.
			r, g, b, a := img.At(x, y).RGBA()
			// Consider pixel black if not transparent and has sufficient darkness.
			// Using luminance calculation: Y = 0.299*R + 0.587*G + 0.114*B.
			luminance := (19595*r + 38470*g + 7471*b) >> 16
			// Black if luminance < 128 (midpoint of 0-65535).
			q.Pixels[y][x] = luminance < 32768 && a > 32768
		}
	}

	return nil
}

// PNG writes the QR code as a PNG image to the provided writer.
// This uses the pre-encoded PNG data for efficiency.
func (q *QRCode) PNG(w io.Writer) error {
	if q.pngData == nil {
		return fmt.Errorf("QR code not encoded, call Encode() first")
	}
	_, err := w.Write(q.pngData)
	return err
}

// PNGData returns the raw PNG image data.
func (q *QRCode) PNGData() ([]byte, error) {
	if q.pngData == nil {
		return nil, fmt.Errorf("QR code not encoded, call Encode() first")
	}
	return q.pngData, nil
}

// Image returns the QR code as an image.Image for integration with other imaging libraries.
func (q *QRCode) Image() (image.Image, error) {
	if q.pngData == nil {
		return nil, fmt.Errorf("QR code not encoded, call Encode() first")
	}
	img, _, err := image.Decode(bytes.NewReader(q.pngData))
	return img, err
}

// ImageWithPadding returns the QR code as an image with a white padding border.
// Padding is useful for display contexts where the QR code needs visual separation.
func (q *QRCode) ImageWithPadding(padding int) (image.Image, error) {
	if q.pngData == nil {
		return nil, fmt.Errorf("QR code not encoded, call Encode() first")
	}

	img, _, err := image.Decode(bytes.NewReader(q.pngData))
	if err != nil {
		return nil, err
	}

	// Create new image with padding.
	paddedSize := img.Bounds().Dx() + 2*padding
	padded := image.NewRGBA(image.Rect(0, 0, paddedSize, paddedSize))

	// Fill with white.
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < paddedSize; y++ {
		for x := 0; x < paddedSize; x++ {
			padded.Set(x, y, white)
		}
	}

	// Copy original QR code into padded image.
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			padded.Set(x+padding, y+padding, img.At(x, y))
		}
	}

	return padded, nil
}

// GenerateTOTPQRCode generates a properly encoded QR code image for TOTP enrollment.
// The URI format should follow the standard otpauth:// scheme:.
//
//	otpauth://totp/Issuer:AccountName?secret=BASE32SECRET&issuer=Issuer&algorithm=SHA1&digits=6&period=30.
//
// Parameters:.
//   - uri: The TOTP provisioning URI from ProvisioningURIFor().
//   - size: Output size in pixels (recommended 200-400 for mobile scanning).
//
// Returns:.
//   - QRCode with full error correction (LevelHigh) for reliable scanning.
//   - Error if encoding fails.
//
// The QR code uses:.
//   - Byte encoding mode for full character support.
//   - High error correction level (30% recovery) for damaged/dirty screens.
//   - Auto version selection based on URI length.
//   - Optimal mask pattern for scannability.
func GenerateTOTPQRCode(uri string, size int) (*QRCode, error) {
	// Validate input.
	if uri == "" {
		return nil, fmt.Errorf("provisioning URI is required")
	}

	// Validate URI format (should start with otpauth://).
	if len(uri) < 10 || uri[:8] != "otpauth:" {
		return nil, fmt.Errorf("invalid TOTP URI format: must start with otpauth://")
	}

	// Ensure recommended size for mobile scanning.
	// QR codes need sufficient modules for reliable scanning.
	if size < 150 {
		size = 150 // Minimum recommended size for mobile cameras
	}
	if size > 1024 {
		size = 1024 // Maximum reasonable size
	}

	// Create QR code with high error correction for TOTP use case.
	// TOTP codes are displayed on screens that may have glare, dirt, or be partially obscured.
	qr := NewQRCode(size)

	// Use highest error correction level (LevelHigh = ~30% recovery).
	// This is critical for:.
	// - Screens with reflections or glare.
	// - Cameras at angles.
	// - Motion blur during capture.
	// - Damaged or dirty screens.
	if err := qr.EncodeWithLevel(uri, LevelHigh); err != nil {
		return nil, fmt.Errorf("failed to generate TOTP QR code: %w", err)
	}

	return qr, nil
}

// GenerateTOTPQRCodeMedium generates a QR code with medium error correction.
// Use this when space is constrained and the QR code will be large enough.
// for reliable scanning without maximum redundancy.
func GenerateTOTPQRCodeMedium(uri string, size int) (*QRCode, error) {
	if uri == "" {
		return nil, fmt.Errorf("provisioning URI is required")
	}
	if size < 100 {
		size = 100
	}
	if size > 1024 {
		size = 1024
	}

	qr := NewQRCode(size)
	if err := qr.EncodeWithLevel(uri, LevelMedium); err != nil {
		return nil, fmt.Errorf("failed to generate TOTP QR code: %w", err)
	}

	return qr, nil
}