package worker

import (
	"context"
	"errors"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"sync"
	"time"

	"github.com/tchorzewski1991/fitbit/core/blockchain/state"
)

type Worker struct {
	state       *state.State
	ticker      *time.Ticker
	ev          EventHandler
	wg          sync.WaitGroup
	shutdown    chan struct{}
	startMining chan bool
	stopMining  chan bool
	shareTx     chan database.BlockTx
}

type EventHandler func(s string, args ...any)

var peerSyncInterval = 90 * time.Second

func Run(s *state.State, ev EventHandler) {
	w := Worker{
		state:       s,
		ticker:      time.NewTicker(peerSyncInterval),
		ev:          ev,
		shutdown:    make(chan struct{}),
		startMining: make(chan bool, 1),
		stopMining:  make(chan bool, 1),
		shareTx:     make(chan database.BlockTx, 10),
	}

	// Register worker instance in state.
	s.RegisterWorker(&w)

	// Before we start peer syncing and mining we need to ensure
	// our node is up-to-date with the chain.
	w.Sync()

	// Load all operations we need to start.
	operations := []func(){
		w.peerSyncer,
		w.txSyncer,
		w.miningListener,
	}

	w.wg.Add(len(operations))

	// We want to make sure all the operations are running before
	// we leave this function call.
	hasStarted := make(chan bool)

	// Start all the operational goroutines.
	for _, operation := range operations {
		go func(op func()) {
			defer w.wg.Done()
			hasStarted <- true
			op()
		}(operation)
	}

	// We want to make sure all the operations are running before
	// we leave this function call.
	for i := 0; i < len(operations); i++ {
		<-hasStarted
	}
}

func (w *Worker) Shutdown() {
	w.ev("[WORKER][Shutdown][Started]")
	defer w.ev("[WORKER][Shutdown][Finished]")

	w.StopMining()

	w.ticker.Stop()

	close(w.shutdown)
	w.wg.Wait()
}

func (w *Worker) StartMining() {
	w.ev("[WORKER][StartMining][Sending signal]")
	defer w.ev("[WORKER][StartMining][Signal sent]")

	select {
	case w.startMining <- true:
	default:
	}
}

func (w *Worker) StopMining() {
	w.ev("[WORKER][StopMining][Sending signal]")
	defer w.ev("[WORKER][StopMining][Signal sent]")

	select {
	case w.stopMining <- true:
	default:
	}
}

func (w *Worker) ShareTx(tx database.BlockTx) {
	w.ev("[WORKER][ShareTx][Sharing tx]")
	defer w.ev("[WORKER][ShareTx][Tx shared]")

	select {
	case w.shareTx <- tx:
	default:
	}
}

func (w *Worker) Sync() {
	w.ev("[WORKER][Sync][Started]")
	defer w.ev("[WORKER][Sync][Finished]")

	for _, peer := range w.state.ExternalPeers() {

		// Send request to get status of external peer.
		w.ev("[WORKER][Sync][Requesting status from peer: %s]", peer.Host)
		status, err := w.state.RequestPeerStatus(peer)
		if err != nil {
			// We have received an error, so we should delete the peer as
			// there is a high chance requested peer is not reachable anymore.
			// This is not a critical issue.
			w.ev("[WORKER][Sync][Status request for peer: %s failed: %s]", peer.Host, err)
			w.state.DeletePeer(peer)
			continue
		}

		// Update the list of peers known to the current node with the list of peers from other node.
		for _, knownPeer := range status.KnownPeers {
			w.state.AddPeer(knownPeer)
		}

		w.ev("[WORKER][Sync][Requesting tx from peer: %s]", peer.Host)
		txs, err := w.state.RequestPeerMempool(peer)
		if err != nil {
			w.ev("[WORKER][Sync][Tx request for peer: %s failed: %s]", peer.Host, err)
		}

		w.ev("[WORKER][Sync][Upserting tx from peer: %s]", peer.Host)
		for _, tx := range txs {
			if err = w.state.UpsertNodeTx(tx); err != nil {
				w.ev("[WORKER][Sync][Tx upsert for peer: %s failed: %s]", peer.Host, err)
			}
		}

		w.ev("[WORKER][Sync][Requesting blocks from peer: %s]", peer.Host)
		blocks, err := w.state.RequestPeerBlocks(peer)
		if err != nil {
			w.ev("[WORKER][Sync][Blocks request for peer: %s failed: %s]", peer.Host, err)
		}

		w.ev("[WORKER][Sync][Processing blocks from peer: %s]", peer.Host)
		for _, block := range blocks {
			if err = w.state.ProcessBlock(block); err != nil {
				w.ev("[WORKER][Sync][Block process for peer: %s failed: %s]", peer.Host, err)
			}
		}
	}

	w.state.SendNodeReady()
}

