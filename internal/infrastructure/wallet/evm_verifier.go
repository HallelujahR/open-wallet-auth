package wallet

import (
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// EVMVerifier validates Ethereum-style addresses and personal_sign signatures.
// EVMVerifier 校验 EVM 地址和 personal_sign 签名。
type EVMVerifier struct{}

// NewEVMVerifier creates an EVM wallet verifier.
// NewEVMVerifier 创建 EVM 钱包签名校验器。
func NewEVMVerifier() *EVMVerifier {
	return &EVMVerifier{}
}

// NormalizeAddress validates and returns the checksum-form EVM address.
// NormalizeAddress 校验地址格式并返回 checksum 形式的 EVM 地址。
func (v *EVMVerifier) NormalizeAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	if !common.IsHexAddress(address) {
		return "", ErrInvalidAddress
	}
	return common.HexToAddress(address).Hex(), nil
}

// VerifyMessage recovers the signer from a personal_sign signature.
// VerifyMessage 从 personal_sign 签名中恢复签名地址并与请求地址比对。
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
// ErrInvalidAddress 表示钱包地址格式不合法。
var ErrInvalidAddress = errString("invalid wallet address")

// ErrInvalidSignature is returned when a wallet signature is malformed.
// ErrInvalidSignature 表示钱包签名格式不合法。
var ErrInvalidSignature = errString("invalid wallet signature")

type errString string

// Error returns the static wallet verifier error text.
// Error 返回钱包校验器的静态错误文本。
func (e errString) Error() string {
	return string(e)
}
