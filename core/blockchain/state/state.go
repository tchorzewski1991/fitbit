package state

import (
	"context"
	"sync"

	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/genesis"
	"github.com/tchorzewski1991/fitbit/core/blockchain/mempool"
	"github.com/tchorzewski1991/fitbit/core/blockchain/network"
)

const oneUnitOfGas = 1

type Worker interface {
	Shutdown()
	StartMining()
	ShareTx(tx database.BlockTx)
	StopMining()
}

type EventHandler func(s string, args ...any)

// Config keeps track over all dependencies necessary for
// proper State initialization.
type Config struct {
	BeneficiaryID database.AccountID
	Host          string
	Genesis       genesis.Genesis
	Storage       database.Storage
	EventHandler  EventHandler
	KnownPeers    *network.PeerSet
}

// State holds all blockchain dependencies and provides core API.
type State struct {
	mu sync.RWMutex

	beneficiaryID database.AccountID
	host          string

	genesis    genesis.Genesis
	mempool    *mempool.Mempool
	db         *database.Database
	knownPeers *network.PeerSet
	ev         EventHandler

	worker Worker
}

// New constructs a new State.
func New(cfg Config) (*State, error) {

	db, err := database.New(cfg.Genesis, cfg.Storage)
	if err != nil {
		return nil, err
	}

	if cfg.EventHandler == nil {
		// Set no-op event handler if event handler has not been set.
		cfg.EventHandler = func(s string, args ...any) {}
	}

	return &State{
		beneficiaryID: cfg.BeneficiaryID,
		host:          cfg.Host,
		genesis:       cfg.Genesis,
		mempool:       mempool.New(),
		db:            db,
		knownPeers:    cfg.KnownPeers,
		ev:            cfg.EventHandler,
	}, nil
}

func (s *State) RegisterWorker(worker Worker) {
	s.worker = worker
}

// Shutdown brings the current instance of State down.
func (s *State) Shutdown() error {

	// Make sure db is properly closed.
	defer func() {
		s.db.Close()
	}()

	// Make sure all blockchain activity is properly stopped.
	s.worker.Shutdown()

	return nil
}

// Genesis returns a copy of genesis information.
func (s *State) Genesis() genesis.Genesis {
	return s.genesis
}

// Accounts returns a copy of all database accounts.
func (s *State) Accounts() database.Accounts {
	return s.db.Accounts()
}

// Account returns a copy of an account requested by given account ID.
func (s *State) Account(accountID database.AccountID) (database.Account, error) {
	return s.db.Account(accountID)
}

