package signature_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchorzewski1991/fitbit/core/blockchain/signature"
	"github.com/tchorzewski1991/fitbit/core/blockchain/testdata"
)

const (
	from = "0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0"
)

func TestHash_Hashable(t *testing.T) {
	hashable := struct {
		X, Y int
	}{
		X: 1,
		Y: 2,
	}
	hash := signature.Hash(hashable)
	assert.Equal(t, "0x", hash[:2])
}

func TestHash_NonHashable(t *testing.T) {
	nonHashable := struct {
		X, Y chan int
	}{}
	hash := signature.Hash(nonHashable)
	assert.Equal(t, signature.ZeroHash, hash)
}

func TestSigning(t *testing.T) {
	priv := testdata.LoadPrivateKey(t)

	data := struct {
		Name string `json:"name"`
	}{
		Name: "John",
	}

	// Sign data with private key
	r, s, v, err := signature.Sign(data, priv)
	assert.Nil(t, err)

	// Verify r, s, v signature
	err = signature.Verify(r, s, v)
	assert.Nil(t, err)

	// Recover public key from data and r, s, v signature
	addr, err := signature.RecoverAddress(data, r, s, v)
	assert.Nil(t, err)

	// Match expected address with one extracted from public key
	assert.Equal(t, from, addr.String())

	// Convert r, s, v signature values to bytes format
	bs, err := signature.ToBytes(r, s, v)
	assert.Nil(t, err)

	// Convert bytes signature to r, s, v format
	recR, recS, recV, err := signature.FromBytes(bs)
	assert.Nil(t, err)

	assert.Equal(t, r, recR)
	assert.Equal(t, s, recS)
	assert.Equal(t, v, recV)
}
