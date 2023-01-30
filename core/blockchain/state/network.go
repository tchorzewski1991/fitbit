package state

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/network"
	"io"
	"net/http"
)

// RequestPeerStatus sends HTTP request to the given peer for the status info.
func (s *State) RequestPeerStatus(peer network.Peer) (network.PeerStatus, error) {
	s.ev("[STATE][RequestPeerStatus][Started new request to: %s]", peer.Host)
	defer s.ev("[STATE][RequestPeerStatus][Request to: %s finished]", peer.Host)

	response, err := sendRequest(http.MethodGet, fmt.Sprintf(peerStatusEndpoint, peer.Host), nil)
	if err != nil {
		s.ev("[STATE][RequestPeerStatus][Got request err: %s]", err)
		return network.PeerStatus{}, err
	}

	var ps network.PeerStatus
	err = json.Unmarshal(response, &ps)
	if err != nil {
		s.ev("[STATE][RequestPeerStatus][Got request err: %s]", err)
		return network.PeerStatus{}, err
	}

	return ps, nil
}

// RequestPeerMempool sends HTTP request to the given peer for the list of uncommited transactions.
func (s *State) RequestPeerMempool(peer network.Peer) ([]database.BlockTx, error) {
	s.ev("[STATE][RequestPeerTxs][Started new request to: %s]", peer.Host)
	defer s.ev("[STATE][RequestPeerTxs][Request to: %s finished]", peer.Host)

	response, err := sendRequest(http.MethodGet, fmt.Sprintf(peerMempoolEndpoint, peer.Host), nil)
	if err != nil {
		s.ev("[STATE][RequestPeerTxs][Got request err: %s]", err)
		return nil, err
	}

	var txs []database.BlockTx
	err = json.Unmarshal(response, &txs)
	if err != nil {
		s.ev("[STATE][RequestPeerTxs][Got request err: %s]", err)
		return nil, err
	}

	return txs, nil
}

// RequestPeerBlocks sends HTTP request to given peer for the list of last n blocks.
func (s *State) RequestPeerBlocks(peer network.Peer) ([]database.Block, error) {
	s.ev("[STATE][RequestPeerBlocks][Started new request to: %s]", peer.Host)
	defer s.ev("[STATE][RequestPeerBlocks][Request to: %s finished]", peer.Host)

	// TODO: make a comment why we have +1 here as it's important.
	from := s.LastBlock().Height()
	if from == 0 {
		from += 1
	}

	response, err := sendRequest(http.MethodGet, fmt.Sprintf(peerBlocksEndpoint, peer.Host, from), nil)
	if err != nil {
		s.ev("[STATE][RequestPeerBlocks][Got request err: %s]", err)
		return nil, err
	}

	var blocksData []database.BlockData
	err = json.Unmarshal(response, &blocksData)
	if err != nil {
		s.ev("[STATE][RequestPeerBlocks][Got request err: %s]", err)
		return nil, err
	}

	var blocks []database.Block
	for _, blockData := range blocksData {
		block, err := blockData.ToBlock()
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, err
}

// SendBlockToPeers sends HTTP request to all known peers with given block.
func (s *State) SendBlockToPeers(block database.Block) error {
	s.ev("[STATE][SendBlockToPeers][Sending started]")
	defer s.ev("[STATE][SendBlockToPeers][Sending finished]")

	data, err := json.Marshal(block.ToBlockData())
	if err != nil {
		s.ev("[STATE][SendBlockToPeers][Preparing block data to send failed: %s]", err)
		return err
	}

	for _, peer := range s.ExternalPeers() {
		s.ev("[STATE][SendBlockToPeers][Started new request to: %s]", peer.Host)
		{
			_, err = sendRequest(http.MethodPost, fmt.Sprintf(submitBlockEndpoint, peer.Host), data)
			if err != nil {
				s.ev("[STATE][SendBlockToPeers][Got request err: %s]", err)
				continue
			}
		}
		s.ev("[STATE][SendBlockToPeers][Request to: %s finished]", peer.Host)
	}

	return nil
}

// SendTxToPeers sends HTTP request to all known peers with given transaction.
func (s *State) SendTxToPeers(tx database.BlockTx) error {
	s.ev("[STATE][SendTxToPeers][Sending started]")
	defer s.ev("[STATE][SendTxToPeers][Sending finished]")

	data, err := json.Marshal(tx)
	if err != nil {
		s.ev("[STATE][SendTxToPeers][Preparing tx to send failed: %s]", err)
		return err
	}

	for _, peer := range s.ExternalPeers() {
		s.ev("[STATE][SendTxToPeers][Started new request to: %s]", peer.Host)
		{
			_, err = sendRequest(http.MethodPost, fmt.Sprintf(submitTxEndpoint, peer.Host), data)
			if err != nil {
				s.ev("[STATE][SendTxToPeers][Got request err: %s]", err)
				continue
			}
		}
		s.ev("[STATE][SendTxToPeers][Request to: %s finished]", peer.Host)
	}

	return nil
}

// SendNodeReady sends HTTP request to all knows peers with info about node readiness.
func (s *State) SendNodeReady() {
	s.ev("[STATE][SendNodeReady][Sending started]")
	defer s.ev("[STATE][SendNodeReady][Sending finished]")

	host := network.Peer{Host: s.host}
	data, err := json.Marshal(host)
	if err != nil {
		s.ev("[STATE][SendNodeReady][Failed to encode host: %s]", err)
		return
	}

	for _, peer := range s.ExternalPeers() {
		s.ev("[STATE][SendNodeReady][Started new request to: %s]", peer.Host)
		{
			_, err = sendRequest(http.MethodPost, fmt.Sprintf(submitPeerEndpoint, peer.Host), data)
			if err != nil {
				s.ev("[STATE][SendNodeReady][Got request err: %s]", err)
			}
		}
		s.ev("[STATE][SendNodeReady][Request to: %s finished]", peer.Host)
	}
}

// private API

const (
	peerStatusEndpoint  = "http://%s/v1/node/status"
	peerBlocksEndpoint  = "http://%s/v1/node/blocks/%d/latest"
	peerMempoolEndpoint = "http://%s/v1/node/tx/uncommited"
	submitBlockEndpoint = "http://%s/v1/node/block"
	submitPeerEndpoint  = "http://%s/v1/node/peer"
	submitTxEndpoint    = "http://%s/v1/node/tx"
)

func sendRequest(method string, url string, body []byte) ([]byte, error) {

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bs, nil
}
