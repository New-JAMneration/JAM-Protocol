package telemetry

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// envelope is the internal queue element. Producers build one under the
// sequencer mutex and send it on c.queue; the writer pops and serialises.
//
// Exactly one of payload / builder is set:
//   - payload, builder=nil → eager Emit. Bytes ready up front.
//   - payload=nil, builder → lazy EmitLazy. Builder runs on the writer
//     goroutine right before write; dropped envelopes pay no encoding cost.
//
// id is the packed event ID (epoch:16 | seq:48). Wire-side parent_id only
// carries the seq; the writer strips epoch via eventIDSeq.
//
// ts is microseconds (currently time.Now().UnixMicro()).
type envelope struct {
	id      uint64
	ts      uint64
	disc    uint8
	payload []byte
	builder func() []byte
}

// tcpClient is the concrete Client backed by one TCP connection. Owns
// the buffered envelope channel, the sequencer (epoch + seq, mutex),
// the drop state (mutex shared with sequencer), and the
// connection-management goroutine. Wire serialisation lives in writer.go.
type tcpClient struct {
	cfg   Config
	seq   *sequencer
	drops *dropState
	queue chan envelope

	// Atomic flags read by Emit*. enabledFlag flips with the connection
	// state; degradedFlag is set by the writer's recover() after a panic.
	enabledFlag  atomic.Bool
	degradedFlag atomic.Bool
	closedFlag   atomic.Bool

	// closeCh fires from Close(); writer drains, connectLoop exits.
	closeCh chan struct{}
	closed  sync.Once

	connWG sync.WaitGroup

	// dialCtx is cancelled by Close (or its CloseTimeout fallback) so an
	// in-progress DialContext returns immediately.
	dialCtx    context.Context
	dialCancel context.CancelFunc

	// currentConn is the live conn (nil between attempts). Close's force
	// path closes it to unblock writer / reader stuck in Write / Read.
	connMu      sync.Mutex
	currentConn net.Conn

	// dialer overrides net.Dialer.DialContext in tests.
	dialer func(ctx context.Context, addr string) (net.Conn, error)
}

// newTCPClient validates cfg, applies defaults, and constructs a client.
// Does not start any goroutine; call start() when ready.
func newTCPClient(cfg Config) (*tcpClient, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg = cfg.withDefaults()
	if cfg.Endpoint == "" {
		return nil, errors.New("telemetry: Endpoint must not be empty for tcpClient")
	}
	dialCtx, dialCancel := context.WithCancel(context.Background())
	return &tcpClient{
		cfg:        cfg,
		seq:        newSequencer(),
		drops:      &dropState{},
		queue:      make(chan envelope, cfg.BufferSize),
		closeCh:    make(chan struct{}),
		dialCtx:    dialCtx,
		dialCancel: dialCancel,
	}, nil
}

// start kicks off the connection-management goroutine. Split from
// newTCPClient so tests can inspect state before any wire activity.
func (c *tcpClient) start() {
	c.connWG.Add(1)
	go c.connectLoop()
}

// Enabled reports whether Emit* will accept new events.
func (c *tcpClient) Enabled() bool {
	if c.closedFlag.Load() {
		return false
	}
	if c.degradedFlag.Load() {
		return false
	}
	return c.enabledFlag.Load()
}

// Close drains the queue, flushes pending drops, and closes the
// connection. Bounded by CloseTimeout: on timeout, dialCancel + force-
// close currentConn unblock anything stuck in Dial / Write / Read,
// then a 1s grace window waits for goroutines to exit. Idempotent.
func (c *tcpClient) Close() error {
	c.closed.Do(func() {
		c.closedFlag.Store(true)
		c.enabledFlag.Store(false)
		close(c.closeCh)
	})

	done := make(chan struct{})
	go func() {
		c.connWG.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(c.cfg.CloseTimeout):
		c.dialCancel()
		c.connMu.Lock()
		conn := c.currentConn
		c.connMu.Unlock()
		if conn != nil {
			_ = conn.Close()
		}
		log.Printf("telemetry: close timeout after %s, forcing", c.cfg.CloseTimeout)
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			log.Printf("telemetry: goroutines still running after force-close")
		}
	}
	return nil
}

