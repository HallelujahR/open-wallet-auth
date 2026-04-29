package wallet

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestEVMVerifierVerifyMessage(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	message := "example.com wants you to sign in"

	signature, err := crypto.Sign(accounts.TextHash([]byte(message)), privateKey)
	if err != nil {
		t.Fatalf("sign message: %v", err)
	}
	signature[64] += 27

	ok, err := NewEVMVerifier().VerifyMessage(address, message, "0x"+hex.EncodeToString(signature))
	if err != nil {
		t.Fatalf("verify returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected signature to verify")
	}
}
