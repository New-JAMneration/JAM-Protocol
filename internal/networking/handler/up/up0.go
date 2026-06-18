package up

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"io"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// framedStream is a persistent UP 0 stream with JAMNP message framing.
type framedStream interface {
	io.Closer
	ReadMessage() ([]byte, error)
	WriteMessage(payload []byte) error
}

// UP0Handler runs UP 0 sessions and fans out outbound announcements to active peers.
type UP0Handler struct {
	Blocks         func() []types.Block
	Finalized      func() (types.HeaderHash, error)
	OnAnnouncement func(Announcement, ed25519.PublicKey) error

	mu       sync.Mutex
	sessions map[string]*UP0Session
}

// UP0Session is one persistent UP 0 stream to a single peer.
type UP0Session struct {
	stream            framedStream
	peerKey           ed25519.PublicKey
	blocksProvider    func() []types.Block
	finalizedProvider func() (types.HeaderHash, error)
	onAnnouncement    func(Announcement, ed25519.PublicKey) error

	mu              sync.Mutex
	cv              ChainView
	ourFinal        BlockRef
	announcedByUs   map[types.HeaderHash]struct{}
	announcedByPeer map[types.HeaderHash]struct{}
}

// Handle runs the UP 0 session: parallel handshake, then an announcement read loop.
func (h *UP0Handler) Handle(ctx context.Context, stream *quic.Stream, peerKey ed25519.PublicKey) error {
	if h == nil {
		return fmt.Errorf("nil UP0Handler")
	}
	session, err := h.newSession(stream, peerKey)
	if err != nil {
		return err
	}

	h.registerSession(peerKey, session)
	defer h.unregisterSession(peerKey)

	if err := session.exchangeHandshake(ctx); err != nil {
		return err
	}
	return session.readLoop(ctx)
}

func (h *UP0Handler) newSession(stream framedStream, peerKey ed25519.PublicKey) (*UP0Session, error) {
	if h.Blocks == nil || h.Finalized == nil {
		return nil, fmt.Errorf("UP0Handler requires Blocks and Finalized providers")
	}
	blocks := h.Blocks()
	finalized, err := h.Finalized()
	if err != nil {
		return nil, fmt.Errorf("finalized provider: %w", err)
	}
	cv, finalRef, err := ViewAtFinalized(blocks, finalized)
	if err != nil {
		return nil, err
	}
	return &UP0Session{
		stream:            stream,
		peerKey:           peerKey,
		blocksProvider:    h.Blocks,
		finalizedProvider: h.Finalized,
		onAnnouncement:    h.OnAnnouncement,
		cv:                cv,
		ourFinal:          finalRef,
		announcedByUs:     make(map[types.HeaderHash]struct{}),
		announcedByPeer:   make(map[types.HeaderHash]struct{}),
	}, nil
}

func (h *UP0Handler) registerSession(peerKey ed25519.PublicKey, session *UP0Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessions == nil {
		h.sessions = make(map[string]*UP0Session)
	}
	h.sessions[string(peerKey)] = session
}

func (h *UP0Handler) unregisterSession(peerKey ed25519.PublicKey) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, string(peerKey))
}

// AnnounceBlock fans out a block announcement to all active UP 0 sessions.
func (h *UP0Handler) AnnounceBlock(header types.Header) {
	if h == nil {
		return
	}
	h.mu.Lock()
	sessions := make([]*UP0Session, 0, len(h.sessions))
	for _, s := range h.sessions {
		sessions = append(sessions, s)
	}
	h.mu.Unlock()

	for _, s := range sessions {
		_ = s.AnnounceBlock(header)
	}
}

func readMessage(ctx context.Context, stream framedStream) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		data, err := stream.ReadMessage()
		ch <- result{data: data, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		return res.data, res.err
	}
}

func (s *UP0Session) exchangeHandshake(ctx context.Context) error {
	type handshakeResult struct {
		hs  Handshake
		err error
	}
	readCh := make(chan handshakeResult, 1)
	go func() {
		data, err := readMessage(ctx, s.stream)
		if err != nil {
			readCh <- handshakeResult{err: err}
			return
		}
		hs, err := DecodeHandshake(data)
		if err != nil {
			readCh <- handshakeResult{err: fmt.Errorf("decode handshake: %w", err)}
			return
		}
		readCh <- handshakeResult{hs: hs}
	}()

	blocks := s.blocksProvider()
	finalized, err := s.finalizedProvider()
	if err != nil {
		return err
	}
	localHS, err := WriteHandshake(s.stream, blocks, finalized)
	if err != nil {
		return fmt.Errorf("send handshake: %w", err)
	}
	s.trackHandshake(localHS, false)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-readCh:
		if res.err != nil {
			return fmt.Errorf("read handshake: %w", res.err)
		}
		s.trackHandshake(res.hs, true)
		return nil
	}
}

func (s *UP0Session) readLoop(ctx context.Context) error {
	for {
		data, err := readMessage(ctx, s.stream)
		if err != nil {
			return err
		}
		ann, err := DecodeAnnouncement(data)
		if err != nil {
			return fmt.Errorf("decode announcement: %w", err)
		}
		if err := s.handleAnnouncement(ann); err != nil {
			return err
		}
	}
}

func (s *UP0Session) handleAnnouncement(ann Announcement) error {
	blockHash, err := hash.ComputeBlockHeaderHash(ann.Header)
	if err != nil {
		return fmt.Errorf("hash announced header: %w", err)
	}

	s.mu.Lock()
	cv := s.cv
	s.mu.Unlock()

	if !cv.IsDescendantOf(blockHash, ann.Final.Hash) {
		return nil
	}

	s.trackAnnounced(true, blockHash)
	if s.onAnnouncement != nil {
		return s.onAnnouncement(ann, s.peerKey)
	}
	return nil
}

// AnnounceBlock sends an announcement when skip rules do not apply.
func (s *UP0Session) AnnounceBlock(header types.Header) error {
	if err := s.refreshChainView(); err != nil {
		return err
	}

	blockHash, err := hash.ComputeBlockHeaderHash(header)
	if err != nil {
		return fmt.Errorf("hash header: %w", err)
	}

	s.mu.Lock()
	cv := s.cv
	finalized := s.ourFinal.Hash
	finalRef := s.ourFinal
	announcedByUs := s.announcedByUs
	announcedByPeer := s.announcedByPeer
	s.mu.Unlock()

	if ShouldSkipAnnouncement(blockHash, finalized, cv, announcedByUs, announcedByPeer) {
		return nil
	}

	ann := Announcement{Header: header, Final: finalRef}
	payload, err := EncodeAnnouncement(ann)
	if err != nil {
		return err
	}
	if err := s.stream.WriteMessage(payload); err != nil {
		return err
	}
	s.trackAnnounced(false, blockHash)
	return nil
}

func (s *UP0Session) refreshChainView() error {
	blocks := s.blocksProvider()
	finalized, err := s.finalizedProvider()
	if err != nil {
		return err
	}
	cv, finalRef, err := ViewAtFinalized(blocks, finalized)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.cv = cv
	s.ourFinal = finalRef
	s.mu.Unlock()
	return nil
}

func (s *UP0Session) trackHandshake(h Handshake, fromPeer bool) {
	for _, ref := range HandshakeRefs(h) {
		s.trackAnnounced(fromPeer, ref.Hash)
	}
}

func (s *UP0Session) trackAnnounced(fromPeer bool, h types.HeaderHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fromPeer {
		s.announcedByPeer[h] = struct{}{}
	} else {
		s.announcedByUs[h] = struct{}{}
	}
}