// QueryBlocksByHeight returns a copy of blocks by given height range.
func (s *State) QueryBlocksByHeight(from, to uint64) ([]database.Block, error) {
	s.ev("[STATE][QueryBlocksByHeight][Start querying blocks from %d, to: %d]", from, to)
	defer s.ev("[STATE][QueryBlocksByHeight][Querying blocks finished]")

	var blocks []database.Block

	for i := from; i <= to; i++ {
		block, err := s.db.ReadBlock(i)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// UpsertWalletTx adds a new wallet transaction to the mempool.
func (s *State) UpsertWalletTx(signedTx database.SignedTx) error {

	// Convert signed tx to proper format.
	tx := database.NewBlockTx(signedTx, s.genesis.GasPrice, oneUnitOfGas)

	// Verify whether tx has a proper signature and data.
	err := tx.Verify(s.genesis.ChainID)
	if err != nil {
		return err
	}

	// Upsert tx to the mempool.
	err = s.mempool.Upsert(tx)
	if err != nil {
		return err
	}

	// Share tx with other peers to let them have a chance to mine a new block.
	s.worker.ShareTx(tx)

	// Start mining of the new block on local node.
	s.worker.StartMining()

	return nil
}

// UpsertNodeTx adds a new node transaction to the mempool.
func (s *State) UpsertNodeTx(tx database.BlockTx) error {

	// Verify whether tx has a proper signature and data.
	err := tx.Verify(s.genesis.ChainID)
	if err != nil {
		return err
	}

	// Upsert tx to the mempool.
	err = s.mempool.Upsert(tx)
	if err != nil {
		return err
	}

	// Start mining of the new block on local node.
	s.worker.StartMining()

	return nil
}

// UncommittedTx returns a copy of all uncommited transactions from the mempool.
func (s *State) UncommittedTx() []database.BlockTx {
	return s.mempool.Select(mempool.SelectAll())
}

// UncommittedTxByAccount returns a copy of all uncommited transactions from the mempool.
func (s *State) UncommittedTxByAccount(accountID database.AccountID) []database.BlockTx {
	return s.mempool.Select(mempool.SelectByAccount(accountID))
}

// MempoolSize returns the current number of transactions in the mempool.
func (s *State) MempoolSize() int {
	return s.mempool.Size()
}

// MineBlock attempts to create a new block using POW consensus algorithm.
func (s *State) MineBlock(ctx context.Context) (database.Block, error) {
	s.ev("[STATE][MineBlock][Started new mining]")
	defer s.ev("[STATE][MineBlock][Mining finished]")

	// Prepare all data necessary for the next block to be mined.
	txs := s.mempool.Select(mempool.SelectAll())
	prevBlock := s.db.LastBlock()
	prevStateRoot := s.db.StateRoot()

	// Mine the block by using Proof of Work consensus algorithm.
	block, err := database.POW(ctx, database.POWArgs{
		BeneficiaryID: s.beneficiaryID,
		Difficulty:    s.genesis.Difficulty,
		Reward:        s.genesis.MiningReward,
		PrevBlock:     prevBlock,
		StateRoot:     prevStateRoot,
		Txs:           txs,
		Ev:            s.ev,
	})
	if err != nil {
		return database.Block{}, err
	}

	// Validate and write the block to the underlying storage.
	s.mu.Lock()
	{
		if err = block.Validate(prevBlock, prevStateRoot); err != nil {
			return database.Block{}, err
		}

		if err = s.db.WriteBlock(block); err != nil {
			return database.Block{}, err
		}

		s.db.UpdateLastBlock(block)

		// Apply mining reward to the node's beneficiary account.
		s.db.ApplyMiningReward(block)

		// Process the transactions and apply balance changes to accounts.
		for _, tx := range txs {
			if err = s.mempool.Remove(tx); err != nil {
				continue
			}

			if err = s.db.ApplyTransaction(block, tx); err != nil {
				continue
			}
		}
	}
	s.mu.Unlock()

	return block, nil
}

// ProcessBlock attempts to create a new block after receiving it from other peers.
func (s *State) ProcessBlock(block database.Block) error {
	s.ev("[STATE][ProcessBlock][Processing new block]")
	defer s.ev("[STATE][ProcessBlock][Block processing finished]")

	prevBlock := s.db.LastBlock()
	prevStateRoot := s.db.StateRoot()

	s.mu.Lock()
	{
		if err := block.Validate(prevBlock, prevStateRoot); err != nil {
			return err
		}

		if err := s.db.WriteBlock(block); err != nil {
			return err
		}

		s.db.UpdateLastBlock(block)

		// Apply mining reward to the node's beneficiary account.
		s.db.ApplyMiningReward(block)

		// Process the transactions and apply balance changes to accounts.
		for _, tx := range block.Tree.Values() {
			if err := s.mempool.Remove(tx); err != nil {
				continue
			}

			if err := s.db.ApplyTransaction(block, tx); err != nil {
				continue
			}
		}
	}
	s.mu.Unlock()

	s.worker.StopMining()

	return nil
}

// KnownPeers returns a copy of all known peers.
func (s *State) KnownPeers() []network.Peer {
	return s.knownPeers.Peers(network.SelectAllPeers())
}

// ExternalPeers returns a copy of all peers excluding the host peer (current node).
func (s *State) ExternalPeers() []network.Peer {
	return s.knownPeers.Peers(network.SelectWithoutPeer(s.host))
}

// AddPeer adds given peer to the list of known peers.
func (s *State) AddPeer(peer network.Peer) bool {
	return s.knownPeers.Add(peer)
}

// DeletePeer removes given peer from the list of known peers.
func (s *State) DeletePeer(peer network.Peer) bool {
	return s.knownPeers.Delete(peer)
}

// LastBlock returns a copy of the last successfully mined block.
func (s *State) LastBlock() database.Block {
	return s.db.LastBlock()
}
