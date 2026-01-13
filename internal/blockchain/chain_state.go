package blockchain

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

var (
	initOnce         sync.Once
	globalChainState *ChainState
)

// ChainState represents a thread-safe global container,
// manages the state and blocks of the blockchain
type ChainState struct {
	// Data access layer
	repo *store.Repository

	// State management
	priorStates        *PriorStates
	intermediateStates *IntermediateStates
	posteriorStates    *PosteriorStates

	// Block management
	unfinalizedBlocks *UnfinalizedBlocks
	finalizedIndex    map[types.HeaderHash]bool
	processingBlock   *ProcessingBlock
	ancestry          *AncestryCache

	posteriorCurrentValidators *PosteriorCurrentValidators
	preStateUnmatchedKeyVals   types.StateKeyVals
	postStateUnmatchedKeyVals  types.StateKeyVals
}

// GetInstance returns the singleton instance of ChainState.
// If the instance doesn't exist, it creates one.
func GetInstance() *ChainState {
	initOnce.Do(func() {
		db := memory.NewDatabase()
		repo := store.NewRepository(db)
		globalChainState = &ChainState{
			repo: repo,

			priorStates:        NewPriorStates(),
			intermediateStates: NewIntermediateStates(),
			posteriorStates:    NewPosteriorStates(),

			unfinalizedBlocks: NewUnfinalizedBlocks(),
			finalizedIndex:    make(map[types.HeaderHash]bool),
			processingBlock:   NewProcessingBlock(),
			ancestry:          NewAncestryCache(),

			posteriorCurrentValidators: NewPosteriorValidators(),
			preStateUnmatchedKeyVals:   types.StateKeyVals{},
			postStateUnmatchedKeyVals:  types.StateKeyVals{},
		}
		logger.Debug("🚀 ChainState initialized")
	})
	return globalChainState
}

func ResetInstance() {
	// reset globalStore
	db := memory.NewDatabase()
	repo := store.NewRepository(db)
	globalChainState = &ChainState{
		repo: repo,

		priorStates:        NewPriorStates(),
		intermediateStates: NewIntermediateStates(),
		posteriorStates:    NewPosteriorStates(),

		unfinalizedBlocks: NewUnfinalizedBlocks(),
		finalizedIndex:    make(map[types.HeaderHash]bool),
		processingBlock:   NewProcessingBlock(),
		ancestry:          NewAncestryCache(),

		posteriorCurrentValidators: NewPosteriorValidators(),
		preStateUnmatchedKeyVals:   types.StateKeyVals{},
		postStateUnmatchedKeyVals:  types.StateKeyVals{},
	}
	logger.Debug("🚀 ChainState reset")
}

// Blockchain interface implementation

func (cs *ChainState) GetBlock(hash types.HeaderHash) (types.Block, error) {
	return cs.GetBlockByHash(hash)
}

func (cs *ChainState) GetBlockNumber(hash types.HeaderHash) (uint32, error) {
	timeSlot, err := cs.repo.GetHeaderTimeSlot(cs.repo.Database(), hash)
	if err != nil {
		return 0, err
	}
	return uint32(timeSlot), nil
}

func (cs *ChainState) GetBlockHashByNumber(number uint32) ([]types.HeaderHash, error) {
	slot := types.TimeSlot(number)
	hashes, err := cs.repo.GetHeaderHashesByTimeSlot(cs.repo.Database(), slot)
	if err != nil {
		return nil, fmt.Errorf("failed to get block hashes: %w", err)
	}
	return hashes, nil
}

func (cs *ChainState) GenesisBlockHash() types.HeaderHash {
	genesisBlock := cs.GetGenesisBlock()
	hash, err := hash.ComputeBlockHeaderHash(genesisBlock.Header)
	if err != nil {
		return types.HeaderHash{}
	}
	return hash
}

func (cs *ChainState) StoreBlockByHash(hash types.HeaderHash, block *types.Block) error {
	cs.AddBlock(*block)
	return nil
}

// --- Block Management ---

func (cs *ChainState) AddBlock(block types.Block) {
	cs.unfinalizedBlocks.AddBlock(block)
	if err := cs.persistBlockMapping(block); err != nil {
		logger.Errorf("AddBlock: failed to index block: %v", err)
	}
}

func (cs *ChainState) GetBlocks() []types.Block {
	return cs.unfinalizedBlocks.GetAllAncientBlocks()
}

