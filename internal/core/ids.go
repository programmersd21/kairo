package core

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
	"time"
)

// NewID returns a compact, time-ordered identifier.
// Format: 10 chars time + 16 chars entropy (Crockford-ish base32, no padding).
func NewID(now time.Time) (string, error) {
	var b [10]byte
	ms := uint64(now.UnixMilli())
	for i := 9; i >= 0; i-- {
		b[i] = byte(ms & 0xFF)
		ms >>= 8
	}
	var r [10]byte
	if _, err := rand.Read(r[:]); err != nil {
		return "", err
	}
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	s := enc.EncodeToString(append(b[:], r[:]...))
	s = strings.ToLower(s)
	return s, nil
}
