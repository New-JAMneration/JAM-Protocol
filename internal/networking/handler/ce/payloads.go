package ce

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// NOTE: These payload types are referenced by CE request encoding helpers
// introduced in the "encode for ce request" change-set.
// They are kept small and local to the CE handler package.

type CE134Payload struct {
	CoreIndex           uint32
	HeaderHash          types.HeaderHash
	WorkPackage         *types.WorkPackage
	SegmentRootMappings []SegmentRootMapping // optional; nil encodes as len++ 0
}

type CE135Payload struct {
	Report    types.WorkReport
	Slot      types.TimeSlot
	Signatures []types.ValidatorSignature // ValidatorIndex ++ Ed25519Signature per entry
}

type CE136Payload struct {
	WorkReportHash types.WorkReportHash
}
