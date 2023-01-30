package private

import (
	"fmt"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/network"
	"github.com/tchorzewski1991/fitbit/core/blockchain/state"
	"github.com/tchorzewski1991/fitbit/core/web"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
}

// Health returns the status of the private node.
func (h Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}

// Status handler provides info about the last block and list of known peers.
func (h Handlers) Status(c *gin.Context) {
	lastBlock := h.State.LastBlock()
	knownPeers := h.State.ExternalPeers()

	c.JSON(http.StatusOK, toStatus(lastBlock, knownPeers))
}

// UncommittedTx handler provides info about all uncommited block transactions.
func (h Handlers) UncommittedTx(c *gin.Context) {
	txs := make([]database.BlockTx, 0)

	for _, dbTx := range h.State.UncommittedTx() {
		txs = append(txs, dbTx)
	}

	c.JSON(http.StatusOK, txs)
}

// BlocksByHeight handler provides the list of blocks specified by given height.
func (h Handlers) BlocksByHeight(c *gin.Context) {

	fromParam := strings.TrimSpace(c.Param("from"))
	if fromParam == "" {
		c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("param :from is mandatory")))
		return
	}

	var from uint64
	switch {
	case fromParam == "latest":
		from = h.State.LastBlock().Height()
	default:
		parsedFrom, err := strconv.ParseUint(fromParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("param :from is invalid: %w", err)))
			return
		}
		from = parsedFrom
	}

	toParam := strings.TrimSpace(c.Param("to"))
	if toParam == "" {
		c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("param :to is mandatory")))
		return
	}

	var to uint64
	switch {
	case toParam == "latest":
		to = h.State.LastBlock().Height()
	default:
		parsedTo, err := strconv.ParseUint(toParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("param :to is invalid: %w", err)))
			return
		}
		to = parsedTo
	}

	blocks, err := h.State.QueryBlocksByHeight(from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("failed to query blocks by height: %w", err)))
		return
	}

	result := make([]database.BlockData, 0)

	for _, block := range blocks {
		result = append(result, block.ToBlockData())
	}

	c.JSON(http.StatusOK, result)
}

// SubmitPeer handler adds new peer to the list of known peers.
func (h Handlers) SubmitPeer(c *gin.Context) {

	var peer network.Peer
	err := c.ShouldBindJSON(&peer)
	if err != nil {
		c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("failed to decode peer: %w", err)))
		return
	}

	h.State.AddPeer(peer)

	c.JSON(http.StatusOK, web.Success())
}

// SubmitBlock handler starts new processing of a proposed block.
func (h Handlers) SubmitBlock(c *gin.Context) {

	var blockData database.BlockData
	err := c.ShouldBindJSON(&blockData)
	if err != nil {
		c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("failed to decode block data: %w", err)))
		return
	}

	block, err := blockData.ToBlock()
	if err != nil {
		c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("failed to prepare block: %w", err)))
		return
	}

	err = h.State.ProcessBlock(block)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, web.Error(fmt.Errorf("failed to process block: %w", err)))
	}

	c.JSON(http.StatusOK, web.Success())
}

// SubmitTx handler adds new transaction to the mempool.
func (h Handlers) SubmitTx(c *gin.Context) {

	var tx database.BlockTx
	err := c.ShouldBindJSON(&tx)
	if err != nil {
		c.JSON(http.StatusBadRequest, web.Error(fmt.Errorf("failed to decode tx data: %w", err)))
		return
	}

	err = h.State.UpsertNodeTx(tx)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, web.Error(fmt.Errorf("failed to upsert block: %w", err)))
	}

	c.JSON(http.StatusOK, web.Success())
}
