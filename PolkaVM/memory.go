package PolkaVM

type MemoryAccess int

type Page struct {
	Value  []byte // 4096 byte (ZP) page data
	Access MemoryAccess
}

type Memory struct {
	Pages map[uint32]*Page // Key: Page Number, Value: Page Data
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
