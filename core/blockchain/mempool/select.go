package mempool

import "github.com/tchorzewski1991/fitbit/core/blockchain/database"

// SelectFunc allows for defining mempool selection strategy in easy and composable way.
type SelectFunc func(tx database.BlockTx) bool

// SelectAll is a mempool selection strategy for selecting all transactions.
var SelectAll = func() SelectFunc {
	return func(tx database.BlockTx) bool {
		return true
	}
}

// SelectByAccount is a mempool selection strategy for selecting account transactions.
var SelectByAccount = func(accountID database.AccountID) SelectFunc {
	return func(tx database.BlockTx) bool {
		if tx.From != accountID || tx.To != accountID {
			return false
		}
		return true
	}
}

// byNonce provides sorting support by the transaction nonce value.
type byNonce []database.BlockTx

func (bn byNonce) Len() int {
	return len(bn)
}

func (bn byNonce) Less(i, j int) bool {
	return bn[i].Nonce < bn[j].Nonce
}

func (bn byNonce) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}
