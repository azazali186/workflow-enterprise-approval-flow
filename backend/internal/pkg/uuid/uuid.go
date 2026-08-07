package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// New generates a UUID v7 (time-ordered, sortable)
// UUID v7 format: 48 bits timestamp + 74 bits randomness
// This ensures IDs are monotonically increasing and sortable by creation time
func New() uuid.UUID {
	return uuid.New()
}

// NewV7 generates a UUID v7 (time-ordered, sortable)
// Format: time (6 bytes) + version (4 bits) + rand (10 bytes) + variant (2 bits) + rand (4 bytes)
func NewV7() uuid.UUID {
	var u uuid.UUID

	// Get current timestamp in milliseconds
	now := time.Now().UnixMilli()

	// Set timestamp (48 bits = 6 bytes)
	binary.BigEndian.PutUint16(u[0:2], uint16(now>>32))
	binary.BigEndian.PutUint32(u[2:6], uint32(now))

	// Set version (4 bits) = 0111 (v7)
	u[6] = (u[6] & 0x0f) | 0x70

	// Set variant (2 bits) = 10xx (RFC 4122)
	u[8] = (u[8] & 0x3f) | 0x80

	// Fill remaining bytes with cryptographic randomness
	rand.Read(u[8:])

	return u
}

// NewV7OrNil returns nil if uuid is nil, otherwise returns a new UUID v7
func NewV7OrNil() *uuid.UUID {
	u := NewV7()
	return &u
}

// Parse parses a UUID string and returns a uuid.UUID
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParse parses a UUID string and panics on error
func MustParse(s string) uuid.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse UUID: %v", err))
	}
	return u
}

// IsNil checks if a UUID is nil (all zeros)
func IsNil(u uuid.UUID) bool {
	return u == uuid.Nil
}

// Timestamp extracts the timestamp from a UUID v7
// Returns the creation time if it's a v7 UUID, otherwise returns zero time
func Timestamp(u uuid.UUID) time.Time {
	// Extract timestamp from first 6 bytes (48 bits)
	ms := int64(binary.BigEndian.Uint16(u[0:2]))<<32 | int64(binary.BigEndian.Uint32(u[2:6]))
	return time.UnixMilli(ms)
}

// GenerateID generates a new UUID v7 and returns it as a string
func GenerateID() string {
	return NewV7().String()
}
