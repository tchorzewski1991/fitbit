package database

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/tchorzewski1991/fitbit/core/blockchain/genesis"
	"github.com/tchorzewski1991/fitbit/core/blockchain/signature"
)

type Storage interface {
	Write(height uint64, data BlockData) error
	Read(height uint64) (*BlockData, error)
	Reset() error
	Close() error
}

type Database struct {
	mu        sync.RWMutex
	genesis   genesis.Genesis
	accounts  Accounts
	storage   Storage
	lastBlock Block
}

// New constructs a new Database.
func New(genesis genesis.Genesis, storage Storage) (*Database, error) {
	db := Database{
		genesis:  genesis,
		accounts: make(Accounts),
		storage:  storage,
	}

	if err := db.loadAccounts(); err != nil {
		return nil, err
	}

	var stepErr error

	for height := uint64(1); stepErr == nil; height++ {

		// Load block by given height.
		block, err := db.ReadBlock(height)
		if err != nil {
			break
		}

		// Validate block according to the previous block and accounts state.
		err = block.Validate(db.lastBlock, db.StateRoot())
		if err != nil {
			stepErr = err
			break
		}

		// Apply all block transactions.
		for _, tx := range block.Tree.Values() {
			err = db.ApplyTransaction(block, tx)
			if err != nil {
				continue
			}
		}

		db.ApplyMiningReward(block)

		db.lastBlock = block
	}
	if stepErr != nil {
		return nil, stepErr
	}

	return &db, nil
}

// Accounts returns the copy of all known Accounts.
func (db *Database) Accounts() Accounts {
	db.mu.RLock()
	defer db.mu.RUnlock()

	accounts := make(Accounts)
	for id, account := range db.accounts {
		accounts[id] = account
	}

	return accounts
}

// Account returns the copy of an account by given AccountID.
func (db *Database) Account(accountID AccountID) (Account, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	account, ok := db.accounts[accountID]
	if !ok {
		return Account{}, errors.New("account not found")
	}

	return account, nil
}

// Remove deletes an account by given AccountID.
func (db *Database) Remove(accountID AccountID) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.accounts, accountID)
	return nil
}

// Reset resets database to the initial (genesis) state.
func (db *Database) Reset() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Ensure db storage is reset to initial state
	if err := db.storage.Reset(); err != nil {
		return err
	}

	// Ensure db accounts are reset and loaded one more time
	db.accounts = make(map[AccountID]Account)
	if err := db.loadAccounts(); err != nil {
		return err
	}

	// Ensure last block is reset to initial value
	db.lastBlock = Block{}

	return nil
}

// WriteBlock writes a new Block to the underlying Storage.
func (db *Database) WriteBlock(block Block) error {
	err := db.storage.Write(block.Height(), block.ToBlockData())
	if err != nil {
		return fmt.Errorf("write block err: %w", err)
	}
	return nil
}

// ReadBlock reads Block from the underlying Storage by given height.
func (db *Database) ReadBlock(height uint64) (Block, error) {
	data, err := db.storage.Read(height)
	if err != nil {
		return Block{}, fmt.Errorf("read block err: %w", err)
	}
	return data.ToBlock()
}

// LastBlock returns the copy of last stored Block.
func (db *Database) LastBlock() Block {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.lastBlock
}

// UpdateLastBlock updates last stored Block.
func (db *Database) UpdateLastBlock(block Block) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.lastBlock = block
}

// ApplyMiningReward updates beneficiary account balance with mining reward defined in genesis file.
func (db *Database) ApplyMiningReward(block Block) {
	db.mu.Lock()
	defer db.mu.Unlock()

	beneficiary := db.accounts[block.Header.BeneficiaryID]
	beneficiary.Balance += block.Header.Reward

	db.accounts[block.Header.BeneficiaryID] = beneficiary
}

func (db *Database) ApplyTransaction(block Block, tx BlockTx) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	from, ok := db.accounts[tx.From]
	if !ok {
		from = Account{ID: tx.From, Balance: 0}
	}

	to, ok := db.accounts[tx.To]
	if !ok {
		to = Account{ID: tx.To, Balance: 0}
	}

	beneficiary, ok := db.accounts[block.Header.BeneficiaryID]
	if !ok {
		beneficiary = Account{ID: tx.To, Balance: 0}
	}

	// Calculate current gas fee based on given transaction.
	gasFee := tx.GasPrice * tx.GasUnits

	// Adjust gas fee according to the balance of 'from' account.
	if gasFee > from.Balance {
		gasFee = from.Balance
	}

	// Update beneficiary account with the gas fee.
	from.Balance -= gasFee
	beneficiary.Balance += gasFee

	// Apply new beneficiary account balance to database.
	db.accounts[tx.From] = from
	db.accounts[block.Header.BeneficiaryID] = beneficiary

	// Perform necessary accounting checks.
	{
		if tx.Nonce != (from.Nonce + 1) {
			return fmt.Errorf("tx invalid, wrong nonce, got: %d, expected: %d", tx.Nonce, from.Nonce+1)
		}

		if from.Balance == 0 || from.Balance < (tx.Value+tx.Tip) {
			return fmt.Errorf("tx invalid, insufficient funds, got: %d, expected: %d", from.Balance, tx.Value+tx.Tip)
		}
	}

	// Update balances between two parties.
	from.Balance -= tx.Value
	to.Balance += tx.Value

	// Update beneficiary account with the tip.
	from.Balance -= tx.Tip
	beneficiary.Balance += tx.Tip

	// Update from nonce for the next transaction check.
	from.Nonce = tx.Nonce

	// Apply final account balances to database.
	db.accounts[tx.From] = from
	db.accounts[tx.To] = to
	db.accounts[block.Header.BeneficiaryID] = beneficiary

	return nil
}

// StateRoot returns a hash based on the known database accounts.
func (db *Database) StateRoot() string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// We don't have many accounts yet, but when the collection will grow
	// we have the possibility to bypass unnecessary allocations by
	// setting slice capacity upfront.
	accounts := make([]Account, 0, len(db.accounts))

	for _, account := range db.accounts {
		accounts = append(accounts, account)
	}

	// Sorting accounts by their ID is mandatory as the order in which we get
	// accounts from the map cannot be determined upfront. Order matters because
	// by changing order hash will be changed as well.
	sort.Sort(byAccountID(accounts))

	return signature.Hash(accounts)
}

// Close closes underlying Storage engine.
func (db *Database) Close() {
	_ = db.storage.Close()
}

// private API

func (db *Database) loadAccounts() error {
	for account, balance := range db.genesis.Balances {
		accountID, err := ToAccountID(account)
		if err != nil {
			return err
		}
		db.accounts[accountID] = Account{
			ID:      accountID,
			Balance: balance,
		}
	}
	return nil
}
