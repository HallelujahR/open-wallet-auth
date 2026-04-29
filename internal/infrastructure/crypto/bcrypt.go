package crypto

import "golang.org/x/crypto/bcrypt"

// BcryptHasher hashes and verifies passwords using bcrypt.
// BcryptHasher 使用 bcrypt 完成密码哈希和校验。
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a bcrypt password hasher.
// NewBcryptHasher 创建 bcrypt 密码哈希器，并在 cost 为空时使用默认值。
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// Hash returns a bcrypt hash for a plain password.
// Hash 返回明文密码对应的 bcrypt 哈希。
func (h *BcryptHasher) Hash(plain string) (string, error) {
	value, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// Compare verifies whether a plain password matches a bcrypt hash.
// Compare 校验明文密码是否匹配 bcrypt 哈希。
func (h *BcryptHasher) Compare(hash string, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
