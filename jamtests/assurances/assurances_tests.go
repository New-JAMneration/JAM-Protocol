package jamtests

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ANSI color codes
var (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

var debugMode = false

// var debugMode = true

func cLog(color string, string string) {
	if debugMode {
		fmt.Printf("%s%s%s\n", color, string, Reset)
	}
}

type AssurancesTestCase struct {
	Input     AssurancesInput  `json:"input"`
	PreState  AssurancesState  `json:"pre_state"`
	Output    AssurancesOutput `json:"output"`
	PostState AssurancesState  `json:"post_state"`
}

type AssurancesInput struct {
	Assurances types.AssurancesExtrinsic `json:"assurances,omitempty"`
	Slot       types.TimeSlot            `json:"slot"`
	Parent     types.HeaderHash          `json:"parent,omitempty"`
}

type AssurancesOutputData struct {
	//  Items removed from ρ† to get ρ'
	Reported []types.WorkReport `json:"reported"`
}

type AssurancesOutput struct {
	Ok  *AssurancesOutputData `json:"ok,omitempty"`
	Err *AssurancesErrorCode  `json:"err,omitempty"`
}

type AssurancesState struct {
	AvailAssignments types.AvailabilityAssignments `json:"avail_assignments"`
	CurrValidators   types.ValidatorsData          `json:"curr_validators"`
}

/*
-- State transition function execution error.
-- Error codes **are not specified** in the the Graypaper.
-- Feel free to ignore the actual value.
*/
type AssurancesErrorCode types.ErrorCode

const (
	BadAttestationParent      AssurancesErrorCode = iota // 0
	BadValidatorIndex                                    // 1
	CoreNotEngaged                                       // 2
	BadSignature                                         // 3
	NotSortedOrUniqueAssurers                            // 4
)

var assurancesErrorMap = map[string]AssurancesErrorCode{
	"bad_attestation_parent":        BadAttestationParent,
	"bad_validator_index":           BadValidatorIndex,
	"core_not_engaged":              CoreNotEngaged,
	"bad_signature":                 BadSignature,
	"not_sorted_or_unique_assurers": NotSortedOrUniqueAssurers,
}

func (e *AssurancesErrorCode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if val, ok := assurancesErrorMap[str]; ok {
			*e = val
			return nil
		}
		return errors.New("invalid error code name: " + str)
	}

	return errors.New("invalid error code format, expected string")
}

// AssurancesInput UnmarshalJSON
func (a *AssurancesInput) UnmarshalJSON(data []byte) error {
	cLog(Cyan, "Unmarshalling AssurancesInput")
	type Alias AssurancesInput
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(a.Assurances) == 0 {
		a.Assurances = nil
	}

	return nil
}

// AssurancesOutputData UnmarshalJSON
func (a *AssurancesOutputData) UnmarshalJSON(data []byte) error {
	cLog(Cyan, "Unmarshalling AssurancesOutput")
	type Alias AssurancesOutputData
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(a.Reported) == 0 {
		a.Reported = nil
	}

	return nil
}

// AssurancesInput
func (a *AssurancesInput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AssurancesInput")
	var err error

	if err = a.Assurances.Decode(d); err != nil {
		return err
	}

	if err = a.Slot.Decode(d); err != nil {
		return err
	}

	if err = a.Parent.Decode(d); err != nil {
		return err
	}

	return nil
}

// AssurancesOutputData
func (a *AssurancesOutputData) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AssurancesOutputData")

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	reported := make([]types.WorkReport, length)
	for i := uint64(0); i < length; i++ {
		if err = reported[i].Decode(d); err != nil {
			return err
		}

		a.Reported = append(a.Reported, reported[i])
	}

	return nil
}

// AssurancesOutput
func (a *AssurancesOutput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AssurancesOutput")
	var err error

	okOrErr, err := d.ReadPointerFlag()
	if err != nil {
		return err
	}

	isOk := okOrErr == 0
	if isOk {
		cLog(Yellow, "AssurancesOutput is ok")

		if a.Ok == nil {
			a.Ok = &AssurancesOutputData{}
		}

		if err = a.Ok.Decode(d); err != nil {
			return err
		}

		return nil
	} else {
		cLog(Yellow, "AssurancesOutput is err")
		cLog(Yellow, "Decoding AssurancesErrorCode")

		// Read a byte as error code
		errByte, err := d.ReadErrorByte()
		if err != nil {
			return err
		}

		a.Err = (*AssurancesErrorCode)(&errByte)

		cLog(Yellow, fmt.Sprintf("SafroleErrorCode: %v", *a.Err))
	}

	return nil
}

// AssurancesState
func (a *AssurancesState) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AssurancesState")
	var err error

	if err = a.AvailAssignments.Decode(d); err != nil {
		return err
	}

	if err = a.CurrValidators.Decode(d); err != nil {
		return err
	}

	return nil
}

// AssurancesTestCase
func (a *AssurancesTestCase) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AssurancesTestCase")
	var err error

	if err = a.Input.Decode(d); err != nil {
		return err
	}

	if err = a.PreState.Decode(d); err != nil {
		return err
	}

	if err = a.Output.Decode(d); err != nil {
		return err
	}

	if err = a.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}