// Emit allocates an event ID, timestamps the event, and non-blockingly
// enqueues it. Channel full → record a drop range. Not enabled →
// InvalidID. Never blocks the caller.
func (c *tcpClient) Emit(disc uint8, payload []byte) uint64 {
	if !c.Enabled() {
		return InvalidID
	}
	c.seq.Lock()
	defer c.seq.Unlock()

	// Re-check under lock: a concurrent Close might have flipped the
	// flag between the outer check and Lock.
	if !c.Enabled() {
		return InvalidID
	}

	id := c.seq.nextID()
	ts := nowMicros()
	env := envelope{id: id, ts: ts, disc: disc, payload: payload}
	select {
	case c.queue <- env:
	default:
		c.drops.record(id, ts)
	}
	return id
}

// EmitLazy is Emit with deferred payload construction. The builder runs
// on the writer goroutine; drops never invoke it. nil builder →
// InvalidID.
func (c *tcpClient) EmitLazy(disc uint8, builder func() []byte) uint64 {
	if builder == nil {
		return InvalidID
	}
	if !c.Enabled() {
		return InvalidID
	}
	c.seq.Lock()
	defer c.seq.Unlock()

	if !c.Enabled() {
		return InvalidID
	}

	id := c.seq.nextID()
	ts := nowMicros()
	env := envelope{id: id, ts: ts, disc: disc, builder: builder}
	select {
	case c.queue <- env:
	default:
		c.drops.record(id, ts)
	}
	return id
}

// EmitFollowup prepends parent's seq (u64 LE) to payload and emits as a
// regular event. Returns InvalidID when parentID is InvalidID, stale
// (different epoch), or the client is not Enabled. Parent validation +
// child ID alloc + try-send are atomic under one sequencer lock; without
// that, a reconnect between validate and emit could let an epoch-N
// parent pair with an epoch-(N+1) child.
func (c *tcpClient) EmitFollowup(disc uint8, parentID uint64, payload []byte) uint64 {
	if parentID == InvalidID {
		return InvalidID
	}
	if !c.Enabled() {
		return InvalidID
	}
	full := make([]byte, 0, 8+len(payload))
	full = append(full, EncodeU64(eventIDSeq(parentID))...)
	full = append(full, payload...)
	return c.emitFollowupAtomic(disc, parentID, full, nil)
}

// EmitFollowupLazy is the lazy variant of EmitFollowup. The builder
// produces only the discriminator-specific payload; the parent_id u64
// prefix is added by the implementation. Same atomicity as EmitFollowup.
func (c *tcpClient) EmitFollowupLazy(disc uint8, parentID uint64, builder func() []byte) uint64 {
	if parentID == InvalidID || builder == nil {
		return InvalidID
	}
	if !c.Enabled() {
		return InvalidID
	}
	parentSeq := eventIDSeq(parentID)
	wrapped := func() []byte {
		inner := builder()
		out := make([]byte, 0, 8+len(inner))
		out = append(out, EncodeU64(parentSeq)...)
		return append(out, inner...)
	}
	return c.emitFollowupAtomic(disc, parentID, nil, wrapped)
}

// emitFollowupAtomic does Lock + validateParent + nextID + try-send (or
// drop record) + Unlock as one critical section. Exactly one of payload
// / builder is set. parent_id prefix is the caller's job.
func (c *tcpClient) emitFollowupAtomic(disc uint8, parentID uint64, payload []byte, builder func() []byte) uint64 {
	c.seq.Lock()
	defer c.seq.Unlock()

	if !c.Enabled() {
		return InvalidID
	}
	if !c.seq.validateParentLocked(parentID) {
		return InvalidID
	}

	id := c.seq.nextID()
	ts := nowMicros()
	env := envelope{id: id, ts: ts, disc: disc, payload: payload, builder: builder}
	select {
	case c.queue <- env:
	default:
		c.drops.record(id, ts)
	}
	return id
}

