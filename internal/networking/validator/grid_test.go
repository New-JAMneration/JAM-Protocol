package validator

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestGridMapper_IsSameIndexCrossEpoch(t *testing.T) {
	var prevKey, nextKey, curKey types.Ed25519Public
	prevKey[0] = 1
	nextKey[0] = 2
	curKey[0] = 3

	g := &GridMapper{
		Previous: types.ValidatorsData{{Ed25519: prevKey}},
		Current:  types.ValidatorsData{{Ed25519: curKey}, {Ed25519: curKey}}, // 2 validators for width
		Next:     types.ValidatorsData{{Ed25519: nextKey}},
	}

	require.True(t, g.IsSameIndexCrossEpoch(0, prevKey), "Previous[0] key")
	require.True(t, g.IsSameIndexCrossEpoch(0, nextKey), "Next[0] key")
	require.False(t, g.IsSameIndexCrossEpoch(0, curKey), "curKey should not match Previous/Next slot unless equal")
	require.False(t, g.IsSameIndexCrossEpoch(1, prevKey), "index 1 out of range for single-entry Previous")
}

func TestValidatorManager_IsNeighbor_CrossEpochSameIndex(t *testing.T) {
	var peerPrev types.Ed25519Public
	peerPrev[0] = 42

	var selfCur types.Ed25519Public
	selfCur[1] = 7

	vm := &ValidatorManager{
		Grid: &GridMapper{
			Previous: types.ValidatorsData{{Ed25519: peerPrev}},
			Current:  types.ValidatorsData{{Ed25519: selfCur}},
			Next:     nil,
		},
		SelfIndex: 0,
		SelfKey:   selfCur,
	}

	require.True(t, vm.IsNeighbor(peerPrev), "same-index Previous validator not in Current")

	var stranger types.Ed25519Public
	stranger[31] = 99
	require.False(t, vm.IsNeighbor(stranger), "unknown key")
}

func TestGridMapper_AllNeighborValidators_includesCrossEpoch(t *testing.T) {
	var prevKey, selfKey types.Ed25519Public
	prevKey[0] = 11
	selfKey[0] = 22

	g := &GridMapper{
		Previous: types.ValidatorsData{{Ed25519: prevKey}},
		Current:  types.ValidatorsData{{Ed25519: selfKey}},
		Next:     nil,
	}
	neighbors := g.AllNeighborValidators(0)
	require.Len(t, neighbors, 1)
	require.Equal(t, prevKey, neighbors[0].Ed25519)
}
