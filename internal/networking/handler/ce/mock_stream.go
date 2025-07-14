package ce

import (
	"bytes"
	"context"
	"time"

	quicgo "github.com/quic-go/quic-go"
)

// --- mockStream implements a minimal in-memory quic.Stream.
type mockStream struct {
	r      *bytes.Buffer // used for reading incoming request bytes
	w      *bytes.Buffer // used for capturing written response bytes
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

func (m *mockStream) Context() context.Context {
	return context.Background()
}

func (m *mockStream) SetDeadline(t time.Time) error      { return nil }
func (m *mockStream) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockStream) SetWriteDeadline(t time.Time) error { return nil }

// Added to satisfy quic.Stream interface.
func (m *mockStream) CancelRead(quicgo.StreamErrorCode)  {}
func (m *mockStream) CancelWrite(quicgo.StreamErrorCode) {}
func (m *mockStream) StreamID() quicgo.StreamID          { return 0 }
