package validator

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestMergeValidators_deduplicatesByEd25519Key(t *testing.T) {
	var k1, k2, k3 types.Ed25519Public
	k1[0] = 1
	k2[0] = 2
	k3[0] = 3

	prev := types.ValidatorsData{{Ed25519: k1}, {Ed25519: k2}}
	cur := types.ValidatorsData{{Ed25519: k2}, {Ed25519: k3}}
	next := types.ValidatorsData{{Ed25519: k1}, {Ed25519: k3}}

	merged := MergeValidators(prev, cur, next)
	require.Len(t, merged, 3)
	require.Equal(t, k1, merged[0].Ed25519)
	require.Equal(t, k2, merged[1].Ed25519)
	require.Equal(t, k3, merged[2].Ed25519)
}

func TestTransportTargets_returnsMergedEpochValidatorsMinusSelf(t *testing.T) {
	var selfKey, k2, k3 types.Ed25519Public
	selfKey[0] = 1
	k2[0] = 2
	k3[0] = 3

	grid := &GridMapper{
		Previous: types.ValidatorsData{{Ed25519: k3}},
		Current:  types.ValidatorsData{{Ed25519: selfKey}, {Ed25519: k2}},
		Next:     types.ValidatorsData{{Ed25519: k2}, {Ed25519: k3}},
	}
	targets := TransportTargets(grid, selfKey)
	require.Len(t, targets, 2)
	require.Equal(t, k3, targets[0].Ed25519)
	require.Equal(t, k2, targets[1].Ed25519)
}

func TestGossipPeerKeys_excludesSelf(t *testing.T) {
	var selfKey, neighborKey types.Ed25519Public
	selfKey[0] = 1
	neighborKey[0] = 2

	vm := &ValidatorManager{
		Grid: &GridMapper{
			Current: types.ValidatorsData{
				{Ed25519: selfKey},
				{Ed25519: neighborKey},
			},
		},
		SelfIndex: 0,
		SelfKey:   selfKey,
	}
	keys := GossipPeerKeys(vm)
	require.Len(t, keys, 1)
	require.Equal(t, neighborKey, keys[0])
}
