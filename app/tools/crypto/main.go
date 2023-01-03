package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

func main() {

	// Client side ====================================================================

	// Generate a new private key
	//priv, err := crypto.GenerateKey()

	// Save the private key to disk
	//err = crypto.SaveECDSA("data/accounts/test.ecdsa", priv)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// Load ecdsa private key from disk
	priv, err := crypto.LoadECDSA("app/tools/crypto/test.ecdsa")
	if err != nil {
		log.Fatal(err)
	}

	// Generate public key out of private key
	pub := priv.PublicKey

	// Derive 20 bytes account address from public key
	address := crypto.PubkeyToAddress(pub).String()
	log.Println("address before sign: ", address)

	// Prepare example tx that will be signed with the private key
	tx := struct {
		Test string
	}{
		Test: "test",
	}

	// Marshal tx to the raw bytes (we need this for Keccak256 hash function)
	txData, err := json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}

	// We want to 'stamp' our data to make sure it is our system that has produced the hash
	stamp := []byte("\x19Fitbit Signed Message:\n32")

	// Hash tx bytes with Keccak256 hash function to the fixed size 32 bytes array
	txHash := crypto.Keccak256(stamp, txData)

	// Sign hashed tx with the private key in order to produce tx signature
	txSig, err := crypto.Sign(txHash, priv)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("tx sig: ", txSig)

	// Server side ==========================================================================

	// Extract uncompressed public key from tx hashed data and tx signature
	unPub, err := crypto.Ecrecover(txHash, txSig)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare the same / different tx in order verify signature
	//invalidTx := struct {
	//	Test string
	//}{
	//	Test: "test!",
	//}
	validTx := struct {
		Test string
	}{
		Test: "test",
	}

	// Marshal tx to the raw bytes (we need this for Keccak256 hash function)
	txData, err = json.Marshal(validTx)
	if err != nil {
		log.Fatal(err)
	}

	stamp = []byte("\x19Fitbit Signed Message:\n32")

	// Hash tx bytes with Keccak256 hash function to the fixed size 32 bytes array
	txHash = crypto.Keccak256(stamp, txData)

	// Verify whether uncompressed pub key has created signature over digest.
	if !crypto.VerifySignature(unPub, txHash, txSig[:crypto.RecoveryIDOffset]) {
		log.Fatal("invalid signature")
	}

	// Unmarshal from uncompressed public key x,y points of the elliptic curve (magic)
	x, y := elliptic.Unmarshal(crypto.S256(), unPub)

	// Rebuild the public key from unmarshalled x,y points
	pub = ecdsa.PublicKey{
		Curve: crypto.S256(),
		X:     x,
		Y:     y,
	}

	// Try to convert public key into an account address
	address = crypto.PubkeyToAddress(pub).String()
	log.Println("address after sign: ", address)
}
