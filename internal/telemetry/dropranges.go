package telemetry

import "fmt"

// dropRange records one contiguous run of dropped event IDs. The
// wire-side Dropped event (discriminator 0) collapses the whole run
// into one summary: header timestamp = firstTS, payload = lastTS +
// count. JamTART advances its implicit counter by count when it sees
// the Dropped event, so the next event the sender writes lines up.
type dropRange struct {
	firstID uint64 // first dropped event ID in this run
	count   uint64 // number of consecutive IDs covered
	firstTS uint64 // microsecond timestamp of the first dropped event
	lastTS  uint64 // microsecond timestamp of the last dropped event
}

// dropState is an ordered list of disjoint dropRanges describing every
// dropped event since the connection started. Producers append via
// record; the writer flushes the head range as a Dropped event when
// it hits the corresponding ID gap in the queue.
//
// Per-connection: discarded on reconnect along with pending envelopes.
//
// Concurrency: not internally synchronised. Producers must hold the
// sequencer mutex during record (atomic with ID alloc + try-send), and
// the writer must hold it for peekFirst / popFirst.
type dropState struct {
	ranges []dropRange
}

// record adds a single dropped (id, ts). Contiguous with the tail
// range → coalesce; otherwise → append.
//
// Caller-enforced invariant: ids must be strictly increasing within a
// connection. Naturally true because the sequencer mutex makes Lock +
// nextID + (try-send | record) atomic. record also panics on
// violation rather than silently producing overlapping ranges (which
// would corrupt wire ordering). O(1).
func (d *dropState) record(id uint64, ts uint64) {
	if n := len(d.ranges); n > 0 {
		tail := &d.ranges[n-1]
		nextValidID := tail.firstID + tail.count // [firstID, firstID+count) is occupied
		if id < nextValidID {
			panic(fmt.Sprintf(
				"telemetry: dropState.record invariant broken: id=%d <= last occupied=%d (tail=%+v)",
				id, nextValidID-1, *tail,
			))
		}
		if id == nextValidID {
			tail.count++
			tail.lastTS = ts
			return
		}
	}
	d.ranges = append(d.ranges, dropRange{
		firstID: id,
		count:   1,
		firstTS: ts,
		lastTS:  ts,
	})
}

// peekFirst returns a copy of the head range (zero + false when empty).
// By value so the writer doesn't pin the slice's backing array.
func (d *dropState) peekFirst() (dropRange, bool) {
	if len(d.ranges) == 0 {
		return dropRange{}, false
	}
	return d.ranges[0], true
}

// popFirst removes the head range. Caller must have flushed it. Zeroes
// the popped slot defensively (cheap; hardens future field additions).
func (d *dropState) popFirst() {
	if len(d.ranges) == 0 {
		return
	}
	d.ranges[0] = dropRange{}
	d.ranges = d.ranges[1:]
}

// empty reports whether there are no pending drop ranges.
func (d *dropState) empty() bool { return len(d.ranges) == 0 }

// reset clears all pending ranges. Used on reconnect: the whole drop
// state belongs to the old epoch.
func (d *dropState) reset() {
	d.ranges = d.ranges[:0]
}

// totalDropped sums count across all ranges. O(len(ranges)).
func (d *dropState) totalDropped() uint64 {
	var total uint64
	for _, r := range d.ranges {
		total += r.count
	}
	return total
}
