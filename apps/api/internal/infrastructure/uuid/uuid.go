// Package uuid provides UUIDv7 generation utilities for the infrastructure layer.
package uuid

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// New generates a new UUID using google/uuid library.
// Falls back to a simple string on error (non-panicking).
func New() string {
	id, err := uuid.NewUUID()
	if err != nil {
		return ""
	}
	return id.String()
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
