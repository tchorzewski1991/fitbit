package testdata

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func LoadPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	priv, err := crypto.HexToECDSA("1cc3ea5a590a4181339569fcd7771cb2d95f203143f116b4cb66556c058e59c4")
	if err != nil {
		t.Fatal(err)
	}
	return priv
}
