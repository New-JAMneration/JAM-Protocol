package topology

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/validator"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

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

func TestDesiredKeySet_usesFullTransportTargets(t *testing.T) {
	var selfKey, peerA, peerB types.Ed25519Public
	selfKey[0] = 1
	peerA[0] = 2
	peerB[0] = 3

	grid := &validator.GridMapper{
		Previous: types.ValidatorsData{{Ed25519: peerB}},
		Current:  types.ValidatorsData{{Ed25519: selfKey}, {Ed25519: peerA}},
		Next:     types.ValidatorsData{{Ed25519: peerA}, {Ed25519: peerB}},
	}
	targets := validator.TransportTargets(grid, selfKey)
	require.Len(t, targets, 2)

	desired := desiredKeySet(targets, selfKey)
	require.Len(t, desired, 2)
}
