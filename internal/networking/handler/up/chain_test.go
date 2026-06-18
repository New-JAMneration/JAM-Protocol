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

func testChain(t *testing.T) ([]types.Block, types.HeaderHash) {
	t.Helper()
	var genesis types.HeaderHash
	genesis[0] = 0x01

	finalizedHeader := types.Header{Parent: genesis, Slot: 10}
	finalized := mustHash(t, finalizedHeader)

	branchA := types.Header{Parent: finalized, Slot: 11}
	aHash := mustHash(t, branchA)
	bHeader := types.Header{Parent: aHash, Slot: 12}

	cHeader := types.Header{Parent: finalized, Slot: 13}
	_ = mustHash(t, cHeader)

	blocks := []types.Block{
		{Header: finalizedHeader},
		{Header: branchA},
		{Header: bHeader},
		{Header: cHeader},
	}
	return blocks, finalized
}

func TestCollectLeavesWithFork(t *testing.T) {
	blocks, finalized := testChain(t)
	cv, err := NewChainView(blocks)
	require.NoError(t, err)

	branchB := blocks[2].Header
	bHash := mustHash(t, branchB)
	branchC := blocks[3].Header
	cHash := mustHash(t, branchC)

	leaves := cv.CollectLeaves(finalized)
	require.Len(t, leaves, 2)

	got := make(map[types.HeaderHash]struct{}, 2)
	for _, leaf := range leaves {
		got[leaf.Hash] = struct{}{}
	}
	require.Contains(t, got, bHash)
	require.Contains(t, got, cHash)
}

func skipAnnouncementFixture(t *testing.T) (ChainView, types.HeaderHash, types.HeaderHash, types.HeaderHash) {
	t.Helper()
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
	return cv, finalized, aHash, bHash
}

func TestShouldSkipAnnouncement(t *testing.T) {
	t.Run("DescendantAlreadyAnnounced", func(t *testing.T) {
		cv, finalized, aHash, bHash := skipAnnouncementFixture(t)
		announcedByUs := map[types.HeaderHash]struct{}{bHash: {}}
		require.True(t, ShouldSkipAnnouncement(aHash, finalized, cv, announcedByUs, nil))
	})

	t.Run("PeerAlreadyAnnounced", func(t *testing.T) {
		cv, finalized, aHash, _ := skipAnnouncementFixture(t)
		announcedByPeer := map[types.HeaderHash]struct{}{aHash: {}}
		require.True(t, ShouldSkipAnnouncement(aHash, finalized, cv, nil, announcedByPeer))
	})

	t.Run("NotFinalizedDescendant", func(t *testing.T) {
		cv, finalized, _, _ := skipAnnouncementFixture(t)
		var orphan types.HeaderHash
		orphan[9] = 9
		require.True(t, ShouldSkipAnnouncement(orphan, finalized, cv, nil, nil))
	})

	t.Run("NoSkipRuleApplies", func(t *testing.T) {
		cv, finalized, aHash, _ := skipAnnouncementFixture(t)
		require.False(t, ShouldSkipAnnouncement(aHash, finalized, cv, nil, nil))
	})
}
