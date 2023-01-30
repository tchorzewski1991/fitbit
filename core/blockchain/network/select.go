package network

// SelectPeerFunc allows for defining peer selection strategy in easy and composable way.
type SelectPeerFunc func(peer Peer) bool

// SelectAllPeers is a peer selection strategy for selecting all known peers.
var SelectAllPeers = func() SelectPeerFunc {
	return func(_ Peer) bool {
		return true
	}
}

// SelectWithoutPeer is a peer selection strategy for selecting peers excluding one identifiable by given host.
var SelectWithoutPeer = func(host string) SelectPeerFunc {
	return func(peer Peer) bool {
		return peer.Host != host
	}
}
