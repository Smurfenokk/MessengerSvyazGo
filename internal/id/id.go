package id

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// GenerateShortID generates an 8-character base32 ID using ULID.
// We use base32 (Crockford alphabet) instead of hex because:
// - Hex: 16^8 = 4.3B combinations, collision likely after ~77k records
// - Base32: 32^8 = 1.1T combinations, collision likely after ~1.3M records
// For a messenger, 1.1T combinations is sufficient for v1.
func GenerateShortID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()[:8]
}
