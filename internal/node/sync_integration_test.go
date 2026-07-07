package node

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

// TestBulkSync_catchesUpToNetworkBest exercises the multi-round bulk loop with a mocked fetch path.
func TestBulkSync_catchesUpToNetworkBest(t *testing.T) {
	blockchain.ResetInstance()
	chain := blockchain.GetInstance()
	genesis := types.Block{Header: types.Header{Slot: 0}, Extrinsic: types.Extrinsic{}}
	require.NoError(t, chain.GenerateGenesisBlock(genesis))
	t.Cleanup(func() { blockchain.ResetInstance() })

	sm := NewSyncManager(chain, nil, &quic.Peer{ID: "local"})
	sm.status = BulkSyncing
	sm.networkBest = &HeadInfo{Hash: types.HeaderHash{9}, Timeslot: 5}

	peer := &quic.Peer{ID: "peer", Ed25519Key: ed25519Key(1)}
	sm.peers["peer"] = peer

	var fetchCalls int
	sm.testFetch = func(_ *quic.Peer, _ types.HeaderHash, _ byte, _ uint32) (int, error) {
		fetchCalls++
		head, err := sm.getCurrentHead()
		if err != nil {
			return 0, err
		}
		if head.Timeslot >= sm.networkBest.Timeslot {
			return 0, nil
		}
		nextSlot := head.Timeslot + 1
		sm.advanceHeadForTest(nextSlot, head.Hash)
		return 1, nil
	}

	err := sm.bulkSync(peer)
	require.NoError(t, err)
	require.Equal(t, Syncing, sm.status)
	require.GreaterOrEqual(t, fetchCalls, 1)

	genesisHead, err := sm.getCurrentHead()
	require.NoError(t, err)
	// Rewind: bulkSync advanced to slot 5; verify we moved forward from genesis slot 0.
	require.Equal(t, types.TimeSlot(5), genesisHead.Timeslot)
}