func (cs *ChainState) GetLatestBlock() types.Block {
	return cs.unfinalizedBlocks.GetLatestBlock()
}

func (cs *ChainState) GetProcessingBlockPointer() *ProcessingBlock {
	return cs.processingBlock
}

func (cs *ChainState) GenerateGenesisBlock(block types.Block) error {
	cs.unfinalizedBlocks.GenerateGenesisBlock(block)
	// Genesis block is always finalized
	cs.finalizedIndex[block.Header.Parent] = true
	if err := cs.persistBlockMapping(block); err != nil {
		logger.Errorf("GenerateGenesisBlock: failed to index block: %v", err)
	}
	return nil
}

// Finalized Blocks Management

// FinalizeBlock marks a block as finalized by its hash
func (cs *ChainState) FinalizeBlock(blockHash types.HeaderHash) {
	cs.finalizedIndex[blockHash] = true
}

// IsBlockFinalized checks if a block is finalized
func (cs *ChainState) IsBlockFinalized(blockHash types.HeaderHash) bool {
	return cs.finalizedIndex[blockHash]
}

// GetFinalizedBlocks returns all finalized blocks
func (cs *ChainState) GetFinalizedBlocks() []types.Block {
	allBlocks := cs.unfinalizedBlocks.GetAllAncientBlocks()

	finalizedBlocksIdx := -1
	for i := len(allBlocks) - 1; i >= 0; i-- {
		block := allBlocks[i]
		if cs.IsBlockFinalized(block.Header.Parent) {
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
func (cs *ChainState) GetFinalizedBlock() types.Block {
	allBlocks := cs.unfinalizedBlocks.GetAllAncientBlocks()

	finalizedBlocksIdx := -1
	found := false
	for i := len(allBlocks) - 1; i >= 0; i-- {
		block := allBlocks[i]
		if cs.IsBlockFinalized(block.Header.Parent) {
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
func (cs *ChainState) GetUnfinalizedBlocks() []types.Block {
	allBlocks := cs.unfinalizedBlocks.GetAllAncientBlocks()
	finalizedBlockIdx := -1

	for i := len(allBlocks) - 1; i >= 0; i-- {
		if cs.IsBlockFinalized(allBlocks[i].Header.Parent) {
			finalizedBlockIdx = i
			break
		}
	}

	return allBlocks[finalizedBlockIdx+1:]
}

// GetLatestFinalizedBlock returns the most recent finalized block
func (cs *ChainState) GetLatestFinalizedBlock() types.Block {
	allBlocks := cs.unfinalizedBlocks.GetAllAncientBlocks()

	// Search from the end to find the latest finalized block
	for i := len(allBlocks) - 1; i >= 0; i-- {
		if cs.IsBlockFinalized(allBlocks[i].Header.Parent) {
			return allBlocks[i]
		}
	}

	return types.Block{}
}

// CleanupOldFinalizedBlocks removes old finalized blocks from memory
// This is a simple implementation - you might want to implement more sophisticated cleanup
func (cs *ChainState) CleanupOldFinalizedBlocks(keepCount int) {
	// TODO: Implement this
}

/*
	State management
*/

func (cs *ChainState) GetPriorStates() *PriorStates {
	return cs.priorStates
}

func (cs *ChainState) GetIntermediateStates() *IntermediateStates {
	return cs.intermediateStates
}

func (cs *ChainState) GetPosteriorStates() *PosteriorStates {
	return cs.posteriorStates
}

func (cs *ChainState) GenerateGenesisState(state types.State) {
	cs.posteriorStates.GenerateGenesisState(state)
	logger.Debug("🚀 Genesis state generated")
}

// post-state update to pre-state
func (cs *ChainState) StateCommit() {
	latestBlock := cs.GetLatestBlock()

	blockHeaderHash, err := hash.ComputeBlockHeaderHash(latestBlock.Header)
	if err != nil {
		logger.Errorf("StateCommit: failed to encode header: %v", err)
	} else {
		posteriorState := cs.GetPosteriorStates().GetState()

		// Persist state for block
		err = cs.PersistStateForBlock(blockHeaderHash, posteriorState)
		if err != nil {
			logger.Errorf("StateCommit: failed to persist state: %v", err)
		} else {
			logger.Debugf("StateCommit: persisted state for block 0x%x", blockHeaderHash[:8])
		}

		// Persist block mapping
		err = cs.repo.SaveBlock(cs.repo.Database(), &latestBlock)
		if err != nil {
			logger.Errorf("StateCommit: failed to persist block: %v", err)
		} else {
			logger.Debugf("StateCommit: persisted block 0x%x", blockHeaderHash[:8])
		}

		existingAncestry := cs.GetAncestry()
		if len(existingAncestry) > 0 {
			// Add to ancestry (avoid duplicating the latest header if it'cs already the last item)
			currentItem := types.AncestryItem{
				Slot:       latestBlock.Header.Slot,
				HeaderHash: blockHeaderHash,
			}
			last := existingAncestry[len(existingAncestry)-1]
			if last.Slot != currentItem.Slot || last.HeaderHash != currentItem.HeaderHash {
				cs.AppendAncestry(types.Ancestry{currentItem})
			} else {
				logger.Debugf("StateCommit: latest header already in ancestry (slot=%d, hash=0x%x), skipping append", currentItem.Slot, currentItem.HeaderHash[:8])
			}
		}
	}

	posterState := cs.GetPosteriorStates().GetState()
	cs.GetPriorStates().SetState(posterState)
	postUnmatchedKeyVal := cs.GetPostStateUnmatchedKeyVals()
	cs.SetPriorStateUnmatchedKeyVals(postUnmatchedKeyVal.DeepCopy())
	cs.GetPosteriorStates().SetState(*NewPosteriorStates().state)
}

/*
	Ancestry management
*/

// AddAncestorHeader is a convenience method that converts Header to AncestryItem and adds it.
func (cs *ChainState) AddAncestorHeader(header types.Header) {
	headerHash, err := hash.ComputeBlockHeaderHash(header)
	if err != nil {
		logger.Errorf("AddAncestorHeader: failed to compute header hash: %v", err)
		return
	}
	cs.AppendAncestry(types.Ancestry{
		{
			Slot:       header.Slot,
			HeaderHash: headerHash,
		},
	})
}

// AppendAncestry appends ancestry items to the blockchain.
func (cs *ChainState) AppendAncestry(ancestry types.Ancestry) {
	cs.ancestry.AppendAncestry(ancestry)
}

// KeepAncestryUpTo keeps only ancestry items up to and including the specified headerHash.
func (cs *ChainState) KeepAncestryUpTo(headerHash types.HeaderHash) {
	cs.ancestry.KeepAncestryUpTo(headerHash)
}

// GetAncestry returns the current ancestry.
func (cs *ChainState) GetAncestry() types.Ancestry {
	return cs.ancestry.GetAncestry()
}

// ClearAncestry clears all ancestry from the blockchain.
func (cs *ChainState) ClearAncestry() {
	cs.ancestry.Clear()
}

/*
	PosteriorCurrentValidators management
*/

func (cs *ChainState) AddPosteriorCurrentValidator(validator types.Validator) {
	cs.posteriorCurrentValidators.AddValidator(validator)
}

func (cs *ChainState) GetPosteriorCurrentValidators() types.ValidatorsData {
	return cs.posteriorCurrentValidators.GetValidators()
}

func (cs *ChainState) GetPosteriorCurrentValidatorByIndex(index types.ValidatorIndex) types.Validator {
	return cs.posteriorCurrentValidators.GetValidatorByIndex(index)
}

/*
	UnmatchedKeyVals management
*/

func (cs *ChainState) GetPriorStateUnmatchedKeyVals() types.StateKeyVals {
	return cs.preStateUnmatchedKeyVals.DeepCopy()
}

func (cs *ChainState) SetPriorStateUnmatchedKeyVals(unmatchedKeyVals types.StateKeyVals) {
	cs.preStateUnmatchedKeyVals = unmatchedKeyVals
}

func (cs *ChainState) GetPostStateUnmatchedKeyVals() types.StateKeyVals {
	return cs.postStateUnmatchedKeyVals.DeepCopy()
}

func (cs *ChainState) SetPostStateUnmatchedKeyVals(unmatchedKeyVals types.StateKeyVals) {
	cs.postStateUnmatchedKeyVals = unmatchedKeyVals
}

/*
	Block retrieval
*/

// GetGenesisBlock retrieves the genesis block from the blockchain.
// This function panics if the genesis block does not exist, as it is assumed to always be present.
// TODO: This is because respecting the current implementation usage of `GenerateGenesisBlock`, but later should ensure genesis block exist at store initialization.
func (cs *ChainState) GetGenesisBlock() types.Block {
	headerHash, err := cs.repo.GetCanonicalHash(cs.repo.Database(), 0)
	if err != nil {
		// This is coding error, genesis block must exist.
		// ChainState instance without genesis block should never happen.
		panic(fmt.Sprintf("genesis block must exist, failed to retrieve genesis block hash: %v", err))
	}

	block, err := cs.repo.GetBlock(cs.repo.Database(), headerHash, 0)
	if err != nil {
		// This is coding error, genesis block must exist.
		// ChainState instance without genesis block should never happen.
		panic(fmt.Sprintf("genesis block must exist, failed to retrieve genesis block: %v", err))
	}
	return *block
}

func (cs *ChainState) GetBlockByHash(headerHash types.HeaderHash) (types.Block, error) {
	timeSlot, err := cs.repo.GetHeaderTimeSlot(cs.repo.Database(), headerHash)
	if err != nil {
		return types.Block{}, err
	}

	block, err := cs.repo.GetBlock(cs.repo.Database(), headerHash, timeSlot)
	if err != nil {
		return types.Block{}, err
	}

	return *block, nil
}

/*
	State persistence
*/

// PersistStateForBlock persists the state for a given block to Redis
func (cs *ChainState) PersistStateForBlock(blockHeaderHash types.HeaderHash, state types.State) error {
	serializedState, err := m.StateEncoder(state)
	if err != nil {
		return fmt.Errorf("failed to encode state: %w", err)
	}

	unmatchedKeyVals := cs.GetPostStateUnmatchedKeyVals()
	fullStateKeyVals := append(serializedState, unmatchedKeyVals...)

	// Sort the fullStateKeyVals by Key to ensure consistent Merklization
	sort.Slice(fullStateKeyVals, func(i, j int) bool {
		return bytes.Compare(fullStateKeyVals[i].Key[:], fullStateKeyVals[j].Key[:]) < 0
	})

	stateRoot := m.MerklizationSerializedState(fullStateKeyVals)

	err = cs.repo.SaveStateRootByHeaderHash(cs.repo.Database(), blockHeaderHash, stateRoot)
	if err != nil {
		return fmt.Errorf("failed to store state root mapping: %w", err)
	}

	err = cs.repo.SaveStateData(cs.repo.Database(), stateRoot, fullStateKeyVals)
	if err != nil {
		return fmt.Errorf("failed to store state data: %w", err)
	}

	return nil
}

// GetStateByBlockHash retrieves state data for a given block from Redis
func (cs *ChainState) GetStateByBlockHash(blockHeaderHash types.HeaderHash) (types.StateKeyVals, error) {
	return cs.repo.GetStateDataByHeaderHash(cs.repo.Database(), blockHeaderHash)
}

func (cs *ChainState) persistBlockMapping(block types.Block) error {
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

func (cs *ChainState) GetBlockAndState(blockHeaderHash types.HeaderHash) (types.Block, types.StateKeyVals, error) {
	block, err := cs.GetBlockByHash(blockHeaderHash)
	if err != nil {
		return types.Block{}, nil, fmt.Errorf("failed to retrieve block for hash 0x%x: %w", blockHeaderHash[:8], err)
	}

	state, err := cs.GetStateByBlockHash(blockHeaderHash)
	if err != nil {
		return types.Block{}, nil, fmt.Errorf("failed to restore state for hash 0x%x: %w", blockHeaderHash[:8], err)
	}

	return block, state, nil
}

func (cs *ChainState) RestoreBlockAndState(blockHeaderHash types.HeaderHash) error {
	block, stateKeyVals, err := cs.GetBlockAndState(blockHeaderHash)
	if err != nil {
		return fmt.Errorf("failed to restore block and state for hash 0x%x: %w", blockHeaderHash[:8], err)
	}

	// Restore state and storage key-vals
	state, unmatchedKeyVals, err := m.StateKeyValsToState(stateKeyVals)
	if err != nil {
		return err
	}

	cs.GetPriorStates().SetState(state)
	cs.SetPriorStateUnmatchedKeyVals(unmatchedKeyVals)
	cs.SetPostStateUnmatchedKeyVals(unmatchedKeyVals.DeepCopy())
	// Keep only blocks up to the restored headerHash (fallback point)
	cs.unfinalizedBlocks.KeepBlocksUpTo(blockHeaderHash)
	// Add the restored block if it'cs not already in the list
	// Check if the latest block matches the restored block to avoid duplicates
	blocks := cs.GetBlocks()
	if len(blocks) == 0 {
		// No blocks found, add the restored block
		cs.AddBlock(block)
	} else {
		// Check if the latest block is the one we're restoring
		latestBlockHash, err := hash.ComputeBlockHeaderHash(blocks[len(blocks)-1].Header)
		if err != nil || latestBlockHash != blockHeaderHash {
			// Latest block doesn't match, add the restored block
			cs.AddBlock(block)
		}
		// Otherwise, the block is already in the list, no need to add
	}

	// Keep only ancestry up to the restored headerHash (fallback point)
	cs.KeepAncestryUpTo(blockHeaderHash)

	// Clear verifier cache when restoring to a different state point
	// as the epoch may have changed
	ClearVerifierCache()

	return nil
}

func BuildStateRootInputKeyValsAndRoot(
	stateKeyVals types.StateKeyVals,
) (merkleInputKeyVals types.StateKeyVals, stateRoot types.StateRoot, err error) {
	state, unmatchedKeyVals, err := m.StateKeyValsToState(stateKeyVals)
	if err != nil {
		return nil, types.StateRoot{}, fmt.Errorf("StateKeyValsToState: %w", err)
	}

	serializedState, err := m.StateEncoder(state)
	if err != nil {
		return nil, types.StateRoot{}, fmt.Errorf("StateEncoder: %w", err)
	}

	merkleInputKeyVals = make(types.StateKeyVals, 0, len(unmatchedKeyVals)+len(serializedState))
	merkleInputKeyVals = append(merkleInputKeyVals, unmatchedKeyVals...)
	merkleInputKeyVals = append(merkleInputKeyVals, serializedState...)

	stateRoot = m.MerklizationSerializedState(merkleInputKeyVals)
	return merkleInputKeyVals, stateRoot, nil
}

func (cs *ChainState) SeedGenesisToBackend(
	ctx context.Context,
	genesisHeader types.Header,
	stateKeyVals types.StateKeyVals,
) (genesisBlockHash types.HeaderHash, genesisStateRoot types.StateRoot, err error) {
	h, err := hash.ComputeBlockHeaderHash(genesisHeader)
	if err != nil {
		return types.HeaderHash{}, types.StateRoot{}, fmt.Errorf("compute genesis header hash: %w", err)
	}
	genesisBlockHash = types.HeaderHash(h)

	merkleInputKeyVals, stateRoot, err := BuildStateRootInputKeyValsAndRoot(stateKeyVals)
	if err != nil {
		return types.HeaderHash{}, types.StateRoot{}, fmt.Errorf("build merkle input + state root: %w", err)
	}
	genesisStateRoot = stateRoot

	// check if non-zero since genesisHeader.ParentStateRoot can be zero
	zero := types.StateRoot{}
	if genesisHeader.ParentStateRoot != zero && genesisStateRoot != genesisHeader.ParentStateRoot {
		return types.HeaderHash{}, types.StateRoot{}, fmt.Errorf(
			"genesis state root mismatch: computed=0x%x header.ParentStateRoot=0x%x",
			genesisStateRoot,
			genesisHeader.ParentStateRoot,
		)
	}

	redisBackend, err := GetRedisBackend()
	if err != nil {
		return types.HeaderHash{}, types.StateRoot{}, fmt.Errorf("failed to get redis backend: %w", err)
	}

	if err := redisBackend.StoreStateData(ctx, genesisStateRoot, merkleInputKeyVals); err != nil {
		return types.HeaderHash{}, types.StateRoot{}, fmt.Errorf("store state_data: %w", err)
	}
	if err := redisBackend.StoreStateRootByBlockHash(ctx, genesisBlockHash, genesisStateRoot); err != nil {
		return types.HeaderHash{}, types.StateRoot{}, fmt.Errorf("store state_root mapping: %w", err)
	}

	return genesisBlockHash, genesisStateRoot, nil
}

// Compile-time interface check
var _ Blockchain = (*ChainState)(nil)
