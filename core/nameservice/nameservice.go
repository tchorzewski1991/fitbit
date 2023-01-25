package nameservice

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
)

// NameService provides lookup functionality for
// account ID to account name mapping.
type NameService struct {
	accounts map[database.AccountID]string
}

// New constructs a new NameService.
func New(accountsPath string) (*NameService, error) {
	accounts := make(map[database.AccountID]string)

	fn := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.New("accounts path is not valid")
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".ecdsa" {
			return errors.New("detected not supported file ext")
		}

		priv, err := crypto.LoadECDSA(path)
		if err != nil {
			return err
		}

		accountID, err := database.PubToAccountID(priv.PublicKey)
		if err != nil {
			return err
		}

		accounts[accountID] = strings.TrimSuffix(info.Name(), ".ecdsa")

		return nil
	}
	err := filepath.Walk(accountsPath, fn)
	if err != nil {
		return nil, err
	}

	return &NameService{accounts: accounts}, nil
}

// FindName returns the account name associated with given account ID.
// It returns given account ID if associated account name does not exist.
func (ns *NameService) FindName(accountID database.AccountID) string {
	if accountName, exist := ns.accounts[accountID]; exist {
		return accountName
	}
	return accountID.String()
}
