package PolkaVM

type MemoryAccess int

type Memory struct {
	Segments MemorySegment
}

type MemorySegment struct {
	Address  uint32       // Starting Address
	Databyte []byte       // Data Content
	Access   MemoryAccess // Access Permissions
}

type MemorySegmentStartEnd struct {
	StartAddress uint32 // Starting Address
	EndAddress   uint32 // Ending Address
	Databyte     []byte // Data Content
}

const (
	MemoryInaccessible MemoryAccess = iota // ∅ Inaccessible
	MemoryReadOnly                         // R Read only
	MemoryReadWrite                        // W Read + Write
)

const ZP = 4096 // PVM memory page size.
