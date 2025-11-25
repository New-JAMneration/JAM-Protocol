package fuzz

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

type FuzzService interface {
	Handshake(PeerInfo) (PeerInfo, error)
	ImportBlock(types.Block) (types.StateRoot, error)
	SetState(types.Header, types.StateKeyVals) (types.StateRoot, error)
	GetState(types.HeaderHash) (types.StateKeyVals, error)
}

type FuzzServiceStub struct{}

func (s *FuzzServiceStub) Handshake(peerInfo PeerInfo) (PeerInfo, error) {
	var response PeerInfo

	if err := response.FromConfig(); err != nil {
		return PeerInfo{}, err
	}

	return response, nil
}

func (s *FuzzServiceStub) ImportBlock(block types.Block) (types.StateRoot, error) {
	// Get the latest block
	storeInstance := store.GetInstance()

	blocks := storeInstance.GetBlocks()
	if len(blocks) > 0 {
		latestBlock := storeInstance.GetLatestBlock()
		encoder := types.NewEncoder()
		encodedLatestHeader, err := encoder.Encode(&latestBlock.Header)
		if err != nil {
			return types.StateRoot{}, err
		}
		latestBlockHash := types.HeaderHash(hash.Blake2bHash(encodedLatestHeader))

		if latestBlockHash != block.Header.Parent {
			// Try the restore block and state
			err := storeInstance.RestoreBlockAndState(block.Header.Parent)
			if err != nil {
				return types.StateRoot{}, fmt.Errorf("failed to restore block and state after parent mismatch: %w", err)
			}
		}
	}

	// Get the latest state root
	latestState := storeInstance.GetPriorStates().GetState()
	serializedState, _ := m.StateEncoder(latestState)
	storageKeyVal := storeInstance.GetStorageKeyVals()
	serializedState = append(storageKeyVal, serializedState...)
	latestStateRoot := m.MerklizationSerializedState(serializedState)

	if latestStateRoot != block.Header.ParentStateRoot {
		return types.StateRoot{}, fmt.Errorf("state_root mismatch: got 0x%x, want 0x%x", latestStateRoot, block.Header.ParentStateRoot)
	}

	storeInstance.AddBlock(block)
	// Run the STF and get the state root
	isProtocolError, err := stf.RunSTF()
	if err != nil {
		if !isProtocolError {
			return types.StateRoot{}, fmt.Errorf("STF runtime error: %v", err)
		}
		return block.Header.ParentStateRoot, err
	}

	latestState = storeInstance.GetPosteriorStates().GetState()
	serializedState, err = m.StateEncoder(latestState)
	if err != nil {
		fmt.Printf("state encoder error: %v\n", err)
		return types.StateRoot{}, err
	}
	storageKeyVal = storeInstance.GetStorageKeyVals()
	serializedState = append(storageKeyVal, serializedState...)
	latestStateRoot = m.MerklizationSerializedState(serializedState)

	// Commit the state and persist the state to Redis
	storeInstance.StateCommit()

	return latestStateRoot, nil
}

func (s *FuzzServiceStub) SetState(header types.Header, stateKeyVals types.StateKeyVals) (types.StateRoot, error) {
	// Reset State and Blocks
	store.ResetInstance()

	// Set State
	storeInstance := store.GetInstance()

	state, storageKeyVal, err := m.StateKeyValsToState(stateKeyVals)
	if err != nil {
		return types.StateRoot{}, err
	}

	storeInstance.GetPosteriorStates().SetState(state)
	// store storage key-val into global variable
	store.GetInstance().SetStorageKeyVals(storageKeyVal)
	serializedState, _ := m.StateEncoder(state)

	stateRoot := m.MerklizationSerializedState(append(storageKeyVal, serializedState...))

	// Use the header to store the mapping
	block := types.Block{
		Header: header,
	}
	storeInstance.AddBlock(block)

	// Commit the state
	storeInstance.StateCommit()

	return stateRoot, nil
}

func (s *FuzzServiceStub) GetState(headerHash types.HeaderHash) (types.StateKeyVals, error) {
	storeInstance := store.GetInstance()
	return storeInstance.GetStateByBlockHash(headerHash)
}
