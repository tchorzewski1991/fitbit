package private

import (
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/network"
)

// status represents the details of the host node which
// will be serialized and moved over the wire.
type status struct {
	LastBlockHash   string      `json:"last_block_hash"`
	LastBlockHeight uint64      `json:"last_block_height"`
	KnownPeers      []knownPeer `json:"known_peers"`
}

func toStatus(lastBlock database.Block, peers []network.Peer) status {
	return status{
		LastBlockHash:   lastBlock.Hash(),
		LastBlockHeight: lastBlock.Height(),
		KnownPeers:      toKnownPeers(peers),
	}
}

// knownPeer represents the details of the peer which
// will be serialized and move over the wire.
type knownPeer struct {
	Host string `json:"host"`
}

func toKnownPeers(peers []network.Peer) []knownPeer {
	knownPeers := make([]knownPeer, len(peers))

	for idx := range peers {
		knownPeers[idx] = knownPeer{
			Host: peers[idx].Host,
		}
	}

	return knownPeers
}
