package signature

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/goccy/go-json"
)

// ZeroHash represents a hash code of zeros.
const ZeroHash string = "0x0000000000000000000000000000000000000000000000000000000000000000"

// fitbitID represents the unique value added to every v value of the ECDSA signature.
const fitbitID = 23

// Sign signs any value using given private key.
// Function returns r, s, v representation of the signature where:
//   - r represents first coordinate of the ecdsa signature
//   - s represents second coordinate of the ecdsa signature
//   - v represents message signing ID, either 0 or 1 (23 or 24 with fitbitID)
func Sign(value any, pk *ecdsa.PrivateKey) (r, s, v *big.Int, err error) {

	if pk == nil {
		return nil, nil, nil, fmt.Errorf("signature sign err: %w", errors.New("private key is mandatory"))
	}

	// Marshal value to the raw bytes (we need this for Keccak256 hash function)
	data, err := json.Marshal(value)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("signature sign err: %w", err)
	}

	// We want to stamp our data to make sure it is our system that has produced the hash
	stamp := []byte(fmt.Sprintf("\x19Fitbit Signed Message:\n%d", len(data)))

	// Hash data with Keccak256 hash function to the fixed 32 bytes array
	hash := crypto.Keccak256(stamp, data)

	// Sign hashed data with the private key
	sig, err := crypto.Sign(hash, pk)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("signature sign err: %w", err)
	}

	// Derive public key associated with the private key
	pub := pk.Public()

	// Ensure derived public key is of *ecdsa.PublicKey type
	_, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, nil, fmt.Errorf("signature sign err: %w", errors.New("public key needs to be ecdsa"))
	}

	// Convert produced signature to r,s,v format
	r, s, v, err = FromBytes(sig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("signature sign err: %w", err)
	}

	return
}

// Verify takes the r, s, v signature representation and verifies its correctness.
//   - r represents first coordinate of the ecdsa signature
//   - s represents second coordinate of the ecdsa signature
//   - v represents message signing ID, either 0 or 1 (23 or 24 with fitbitID)
func Verify(r, s, v *big.Int) error {

	// Message signing ID should be either 0 or 1.
	uint64V := v.Uint64() - fitbitID
	if uint64V != 0 && uint64V != 1 {
		return errors.New("message signing id is invalid")
	}

	// Validate correctness of the given signature coordinates
	if !crypto.ValidateSignatureValues(byte(uint64V), r, s, false) {
		return errors.New("signature is invalid")
	}

	return nil
}

// RecoverAddress takes any value and r, s, v signature representation and recovers
// Etherium account address associated with that signature.
func RecoverAddress(value any, r, s, v *big.Int) (common.Address, error) {

	// Marshal value to the raw bytes (we need this for Keccak256 hash function).
	data, err := json.Marshal(value)
	if err != nil {
		return common.Address{}, fmt.Errorf("recover pubkey err: %w", err)
	}

	// We want to stamp our data to make sure it is our system that has produced the hash.
	stamp := []byte(fmt.Sprintf("\x19Fitbit Signed Message:\n%d", len(data)))

	// Hash data with Keccak256 hash function to the fixed 32 bytes array.
	hash := crypto.Keccak256(stamp, data)

	// Convert r, s, v signature format to the []byte signature representation.
	sig, err := ToBytes(r, s, v)
	if err != nil {
		return common.Address{}, fmt.Errorf("recover pubkey err: %w", err)
	}

	// Extract public key associated with the hash and the signature.
	pub, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return common.Address{}, fmt.Errorf("recover pubkey err: %w", err)
	}

	return crypto.PubkeyToAddress(*pub), err
}

// Hash returns unique representation of the value.
func Hash(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ZeroHash
	}

	hash := sha256.Sum256(data)

	return hexutil.Encode(hash[:])
}

// ToBytes takes r, s, v signature format and converts it to the []byte format.
func ToBytes(r, s, v *big.Int) ([]byte, error) {
	if r == nil || s == nil || v == nil {
		return nil, errors.New("cannot convert signature to bytes")
	}

	sig := make([]byte, crypto.SignatureLength)

	rbs := make([]byte, 32)
	r.FillBytes(rbs)
	copy(sig[:32], rbs)

	sbs := make([]byte, 32)
	s.FillBytes(sbs)
	copy(sig[32:], sbs)

	sig[64] = byte(v.Uint64() - fitbitID)

	return sig, nil
}

// FromBytes takes signature in []byte format and converts it to r, s, v format.
func FromBytes(sig []byte) (r, s, v *big.Int, err error) {
	if len(sig) < 64 {
		return nil, nil, nil, errors.New("cannot convert signature to r, s, v")
	}

	r = big.NewInt(0).SetBytes(sig[:32])
	s = big.NewInt(0).SetBytes(sig[32:64])
	v = big.NewInt(0).SetBytes([]byte{sig[64] + fitbitID})

	return
}
