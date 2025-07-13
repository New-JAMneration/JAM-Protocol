package fuzz

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type FuzzService interface {
	Handshake(PeerInfo) (PeerInfo, error)
	ImportBlock(types.Block) (StateRoot, error)
	SetState(types.Header, State) (StateRoot, error)
	GetState(types.HeaderHash) (State, error)
}

type FuzzServiceStub struct {
	// TODO: fill in whatever dependency you need
}

func (s *FuzzServiceStub) Handshake(peerInfo PeerInfo) (PeerInfo, error) {
	log.Println("Received Handshake:")
	log.Printf("  Name: %s\n", peerInfo.Name)
	log.Printf("  App Version: %v\n", peerInfo.AppVersion)
	log.Printf("  JAM Version: %v\n", peerInfo.JamVersion)

	var response PeerInfo

	if err := response.FromConfig(); err != nil {
		return PeerInfo{}, err
	}

	return response, nil
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
