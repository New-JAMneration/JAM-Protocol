package telemetry

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Event ID layout (helpers in sequencer.go: makeEventID, eventIDEpoch,
// eventIDSeq, plus the InvalidID constant in client.go).
// ---------------------------------------------------------------------------

// makeEventID + eventIDEpoch + eventIDSeq must be exact inverses across the
// full layout: 16-bit epoch in upper bits, 48-bit seq in lower bits.
func TestEventID_PackUnpack(t *testing.T) {
	tests := []struct {
		name  string
		epoch uint16
		seq   uint64
	}{
		{"all zero", 0, 0},
		{"epoch=1, seq=0 (sequencer initial)", 1, 0},
		{"epoch=0, seq=1", 0, 1},
		{"max epoch, max seq", 0xFFFF, eventIDSeqMask},
		{"middle bits set", 0x1234, 0xABCDEF123456},
		{"epoch=1, seq just below 2^48", 1, eventIDSeqMask},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id := makeEventID(tc.epoch, tc.seq)
			require.Equal(t, tc.epoch, eventIDEpoch(id), "epoch")
			require.Equal(t, tc.seq, eventIDSeq(id), "seq")
		})
	}
}

// seq above 2^48-1 must be silently truncated by the mask. This codifies
// the contract: callers that overflow 48-bit space lose the upper bits
// rather than corrupting the epoch field.
func TestEventID_SeqAboveMaskTruncated(t *testing.T) {
	overflow := (uint64(1) << eventIDSeqBits) | uint64(5) // bit 48 set + low bits 5
	id := makeEventID(7, overflow)

	require.Equal(t, uint16(7), eventIDEpoch(id), "epoch must not be polluted by overflow seq")
	require.Equal(t, uint64(5), eventIDSeq(id), "seq must be truncated to 48 bits")
}

// InvalidID corresponds to (epoch=0xFFFF, seq=2^48-1). A real connection
// would need 65 535 reconnects AND 281T events on the last connection to
// reach this point, so the sentinel is safe in any practical run.
func TestEventID_InvalidIDDecomposition(t *testing.T) {
	require.Equal(t, uint16(0xFFFF), eventIDEpoch(InvalidID))
	require.Equal(t, eventIDSeqMask, eventIDSeq(InvalidID))
}

// Layout sanity: the boundary between epoch and seq must be at bit 48.
// If someone changes the constant, this test fails before it can mask a
// silent layout shift.
func TestEventID_LayoutInvariants(t *testing.T) {
	require.Equal(t, 48, eventIDSeqBits, "seq is the lower 48 bits")
	require.Equal(t, uint64(0x0000FFFFFFFFFFFF), eventIDSeqMask)

	// Smallest non-zero epoch must not bleed into the seq field.
	id := makeEventID(1, 0)
	require.Equal(t, uint64(1)<<48, id, "epoch=1, seq=0 must be 0x0001_0000_0000_0000")
}

// ---------------------------------------------------------------------------
// Sequencer (state machine).
// ---------------------------------------------------------------------------

// First nextID call returns epoch=1, seq=0 — the JIP-3-mandated first event
// ID is 0 in JamTART's per-connection view, and our epoch starts at 1 so
// zero-initialized uint64 fields don't accidentally validate later.
func TestSequencer_StartsAtEpoch1Seq0(t *testing.T) {
	s := newSequencer()

	s.Lock()
	id := s.nextID()
	s.Unlock()

	require.Equal(t, uint16(1), eventIDEpoch(id))
	require.Equal(t, uint64(0), eventIDSeq(id))
}

// Sequential nextID calls within one connection produce seq 0, 1, 2, ...
// epoch unchanged.
func TestSequencer_SeqMonotonicWithinEpoch(t *testing.T) {
	s := newSequencer()

	const n = 5
	ids := make([]uint64, n)
	for i := 0; i < n; i++ {
		s.Lock()
		ids[i] = s.nextID()
		s.Unlock()
	}

	for i, id := range ids {
		assert.Equal(t, uint16(1), eventIDEpoch(id), "epoch unchanged within one connection")
		assert.Equal(t, uint64(i), eventIDSeq(id), "seq increments by 1 per nextID")
	}
}

// bumpEpoch advances epoch and resets seq counter to 0. Models a TCP
// reconnect: JamTART sees a fresh stream of seq 0, 1, 2, ... after Node
// Info, but our internal IDs carry the new epoch in the upper bits so old
// stored parent IDs are detectable as stale.
func TestSequencer_BumpEpochResetsSeq(t *testing.T) {
	s := newSequencer()

	// Burn 3 IDs in epoch 1.
	s.Lock()
	s.nextID()
	s.nextID()
	s.nextID()
	s.Unlock()

	s.bumpEpoch()

	s.Lock()
	id := s.nextID()
	s.Unlock()

	require.Equal(t, uint16(2), eventIDEpoch(id), "epoch must advance to 2")
	require.Equal(t, uint64(0), eventIDSeq(id), "seq must reset to 0")
}

