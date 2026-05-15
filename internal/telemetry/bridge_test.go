package telemetry

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Fake Client for tracking bridge behaviour. Not thread-safe across
// inspection / call (tests inspect only after the handler returns).
// ---------------------------------------------------------------------------

type fakeClient struct {
	enabled atomic.Bool

	emitCount     atomic.Int32
	lazyCount     atomic.Int32
	followupCount atomic.Int32

	// Last-seen fields are guarded by mu so the concurrent-invocation
	// test (and the race detector) is happy. Single-threaded tests can
	// read them after the handler returns without locking, but the
	// helpers below lock anyway to keep the API simple.
	mu               sync.Mutex
	emitLastDisc     uint8
	emitLastPayload  []byte
	lazyLastDisc     uint8
	lazyLastBuilder  func() []byte
	followupLastDisc uint8
	followupParent   uint64
	followupPayload  []byte
}

func newFakeClient(enabled bool) *fakeClient {
	c := &fakeClient{}
	c.enabled.Store(enabled)
	return c
}

func (c *fakeClient) Enabled() bool { return c.enabled.Load() }

func (c *fakeClient) Emit(disc uint8, payload []byte) uint64 {
	c.emitCount.Add(1)
	c.mu.Lock()
	c.emitLastDisc = disc
	c.emitLastPayload = append([]byte(nil), payload...)
	c.mu.Unlock()
	return uint64(c.emitCount.Load())
}

func (c *fakeClient) EmitLazy(disc uint8, builder func() []byte) uint64 {
	c.lazyCount.Add(1)
	c.mu.Lock()
	c.lazyLastDisc = disc
	c.lazyLastBuilder = builder
	c.mu.Unlock()
	return uint64(c.lazyCount.Load())
}

func (c *fakeClient) EmitFollowup(disc uint8, parentID uint64, payload []byte) uint64 {
	c.followupCount.Add(1)
	c.mu.Lock()
	c.followupLastDisc = disc
	c.followupParent = parentID
	c.followupPayload = append([]byte(nil), payload...)
	c.mu.Unlock()
	return uint64(c.followupCount.Load())
}

func (c *fakeClient) EmitFollowupLazy(disc uint8, parentID uint64, builder func() []byte) uint64 {
	return c.EmitFollowup(disc, parentID, builder())
}

// snapshot getters: single-threaded tests use them so race detector
// doesn't trip if some test we add later accidentally goes concurrent.
func (c *fakeClient) lastEmit() (disc uint8, payload []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.emitLastDisc, c.emitLastPayload
}
func (c *fakeClient) lastLazyBuilder() func() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lazyLastBuilder
}
func (c *fakeClient) lastFollowup() (disc uint8, parent uint64, payload []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.followupLastDisc, c.followupParent, c.followupPayload
}

func (c *fakeClient) Close() error { return nil }

// ---------------------------------------------------------------------------
// Bridge: happy path + invariants
// ---------------------------------------------------------------------------

func TestBridge_EncodesAndEmitsOnEnabledClient(t *testing.T) {
	fc := newFakeClient(true)
	h := Bridge(fc, 7, func(ev any) []byte {
		return []byte{byte(ev.(int))}
	})

	if err := h(context.Background(), 42); err != nil {
		t.Fatalf("handler returned err: %v", err)
	}
	if got := fc.emitCount.Load(); got != 1 {
		t.Errorf("emit count = %d, want 1", got)
	}
	disc, payload := fc.lastEmit()
	if disc != 7 {
		t.Errorf("disc = %d, want 7", disc)
	}
	if len(payload) != 1 || payload[0] != 42 {
		t.Errorf("payload = %v, want [42]", payload)
	}
}

// Invariant 1: handler never propagates errors to the publisher.
func TestBridge_ReturnsNilOnEncoderPanic(t *testing.T) {
	fc := newFakeClient(true)
	h := Bridge(fc, 7, func(ev any) []byte {
		panic("encoder boom")
	})

	if err := h(context.Background(), nil); err != nil {
		t.Errorf("handler returned err: %v (must always be nil)", err)
	}
	if got := fc.emitCount.Load(); got != 0 {
		t.Errorf("emit count after panic = %d, want 0", got)
	}
}

// Invariant 2: disabled client → no-op, no panic.
func TestBridge_NoOpOnDisabledClient(t *testing.T) {
	fc := newFakeClient(false)
	called := false
	h := Bridge(fc, 7, func(ev any) []byte {
		called = true
		return nil
	})

	if err := h(context.Background(), 42); err != nil {
		t.Errorf("handler returned err: %v", err)
	}
	if called {
		t.Errorf("encoder should not run when client disabled")
	}
	if got := fc.emitCount.Load(); got != 0 {
		t.Errorf("emit count = %d, want 0", got)
	}
}

// Invariant 3: handler returns synchronously even if downstream is slow.
// Bridge itself doesn't block; tel.Emit is non-blocking by contract.
// This test is a sanity check against a future refactor that
// accidentally introduces a sync call.
func TestBridge_NeverBlocks(t *testing.T) {
	fc := newFakeClient(true)
	h := Bridge(fc, 7, func(ev any) []byte { return nil })

	done := make(chan struct{})
	go func() {
		_ = h(context.Background(), 0)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("handler did not return within 500ms")
	}
}

// ---------------------------------------------------------------------------
// BridgeLazy: snapshot is captured at handler time, encode runs later.
// ---------------------------------------------------------------------------

type mutableEvent struct {
	value int
}

