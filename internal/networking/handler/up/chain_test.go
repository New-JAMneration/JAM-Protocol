package up

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/stretchr/testify/require"
)

func mustHash(t *testing.T, header types.Header) types.HeaderHash {
	t.Helper()
	h, err := hash.ComputeBlockHeaderHash(header)
	require.NoError(t, err)
	return h
}

func TestCollectLeaves_fork(t *testing.T) {
	var genesis types.HeaderHash
	genesis[0] = 0x01

	finalizedHeader := types.Header{Parent: genesis, Slot: 10}
	finalized := mustHash(t, finalizedHeader)

	branchA := types.Header{Parent: finalized, Slot: 11}
	aHash := mustHash(t, branchA)
	bHeader := types.Header{Parent: aHash, Slot: 12}
	bHash := mustHash(t, bHeader)

	cHeader := types.Header{Parent: finalized, Slot: 13}
	cHash := mustHash(t, cHeader)

	blocks := []types.Block{
		{Header: finalizedHeader},
		{Header: branchA},
		{Header: bHeader},
		{Header: cHeader},
	}

	cv, err := NewChainView(blocks)
	require.NoError(t, err)

	leaves := cv.CollectLeaves(finalized)
	require.Len(t, leaves, 2)

	got := make(map[types.HeaderHash]struct{}, 2)
	for _, leaf := range leaves {
		got[leaf.Hash] = struct{}{}
	}
	_, hasB := got[bHash]
	_, hasC := got[cHash]
	require.True(t, hasB, "expected tip b")
	require.True(t, hasC, "expected tip c")
}

func TestShouldSkipAnnouncement(t *testing.T) {
	var genesis types.HeaderHash
	genesis[0] = 1

	finalizedHeader := types.Header{Parent: genesis, Slot: 0}
	finalized := mustHash(t, finalizedHeader)

	aHeader := types.Header{Parent: finalized, Slot: 1}
	aHash := mustHash(t, aHeader)
	bHeader := types.Header{Parent: aHash, Slot: 2}
	bHash := mustHash(t, bHeader)

	blocks := []types.Block{
		{Header: finalizedHeader},
		{Header: aHeader},
		{Header: bHeader},
	}

	cv, err := NewChainView(blocks)
	require.NoError(t, err)

	announcedByUs := map[types.HeaderHash]struct{}{bHash: {}}
	require.True(t, ShouldSkipAnnouncement(aHash, finalized, cv, announcedByUs, nil),
		"skip when we already announced a descendant")

	announcedByPeer := map[types.HeaderHash]struct{}{aHash: {}}
	require.True(t, ShouldSkipAnnouncement(aHash, finalized, cv, nil, announcedByPeer),
		"skip when peer announced the block")

	var orphan types.HeaderHash
	orphan[9] = 9
	require.True(t, ShouldSkipAnnouncement(orphan, finalized, cv, nil, nil),
		"skip when block is not a finalized descendant")

	require.False(t, ShouldSkipAnnouncement(aHash, finalized, cv, nil, nil),
		"should announce when no skip rule applies")
}
