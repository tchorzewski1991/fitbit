package disk

import (
	"fmt"
	"os"
	"path"

	"github.com/goccy/go-json"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
)

// Disk represents the database.Storage implementation we can use
// for storing and reading blocks of the disk from their own separate files.
type Disk struct {
	dataPath string
}

// New constructs a new Disk.
// This function takes care about building any subdirectories structure
// along the given path.
func New(dataPath string) (*Disk, error) {
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return nil, err
	}
	return &Disk{dataPath: dataPath}, nil
}

// Write writes the database.BlockData on the disk in a JSON file named by given height.
// JSON file format is the subject to change in the future, and it is the only supported
// data serialization strategy atm.
func (d *Disk) Write(height int, data database.BlockData) error {

	// Marshall data into JSON.
	bs, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	// Open or create a new JSON file named by given block height.
	f, err := os.OpenFile(d.filePath(height), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// Write JSON bytes to the file.
	_, err = f.Write(bs)
	if err != nil {
		return err
	}

	return nil
}

// Read reads the JSON file named by given height and decodes it into database.BlockData.
// JSON file format is the subject to change in the future, and it is the only supported
// data serialization strategy atm.
func (d *Disk) Read(height int) (*database.BlockData, error) {

	// Open a JSON file named by given height.
	f, err := os.OpenFile(d.filePath(height), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	// Decode JSON file into database.BlockData.
	var data database.BlockData
	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// Reset removes all the blocks from the disk and recreates the subdirectory structure
// defined by the path given during initialization.
func (d *Disk) Reset() error {
	if err := os.RemoveAll(d.dataPath); err != nil {
		return err
	}
	return os.MkdirAll(d.dataPath, 0755)
}

func (d *Disk) filePath(blockHeight int) string {
	return path.Join(d.dataPath, fmt.Sprintf("%d.json", blockHeight))
}
