package node

import (
	"errors"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestPeerCandidates_prefersPrimary(t *testing.T) {
	sm := NewSyncManager(nil, nil, nil)
	primary := &quic.Peer{ID: "primary", Ed25519Key: ed25519Key(1)}
	other := &quic.Peer{ID: "other", Ed25519Key: ed25519Key(2)}
	sm.peers["other"] = other

	got := sm.peerCandidates(primary)
	require.Len(t, got, 2)
	require.Equal(t, "primary", peerID(got[0]))
}

func TestFetchWithPeerFallback_triesNextPeer(t *testing.T) {
	sm := NewSyncManager(nil, nil, &quic.Peer{ID: "local"})
	bad := &quic.Peer{ID: "bad", Ed25519Key: ed25519Key(1)}
	good := &quic.Peer{ID: "good", Ed25519Key: ed25519Key(2)}
	sm.peers["bad"] = bad
	sm.peers["good"] = good

	sm.testFetchAndStore = func(peer *quic.Peer, _ types.HeaderHash, _ byte, _ uint32) (int, error) {
		if peerID(peer) == "bad" {
			return 0, errors.New("disconnect")
		}
		return 2, nil
	}

	imported, err := sm.fetchWithPeerFallback(bad, types.HeaderHash{}, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 2, imported)
}

func ed25519Key(b byte) []byte {
	key := make([]byte, 32)
	key[0] = b
	return key
}
