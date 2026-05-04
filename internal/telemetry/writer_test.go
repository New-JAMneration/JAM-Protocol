package telemetry

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
)

func TestWriteFrame_LengthPrefixAndPayload(t *testing.T) {
	t.Run("EmptyPayload", func(t *testing.T) {
		var buf bytes.Buffer
		if err := writeFrame(&buf, nil); err != nil {
			t.Fatalf("writeFrame: %v", err)
		}
		got := buf.Bytes()
		if len(got) != 4 {
			t.Fatalf("got %d bytes, want 4", len(got))
		}
		if l := binary.LittleEndian.Uint32(got); l != 0 {
			t.Errorf("length prefix = %d, want 0", l)
		}
	})

	t.Run("SmallPayload", func(t *testing.T) {
		payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}
		var buf bytes.Buffer
		if err := writeFrame(&buf, payload); err != nil {
			t.Fatalf("writeFrame: %v", err)
		}
		got := buf.Bytes()
		if len(got) != 8 {
			t.Fatalf("got %d bytes, want 8", len(got))
		}
		if l := binary.LittleEndian.Uint32(got[:4]); l != 4 {
			t.Errorf("length prefix = %d, want 4", l)
		}
		if !bytes.Equal(got[4:], payload) {
			t.Errorf("payload mismatch: got % x", got[4:])
		}
	})
}

func TestTCPClient_writeEvent_Layout(t *testing.T) {
	c := &tcpClient{cfg: Config{}.withDefaults()}
	var buf bytes.Buffer
	env := envelope{
		id:      makeEventID(1, 7),
		ts:      0x1122334455667788,
		disc:    42,
		payload: []byte{0xAA, 0xBB},
	}
	if err := c.writeEvent(&buf, &env); err != nil {
		t.Fatalf("writeEvent: %v", err)
	}
	got := buf.Bytes()

	// frame: 4-byte length prefix + body. body = u64 LE ts + u8 disc + payload.
	wantBody := []byte{
		0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, // ts LE
		42,         // disc
		0xAA, 0xBB, // payload
	}
	if l := binary.LittleEndian.Uint32(got[:4]); int(l) != len(wantBody) {
		t.Errorf("length prefix = %d, want %d", l, len(wantBody))
	}
	if !bytes.Equal(got[4:], wantBody) {
		t.Errorf("body mismatch:\n got %x\nwant %x", got[4:], wantBody)
	}
}

func TestTCPClient_writeEvent_RunsLazyBuilder(t *testing.T) {
	c := &tcpClient{cfg: Config{}.withDefaults()}
	called := false
	env := envelope{
		id:   makeEventID(1, 0),
		ts:   1,
		disc: 5,
		builder: func() []byte {
			called = true
			return []byte{0xCC}
		},
	}
	var buf bytes.Buffer
	if err := c.writeEvent(&buf, &env); err != nil {
		t.Fatalf("writeEvent: %v", err)
	}
	if !called {
		t.Fatalf("builder should have been called")
	}
	if env.builder != nil {
		t.Errorf("builder should be cleared after run")
	}
	if !bytes.Equal(env.payload, []byte{0xCC}) {
		t.Errorf("payload not populated: % x", env.payload)
	}
}

func TestTCPClient_writeDropped_Layout(t *testing.T) {
	c := &tcpClient{cfg: Config{}.withDefaults()}
	var buf bytes.Buffer
	r := dropRange{
		firstID: makeEventID(1, 100),
		count:   7,
		firstTS: 0x0000000000000001,
		lastTS:  0x0000000000000099,
	}
	if err := c.writeDropped(&buf, r); err != nil {
		t.Fatalf("writeDropped: %v", err)
	}
	got := buf.Bytes()

	// body = u64 LE firstTS + u8 disc=0 + u64 LE lastTS + u64 LE count
	wantBody := []byte{
		0x01, 0, 0, 0, 0, 0, 0, 0, // firstTS = 1
		0,                         // disc = 0 (Dropped)
		0x99, 0, 0, 0, 0, 0, 0, 0, // lastTS = 0x99
		0x07, 0, 0, 0, 0, 0, 0, 0, // count = 7
	}
	if l := binary.LittleEndian.Uint32(got[:4]); int(l) != len(wantBody) {
		t.Errorf("length prefix = %d, want %d", l, len(wantBody))
	}
	if !bytes.Equal(got[4:], wantBody) {
		t.Errorf("body mismatch:\n got %x\nwant %x", got[4:], wantBody)
	}
}

// ---------------------------------------------------------------------------
// flushReadyDrops claims (peek+pop) the head range under one lock so a
// producer cannot grow it between when we sample its count and when we
// remove it. Without atomic claim, a racing contiguous drop merges into
// the in-flight head, the wire writes count=N (pre-claim copy), but
// popFirst removes the count=N+1 expanded range — losing one event from
// JamTART's implicit counter.
// ---------------------------------------------------------------------------

// hookWriter runs a callback before each Write so tests can observe state
// at exactly the moment the wire write happens.
type hookWriter struct {
	onWrite func()
	buf     bytes.Buffer
}

func (h *hookWriter) Write(b []byte) (int, error) {
	if h.onWrite != nil {
		h.onWrite()
	}
	return h.buf.Write(b)
}

