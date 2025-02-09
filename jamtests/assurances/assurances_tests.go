package jamtests

import (
	"encoding/json"
	"errors"

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
	Reported []types.WorkReport `json:"reported"`
}

type AssuranceOutput struct {
	Ok  *AssuranceOutputData `json:"ok,omitempty"`
	Err *AssuranceErrorCode  `json:"err,omitempty"`
}

type AssuranceState struct {
	AvailAssignments types.AvailabilityAssignments `json:"avail_assignments"`
	CurrValidators   types.ValidatorsData          `json:"curr_validators"`
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

var assuranceErrorMap = map[string]AssuranceErrorCode{
	"bad_attestation_parent":        BadAttestationParent,
	"bad_validator_index":           BadValidatorIndex,
	"core_not_engaged":              CoreNotEngaged,
	"bad_signature":                 BadSignature,
	"not_sorted_or_unique_assurers": NotSortedOrUniqueAssurers,
}

func (e *AssuranceErrorCode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if val, ok := assuranceErrorMap[str]; ok {
			*e = val
			return nil
		}
		return errors.New("invalid error code name: " + str)
	}

	return errors.New("invalid error code format, expected string")
}
