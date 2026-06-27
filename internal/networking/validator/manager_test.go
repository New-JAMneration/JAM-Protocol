package validator

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestValidatorManager_ShouldOpenUP0(t *testing.T) {
	var selfKey, neighborKey, strangerKey, fullNodePeer types.Ed25519Public
	selfKey[0] = 1
	neighborKey[0] = 2
	strangerKey[0] = 3
	fullNodePeer[31] = 99

	g := &GridMapper{
		Current: types.ValidatorsData{
			{Ed25519: selfKey},
			{Ed25519: neighborKey},
			{Ed25519: types.Ed25519Public{4}},
			{Ed25519: strangerKey},
		},
	}
	vm := &ValidatorManager{
		Grid:      g,
		SelfIndex: 0,
		SelfKey:   selfKey,
	}

	require.True(t, vm.ShouldOpenUP0(fullNodePeer, false), "full node opens to everyone")
	require.True(t, vm.ShouldOpenUP0(fullNodePeer, true), "unknown peer treated as non-validator")
	require.True(t, vm.ShouldOpenUP0(neighborKey, true), "grid neighbour validator")
	require.False(t, vm.ShouldOpenUP0(strangerKey, true), "non-neighbour validator")

	var nilVM *ValidatorManager
	require.False(t, nilVM.ShouldOpenUP0(neighborKey, true))
}

func TestGridMapper_IsKnownValidator(t *testing.T) {
	var curKey, prevKey types.Ed25519Public
	curKey[0] = 1
	prevKey[0] = 2

	g := &GridMapper{
		Previous: types.ValidatorsData{{Ed25519: prevKey}},
		Current:  types.ValidatorsData{{Ed25519: curKey}},
	}
	require.True(t, g.IsKnownValidator(curKey))
	require.True(t, g.IsKnownValidator(prevKey))

	var unknown types.Ed25519Public
	unknown[31] = 7
	require.False(t, g.IsKnownValidator(unknown))
}
