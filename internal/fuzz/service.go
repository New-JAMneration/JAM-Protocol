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

// ImportBlock receives a block and determines if the block is valid or mutant(fork)
// in mutant case, fallback to target state by identifying parent block
func (s *FuzzServiceStub) ImportBlock(block types.Block) (types.StateRoot, error) {
	// Build context for logging
	headerHash, err := hash.ComputeBlockHeaderHash(block.Header)
	if err != nil {
		return types.StateRoot{}, fmt.Errorf("error computing header hash: %w", err)
	}
	hashStr := hex.EncodeToString(headerHash[:])
	slot := uint32(block.Header.Slot) % uint32(types.EpochLength)
	epoch := uint32(block.Header.Slot) / uint32(types.EpochLength)
	ctx := logger.FormatContext(hashStr, slot, epoch, "ImportBlock")
	logger.Debugf("%s Processing...", ctx)

	// Get the latest block
	storeInstance := store.GetInstance()

	blocks := storeInstance.GetBlocks()
	if len(blocks) > 0 {
		latestBlock := storeInstance.GetLatestBlock()

		latestBlockHash, err := hash.ComputeBlockHeaderHash(latestBlock.Header)
		if err != nil {
			return types.StateRoot{}, fmt.Errorf("error computing latest block hash: %w", err)
		}

		if latestBlockHash != block.Header.Parent && latestBlockHash != headerHash {
			logger.Debugf("%s parent mismatch, trying to restore block and state", ctx)
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
	priorUnmatchedKeyVals := storeInstance.GetPriorStateUnmatchedKeyVals()
	serializedState = append(priorUnmatchedKeyVals, serializedState...)
	latestStateRoot := m.MerklizationSerializedState(serializedState)

	storeInstance.AddBlock(block)
	logger.Infof("%s Block 0x%x... added for ImportBlock", ctx, headerHash[:8])

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
		return latestStateRoot, err
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

	// Commit the state and persist the state to Redis
	storeInstance.StateCommit()

	return latestStateRoot, nil
}

func (s *FuzzServiceStub) SetState(header types.Header, stateKeyVals types.StateKeyVals, ancestry types.Ancestry) (types.StateRoot, error) {
	// Build context for logging
	headerHash, _ := hash.ComputeBlockHeaderHash(header)
	hashStr := hex.EncodeToString(headerHash[:])
	slot := uint32(header.Slot) % uint32(types.EpochLength)
	epoch := uint32(header.Slot) / uint32(types.EpochLength)
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
		// logger.Debugf("%s Ancestry items: %v", ctx, ancestry)
	}

	state, unmatchedKeyVals, err := m.StateKeyValsToState(stateKeyVals)
	if err != nil {
		logger.Errorf("%s failed to convert state keyvals: %v", ctx, err)
		return types.StateRoot{}, err
	}

	// For genesis SetState (from genesis.json), we *do* want to persist the
	// header (w/o validation) and state to Redis so that later ImportBlock
	// calls can restore using the genesis header hash via RestoreBlockAndState.

	// Prepare posterior state to match the initialized state.
	storeInstance.SetPostStateUnmatchedKeyVals(unmatchedKeyVals.DeepCopy())
	storeInstance.GetPosteriorStates().SetState(state)

	// Add the genesis block to the in-memory block list.
	genesisBlock := types.Block{
		Header: header,
	}
	storeInstance.AddBlock(genesisBlock)

	// Persist block + state to Redis and record ancestry via StateCommit.
	storeInstance.StateCommit()

	// Empty posterior state
	posteriorStates := store.NewPosteriorStates()
	storeInstance.GetPosteriorStates().SetState(posteriorStates.GetState())

	serializedState, _ := m.StateEncoder(state)

	stateRoot := m.MerklizationSerializedState(append(unmatchedKeyVals, serializedState...))

	logger.Infof("%s Init completed, header hash: 0x%x..., state root: 0x%x", ctx, headerHash[:8], stateRoot)
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
		epoch = uint32(block.Header.Slot) / uint32(types.EpochLength)
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
