package mempool_test

import (
	"crypto/ecdsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/mempool"
	"github.com/tchorzewski1991/fitbit/core/blockchain/testdata"
)

func TestMempool(t *testing.T) {
	// Setup test data
	priv := testdata.LoadPrivateKey(t)

	from, err := database.PubToAccountID(priv.PublicKey)
	assert.Nil(t, err)

	to, err := database.ToAccountID("0x0ce5ba68586c85880B0900D0dEe0eEcBB33040e0")
	assert.Nil(t, err)

	blockTx1 := prepareBlockTx(t, blockTxArgs{
		priv:  priv,
		from:  from,
		to:    to,
		nonce: 1,
	})
	blockTx2 := prepareBlockTx(t, blockTxArgs{
		priv:  priv,
		from:  from,
		to:    to,
		nonce: 2,
	})

	// Run test

	m := mempool.New()
	assert.Equal(t, 0, m.Size())

	err = m.Upsert(blockTx1)
	assert.Nil(t, err)

	err = m.Upsert(blockTx2)
	assert.Nil(t, err)

	assert.Equal(t, 2, m.Size())

	txs := m.Select(mempool.SelectAll())
	assert.Equal(t, 2, len(txs))
	assert.Equal(t, blockTx1.Timestamp, txs[0].Timestamp)
	assert.Equal(t, blockTx2.Timestamp, txs[1].Timestamp)

	err = m.Remove(blockTx1)
	assert.Nil(t, err)

	assert.Equal(t, 1, m.Size())

	m.Truncate()
	assert.Equal(t, 0, m.Size())
}

type blockTxArgs struct {
	priv  *ecdsa.PrivateKey
	from  database.AccountID
	to    database.AccountID
	nonce uint64
}

func prepareBlockTx(t *testing.T, args blockTxArgs) database.BlockTx {
	tx := database.Tx{
		ChainID: 1,
		Nonce:   args.nonce,
		From:    args.from,
		To:      args.to,
		Value:   100,
		Tip:     10,
	}
	signedTx, err := tx.Sign(args.priv)
	assert.Nil(t, err)

	return database.NewBlockTx(signedTx, 0, 0)
}
