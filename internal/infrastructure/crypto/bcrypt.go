package crypto

import "golang.org/x/crypto/bcrypt"

// BcryptHasher hashes and verifies passwords using bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a bcrypt password hasher.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(plain string) (string, error) {
	value, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func (h *BcryptHasher) Compare(hash string, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
