package quic

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	quicgo "github.com/quic-go/quic-go"
)

// Can be changed to other values if needed.
const MaxMsgSize = 256 * 1024 * 1024

// Stream encapsulates a quic.Stream to provide a higher-level abstraction
// for reading from and writing to QUIC streams.
type Stream struct {
	Stream quicgo.Stream
}

func (s *Stream) Read(p []byte) (int, error) {
	return s.Stream.Read(p)
}

func (s *Stream) ReadFull(p []byte) error {
	off := 0
	for off < len(p) {
		n, err := s.Read(p[off:])
		if n > 0 {
			off += n
		}
		if err != nil {
			// if the stream is closed, return nil
			if err == io.EOF && off == len(p) {
				return nil
			}
			if err == io.EOF {
				return io.ErrUnexpectedEOF
			}
			return err
		}
	}
	return nil
}

func (s *Stream) Write(p []byte) (int, error) {
	return s.Stream.Write(p)
}

func (s *Stream) writeFull(p []byte) error {
	for len(p) > 0 {
		n, err := s.Write(p)
		if n > 0 {
			p = p[n:]
		}
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
	}
	return nil
}

func (s *Stream) Close() error {
	return s.Stream.Close()
}

func (s *Stream) SetReadDeadline(t time.Time) error {
	return s.Stream.SetReadDeadline(t)
}

func Copy(dest quicgo.Stream, src quicgo.Stream) error {
	_, err := io.Copy(dest, src)
	return err
}

// After opening a stream, the stream initiator must send a single byte identifying the stream kind.
func (s *Stream) WriteStreamKind(kind byte) error {
	return s.writeFull([]byte{kind})
}

func (s *Stream) ReadStreamKind() (byte, error) {
	var b [1]byte
	if err := s.ReadFull(b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}

// Message format: [len][payload...]
func (s *Stream) WriteMessage(payload []byte) error {
	if uint64(len(payload)) > MaxMsgSize {
		return fmt.Errorf("payload too large: %d > %d", len(payload), MaxMsgSize)
	}

	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(payload)))
	if err := s.writeFull(hdr[:]); err != nil {
		return err
	}
	return s.writeFull(payload)
}

// Message format: [len][payload...]
func (s *Stream) ReadMessage() ([]byte, error) {
	var hdr [4]byte
	if err := s.ReadFull(hdr[:]); err != nil {
		return nil, err
	}
	n := binary.LittleEndian.Uint32(hdr[:])

	if n > MaxMsgSize {
		return nil, fmt.Errorf("message too large: %d > %d", n, MaxMsgSize)
	}

	buf := make([]byte, n)
	if err := s.ReadFull(buf); err != nil {
		return nil, err
	}
	return buf, nil
}
