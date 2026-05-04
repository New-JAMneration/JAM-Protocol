package telemetry

import "sync"

// Event ID layout: uint64 = (epoch << 48) | seq.
//
// JamTART numbers events 0, 1, 2, ... per connection. The epoch in the
// upper 16 bits lets a parent ID stored across a reconnect (which resets
// JamTART's counter) be detected as stale — EmitFollowup rejects parents
// whose epoch ≠ current.
//
// epoch starts at 1, ++ per (re)connect; bumpEpoch refuses to wrap past
// 0xFFFF. seq is 48-bit (2^48 ≈ 281T events per connection). Wire only
// carries the seq half; epoch is internal.
const (
	eventIDSeqBits = 48
	eventIDSeqMask = (uint64(1) << eventIDSeqBits) - 1
)

// makeEventID packs (epoch, seq) into the public uint64. seq is masked
// to 48 bits; an overflowing seq silently truncates.
func makeEventID(epoch uint16, seq uint64) uint64 {
	return (uint64(epoch) << eventIDSeqBits) | (seq & eventIDSeqMask)
}

// eventIDEpoch returns the upper 16 bits of id.
func eventIDEpoch(id uint64) uint16 {
	return uint16(id >> eventIDSeqBits)
}

// eventIDSeq returns the lower 48 bits of id — the value JamTART
// derives from arrival order and the value sent in parent_id payloads.
func eventIDSeq(id uint64) uint64 {
	return id & eventIDSeqMask
}

// sequencer serializes event-ID allocation against drop-range recording
// so wire order matches event-ID order under concurrent producers.
//
// Producer rule: hold Lock from nextID through the matching try-send (or
// drop-range record). Without this, goroutine A could allocate ID 10,
// be preempted, B could allocate 11 and enqueue first, and the writer
// would see queue head 11 ≠ expectedWireID without a real drop —
// JamTART's correlation breaks.
//
// epoch starts at 1 so a zero-value uint64 (never-set struct field)
// can't accidentally validate as a real event ID. bumpEpoch on each
// reconnect resets seqCounter to 0, matching JamTART's per-connection
// view.
type sequencer struct {
	mu           sync.Mutex
	currentEpoch uint16 // guarded by mu
	seqCounter   uint64 // guarded by mu; next seq to allocate
}

// newSequencer returns a sequencer at epoch 1, seq 0.
func newSequencer() *sequencer {
	return &sequencer{currentEpoch: 1}
}

// Lock acquires the sequencer mutex. Producer-side callers must hold it
// from nextID through the matching try-send or drop-range record.
func (s *sequencer) Lock() { s.mu.Lock() }

// Unlock releases the sequencer mutex.
func (s *sequencer) Unlock() { s.mu.Unlock() }

// nextID allocates the next packed event ID. Caller must hold Lock.
func (s *sequencer) nextID() uint64 {
	seq := s.seqCounter
	s.seqCounter++
	return makeEventID(s.currentEpoch, seq)
}

// bumpEpoch advances epoch by 1 and resets seqCounter to 0. Returns
// false when the next epoch would wrap past 0xFFFF — connectLoop must
// then degrade the client. Wrapping to 0 would silently validate
// zero-value parent IDs and alias stale parents from epoch 0xFFFF.
//
// 16 bits = 65535 reconnects per process. With default 1s ReconnectMin,
// a flapping accept-then-close peer reaches it in ~18h; healthy
// operation never does.
//
// Acquires its own lock — must NOT be called while holding Lock.
func (s *sequencer) bumpEpoch() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentEpoch == ^uint16(0) {
		return false
	}
	s.currentEpoch++
	s.seqCounter = 0
	return true
}

// validateParent reports whether parentID is from the current
// connection. False for InvalidID, stale (previous epoch), and
// zero-value uint64 (epoch 0 is never used). Acquires its own lock.
//
// Follow-up emit paths use validateParentLocked: parent validation,
// ID allocation, and try-send must happen under one lock to keep a
// reconnect from slipping in between (which would let a parent from
// epoch N pair with a child in epoch N+1).
func (s *sequencer) validateParent(parentID uint64) bool {
	if parentID == InvalidID {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.validateParentLocked(parentID)
}

// validateParentLocked is the lock-not-acquired variant. Caller must
// hold Lock.
func (s *sequencer) validateParentLocked(parentID uint64) bool {
	if parentID == InvalidID {
		return false
	}
	return eventIDEpoch(parentID) == s.currentEpoch
}

// snapshot returns a copy of (epoch, nextSeq) for tests and diagnostics.
// Acquires its own lock.
func (s *sequencer) snapshot() (epoch uint16, nextSeq uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentEpoch, s.seqCounter
}
