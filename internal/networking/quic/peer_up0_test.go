package quic

import (
	"context"
	"crypto/ed25519"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPeerRunUPStreamRejectsOlderDuplicate(t *testing.T) {
	peer := &Peer{upStreams: newUPStreamRegistry()}
	peerKey := ed25519.PublicKey{1}

	block := make(chan struct{})
	var handlerCalls atomic.Int32
	handler := func(context.Context, *Stream, ed25519.PublicKey) error {
		handlerCalls.Add(1)
		<-block
		return nil
	}

	high := &mockUPStream{id: 8}
	go peer.runUPStream(context.Background(), upStreamKind, &Stream{Stream: high}, peerKey, handler)
	require.Eventually(t, func() bool {
		return handlerCalls.Load() == 1
	}, time.Second, 10*time.Millisecond)

	older := &mockUPStream{id: 6}
	peer.runUPStream(context.Background(), upStreamKind, &Stream{Stream: older}, peerKey, handler)

	require.Equal(t, int32(1), handlerCalls.Load())
	require.True(t, older.canceled)
	require.False(t, high.canceled)

	close(block)
}

func TestPeerRunUPStreamSupersedesActiveStream(t *testing.T) {
	peer := &Peer{upStreams: newUPStreamRegistry()}
	peerKey := ed25519.PublicKey{1}

	block := make(chan struct{})
	var handlerCalls atomic.Int32
	handler := func(context.Context, *Stream, ed25519.PublicKey) error {
		handlerCalls.Add(1)
		<-block
		return nil
	}

	low := &mockUPStream{id: 4}
	ctx := context.Background()

	go peer.runUPStream(ctx, upStreamKind, &Stream{Stream: low}, peerKey, handler)
	require.Eventually(t, func() bool {
		return handlerCalls.Load() == 1
	}, time.Second, 10*time.Millisecond)

	high := &mockUPStream{id: 8}
	go peer.runUPStream(ctx, upStreamKind, &Stream{Stream: high}, peerKey, handler)

	require.Eventually(t, func() bool {
		return low.canceled
	}, time.Second, 10*time.Millisecond)
	require.Eventually(t, func() bool {
		return handlerCalls.Load() == 2
	}, time.Second, 10*time.Millisecond)
	require.False(t, high.canceled)

	close(block)
}