func (w *Worker) miningListener() {
	w.ev("[WORKER][miningListener][Started]")
	defer w.ev("[WORKER][miningListener][Stopped]")

	for {
		select {
		case <-w.startMining:
			w.ev("[WORKER][miningListener][Received start mining signal]")

			if !w.isShutdown() {
				w.runMining()
			}
		case <-w.shutdown:
			w.ev("[WORKER][miningListener][Received shutdown signal]")
			return
		}
	}
}

func (w *Worker) peerSyncer() {
	w.ev("[WORKER][peerSyncer][Started]")
	defer w.ev("[WORKER][peerSyncer][Stopped]")

	w.runPeerSync()

	for {
		select {
		case <-w.ticker.C:
			w.ev("[WORKER][peerSyncer][Received peer sync signal]")

			if !w.isShutdown() {
				w.runPeerSync()
			}
		case <-w.shutdown:
			w.ev("[WORKER][peerSyncer][Received shutdown signal]")
			return
		}
	}
}

func (w *Worker) txSyncer() {
	w.ev("[WORKER][txSyncer][Started]")
	defer w.ev("[WORKER][txSyncer][Stopped]")

	for {
		select {
		case tx := <-w.shareTx:
			w.ev("[WORKER][txSyncer][Received tx share signal]")

			if !w.isShutdown() {
				w.runTxShare(tx)
			}
		case <-w.shutdown:
			w.ev("[WORKER][txSyncer][Received shutdown signal]")
			return
		}
	}
}

func (w *Worker) runMining() {
	w.ev("[WORKER][runMining][Started new mining]")
	defer w.ev("[WORKER][runMining][Mining finished]")

	// Ensure there are transactions in the mempool.
	if size := w.state.MempoolSize(); size == 0 {
		w.ev("[WORKER][runMining][Mempool is empty]")
		return
	}

	//return // uncomment it if you want to skip mining

	// Verify whether new mining should be started at the end of current mining.
	defer func() {
		if size := w.state.MempoolSize(); size > 0 {
			w.ev("[WORKER][runMining][Signaling new mining]")
			w.StartMining()
		}
	}()

	select {
	case <-w.stopMining:
		w.ev("[WORKER][runMining][Draining stop mining channel]")
	default:
	}

	// Initialize new context to make possible to cancel mining.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize new wait group to make sure this function does not return
	// until both schedule goroutines are done.
	var wg sync.WaitGroup
	wg.Add(2)

	// Schedule ctx cancellation aware goroutine.
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		// This select statement will block until one of the signals appear.
		// After the signaling happen the goroutine will terminate causing
		// execution of the defer statement, where the ctx cancellation
		// function has been placed.
		select {
		case <-w.stopMining:
			w.ev("[WORKER][runMining][CtxAwareG][Received stop mining signal]")
		case <-ctx.Done():
			w.ev("[WORKER][runMining][CtxAwareG][Received ctx done signal]")
		}
	}()

	// Schedule block mining goroutine.
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		block, err := w.state.MineBlock(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				w.ev("[WORKER][runMining][MiningG][Mining cancelled]")
			} else {
				w.ev("[WORKER][runMining][MiningG][Mining err: %w]", err)
			}
			return
		}
		w.ev("[WORKER][runMining][MiningG][Mining finished]")

		err = w.state.SendBlockToPeers(block)
		if err != nil {
			w.ev("[WORKER][runMining][MiningG][Sending block to peers failed: %s]", err)
		}
		w.ev("[WORKER][runMining][MiningG][Sending block to peers finished]")
	}()

	// Wait for both goroutines to terminate.
	wg.Wait()
}

func (w *Worker) runPeerSync() {
	w.ev("[WORKER][runPeerSync][Started new peer sync]")
	defer w.ev("[WORKER][runPeerSync][Peer sync finished]")

	for _, peer := range w.state.ExternalPeers() {

		// Send request to get status of external peer.
		status, err := w.state.RequestPeerStatus(peer)
		if err != nil {
			// We have received an error, so we should delete the peer as
			// there is a high chance requested peer is not reachable anymore.
			w.state.DeletePeer(peer)
			continue
		}

		// Update the list of peers known to the current node with the list of peers from other node.
		for _, knownPeer := range status.KnownPeers {
			w.state.AddPeer(knownPeer)
		}
	}
}

func (w *Worker) runTxShare(tx database.BlockTx) {
	w.ev("[WORKER][runTxShare][Started new tx share]")
	defer w.ev("[WORKER][runTxShare][Tx share finished]")

	err := w.state.SendTxToPeers(tx)
	if err != nil {
		w.ev("[WORKER][runTxShare][Sending tx to peers failed: %s]", err)
	}
}

func (w *Worker) isShutdown() bool {
	select {
	case <-w.shutdown:
		return true
	default:
		return false
	}
}
