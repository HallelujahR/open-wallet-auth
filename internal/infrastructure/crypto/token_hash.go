package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

type SHA256TokenHasher struct{}

func NewSHA256TokenHasher() SHA256TokenHasher {
	return SHA256TokenHasher{}
}

func (SHA256TokenHasher) HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
