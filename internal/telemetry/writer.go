package telemetry

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// JIP-3 reserves discriminator 0 for the Dropped event. JamTART
// advances its implicit counter by the dropped count (not 1) on a
// Dropped event, keeping wire order aligned with event IDs across a
// partial drop.
const discriminatorDropped uint8 = 0

// writeLoop is the single writer goroutine for one TCP connection.
// It serialises c.queue → wire in ascending event-ID order, flushing
// drop ranges from c.drops first whenever the next envelope's seq is
// past the writer's expectedWireID.
//
// Returns:
//   - nil on a clean Close (closeCh fired, queue + drops drained).
//   - the I/O error on any write failure → connectLoop reconnects.
//   - a synthetic error after a panic recover → client marked degraded.
func (c *tcpClient) writeLoop(conn net.Conn, peerClosed <-chan struct{}) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("telemetry: writer panic: %v", r)
			c.degradedFlag.Store(true)
			retErr = fmt.Errorf("telemetry writer degraded after panic: %v", r)
		}
	}()

	var expectedWireID uint64
	var pending *envelope

	ticker := time.NewTicker(c.cfg.TailDropInterval)
	defer ticker.Stop()

	closing := false

	for {
		// 1. Flush leading drop ranges that line up with expectedWireID.
		if _, err := c.flushReadyDrops(io.Writer(conn), &expectedWireID); err != nil {
			return err
		}

		// 2. Write the pending envelope if it's now ready.
		if pending != nil && eventIDSeq(pending.id) == expectedWireID {
			if err := c.writeEvent(conn, pending); err != nil {
				return err
			}
			expectedWireID++
			pending = nil
			continue
		}

		// 3. Pending but expectedWireID still behind: a producer must be
		// mid-Lock recording a drop. Re-acquire the lock; once we get
		// it, the drop range (if any) is observable.
		if pending != nil {
			c.seq.Lock()
			head, ok := c.drops.peekFirst()
			c.seq.Unlock()
			if ok && eventIDSeq(head.firstID) == expectedWireID {
				continue
			}
			// Pending.seq > expected with no matching drop = invariant
			// broken. Bail; the conn drops and reconnect resets state.
			return fmt.Errorf(
				"telemetry: wire alignment lost: pending seq=%d expected=%d, no drop range",
				eventIDSeq(pending.id), expectedWireID,
			)
		}

		// 4. Queue + drops idle. In closing mode, drain non-blockingly;
		// otherwise wait for input.
		if closing {
			select {
			case env := <-c.queue:
				pending = &env
				continue
			default:
				c.seq.Lock()
				_, hasDrop := c.drops.peekFirst()
				c.seq.Unlock()
				if !hasDrop {
					return nil
				}
				continue // tail drops left; loop flushes them
			}
		}

		select {
		case <-c.closeCh:
			closing = true
		case <-peerClosed:
			return fmt.Errorf("telemetry: peer closed connection")
		case env := <-c.queue:
			pending = &env
		case <-ticker.C:
			// next iteration's flushReadyDrops handles tail ranges
		}
	}
}

// flushReadyDrops writes head drop ranges that line up with
// *expectedWireID as Dropped events, advancing *expectedWireID by each
// range's count. Returns the number flushed and the first I/O error.
//
// peekFirst + popFirst happen under one lock so a producer recording a
// contiguous next-id drop can't grow the in-flight head range. Without
// this claim-before-write, the wire would write count=N (sampled
// pre-merge) but pop the count=N+1 expanded range, dropping one event
// from JamTART's counter.
func (c *tcpClient) flushReadyDrops(w io.Writer, expectedWireID *uint64) (int, error) {
	flushed := 0
	for {
		c.seq.Lock()
		head, ok := c.drops.peekFirst()
		if !ok || eventIDSeq(head.firstID) != *expectedWireID {
			c.seq.Unlock()
			return flushed, nil
		}
		c.drops.popFirst() // claim before write
		c.seq.Unlock()
		if err := c.writeDropped(w, head); err != nil {
			return flushed, err
		}
		*expectedWireID += head.count
		flushed++
	}
}

// ---------------------------------------------------------------------------
// Wire encoders / framing
// ---------------------------------------------------------------------------

// writeFrame writes a length-prefixed frame: u32 LE length + payload.
// Both writes go through writeAll so a partial-write io.Writer can't
// silently truncate the frame.
func writeFrame(w io.Writer, payload []byte) error {
	if uint64(len(payload)) > uint64(^uint32(0)) {
		return fmt.Errorf("telemetry: frame too large: %d bytes", len(payload))
	}
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(payload)))
	if err := writeAll(w, hdr[:]); err != nil {
		return fmt.Errorf("telemetry: write length prefix: %w", err)
	}
	if err := writeAll(w, payload); err != nil {
		return fmt.Errorf("telemetry: write payload: %w", err)
	}
	return nil
}

// writeAll loops Write until all of b is written. The io.Writer contract
// permits (n < len(p), err == nil) for wrappers, so a single Write would
// silently truncate. Returns io.ErrShortWrite on (0, nil) to avoid
// looping forever.
func writeAll(w io.Writer, b []byte) error {
	for len(b) > 0 {
		n, err := w.Write(b)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		b = b[n:]
	}
	return nil
}

// writeNodeInfo encodes and frames the connection-initial NodeInfo. Per
// JIP-3 it has no Timestamp / Discriminator header.
func (c *tcpClient) writeNodeInfo(w io.Writer) error {
	payload, err := c.cfg.NodeInfo.Encode()
	if err != nil {
		return fmt.Errorf("telemetry: encode node info: %w", err)
	}
	return writeFrame(w, payload)
}

// writeEvent writes one event frame: u64 ts + u8 disc + payload. Lazy
// envelopes have their builder run here so dropped events never pay
// encoding cost.
func (c *tcpClient) writeEvent(w io.Writer, env *envelope) error {
	if env.builder != nil {
		env.payload = env.builder()
		env.builder = nil
	}
	body := make([]byte, 0, 8+1+len(env.payload))
	body = append(body, EncodeU64(env.ts)...)
	body = append(body, EncodeU8(env.disc)...)
	body = append(body, env.payload...)
	return writeFrame(w, body)
}

// writeDropped writes one Dropped event for r:
//
//	header ts = r.firstTS, disc = 0
//	payload   = u64 r.lastTS + u64 r.count
//
// JamTART derives implicit_id = r.firstID and advances its counter by
// r.count, keeping the next event's ID aligned with the sender.
func (c *tcpClient) writeDropped(w io.Writer, r dropRange) error {
	body := make([]byte, 0, 8+1+8+8)
	body = append(body, EncodeU64(r.firstTS)...)
	body = append(body, EncodeU8(discriminatorDropped)...)
	body = append(body, EncodeU64(r.lastTS)...)
	body = append(body, EncodeU64(r.count)...)
	return writeFrame(w, body)
}
