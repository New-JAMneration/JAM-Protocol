package quic

import (
	"context"
	"testing"
	"time"

	quicgo "github.com/quic-go/quic-go"
	"github.com/stretchr/testify/require"
)

type mockUPStream struct {
	id       quicgo.StreamID
	canceled bool
}

func (m *mockUPStream) StreamID() quicgo.StreamID { return m.id }
func (m *mockUPStream) Read([]byte) (int, error)  { return 0, nil }
func (m *mockUPStream) Write([]byte) (int, error) { return 0, nil }
func (m *mockUPStream) Close() error              { return nil }
func (m *mockUPStream) CancelRead(quicgo.StreamErrorCode) {
	m.canceled = true
}
func (m *mockUPStream) CancelWrite(quicgo.StreamErrorCode) {
	m.canceled = true
}
func (m *mockUPStream) SetReadDeadline(time.Time) error  { return nil }
func (m *mockUPStream) SetWriteDeadline(time.Time) error { return nil }
func (m *mockUPStream) SetDeadline(time.Time) error      { return nil }
func (m *mockUPStream) Context() context.Context         { return context.Background() }

func TestUPStreamRegistryKeepsGreatestStreamID(t *testing.T) {
	reg := newUPStreamRegistry()
	peer := "peer-a"

	low := &mockUPStream{id: 4}
	require.True(t, reg.admit(peer, upStreamKind, low.id, low))

	high := &mockUPStream{id: 8}
	require.True(t, reg.admit(peer, upStreamKind, high.id, high))

	older := &mockUPStream{id: 6}
	require.False(t, reg.admit(peer, upStreamKind, older.id, older))
	require.True(t, older.canceled)
	require.False(t, low.canceled)

	newer := &mockUPStream{id: 12}
	require.True(t, reg.admit(peer, upStreamKind, newer.id, newer))
	require.True(t, high.canceled)
}

func TestUPStreamRegistryReleaseOnlyCurrent(t *testing.T) {
	reg := newUPStreamRegistry()
	peer := "peer-a"
	stream := &mockUPStream{id: 4}
	require.True(t, reg.admit(peer, upStreamKind, stream.id, stream))

	reg.release(peer, upStreamKind, 2)
	_, exists := reg.active[peer][upStreamKind]
	require.True(t, exists)

	reg.release(peer, upStreamKind, stream.id)
	_, exists = reg.active[peer][upStreamKind]
	require.False(t, exists)
}
