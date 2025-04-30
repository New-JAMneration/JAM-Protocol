package jamtests

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
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

func (a *AssuranceErrorCode) Error() string {
	if a == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *a)
}

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

// AssurancesInput UnmarshalJSON
func (a *AssuranceInput) UnmarshalJSON(data []byte) error {
	cLog(Cyan, "Unmarshalling AssurancesInput")
	type Alias AssuranceInput
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
func (a *AssuranceOutputData) UnmarshalJSON(data []byte) error {
	cLog(Cyan, "Unmarshalling AssurancesOutput")
	type Alias AssuranceOutputData
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
func (a *AssuranceInput) Decode(d *types.Decoder) error {
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
func (a *AssuranceOutputData) Decode(d *types.Decoder) error {
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
func (a *AssuranceOutput) Decode(d *types.Decoder) error {
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
			a.Ok = &AssuranceOutputData{}
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

		a.Err = (*AssuranceErrorCode)(&errByte)

		cLog(Yellow, fmt.Sprintf("SafroleErrorCode: %v", *a.Err))
	}

	return nil
}

// AssurancesState
func (a *AssuranceState) Decode(d *types.Decoder) error {
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
func (a *AssuranceTestCase) Decode(d *types.Decoder) error {
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

// Encode
type Encodable interface {
	Encode(e *types.Encoder) error
}

// AssurancesInput
func (a *AssuranceInput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AssurancesInput")
	var err error

	if err = a.Assurances.Encode(e); err != nil {
		return err
	}

	if err = a.Slot.Encode(e); err != nil {
		return err
	}

	if err = a.Parent.Encode(e); err != nil {
		return err
	}

	return nil
}

// AssurancesOutputData
func (a *AssuranceOutputData) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AssurancesOutputData")

	if err := e.EncodeLength(uint64(len(a.Reported))); err != nil {
		return err
	}

	for _, report := range a.Reported {
		if err := report.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AssurancesOutput
func (a *AssuranceOutput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AssurancesOutput")

	if a.Ok != nil && a.Err != nil {
		return errors.New("both Ok and Err are set")
	}

	if a.Ok == nil && a.Err == nil {
		return errors.New("neither Ok nor Err are set")
	}

	if a.Ok != nil {
		cLog(Yellow, "AssurancesOutput is ok")
		if err := e.WriteByte(0); err != nil {
			return err
		}

		// Encode AssuranceOutputData
		if err := a.Ok.Encode(e); err != nil {
			return err
		}

		return nil
	}

	if a.Err != nil {
		cLog(Yellow, "AssurancesOutput is err")
		if err := e.WriteByte(1); err != nil {
			return err
		}

		// Encode DisputeErrorCode
		if err := e.WriteByte(byte(*a.Err)); err != nil {
			return err
		}

		cLog(Yellow, fmt.Sprintf("AssuranceErrorCode: %d", *a.Err))

		return nil
	}

	return nil
}

// AssurancesState
func (a *AssuranceState) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AssurancesState")
	var err error

	if err = a.AvailAssignments.Encode(e); err != nil {
		return err
	}

	if err = a.CurrValidators.Encode(e); err != nil {
		return err
	}

	return nil
}

// AssurancesTestCase
func (a *AssuranceTestCase) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AssurancesTestCase")
	var err error

	if err = a.Input.Encode(e); err != nil {
		return err
	}

	if err = a.PreState.Encode(e); err != nil {
		return err
	}

	if err = a.Output.Encode(e); err != nil {
		return err
	}

	if err = a.PostState.Encode(e); err != nil {
		return err
	}

	return nil
}

func (a *AssuranceTestCase) Dump() error {
	s := store.GetInstance()

	// Add block
	header := types.Header{Slot: a.Input.Slot, Parent: a.Input.Parent}
	block := types.Block{Header: header}
	s.AddBlock(block)

	s.GetPosteriorStates().SetKappa(a.PreState.CurrValidators)
	s.GetIntermediateStates().SetRhoDagger(a.PreState.AvailAssignments)

	return nil
}

func (a *AssuranceTestCase) GetPostState() interface{} {
	return a.PostState
}

func (a *AssuranceTestCase) GetOutput() interface{} {
	return a.Output
}

func (a *AssuranceTestCase) ExpectError() error {
	if a.Output.Err == nil {
		return nil
	}
	return a.Output.Err
}

func (a *AssuranceTestCase) Validate() error {
	return nil
}
