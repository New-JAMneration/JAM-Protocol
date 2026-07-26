package node

import (
	"context"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestSyncManager_storeBlocks_publishesBlockImported(t *testing.T) {
	blockchain.ResetInstance()
	chain := blockchain.GetInstance()
	genesis := types.Block{Header: types.Header{Slot: 0}, Extrinsic: types.Extrinsic{}}
	require.NoError(t, chain.GenerateGenesisBlock(genesis))
	t.Cleanup(func() { blockchain.ResetInstance() })

	eventBus := quic.NewEventBus()
	imported := make(chan quic.HeadInfo, 1)
	eventBus.Subscribe(quic.BlockImported, func(ctx context.Context, event quic.Event) error {
		ev, ok := event.(*quic.BlockImportedEvent)
		if !ok {
			t.Fatalf("unexpected event type %T", event)
		}
		imported <- ev.Head
		return nil
	})

	sm := NewSyncManager(chain, eventBus, nil)
	block := types.Block{
		Header: types.Header{Slot: 3, Parent: chain.GenesisBlockHash()},
	}
	require.NoError(t, sm.storeBlocks([]types.Block{block}))

	select {
	case head := <-imported:
		require.Equal(t, types.TimeSlot(3), head.Timeslot)
	default:
		t.Fatal("expected BlockImported event")
	}
}
