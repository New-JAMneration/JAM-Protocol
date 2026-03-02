package ce

import (
	"bytes"
	"context"
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
