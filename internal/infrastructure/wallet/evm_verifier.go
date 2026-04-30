package wallet

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/sha3"
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
	raw := strings.TrimPrefix(address, "0x")
	raw = strings.TrimPrefix(raw, "0X")
	if len(raw) != 40 || !isHex(raw) {
		return "", ErrInvalidAddress
	}
	return checksumAddress(strings.ToLower(raw)), nil
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

	compactSig := make([]byte, 65)
	compactSig[0] = sig[64] + 27
	copy(compactSig[1:33], sig[:32])
	copy(compactSig[33:], sig[32:64])

	pubKey, _, err := ecdsa.RecoverCompact(compactSig, personalSignHash(message))
	if err != nil {
		return false, err
	}
	recovered := publicKeyToAddress(pubKey.SerializeUncompressed())
	return strings.EqualFold(recovered, normalized), nil
}

// personalSignHash returns the EIP-191 hash used by eth_personalSign.
// personalSignHash 返回 eth_personalSign 使用的 EIP-191 消息哈希。
func personalSignHash(message string) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	return keccak256([]byte(prefix + message))
}

// publicKeyToAddress derives an EIP-55 checksum address from an uncompressed public key.
// publicKeyToAddress 从未压缩公钥推导 EIP-55 checksum 地址。
func publicKeyToAddress(uncompressed []byte) string {
	hash := keccak256(uncompressed[1:])
	return checksumAddress(hex.EncodeToString(hash[12:]))
}

// checksumAddress formats a lowercase hex address using the EIP-55 checksum.
// checksumAddress 按 EIP-55 checksum 格式化小写十六进制地址。
func checksumAddress(lowerHex string) string {
	hash := hex.EncodeToString(keccak256([]byte(lowerHex)))
	var builder strings.Builder
	builder.WriteString("0x")
	for i, char := range lowerHex {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
			continue
		}
		nibble, _ := strconv.ParseUint(hash[i:i+1], 16, 8)
		if nibble >= 8 {
			builder.WriteString(strings.ToUpper(string(char)))
			continue
		}
		builder.WriteRune(char)
	}
	return builder.String()
}

// keccak256 hashes data with Ethereum's legacy Keccak-256 variant.
// keccak256 使用以太坊采用的 legacy Keccak-256 变体计算哈希。
func keccak256(data []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write(data)
	return hasher.Sum(nil)
}

// isHex reports whether a string contains only hexadecimal characters.
// isHex 判断字符串是否只包含十六进制字符。
func isHex(value string) bool {
	for _, char := range value {
		if (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F') {
			continue
		}
		return false
	}
	return true
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
