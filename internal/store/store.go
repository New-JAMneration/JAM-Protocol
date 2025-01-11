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
	mu sync.RWMutex

	// INFO: Add more fields here
	blocks                     *Blocks
	priorStates                *PriorStates
	intermediateStates         *IntermediateStates
	posteriorStates            *PosteriorStates
	ancestorHeaders            *AncestorHeaders
	intermediateHeader         *IntermediateHeader
	posteriorCurrentValidators *PosteriorCurrentValidators
	beefyCommitmentOutput      *BeefyCommitmentOutputs // This is tmp used waiting for more def in GP
}

// GetInstance returns the singleton instance of Store.
// If the instance doesn't exist, it creates one.
func GetInstance() *Store {
	initOnce.Do(func() {
		globalStore = &Store{
			blocks:                     NewBlocks(),
			priorStates:                NewPriorStates(),
			intermediateStates:         NewIntermediateStates(),
			posteriorStates:            NewPosteriorStates(),
			ancestorHeaders:            NewAncestorHeaders(),
			intermediateHeader:         NewIntermediateHeader(),
			posteriorCurrentValidators: NewPosteriorValidators(),
			beefyCommitmentOutput:      NewBeefyCommitmentOutput(), // This is tmp used waiting for more def in GP
		}
		log.Println("🚀 Store initialized")
	})
	return globalStore
}

func (s *Store) AddBlock(block types.Block) {
	s.blocks.AddBlock(block)
}

func (s *Store) GetBlocks() []types.Block {
	return s.blocks.GetAllAncientBlocks()
}

func (s *Store) GetBlock() types.Block {
	return s.blocks.GetLatestBlock()
}

func (s *Store) GetLatestBlock() types.Block {
	return s.blocks.GetLatestBlock()
}

func (s *Store) GenerateGenesisBlock(block types.Block) {
	s.blocks.GenerateGenesisBlock(block)
	s.blocks.AddBlock(block)
	log.Println("🚀 Genesis block generated")
}

// Get
func (s *Store) GetPriorState() types.State {
	return s.priorStates.GetState()
}

// Set
func (s *Store) GetPriorStates() *PriorStates {
	return s.priorStates
}

// // Get
// func (s *Store) GetIntermediateState() types.State {
// 	return s.intermediateStates.GetState()
// }

// Set
func (s *Store) GetIntermediateStates() *IntermediateStates {
	return s.intermediateStates
}

// Get
func (s *Store) GetPosteriorState() types.State {
	return s.posteriorStates.GetState()
}

// Set
func (s *Store) GetPosteriorStates() *PosteriorStates {
	return s.posteriorStates
}

func (s *Store) GenerateGenesisState(state types.State) {
	s.priorStates.GenerateGenesisState(state)
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

// IntermediateHeader

func (s *Store) AddIntermediateHeader(header types.Header) {
	s.intermediateHeader.AddHeader(header)
}

func (s *Store) GetIntermediateHeader() types.Header {
	return s.intermediateHeader.GetHeader()
}

func (s *Store) ResetIntermediateHeader() {
	s.intermediateHeader.ResetHeader()
}

// BeefyCommitmentOutput (This is tmp used waiting for more def in GP)

// Get
func (s *Store) GetBeefyCommitmentOutput() types.BeefyCommitmentOutput {
	return s.beefyCommitmentOutput.GetBeefyCommitmentOutput()
}

// Set
func (s *Store) GetBeefyCommitmentOutputs() *BeefyCommitmentOutputs {
	return s.beefyCommitmentOutput
}
