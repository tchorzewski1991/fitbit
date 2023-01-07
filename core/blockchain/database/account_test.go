package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
)

func TestToAccountID(t *testing.T) {
	hexID := "0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0"
	accountID, err := database.ToAccountID(hexID)
	assert.Nil(t, err)

	expectedAccountID := database.AccountID(hexID)
	assert.Equal(t, accountID, expectedAccountID)

	// Ensures hexID is not empty
	hexID = ""
	_, err = database.ToAccountID(hexID)
	assert.EqualError(t, err, "invalid account ID format: value is empty")

	// Ensures hexID contains 0x prefix
	hexID = "0ee5ba68586c85880B0900D0dEe0eEcBB37010e0"
	_, err = database.ToAccountID(hexID)
	assert.EqualError(t, err, "invalid account ID format: 0x prefix not found")

	// Ensures hexID has valid length
	hexID = "0x0"
	_, err = database.ToAccountID(hexID)
	assert.EqualError(t, err, "invalid account ID format: length is too small")

	// Ensures hexID contains valid chars (first char after 0x)
	hexID = "0xHee5ba68586c85880B0900D0dEe0eEcBB37010e0"
	_, err = database.ToAccountID(hexID)
	assert.EqualError(t, err, "invalid account ID format: invalid char found")
}

func TestAccountID_Verify(t *testing.T) {
	hexID := "0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0"
	accountID := database.AccountID(hexID)
	err := accountID.Verify()
	assert.Nil(t, err)

	hexID = "0ee5ba68586c85880B0900D0dEe0eEcBB37010e0"
	accountID = database.AccountID(hexID)
	err = accountID.Verify()
	assert.EqualError(t, err, "invalid account ID format: 0x prefix not found")
}
