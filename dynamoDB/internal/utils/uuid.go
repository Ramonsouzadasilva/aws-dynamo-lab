package utils

import (
	"crypto/rand"
	"fmt"
)

// GenerateUUID generates a UUID v4 string using cryptographically secure random bytes.
func GenerateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to time-based generation or just panic if entropy source is broken
		panic(err)
	}
	
	// Set the version (4) and variant (RFC4122)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
