// Package uuid provides UUIDv7 generation utilities for the infrastructure layer.
package uuid

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrGenerationFailed is returned when UUID generation fails.
var ErrGenerationFailed = errors.New("failed to generate UUID: crypto/rand unavailable")

var (
	uuidMu        sync.Mutex
	lastTimestamp int64
	counter       uint16
)

// New generates a new UUIDv7 string using google/uuid library.
// Thread-safe and monotonically increasing within the same millisecond.
// Panics if crypto/rand is unavailable (critical security failure).
func New() string {
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

	// Generate random bytes for the random portion
	var randBytes [8]byte
	if _, err := rand.Read(randBytes[:]); err != nil {
		panic(ErrGenerationFailed)
	}

	// Build UUIDv7
	timestampHi := uint32(now >> 16)
	timestampMid := uint16(now & 0xffff)
	versionedTimeHi := (timestampHi << 4) | 0x7
	variantRandA := (randBytes[0] & 0x3f) | 0x80
	group3 := (0x7000 | (counter & 0x0fff))
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

// Parse validates that a string is a properly formatted UUID.
func Parse(s string) error {
	_, err := uuid.Parse(s)
	return err
}

// IsValid returns true if the string is a valid UUID.
func IsValid(s string) bool {
	return Parse(s) == nil
}

// ExtractTime extracts the Unix timestamp from a UUIDv7.
func ExtractTime(uuidStr string) (time.Time, error) {
	if err := parseV7(uuidStr); err != nil {
		return time.Time{}, err
	}

	timestampHex := uuidStr[0:8] + uuidStr[9:13]
	var timestamp int64
	_, err := fmt.Sscanf(timestampHex, "%012x", &timestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("uuid: failed to parse timestamp: %w", err)
	}

	return time.UnixMilli(timestamp >> 4), nil
}

func parseV7(s string) error {
	if len(s) != 36 {
		return fmt.Errorf("uuid: invalid length %d", len(s))
	}

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

	versionNibble := hexCharToInt(rune(s[14]))
	if versionNibble != 7 {
		return fmt.Errorf("uuid: invalid version %d, expected 7", versionNibble)
	}

	variantNibble := hexCharToInt(rune(s[19]))
	if (variantNibble & 0xc) != 0x8 {
		return fmt.Errorf("uuid: invalid variant")
	}

	return nil
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
