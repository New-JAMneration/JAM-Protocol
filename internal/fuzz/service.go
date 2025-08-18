package fuzz

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

type FuzzService interface {
	Handshake(PeerInfo) (PeerInfo, error)
	ImportBlock(types.Block) (types.StateRoot, error)
	SetState(types.Header, types.StateKeyVals) (types.StateRoot, error)
	GetState(types.HeaderHash) (types.StateKeyVals, error)
}

type FuzzServiceStub struct {
	// TODO: fill in whatever dependency you need
	Blockchain blockchain.Blockchain
}

func (s *FuzzServiceStub) Handshake(peerInfo PeerInfo) (PeerInfo, error) {
	var response PeerInfo

	if err := response.FromConfig(); err != nil {
		return PeerInfo{}, err
	}

	return response, nil
}

func (s *FuzzServiceStub) ImportBlock(block types.Block) (types.StateRoot, error) {
	// Do some validation on the block here
	if s.Blockchain.GetLatestBlock().Header.Parent != block.Header.Parent {
		return types.StateRoot{}, fmt.Errorf("parent mismatch: got %x, want %x", s.Blockchain.GetLatestBlock().Header.Parent, block.Header.Parent)
	}

	if s.Blockchain.GetLatestBlock().Header.ParentStateRoot != block.Header.ParentStateRoot {
		return types.StateRoot{}, fmt.Errorf("state_root mismatch: got %x, want %x", s.Blockchain.GetLatestBlock().Header.ParentStateRoot, block.Header.ParentStateRoot)
	}

	// Store the block in the blockchain (into redis)S
	blockHash := block.Header.Parent
	err := s.Blockchain.StoreBlockByHash(blockHash, &block)
	if err != nil {
		return types.StateRoot{}, err
	}

	// Run the STF and get the state root
	err = stf.RunSTF()
	if err != nil {
		return types.StateRoot{}, err
	}

	latestBlock := s.Blockchain.GetLatestBlock()

	return latestBlock.Header.ParentStateRoot, nil
}

func (s *FuzzServiceStub) SetState(header types.Header, stateKeyVals types.StateKeyVals) (types.StateRoot, error) {
	storeInstance := store.GetInstance()

	state, err := m.StateKeyValsToState(stateKeyVals)
	if err != nil {
		return types.StateRoot{}, err
	}

	storeInstance.SetPosteriorStatesInstance(store.NewPosteriorStates(store.WithState(state)))

	stateRoot := m.MerklizationState(state)

	return stateRoot, nil
}

func (s *FuzzServiceStub) GetState(hash types.HeaderHash) (types.StateKeyVals, error) {
	storeInstance := store.GetInstance()

	if hash != storeInstance.GetProcessingBlockPointer().GetBlock().Header.Parent {
		return types.StateKeyVals{}, fmt.Errorf("hash mismatch: got %x, want %x", hash, storeInstance.GetProcessingBlockPointer().Header.Hash)
	}

	state := storeInstance.GetPosteriorStates().GetState()

	encodedState, err := m.StateEncoder(state)
	if err != nil {
		return encodedState, err
	}

	return types.StateKeyVals{}, nil
}
