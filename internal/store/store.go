package store

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

var (
	initOnce    sync.Once
	globalStore *Store
)

// Store represents a thread-safe global state container
type Store struct {
	// INFO: Add more fields here
	unfinalizedBlocks          *UnfinalizedBlocks
	finalizedIndex             map[types.HeaderHash]bool
	processingBlock            *ProcessingBlock
	priorStates                *PriorStates
	intermediateStates         *IntermediateStates
	posteriorStates            *PosteriorStates
	ancestry                   *AncestryStore
	posteriorCurrentValidators *PosteriorCurrentValidators
	preStateUnmatchedKeyVals   types.StateKeyVals
	postStateUnmatchedKeyVals  types.StateKeyVals
}

// GetInstance returns the singleton instance of Store.
// If the instance doesn't exist, it creates one.
func GetInstance() *Store {
	initOnce.Do(func() {
		globalStore = &Store{
			unfinalizedBlocks:          NewUnfinalizedBlocks(),
			finalizedIndex:             make(map[types.HeaderHash]bool),
			processingBlock:            NewProcessingBlock(),
			priorStates:                NewPriorStates(),
			intermediateStates:         NewIntermediateStates(),
			posteriorStates:            NewPosteriorStates(),
			ancestry:                   NewAncestryStore(),
			posteriorCurrentValidators: NewPosteriorValidators(),
			preStateUnmatchedKeyVals:   types.StateKeyVals{},
			postStateUnmatchedKeyVals:  types.StateKeyVals{},
		}
		logger.Debug("🚀 Store initialized")
	})
	return globalStore
}

func ResetInstance() {
	// reset globalStore
	globalStore = &Store{
		unfinalizedBlocks:          NewUnfinalizedBlocks(),
		finalizedIndex:             make(map[types.HeaderHash]bool),
		processingBlock:            NewProcessingBlock(),
		priorStates:                NewPriorStates(),
		intermediateStates:         NewIntermediateStates(),
		posteriorStates:            NewPosteriorStates(),
		ancestry:                   NewAncestryStore(),
		posteriorCurrentValidators: NewPosteriorValidators(),
		preStateUnmatchedKeyVals:   types.StateKeyVals{},
		postStateUnmatchedKeyVals:  types.StateKeyVals{},
	}
	logger.Debug("🚀 Store reset")
}

func (s *Store) AddBlock(block types.Block) {
	s.unfinalizedBlocks.AddBlock(block)
	if err := s.persistBlockMapping(block); err != nil {
		logger.Errorf("AddBlock: failed to index block: %v", err)
	}
}

func (s *Store) CleanupBlock() {
	s.unfinalizedBlocks = NewUnfinalizedBlocks()
}

// KeepBlocksUpTo keeps only blocks up to and including the specified headerHash.
func (s *Store) KeepBlocksUpTo(headerHash types.HeaderHash) {
	s.unfinalizedBlocks.KeepBlocksUpTo(headerHash)
}

func (s *Store) GetBlocks() []types.Block {
	return s.unfinalizedBlocks.GetAllAncientBlocks()
}

func (s *Store) GetBlock() types.Block {
	return s.unfinalizedBlocks.GetLatestBlock()
}

func (s *Store) GetLatestBlock() types.Block {
	return s.unfinalizedBlocks.GetLatestBlock()
}

func (s *Store) GetProcessingBlockPointer() *ProcessingBlock {
	return s.processingBlock
}

func (s *Store) GenerateGenesisBlock(block types.Block) {
	s.unfinalizedBlocks.GenerateGenesisBlock(block)
	// Genesis block is always finalized
	s.finalizedIndex[block.Header.Parent] = true
	if err := s.persistBlockMapping(block); err != nil {
		logger.Errorf("GenerateGenesisBlock: failed to index block: %v", err)
	}
}

// Finalized Blocks Management

// FinalizeBlock marks a block as finalized by its hash
func (s *Store) FinalizeBlock(blockHash types.HeaderHash) {
	s.finalizedIndex[blockHash] = true
}

// IsBlockFinalized checks if a block is finalized
func (s *Store) IsBlockFinalized(blockHash types.HeaderHash) bool {
	return s.finalizedIndex[blockHash]
}