func TestBridgeLazy_SnapshotCapturedAtCallTime(t *testing.T) {
	fc := newFakeClient(true)

	h := BridgeLazy(fc, 9,
		func(ev any) any {
			e := ev.(*mutableEvent)
			return e.value // copy out by value
		},
		func(snap any) []byte {
			return []byte{byte(snap.(int))}
		},
	)

	ev := &mutableEvent{value: 42}
	if err := h(context.Background(), ev); err != nil {
		t.Fatalf("handler returned err: %v", err)
	}

	// Publisher mutates ev after Publish returns. snapshot should have
	// captured the pre-mutation value, so the deferred builder must
	// produce 42, not 99.
	ev.value = 99

	b := fc.lastLazyBuilder()
	if b == nil {
		t.Fatalf("builder was not enqueued")
	}
	out := b()
	if len(out) != 1 || out[0] != 42 {
		t.Errorf("snapshot leaked mutation: got %v, want [42]", out)
	}
}

func TestBridgeLazy_NoOpOnDisabledClient(t *testing.T) {
	fc := newFakeClient(false)
	snapCalls := 0
	encodeCalls := 0
	h := BridgeLazy(fc, 9,
		func(ev any) any { snapCalls++; return ev },
		func(snap any) []byte { encodeCalls++; return nil },
	)

	if err := h(context.Background(), 1); err != nil {
		t.Errorf("err: %v", err)
	}
	if snapCalls != 0 {
		t.Errorf("snapshot called %d times on disabled, want 0", snapCalls)
	}
	if encodeCalls != 0 {
		t.Errorf("encode called %d times on disabled, want 0", encodeCalls)
	}
}

func TestBridgeLazy_ReturnsNilOnSnapshotPanic(t *testing.T) {
	fc := newFakeClient(true)
	h := BridgeLazy(fc, 9,
		func(ev any) any { panic("snapshot boom") },
		func(snap any) []byte { return nil },
	)

	if err := h(context.Background(), 0); err != nil {
		t.Errorf("handler returned err: %v", err)
	}
	if got := fc.lazyCount.Load(); got != 0 {
		t.Errorf("lazy enqueued %d times after snapshot panic, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// BridgeFollowup
// ---------------------------------------------------------------------------

// BridgeFollowup takes an extract func that returns (parentID, payload)
// per event — the parent is captured per-call, not at construction.
// Real callsites read the parent from a request-tracking struct attached
// to ev (e.g. the original request's TelemetryEventID).
func TestBridgeFollowup_ExtractsParentPerCall(t *testing.T) {
	fc := newFakeClient(true)
	// Each event carries its own parent ID; extract surfaces it.
	type req struct {
		ParentID uint64
		Body     byte
	}
	h := BridgeFollowup(fc, 11, func(ev any) (uint64, []byte) {
		r := ev.(*req)
		return r.ParentID, []byte{r.Body}
	})

	if err := h(context.Background(), &req{ParentID: 0xAAAA, Body: 5}); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := h(context.Background(), &req{ParentID: 0xBBBB, Body: 9}); err != nil {
		t.Fatalf("second: %v", err)
	}

	if got := fc.followupCount.Load(); got != 2 {
		t.Fatalf("followup count = %d, want 2", got)
	}
	// Last followup must carry the per-call parent (0xBBBB), not the
	// first call's (0xAAAA).
	disc, parent, payload := fc.lastFollowup()
	if disc != 11 {
		t.Errorf("disc = %d, want 11", disc)
	}
	if parent != 0xBBBB {
		t.Errorf("parent = 0x%x, want 0xBBBB (per-call, not closure-captured)", parent)
	}
	if len(payload) != 1 || payload[0] != 9 {
		t.Errorf("payload = %v, want [9]", payload)
	}
}

func TestBridgeFollowup_NoOpOnDisabledClient(t *testing.T) {
	fc := newFakeClient(false)
	called := false
	h := BridgeFollowup(fc, 11, func(ev any) (uint64, []byte) {
		called = true
		return 1, nil
	})
	_ = h(context.Background(), 0)
	if called {
		t.Errorf("extract ran on disabled client")
	}
}

// Mirrors TestBridge_ReturnsNilOnEncoderPanic and
// TestBridgeLazy_ReturnsNilOnSnapshotPanic. The user-supplied closure
// (extract) sits inside the same defer recoverBridgePanic boundary as
// the other two helpers — verify symmetry.
func TestBridgeFollowup_ReturnsNilOnExtractPanic(t *testing.T) {
	fc := newFakeClient(true)
	h := BridgeFollowup(fc, 11, func(ev any) (uint64, []byte) {
		panic("extract boom")
	})

	if err := h(context.Background(), nil); err != nil {
		t.Errorf("handler returned err: %v (must always be nil)", err)
	}
	if got := fc.followupCount.Load(); got != 0 {
		t.Errorf("followup count after panic = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// Concurrent invocation: bridge handlers are entered from multiple
// goroutines (EventBus publishers across subsystems). Confirm
// concurrent Bridge calls don't race.
// ---------------------------------------------------------------------------

func TestBridge_ConcurrentInvocationsRaceClean(t *testing.T) {
	fc := newFakeClient(true)
	h := Bridge(fc, 7, func(ev any) []byte {
		return []byte{byte(ev.(int) & 0xFF)}
	})

	const G = 16
	const Per = 100
	var wg sync.WaitGroup
	for g := 0; g < G; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < Per; i++ {
				_ = h(context.Background(), g*Per+i)
			}
		}(g)
	}
	wg.Wait()

	if got := fc.emitCount.Load(); got != int32(G*Per) {
		t.Errorf("emit count = %d, want %d", got, G*Per)
	}
}
