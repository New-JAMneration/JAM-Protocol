package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type ProcessingBlock struct {
	mu    sync.RWMutex
	block *types.Block
}

func NewProcessingBlock() *ProcessingBlock {
	return &ProcessingBlock{
		block: &types.Block{},
	}
}

func (b *ProcessingBlock) SetBlock(block types.Block) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block = &block
}

func (b *ProcessingBlock) GetBlock() types.Block {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return *b.block
}

// Get Extrinsics from the block
func (b *ProcessingBlock) GetExtrinsics() types.Extrinsic {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Extrinsic
}

func (b *ProcessingBlock) GetTicketsExtrinsic() types.TicketsExtrinsic {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Extrinsic.Tickets
}

func (b *ProcessingBlock) SetTicketsExtrinsic(tickets types.TicketsExtrinsic) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Extrinsic.Tickets = tickets
}

func (b *ProcessingBlock) GetPreimagesExtrinsic() types.PreimagesExtrinsic {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Extrinsic.Preimages
}

func (b *ProcessingBlock) SetPreimagesExtrinsic(preimages types.PreimagesExtrinsic) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Extrinsic.Preimages = preimages
}

func (b *ProcessingBlock) GetGuaranteesExtrinsic() types.GuaranteesExtrinsic {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Extrinsic.Guarantees
}

func (b *ProcessingBlock) SetGuaranteesExtrinsic(guarantees types.GuaranteesExtrinsic) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Extrinsic.Guarantees = guarantees
}

func (b *ProcessingBlock) GetAssurancesExtrinsic() types.AssurancesExtrinsic {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Extrinsic.Assurances
}

func (b *ProcessingBlock) SetAssurancesExtrinsic(assurances types.AssurancesExtrinsic) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Extrinsic.Assurances = assurances
}

func (b *ProcessingBlock) GetDisputesExtrinsic() types.DisputesExtrinsic {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Extrinsic.Disputes
}

func (b *ProcessingBlock) SetDisputesExtrinsic(disputes types.DisputesExtrinsic) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Extrinsic.Disputes = disputes
}
