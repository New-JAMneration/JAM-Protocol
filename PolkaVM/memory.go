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