// GetFinalizedBlocks returns all finalized blocks
func (s *Store) GetFinalizedBlocks() []types.Block {
	allBlocks := s.unfinalizedBlocks.GetAllAncientBlocks()

	finalizedBlocksIdx := -1
	for i := len(allBlocks) - 1; i >= 0; i-- {
		block := allBlocks[i]
		if s.IsBlockFinalized(block.Header.Parent) {
			finalizedBlocksIdx = i
			break
		}
	}

	if finalizedBlocksIdx == -1 {
		return []types.Block{}
	}

	return allBlocks[:finalizedBlocksIdx+1]
}

// GetFinalizedBlocks returns all finalized blocks
func (s *Store) GetFinalizedBlock() types.Block {
	allBlocks := s.unfinalizedBlocks.GetAllAncientBlocks()

	finalizedBlocksIdx := -1
	found := false
	for i := len(allBlocks) - 1; i >= 0; i-- {
		block := allBlocks[i]
		if s.IsBlockFinalized(block.Header.Parent) {
			finalizedBlocksIdx = i
			found = true
			break
		}
	}

	if !found {
		return types.Block{}
	}

	return allBlocks[finalizedBlocksIdx]
}

// GetUnfinalizedBlocks returns all unfinalized blocks
func (s *Store) GetUnfinalizedBlocks() []types.Block {
	allBlocks := s.unfinalizedBlocks.GetAllAncientBlocks()
	finalizedBlockIdx := -1

	for i := len(allBlocks) - 1; i >= 0; i-- {
		if s.IsBlockFinalized(allBlocks[i].Header.Parent) {
			finalizedBlockIdx = i
			break
		}
	}

	return allBlocks[finalizedBlockIdx+1:]
}

// GetLatestFinalizedBlock returns the most recent finalized block
func (s *Store) GetLatestFinalizedBlock() types.Block {
	allBlocks := s.unfinalizedBlocks.GetAllAncientBlocks()

	// Search from the end to find the latest finalized block
	for i := len(allBlocks) - 1; i >= 0; i-- {
		if s.IsBlockFinalized(allBlocks[i].Header.Parent) {
			return allBlocks[i]
		}
	}

	return types.Block{}
}

// CleanupOldFinalizedBlocks removes old finalized blocks from memory
// This is a simple implementation - you might want to implement more sophisticated cleanup
func (s *Store) CleanupOldFinalizedBlocks(keepCount int) {
	// TODO: Implement this
}

// Set
func (s *Store) GetPriorStates() *PriorStates {
	return s.priorStates
}

func (s *Store) GetIntermediateStates() *IntermediateStates {
	return s.intermediateStates
}

func (s *Store) GetPosteriorStates() *PosteriorStates {
	return s.posteriorStates
}

func (s *Store) GenerateGenesisState(state types.State) {
	s.posteriorStates.GenerateGenesisState(state)
	logger.Debug("🚀 Genesis state generated")
}

// Ancestry methods (replaces AncestorHeaders)

// AddAncestorHeader is a convenience method that converts Header to AncestryItem and adds it.
func (s *Store) AddAncestorHeader(header types.Header) {
	headerHash, err := hash.ComputeBlockHeaderHash(header)
	if err != nil {
		logger.Errorf("AddAncestorHeader: failed to compute header hash: %v", err)
		return
	}
	s.AppendAncestry(types.Ancestry{
		{
			Slot:       header.Slot,
			HeaderHash: headerHash,
		},
	})
}

// AppendAncestry appends ancestry items to the store.
func (s *Store) AppendAncestry(ancestry types.Ancestry) {
	s.ancestry.AppendAncestry(ancestry)
}

// KeepAncestryUpTo keeps only ancestry items up to and including the specified headerHash.
func (s *Store) KeepAncestryUpTo(headerHash types.HeaderHash) {
	s.ancestry.KeepAncestryUpTo(headerHash)
}

// GetAncestry returns the current ancestry.
func (s *Store) GetAncestry() types.Ancestry {
	return s.ancestry.GetAncestry()
}

// ClearAncestry clears all ancestry from the store.
func (s *Store) ClearAncestry() {
	s.ancestry.Clear()
}

// PosteriorCurrentValidators

func (s *Store) AddPosteriorCurrentValidator(validator types.Validator) {
	s.posteriorCurrentValidators.AddValidator(validator)
}

func (s *Store) GetPosteriorCurrentValidators() types.ValidatorsData {
	return s.posteriorCurrentValidators.GetValidators()
}

