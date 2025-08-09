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

func (s *FuzzServiceStub) SetState(header types.Header, state types.StateKeyVals) (types.StateRoot, error) {
	storeInstance := store.GetInstance()

	for _, value := range state {
		decoder := types.NewDecoder()
		switch value.Key {
		case m.StateWrapper{StateIndex: 1}.StateKeyConstruct():
			val := types.AuthPools{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetAlpha(val)
		case m.StateWrapper{StateIndex: 2}.StateKeyConstruct():
			val := types.AuthQueues{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetVarphi(val)
		case m.StateWrapper{StateIndex: 3}.StateKeyConstruct():
			val := types.Beta{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetBeta(val)
		case m.StateWrapper{StateIndex: 4}.StateKeyConstruct():
			val := types.Gamma{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetGamma(val)
		case m.StateWrapper{StateIndex: 5}.StateKeyConstruct():
			val := types.DisputesRecords{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetPsi(val)
		case m.StateWrapper{StateIndex: 6}.StateKeyConstruct():
			val := types.EntropyBuffer{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetEta(val)
		case m.StateWrapper{StateIndex: 7}.StateKeyConstruct():
			val := types.ValidatorsData{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetIota(val)
		case m.StateWrapper{StateIndex: 8}.StateKeyConstruct():
			val := types.ValidatorsData{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetKappa(val)
		case m.StateWrapper{StateIndex: 9}.StateKeyConstruct():
			val := types.ValidatorsData{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetLambda(val)
		case m.StateWrapper{StateIndex: 10}.StateKeyConstruct():
			val := types.AvailabilityAssignments{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetRho(val)
		case m.StateWrapper{StateIndex: 11}.StateKeyConstruct():
			val := make([]types.TimeSlot, 1)
			val[0].Decode(decoder)
			storeInstance.GetPriorStates().SetTau(val[0])
		case m.StateWrapper{StateIndex: 12}.StateKeyConstruct():
			val := types.Privileges{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetChi(val)
		case m.StateWrapper{StateIndex: 13}.StateKeyConstruct():
			val := types.Statistics{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetPi(val)
		case m.StateWrapper{StateIndex: 14}.StateKeyConstruct():
			val := types.ReadyQueue{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetTheta(val)
		case m.StateWrapper{StateIndex: 15}.StateKeyConstruct():
			val := types.AccumulatedQueue{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetXi(val)
		case m.StateWrapper{StateIndex: 16}.StateKeyConstruct():
			val := types.AccumulatedServiceOutput{}
			decoder.Decode(value.Value, &val)
			storeInstance.GetPriorStates().SetLastAccOut(val)
		}

		// TODO Delta 1 - 4

	}
	// TODO
	return types.StateRoot{}, ErrNotImpl
}

func (s *FuzzServiceStub) GetState(hash types.HeaderHash) (types.StateKeyVals, error) {
	storeInstance := store.GetInstance()
	state := storeInstance.GetPosteriorStates().GetState()
	encodedState, err := m.StateEncoder(state)
	if err != nil {
		return encodedState, err
	}
	return types.StateKeyVals{}, nil
}
