package mempool

import (
	"fmt"
	"sort"
	"sync"

	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
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

// Size returns the current number of transactions in the Mempool.
func (m *Mempool) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pool)
}

// Upsert adds a new transaction to Mempool.
func (m *Mempool) Upsert(tx database.BlockTx) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := prepareKey(tx)

	m.pool[key] = tx

	return nil
}

// Remove deletes a transaction from Mempool.
func (m *Mempool) Remove(tx database.BlockTx) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := prepareKey(tx)

	delete(m.pool, key)

	return nil
}

// Select returns a copy of all transactions from Mempool.
func (m *Mempool) Select(filter SelectFunc) []database.BlockTx {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var txs []database.BlockTx

	for _, tx := range m.pool {
		if filter(tx) {
			txs = append(txs, tx)
		}
	}

	sort.Sort(byNonce(txs))

	return txs
}

// Truncate deletes all transactions from Mempool.
func (m *Mempool) Truncate() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pool = make(map[string]database.BlockTx)
}

// prepareKey builds a new mempool key based on transaction from address and nonce.
func prepareKey(tx database.BlockTx) string {
	return fmt.Sprintf("%s:%d", tx.From, tx.Nonce)
}
