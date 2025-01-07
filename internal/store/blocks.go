package store

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type Blocks struct {
	mu     sync.RWMutex
	blocks []types.Block
}

func NewBlocks() *Blocks {
	return &Blocks{
		blocks: make([]types.Block, 0),
	}
}

func (b *Blocks) AddBlock(block types.Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blocks = append(b.blocks, block)
}

func (b *Blocks) GetBlocks() []types.Block {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.blocks
}

func (b *Blocks) GetLatestBlock() types.Block {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.blocks[len(b.blocks)-1]
}

func (b *Blocks) GenerateGenesisBlock(block types.Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blocks = append(b.blocks, block)
}
