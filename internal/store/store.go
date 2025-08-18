package store

import (
	"log"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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
	ancestorHeaders            *AncestorHeaders
	posteriorCurrentValidators *PosteriorCurrentValidators
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
			ancestorHeaders:            NewAncestorHeaders(),
			posteriorCurrentValidators: NewPosteriorValidators(),
		}
		log.Println("🚀 Store initialized")
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
		ancestorHeaders:            NewAncestorHeaders(),
		posteriorCurrentValidators: NewPosteriorValidators(),
	}
	log.Println("🚀 Store reset")
}

func (s *Store) AddBlock(block types.Block) {
	s.unfinalizedBlocks.AddBlock(block)
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
	posterState := s.GetPosteriorStates().GetState()
	s.GetPriorStates().SetState(posterState)

	// s.GetPosteriorStates().SetState(*NewPosteriorStates().state)
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
