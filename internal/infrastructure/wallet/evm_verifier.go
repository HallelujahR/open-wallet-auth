package wallet

import (
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// EVMVerifier validates Ethereum-style addresses and personal_sign signatures.
type EVMVerifier struct{}

// NewEVMVerifier creates an EVM wallet verifier.
func NewEVMVerifier() *EVMVerifier {
	return &EVMVerifier{}
}

func (v *EVMVerifier) NormalizeAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	if !common.IsHexAddress(address) {
		return "", ErrInvalidAddress
	}
	return common.HexToAddress(address).Hex(), nil
}

func (v *EVMVerifier) VerifyMessage(address string, message string, signature string) (bool, error) {
	normalized, err := v.NormalizeAddress(address)
	if err != nil {
		return false, err
	}

	rawSig := strings.TrimPrefix(strings.TrimSpace(signature), "0x")
	sig, err := hex.DecodeString(rawSig)
	if err != nil || len(sig) != 65 {
		return false, ErrInvalidSignature
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	pubKey, err := crypto.SigToPub(accounts.TextHash([]byte(message)), sig)
	if err != nil {
		return false, err
	}
	recovered := crypto.PubkeyToAddress(*pubKey).Hex()
	return strings.EqualFold(recovered, normalized), nil
}

// ErrInvalidAddress is returned when a wallet address is malformed.
var ErrInvalidAddress = errString("invalid wallet address")

// ErrInvalidSignature is returned when a wallet signature is malformed.
var ErrInvalidSignature = errString("invalid wallet signature")

type errString string

func (e errString) Error() string {
	return string(e)
}
