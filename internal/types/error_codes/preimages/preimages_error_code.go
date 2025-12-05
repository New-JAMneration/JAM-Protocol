package types

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

const (
	PreimageUnneeded types.ErrorCode = iota // 0
	PrimagesNotSortedUnique
)

// This map provides human-readable messages following the fuzz-proto examples
var PreimagesErrorCodeMessages = map[types.ErrorCode]string{
	PreimageUnneeded:        "preimage not required", // matches fuzz-proto example
	PrimagesNotSortedUnique: "preimages not sorted and unique",
}
