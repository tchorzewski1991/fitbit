package disk_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/storage/disk"
)

func TestDisk(t *testing.T) {
	// Initialize new disk
	d, err := disk.New("testdata")
	assert.Nil(t, err)

	// Write block data to disk
	err = d.Write(1, database.BlockData{})
	assert.Nil(t, err)

	// Read block data from disk
	_, err = d.Read(1)
	assert.Nil(t, err)

	// Reset disk state
	err = d.Reset()
	assert.Nil(t, err)

	// Read block data from disk one more time and asset err
	_, err = d.Read(1)
	assert.NotNil(t, err)
}
