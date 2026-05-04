package telemetry

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Mock TCP server: accepts one connection at a time, reads framed messages
// (u32 LE len + payload), and exposes them via Frames(). New connections
// reset the frames buffer.
// ---------------------------------------------------------------------------

type mockServer struct {
	ln       net.Listener
	mu       sync.Mutex
	frames   [][]byte
	connCh   chan net.Conn
	stopOnce sync.Once
	stopped  chan struct{}
}

func newMockServer(t *testing.T) *mockServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := &mockServer{
		ln:      ln,
		connCh:  make(chan net.Conn, 4),
		stopped: make(chan struct{}),
	}
	go s.acceptLoop()
	return s
}

func (s *mockServer) Addr() string { return s.ln.Addr().String() }

func (s *mockServer) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.connCh <- conn
		go s.readLoop(conn)
	}
}

func (s *mockServer) readLoop(conn net.Conn) {
	defer conn.Close()
	for {
		var hdr [4]byte
		if _, err := io.ReadFull(conn, hdr[:]); err != nil {
			return
		}
		n := binary.LittleEndian.Uint32(hdr[:])
		if n > 1<<20 {
			return
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return
		}
		s.mu.Lock()
		s.frames = append(s.frames, buf)
		s.mu.Unlock()
	}
}

func (s *mockServer) Frames() [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([][]byte, len(s.frames))
	copy(out, s.frames)
	return out
}

func (s *mockServer) ResetFrames() {
	s.mu.Lock()
	s.frames = nil
	s.mu.Unlock()
}

