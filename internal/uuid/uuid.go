package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// New returns a random UUID v4 string (RFC 4122, 8-4-4-4-12 hex with dashes).
func New() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%s",
		uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]),
		uint16(b[4])<<8|uint16(b[5]),
		uint16(b[6])<<8|uint16(b[7]),
		uint16(b[8])<<8|uint16(b[9]),
		hex.EncodeToString(b[10:16]),
	)
}
