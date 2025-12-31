package store

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
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

// KeepBlocksUpTo keeps only blocks up to and including the specified headerHash.
// If the headerHash is not found, it clears all blocks.
func (b *UnfinalizedBlocks) KeepBlocksUpTo(headerHash types.HeaderHash) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Find the index of the headerHash in blocks
	foundIdx := -1
	for i := len(b.blocks) - 1; i >= 0; i-- {
		blockHeaderHash, err := hash.ComputeBlockHeaderHash(b.blocks[i].Header)
		if err != nil {
			continue
		}
		if blockHeaderHash == headerHash {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		// HeaderHash not found, clear all blocks
		b.blocks = make([]types.Block, 0)
		return
	}

	// Keep only blocks up to and including the found index
	b.blocks = b.blocks[:foundIdx+1]
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
