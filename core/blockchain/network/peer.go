package network

import "sync"

// Peer represents the details of the node on p2p network.
type Peer struct {
	Host string
}

// NewPeer constructs a new Peer.
func NewPeer(host string) Peer {
	return Peer{Host: host}
}

// Match checks whether given host matches this peer host.
func (p Peer) Match(host string) bool {
	return p.Host == host
}

// PeerSet represents the collection of all known Peer(s) on p2p network.
type PeerSet struct {
	mu  sync.RWMutex
	set map[Peer]struct{}
}

// NewPeerSet constructs a new PeerSet.
func NewPeerSet() *PeerSet {
	return &PeerSet{set: make(map[Peer]struct{})}
}

// Add registers given Peer to the peers set.
func (s *PeerSet) Add(peer Peer) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exist := s.set[peer]
	if !exist {
		s.set[peer] = struct{}{}
		return true
	}

	return false
}

// Delete removes given Peer from the peers set.
func (s *PeerSet) Delete(peer Peer) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exist := s.set[peer]
	if exist {
		delete(s.set, peer)
		return true
	}

	return false
}

// Peers returns the copy of known peers excluding the peer specified by given host.
func (s *PeerSet) Peers(selector SelectPeerFunc) []Peer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if selector == nil {
		selector = SelectAllPeers()
	}

	var peers []Peer

	for peer := range s.set {
		if selector(peer) {
			peers = append(peers, peer)
		}
	}

	return peers
}

type PeerStatus struct {
	LatestBlockHash   string `json:"latest_block_hash"`
	LatestBlockNumber uint64 `json:"latest_block_number"`
	KnownPeers        []Peer `json:"known_peers"`
}
