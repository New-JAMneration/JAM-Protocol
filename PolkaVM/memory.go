package polkavm

import "github.com/New-JAMneration/JAM-Protocol/internal/input/jam_types"

type MemoryAccess int

type Memory struct {
	Segments MemorySegment
}

type MemorySegment struct {
	Address  jam_types.U32          // Starting Address
	Databyte jam_types.ByteSequence // Data Content
	Access   MemoryAccess           // Access Permissions
}

type MemorySegmentStartEnd struct {
	StartAddress jam_types.U32          // Starting Address
	EndAddress   jam_types.U32          // Ending Address
	Databyte     jam_types.ByteSequence // Data Content
}

const (
	MemoryInaccessible MemoryAccess = iota // ∅ Inaccessible
	MemoryReadOnly                         // R Read only
	MemoryReadWrite                        // W Read + Write
)
