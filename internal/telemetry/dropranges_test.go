package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Use makeEventID throughout so the tests exercise the real packed layout,
// not raw integers. ts values are also synthetic (not real μs) but
// distinguishable.
func id(seq uint64) uint64 { return makeEventID(1, seq) }

// Single drop creates one range with count=1 and firstTS == lastTS.
func TestDropState_RecordSingle(t *testing.T) {
	var d dropState
	d.record(id(5), 100)

	require.False(t, d.empty())
	r, ok := d.peekFirst()
	require.True(t, ok)
	assert.Equal(t, id(5), r.firstID)
	assert.Equal(t, uint64(1), r.count)
	assert.Equal(t, uint64(100), r.firstTS)
	assert.Equal(t, uint64(100), r.lastTS)
}

// Consecutive drops within one run coalesce into a single range whose
// count grows and whose lastTS tracks the last drop. firstTS stays at the
// start of the run.
func TestDropState_RecordContiguousMerges(t *testing.T) {
	var d dropState
	d.record(id(10), 200)
	d.record(id(11), 210)
	d.record(id(12), 220)

	require.Len(t, d.ranges, 1, "three contiguous drops collapse to one range")
	r, _ := d.peekFirst()
	assert.Equal(t, id(10), r.firstID)
	assert.Equal(t, uint64(3), r.count)
	assert.Equal(t, uint64(200), r.firstTS)
	assert.Equal(t, uint64(220), r.lastTS)
}

// A success between two drops splits them into separate ranges. This is
// the case writer-side ordering depends on: each range corresponds to a
// distinct wire-side gap.
func TestDropState_NonContiguousAppends(t *testing.T) {
	var d dropState
	d.record(id(10), 100) // run 1: id 10
	// id 11 was a successful emit (not recorded here)
	d.record(id(12), 120) // run 2: id 12

	require.Len(t, d.ranges, 2)
	assert.Equal(t, id(10), d.ranges[0].firstID)
	assert.Equal(t, uint64(1), d.ranges[0].count)
	assert.Equal(t, id(12), d.ranges[1].firstID)
	assert.Equal(t, uint64(1), d.ranges[1].count)
}

// Mixed pattern: drop, success, drop, drop, success, drop.
// Producer sees: drop 100, success 101 (not recorded), drop 102, drop 103,
// success 104 (not recorded), drop 105. Expected ranges:
//
//	[{first=100, count=1}, {first=102, count=2}, {first=105, count=1}]
func TestDropState_InterleavedDropsAndSuccesses(t *testing.T) {
	var d dropState
	d.record(id(100), 1000)
	d.record(id(102), 1020)
	d.record(id(103), 1030)
	d.record(id(105), 1050)

	require.Len(t, d.ranges, 3)

	assert.Equal(t, id(100), d.ranges[0].firstID)
	assert.Equal(t, uint64(1), d.ranges[0].count)

	assert.Equal(t, id(102), d.ranges[1].firstID)
	assert.Equal(t, uint64(2), d.ranges[1].count)
	assert.Equal(t, uint64(1020), d.ranges[1].firstTS)
	assert.Equal(t, uint64(1030), d.ranges[1].lastTS)

	assert.Equal(t, id(105), d.ranges[2].firstID)
	assert.Equal(t, uint64(1), d.ranges[2].count)
}

// peekFirst on an empty state returns false; popFirst is a no-op.
func TestDropState_PeekPopEmpty(t *testing.T) {
	var d dropState
	require.True(t, d.empty())

	_, ok := d.peekFirst()
	require.False(t, ok)

	// Should not panic.
	d.popFirst()
	require.True(t, d.empty())
}

// peekFirst returns the head; popFirst removes it; repeat to drain.
func TestDropState_PeekPopLifecycle(t *testing.T) {
	var d dropState
	d.record(id(10), 100)
	d.record(id(20), 200) // non-contiguous: separate range
	d.record(id(30), 300)

	require.Len(t, d.ranges, 3)

	r1, _ := d.peekFirst()
	assert.Equal(t, id(10), r1.firstID)

	d.popFirst()
	require.Len(t, d.ranges, 2)

	r2, _ := d.peekFirst()
	assert.Equal(t, id(20), r2.firstID, "after popFirst, head is the next range")

	d.popFirst()
	d.popFirst()
	require.True(t, d.empty(), "drained")
}

// peekFirst returns by value: mutating the returned struct must not affect
// the stored range. (Defensive — guards against accidental aliasing.)
func TestDropState_PeekFirstByValue(t *testing.T) {
	var d dropState
	d.record(id(7), 700)

	r, _ := d.peekFirst()
	r.count = 999

	stored, _ := d.peekFirst()
	assert.Equal(t, uint64(1), stored.count, "external mutation must not change stored range")
}