// connectLoop dials, sends NodeInfo, runs writeLoop until disconnect or
// Close, then reconnects with exponential backoff. On disconnect (other
// than Close) it bumps epoch, resets drop state, and drains stale
// queued envelopes.
func (c *tcpClient) connectLoop() {
	defer c.connWG.Done()

	backoff := c.cfg.ReconnectMin
	for {
		if c.closedFlag.Load() {
			return
		}
		if c.degradedFlag.Load() {
			return
		}

		conn, err := c.dial()
		if err != nil {
			log.Printf("telemetry: dial %s: %v (retry in %s)", c.cfg.Endpoint, err, backoff)
			if !c.sleepOrClose(backoff) {
				return
			}
			backoff = nextBackoff(backoff, c.cfg.ReconnectMax)
			continue
		}

		// Register conn BEFORE writeNodeInfo so Close's force path can
		// find it: writeNodeInfo can block for minutes on a half-open
		// peer (TCP retransmits), and currentConn=nil during that window
		// would let connectLoop / reader outlive Close + grace.
		c.connMu.Lock()
		c.currentConn = conn
		c.connMu.Unlock()

		if err := c.writeNodeInfo(conn); err != nil {
			log.Printf("telemetry: send node info: %v", err)
			c.connMu.Lock()
			c.currentConn = nil
			c.connMu.Unlock()
			_ = conn.Close()
			if !c.sleepOrClose(backoff) {
				return
			}
			backoff = nextBackoff(backoff, c.cfg.ReconnectMax)
			continue
		}

		// Wire is up: producers may now Emit.
		backoff = c.cfg.ReconnectMin
		c.enabledFlag.Store(true)

		// Read-side liveness watcher. JamTART doesn't send anything, so
		// any Read error means the connection is gone — closing
		// peerClosed lets writeLoop exit even when no events are being
		// emitted (a passive writer can't see TCP half-closure without
		// traffic). Tracked by connWG so Close's "all goroutines
		// exited" contract holds across the conn.Close → Read-error
		// race.
		peerClosed := make(chan struct{})
		c.connWG.Add(1)
		go func() {
			defer c.connWG.Done()
			var buf [256]byte
			for {
				n, err := conn.Read(buf[:])
				if err != nil {
					close(peerClosed)
					return
				}
				// (0, nil) is permitted but discouraged. A wrapper Reader
				// could return it; sleep briefly to avoid a busy loop.
				if n == 0 {
					select {
					case <-peerClosed:
						return
					case <-time.After(10 * time.Millisecond):
					}
				}
			}
		}()

		err = c.writeLoop(conn, peerClosed)
		c.enabledFlag.Store(false)
		c.connMu.Lock()
		c.currentConn = nil
		c.connMu.Unlock()
		_ = conn.Close()

		if c.closedFlag.Load() {
			return
		}
		if c.degradedFlag.Load() {
			// Writer panicked. Stay degraded for the process lifetime
			// (design doc invariant 3).
			return
		}

		// Disconnect: bump epoch (in-flight parent IDs become stale),
		// reset drops, drain stale queue entries.
		if !c.seq.bumpEpoch() {
			log.Printf("telemetry: epoch counter exhausted (16-bit wrap); degrading client")
			c.degradedFlag.Store(true)
			return
		}
		c.seq.Lock()
		c.drops.reset()
		c.drainQueueLocked()
		c.seq.Unlock()

		log.Printf("telemetry: connection lost: %v (reconnect in %s)", err, backoff)
		if !c.sleepOrClose(backoff) {
			return
		}
		backoff = nextBackoff(backoff, c.cfg.ReconnectMax)
	}
}

// dial calls the configured dialer (or net.Dialer.DialContext), passing
// dialCtx so Close's force path can cancel an in-flight attempt.
func (c *tcpClient) dial() (net.Conn, error) {
	if c.dialer != nil {
		return c.dialer(c.dialCtx, c.cfg.Endpoint)
	}
	d := net.Dialer{Timeout: 10 * time.Second}
	return d.DialContext(c.dialCtx, "tcp", c.cfg.Endpoint)
}

// sleepOrClose waits for d or closeCh. Returns false on Close.
func (c *tcpClient) sleepOrClose(d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-c.closeCh:
		return false
	}
}

// drainQueueLocked empties c.queue without writing. Called under
// sequencer lock between connections so producers can't enqueue while we
// drain. Stale envelopes have no meaning on the new connection.
func (c *tcpClient) drainQueueLocked() {
	for {
		select {
		case <-c.queue:
		default:
			return
		}
	}
}

// nextBackoff doubles cur, capped at max.
func nextBackoff(cur, max time.Duration) time.Duration {
	next := cur * 2
	if next > max {
		next = max
	}
	return next
}

// nowMicros returns microseconds since Unix epoch. JIP-3 specifies "JAM
// Common Era" epoch; we use Unix μs until the spec reference is
// finalised. TODO(jip3-jam-common-era): adjust the offset.
func nowMicros() uint64 {
	return uint64(time.Now().UnixMicro())
}