// validateParent must accept current-epoch IDs, reject stale IDs, reject
// InvalidID, and reject the zero value (epoch=0 is never a real epoch).
func TestSequencer_ValidateParent(t *testing.T) {
	s := newSequencer()

	// Live ID from current epoch.
	s.Lock()
	live := s.nextID()
	s.Unlock()
	require.True(t, s.validateParent(live), "current-epoch ID must validate")

	// InvalidID sentinel.
	require.False(t, s.validateParent(InvalidID), "InvalidID must not validate")

	// Zero-init uint64 (epoch=0 is never used by a live sequencer).
	require.False(t, s.validateParent(uint64(0)), "zero value must not validate")

	// After reconnect, the previously live ID is now stale.
	s.bumpEpoch()
	require.False(t, s.validateParent(live), "ID from previous epoch must not validate")

	// The new connection's first ID validates.
	s.Lock()
	fresh := s.nextID()
	s.Unlock()
	require.True(t, s.validateParent(fresh), "current-epoch ID after reconnect must validate")
}

// snapshot reports the current state without consuming an ID.
func TestSequencer_Snapshot(t *testing.T) {
	s := newSequencer()
	epoch, nextSeq := s.snapshot()
	require.Equal(t, uint16(1), epoch)
	require.Equal(t, uint64(0), nextSeq)

	s.Lock()
	s.nextID()
	s.nextID()
	s.Unlock()

	epoch, nextSeq = s.snapshot()
	require.Equal(t, uint16(1), epoch)
	require.Equal(t, uint64(2), nextSeq, "next seq should be 2 after two allocations")

	s.bumpEpoch()
	epoch, nextSeq = s.snapshot()
	require.Equal(t, uint16(2), epoch)
	require.Equal(t, uint64(0), nextSeq, "seq resets on epoch bump")
}

// Concurrent producers each take Lock, allocate one ID, Unlock. After all
// goroutines complete, every ID must be unique and the seq range must be
// exactly [0, total). This is the producer-side correctness invariant for
// wire-order == event-ID-order.
func TestSequencer_ConcurrentAllocateUnique(t *testing.T) {
	s := newSequencer()

	const goroutines = 10
	const perGoroutine = 100
	const total = goroutines * perGoroutine

	var wg sync.WaitGroup
	var dupCount atomic.Int64
	seen := sync.Map{} // map[uint64]struct{}

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				s.Lock()
				id := s.nextID()
				s.Unlock()

				if _, loaded := seen.LoadOrStore(id, struct{}{}); loaded {
					dupCount.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	require.Equal(t, int64(0), dupCount.Load(), "no duplicate IDs across concurrent producers")

	// Count of unique IDs.
	count := 0
	seen.Range(func(_, _ any) bool { count++; return true })
	require.Equal(t, total, count)

	// Counter has advanced to total.
	_, nextSeq := s.snapshot()
	require.Equal(t, uint64(total), nextSeq)
}

// All concurrently allocated IDs share the same epoch (1) since no
// reconnect happened during the test.
func TestSequencer_ConcurrentAllocateAllSameEpoch(t *testing.T) {
	s := newSequencer()

	const goroutines = 8
	const perGoroutine = 50

	idsCh := make(chan uint64, goroutines*perGoroutine)
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				s.Lock()
				idsCh <- s.nextID()
				s.Unlock()
			}
		}()
	}
	wg.Wait()
	close(idsCh)

	for id := range idsCh {
		require.Equal(t, uint16(1), eventIDEpoch(id))
	}
}

// bumpEpoch refuses to wrap past the 16-bit ceiling. Wrapping to 0 would
// silently validate zero-value parent IDs (epoch 0 is the never-set
// sentinel) and could alias stale parents from epoch 0xFFFF. Connection
// loop must degrade the client when this returns false.
//
// Slow-ish because it actually loops 65534 bumps — the real coverage is
// worth the ~1ms cost vs mocking the ceiling.
func TestSequencer_BumpEpochRefusesWrap(t *testing.T) {
	s := newSequencer()
	// Initial epoch is 1. Each bump goes to 2, 3, ..., 65535.
	// That's 65534 successful bumps. The 65535th would wrap.
	for i := 0; i < 65534; i++ {
		if !s.bumpEpoch() {
			t.Fatalf("bump %d failed unexpectedly", i)
		}
	}
	ep, _ := s.snapshot()
	if ep != ^uint16(0) {
		t.Fatalf("after 65534 bumps, epoch = %d, want 65535", ep)
	}
	if s.bumpEpoch() {
		t.Errorf("expected bumpEpoch to refuse at epoch 65535 (would wrap to 0)")
	}
	// Sanity: epoch unchanged after refused bump.
	ep2, _ := s.snapshot()
	if ep2 != ^uint16(0) {
		t.Errorf("refused bump mutated state: epoch = %d, want 65535", ep2)
	}
}
