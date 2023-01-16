package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/testdata"
)

const defaultChainID = uint16(1)

func TestTx_SignAndVerify(t *testing.T) {
	priv := testdata.LoadPrivateKey(t)
	params := defaultTxParams(t)

	// Build and sign valid tx with nil private key
	tx := buildTx(&params)
	_, err := tx.Sign(nil)
	assert.EqualError(t, err, "tx sign err: signature sign err: private key is mandatory")

	// Build and sign valid tx with proper private key
	signedTx, err := tx.Sign(priv)
	assert.Nil(t, err)

	// Verify signed tx with default chain ID
	err = signedTx.Verify(defaultChainID)
	assert.Nil(t, err)
}

func TestTx_SignAndVerifyChainID(t *testing.T) {
	priv := testdata.LoadPrivateKey(t)
	params := defaultTxParams(t)

	// Build and sign tx with default chain ID
	tx := buildTx(&params)
	signedTx, err := tx.Sign(priv)
	assert.Nil(t, err)

	// Verify signed tx with other chain ID
	err = signedTx.Verify(2)
	assert.EqualError(t, err, "tx chain ID: 1 is not valid")
}

func TestTx_SignAndVerifyFrom(t *testing.T) {
	priv := testdata.LoadPrivateKey(t)
	params := defaultTxParams(t)

	// Build and sign tx with empty from address
	params.from = ""
	tx := buildTx(&params)
	signedTx, err := tx.Sign(priv)
	assert.Nil(t, err)

	// Verify signed tx with default chain ID
	err = signedTx.Verify(defaultChainID)
	assert.EqualError(t, err, "tx from:  is not valid")
}

func TestTx_SignAndVerifyTo(t *testing.T) {
	priv := testdata.LoadPrivateKey(t)
	params := defaultTxParams(t)

	// Build and sign tx with empty to address
	params.to = ""
	tx := buildTx(&params)
	signedTx, err := tx.Sign(priv)
	assert.Nil(t, err)

	// Verify signed tx with default chain ID
	err = signedTx.Verify(defaultChainID)
	assert.EqualError(t, err, "tx to:  is not valid")
}

func TestTx_SignAndVerifyFromToEqual(t *testing.T) {
	priv := testdata.LoadPrivateKey(t)
	params := defaultTxParams(t)

	// Build and sign tx with from address the same as to address
	params.from = params.to
	tx := buildTx(&params)
	signedTx, err := tx.Sign(priv)
	assert.Nil(t, err)

	// Verify signed tx with default chain ID
	err = signedTx.Verify(defaultChainID)
	assert.EqualError(t, err, "cannot send from: 0x0ce5ba68586c85880B0900D0dEe0eEcBB33040e0 to: 0x0ce5ba68586c85880B0900D0dEe0eEcBB33040e0")
}

func TestTx_SignAndVerifySignatureAddress(t *testing.T) {
	priv := testdata.LoadPrivateKey(t)
	params := defaultTxParams(t)

	// Build and sign tx
	params.from = "0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e1"
	tx := buildTx(&params)
	signedTx, err := tx.Sign(priv)
	assert.Nil(t, err)

	// Verify signed tx with default chain ID
	err = signedTx.Verify(defaultChainID)
	assert.EqualError(t, err, "tx from: 0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e1 does not match signature address: 0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0")
}

// Helper functions

type txParams struct {
	chainID uint16
	from    database.AccountID
	to      database.AccountID
}

func defaultTxParams(t *testing.T) txParams {
	return txParams{
		chainID: defaultChainID,
		from:    loadAccountID(t, "0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0"),
		to:      loadAccountID(t, "0x0ce5ba68586c85880B0900D0dEe0eEcBB33040e0"),
	}
}

func buildTx(params *txParams) database.Tx {
	return database.Tx{
		ChainID: params.chainID,
		Nonce:   1,
		From:    params.from,
		To:      params.to,
		Value:   100,
		Tip:     10,
		Data:    nil,
	}
}

func loadAccountID(t *testing.T, hexID string) database.AccountID {
	id, err := database.ToAccountID(hexID)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
