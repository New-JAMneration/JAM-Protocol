package store

import (
	"log"
	"sync"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
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
	states                     *States
	ancestorHeaders            *AncestorHeaders
	intermediateHeader         *IntermediateHeader
	posteriorCurrentValidators *PosteriorCurrentValidators
}

// GetInstance returns the singleton instance of Store.
// If the instance doesn't exist, it creates one.
func GetInstance() *Store {
	initOnce.Do(func() {
		globalStore = &Store{
			blocks:                     NewBlocks(),
			states:                     NewStates(),
			ancestorHeaders:            NewAncestorHeaders(),
			intermediateHeader:         NewIntermediateHeader(),
			posteriorCurrentValidators: NewPosteriorValidators(),
		}
		log.Println("🚀 Store initialized")
	})
	return globalStore
}

func (s *Store) AddBlock(block jamTypes.Block) {
	s.blocks.AddBlock(block)
}

func (s *Store) GetBlocks() []jamTypes.Block {
	return s.blocks.GetBlocks()
}

func (s *Store) GenerateGenesisBlock(block jamTypes.Block) {
	s.blocks.GenerateGenesisBlock(block)
	s.blocks.AddBlock(block)
	log.Println("🚀 Genesis block generated")
}

func (s *Store) GetState() jamTypes.State {
	return s.states.GetState()
}

func (s *Store) GetStates() States {
	return *s.states
}

func (s *Store) GenerateGenesisState(state jamTypes.State) {
	s.states.GenerateGenesisState(state)
	log.Println("🚀 Genesis state generated")
}

// AncestorHeaders

func (s *Store) AddAncestorHeader(header jamTypes.Header) {
	s.ancestorHeaders.AddHeader(header)
}

func (s *Store) GetAncestorHeaders() []jamTypes.Header {
	return s.ancestorHeaders.GetHeaders()
}

// PosteriorCurrentValidators

func (s *Store) AddPosteriorCurrentValidator(validator jamTypes.Validator) {
	s.posteriorCurrentValidators.AddValidator(validator)
}

func (s *Store) GetPosteriorCurrentValidators() jamTypes.ValidatorsData {
	return s.posteriorCurrentValidators.GetValidators()
}

func (s *Store) GetPosteriorCurrentValidatorByIndex(index jamTypes.ValidatorIndex) jamTypes.Validator {
	return s.posteriorCurrentValidators.GetValidatorByIndex(index)
}

// IntermediateHeader

func (s *Store) AddIntermediateHeader(header jamTypes.Header) {
	s.intermediateHeader.AddHeader(header)
}

func (s *Store) GetIntermediateHeader() jamTypes.Header {
	return s.intermediateHeader.GetHeader()
}

func (s *Store) ResetIntermediateHeader() {
	s.intermediateHeader.ResetHeader()
}
