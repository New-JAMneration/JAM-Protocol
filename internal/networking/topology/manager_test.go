package topology

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestEpochTransitionDelaySlots(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })

	types.EpochLength = 12
	require.Equal(t, 1, EpochTransitionDelaySlots())

	types.EpochLength = 600
	require.Equal(t, 20, EpochTransitionDelaySlots())
}

func TestCanApplyEpochTransition_respectsDelay(t *testing.T) {
	backup := types.EpochLength
	t.Cleanup(func() { types.EpochLength = backup })
	types.EpochLength = 12

	m := &Manager{
		pendingEpoch: &pendingEpochTransition{
			targetEpoch:    1,
			epochStartSlot: 12,
		},
	}

	block := types.Block{Header: types.Header{Slot: 12}}
	require.False(t, m.canApplyEpochTransition(block), "slot 12 is epoch start, delay not met")

	block.Header.Slot = 13
	require.True(t, m.canApplyEpochTransition(block), "slot 13 meets delay of 1 for E=12")
}
