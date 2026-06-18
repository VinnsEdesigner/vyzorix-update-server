// Package storage provides SQLite database operations.
package storage

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrUUIDGenerationFailed is returned when UUID generation fails.
var ErrUUIDGenerationFailed = errors.New("failed to generate UUID: crypto/rand unavailable")

var (
	uuidMu         sync.Mutex
	lastTimestamp  int64
	counter        uint16
)

// NewUUIDv7 generates a new UUIDv7 string using google/uuid library.
// Thread-safe and monotonically increasing within the same millisecond.
// Falls back to standard UUID v4 if crypto/rand is unavailable (should never happen in production).
func NewUUIDv7() string {
	uuidMu.Lock()
	defer uuidMu.Unlock()

	now := time.Now().UnixMilli()

	// Reset counter if we're in a new millisecond
	if now != lastTimestamp {
		lastTimestamp = now
		counter = 0
	} else {
		// Increment counter for same-millisecond UUIDs
		counter++
		// If counter overflows, spin until the next millisecond
		if counter == 0 {
			for now == lastTimestamp {
				time.Sleep(time.Microsecond)
				now = time.Now().UnixMilli()
			}
			lastTimestamp = now
			counter = 0
		}
	}

	// Generate random bytes for the random portion (only 8 bytes needed for groups 4+5)
	var randBytes [8]byte
	if _, err := rand.Read(randBytes[:]); err != nil {
		// In production, crypto/rand NEVER fails. If it does, the system is in a critical
		// state. We panic rather than falling back to predictable random.
		panic(ErrUUIDGenerationFailed)
	}

	// Build UUIDv7.
	// UUID format: xxxxxxxx-xxxx-7xxx-yxxx-xxxxxxxxxxxx
	// timestamp_hi: top 32 bits of 48-bit timestamp.
	// timestamp_mid: bottom 16 bits of 48-bit timestamp.
	timestampHi := uint32(now >> 16) // #nosec G115 -- conversion from int64 to uint32 is safe here since timestamp is a Unix timestamp in milliseconds.
	timestampMid := uint16(now & 0xffff) // #nosec G115 -- bottom 16 bits of a positive timestamp fit in uint16.

	// version 7 in the 4 most significant bits of timestamp_hi.
	versionedTimeHi := (timestampHi << 4) | 0x7

	// variant 1 (10xx) in the 2 most significant bits of rand_a (first byte of group 4).
	variantRandA := (randBytes[0] & 0x3f) | 0x80

	// Group 3: version 7 (4 bits) + counter (12 bits)
	group3 := (0x7000 | (counter & 0x0fff))

	// Group 4: variant (2 bits) + rand_a (8 bits) = 10 bits from randBytes[0], 6 bits from randBytes[1]
	group4 := uint16(variantRandA)<<8 | uint16(randBytes[1])

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%02x%02x%02x%02x%02x%02x",
		versionedTimeHi,
		timestampMid,
		group3,
		group4,
		randBytes[2],
		randBytes[3],
		randBytes[4],
		randBytes[5],
		randBytes[6],
		randBytes[7],
	)
}

// MustNewUUIDv7 generates a new UUIDv7 or panics if generation fails.
// Use this when failure to generate a UUID is a fatal error.
func MustNewUUIDv7() string {
	uuidStr := NewUUIDv7()
	if uuidStr == "" {
		panic("MustNewUUIDv7: generated empty UUID")
	}
	return uuidStr
}

// ParseUUIDv7 validates that a string is a properly formatted UUIDv7.
// Returns nil if valid, error if invalid.
func ParseUUIDv7(s string) error {
	if len(s) != 36 {
		return fmt.Errorf("uuid: invalid length %d", len(s))
	}

	// Check format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return fmt.Errorf("uuid: invalid hyphen position at %d", i)
			}
			continue
		}
		if !isHexChar(c) {
			return fmt.Errorf("uuid: invalid hex char at %d", i)
		}
	}

	// Check version (character after 2nd hyphen, bits 12-15 should be 7)
	// Format: xxxxxxxx-xxxx-Mxxx-yxxx-xxxxxxxxxxxx
	// M is at index 14 (0-based)
	versionNibble := hexCharToInt(rune(s[14]))
	if versionNibble != 7 {
		return fmt.Errorf("uuid: invalid version %d, expected 7", versionNibble)
	}

	// Check variant (first char of 4th group, bits 17-18 should be 10xx)
	// y is at index 19 (0-based)
	variantNibble := hexCharToInt(rune(s[19]))
	if (variantNibble & 0xc) != 0x8 {
		return fmt.Errorf("uuid: invalid variant")
	}

	return nil
}

// ParseUUID validates that a string is a properly formatted UUID (any version).
// Uses google/uuid library for validation.
func ParseUUID(s string) error {
	_, err := uuid.Parse(s)
	return err
}

func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func hexCharToInt(c rune) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return 0
	}
}

// UUIDv7ToTime extracts the Unix timestamp (milliseconds) from a UUIDv7.
func UUIDv7ToTime(uuidStr string) (time.Time, error) {
	if err := ParseUUIDv7(uuidStr); err != nil {
		return time.Time{}, err
	}

	// Extract timestamp from first 12 characters (excluding hyphen)
	// Format: xxxxxxxx-xxxx-xxxx...
	// timestamp is encoded in first 48 bits (12 hex chars)
	// The version nibble is in bits 12-15, so we shift right by 4 to get the actual timestamp.
	timestampHex := uuidStr[0:8] + uuidStr[9:13]

	var timestamp int64
	_, err := fmt.Sscanf(timestampHex, "%012x", &timestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("uuid: failed to parse timestamp: %w", err)
	}

	// Shift right by 4 to remove version bits at the top of the 48-bit value.
	// Since timestamps are positive (Unix milliseconds), this is safe.
	timestamp = timestamp >> 4

	return time.UnixMilli(timestamp), nil
}

// IsUUIDv7 returns true if the string is a valid UUIDv7.
func IsUUIDv7(s string) bool {
	return ParseUUIDv7(s) == nil
}

// IsValidUUID returns true if the string is a valid UUID (any version).
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}