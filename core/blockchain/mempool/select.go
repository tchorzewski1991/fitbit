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
