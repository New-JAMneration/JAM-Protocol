package blockchain

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	pebbledb "github.com/New-JAMneration/JAM-Protocol/internal/database/provider/pebble"
	redisdb "github.com/New-JAMneration/JAM-Protocol/internal/database/provider/redis"
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzzenv"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

var (
	initOnce           sync.Once
	persistentDBOnce   sync.Once
	globalChainState   *ChainState
	globalPersistentDB database.Database
)

// ChainState represents a thread-safe global container,
// manages the state and blocks of the blockchain
type ChainState struct {
	// Data access layer
	repo           *store.Repository
	persistentRepo *store.Repository

	// Trie node persistence and reference counting
	trieStore *store.Trie

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

	// cache for leaf level merklization
	keyLevelCache *KeyLevelCache

	// ring buffer tracking recent state roots stored in repo (memory),
	// used to evict old entries and bound memory usage.
	recentStateRoots []types.StateRoot

	// lastCommittedStateRoot caches the state root from the most recent
	// successful StateCommit, avoiding redundant serialize+merklize.
	lastCommittedStateRoot types.StateRoot

	// tracks recent block hashes persisted to disk, used for fuzz-mode block pruning
	persistedBlockHashes []types.HeaderHash
}

func getPersistentDatabase() database.Database {
	persistentDBOnce.Do(func() {
		dbConfig := config.Config.Database
		switch dbConfig.Type {
		case "pebble":
			db, err := pebbledb.NewDatabase(dbConfig.DataDir, false)
			if err != nil {
				if strings.Contains(err.Error(), "lock") {
					logger.Errorf("Failed to initialize Pebble database: %v. Database may be locked by another process or previous instance. Please ensure no other process is using the database at %s", err, dbConfig.DataDir)
				} else {
					logger.Errorf("Failed to initialize Pebble database: %v", err)
				}
				globalPersistentDB = memory.NewDatabase()
			} else {
				globalPersistentDB = db
			}
		case "redis":
			redisConfig := config.Config.Redis
			globalPersistentDB = redisdb.NewDatabase(redisConfig.Address, redisConfig.Password, redisConfig.Port)
		default:
			logger.Warnf("Unknown database type: %s, using memory database", dbConfig.Type)
			globalPersistentDB = memory.NewDatabase()
		}
	})
	return globalPersistentDB
}

// GetInstance returns the singleton instance of ChainState.
// If the instance doesn't exist, it creates one.
func GetInstance() *ChainState {
	initOnce.Do(func() {
		repo := store.NewRepository(memory.NewDatabase())
		persistentDB := getPersistentDatabase()
		persistentRepo := store.NewRepository(persistentDB)

		globalChainState = &ChainState{
			repo:           repo,
			persistentRepo: persistentRepo,
			trieStore:      store.NewTrie(persistentDB),

			priorStates:        NewPriorStates(),
			intermediateStates: NewIntermediateStates(),
			posteriorStates:    NewPosteriorStates(),

			unfinalizedBlocks: NewUnfinalizedBlocks(),
			finalizedIndex:    make(map[types.HeaderHash]bool),
			processingBlock:   NewProcessingBlock(),
			ancestry:          NewAncestryCache(),

			posteriorCurrentValidators: NewPosteriorValidators(),

			keyLevelCache: NewKeyLevelCache(),
		}
		logger.Debug("🚀 ChainState initialized")
	})
	return globalChainState
}

func ResetInstance() {
	repo := store.NewRepository(memory.NewDatabase())
	persistentDB := getPersistentDatabase()
	persistentRepo := store.NewRepository(persistentDB)

	// Clean up trie data from previous session to avoid cross-session refcount accumulation.
	trieStore := store.NewTrie(persistentDB)
	if err := trieStore.DeleteAll(); err != nil {
		logger.Warnf("ResetInstance: failed to clean trie data: %v", err)
	}

	globalChainState = &ChainState{
		repo:           repo,
		persistentRepo: persistentRepo,
		trieStore:      trieStore,

		priorStates:        NewPriorStates(),
		intermediateStates: NewIntermediateStates(),
		posteriorStates:    NewPosteriorStates(),

		unfinalizedBlocks: NewUnfinalizedBlocks(),
		finalizedIndex:    make(map[types.HeaderHash]bool),
		processingBlock:   NewProcessingBlock(),
		ancestry:          NewAncestryCache(),

		posteriorCurrentValidators: NewPosteriorValidators(),

		keyLevelCache: NewKeyLevelCache(),
	}
	logger.Debug("🚀 ChainState reset")
}

