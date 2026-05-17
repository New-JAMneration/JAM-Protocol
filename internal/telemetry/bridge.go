package telemetry

import (
	"context"
	"log"
)

// Handler is the EventBus subscriber signature, aliased (not a named
// type) so a Bridge / BridgeLazy / BridgeFollowup return value is
// implicitly assignable to quic.Handler without explicit conversion at
// the callsite. The compile-time guard in bridge_eventbus_compat_test.go
// enforces this — if EventBus's signature ever drifts, that test file
// stops compiling and CI catches it before it lands.
type Handler = func(ctx context.Context, ev any) error

// Bridge wraps an encoder into an EventBus-compatible Handler that emits
// the encoded payload via tel.Emit. Use for cheap encoders.
//
// Three invariants enforced for every event:
//
//  1. Non-blocking — Emit is non-blocking already.
//  2. Always returns nil — handler MUST NOT propagate errors to the
//     publisher's main flow (telemetry is best-effort observability,
//     not a correctness path).
//  3. Swallows panic + encoder errors — recovered, logged, dropped.
//
// Disabled client → no-op.
//
// Panics on nil tel or nil encoder (fail-fast at construction; a
// runtime panic in the first published event would be slower to
// diagnose).
func Bridge(tel Client, disc uint8, encoder func(ev any) []byte) Handler {
	if tel == nil {
		panic("telemetry.Bridge: tel must not be nil")
	}
	if encoder == nil {
		panic("telemetry.Bridge: encoder must not be nil")
	}
	return func(_ context.Context, ev any) error {
		if !tel.Enabled() {
			return nil
		}
		defer recoverBridgePanic(disc)
		payload := encoder(ev)
		tel.Emit(disc, payload)
		return nil
	}
}

// BridgeLazy is Bridge with deferred encoding. snapshot runs
// synchronously inside the handler; encode runs on the writer goroutine.
//
// Use when:
//   - encoding is expensive (Block Outline, Cost structs) and the event
//     might be dropped, OR
//   - ev carries pointers whose targets the publisher may mutate after
//     Publish — snapshot captures the immutable shape now, encode runs
//     later against the snapshot.
//
// CRITICAL: snapshot must copy mutable fields out of ev. Returning ev
// itself (or a struct that shares pointers with ev) defeats the
// deferred-encoding guarantee — the publisher can mutate ev's pointees
// while encode is still queued in the writer goroutine, and you'll
// race.
//
//	Bad:  func(ev any) any { return ev }
//	Good: func(ev any) any { e := ev.(*Event); return EventSnap{Hash: e.Hash, Slot: e.Slot} }
//
// Non-blocking + always-returns-nil invariants are the same as Bridge.
// Panic semantics differ by phase:
//   - snapshot panics run inside the handler → swallowed by the bridge,
//     handler returns nil, event is dropped.
//   - encode panics run inside the writer goroutine — outside this
//     bridge's recover boundary. The writer's own recover marks the
//     client degraded (design doc invariant 3); telemetry stops for the
//     lifetime of the process. Treat encode like a hot path: handle nil
//     pointers, bound array indices, etc.
//
// Panics on nil tel / snapshot / encode at construction.
func BridgeLazy(tel Client, disc uint8, snapshot func(ev any) any, encode func(snap any) []byte) Handler {
	if tel == nil {
		panic("telemetry.BridgeLazy: tel must not be nil")
	}
	if snapshot == nil {
		panic("telemetry.BridgeLazy: snapshot must not be nil")
	}
	if encode == nil {
		panic("telemetry.BridgeLazy: encode must not be nil")
	}
	return func(_ context.Context, ev any) error {
		if !tel.Enabled() {
			return nil
		}
		defer recoverBridgePanic(disc)
		snap := snapshot(ev)
		tel.EmitLazy(disc, func() []byte {
			return encode(snap)
		})
		return nil
	}
}

// BridgeFollowup is Bridge for cause-effect chained events. extract
// runs at handler time and returns both the parent ID and the
// encoded payload for the same event — typical use is reading a
// request-tracking struct that carries the original event's ID:
//
//	telemetry.BridgeFollowup(tel, 66, func(ev any) (uint64, []byte) {
//	    e := ev.(*BlockRequestSent)
//	    return e.RequestParentID, encodeBlockRequestSent(e)
//	})
//
// EmitFollowup rejects stale parents (different epoch) internally;
// extract just needs to surface the parent the caller stored. Returning
// InvalidID short-circuits without emitting (no error to caller).
//
// Same non-blocking / return-nil / panic-swallow invariants as Bridge.
// Panics on nil tel / extract at construction.
func BridgeFollowup(tel Client, disc uint8, extract func(ev any) (parentID uint64, payload []byte)) Handler {
	if tel == nil {
		panic("telemetry.BridgeFollowup: tel must not be nil")
	}
	if extract == nil {
		panic("telemetry.BridgeFollowup: extract must not be nil")
	}
	return func(_ context.Context, ev any) error {
		if !tel.Enabled() {
			return nil
		}
		defer recoverBridgePanic(disc)
		parentID, payload := extract(ev)
		tel.EmitFollowup(disc, parentID, payload)
		return nil
	}
}

// recoverBridgePanic logs and swallows a panic from inside a Bridge
// handler. Returns to the deferred path; the handler still returns nil
// (unnamed return → zero-value error → nil).
//
// Bridge handlers must never break the publisher. A panic in the
// encoder, in a nil-deref inside snapshot, or in any user closure
// stops at this boundary.
func recoverBridgePanic(disc uint8) {
	if r := recover(); r != nil {
		log.Printf("telemetry: bridge handler panic (disc=%d): %v", disc, r)
	}
}
