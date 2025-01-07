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

func (i *IntermediateHeader) SetSeal(seal types.BandersnatchVrfSignature) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Header.Seal = seal
}
