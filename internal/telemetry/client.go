// Package telemetry implements a JIP-3 telemetry client streaming node
// events to JamTART (or any compatible aggregator) over TCP.
//
// JIP-3 spec: https://github.com/polkadot-fellows/JIPs/blob/main/JIP-3.md
// JamTART:    https://github.com/paritytech/jamtart
//
// Wire protocol: length-prefixed messages on a single TCP connection.
// First message is Node Info (raw struct, no header). Every event after
// that is Timestamp + Discriminator + JAM-encoded payload.
//
// Usage:
//
//	tel, _ := telemetry.New(cfg)   // disabled if cfg.Endpoint == ""
//	defer tel.Close()
//	id := tel.Emit(disc, payload)
//	tel.EmitFollowup(disc, id, p)
//
// Design notes (event ID layout, drop window, bridge invariants) live
// in the planning comment on issue #775:
// https://github.com/New-JAMneration/JAM-Protocol/issues/775#issuecomment-4324368926
package telemetry

// Client is the telemetry interface used by the rest of the node.
// NewDisabled returns a no-op for when --telemetry is unset.
type Client interface {
	// Enabled reports whether Emit* will accept events. Guard expensive
	// payload construction with this, or use EmitLazy.
	Enabled() bool

	// Emit allocates an event ID, timestamps it, enqueues, and returns
	// the ID. Returns InvalidID if the client can't accept the event.
	// payload is the discriminator-specific bytes; the writer adds the
	// universal Timestamp + Discriminator header.
	//
	// The returned uint64 packs (epoch << 48) | seq. JamTART sees only
	// the 48-bit seq on the wire (parent_id fields carry seq alone), but
	// callers should store the full 64-bit value: epoch is needed by
	// EmitFollowup to detect stale parents from a previous connection.
	Emit(disc uint8, payload []byte) uint64

	// EmitLazy defers payload construction to the writer goroutine, so
	// dropped events pay no encoding cost.
	//
	// builder must be pure: its side effects may be silently skipped on
	// drop, disable, degrade, or reconnect-drain. Put metrics or resource
	// release at the callsite, not inside builder.
	EmitLazy(disc uint8, builder func() []byte) uint64

	// EmitFollowup emits an event whose payload is prefixed with parent's
	// seq (8 bytes LE). Returns InvalidID if parentID is InvalidID or
	// from a previous connection.
	EmitFollowup(disc uint8, parentID uint64, payload []byte) uint64

	// EmitFollowupLazy is the lazy variant of EmitFollowup. Same purity
	// contract as EmitLazy.
	EmitFollowupLazy(disc uint8, parentID uint64, builder func() []byte) uint64

	// Close drains the queue, flushes pending drop ranges as Dropped
	// events, and FINs the connection. Bounded by CloseTimeout (default
	// 5s); on timeout, dial / writer / reader are force-closed.
	// Idempotent.
	Close() error
}

// InvalidID signals that no event ID could be allocated: client is
// disabled, between connections, degraded after a writer panic, or
// the parent ID is stale (epoch mismatch). Callers must not follow up
// on it; passing it as parentID short-circuits.
//
// Value is ^uint64(0). Layout helpers in sequencer.go.
const InvalidID uint64 = ^uint64(0)

// New returns a Client. Empty Endpoint → disabled (no-op). Otherwise
// starts a connection goroutine: dial, send NodeInfo, run writer,
// reconnect on disconnect with exponential backoff. Emit* before the
// first connection returns InvalidID.
//
// Errors only on invalid config (negative durations, inverted Reconnect
// bounds, negative BufferSize). Network failures surface via reconnect
// logs and Enabled().
func New(cfg Config) (Client, error) {
	if cfg.Endpoint == "" {
		return NewDisabled(), nil
	}
	tc, err := newTCPClient(cfg)
	if err != nil {
		return nil, err
	}
	tc.start()
	return tc, nil
}

// disabledClient is the no-op Client returned when --telemetry is unset.
type disabledClient struct{}

// NewDisabled returns a Client that ignores every Emit. Use in tests
// that don't exercise telemetry.
func NewDisabled() Client {
	return disabledClient{}
}

func (disabledClient) Enabled() bool { return false }

func (disabledClient) Emit(uint8, []byte) uint64 { return InvalidID }

func (disabledClient) EmitLazy(uint8, func() []byte) uint64 { return InvalidID }

func (disabledClient) EmitFollowup(uint8, uint64, []byte) uint64 { return InvalidID }

func (disabledClient) EmitFollowupLazy(uint8, uint64, func() []byte) uint64 { return InvalidID }

func (disabledClient) Close() error { return nil }
