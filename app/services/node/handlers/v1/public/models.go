package public

import (
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
)

// account represents the details of the account which
// will be serialized and moved over the wire.
type account struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Nonce   uint64 `json:"nonce"`
	Balance uint64 `json:"balance"`
}

func toAccount(h Handlers, dbAccount database.Account) account {
	return account{
		ID:      string(dbAccount.ID),
		Name:    h.NameService.FindName(dbAccount.ID),
		Nonce:   dbAccount.Nonce,
		Balance: dbAccount.Balance,
	}
}

// uncommitedTx represents the details of the transaction which
// will be serialized and moved over the wire.
type uncommitedTx struct {
	From      database.AccountID `json:"from"`
	FromName  string             `json:"from_name"`
	To        database.AccountID `json:"to"`
	ToName    string             `json:"to_name"`
	ChainID   uint16             `json:"chain_id"`
	Nonce     uint64             `json:"nonce"`
	Value     uint64             `json:"value"`
	Data      []byte             `json:"data"`
	Tip       uint64             `json:"tip"`
	GasPrice  uint64             `json:"gas_price"`
	GasUnits  uint64             `json:"gas_units"`
	Signature string             `json:"signature"`
	Timestamp uint64             `json:"timestamp"`
}

func toUncommittedTx(h Handlers, dbTx database.BlockTx) uncommitedTx {
	return uncommitedTx{
		From:      dbTx.From,
		FromName:  h.NameService.FindName(dbTx.From),
		To:        dbTx.To,
		ToName:    h.NameService.FindName(dbTx.To),
		ChainID:   dbTx.ChainID,
		Nonce:     dbTx.Nonce,
		Value:     dbTx.Value,
		Data:      dbTx.Data,
		Tip:       dbTx.Tip,
		GasPrice:  dbTx.GasPrice,
		GasUnits:  dbTx.GasUnits,
		Signature: dbTx.Signature(),
		Timestamp: dbTx.Timestamp,
	}
}
