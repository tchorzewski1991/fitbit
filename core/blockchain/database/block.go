package database

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
	"github.com/tchorzewski1991/fitbit/core/blockchain/signature"
)

// Block orchestrates a batch of transactions together.
type Block struct {
	Header BlockHeader
	Tree   *merkle.Tree[BlockTx]
}

// BlockHeader represents common information about each block.
type BlockHeader struct {
	Height        uint64    `json:"height"`
	PrevHash      string    `json:"prev_hash"`
	Timestamp     uint64    `json:"timestamp"`
	BeneficiaryID AccountID `json:"beneficiary"`
	Difficulty    uint16    `json:"difficulty"`
	Reward        uint64    `json:"reward"`
	StateRoot     string    `json:"state_root"`
	TxRoot        string    `json:"tx_root"`
	Nonce         uint64    `json:"nonce"`
}

// BlockData represents what can be serialized to disk or over the network.
type BlockData struct {
	Hash   string      `json:"hash"`
	Header BlockHeader `json:"header"`
	Txs    []BlockTx   `json:"txs"`
}

// POWArgs represents a set of arguments necessary to run Proof of Work.
type POWArgs struct {
	BeneficiaryID AccountID
	Difficulty    uint16
	Reward        uint64
	PrevBlock     Block
	StateRoot     string
	Txs           []BlockTx
	Ev            func(s string, args ...any)
}

// POW constructs a new Block by finding a nonce that solves a cryptographic challenge.
func POW(ctx context.Context, args POWArgs) (Block, error) {
	args.Ev("[DB][POW][Started]")
	defer args.Ev("[DB][POW][Finished]")

	// Get the hash value of a previous block.
	// When it is the first block then we zero hash value will be used.
	prevBlockHash := args.PrevBlock.Hash()

	// Get the height of the previous block.
	// When it is the first block then its height will be zero.
	prevBlockHeight := args.PrevBlock.Height()

	// Keep transactions cryptographically safe by constructing a new merkle tree.
	// The root hash of this tree will be kept as part of the block.
	tree, err := merkle.NewTree(args.Txs)
	if err != nil {
		return Block{}, err
	}

	// Initialize a new Block where nonce is not set.
	// Proper nonce will set by running a POW algorithm.
	block := Block{
		Header: BlockHeader{
			Height:        prevBlockHeight + 1,
			PrevHash:      prevBlockHash,
			Timestamp:     uint64(time.Now().UTC().Unix()),
			BeneficiaryID: args.BeneficiaryID,
			Difficulty:    args.Difficulty,
			Reward:        args.Reward,
			StateRoot:     args.StateRoot,
			TxRoot:        tree.RootHex(),
			Nonce:         0, // Nonce will be calculated by performing POW below.
		},
		Tree: tree,
	}

	err = performPOW(ctx, &block, args.Ev)
	if err != nil {
		return Block{}, err
	}

	return block, nil
}

// Hash builds up unique string representation of the Block.
// It returns signature.ZeroHash if the block height is 0.
func (b Block) Hash() string {
	if b.Height() > 0 {
		return signature.Hash(b.Header)
	}
	return signature.ZeroHash
}

// Height returns the number of current Block.
func (b Block) Height() uint64 {
	return b.Header.Height
}

// ToBlockData constructs a new BlockData from Block.
func (b Block) ToBlockData() BlockData {
	return BlockData{
		Hash:   b.Hash(),
		Header: b.Header,
		Txs:    b.Tree.Values(),
	}
}

// Validate verifies through the core set of check whether given Block is valid.
func (b Block) Validate(prevBlock Block, prevStateRoot string) error {

	// Verify if we do not have a fork.
	height := b.Header.Height
	prevHeight := prevBlock.Header.Height
	if (height - prevHeight) != 1 {
		return fmt.Errorf("fork check failed: height: %d | prev height: %d", height, prevHeight)
	}

	// Verify if difficulty is the same or greater than parent block difficulty.
	difficulty := b.Header.Difficulty
	prevDifficulty := prevBlock.Header.Difficulty
	if difficulty < prevDifficulty {
		return fmt.Errorf("difficulty check failed: difficulty: %d | prev difficulty: %d", difficulty, prevDifficulty)
	}

	// Verify if hash has been actually solved.
	hash := b.Hash()
	refHash := buildReferenceHash(difficulty)
	if !isSolved(difficulty, refHash, hash) {
		return fmt.Errorf("hash solved check failed: hash: %s | ref hash: %s", hash, refHash)
	}

	// Verify if previous block hash is the same as previous block hash of the validated block.
	prevHash := b.Header.PrevHash
	prevBlockHash := prevBlock.Hash()
	if prevHash != prevBlockHash {
		return fmt.Errorf("prev hash check failed: prev hash: %s | prev block hash: %s", prevHash, prevBlockHash)
	}

	// Verify if previous block state root is the same as state root of the validated block.
	stateRoot := b.Header.StateRoot
	if stateRoot != prevStateRoot {
		return fmt.Errorf("state root check failed: state root: %s | prev state root: %s", stateRoot, prevStateRoot)
	}

	// Verify if merkle root does not match transactions.
	txRoot := b.Header.TxRoot
	treeRoot := b.Tree.RootHex()
	if txRoot != treeRoot {
		return fmt.Errorf("tx root check failed: tx root: %s | tx tree root: %s", txRoot, treeRoot)
	}

	return nil
}

// ToBlock constructs a new Block out of BlockData.
func (bd BlockData) ToBlock() (Block, error) {
	// Generate new merkle tree out of block data transactions.
	tree, err := merkle.NewTree(bd.Txs)
	if err != nil {
		return Block{}, err
	}
	// Build new block.
	return Block{
		Header: bd.Header,
		Tree:   tree,
	}, nil
}

func performPOW(ctx context.Context, block *Block, ev func(s string, args ...any)) error {

	// Start with setting new, random nonce.
	nonce, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return err
	}
	block.Header.Nonce = nonce.Uint64()

	difficulty := block.Header.Difficulty
	refHash := buildReferenceHash(difficulty)

	var attempts int

	for {
		attempts++

		if attempts%1_000_000 == 0 {
			ev("[DB][POW][Attempts: %d]", attempts)
		}

		// Ensure we are not timed-out or cancelled.
		if err = ctx.Err(); err != nil {
			ev("[DB][POW][Mining cancelled]")
			return err
		}

		// Calculate a new hash based on initial nonce.
		hash := block.Hash()

		// Verify whether we solve the cryptographic puzzle.
		if !isSolved(difficulty, refHash, hash) {
			block.Header.Nonce++
			continue
		}

		// Ensure we are not timed-out or cancelled.
		if err = ctx.Err(); err != nil {
			ev("[DB][POW][Mining cancelled]")
			return err
		}

		ev("[DB][POW][Solved! Prev: %s New: %s]", block.Header.PrevHash, hash)
		ev("[DB][POW][Solved! Attempts: %d]", attempts)

		// Reaching this code means we solved the puzzle.
		return nil
	}
}

func buildReferenceHash(difficulty uint16) string {
	b := strings.Builder{}
	b.WriteString("0x")

	for difficulty > 0 {
		b.WriteString("0")
		difficulty--
	}

	return b.String()
}

func isSolved(difficulty uint16, refHash, hash string) bool {
	return refHash[2:2+difficulty] == hash[2:2+difficulty]
}
