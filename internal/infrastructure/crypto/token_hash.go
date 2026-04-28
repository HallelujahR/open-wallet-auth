package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256TokenHasher hashes opaque tokens before persistence.
type SHA256TokenHasher struct{}

// NewSHA256TokenHasher creates a SHA-256 token hasher.
func NewSHA256TokenHasher() SHA256TokenHasher {
	return SHA256TokenHasher{}
}

func (SHA256TokenHasher) HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
