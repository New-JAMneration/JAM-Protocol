package store

import (
	"fmt"
	"log"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/repository"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

var (
	initOnce    sync.Once
	globalStore *Store
)

// Store represents a thread-safe global state container
type Store struct {
	repo *repository.Repository

	// INFO: Add more fields here
	unfinalizedBlocks          *UnfinalizedBlocks
	finalizedIndex             map[types.HeaderHash]bool
	processingBlock            *ProcessingBlock
	priorStates                *PriorStates
	intermediateStates         *IntermediateStates
	posteriorStates            *PosteriorStates
	ancestorHeaders            *AncestorHeaders
	posteriorCurrentValidators *PosteriorCurrentValidators
	storageKeyVals             types.StateKeyVals
}

// GetInstance returns the singleton instance of Store.
// If the instance doesn't exist, it creates one.
func GetInstance() *Store {
	initOnce.Do(func() {
		db := memory.NewDatabase()
		repo := repository.NewRepository(db)
		globalStore = &Store{
			repo:                       repo,
			unfinalizedBlocks:          NewUnfinalizedBlocks(),
			finalizedIndex:             make(map[types.HeaderHash]bool),
			processingBlock:            NewProcessingBlock(),
			priorStates:                NewPriorStates(),
			intermediateStates:         NewIntermediateStates(),
			posteriorStates:            NewPosteriorStates(),
			ancestorHeaders:            NewAncestorHeaders(),
			posteriorCurrentValidators: NewPosteriorValidators(),
			storageKeyVals:             types.StateKeyVals{},
		}
		log.Println("🚀 Store initialized")
	})
	return globalStore
}

func ResetInstance() {
	// reset globalStore
	db := memory.NewDatabase()
	repo := repository.NewRepository(db)
	globalStore = &Store{
		repo:                       repo,
		unfinalizedBlocks:          NewUnfinalizedBlocks(),
		finalizedIndex:             make(map[types.HeaderHash]bool),
		processingBlock:            NewProcessingBlock(),
		priorStates:                NewPriorStates(),
		intermediateStates:         NewIntermediateStates(),
		posteriorStates:            NewPosteriorStates(),
		ancestorHeaders:            NewAncestorHeaders(),
		posteriorCurrentValidators: NewPosteriorValidators(),
		storageKeyVals:             types.StateKeyVals{},
	}
	log.Println("🚀 Store reset")
}

func (s *Store) AddBlock(block types.Block) {
	s.unfinalizedBlocks.AddBlock(block)
	if err := s.repo.SaveBlock(s.repo.Database(), &block); err != nil {
		log.Printf("AddBlock: failed to index block: %v", err)
	}
}

func (s *Store) CleanupBlock() {
	s.unfinalizedBlocks = NewUnfinalizedBlocks()
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

func (s *Store) GenerateGenesisBlock(block types.Block) error {
	s.unfinalizedBlocks.GenerateGenesisBlock(block)
	// Genesis block is always finalized
	s.finalizedIndex[block.Header.Parent] = true
	if err := s.repo.SaveBlock(s.repo.Database(), &block); err != nil {
		log.Printf("GenerateGenesisBlock: failed to index block: %v", err)
		return err
	}
	return nil
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
	log.Println("🚀 Genesis state generated")
}

// AncestorHeaders

func (s *Store) AddAncestorHeader(header types.Header) {
	s.ancestorHeaders.AddHeader(header)
}

func (s *Store) GetAncestorHeaders() []types.Header {
	return s.ancestorHeaders.GetHeaders()
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
	blocks := s.GetBlocks()
	if len(blocks) == 0 {
		posterState := s.GetPosteriorStates().GetState()
		s.GetPriorStates().SetState(posterState)
		s.GetPosteriorStates().SetState(*NewPosteriorStates().state)
		return
	}

	latestBlock := s.GetLatestBlock()

	blockHeaderHash, err := computeBlockHeaderHash(latestBlock.Header)
	if err != nil {
		log.Printf("StateCommit: failed to encode header: %v", err)
	} else {
		posteriorState := s.GetPosteriorStates().GetState()

		// Persist state for block
		err = s.PersistStateForBlock(blockHeaderHash, posteriorState)
		if err != nil {
			log.Printf("StateCommit: failed to persist state: %v", err)
		} else {
			log.Printf("StateCommit: persisted state for block 0x%x", blockHeaderHash[:8])
		}

		// Persist block mapping
		err = s.repo.SaveBlock(s.repo.Database(), &latestBlock)
		if err != nil {
			log.Printf("StateCommit: failed to persist block: %v", err)
		} else {
			log.Printf("StateCommit: persisted block 0x%x", blockHeaderHash[:8])
		}
	}

	posterState := s.GetPosteriorStates().GetState()
	s.GetPriorStates().SetState(posterState)
	s.GetPosteriorStates().SetState(*NewPosteriorStates().state)
}

// PersistStateForBlock persists the state for a given block to Redis
func (s *Store) PersistStateForBlock(blockHeaderHash types.HeaderHash, state types.State) error {
	serializedState, err := m.StateEncoder(state)
	if err != nil {
		return fmt.Errorf("failed to encode state: %w", err)
	}

	storageKeyVals := s.GetStorageKeyVals()
	fullStateKeyVals := append(storageKeyVals, serializedState...)

	stateRoot := m.MerklizationSerializedState(fullStateKeyVals)

	err = s.repo.SaveStateRootByHeaderHash(s.repo.Database(), blockHeaderHash, stateRoot)
	if err != nil {
		return fmt.Errorf("failed to store state root mapping: %w", err)
	}

	err = s.repo.SaveStateData(s.repo.Database(), stateRoot, fullStateKeyVals)
	if err != nil {
		return fmt.Errorf("failed to store state data: %w", err)
	}

	return nil
}

// GetStateByBlockHash retrieves state data for a given block from Redis
func (s *Store) GetStateByBlockHash(blockHeaderHash types.HeaderHash) (types.StateKeyVals, error) {
	return s.repo.GetStateDataByHeaderHash(s.repo.Database(), blockHeaderHash)
}

// StorageKeyVals
func (s *Store) GetStorageKeyVals() types.StateKeyVals {
	return s.storageKeyVals
}

func (s *Store) SetStorageKeyVals(storageKeyVals types.StateKeyVals) {
	s.storageKeyVals = storageKeyVals
}

// GetGenesisBlock retrieves the genesis block from the store.
// This function panics if the genesis block does not exist, as it is assumed to always be present.
// TODO: This is because respecting the current implementation usage of `GenerateGenesisBlock`, but later should ensure genesis block exist at store initialization.
func (s *Store) GetGenesisBlock() types.Block {
	headerHash, err := s.repo.GetCanonicalHash(s.repo.Database(), 0)
	if err != nil {
		// This is coding error, genesis block must exist.
		// Store instance without genesis block should never happen.
		panic(fmt.Sprintf("genesis block must exist, failed to retrieve genesis block hash: %v", err))
	}

	block, err := s.repo.GetBlock(s.repo.Database(), headerHash, 0)
	if err != nil {
		// This is coding error, genesis block must exist.
		// Store instance without genesis block should never happen.
		panic(fmt.Sprintf("genesis block must exist, failed to retrieve genesis block: %v", err))
	}
	return *block
}

func (s *Store) GetBlockByHash(headerHash types.HeaderHash) (types.Block, error) {
	timeSlot, err := s.repo.GetHeaderTimeSlot(s.repo.Database(), headerHash)
	if err != nil {
		return types.Block{}, err
	}

	block, err := s.repo.GetBlock(s.repo.Database(), headerHash, timeSlot)
	if err != nil {
		return types.Block{}, err
	}

	return *block, nil
}

func computeBlockHeaderHash(header types.Header) (types.HeaderHash, error) {
	encoder := types.NewEncoder()
	encodedHeader, err := encoder.Encode(&header)
	if err != nil {
		return types.HeaderHash{}, err
	}
	return types.HeaderHash(hash.Blake2bHash(encodedHeader)), nil
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
	state, storageKeyVal, err := merklization.StateKeyValsToState(stateKeyVals)
	if err != nil {
		return err
	}

	s.GetPosteriorStates().SetState(state)
	s.SetStorageKeyVals(storageKeyVal)

	// Restore block
	s.CleanupBlock()
	s.AddBlock(block)

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
