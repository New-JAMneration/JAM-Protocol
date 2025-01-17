package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type IntermediateHeader struct {
	mu     sync.RWMutex
	Header types.Header
}

func NewIntermediateHeader() *IntermediateHeader {
	return &IntermediateHeader{
		Header: types.Header{},
	}
}

// AddHeader adds a header to the intermediateHeaders
func (i *IntermediateHeader) AddHeader(header types.Header) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header = header
}

// GetHeader returns the intermediateHeader
func (i *IntermediateHeader) GetHeader() types.Header {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.Header
}

// ResetHeader resets the intermediateHeader
func (i *IntermediateHeader) ResetHeader() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header = types.Header{}
}

func (i *IntermediateHeader) GetTicketsMark() *types.TicketsMark {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.Header.TicketsMark
}

func (i *IntermediateHeader) SetTicketsMark(ticketsMark *types.TicketsMark) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header.TicketsMark = ticketsMark
}

func (i *IntermediateHeader) SetSeal(seal types.BandersnatchVrfSignature) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header.Seal = seal
}

func (i *IntermediateHeader) SetEntropySource(entropy types.BandersnatchVrfSignature) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header.EntropySource = entropy
}

func (i *IntermediateHeader) SetAuthorIndex(index types.ValidatorIndex) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header.AuthorIndex = index
}

func (i *IntermediateHeader) SetParent(parent types.HeaderHash) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header.Parent = parent
}

func (i *IntermediateHeader) SetParentStateRoot(parent_state_root types.StateRoot) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header.ParentStateRoot = parent_state_root
}

func (i *IntermediateHeader) SetExtrinsicHash(extrinsic_hash types.OpaqueHash) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header.ExtrinsicHash = extrinsic_hash
}

func (i *IntermediateHeader) SetSlot(timeslot types.TimeSlot) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header.Slot = timeslot
}

func (i *IntermediateHeader) SetEpochMark(epoch_mark types.EpochMark) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Assign the entire EpochMark struct to the pointer
	i.Header.EpochMark = &epoch_mark
}

func (i *IntermediateHeader) SetHeader(header types.Header) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Assign the entire EpochMark struct to the pointer
	i.Header = header
}
