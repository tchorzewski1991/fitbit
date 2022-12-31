package genesis

import (
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-json"
)

type Genesis struct {
	Date         time.Time         `json:"date"`
	ChainID      uint16            `json:"chain_id"`
	TxPerBlock   uint16            `json:"tx_per_block"`
	Difficulty   uint16            `json:"difficulty"`
	MiningReward uint64            `json:"mining_reward"`
	GasPrice     uint64            `json:"gas_price"`
	Balances     map[string]uint64 `json:"balances"`
}

func Load() (Genesis, error) {
	f, err := os.Open("data/genesis.json")
	if err != nil {
		return Genesis{}, fmt.Errorf("failed to open genesis file: %w", err)
	}

	var gen Genesis
	err = json.NewDecoder(f).Decode(&gen)
	if err != nil {
		return Genesis{}, fmt.Errorf("failed to encode genesis file: %w", err)
	}

	return gen, nil
}
