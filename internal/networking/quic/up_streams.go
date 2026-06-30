package quic

import (
	"sync"

	quicgo "github.com/quic-go/quic-go"
)

// upStreamKind is JAMNP-S UP 0 (block announcement).
const upStreamKind byte = 0

type upStreamRegistry struct {
	mu     sync.Mutex
	active map[string]map[byte]trackedUPStream
}

type trackedUPStream struct {
	id     quicgo.StreamID
	stream quicgo.Stream
}

func newUPStreamRegistry() *upStreamRegistry {
	return &upStreamRegistry{
		active: make(map[string]map[byte]trackedUPStream),
	}
}

// admit records a UP stream and returns whether the caller should run the handler.
// Per JAMNP-S, only the UP stream with the greatest QUIC stream ID is kept.
func (r *upStreamRegistry) admit(peerKey string, kind byte, id quicgo.StreamID, stream quicgo.Stream) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	byKind, ok := r.active[peerKey]
	if !ok {
		byKind = make(map[byte]trackedUPStream)
		r.active[peerKey] = byKind
	}

	current, exists := byKind[kind]
	if !exists {
		byKind[kind] = trackedUPStream{id: id, stream: stream}
		return true
	}
	if id == current.id {
		return true
	}
	if id < current.id {
		cancelStream(stream)
		return false
	}

	cancelStream(current.stream)
	byKind[kind] = trackedUPStream{id: id, stream: stream}
	return true
}

func (r *upStreamRegistry) closeKind(peerKey string, kind byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	byKind, ok := r.active[peerKey]
	if !ok {
		return
	}
	if tracked, ok := byKind[kind]; ok {
		cancelStream(tracked.stream)
		delete(byKind, kind)
	}
	if len(byKind) == 0 {
		delete(r.active, peerKey)
	}
}

func (r *upStreamRegistry) closeExcept(eligible map[string]struct{}, kind byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for peerKey, byKind := range r.active {
		if _, ok := eligible[peerKey]; ok {
			continue
		}
		if tracked, ok := byKind[kind]; ok {
			cancelStream(tracked.stream)
			delete(byKind, kind)
		}
		if len(byKind) == 0 {
			delete(r.active, peerKey)
		}
	}
}

func (r *upStreamRegistry) hasKind(peerKey string, kind byte) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	byKind, ok := r.active[peerKey]
	if !ok {
		return false
	}
	_, ok = byKind[kind]
	return ok
}

func (r *upStreamRegistry) release(peerKey string, kind byte, id quicgo.StreamID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	byKind, ok := r.active[peerKey]
	if !ok {
		return
	}
	current, exists := byKind[kind]
	if exists && current.id == id {
		delete(byKind, kind)
	}
	if len(byKind) == 0 {
		delete(r.active, peerKey)
	}
}

func cancelStream(stream quicgo.Stream) {
	if stream == nil {
		return
	}
	stream.CancelRead(0)
	stream.CancelWrite(0)
}
