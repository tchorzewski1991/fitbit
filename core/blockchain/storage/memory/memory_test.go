package memory_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/storage/memory"
	"testing"
)

func TestMemory(t *testing.T) {
	// Initialize new memory
	m, err := memory.New()
	assert.Nil(t, err)

	// Read block data from memory and assert err (height is too low)
	_, err = m.Read(0)
	assert.NotNil(t, err)

	// Read block data from memory and assert err (block does not exist yet)
	_, err = m.Read(1)
	assert.NotNil(t, err)

	// Write block data to memory
	err = m.Write(1, database.BlockData{})
	assert.Nil(t, err)

	// Read block data from memory (block should already exist)
	_, err = m.Read(1)
	assert.Nil(t, err)

	// Write block data to memory with too big height
	err = m.Write(3, database.BlockData{})
	assert.NotNil(t, err)

	// Read block data from memory with too big height
	_, err = m.Read(2)
	assert.NotNil(t, err)

	// Reset memory state
	err = m.Reset()
	assert.Nil(t, err)

	// Read block data from memory one more time and assert err
	_, err = m.Read(1)
	assert.NotNil(t, err)
}
