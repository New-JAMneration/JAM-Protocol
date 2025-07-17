package fuzz

import (
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
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
	Blockchain blockchain.Blockchain
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
	// Do some validation on the block here
	if s.Blockchain.GetLatestBlock().Header.Parent != block.Header.Parent {
		return StateRoot{}, fmt.Errorf("parent mismatch: got %x, want %x", s.Blockchain.GetLatestBlock().Header.Parent, block.Header.Parent)
	}

	if s.Blockchain.GetLatestBlock().Header.ParentStateRoot != block.Header.ParentStateRoot {
		return StateRoot{}, fmt.Errorf("state_root mismatch: got %x, want %x", s.Blockchain.GetLatestBlock().Header.ParentStateRoot, block.Header.ParentStateRoot)
	}

	// Store the block in the blockchain (into redis)S
	blockHash := block.Header.Parent
	err := s.Blockchain.StoreBlockByHash(blockHash, &block)
	if err != nil {
		return StateRoot{}, err
	}

	// Run the STF and get the state root
	err = stf.RunSTF()
	if err != nil {
		return StateRoot{}, err
	}

	latestBlock := s.Blockchain.GetLatestBlock()

	return StateRoot(latestBlock.Header.ParentStateRoot), nil
}

func (s *FuzzServiceStub) SetState(header types.Header, state State) (StateRoot, error) {
	// TODO
	return StateRoot{}, ErrNotImpl
}

func (s *FuzzServiceStub) GetState(hash types.HeaderHash) (State, error) {
	// TODO
	return State{}, ErrNotImpl
}
