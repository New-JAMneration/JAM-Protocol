package fuzz

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

type FuzzService interface {
	Handshake(PeerInfo) (PeerInfo, error)
	ImportBlock(types.Block) (types.StateRoot, error)
	SetState(types.Header, types.StateKeyVals, types.Ancestry) (types.StateRoot, error)
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
	// Build context for logging
	headerHash, _ := hash.ComputeBlockHeaderHash(block.Header)
	hashStr := hex.EncodeToString(headerHash[:])
	slot := uint32(block.Header.Slot) % uint32(types.EpochLength)
	epoch := slot / uint32(types.EpochLength)
	ctx := logger.FormatContext(hashStr, slot, epoch, "ImportBlock")
	logger.Debugf("%s Processing...", ctx)

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
			logger.Debugf("%s parent mismatch, trying to restore block and state", ctx)
			// Try the restore block and state
			err := storeInstance.RestoreBlockAndState(block.Header.Parent)
			if err != nil {
				return types.StateRoot{}, fmt.Errorf("failed to restore block and state after parent mismatch: %w", err)
			}
		}
	} else {
		// Initialize the ring verifier cache
		store.ClearVerifierCache()
	}

	// Get the latest state root
	latestState := storeInstance.GetPriorStates().GetState()
	serializedState, _ := m.StateEncoder(latestState)
	priorUnmatchedKeyVals := storeInstance.GetPriorStateUnmatchedKeyVals()
	serializedState = append(priorUnmatchedKeyVals, serializedState...)
	latestStateRoot := m.MerklizationSerializedState(serializedState)

	if latestStateRoot != block.Header.ParentStateRoot {
		return types.StateRoot{}, fmt.Errorf("state_root mismatch: got 0x%x, want 0x%x", latestStateRoot, block.Header.ParentStateRoot)
	}

	storeInstance.AddBlock(block)
	logger.Infof("%s Block added: 0x%x", ctx, headerHash)

	// Run the STF and get the state root
	isProtocolError, err := stf.RunSTF()
	if err != nil {
		if !isProtocolError {
			// Runtime error: unexpected bug, should terminate the program
			// Note: We return the error here, caller (server) should decide to close connection
			// If this is a standalone node, caller should call logger.Fatal()
			logger.Errorf("%s STF runtime error (unexpected bug): %v", ctx, err)
			return types.StateRoot{}, fmt.Errorf("STF runtime error: %w", err)
		}
		// Protocol error: block is invalid, but node should continue
		logger.Errorf("%s [PROTOCOL] block invalid: %v", ctx, err)
		return block.Header.ParentStateRoot, err
	}

	latestState = storeInstance.GetPosteriorStates().GetState()
	serializedState, err = m.StateEncoder(latestState)
	if err != nil {
		logger.Errorf("%s state encoder error: %v", ctx, err)
		return types.StateRoot{}, err
	}
	postUnmatchedKeyVals := storeInstance.GetPostStateUnmatchedKeyVals()
	serializedState = append(postUnmatchedKeyVals, serializedState...)
	latestStateRoot = m.MerklizationSerializedState(serializedState)

	// Append current block header to ancestry on successful import
	currentAncestryItem := types.Ancestry{
		{
			Slot:       block.Header.Slot,
			HeaderHash: headerHash,
		},
	}
	storeInstance.AppendAncestry(currentAncestryItem)
	logger.Debugf("%s Added block to ancestry: slot=%d, hash=0x%x", ctx, block.Header.Slot, headerHash[:8])

	// Commit the state and persist the state to Redis
	storeInstance.StateCommit()

	return latestStateRoot, nil
}

func (s *FuzzServiceStub) SetState(header types.Header, stateKeyVals types.StateKeyVals, ancestry types.Ancestry) (types.StateRoot, error) {
	// Build context for logging
	headerHash, _ := hash.ComputeBlockHeaderHash(header)
	hashStr := hex.EncodeToString(headerHash[:])
	slot := uint32(header.Slot) % uint32(types.EpochLength)
	epoch := slot / uint32(types.EpochLength)
	ctx := logger.FormatContext(hashStr, slot, epoch, "SetState")
	logger.Debugf("%s Processing...", ctx)

	// Reset State and Blocks
	store.ClearVerifierCache()
	store.ResetInstance()

	// Set State
	storeInstance := store.GetInstance()

	// Append ancestry if provided
	if len(ancestry) > 0 {
		storeInstance.AppendAncestry(ancestry)
		logger.Debugf("%s Appended %d ancestry items", ctx, len(ancestry))
	}

	state, unmatchedKeyVals, err := m.StateKeyValsToState(stateKeyVals)
	if err != nil {
		logger.Errorf("%s failed to convert state keyvals: %v", ctx, err)
		return types.StateRoot{}, err
	}

	storeInstance.GetPosteriorStates().SetState(state)
	// store storage key-val into global variable
	store.GetInstance().SetPostStateUnmatchedKeyVals(unmatchedKeyVals)
	serializedState, _ := m.StateEncoder(state)

	stateRoot := m.MerklizationSerializedState(append(unmatchedKeyVals, serializedState...))

	block := types.Block{
		Header: header,
	}
	storeInstance.AddBlock(block)

	// Add current header to ancestry
	currentAncestryItem := types.Ancestry{
		{
			Slot:       header.Slot,
			HeaderHash: headerHash,
		},
	}
	storeInstance.AppendAncestry(currentAncestryItem)

	// Commit the state
	storeInstance.StateCommit()

	logger.Infof("%s Init completed, header hash: 0x%x, state root: 0x%x", ctx, headerHash, stateRoot)
	return stateRoot, nil
}

func (s *FuzzServiceStub) GetState(headerHash types.HeaderHash) (types.StateKeyVals, error) {
	// Build context for logging - try to get slot/epoch from block if available
	hashStr := hex.EncodeToString(headerHash[:])
	slot := uint32(0)
	epoch := uint32(0)
	storeInstance := store.GetInstance()
	// Try to get block to extract slot/epoch
	block, err := storeInstance.GetBlockByHash(headerHash)
	if err == nil {
		slot = uint32(block.Header.Slot) % uint32(types.EpochLength)
		epoch = slot / uint32(types.EpochLength)
	}

	ctx := logger.FormatContext(hashStr, slot, epoch, "GetState")

	state, err := storeInstance.GetStateByBlockHash(headerHash)
	if err != nil {
		logger.Errorf("%s failed to get state: %v", ctx, err)
		return nil, err
	}

	logger.Infof("%s Retrieved %d key-value pairs", ctx, len(state))
	return state, nil
}
