package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256TokenHasher hashes opaque tokens before persistence.
// SHA256TokenHasher 在不透明 token 入库前执行 SHA-256 单向哈希。
type SHA256TokenHasher struct{}

// NewSHA256TokenHasher creates a SHA-256 token hasher.
// NewSHA256TokenHasher 创建 SHA-256 token 哈希器。
func NewSHA256TokenHasher() SHA256TokenHasher {
	return SHA256TokenHasher{}
}

// HashToken returns the hex SHA-256 digest of an opaque token.
// HashToken 返回不透明 token 的十六进制 SHA-256 摘要。
func (SHA256TokenHasher) HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
