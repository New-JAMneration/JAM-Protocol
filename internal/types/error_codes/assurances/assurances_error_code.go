package types

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

const (
	BadAttestationParent      types.ErrorCode = iota // 0
	BadValidatorIndex                                // 1
	CoreNotEngaged                                   // 2
	BadSignature                                     // 3
	NotSortedOrUniqueAssurers                        // 4
)
