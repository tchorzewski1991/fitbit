package mempool

import (
	"fmt"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"sync"
)

// Mempool represents the cache for waiting transactions.
type Mempool struct {
	mu   sync.RWMutex
	pool map[string]database.BlockTx
}

// New constructs a new Mempool.
func New() *Mempool {
	return &Mempool{
		pool: make(map[string]database.BlockTx),
	}
}

// Count returns the current number of transactions in the Mempool.
func (m *Mempool) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pool)
}

// Upsert adds a new transaction to Mempool.
// TODO: Should we check err on key preparation step?
func (m *Mempool) Upsert(tx database.BlockTx) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := prepareKey(tx)

	m.pool[key] = tx

	return nil
}

// Remove deletes a transaction from Mempool.
// TODO: Should we check err on key preparation step?
func (m *Mempool) Remove(tx database.BlockTx) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := prepareKey(tx)

	delete(m.pool, key)

	return nil
}

// Get returns all transactions from Mempool.
func (m *Mempool) Get() []database.BlockTx {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var txs []database.BlockTx

	for _, tx := range m.pool {
		txs = append(txs, tx)
	}

	return txs
}

// Truncate deletes all transactions from Mempool
func (m *Mempool) Truncate() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pool = make(map[string]database.BlockTx)
}

// prepareKey builds a new mempool key based on transaction from address and nonce.
func prepareKey(tx database.BlockTx) string {
	return fmt.Sprintf("%s:%d", tx.From, tx.Nonce)
}
