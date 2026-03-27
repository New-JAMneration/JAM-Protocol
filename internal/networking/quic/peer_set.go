package quic

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"sync"
)

type PeerSet struct {
	mu             sync.RWMutex
	byEd25519Key   map[string]*Peer
	byAddr         map[string]*Peer
	byValidatorIdx map[uint16]*Peer
}

func NewPeerSet() *PeerSet {
	return &PeerSet{
		byEd25519Key:   make(map[string]*Peer),
		byAddr:         make(map[string]*Peer),
		byValidatorIdx: make(map[uint16]*Peer),
	}
}

func peerKeyString(p *Peer) string {
	return hex.EncodeToString(p.Ed25519Key)
}

func (ps *PeerSet) Add(p *Peer, addr string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(p.Ed25519Key) != 0 && len(p.Ed25519Key) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid ed25519 key length: %d", len(p.Ed25519Key))
	}
	if len(p.Ed25519Key) == ed25519.PublicKeySize {
		ps.byEd25519Key[peerKeyString(p)] = p
	}
	if addr != "" {
		ps.byAddr[addr] = p
	}
	if p.ValidatorIndex != nil {
		ps.byValidatorIdx[*p.ValidatorIndex] = p
	}
	return nil
}

func (ps *PeerSet) Remove(p *Peer, addr string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(p.Ed25519Key) > 0 {
		delete(ps.byEd25519Key, peerKeyString(p))
	}
	if addr != "" {
		delete(ps.byAddr, addr)
	}
	if p.ValidatorIndex != nil {
		delete(ps.byValidatorIdx, *p.ValidatorIndex)
	}
}

func (ps *PeerSet) GetByKey(key string) (*Peer, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.byEd25519Key[key]
	return p, ok
}

func (ps *PeerSet) GetByAddr(addr string) (*Peer, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.byAddr[addr]
	return p, ok
}

func (ps *PeerSet) GetByValidatorIndex(idx uint16) (*Peer, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.byValidatorIdx[idx]
	return p, ok
}

func (ps *PeerSet) All() []*Peer {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	out := make([]*Peer, 0, len(ps.byAddr))
	for _, p := range ps.byAddr {
		out = append(out, p)
	}
	return out
}
