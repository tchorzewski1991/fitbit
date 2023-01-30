package public

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/state"
	"github.com/tchorzewski1991/fitbit/core/nameservice"
	"github.com/tchorzewski1991/fitbit/core/web"
	"go.uber.org/zap"
)

type Handlers struct {
	Log         *zap.SugaredLogger
	State       *state.State
	NameService *nameservice.NameService
}

// Health handler provides info about the status of the public node.
func (h Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}

// Genesis handler provides info about the genesis manifest.
func (h Handlers) Genesis(c *gin.Context) {
	c.JSON(http.StatusOK, h.State.Genesis())
}

// Accounts handler provides info about all account balances.
func (h Handlers) Accounts(c *gin.Context) {
	accounts := make([]account, 0)

	for _, dbAccount := range h.State.Accounts() {
		accounts = append(accounts, toAccount(h, dbAccount))
	}

	c.JSON(http.StatusOK, accounts)
}

// Account handler provides info about specific account balance.
func (h Handlers) Account(c *gin.Context) {
	accountID, err := database.ToAccountID(c.Param("address"))
	if err != nil {
		c.JSON(http.StatusBadRequest, web.Error(err))
		return
	}

	dbAccount, err := h.State.Account(accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, web.Error(err))
		return
	}

	c.JSON(http.StatusOK, toAccount(h, dbAccount))
}

// SubmitWalletTx handler adds new transaction to the mempool.
func (h Handlers) SubmitWalletTx(c *gin.Context) {

	var tx database.SignedTx
	err := c.ShouldBindJSON(&tx)
	if err != nil {
		c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("failed to decode tx: %w", err)))
		return
	}

	err = h.State.UpsertWalletTx(tx)
	if err != nil {
		c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("failed to upsert tx: %w", err)))
		return
	}

	c.JSON(http.StatusOK, web.Success())
}

// UncommittedWalletTx handler provides info about all uncommited transactions.
func (h Handlers) UncommittedWalletTx(c *gin.Context) {
	txs := make([]uncommitedTx, 0)

	for _, dbTx := range h.State.UncommittedTx() {
		txs = append(txs, toUncommittedTx(h, dbTx))
	}

	c.JSON(http.StatusOK, txs)
}

// UncommittedWalletAccountTx handler provides info about all uncommited account transactions.
func (h Handlers) UncommittedWalletAccountTx(c *gin.Context) {
	accountID, err := database.ToAccountID(c.Param("address"))
	if err != nil {
		c.JSON(http.StatusBadRequest, web.Error(err))
		return
	}

	txs := make([]uncommitedTx, 0)

	for _, dbTx := range h.State.UncommittedTxByAccount(accountID) {
		txs = append(txs, toUncommittedTx(h, dbTx))
	}

	c.JSON(http.StatusOK, txs)
}
