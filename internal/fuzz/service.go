package fuzz

import (
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type FuzzService interface {
	Handshake(PeerInfo) (PeerInfo, error)
	ImportBlock(types.Block) (StateRoot, error)
	SetState(types.Header, State) (StateRoot, error)
	GetState(types.HeaderHash) (State, error)
}

var ErrNotImpl = errors.New("not implemented")

type FuzzServiceStub struct {
	// TODO: fill in whatever dependency you need
}

func (s *FuzzServiceStub) Handshake(peerInfo PeerInfo) (PeerInfo, error) {
	// TODO
	return PeerInfo{}, ErrNotImpl
}

func (s *FuzzServiceStub) ImportBlock(block types.Block) (StateRoot, error) {
	// TODO
	return StateRoot{}, ErrNotImpl
}

func (s *FuzzServiceStub) SetState(header types.Header, state State) (StateRoot, error) {
	// TODO
	return StateRoot{}, ErrNotImpl
}

func (s *FuzzServiceStub) GetState(hash types.HeaderHash) (State, error) {
	// TODO
	return State{}, ErrNotImpl
}
