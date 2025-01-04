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
	blocks *Blocks
	states *States
}

// GetInstance returns the singleton instance of Store.
// If the instance doesn't exist, it creates one.
func GetInstance() *Store {
	initOnce.Do(func() {
		globalStore = &Store{
			blocks: NewBlocks(),
			states: NewStates(),
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
