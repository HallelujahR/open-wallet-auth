package wallet

import (
	"encoding/hex"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

func TestEVMVerifierVerifyMessage(t *testing.T) {
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	address := publicKeyToAddress(privateKey.PubKey().SerializeUncompressed())
	message := "example.com wants you to sign in"

	compactSignature := ecdsa.SignCompact(privateKey, personalSignHash(message), false)
	signature := append([]byte{}, compactSignature[1:]...)
	signature = append(signature, compactSignature[0])

	ok, err := NewEVMVerifier().VerifyMessage(address, message, "0x"+hex.EncodeToString(signature))
	if err != nil {
		t.Fatalf("verify returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected signature to verify")
	}
}
