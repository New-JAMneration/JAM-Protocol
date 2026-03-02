package ce

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	quicgo "github.com/quic-go/quic-go"
)

// --- mockStream implements quic.MessageStream for tests (Read, Write, Close, WriteMessage).
type mockStream struct {
	r      *bytes.Buffer
	w      *bytes.Buffer
	closed bool
}

func newMockStream(initialData []byte) *mockStream {
	return &mockStream{
		r: bytes.NewBuffer(initialData),
		w: new(bytes.Buffer),
	}
}

func (m *mockStream) Read(p []byte) (int, error) {
	return m.r.Read(p)
}

func (m *mockStream) Write(p []byte) (int, error) {
	return m.w.Write(p)
}

func (m *mockStream) Close() error {
	m.closed = true
	return nil
}

func (m *mockStream) WriteMessage(payload []byte) error {
	return quic.WriteMessageFrame(m.w, payload)
}

// ReadMessage reads one JAMNP-framed message (4-byte LE length + payload) from the stream.
func (m *mockStream) ReadMessage() ([]byte, error) {
	var hdr [4]byte
	if n, err := m.r.Read(hdr[:]); err != nil || n != 4 {
		if err != nil {
			return nil, err
		}
		return nil, io.ErrUnexpectedEOF
	}
	msgLen := binary.LittleEndian.Uint32(hdr[:])
	if msgLen > quic.MaxMsgSize {
		return nil, fmt.Errorf("message too large: %d", msgLen)
	}
	buf := make([]byte, msgLen)
	if _, err := io.ReadFull(m.r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (m *mockStream) Context() context.Context {
	return context.Background()
}

func (m *mockStream) SetDeadline(t time.Time) error      { return nil }
func (m *mockStream) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockStream) SetWriteDeadline(t time.Time) error { return nil }

// Satisfy quicgo.Stream for tests that type-assert.
func (m *mockStream) CancelRead(quicgo.StreamErrorCode)  {}
func (m *mockStream) CancelWrite(quicgo.StreamErrorCode) {}
func (m *mockStream) StreamID() quicgo.StreamID          { return 0 }