// Blockchain interface implementation

// Database returns the persistent database (same as PersistentDatabase). Implements Blockchain.
func (cs *ChainState) Database() database.Database {
	return cs.persistentRepo.Database()
}

// PersistentDatabase returns the persistent database (same as used by persistentRepo).
// Used by CE handlers via type assertion when bc is *ChainState.
func (cs *ChainState) PersistentDatabase() database.Database {
	return cs.persistentRepo.Database()
}

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

func (cs *ChainState) GetCurrentHead() (types.Block, error) {
	// For now, treat the latest in-memory block as the current head.
	return cs.GetLatestBlock(), nil
}

func (cs *ChainState) SetCurrentHead(hash types.HeaderHash) {
	// NOTE: ChainState currently tracks head implicitly via in-memory ordering.
	// This is a placeholder for node/CE compatibility.
}

func (cs *ChainState) GetStateAt(headerHash types.HeaderHash) (types.StateKeyVals, error) {
	return cs.GetStateByBlockHash(headerHash)
}

func (cs *ChainState) GetStateRange(
	headerHash types.HeaderHash,
	keyStart types.StateKey,
	keyEnd types.StateKey,
	maxSize uint32,
) (types.StateKeyVals, error) {
	stateKeyVals, err := cs.GetStateAt(headerHash)
	if err != nil {
		return nil, err
	}

	// Ensure consistent ordering (required by CE).
	sort.Slice(stateKeyVals, func(i, j int) bool {
		return bytes.Compare(stateKeyVals[i].Key[:], stateKeyVals[j].Key[:]) < 0
	})

	out := make(types.StateKeyVals, 0)
	for _, kv := range stateKeyVals {
		if bytes.Compare(kv.Key[:], keyStart[:]) >= 0 && bytes.Compare(kv.Key[:], keyEnd[:]) < 0 {
			out = append(out, kv)
			if maxSize > 0 && uint32(len(out)) >= maxSize {
				break
			}
		}
	}

	return out, nil
}

