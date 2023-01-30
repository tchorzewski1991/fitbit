package memory

import (
	"fmt"
	"sync"

	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
)

// Memory represents the database.Storage implementation we can use
// for storing and reading blocks from memory using a slice.
type Memory struct {
	mu     sync.RWMutex
	blocks []database.BlockData
}

// New constructs a new Memory.
func New() (*Memory, error) {
	return &Memory{}, nil
}

// Write writes the database.BlockData to memory protecting the order of blocks based on given height.
func (m *Memory) Write(height uint64, data database.BlockData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if int(height) != len(m.blocks)+1 {
		return fmt.Errorf("cannot write block with height: %d to the chain of len: %d", height, len(m.blocks))
	}

	m.blocks = append(m.blocks, data)

	return nil
}

// Read reads the database.BlockData from memory by given height.
func (m *Memory) Read(height uint64) (*database.BlockData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if height == 0 || int(height) > len(m.blocks) {
		return nil, fmt.Errorf("cannot read block with height: %d from the chain of len: %d", height, len(m.blocks))
	}

	return &m.blocks[height-1], nil
}

// Reset removes all the blocks from memory.
func (m *Memory) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blocks = nil
	return nil
}

// Close is a no-op method required by database.Storage interface.
func (m *Memory) Close() error {
	return nil
}
