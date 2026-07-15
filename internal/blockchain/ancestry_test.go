package blockchain

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestAppendCommittedAncestryStartsEmptyAndDeduplicates(t *testing.T) {
	ResetInstance()
	t.Cleanup(ResetInstance)
	cs := GetInstance()
	header := types.Header{Slot: 7}
	headerHash := types.HeaderHash{0x11}
	stateRoot := types.StateRoot{0x22}

	cs.appendCommittedAncestry(header, headerHash, stateRoot)
	cs.appendCommittedAncestry(header, headerHash, stateRoot)

	got := cs.GetAncestry()
	if len(got) != 1 {
		t.Fatalf("ancestry length = %d, want 1", len(got))
	}
	if got[0].Slot != header.Slot || got[0].HeaderHash != headerHash || got[0].StateRoot != stateRoot {
		t.Fatalf("ancestry item = %+v, want slot/hash/state root from committed block", got[0])
	}
}

func TestAncestryCacheKeepsMostRecentLookupWindow(t *testing.T) {
	cache := NewAncestryCache()
	items := make(types.Ancestry, types.MaxLookupAge+1)
	for i := range items {
		items[i].Slot = types.TimeSlot(i)
	}

	cache.AppendAncestry(items)

	got := cache.GetAncestry()
	if len(got) != types.MaxLookupAge {
		t.Fatalf("ancestry length = %d, want %d", len(got), types.MaxLookupAge)
	}
	if got[0].Slot != 1 {
		t.Fatalf("oldest retained slot = %d, want 1", got[0].Slot)
	}
	if got[len(got)-1].Slot != types.TimeSlot(types.MaxLookupAge) {
		t.Fatalf("newest retained slot = %d, want %d", got[len(got)-1].Slot, types.MaxLookupAge)
	}
}