func (cs *ChainState) GetBoundaryNodes(
	_ types.HeaderHash,
	_ types.StateKey,
	_ types.StateKey,
	_ uint32,
) ([]types.BoundaryNode, error) {
	// Boundary-node generation is not yet implemented for ChainState.
	// Returning an empty set keeps CE129 compatible in bootstrap/test contexts.
	return []types.BoundaryNode{}, nil
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

// StateCommit persists the posterior state, updates prior state, and returns the committed state root.
func (cs *ChainState) StateCommit() (types.StateRoot, error) {
	latestBlock := cs.GetLatestBlock()

	blockHeaderHash, err := hash.ComputeBlockHeaderHash(latestBlock.Header)
	if err != nil {
		return types.StateRoot{}, fmt.Errorf("StateCommit: failed to encode header: %w", err)
	}

	posteriorState := cs.GetPosteriorStates().GetState()

	stateRoot, err := cs.PersistStateForBlock(blockHeaderHash, posteriorState)
	if err != nil {
		return types.StateRoot{}, fmt.Errorf("StateCommit: failed to persist state: %w", err)
	}
	logger.Debugf("StateCommit: persisted state for block 0x%x", blockHeaderHash[:8])

	cs.lastCommittedStateRoot = stateRoot

	// Persist block mapping
	if err := cs.repo.SaveBlock(cs.repo.Database(), &latestBlock); err != nil {
		logger.Errorf("StateCommit: failed to persist block: %v", err)
	} else {
		logger.Debugf("StateCommit: persisted block 0x%x", blockHeaderHash[:8])
	}

	existingAncestry := cs.GetAncestry()
	if len(existingAncestry) > 0 {
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

	posterState := cs.GetPosteriorStates().GetState()
	cs.GetPriorStates().SetState(posterState)
	cs.GetPosteriorStates().SetState(*NewPosteriorStates().state)

	return stateRoot, nil
}

// StateCommitWithPreComputedState persists state using pre-computed stateRoot and
// fullStateKeyVals, skipping the redundant StateEncoder + Merklize in PersistStateForBlock.
func (cs *ChainState) StateCommitWithPreComputedState(
	blockHeaderHash types.HeaderHash,
	stateRoot types.StateRoot,
	fullStateKeyVals types.StateKeyVals,
) {
	sort.Slice(fullStateKeyVals, func(i, j int) bool {
		return bytes.Compare(fullStateKeyVals[i].Key[:], fullStateKeyVals[j].Key[:]) < 0
	})

	err := cs.repo.SaveStateRootByHeaderHash(cs.repo.Database(), blockHeaderHash, stateRoot)
	if err != nil {
		logger.Errorf("StateCommitWithPreComputedState: failed to store state root mapping: %v", err)
	}

	// Write state data to both memory (for fast reads) and disk (for persistence).
	err = cs.repo.SaveStateData(cs.repo.Database(), stateRoot, fullStateKeyVals)
	if err != nil {
		logger.Errorf("StateCommitWithPreComputedState: failed to store state data to memory: %v", err)
	} else {
		logger.Debugf("StateCommitWithPreComputedState: persisted state for block 0x%x", blockHeaderHash[:8])
	}
	if err = cs.persistentRepo.SaveStateData(cs.persistentRepo.Database(), stateRoot, fullStateKeyVals); err != nil {
		logger.Warnf("StateCommitWithPreComputedState: failed to store state data to disk: %v", err)
	}

	// Evict oldest state data from memory when exceeding MaxLookupAge entries.
	cs.recentStateRoots = append(cs.recentStateRoots, stateRoot)
	if len(cs.recentStateRoots) > types.MaxLookupAge {
		evict := cs.recentStateRoots[0]
		cs.recentStateRoots = cs.recentStateRoots[1:]
		cs.repo.DeleteStateData(cs.repo.Database(), evict)
		if fuzzenv.Enabled() {
			cs.persistentRepo.DeleteStateData(cs.persistentRepo.Database(), evict)
			if err := cs.trieStore.DeleteTrie(types.OpaqueHash(evict)); err != nil {
				logger.Warnf("StateCommitWithPreComputedState: failed to delete evicted trie %x: %v", evict[:8], err)
			}
		}
	}

	latestBlock := cs.GetLatestBlock()

	// Persist block mapping
	err = cs.repo.SaveBlock(cs.repo.Database(), &latestBlock)
	if err != nil {
		logger.Errorf("StateCommitWithPreComputedState: failed to persist block: %v", err)
	} else {
		logger.Debugf("StateCommitWithPreComputedState: persisted block 0x%x", blockHeaderHash[:8])
	}

	existingAncestry := cs.GetAncestry()
	if len(existingAncestry) > 0 {
		currentItem := types.AncestryItem{
			Slot:       latestBlock.Header.Slot,
			HeaderHash: blockHeaderHash,
		}
		last := existingAncestry[len(existingAncestry)-1]
		if last.Slot != currentItem.Slot || last.HeaderHash != currentItem.HeaderHash {
			cs.AppendAncestry(types.Ancestry{currentItem})
		} else {
			logger.Debugf("StateCommitWithPreComputedState: latest header already in ancestry (slot=%d, hash=0x%x), skipping append", currentItem.Slot, currentItem.HeaderHash[:8])
		}
	}

	posterState := cs.GetPosteriorStates().GetState()
	cs.GetPriorStates().SetState(posterState)
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

// PruneOldPersistentBlocks deletes old blocks and header mappings from persistent storage,
// keeping only the most recent FuzzPersistentRetainBlocks entries.
// Called after each successful ImportBlock in fuzz mode to prevent disk exhaustion.
func (cs *ChainState) PruneOldPersistentBlocks(headerHash types.HeaderHash) {
	cs.persistedBlockHashes = append(cs.persistedBlockHashes, headerHash)
	if len(cs.persistedBlockHashes) <= fuzzenv.FuzzPersistentRetainBlocks {
		return
	}
	cutoff := len(cs.persistedBlockHashes) - fuzzenv.FuzzPersistentRetainBlocks
	for _, old := range cs.persistedBlockHashes[:cutoff] {
		cs.persistentRepo.DeleteBlockByHash(cs.persistentRepo.Database(), types.OpaqueHash(old))
		cs.persistentRepo.DeleteHeaderTimeSlot(cs.persistentRepo.Database(), old)
	}
	cs.persistedBlockHashes = cs.persistedBlockHashes[cutoff:]
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

// GetGenesisBlockMaybe retrieves the genesis block if it exists.
// Unlike GetGenesisBlock, it does not panic when the genesis block is missing.
func (cs *ChainState) GetGenesisBlockMaybe() (*types.Block, error) {
	headerHash, err := cs.repo.GetCanonicalHash(cs.repo.Database(), 0)
	if err != nil {
		return nil, err
	}

	block, err := cs.repo.GetBlock(cs.repo.Database(), headerHash, 0)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (cs *ChainState) GetBlockByHash(headerHash types.HeaderHash) (types.Block, error) {
	timeSlot, err := cs.repo.GetHeaderTimeSlot(cs.repo.Database(), headerHash)
	if err != nil {
		timeSlot, err = cs.persistentRepo.GetHeaderTimeSlot(cs.persistentRepo.Database(), headerHash)
		if err != nil {
			return types.Block{}, err
		}
		block, err := cs.persistentRepo.GetBlockByHash(cs.persistentRepo.Database(), types.OpaqueHash(headerHash))
		if err != nil {
			return types.Block{}, err
		}
		return *block, nil
	}

	block, err := cs.repo.GetBlock(cs.repo.Database(), headerHash, timeSlot)
	if err != nil {
		block, err := cs.persistentRepo.GetBlockByHash(cs.persistentRepo.Database(), types.OpaqueHash(headerHash))
		if err != nil {
			return types.Block{}, err
		}
		return *block, nil
	}

	return *block, nil
}

/*
	State persistence
*/

// PersistStateForBlock persists the state for a given block and returns the computed state root.
func (cs *ChainState) PersistStateForBlock(blockHeaderHash types.HeaderHash, state types.State) (types.StateRoot, error) {
	serializedState, err := m.StateEncoder(state)
	if err != nil {
		return types.StateRoot{}, fmt.Errorf("failed to encode state: %w", err)
	}

	// Method A: serializedState already contains every storage / lookup-meta
	// entry from globalKV; no fallback pool to merge in.
	fullStateKeyVals := serializedState

	// Sort by Key to ensure consistent Merklization
	sort.Slice(fullStateKeyVals, func(i, j int) bool {
		return bytes.Compare(fullStateKeyVals[i].Key[:], fullStateKeyVals[j].Key[:]) < 0
	})

	stateRoot, err := cs.persistStateForBlockMerklize(fullStateKeyVals)
	if err != nil {
		return types.StateRoot{}, err
	}

	err = cs.repo.SaveStateRootByHeaderHash(cs.repo.Database(), blockHeaderHash, stateRoot)
	if err != nil {
		return types.StateRoot{}, fmt.Errorf("failed to store state root mapping: %w", err)
	}

	// Write state data to both memory (for fast reads) and disk (for persistence).
	err = cs.repo.SaveStateData(cs.repo.Database(), stateRoot, fullStateKeyVals)
	if err != nil {
		return types.StateRoot{}, fmt.Errorf("failed to store state data to memory: %w", err)
	}
	if err = cs.persistentRepo.SaveStateData(cs.persistentRepo.Database(), stateRoot, fullStateKeyVals); err != nil {
		logger.Warnf("PersistStateForBlock: failed to store state data to disk: %v", err)
	}

	// Evict oldest state data from memory when exceeding MaxLookupAge entries.
	cs.recentStateRoots = append(cs.recentStateRoots, stateRoot)
	if len(cs.recentStateRoots) > types.MaxLookupAge {
		evict := cs.recentStateRoots[0]
		cs.recentStateRoots = cs.recentStateRoots[1:]
		cs.repo.DeleteStateData(cs.repo.Database(), evict)
		if fuzzenv.Enabled() {
			cs.persistentRepo.DeleteStateData(cs.persistentRepo.Database(), evict)
			if err := cs.trieStore.DeleteTrie(types.OpaqueHash(evict)); err != nil {
				logger.Warnf("PersistStateForBlock: failed to delete evicted trie %x: %v", evict[:8], err)
			}
		}
	}

	return stateRoot, nil
}

// persistStateForBlockMerklize uses incremental merklize when prior state is available,
// falling back to full MerklizeAndCommit otherwise.
func (cs *ChainState) persistStateForBlockMerklize(fullStateKeyVals types.StateKeyVals) (types.StateRoot, error) {
	// Try incremental path
	if len(cs.recentStateRoots) > 0 {
		priorRoot := cs.recentStateRoots[len(cs.recentStateRoots)-1]
		priorKVs, err := cs.repo.GetStateData(cs.repo.Database(), priorRoot)
		if err == nil {
			trieExists, _ := cs.trieStore.TrieExists(types.OpaqueHash(priorRoot))
			if trieExists {
				dirtyEntries := store.DiffSortedKeyVals(priorKVs, fullStateKeyVals)
				for _, entry := range dirtyEntries {
					if entry.IsDelete {
						cs.keyLevelCache.Invalidate(entry.Key)
					}
				}
				root, err := cs.incrementalMerklizeAndCommit(priorRoot, dirtyEntries)
				if err == nil {
					return root, nil
				}
				logger.Warnf("PersistStateForBlock: incremental failed, fallback to full: %v", err)
			}
		}
	}

	// Fallback: full merklize
	return cs.trieStore.MerklizeAndCommit(fullStateKeyVals)
}

// incrementalMerklizeAndCommit runs incremental merklize with persistence callbacks,
// writes new nodes to DB via batch, then increments refcounts.
func (cs *ChainState) incrementalMerklizeAndCommit(
	priorRoot types.StateRoot,
	dirtyEntries []store.DirtyEntry,
) (types.StateRoot, error) {
	batch := cs.persistentRepo.Database().NewBatch()
	var newNodes []types.OpaqueHash

	storeNode := func(nodeHash types.OpaqueHash, node m.TrieNode) error {
		newNodes = append(newNodes, nodeHash)
		return batch.Put(makeTrieKey(store.TrieNodePrefix(), nodeHash[1:]), node[:])
	}

	storeValue := func(value []byte) error {
		valueHash := hash.Blake2bHash(value)
		return batch.Put(makeTrieKey(store.TrieNodeValuePrefix(), valueHash[:]), value)
	}

	incrRoot, err := store.IncrementalMerklize(
		types.OpaqueHash(priorRoot),
		dirtyEntries,
		cs.trieStore,
		storeNode, storeValue,
	)
	if err != nil {
		batch.Close()
		return types.StateRoot{}, err
	}

	if err := batch.Commit(); err != nil {
		batch.Close()
		return types.StateRoot{}, fmt.Errorf("incremental batch commit: %w", err)
	}
	batch.Close()

	for _, nodeHash := range newNodes {
		if err := cs.trieStore.IncreaseNodeRefCount(nodeHash); err != nil {
			return types.StateRoot{}, fmt.Errorf("incremental increase ref count: %w", err)
		}
	}

	return types.StateRoot(incrRoot), nil
}

func makeTrieKey(prefix byte, suffix []byte) []byte {
	key := make([]byte, 1+len(suffix))
	key[0] = prefix
	copy(key[1:], suffix)
	return key
}

// merklizeWithKeyCache computes state root using key-level cache.
// This optimization caches leaf hashes for individual keys, so unchanged keys
// don't need to recompute their leaf hashes during merklization.
// The cache callback is get-or-compute: on miss it computes once, stores, and returns
// the hash so merklization uses it without recomputing.
func (cs *ChainState) merklizeWithKeyCache(fullStateKeyVals types.StateKeyVals) (types.StateRoot, error) {
	cacheFn := func(key types.StateKey, value []byte) types.OpaqueHash {
		leafHash, valueHash, ok := cs.keyLevelCache.GetLeafHash(key, value)
		if ok {
			return leafHash
		}
		// Safety cap: clear entire cache if over limit (EpochLength * 50)
		if cs.keyLevelCache.Len() >= types.MaxKeyLevelCacheSize {
			cs.ClearKeyLevelCache()
		}
		leafHash = m.EncodeLeafNodeHash(key, value)
		cs.keyLevelCache.PutLeafHash(key, valueHash, leafHash)
		return leafHash
	}

	return m.MerklizationSerializedStateWithCache(fullStateKeyVals, cacheFn, nil, nil)
}

// ComputeStateRootWithCache computes the state root for given state key-values using the key-level cache.
// This is a public method that can be used by other packages (e.g., stf) to compute state roots with caching.
// The stateKeyVals should already be sorted by Key for consistent Merklization.
func (cs *ChainState) ComputeStateRootWithCache(stateKeyVals types.StateKeyVals) (types.StateRoot, error) {
	return cs.merklizeWithKeyCache(stateKeyVals)
}

// ClearKeyLevelCache clears the key-level merklization cache.
func (cs *ChainState) ClearKeyLevelCache() {
	if cs.keyLevelCache != nil {
		cs.keyLevelCache.Clear()
	}
}

// LastCommittedStateRoot returns the state root from the most recent StateCommit.
func (cs *ChainState) LastCommittedStateRoot() types.StateRoot {
	return cs.lastCommittedStateRoot
}

// GetStateByBlockHash retrieves state data for a given block from persistent database
func (cs *ChainState) GetStateByBlockHash(blockHeaderHash types.HeaderHash) (types.StateKeyVals, error) {
	stateKeyVals, err := cs.repo.GetStateDataByHeaderHash(cs.repo.Database(), blockHeaderHash)
	if err == nil {
		return stateKeyVals, nil
	}
	return cs.persistentRepo.GetStateDataByHeaderHash(cs.persistentRepo.Database(), blockHeaderHash)
}

func (cs *ChainState) GetStateRootByBlockHash(blockHeaderHash types.HeaderHash) (types.StateRoot, error) {
	stateRoot, err := cs.repo.GetStateRootByHeaderHash(cs.repo.Database(), blockHeaderHash)
	if err == nil {
		return stateRoot, nil
	}
	return cs.persistentRepo.GetStateRootByHeaderHash(cs.persistentRepo.Database(), blockHeaderHash)
}

func (cs *ChainState) GetHashSegmentMap() (map[types.OpaqueHash]types.OpaqueHash, error) {
	return cs.persistentRepo.GetHashSegmentMap(cs.persistentRepo.Database())
}

func (cs *ChainState) SetHashSegmentMap(hashSegmentMap map[string]string) error {
	return cs.persistentRepo.SetHashSegmentMap(cs.persistentRepo.Database(), hashSegmentMap)
}

func (cs *ChainState) SetHashSegmentMapWithLimit(wpHash, segmentRoot types.OpaqueHash) (map[types.OpaqueHash]types.OpaqueHash, error) {
	db := cs.persistentRepo.Database()
	return cs.persistentRepo.SetHashSegmentMapWithLimit(db, db, wpHash, segmentRoot)
}

func (cs *ChainState) GetSegmentErasureMap(segmentRoot types.OpaqueHash) (types.OpaqueHash, error) {
	return cs.persistentRepo.GetSegmentErasureMap(cs.persistentRepo.Database(), segmentRoot)
}

func (cs *ChainState) SetSegmentErasureMap(segmentRoot, erasureRoot types.OpaqueHash) error {
	return cs.persistentRepo.SetSegmentErasureMap(cs.persistentRepo.Database(), segmentRoot, erasureRoot, types.SegmentErasureTTL)
}

func (cs *ChainState) SaveBlockByHashToPersistent(hash types.OpaqueHash, block *types.Block) error {
	return cs.persistentRepo.SaveBlockByHash(cs.persistentRepo.Database(), hash, block)
}

func (cs *ChainState) persistBlockMapping(block types.Block) error {
	headerHash, err := hash.ComputeBlockHeaderHash(block.Header)
	if err != nil {
		return fmt.Errorf("failed to compute block header hash: %w", err)
	}

	if err := cs.SaveBlockByHashToPersistent(types.OpaqueHash(headerHash), &block); err != nil {
		return fmt.Errorf("failed to persist block: %w", err)
	}

	if err := cs.persistentRepo.SaveHeaderTimeSlot(cs.persistentRepo.Database(), headerHash, block.Header.Slot); err != nil {
		return fmt.Errorf("failed to persist header time slot: %w", err)
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

	// Restore state. Method A keeps storage/lookup-meta in globalKV, so
	// StateKeyValsToState no longer returns an unmatched-keyvals pool.
	state, err := m.StateKeyValsToState(stateKeyVals)
	if err != nil {
		return err
	}

	return cs.restoreWithState(blockHeaderHash, block, state)
}

// RestoreStateFromSnapshot restores block/ancestry management like
// RestoreBlockAndState, but uses the provided state instead of reading from DB.
func (cs *ChainState) RestoreStateFromSnapshot(
	blockHeaderHash types.HeaderHash,
	state types.State,
) error {
	block, err := cs.GetBlockByHash(blockHeaderHash)
	if err != nil {
		return fmt.Errorf("failed to get block for hash 0x%x: %w", blockHeaderHash[:8], err)
	}

	return cs.restoreWithState(blockHeaderHash, block, state)
}

func (cs *ChainState) restoreWithState(
	blockHeaderHash types.HeaderHash,
	block types.Block,
	state types.State,
) error {
	cs.GetPriorStates().SetState(state)
	// Keep only blocks up to the restored headerHash (fallback point)
	cs.unfinalizedBlocks.KeepBlocksUpTo(blockHeaderHash)
	// Add the restored block if it's not already in the list
	blocks := cs.GetBlocks()
	if len(blocks) == 0 {
		cs.AddBlock(block)
	} else {
		latestBlockHash, err := hash.ComputeBlockHeaderHash(blocks[len(blocks)-1].Header)
		if err != nil || latestBlockHash != blockHeaderHash {
			cs.AddBlock(block)
		}
	}

	// Keep only ancestry up to the restored headerHash (fallback point)
	cs.KeepAncestryUpTo(blockHeaderHash)

	// Reset recentStateRoots and lastCommittedStateRoot to the restored block's
	// state root. This ensures the next block's diff/trie operations use the
	// correct prior root instead of stale values from the old fork.
	restoredRoot, err := cs.repo.GetStateRootByHeaderHash(cs.repo.Database(), blockHeaderHash)
	if err != nil {
		// State root not in memory repo — reset to zero (fallback to full merklize).
		cs.recentStateRoots = nil
		cs.lastCommittedStateRoot = types.StateRoot{}
	} else {
		cs.recentStateRoots = []types.StateRoot{restoredRoot}
		cs.lastCommittedStateRoot = restoredRoot
	}

	// Clear verifier cache when restoring to a different state point
	// as the epoch may have changed
	ClearVerifierCache()

	return nil
}

func (cs *ChainState) BuildStateRootInputKeyValsAndRoot(
	stateKeyVals types.StateKeyVals,
) (merkleInputKeyVals types.StateKeyVals, stateRoot types.StateRoot, err error) {
	state, err := m.StateKeyValsToState(stateKeyVals)
	if err != nil {
		return nil, types.StateRoot{}, fmt.Errorf("StateKeyValsToState: %w", err)
	}

	serializedState, err := m.StateEncoder(state)
	if err != nil {
		return nil, types.StateRoot{}, fmt.Errorf("StateEncoder: %w", err)
	}

	// Method A: serializedState already contains every storage/lookup-meta
	// entry from globalKV, so there is no fallback pool to merge in.
	merkleInputKeyVals = serializedState
	stateRoot, err = cs.merklizeWithKeyCache(merkleInputKeyVals)
	if err != nil {
		return nil, types.StateRoot{}, fmt.Errorf("merklizeWithKeyCache: %w", err)
	}
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

	merkleInputKeyVals, stateRoot, err := cs.BuildStateRootInputKeyValsAndRoot(stateKeyVals)
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

	db := cs.persistentRepo.Database()
	if err := cs.persistentRepo.SaveStateData(db, genesisStateRoot, merkleInputKeyVals); err != nil {
		return types.HeaderHash{}, types.StateRoot{}, fmt.Errorf("store state_data: %w", err)
	}
	if err := cs.persistentRepo.SaveStateRootByHeaderHash(db, genesisBlockHash, genesisStateRoot); err != nil {
		return types.HeaderHash{}, types.StateRoot{}, fmt.Errorf("store state_root mapping: %w", err)
	}

	cs.lastCommittedStateRoot = genesisStateRoot

	return genesisBlockHash, genesisStateRoot, nil
}

// Compile-time interface check
var _ Blockchain = (*ChainState)(nil)