func (s *Store) GetPosteriorCurrentValidatorByIndex(index types.ValidatorIndex) types.Validator {
	return s.posteriorCurrentValidators.GetValidatorByIndex(index)
}

// post-state update to pre-state
func (s *Store) StateCommit() {
	latestBlock := s.GetLatestBlock()

	blockHeaderHash, err := hash.ComputeBlockHeaderHash(latestBlock.Header)
	if err != nil {
		logger.Errorf("StateCommit: failed to encode header: %v", err)
	} else {
		posteriorState := s.GetPosteriorStates().GetState()

		// Persist state for block
		err = s.PersistStateForBlock(blockHeaderHash, posteriorState)
		if err != nil {
			logger.Errorf("StateCommit: failed to persist state: %v", err)
		} else {
			logger.Debugf("StateCommit: persisted state for block 0x%x", blockHeaderHash[:8])
		}

		// Persist block mapping
		err = s.persistBlockMapping(latestBlock)
		if err != nil {
			logger.Errorf("StateCommit: failed to persist block: %v", err)
		} else {
			logger.Debugf("StateCommit: persisted block 0x%x", blockHeaderHash[:8])
		}

		// Add to ancestry (avoid duplicating the latest header if it's already the last item)
		currentItem := types.AncestryItem{
			Slot:       latestBlock.Header.Slot,
			HeaderHash: blockHeaderHash,
		}
		existingAncestry := s.GetAncestry()
		if len(existingAncestry) == 0 {
			s.AppendAncestry(types.Ancestry{currentItem})
		} else {
			last := existingAncestry[len(existingAncestry)-1]
			if last.Slot != currentItem.Slot || last.HeaderHash != currentItem.HeaderHash {
				s.AppendAncestry(types.Ancestry{currentItem})
			} else {
				logger.Debugf("StateCommit: latest header already in ancestry (slot=%d, hash=0x%x), skipping append", currentItem.Slot, currentItem.HeaderHash[:8])
			}
		}
	}

	posterState := s.GetPosteriorStates().GetState()
	s.GetPriorStates().SetState(posterState)
	postUnmatchedKeyVal := s.GetPostStateUnmatchedKeyVals()
	s.SetPriorStateUnmatchedKeyVals(postUnmatchedKeyVal.DeepCopy())
	s.GetPosteriorStates().SetState(*NewPosteriorStates().state)
}

// PersistStateForBlock persists the state for a given block to Redis
func (s *Store) PersistStateForBlock(blockHeaderHash types.HeaderHash, state types.State) error {
	redisBackend, err := GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get redis backend: %w", err)
	}

	serializedState, err := m.StateEncoder(state)
	if err != nil {
		return fmt.Errorf("failed to encode state: %w", err)
	}

	unmatchedKeyVals := s.GetPostStateUnmatchedKeyVals()
	fullStateKeyVals := append(serializedState, unmatchedKeyVals...)

	// Sort the fullStateKeyVals by Key to ensure consistent Merklization
	sort.Slice(fullStateKeyVals, func(i, j int) bool {
		return bytes.Compare(fullStateKeyVals[i].Key[:], fullStateKeyVals[j].Key[:]) < 0
	})

	stateRoot := m.MerklizationSerializedState(fullStateKeyVals)

	ctx := context.Background()

	err = redisBackend.StoreStateRootByBlockHash(ctx, blockHeaderHash, stateRoot)
	if err != nil {
		return fmt.Errorf("failed to store state root mapping: %w", err)
	}

	err = redisBackend.StoreStateData(ctx, stateRoot, fullStateKeyVals)
	if err != nil {
		return fmt.Errorf("failed to store state data: %w", err)
	}

	return nil
}

// GetStateByBlockHash retrieves state data for a given block from Redis
func (s *Store) GetStateByBlockHash(blockHeaderHash types.HeaderHash) (types.StateKeyVals, error) {
	redisBackend, err := GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis backend: %w", err)
	}

	ctx := context.Background()
	return redisBackend.GetStateByBlockHash(ctx, blockHeaderHash)
}

// UnmatchedKeyVals
func (s *Store) GetPriorStateUnmatchedKeyVals() types.StateKeyVals {
	return s.preStateUnmatchedKeyVals
}

