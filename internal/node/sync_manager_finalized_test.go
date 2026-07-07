package node

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestOnPeerUpdated_tracksNetworkFinalizedBest(t *testing.T) {
	blockchain.ResetInstance()
	chain := blockchain.GetInstance()
	require.NoError(t, chain.GenerateGenesisBlock(types.Block{
		Header:    types.Header{Slot: 0},
		Extrinsic: types.Extrinsic{},
	}))
	t.Cleanup(func() { blockchain.ResetInstance() })

	sm := NewSyncManager(chain, nil, nil)
	peer := &quic.Peer{
		ID: "peer-a",
		Finalized: quic.HeadInfo{
			Hash:     types.HeaderHash{1},
			Timeslot: 42,
		},
	}

	err := sm.onPeerUpdated(peer, nil)
	require.NoError(t, err)
	require.NotNil(t, sm.networkFinalizedBest)
	require.Equal(t, types.TimeSlot(42), sm.networkFinalizedBest.Timeslot)
}
