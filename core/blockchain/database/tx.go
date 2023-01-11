package database

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tchorzewski1991/fitbit/core/blockchain/signature"
)

// Tx represents transactional change between two accounts.
type Tx struct {
	ChainID uint64    `json:"chain_id"`
	Nonce   uint64    `json:"nonce"`
	From    AccountID `json:"from"`
	To      AccountID `json:"to"`
	Value   uint64    `json:"value"`
	Tip     uint64    `json:"tip"`
	Data    []byte    `json:"data"`
}

// Sign uses ECDSA private key to sign the Tx.
func (tx Tx) Sign(pk *ecdsa.PrivateKey) (SignedTx, error) {

	// Sign tx with given private key
	r, s, v, err := signature.Sign(tx, pk)
	if err != nil {
		return SignedTx{}, fmt.Errorf("tx sign err: %w", err)
	}

	return SignedTx{
		Tx: tx,
		R:  r,
		S:  s,
		V:  v,
	}, nil
}

// SignedTx represents signed version of the transaction.
// SignedTx is how clients like 3rd party wallets can include
// transaction into the fitbit blockchain.
type SignedTx struct {
	Tx
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`
	V *big.Int `json:"v"`
}

// Verify checks whether the transaction has a proper signature.
func (tx SignedTx) Verify(chainID uint64) error {

	if tx.ChainID != chainID {
		return fmt.Errorf("chain ID: %d is not valid", chainID)
	}

	if err := tx.From.Verify(); err != nil {
		return fmt.Errorf("tx from: %s is not valid", tx.From)
	}

	if err := tx.To.Verify(); err != nil {
		return fmt.Errorf("tx to: %s is not valid", tx.To)
	}

	if tx.From == tx.To {
		return fmt.Errorf("cannot send from: %s to: %s", tx.From, tx.To)
	}

	addr, err := signature.RecoverAddress(tx.Tx, tx.R, tx.S, tx.V)
	if err != nil {
		return errors.New("cannot recover address from signature")
	}

	from, _ := ToAccountID(addr.String())

	if tx.From != from {
		return fmt.Errorf("tx from: %s does not match signature address: %s", tx.From, from)
	}

	return nil
}

// BlockTx represents transaction that will be saved to the blockchain.
type BlockTx struct {
	SignedTx
	Timestamp uint64 `json:"timestamp"`
	GasPrice  uint64 `json:"gas_price"`
	GasUnits  uint64 `json:"gas_units"`
}

// NewBlockTx constructs a new block Tx.
func NewBlockTx(tx SignedTx, gasPrice, gasUnits uint64) BlockTx {
	return BlockTx{
		SignedTx:  tx,
		Timestamp: uint64(time.Now().UnixNano()),
		GasPrice:  gasPrice,
		GasUnits:  gasUnits,
	}
}
