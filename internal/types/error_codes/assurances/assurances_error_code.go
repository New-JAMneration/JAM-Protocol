package types

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

const (
	BadAttestationParent      types.ErrorCode = iota // 0
	BadValidatorIndex                                // 1
	CoreNotEngaged                                   // 2
	BadSignature                                     // 3
	NotSortedOrUniqueAssurers                        // 4
)

// This map provides human-readable messages following the fuzz-proto examples
var AssurancesErrorCodeMessages = map[types.ErrorCode]string{
	BadAttestationParent:      "bad attestation parent",
	BadValidatorIndex:         "bad attestation validator index", // matches fuzz-proto example
	CoreNotEngaged:            "core not engaged",
	BadSignature:              "bad attestation signature",
	NotSortedOrUniqueAssurers: "not sorted or unique assurers",
}
