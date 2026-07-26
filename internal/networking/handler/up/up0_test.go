package up

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	quicgo "github.com/quic-go/quic-go"
	"github.com/stretchr/testify/require"
)

func testHandler(t *testing.T, blocks []types.Block, finalized types.HeaderHash) *UP0Handler {
	t.Helper()
	return &UP0Handler{
		Blocks:    func() []types.Block { return blocks },
		Finalized: func() (types.HeaderHash, error) { return finalized, nil },
	}
}

func readFramedMessage(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	msgLen := binary.LittleEndian.Uint32(hdr[:])
	if msgLen > quic.MaxMsgSize {
		return nil, fmt.Errorf("message too large: %d", msgLen)
	}
	buf := make([]byte, msgLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

type testUP0Stream struct {
	reader io.Reader
	writer io.Writer
	record *bytes.Buffer
}

func newTestUP0Stream(initialRead []byte) *testUP0Stream {
	return &testUP0Stream{
		reader: bytes.NewBuffer(initialRead),
		record: new(bytes.Buffer),
	}
}

func newLinkedTestUP0Streams() (*testUP0Stream, *testUP0Stream) {
	aRead, aWrite := io.Pipe()
	bRead, bWrite := io.Pipe()
	return &testUP0Stream{reader: aRead, writer: bWrite},
		&testUP0Stream{reader: bRead, writer: aWrite}
}

func (s *testUP0Stream) Close() error {
	if c, ok := s.reader.(io.Closer); ok {
		_ = c.Close()
	}
	if c, ok := s.writer.(io.Closer); ok {
		_ = c.Close()
	}
	return nil
}

func (s *testUP0Stream) WriteMessage(payload []byte) error {
	if s.writer != nil {
		return quic.WriteMessageFrame(s.writer, payload)
	}
	return quic.WriteMessageFrame(s.record, payload)
}

func (s *testUP0Stream) ReadMessage() ([]byte, error) {
	return readFramedMessage(s.reader)
}

func (s *testUP0Stream) writtenPayloads() [][]byte {
	var out [][]byte
	r := bytes.NewReader(s.record.Bytes())
	for {
		payload, err := readFramedMessage(r)
		if err != nil {
			break
		}
		out = append(out, payload)
	}
	return out
}

type pipeQuicStream struct {
	inner *testUP0Stream
}

func asQuicStream(inner *testUP0Stream) *quic.Stream {
	return &quic.Stream{Stream: &pipeQuicStream{inner: inner}}
}

func (s *pipeQuicStream) Read(p []byte) (int, error)         { return s.inner.reader.Read(p) }
func (s *pipeQuicStream) Write(p []byte) (int, error)        { return s.inner.writer.Write(p) }
func (s *pipeQuicStream) Close() error                       { return s.inner.Close() }
func (s *pipeQuicStream) StreamID() quicgo.StreamID          { return 1 }
func (s *pipeQuicStream) CancelRead(quicgo.StreamErrorCode)  {}
func (s *pipeQuicStream) CancelWrite(quicgo.StreamErrorCode) {}
func (s *pipeQuicStream) SetReadDeadline(time.Time) error    { return nil }
func (s *pipeQuicStream) SetWriteDeadline(time.Time) error   { return nil }
func (s *pipeQuicStream) SetDeadline(time.Time) error        { return nil }
func (s *pipeQuicStream) Context() context.Context           { return context.Background() }

func TestHandshakeExchangeConcurrent(t *testing.T) {
	blocks, finalized := testChain(t)
	streamA, streamB := newLinkedTestUP0Streams()
	handler := testHandler(t, blocks, finalized)

	sessionA, err := handler.newSession(streamA, ed25519.PublicKey{1})
	require.NoError(t, err)
	sessionB, err := handler.newSession(streamB, ed25519.PublicKey{2})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	errCh := make(chan error, 2)
	go func() { errCh <- sessionA.exchangeHandshake(ctx) }()
	go func() { errCh <- sessionB.exchangeHandshake(ctx) }()
	require.NoError(t, <-errCh)
	require.NoError(t, <-errCh)
}

func TestAnnounceBlockSkipsPeerKnownBlock(t *testing.T) {
	blocks, finalized := testChain(t)
	stream := newTestUP0Stream(nil)
	session, err := testHandler(t, blocks, finalized).newSession(stream, ed25519.PublicKey{1})
	require.NoError(t, err)

	branchA := blocks[1].Header
	aHash := mustHash(t, branchA)
	session.trackHandshake(Handshake{
		Final:  BlockRef{Hash: finalized, Slot: 10},
		Leaves: []BlockRef{{Hash: aHash, Slot: branchA.Slot}},
	}, true)

	require.NoError(t, session.AnnounceBlock(branchA))
	require.Empty(t, stream.writtenPayloads())

	bHeader := blocks[2].Header
	require.NoError(t, session.AnnounceBlock(bHeader))
	payloads := stream.writtenPayloads()
	require.Len(t, payloads, 1)

	ann, err := DecodeAnnouncement(payloads[0])
	require.NoError(t, err)
	require.Equal(t, bHeader.Slot, ann.Header.Slot)
}

func TestAnnouncementInvokesCallback(t *testing.T) {
	blocks, finalized := testChain(t)
	stream := newTestUP0Stream(nil)

	var received Announcement
	var peerKey ed25519.PublicKey
	handler := testHandler(t, blocks, finalized)
	handler.OnAnnouncement = func(ann Announcement, pk ed25519.PublicKey) error {
		received = ann
		peerKey = pk
		return nil
	}

	wantKey := ed25519.PublicKey{9}
	session, err := handler.newSession(stream, wantKey)
	require.NoError(t, err)

	bHeader := blocks[2].Header
	require.NoError(t, session.handleAnnouncement(Announcement{
		Header: bHeader,
		Final:  BlockRef{Hash: finalized, Slot: 10},
	}))
	require.Equal(t, bHeader.Slot, received.Header.Slot)
	require.Equal(t, wantKey, peerKey)
}

func TestUP0HandlerHandleInvokesOnAnnouncement(t *testing.T) {
	blocks, finalized := testChain(t)
	localStream, remoteStream := newLinkedTestUP0Streams()

	var received Announcement
	var peerKey ed25519.PublicKey
	handler := testHandler(t, blocks, finalized)
	handler.OnAnnouncement = func(ann Announcement, pk ed25519.PublicKey) error {
		received = ann
		peerKey = pk
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.Handle(ctx, asQuicStream(localStream), ed25519.PublicKey{1})
	}()

	remoteSession, err := handler.newSession(remoteStream, ed25519.PublicKey{2})
	require.NoError(t, err)
	require.NoError(t, remoteSession.exchangeHandshake(ctx))

	bHeader := blocks[2].Header
	ann := Announcement{Header: bHeader, Final: BlockRef{Hash: finalized, Slot: blocks[0].Header.Slot}}
	payload, err := EncodeAnnouncement(ann)
	require.NoError(t, err)
	require.NoError(t, remoteStream.WriteMessage(payload))

	require.Eventually(t, func() bool {
		return received.Header.Slot == bHeader.Slot
	}, time.Second, 10*time.Millisecond)
	require.Equal(t, ed25519.PublicKey{1}, peerKey)

	cancel()
	<-errCh
}

func TestUP0SessionReadLoopInvokesCallback(t *testing.T) {
	blocks, finalized := testChain(t)
	bHeader := blocks[2].Header
	branchD := types.Header{Parent: mustHash(t, bHeader), Slot: 14}
	blocks = append(blocks, types.Block{Header: branchD})

	ann := Announcement{Header: branchD, Final: BlockRef{Hash: finalized, Slot: 10}}
	payload, err := EncodeAnnouncement(ann)
	require.NoError(t, err)

	var framed bytes.Buffer
	require.NoError(t, quic.WriteMessageFrame(&framed, payload))
	stream := newTestUP0Stream(framed.Bytes())

	var received Announcement
	handler := testHandler(t, blocks, finalized)
	handler.OnAnnouncement = func(ann Announcement, pk ed25519.PublicKey) error {
		received = ann
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	session, err := handler.newSession(stream, ed25519.PublicKey{2})
	require.NoError(t, err)

	errCh := make(chan error, 1)
	go func() { errCh <- session.readLoop(ctx) }()

	require.Eventually(t, func() bool {
		return received.Header.Slot == branchD.Slot
	}, time.Second, 10*time.Millisecond)

	cancel()
	<-errCh
}

func TestUP0HandlerHandleRejectsMalformedAnnouncement(t *testing.T) {
	blocks, finalized := testChain(t)
	localStream, remoteStream := newLinkedTestUP0Streams()
	handler := testHandler(t, blocks, finalized)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.Handle(ctx, asQuicStream(localStream), ed25519.PublicKey{1})
	}()

	remoteSession, err := handler.newSession(remoteStream, ed25519.PublicKey{2})
	require.NoError(t, err)
	require.NoError(t, remoteSession.exchangeHandshake(ctx))
	require.NoError(t, remoteStream.WriteMessage([]byte{0x01, 0x02, 0x03}))

	err = <-errCh
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode announcement")
}

func TestUP0HandlerDropsNonDescendantAnnouncement(t *testing.T) {
	blocks, finalized := testChain(t)
	stream := newTestUP0Stream(nil)

	called := false
	handler := testHandler(t, blocks, finalized)
	handler.OnAnnouncement = func(Announcement, ed25519.PublicKey) error {
		called = true
		return nil
	}

	session, err := handler.newSession(stream, ed25519.PublicKey{1})
	require.NoError(t, err)

	var unrelated types.HeaderHash
	unrelated[0] = 0x99
	require.NoError(t, session.handleAnnouncement(Announcement{
		Header: blocks[2].Header,
		Final:  BlockRef{Hash: unrelated, Slot: 99},
	}))
	require.False(t, called)
}

func TestUP0HandlerUnregisterSessionKeepsReplacement(t *testing.T) {
	handler := &UP0Handler{sessions: make(map[string]*UP0Session)}
	peerKey := ed25519.PublicKey{1}

	first := &UP0Session{peerKey: peerKey}
	second := &UP0Session{peerKey: peerKey}

	handler.registerSession(peerKey, first)
	handler.registerSession(peerKey, second)
	handler.unregisterSession(peerKey, first)

	handler.mu.Lock()
	active := handler.sessions[string(peerKey)]
	handler.mu.Unlock()
	require.Equal(t, second, active)
}
