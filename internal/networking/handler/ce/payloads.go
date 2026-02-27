package ce

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// NOTE: These payload types are referenced by CE request encoding helpers
// introduced in the "encode for ce request" change-set.
// They are kept small and local to the CE handler package.

type CE134Payload struct {
	CoreIndex   uint32
	HeaderHash  types.HeaderHash
	WorkPackage *types.WorkPackage
}

type CE135Payload struct {
	CoreIndex uint32
	Report    types.WorkReport
	Signature types.Ed25519Signature
}

type CE136Payload struct {
	CoreIndex uint32
}
