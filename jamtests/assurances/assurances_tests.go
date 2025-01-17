package jamtests

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type AssuranceTestCase struct {
	Input     AssuranceInput  `json:"input"`
	PreState  AssuranceState  `json:"pre_state"`
	Output    AssuranceOutput `json:"output"`
	PostState AssuranceState  `json:"post_state"`
}

type AssuranceInput struct {
	Assurances types.AssurancesExtrinsic `json:"assurances,omitempty"`
	Slot       types.TimeSlot            `json:"slot"`
	Parent     types.HeaderHash          `json:"parent,omitempty"`
}

type AssuranceOutputData struct {
	//  Items removed from ρ† to get ρ'
	Reported []types.WorkReport `json:"report"`
}

type AssuranceOutput struct {
	Ok  *AssuranceOutputData `json:"ok,omitempty"`
	Err *AssuranceErrorCode  `json:"err,omitempty"`
}

type AssuranceState struct {
	Rho   types.AvailabilityAssignments `json:"rho"`
	Kappa types.ValidatorsData          `json:"kappa"`
}

/*
-- State transition function execution error.
-- Error codes **are not specified** in the the Graypaper.
-- Feel free to ignore the actual value.
*/
type AssuranceErrorCode types.ErrorCode

const (
	BadAttestationParent      AssuranceErrorCode = iota // 0
	BadValidatorIndex                                   // 1
	CoreNotEngaged                                      // 2
	BadSignature                                        // 3
	NotSortedOrUniqueAssurers                           // 4
)
