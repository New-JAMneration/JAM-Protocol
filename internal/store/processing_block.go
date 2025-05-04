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

func (b *ProcessingBlock) SetHeader(header types.Header) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header = header
}

func (b *ProcessingBlock) GetHeader() types.Header {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header
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

func (b *ProcessingBlock) GetParent() types.HeaderHash {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.Parent
}

func (b *ProcessingBlock) SetParent(parent types.HeaderHash) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.Parent = parent
}

func (b *ProcessingBlock) GetParentStateRoot() types.StateRoot {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.ParentStateRoot
}

func (b *ProcessingBlock) SetParentStateRoot(parentStateRoot types.StateRoot) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.ParentStateRoot = parentStateRoot
}

func (b *ProcessingBlock) GetExtrinsicHash() types.OpaqueHash {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.ExtrinsicHash
}

func (b *ProcessingBlock) SetExtrinsicHash(extrinsicHash types.OpaqueHash) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.ExtrinsicHash = extrinsicHash
}

func (b *ProcessingBlock) GetSlot() types.TimeSlot {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.Slot
}

func (b *ProcessingBlock) SetSlot(slot types.TimeSlot) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.Slot = slot
}

func (b *ProcessingBlock) GetEpochMark() *types.EpochMark {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.EpochMark
}

func (b *ProcessingBlock) SetEpochMark(epochMark *types.EpochMark) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.EpochMark = epochMark
}

func (b *ProcessingBlock) GetTicketsMark() *types.TicketsMark {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.TicketsMark
}

func (b *ProcessingBlock) SetTicketsMark(ticketsMark *types.TicketsMark) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.TicketsMark = ticketsMark
}

func (b *ProcessingBlock) GetOffendersMark() types.OffendersMark {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.OffendersMark
}

func (b *ProcessingBlock) SetOffendersMark(offendersMark types.OffendersMark) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.OffendersMark = offendersMark
}

func (b *ProcessingBlock) GetAuthorIndex() types.ValidatorIndex {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.AuthorIndex
}

func (b *ProcessingBlock) SetAuthorIndex(authorIndex types.ValidatorIndex) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.AuthorIndex = authorIndex
}

func (b *ProcessingBlock) GetEntropySource() types.BandersnatchVrfSignature {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.EntropySource
}

func (b *ProcessingBlock) SetEntropySource(entropySource types.BandersnatchVrfSignature) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.EntropySource = entropySource
}

func (b *ProcessingBlock) GetSeal() types.BandersnatchVrfSignature {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.block.Header.Seal
}

func (b *ProcessingBlock) SetSeal(seal types.BandersnatchVrfSignature) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.block.Header.Seal = seal
}