func (s *Store) SetPriorStateUnmatchedKeyVals(unmatchedKeyVals types.StateKeyVals) {
	s.preStateUnmatchedKeyVals = unmatchedKeyVals
}

func (s *Store) GetPostStateUnmatchedKeyVals() types.StateKeyVals {
	return s.postStateUnmatchedKeyVals
}

func (s *Store) SetPostStateUnmatchedKeyVals(unmatchedKeyVals types.StateKeyVals) {
	s.postStateUnmatchedKeyVals = unmatchedKeyVals
}

func (s *Store) GetBlockByHash(headerHash types.HeaderHash) (types.Block, error) {
	redisBackend, err := GetRedisBackend()
	if err != nil {
		return types.Block{}, fmt.Errorf("failed to get redis backend: %w", err)
	}

	ctx := context.Background()
	block, err := redisBackend.GetBlockByHash(ctx, types.OpaqueHash(headerHash))
	if err != nil {
		return types.Block{}, err
	}
	if block == nil {
		return types.Block{}, fmt.Errorf("block not found for hash 0x%x", headerHash[:8])
	}

	return *block, nil
}

func (s *Store) persistBlockMapping(block types.Block) error {
	headerHash, err := hash.ComputeBlockHeaderHash(block.Header)
	if err != nil {
		return fmt.Errorf("failed to compute block header hash: %w", err)
	}

	redisBackend, err := GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get redis backend: %w", err)
	}

	ctx := context.Background()
	if err := redisBackend.StoreBlockByHash(ctx, &block, types.OpaqueHash(headerHash)); err != nil {
		return fmt.Errorf("failed to persist block in redis: %w", err)
	}

	return nil
}

func (s *Store) GetBlockAndState(blockHeaderHash types.HeaderHash) (types.Block, types.StateKeyVals, error) {
	block, err := s.GetBlockByHash(blockHeaderHash)
	if err != nil {
		return types.Block{}, nil, fmt.Errorf("failed to retrieve block for hash 0x%x: %w", blockHeaderHash[:8], err)
	}

	state, err := s.GetStateByBlockHash(blockHeaderHash)
	if err != nil {
		return types.Block{}, nil, fmt.Errorf("failed to restore state for hash 0x%x: %w", blockHeaderHash[:8], err)
	}

	return block, state, nil
}

func (s *Store) RestoreBlockAndState(blockHeaderHash types.HeaderHash) error {
	block, stateKeyVals, err := s.GetBlockAndState(blockHeaderHash)
	if err != nil {
		return fmt.Errorf("failed to restore block and state for hash 0x%x: %w", blockHeaderHash[:8], err)
	}

	// Restore state and storage key-vals
	state, unmatchedKeyVals, err := m.StateKeyValsToState(stateKeyVals)
	if err != nil {
		return err
	}

	s.GetPriorStates().SetState(state)
	s.SetPriorStateUnmatchedKeyVals(unmatchedKeyVals)
	s.SetPostStateUnmatchedKeyVals(unmatchedKeyVals.DeepCopy())
	// Keep only blocks up to the restored headerHash (fallback point)
	s.KeepBlocksUpTo(blockHeaderHash)
	// Add the restored block if it's not already in the list
	// Check if the latest block matches the restored block to avoid duplicates
	blocks := s.GetBlocks()
	if len(blocks) == 0 {
		// No blocks found, add the restored block
		s.AddBlock(block)
	} else {
		// Check if the latest block is the one we're restoring
		latestBlockHash, err := hash.ComputeBlockHeaderHash(blocks[len(blocks)-1].Header)
		if err != nil || latestBlockHash != blockHeaderHash {
			// Latest block doesn't match, add the restored block
			s.AddBlock(block)
		}
		// Otherwise, the block is already in the list, no need to add
	}

	// Keep only ancestry up to the restored headerHash (fallback point)
	s.KeepAncestryUpTo(blockHeaderHash)

	return nil
}

// // ServiceAccountDerivatives (This is tmp used waiting for more testvector to verify)

// // Get
// func (s *Store) GetServiceAccountDerivatives() types.ServiceAccountDerivatives {
// 	return s.serviceAccountDerivatives.GetServiceAccountDerivatives()
// }

// // Set
// func (s *Store) GetServiceAccountDerivatives() *ServiceAccountDerivatives {
// 	return s.serviceAccountDerivatives
// }