func TestFlushReadyDrops_ClaimsRangeBeforeWrite(t *testing.T) {
	c := &tcpClient{cfg: Config{}.withDefaults(), seq: newSequencer(), drops: &dropState{}}

	// Pre-populate one drop range at seq=0 count=1.
	c.seq.Lock()
	c.drops.record(makeEventID(1, 0), 1000)
	c.seq.Unlock()

	expected := uint64(0)
	var observedPopped int32 // 1 = popped before write, 0 = still present
	w := &hookWriter{
		onWrite: func() {
			c.seq.Lock()
			_, hasHead := c.drops.peekFirst()
			c.seq.Unlock()
			if !hasHead {
				atomic.StoreInt32(&observedPopped, 1)
			}
		},
	}

	n, err := c.flushReadyDrops(w, &expected)
	if err != nil {
		t.Fatalf("flushReadyDrops: %v", err)
	}
	if n != 1 {
		t.Errorf("flushed = %d, want 1", n)
	}
	if expected != 1 {
		t.Errorf("expectedWireID = %d, want 1", expected)
	}
	if atomic.LoadInt32(&observedPopped) != 1 {
		t.Errorf("range was not popped before writeDropped — producers could still merge into it")
	}
}

// Verifies a producer recording a contiguous drop while a Dropped frame is
// being written cannot get its event silently absorbed. Without the fix,
// the racing drop would merge the head from {firstID:0,count:1} to
// {firstID:0,count:2}; we'd write count=1 (pre-claim copy) and pop the
// count=2 range, losing seq=1 from JamTART's counter. With the fix, the
// racing drop becomes a fresh range and gets flushed on the next loop iter
// as a SECOND Dropped(count=1) frame. flushed == 2 proves both events
// reached the wire.
func TestFlushReadyDrops_RacingContiguousDropProducesTwoFrames(t *testing.T) {
	c := &tcpClient{cfg: Config{}.withDefaults(), seq: newSequencer(), drops: &dropState{}}

	// Pre-populate the head range with seq=0 (count=1).
	c.seq.Lock()
	c.drops.record(makeEventID(1, 0), 1000)
	c.seq.Unlock()

	expected := uint64(0)
	var racingDropDone sync.WaitGroup
	racingDropDone.Add(1)
	var raceFired int32

	w := &hookWriter{}
	w.onWrite = func() {
		// Once (on the first hook call), simulate a producer recording
		// a contiguous drop for seq=1. The fix guarantees this cannot
		// merge into the in-flight head (already popped); it must start
		// a fresh range.
		if !atomic.CompareAndSwapInt32(&raceFired, 0, 1) {
			return
		}
		go func() {
			defer racingDropDone.Done()
			c.seq.Lock()
			c.drops.record(makeEventID(1, 1), 2000)
			c.seq.Unlock()
		}()
		racingDropDone.Wait()
	}

	flushed, err := c.flushReadyDrops(w, &expected)
	if err != nil {
		t.Fatalf("flushReadyDrops: %v", err)
	}
	if flushed != 2 {
		t.Errorf("flushed = %d, want 2 (one per dropped event, proving the racing drop was not absorbed into the in-flight frame)", flushed)
	}
	if expected != 2 {
		t.Errorf("expectedWireID = %d, want 2", expected)
	}
}

// ---------------------------------------------------------------------------
// writeAll loops on partial writes. The io.Writer contract permits
// (n < len(p), err == nil) for some implementations (notably wrappers);
// without the loop, writeFrame would silently truncate frames.
// ---------------------------------------------------------------------------

// oneByteAtATimeWriter writes exactly one byte per Write call. Standard
// io.Writer contract permits this for any wrapper.
type oneByteAtATimeWriter struct {
	buf bytes.Buffer
}

func (o *oneByteAtATimeWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return o.buf.Write(p[:1])
}

func TestWriteAll_LoopsOnPartialWrite(t *testing.T) {
	w := &oneByteAtATimeWriter{}
	want := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x42}
	if err := writeAll(w, want); err != nil {
		t.Fatalf("writeAll: %v", err)
	}
	if !bytes.Equal(w.buf.Bytes(), want) {
		t.Errorf("got %x, want %x", w.buf.Bytes(), want)
	}
}

// zeroWriter returns (0, nil) — a contract violation but possible from
// some buggy wrappers. writeAll must detect and return ErrShortWrite
// rather than spin forever.
type zeroWriter struct{}

func (zeroWriter) Write(p []byte) (int, error) { return 0, nil }

func TestWriteAll_DetectsZeroWriteWithNilError(t *testing.T) {
	err := writeAll(zeroWriter{}, []byte{1, 2, 3})
	if !errors.Is(err, io.ErrShortWrite) {
		t.Errorf("got err=%v, want io.ErrShortWrite", err)
	}
}

func TestWriteFrame_PartialWriteSurvives(t *testing.T) {
	w := &oneByteAtATimeWriter{}
	payload := []byte{0xAA, 0xBB, 0xCC}
	if err := writeFrame(w, payload); err != nil {
		t.Fatalf("writeFrame: %v", err)
	}
	got := w.buf.Bytes()
	if len(got) != 4+len(payload) {
		t.Fatalf("got %d bytes, want %d", len(got), 4+len(payload))
	}
	// Verify length prefix.
	if got[0] != 3 || got[1] != 0 || got[2] != 0 || got[3] != 0 {
		t.Errorf("length prefix = % x, want 03 00 00 00", got[:4])
	}
	if !bytes.Equal(got[4:], payload) {
		t.Errorf("payload = % x, want % x", got[4:], payload)
	}
}
