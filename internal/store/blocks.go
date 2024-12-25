package store

import (
	"sync"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

type Blocks struct {
	mu     sync.RWMutex
	blocks []jamTypes.Block
}

func NewBlocks() *Blocks {
	return &Blocks{
		blocks: make([]jamTypes.Block, 0),
	}
}

func (b *Blocks) AddBlock(block jamTypes.Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blocks = append(b.blocks, block)
}

func (b *Blocks) GetBlocks() []jamTypes.Block {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.blocks
}

func (b *Blocks) GenerateGenesisBlock(block jamTypes.Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blocks = append(b.blocks, block)
}
