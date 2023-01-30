package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/genesis"
	"github.com/tchorzewski1991/fitbit/core/blockchain/storage/memory"
)

func TestDatabase_Reset(t *testing.T) {
	db, err := database.New(mockGenesis(), mockStorage(t))
	assert.Nil(t, err)

	// Check how many accounts we have before remove
	assert.Equal(t, 2, len(db.Accounts()))

	// Remove account1 on purpose
	err = db.Remove("0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0")
	assert.Nil(t, err)

	// Check how many accounts we have after remove
	assert.Equal(t, 1, len(db.Accounts()))

	// Reset db
	err = db.Reset()
	assert.Nil(t, err)

	// Check how many accounts we have after reset
	assert.Equal(t, 2, len(db.Accounts()))
}

func TestDatabase_Accounts(t *testing.T) {
	db, err := database.New(mockGenesis(), mockStorage(t))
	assert.Nil(t, err)

	expectedAccounts := database.Accounts{
		"0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0": database.Account{
			ID:      "0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0",
			Nonce:   0,
			Balance: 100,
		},
		"0x0ce5ba68586c85880B0900D0dEe0eEcBB33040e0": database.Account{
			ID:      "0x0ce5ba68586c85880B0900D0dEe0eEcBB33040e0",
			Nonce:   0,
			Balance: 200,
		},
	}
	actualAccounts := db.Accounts()
	assert.Equal(t, expectedAccounts, actualAccounts)

	// Modify actual accounts and make sure it is the copy of the original accounts
	actualAccounts["0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0"] = database.Account{ID: "modified"}

	assert.Equal(t, expectedAccounts, db.Accounts())
}

func TestDatabase_Account(t *testing.T) {
	db, err := database.New(mockGenesis(), mockStorage(t))
	assert.Nil(t, err)

	// Query for not existing account and assert the error.
	_, err = db.Account("not-existing-account-address")
	assert.EqualError(t, err, "account not found")

	// Query for existing account.
	expectedAccount := database.Account{
		ID:      "0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0",
		Nonce:   0,
		Balance: 100,
	}
	account, err := db.Account("0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0")
	assert.Nil(t, err)
	assert.Equal(t, expectedAccount, account)
}

// Helper functions

func mockGenesis() genesis.Genesis {
	return genesis.Genesis{
		Date:         time.Time{},
		ChainID:      1,
		TxPerBlock:   10,
		Difficulty:   1,
		MiningReward: 1,
		GasPrice:     1,
		Balances: map[string]uint64{
			"0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0": 100,
			"0x0ce5ba68586c85880B0900D0dEe0eEcBB33040e0": 200,
		},
	}
}

func mockStorage(t *testing.T) database.Storage {
	storage, err := memory.New()
	assert.Nil(t, err)
	return storage
}