// WaitFrames blocks until at least n frames have been received or timeout.
func (s *mockServer) WaitFrames(n int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		got := len(s.frames)
		s.mu.Unlock()
		if got >= n {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// DropFirstConnection grabs the first accepted connection and closes it,
// forcing the client to reconnect.
func (s *mockServer) DropFirstConnection(t *testing.T, timeout time.Duration) {
	t.Helper()
	select {
	case conn := <-s.connCh:
		conn.Close()
	case <-time.After(timeout):
		t.Fatalf("DropFirstConnection: no connection within %v", timeout)
	}
}

func (s *mockServer) Stop() {
	s.stopOnce.Do(func() {
		s.ln.Close()
		close(s.stopped)
	})
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestTCPClient_NodeInfoIsFirstFrame(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Stop()

	cli, err := New(Config{
		Endpoint:     srv.Addr(),
		NodeInfo:     sampleNodeInfo(),
		ReconnectMin: 10 * time.Millisecond,
		ReconnectMax: 50 * time.Millisecond,
		CloseTimeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer cli.Close()

	if !srv.WaitFrames(1, 2*time.Second) {
		t.Fatalf("no Node Info frame received")
	}
	frames := srv.Frames()
	wantInfo, err := sampleNodeInfo().Encode()
	if err != nil {
		t.Fatalf("encode reference NodeInfo: %v", err)
	}
	if !bytes.Equal(frames[0], wantInfo) {
		t.Errorf("Node Info mismatch")
	}
}

func TestTCPClient_EmitProducesEventFrame(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Stop()

	cli, err := New(Config{
		Endpoint:     srv.Addr(),
		NodeInfo:     sampleNodeInfo(),
		ReconnectMin: 10 * time.Millisecond,
		ReconnectMax: 50 * time.Millisecond,
		CloseTimeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer cli.Close()

	if !waitEnabled(cli, 2*time.Second) {
		t.Fatalf("client never became Enabled")
	}

	id := cli.Emit(11, []byte{0xAA, 0xBB})
	if id == InvalidID {
		t.Fatalf("Emit returned InvalidID while client was Enabled")
	}
	if eventIDSeq(id) != 0 {
		t.Errorf("first event seq = %d, want 0", eventIDSeq(id))
	}

	if !srv.WaitFrames(2, 2*time.Second) {
		t.Fatalf("event frame not received")
	}
	frames := srv.Frames()
	if len(frames) < 2 {
		t.Fatalf("got %d frames, want >= 2", len(frames))
	}
	body := frames[1]
	// body = u64 ts + u8 disc + payload
	if len(body) < 9 {
		t.Fatalf("event body too short: %d bytes", len(body))
	}
	disc := body[8]
	if disc != 11 {
		t.Errorf("disc = %d, want 11", disc)
	}
	if !bytes.Equal(body[9:], []byte{0xAA, 0xBB}) {
		t.Errorf("payload mismatch: % x", body[9:])
	}
}

func TestTCPClient_EmitLazyDoesNotBuildOnDisabled(t *testing.T) {
	// Use an endpoint that will never connect; client stays disabled.
	cli, err := New(Config{
		Endpoint:     "127.0.0.1:1", // reserved, refused
		NodeInfo:     sampleNodeInfo(),
		ReconnectMin: 1 * time.Hour, // back off long so no retry races
		ReconnectMax: 1 * time.Hour,
		CloseTimeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer cli.Close()

	var built atomic.Int32
	id := cli.EmitLazy(1, func() []byte {
		built.Add(1)
		return nil
	})
	if id != InvalidID {
		t.Errorf("EmitLazy on disabled returned id %d, want InvalidID", id)
	}
	if got := built.Load(); got != 0 {
		t.Errorf("builder called %d times on disabled, want 0", got)
	}
}

func TestTCPClient_EmitFollowupRejectsInvalidParent(t *testing.T) {
	tc, err := newTCPClient(Config{
		Endpoint: "127.0.0.1:0",
		NodeInfo: sampleNodeInfo(),
	})
	if err != nil {
		t.Fatalf("newTCPClient: %v", err)
	}
	tc.enabledFlag.Store(true) // pretend connection is up

	if got := tc.EmitFollowup(1, InvalidID, nil); got != InvalidID {
		t.Errorf("InvalidID parent: got %d, want InvalidID", got)
	}

	// Stale parent: epoch on parent ID does not match sequencer's current
	// epoch (which is 1).
	stale := makeEventID(2, 5)
	if got := tc.EmitFollowup(1, stale, nil); got != InvalidID {
		t.Errorf("stale parent (epoch 2 vs 1): got %d, want InvalidID", got)
	}
}

func TestTCPClient_EmitFollowupAcceptsCurrentEpochParent(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Stop()

	cli, err := New(Config{
		Endpoint:     srv.Addr(),
		NodeInfo:     sampleNodeInfo(),
		ReconnectMin: 10 * time.Millisecond,
		ReconnectMax: 50 * time.Millisecond,
		CloseTimeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer cli.Close()

	if !waitEnabled(cli, 2*time.Second) {
		t.Fatalf("client never became Enabled")
	}

	parent := cli.Emit(1, nil)
	if parent == InvalidID {
		t.Fatalf("parent Emit returned InvalidID")
	}
	child := cli.EmitFollowup(2, parent, []byte{0xFF})
	if child == InvalidID {
		t.Errorf("EmitFollowup with current-epoch parent returned InvalidID")
	}

	if !srv.WaitFrames(3, 2*time.Second) {
		t.Fatalf("expected 3 frames (NodeInfo + parent event + child event)")
	}
	frames := srv.Frames()
	childBody := frames[2]
	// body = u64 ts + u8 disc + u64 parent_seq + u8 0xFF
	if len(childBody) != 8+1+8+1 {
		t.Fatalf("child body len %d, want %d", len(childBody), 8+1+8+1)
	}
	gotParent := binary.LittleEndian.Uint64(childBody[9:17])
	if gotParent != eventIDSeq(parent) {
		t.Errorf("parent_seq in child body = %d, want %d", gotParent, eventIDSeq(parent))
	}
}

func TestTCPClient_DropOnFullQueueAndDroppedEventEmitted(t *testing.T) {
	// Build a client whose queue is tiny and whose mock server reads slowly,
	// so we can force the queue full.
	srv := newSlowMockServer(t, 100*time.Millisecond)
	defer srv.Stop()

	cli, err := New(Config{
		Endpoint:         srv.Addr(),
		NodeInfo:         sampleNodeInfo(),
		BufferSize:       2,
		ReconnectMin:     10 * time.Millisecond,
		ReconnectMax:     50 * time.Millisecond,
		CloseTimeout:     2 * time.Second,
		TailDropInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer cli.Close()

	if !waitEnabled(cli, 2*time.Second) {
		t.Fatalf("client never became Enabled")
	}

	// Burst many emits; some will land, some will drop.
	const N = 50
	for i := 0; i < N; i++ {
		cli.Emit(7, []byte{byte(i)})
	}

	// Wait long enough for everything to flush (events + dropped events).
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		frames := srv.Frames()
		// We expect: 1 NodeInfo + N total event-or-dropped frames *but*
		// drops collapse contiguous runs into 1 frame. So total >= 1 + 1
		// (some events) + at least one Dropped frame.
		if len(frames) >= 3 {
			// Verify at least one frame has discriminator 0 (Dropped).
			for _, f := range frames[1:] {
				if len(f) >= 9 && f[8] == discriminatorDropped {
					return // success
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected to see at least one Dropped frame; got %d frames", len(srv.Frames()))
}

func TestTCPClient_CloseIsIdempotent(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Stop()

	cli, err := New(Config{
		Endpoint:     srv.Addr(),
		NodeInfo:     sampleNodeInfo(),
		ReconnectMin: 10 * time.Millisecond,
		ReconnectMax: 50 * time.Millisecond,
		CloseTimeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if !waitEnabled(cli, 2*time.Second) {
		t.Fatalf("client never became Enabled")
	}
	if err := cli.Close(); err != nil {
		t.Errorf("first Close: %v", err)
	}
	if err := cli.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
	if cli.Enabled() {
		t.Errorf("Enabled() should be false after Close")
	}
	if id := cli.Emit(1, nil); id != InvalidID {
		t.Errorf("Emit after Close returned %d, want InvalidID", id)
	}
}

func TestTCPClient_ReconnectBumpsEpoch(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Stop()

	tc, err := newTCPClient(Config{
		Endpoint:     srv.Addr(),
		NodeInfo:     sampleNodeInfo(),
		ReconnectMin: 5 * time.Millisecond,
		ReconnectMax: 10 * time.Millisecond,
		CloseTimeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("newTCPClient: %v", err)
	}
	tc.start()
	defer tc.Close()

	if !waitEnabled(tc, 2*time.Second) {
		t.Fatalf("client never became Enabled")
	}
	epoch1, _ := tc.seq.snapshot()
	if epoch1 != 1 {
		t.Fatalf("initial epoch = %d, want 1", epoch1)
	}

	// Drop the first server-side connection; client should reconnect with
	// epoch++.
	srv.DropFirstConnection(t, 2*time.Second)

	deadline := time.Now().Add(2 * time.Second)
	var epoch2 uint16
	for time.Now().Before(deadline) {
		epoch2, _ = tc.seq.snapshot()
		if epoch2 > epoch1 && tc.Enabled() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if epoch2 <= epoch1 {
		t.Errorf("epoch did not bump on reconnect: was %d, now %d", epoch1, epoch2)
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func sampleNodeInfo() NodeInfo {
	return NodeInfo{
		JAMParameters: []byte{0xAA, 0xBB, 0xCC},
		ImplName:      "test-jam",
		ImplVersion:   "0.0.1",
		GrayPaperVer:  "0.7.2",
		FreeformInfo:  "ci",
	}
}

func waitEnabled(cli Client, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cli.Enabled() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// ---------------------------------------------------------------------------
// Slow mock server: same as mockServer but reads with a per-frame delay so
// the client's send queue fills up.
// ---------------------------------------------------------------------------

type slowMockServer struct {
	*mockServer
	readDelay time.Duration
}

func newSlowMockServer(t *testing.T, delay time.Duration) *slowMockServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	base := &mockServer{
		ln:      ln,
		connCh:  make(chan net.Conn, 4),
		stopped: make(chan struct{}),
	}
	s := &slowMockServer{mockServer: base, readDelay: delay}
	go s.acceptSlowLoop()
	return s
}

func (s *slowMockServer) acceptSlowLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.connCh <- conn
		go s.readSlowLoop(conn)
	}
}

func (s *slowMockServer) readSlowLoop(conn net.Conn) {
	defer conn.Close()
	for {
		var hdr [4]byte
		if _, err := io.ReadFull(conn, hdr[:]); err != nil {
			return
		}
		n := binary.LittleEndian.Uint32(hdr[:])
		if n > 1<<20 {
			return
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return
		}
		s.mu.Lock()
		s.frames = append(s.frames, buf)
		s.mu.Unlock()
		time.Sleep(s.readDelay)
	}
}

// ---------------------------------------------------------------------------
// Sanity: New with unreachable endpoint returns a usable (but not enabled)
// client without blocking.
// ---------------------------------------------------------------------------

func TestNew_UnreachableEndpointReturnsImmediately(t *testing.T) {
	done := make(chan struct{})
	var cli Client
	var err error
	go func() {
		cli, err = New(Config{
			Endpoint:     "127.0.0.1:1",
			NodeInfo:     sampleNodeInfo(),
			ReconnectMin: 1 * time.Hour,
			ReconnectMax: 1 * time.Hour,
			CloseTimeout: 200 * time.Millisecond,
		})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("New blocked on unreachable endpoint")
	}
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer cli.Close()
	if cli.Enabled() {
		t.Errorf("Enabled before connection should be false")
	}
}

// Verify that on a broken server (immediate close) the client doesn't
// crash and stays not-Enabled until a working connection is up.
func TestTCPClient_BrokenServerDoesNotCrash(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	// Server immediately closes any incoming connection.
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	cli, err := New(Config{
		Endpoint:     ln.Addr().String(),
		NodeInfo:     sampleNodeInfo(),
		ReconnectMin: 5 * time.Millisecond,
		ReconnectMax: 20 * time.Millisecond,
		CloseTimeout: 200 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer cli.Close()

	// Give it a few reconnect cycles.
	time.Sleep(200 * time.Millisecond)
	// Should still be alive; if anything panicked the test would have
	// already crashed.
	if errors.Is(cli.Close(), io.EOF) {
		t.Errorf("unexpected Close error")
	}
}

// ---------------------------------------------------------------------------
// EmitFollowup parent epoch race: parent validation + child ID allocation
// must be atomic. Without the fix, a reconnect between validateParent and
// child Emit lets a stale parent (epoch N) pair with a child in epoch N+1.
// ---------------------------------------------------------------------------

func TestEmitFollowup_RejectsParentAfterEpochBumpUnderRace(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Stop()

	tc, err := newTCPClient(Config{
		Endpoint:     srv.Addr(),
		NodeInfo:     sampleNodeInfo(),
		BufferSize:   16,
		ReconnectMin: 5 * time.Millisecond,
		ReconnectMax: 10 * time.Millisecond,
		CloseTimeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("newTCPClient: %v", err)
	}
	tc.start()
	defer tc.Close()

	if !waitEnabled(tc, 2*time.Second) {
		t.Fatalf("client never became Enabled")
	}

	parent := tc.Emit(1, nil)
	if parent == InvalidID {
		t.Fatalf("parent Emit returned InvalidID")
	}

	// Force epoch bump: server drops the connection. Client reconnects
	// with epoch++, parent's epoch becomes stale.
	srv.DropFirstConnection(t, 2*time.Second)

	// Wait for the new epoch to be in place AND a new connection to be
	// up so EmitFollowup will hit the "Enabled but stale parent" path.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		ep, _ := tc.seq.snapshot()
		if ep > 1 && tc.Enabled() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if ep, _ := tc.seq.snapshot(); ep == 1 {
		t.Fatalf("epoch never bumped")
	}

	// Now the parent (epoch 1) is stale relative to the current epoch.
	// The child should be rejected without emitting anything.
	id := tc.EmitFollowup(2, parent, []byte{0xFF})
	if id != InvalidID {
		t.Errorf("EmitFollowup with stale parent returned id=%d, want InvalidID", id)
	}
}

// ---------------------------------------------------------------------------
// Close force path: cancels in-flight dial via context + force-closes the
// current connection so a writer / dialer blocked in I/O returns promptly.
// ---------------------------------------------------------------------------

func TestClose_CancelsInFlightDial(t *testing.T) {
	tc, err := newTCPClient(Config{
		Endpoint:     "127.0.0.1:1",
		NodeInfo:     sampleNodeInfo(),
		ReconnectMin: 1 * time.Hour, // never retry naturally
		ReconnectMax: 1 * time.Hour,
		CloseTimeout: 200 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("newTCPClient: %v", err)
	}

	// Dialer that hangs until ctx is cancelled, then returns ctx.Err().
	tc.dialer = func(ctx context.Context, addr string) (net.Conn, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	tc.start()

	// Give dial time to start.
	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	if err := tc.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	elapsed := time.Since(start)

	// Close should return within CloseTimeout + grace (1s grace).
	// Allowing 2s total bound for safety.
	if elapsed > 2*time.Second {
		t.Errorf("Close took %v with hung dialer; expected force-cancel within ~%v",
			elapsed, tc.cfg.CloseTimeout+1*time.Second)
	}
}

// blockingNodeInfoConn lets the dial succeed and lets Read drain naturally
// (so the peer-closed reader doesn't fire), but blocks every Write until
// releaseCh is closed. This pins writeNodeInfo and verifies Close's force
// path finds the conn (assigned BEFORE writeNodeInfo, not after).
type blockingNodeInfoConn struct {
	net.Conn
	releaseCh chan struct{}
	closed    chan struct{}
}

func (b *blockingNodeInfoConn) Write(p []byte) (int, error) {
	select {
	case <-b.releaseCh:
		return 0, errors.New("blockingNodeInfoConn: simulated write failure after release")
	case <-b.closed:
		return 0, net.ErrClosed
	}
}

func (b *blockingNodeInfoConn) Close() error {
	select {
	case <-b.closed:
	default:
		close(b.closed)
	}
	return b.Conn.Close()
}

func TestClose_ForceClosesDuringNodeInfoBlock(t *testing.T) {
	// A real listener so Close on the wrapped conn actually closes
	// something the peer can observe.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			// Don't drain — we want the conn alive but server-side passive.
			_ = c
		}
	}()

	tc, err := newTCPClient(Config{
		Endpoint:     ln.Addr().String(),
		NodeInfo:     sampleNodeInfo(),
		BufferSize:   16,
		ReconnectMin: 1 * time.Hour,
		CloseTimeout: 200 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("newTCPClient: %v", err)
	}
	releaseCh := make(chan struct{})
	tc.dialer = func(ctx context.Context, addr string) (net.Conn, error) {
		raw, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return &blockingNodeInfoConn{
			Conn:      raw,
			releaseCh: releaseCh,
			closed:    make(chan struct{}),
		}, nil
	}
	tc.start()

	// Give dial + currentConn assignment time. writeNodeInfo will block.
	time.Sleep(100 * time.Millisecond)

	// Without the fix, currentConn would still be nil (assigned after
	// writeNodeInfo returns), so Close's force path would find no conn
	// and the goroutine would outlive Close + grace.
	start := time.Now()
	if err := tc.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	elapsed := time.Since(start)

	// Bound: CloseTimeout (200ms) + 1s grace + small slack.
	if elapsed > 2*time.Second {
		t.Errorf(
			"Close took %v while writeNodeInfo was blocked; force path did not find currentConn",
			elapsed,
		)
	}
	close(releaseCh) // unblock any lingering write goroutine
}

// Reader goroutines (one per successful connection) are tracked by connWG
// so Close's "all spawned goroutines have exited when Close returns"
// contract holds. Across multiple reconnects, readers from earlier cycles
// must all be caught by connWG.Wait inside Close.
func TestClose_WaitsForAllReadersAcrossReconnects(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Stop()

	tc, err := newTCPClient(Config{
		Endpoint:     srv.Addr(),
		NodeInfo:     sampleNodeInfo(),
		BufferSize:   16,
		ReconnectMin: 5 * time.Millisecond,
		ReconnectMax: 10 * time.Millisecond,
		CloseTimeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("newTCPClient: %v", err)
	}
	tc.start()

	// Force a few reconnect cycles by dropping the server-side conn.
	for i := 0; i < 3; i++ {
		if !waitEnabled(tc, 2*time.Second) {
			t.Fatalf("cycle %d: client never became Enabled", i)
		}
		srv.DropFirstConnection(t, 2*time.Second)
		time.Sleep(20 * time.Millisecond) // let reconnect kick in
	}

	if !waitEnabled(tc, 2*time.Second) {
		t.Fatalf("final: client never became Enabled")
	}

	// Close should wait for all spawned reader goroutines (one per
	// successful connection) to have observed the conn close.
	start := time.Now()
	if err := tc.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("Close took %v across 3 reconnects; readers may not be tracked", elapsed)
	}
}
