package testdata

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func LoadPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	priv, err := crypto.LoadECDSA("testdata/test.ecdsa")
	if err != nil {
		t.Fatal(err)
	}
	return priv
}
