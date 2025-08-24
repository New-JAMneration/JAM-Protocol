package blockchain

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type UnfinalizedBlocks struct {
	mu     sync.RWMutex
	blocks []types.Block
}

// New one empty slice for blocks
func NewUnfinalizedBlocks() *UnfinalizedBlocks {
	return &UnfinalizedBlocks{
		blocks: make([]types.Block, 0),
	}
}

// Add block to ancient blocks for storage
func (b *UnfinalizedBlocks) AddBlock(block types.Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blocks = append(b.blocks, block)
}

// Get all ancient blocks
func (b *UnfinalizedBlocks) GetAllAncientBlocks() []types.Block {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.blocks
}

// Get latest block
func (b *UnfinalizedBlocks) GetLatestBlock() types.Block {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.blocks[len(b.blocks)-1]
}

func (b *UnfinalizedBlocks) GenerateGenesisBlock(block types.Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blocks = append(b.blocks, block)
}

// Set Header to latest block
func (b *UnfinalizedBlocks) SetHeader(header types.Header) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blocks[len(b.blocks)-1].Header = header
}

// Set Extrinsic to latest block
func (b *UnfinalizedBlocks) SetExtrinsic(extrinsic types.Extrinsic) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blocks[len(b.blocks)-1].Extrinsic = extrinsic
}
