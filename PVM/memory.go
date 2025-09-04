package PVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type MemoryAccess int

type Page struct {
	Value  []byte // 4096 byte (ZP) page data
	Access MemoryAccess
}

type Memory struct {
	Pages       map[uint32]*Page // Key: Page Number, Value: Page Data
	heapPointer uint64
}

const (
	MemoryInaccessible MemoryAccess = iota // ∅ Inaccessible
	MemoryReadOnly                         // R Read only
	MemoryReadWrite                        // W Read + Write
)

func (m *Memory) GetPageAccess(index uint32) MemoryAccess {
	page, found := m.Pages[index]
	if !found {
		return MemoryInaccessible
	}

	return page.Access
}

// Read reads data from memory, might cross many pages
func (m *Memory) Read(start uint64, offset uint64) types.ByteSequence {
	buffer := make([]byte, offset)

	pageNumber := uint32(start / ZP)
	pageIndex := start % ZP

	for copied := uint64(0); copied < offset; {
		copied += uint64(copy(buffer[copied:], m.Pages[pageNumber].Value[pageIndex:]))
		pageNumber++
		pageIndex = 0
	}

	return buffer
}

func (m *Memory) Write(start uint64, offset uint64, data types.ByteSequence) {
	pageNumber := uint32(start / ZP)
	pageIndex := start % ZP

	for copied := uint64(0); copied < offset; {
		copied += uint64(copy(m.Pages[pageNumber].Value[pageIndex:], data[copied:]))
		pageNumber++
		pageIndex = 0
	}
}