// reset clears all pending ranges. Used on reconnect.
func TestDropState_Reset(t *testing.T) {
	var d dropState
	d.record(id(1), 10)
	d.record(id(3), 30) // non-contiguous
	d.record(id(4), 40)

	require.False(t, d.empty())
	d.reset()
	require.True(t, d.empty())
	_, ok := d.peekFirst()
	require.False(t, ok)

	// Reset state must accept new records normally.
	d.record(id(0), 0)
	require.Len(t, d.ranges, 1)
	r, _ := d.peekFirst()
	assert.Equal(t, id(0), r.firstID)
}

// totalDropped sums count across all ranges; useful for writer-side
// diagnostics and tail-drop flush decisions.
func TestDropState_TotalDropped(t *testing.T) {
	var d dropState
	assert.Equal(t, uint64(0), d.totalDropped())

	d.record(id(10), 100)
	d.record(id(11), 110) // merges into range 1: count = 2
	d.record(id(20), 200) // new range: count = 1
	d.record(id(21), 210) // merges into range 2: count = 2
	d.record(id(22), 220) // merges into range 2: count = 3

	require.Len(t, d.ranges, 2)
	assert.Equal(t, uint64(2), d.ranges[0].count)
	assert.Equal(t, uint64(3), d.ranges[1].count)
	assert.Equal(t, uint64(5), d.totalDropped())
}

// Drop-as-first-event edge case: id=0 (the first event in a connection)
// drops. firstID = 0 must be encoded correctly, not confused with a sentinel.
func TestDropState_DropAsFirstEvent(t *testing.T) {
	var d dropState
	d.record(id(0), 1)
	d.record(id(1), 2)
	d.record(id(2), 3)

	require.Len(t, d.ranges, 1)
	r, _ := d.peekFirst()
	assert.Equal(t, id(0), r.firstID, "firstID can be the very first event ID")
	assert.Equal(t, uint64(3), r.count)
}

// Many alternating drops produce one range per drop (worst case for
// memory). This isn't a property test — just confirms slice growth works.
func TestDropState_HighlyInterleaved(t *testing.T) {
	var d dropState
	const n = 100
	// Drops at every even seq: 0, 2, 4, ..., 198. Each is its own range
	// because odd seqs (successes) split them.
	for i := uint64(0); i < n; i++ {
		d.record(id(i*2), i*10)
	}
	require.Len(t, d.ranges, n)
	assert.Equal(t, uint64(n), d.totalDropped())
}

// ---------------------------------------------------------------------------
// Runtime invariant guard: dropState.record panics on non-increasing id.
// The producer-side sequencer mutex enforces strictly-increasing ids in
// production, but a future refactor breaking that invariant would
// silently produce overlapping ranges and corrupt wire ordering — the
// panic surfaces such a regression at test time.
// ---------------------------------------------------------------------------

func TestDropState_RecordPanicsOnDuplicateID(t *testing.T) {
	d := &dropState{}
	d.record(makeEventID(1, 5), 100)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on duplicate id")
		}
	}()
	d.record(makeEventID(1, 5), 200) // same id, invariant violated
}

func TestDropState_RecordPanicsOnDecreasingID(t *testing.T) {
	d := &dropState{}
	d.record(makeEventID(1, 10), 100)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on decreasing id")
		}
	}()
	d.record(makeEventID(1, 5), 200) // earlier id, invariant violated
}

func TestDropState_RecordPanicsOnInsideExistingRange(t *testing.T) {
	d := &dropState{}
	// Record contiguous: 5, 6, 7 -> one range {firstID:5, count:3}.
	d.record(makeEventID(1, 5), 100)
	d.record(makeEventID(1, 6), 200)
	d.record(makeEventID(1, 7), 300)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on id inside existing range")
		}
	}()
	d.record(makeEventID(1, 6), 400) // inside [5, 8), invariant violated
}

func TestDropState_RecordAcceptsValidContiguousAndDisjoint(t *testing.T) {
	d := &dropState{}
	// Mix of contiguous merges and disjoint appends.
	d.record(makeEventID(1, 5), 100)
	d.record(makeEventID(1, 6), 200)  // contiguous, merge
	d.record(makeEventID(1, 9), 300)  // disjoint, new range
	d.record(makeEventID(1, 10), 400) // contiguous to second range
	if len(d.ranges) != 2 {
		t.Errorf("got %d ranges, want 2", len(d.ranges))
	}
	if d.ranges[0].count != 2 {
		t.Errorf("range 0 count = %d, want 2", d.ranges[0].count)
	}
	if d.ranges[1].count != 2 {
		t.Errorf("range 1 count = %d, want 2", d.ranges[1].count)
	}
}
