package database

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
)

const addressLength = 20

type AccountID string

type Accounts map[AccountID]Account

type Account struct {
	ID      AccountID
	Nonce   uint64
	Balance uint64
}

// ToAccountID is a constructor for a new AccountID.
// This function takes hex encoded string and verifies whether its
// underlying value conforms to the AccountID format requirements.
func ToAccountID(hexID string) (AccountID, error) {
	if err := verify(hexID); err != nil {
		return "", err
	}
	return AccountID(hexID), nil
}

// PubToAccountID is a constructor for a new AccountID.
// This function takes ECDSA public key and converts it to AccountID.
func PubToAccountID(pub ecdsa.PublicKey) (AccountID, error) {
	return ToAccountID(crypto.PubkeyToAddress(pub).String())
}

// Verify ensures the underlying value conforms to the AccountID format requirements.
func (id AccountID) Verify() error {
	if err := verify(string(id)); err != nil {
		return err
	}
	return nil
}

func (id AccountID) String() string {
	return string(id)
}

// verify ensures the format of given hex ID conforms to the requirements of AccountID format.
func verify(hexID string) error {
	if len(hexID) == 0 {
		return errors.New("invalid account ID format: value is empty")
	}
	if !has0xPrefix(hexID) {
		return errors.New("invalid account ID format: 0x prefix not found")
	}
	if !hasProperLen(hexID[2:]) {
		return errors.New("invalid account ID format: length is too small")
	}
	if !hasProperChars(hexID[2:]) {
		return errors.New("invalid account ID format: invalid char found")
	}
	return nil
}

func has0xPrefix(hexID string) bool {
	return hexID[0] == '0' && (hexID[1] == 'x' || hexID[1] == 'X')
}

func hasProperLen(hexID string) bool {
	return len(hexID) == 2*addressLength
}

func hasProperChars(hexID string) bool {
	for _, b := range []byte(hexID) {
		if !isHexChar(b) {
			return false
		}
	}
	return true
}

func isHexChar(b byte) bool {
	return ('0' <= b && b <= '9') || ('a' <= b && b <= 'f') || ('A' <= b && b <= 'F')
}
