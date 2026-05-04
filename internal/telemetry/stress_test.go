package telemetry

import (
	"encoding/binary"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestTCPClient_ConcurrentEmitOrder exercises the sequencer mutex's core
// invariant: under contention from many goroutines, every Emit gets a
// unique monotonically-increasing event ID, and the wire frames preserve
// per-goroutine order (i.e. goroutine G's frames arrive on the wire in
// the same order G emitted them).
//
// Cross-goroutine wire order is NOT specified — that's by design — so we
// don't assert anything about which goroutine "wins" any given slot.
func TestTCPClient_ConcurrentEmitOrder(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Stop()

	cli, err := New(Config{
		Endpoint:     srv.Addr(),
		NodeInfo:     sampleNodeInfo(),
		BufferSize:   2048, // big enough so no drops (N=400 below)
		ReconnectMin: 10 * time.Millisecond,
		ReconnectMax: 50 * time.Millisecond,
		CloseTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if !waitEnabled(cli, 2*time.Second) {
		t.Fatalf("client never became Enabled")
	}

	const G = 8     // goroutines
	const PerG = 50 // emits per goroutine
	const N = G * PerG

	allIDs := make([][]uint64, G) // allIDs[g][k] = event ID of g's k-th emit
	for g := range allIDs {
		allIDs[g] = make([]uint64, PerG)
	}

	var wg sync.WaitGroup
	for g := 0; g < G; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for k := 0; k < PerG; k++ {
				// Payload encodes (goroutine_id, per_goroutine_counter)
				// so the server can recover per-goroutine order on the wire.
				payload := []byte{byte(g), byte(k >> 8), byte(k)}
				id := cli.Emit(7, payload)
				if id == InvalidID {
					t.Errorf("g=%d k=%d: Emit returned InvalidID", g, k)
					return
				}
				allIDs[g][k] = id
			}
		}(g)
	}
	wg.Wait()
	cli.Close()
	// Server's readLoop is concurrent; give it time to drain the kernel
	// receive buffer after the client has FIN'd.
	if !srv.WaitFrames(1+N, 5*time.Second) {
		t.Fatalf("server did not receive %d frames; got %d", 1+N, len(srv.Frames()))
	}

	frames := srv.Frames()

	// Property 1: all returned IDs are unique and densely cover [0, N).
	seen := make(map[uint64]bool, N)
	for g := 0; g < G; g++ {
		for k, id := range allIDs[g] {
			seq := eventIDSeq(id)
			if seq >= N {
				t.Errorf("g=%d k=%d: seq %d >= N=%d", g, k, seq, N)
			}
			if seen[seq] {
				t.Errorf("g=%d k=%d: duplicate seq %d", g, k, seq)
			}
			seen[seq] = true
		}
	}
	if len(seen) != N {
		t.Errorf("got %d unique seqs, want %d", len(seen), N)
	}

	// Property 2: per-goroutine order preserved on wire. Walk wire frames,
	// extract (g, k); verify each goroutine's k values appear in 0..PerG-1
	// ascending order on the wire.
	wireKByG := make([][]int, G)
	for _, f := range frames[1:] { // skip NodeInfo
		// frame body = u64 ts + u8 disc + payload
		if len(f) != 8+1+3 {
			continue // skip Dropped or unexpected
		}
		if f[8] == discriminatorDropped {
			continue
		}
		g := int(f[9])
		k := int(uint16(f[10])<<8 | uint16(f[11]))
		if g < 0 || g >= G {
			continue
		}
		wireKByG[g] = append(wireKByG[g], k)
	}
	for g, ks := range wireKByG {
		if len(ks) != PerG {
			t.Errorf("g=%d: got %d wire frames, want %d", g, len(ks), PerG)
			continue
		}
		for i := 1; i < len(ks); i++ {
			if ks[i] <= ks[i-1] {
				t.Errorf("g=%d: wire order broken at i=%d (k[%d]=%d, k[%d]=%d)",
					g, i, i-1, ks[i-1], i, ks[i])
				break
			}
		}
	}
}

// TestTCPClient_StressOverflow_WireAlignment is the Q3-critical test. With
// a small buffer + slow server we force the queue full, get Dropped events
// generated, then verify that on the wire the JamTART implicit-ID counter
// is preserved: the sum of (event_count + Dropped.count summed across all
// Dropped frames) equals the total number of event IDs the producer
// allocated. This is the central JIP-3 wire-side invariant for the foundation.
func TestTCPClient_StressOverflow_WireAlignment(t *testing.T) {
	// Slow server: reads frames with a small per-frame delay so the
	// client's send buffer fills up.
	srv := newSlowMockServer(t, 5*time.Millisecond)
	defer srv.Stop()

	cli, err := New(Config{
		Endpoint:         srv.Addr(),
		NodeInfo:         sampleNodeInfo(),
		BufferSize:       1, // smallest possible; under slow CI the producer can otherwise outpace the writer enough that the queue rarely fills
		ReconnectMin:     10 * time.Millisecond,
		ReconnectMax:     50 * time.Millisecond,
		CloseTimeout:     15 * time.Second, // generous: needs to drain
		TailDropInterval: 20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if !waitEnabled(cli, 2*time.Second) {
		t.Fatalf("client never became Enabled")
	}

	// Larger N (was 200) so even a slow CI hits the buffer-full path
	// reliably and the assertion at the end always sees gotDropped > 0.
	const N = 500
	allocated := 0
	for i := 0; i < N; i++ {
		// payload is just 2 bytes for compactness
		id := cli.Emit(7, []byte{byte(i >> 8), byte(i)})
		if id != InvalidID {
			allocated++
		}
	}
	cli.Close()

	// Wait for server to drain its receive buffer. We don't know exactly
	// how many frames will arrive (depends on drop coalescing), but we
	// know it'll be at least 2 (NodeInfo + 1 event-or-drop) and at most
	// 1+N. Give the slow server enough time to catch up.
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		f := srv.Frames()
		if len(f) >= 2 {
			// Sanity: total reflects N. Stop once it does.
			var ge int
			var gd uint64
			for _, fr := range f[1:] {
				if len(fr) >= 9 && fr[8] == discriminatorDropped {
					if len(fr) >= 17+8 {
						gd += binary.LittleEndian.Uint64(fr[17:25])
					}
				} else {
					ge++
				}
			}
			if uint64(ge)+gd >= uint64(allocated) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	frames := srv.Frames()
	if len(frames) < 2 {
		t.Fatalf("got %d frames; expected at least NodeInfo + something", len(frames))
	}

	var (
		gotEvents  int
		gotDropped uint64
	)
	for _, f := range frames[1:] { // skip NodeInfo
		if len(f) < 9 {
			t.Errorf("short frame: % x", f)
			continue
		}
		if f[8] == discriminatorDropped {
			// body = u64 firstTS + u8 disc + u64 lastTS + u64 count
			if len(f) != 8+1+8+8 {
				t.Errorf("Dropped frame wrong size: %d, want %d", len(f), 8+1+8+8)
				continue
			}
			count := binary.LittleEndian.Uint64(f[17:25])
			gotDropped += count
		} else {
			gotEvents++
		}
	}
	gotTotal := uint64(gotEvents) + gotDropped
	wantTotal := uint64(allocated)
	if gotTotal != wantTotal {
		t.Errorf(
			"wire alignment broken: events(%d) + dropped(%d) = %d, want allocated=%d",
			gotEvents, gotDropped, gotTotal, wantTotal,
		)
	}
	if gotDropped == 0 {
		t.Errorf("expected at least one drop with BufferSize=4 and slow server, got 0 (test setup may be too fast)")
	}
	t.Logf("alloc=%d events=%d dropped=%d total=%d (drop rate %.1f%%)",
		allocated, gotEvents, gotDropped, gotTotal,
		float64(gotDropped)*100.0/float64(allocated))
}

// TestTCPClient_CloseDrainsAllEnqueued verifies that Close blocks long
// enough for the writer to flush all queued envelopes (and any pending
// drops). With a generous CloseTimeout and a healthy server, no events
// should be silently lost.
func TestTCPClient_CloseDrainsAllEnqueued(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Stop()

	cli, err := New(Config{
		Endpoint:     srv.Addr(),
		NodeInfo:     sampleNodeInfo(),
		BufferSize:   256, // big enough that no drops happen
		ReconnectMin: 10 * time.Millisecond,
		ReconnectMax: 50 * time.Millisecond,
		CloseTimeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if !waitEnabled(cli, 2*time.Second) {
		t.Fatalf("client never became Enabled")
	}

	const N = 100
	enqueued := 0
	for i := 0; i < N; i++ {
		if id := cli.Emit(9, []byte{byte(i)}); id != InvalidID {
			enqueued++
		}
	}

	// Close should drain.
	if err := cli.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	// Wait for server to finish reading the kernel buffer post-FIN.
	if !srv.WaitFrames(1+enqueued, 5*time.Second) {
		t.Fatalf("server did not receive %d frames; got %d", 1+enqueued, len(srv.Frames()))
	}

	frames := srv.Frames()
	var gotEvents int
	var gotDropped uint64
	for _, f := range frames[1:] {
		if len(f) < 9 {
			continue
		}
		if f[8] == discriminatorDropped {
			gotDropped += binary.LittleEndian.Uint64(f[17:25])
		} else {
			gotEvents++
		}
	}
	total := uint64(gotEvents) + gotDropped
	if total != uint64(enqueued) {
		t.Errorf("Close lost events: enqueued=%d, on wire (events+dropped)=%d",
			enqueued, total)
	}
	if gotDropped != 0 {
		t.Errorf("expected no drops with healthy server + big buffer, got %d", gotDropped)
	}
}

// Sanity guard against test flakiness: with very tight timing, occasionally
// the slow server's queue read is fast enough that no drops happen. We
// already assert "expected at least one drop" in
// TestTCPClient_StressOverflow_WireAlignment; this helper documents the
// dependency for future maintainers.
var _ = atomic.Int32{}
