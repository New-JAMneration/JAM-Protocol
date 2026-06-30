package epochclock

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestTimeslotsFrom_usesHeadAndFinalized(t *testing.T) {
	blockchain.ResetInstance()
	chain := blockchain.GetInstance()
	genesis := types.Block{Header: types.Header{Slot: 0}, Extrinsic: types.Extrinsic{}}
	require.NoError(t, chain.GenerateGenesisBlock(genesis))
	t.Cleanup(func() { blockchain.ResetInstance() })

	ts := TimeslotsFrom(chain)
	require.Equal(t, types.TimeSlot(0), ts.BestHead)
	require.Equal(t, types.TimeSlot(0), ts.Finalized)

	chain.AddBlock(types.Block{
		Header: types.Header{Slot: 5, Parent: chain.GenesisBlockHash()},
	})
	ts = TimeslotsFrom(chain)
	require.Equal(t, types.TimeSlot(5), ts.BestHead)
}

func TestEpochTransitionDelaySlots(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })

	types.EpochLength = 12
	require.Equal(t, 1, EpochTransitionDelaySlots())

	types.EpochLength = 600
	require.Equal(t, 20, EpochTransitionDelaySlots())
}

func TestSafroleDelaySlots(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })

	types.EpochLength = 60
	require.Equal(t, 1, SafroleStep1DelaySlots())
	require.Equal(t, 3, SafroleStep2DelaySlots())
}

func TestCanApplyEpochTransition_respectsDelay(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })
	types.EpochLength = 12

	require.False(t, CanApplyEpochTransition(12, 1, 12))
	require.True(t, CanApplyEpochTransition(13, 1, 12))
}

func TestConnectivityApplied_SafroleDelays(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })
	types.EpochLength = 60

	applied := ConnectivityApplied{
		Epoch:          1,
		EpochStartSlot: 60,
		AppliedAtSlot:  65,
	}

	require.False(t, applied.CanSendSafroleStep1(65))
	require.True(t, applied.CanSendSafroleStep1(66))

	require.False(t, applied.CanForwardSafroleStep2(67))
	require.True(t, applied.CanForwardSafroleStep2(68))
}
