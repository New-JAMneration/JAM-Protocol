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

type DisputesTestCase struct {
	Input     DisputesInput  `json:"input"`
	PreState  DisputesState  `json:"pre_state"`
	Output    DisputesOutput `json:"output"`
	PostState DisputesState  `json:"post_state"`
}

type DisputesInput struct {
	Disputes types.DisputesExtrinsic `json:"disputes"`
}

type DisputesOutputData struct {
	OffendersMark types.OffendersMark `json:"offenders_mark"`
}

type DisputesOutput struct {
	Ok  *DisputesOutputData `json:"ok,omitempty"`
	Err *DisputesErrorCode  `json:"err,omitempty"`
}

type DisputesState struct {
	Psi    types.DisputesRecords         `json:"psi"`
	Rho    types.AvailabilityAssignments `json:"rho"`
	Tau    types.TimeSlot                `json:"tau"`
	Kappa  types.ValidatorsData          `json:"kappa"`
	Lambda types.ValidatorsData          `json:"lambda"`
}

type DisputesErrorCode types.ErrorCode

const (
	AlreadyJudged             DisputesErrorCode = iota // 0
	BadVoteSplit                                       // 1
	VerdictsNotSortedUnique                            // 2
	JudgementsNotSortedUnique                          // 3
	CulpritsNotSortedUnique                            // 4
	FaultsNotSortedUnique                              // 5
	NotEnoughCulprits                                  // 6
	NotEnoughFaults                                    // 7
	CulpritsVerdictNotBad                              // 8
	FaultVerdictWrong                                  // 9
	OffenderAlreadyReported                            // 10
	BadJudgementAge                                    // 11
	BadValidatorIndex                                  // 12
	BadSignature                                       // 13
)

var disputesErrorMap = map[string]DisputesErrorCode{
	"already_judged":               AlreadyJudged,
	"bad_vote_split":               BadVoteSplit,
	"verdicts_not_sorted_unique":   VerdictsNotSortedUnique,
	"judgements_not_sorted_unique": JudgementsNotSortedUnique,
	"culprits_not_sorted_unique":   CulpritsNotSortedUnique,
	"faults_not_sorted_unique":     FaultsNotSortedUnique,
	"not_enough_culprits":          NotEnoughCulprits,
	"not_enough_faults":            NotEnoughFaults,
	"culprits_verdict_not_bad":     CulpritsVerdictNotBad,
	"fault_verdict_wrong":          FaultVerdictWrong,
	"offender_already_reported":    OffenderAlreadyReported,
	"bad_judgement_age":            BadJudgementAge,
	"bad_validator_index":          BadValidatorIndex,
	"bad_signature":                BadSignature,
}

func (e *DisputesErrorCode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if val, ok := disputesErrorMap[str]; ok {
			*e = val
			return nil
		}
		return errors.New("invalid error code name: " + str)
	}
	return errors.New("invalid error code format, expected string")
}

// DisputesInput
func (di *DisputesInput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding DisputesInput")
	var err error

	if err = di.Disputes.Decode(d); err != nil {
		return nil
	}

	return nil
}

// DisputesOutputData
func (dod *DisputesOutputData) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding DisputesOutputData")
	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	dod.OffendersMark = make(types.OffendersMark, length)
	for i := uint64(0); i < length; i++ {
		if err = dod.OffendersMark[i].Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// DisputesOutput
func (do *DisputesOutput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding DisputesOutput")
	var err error

	okOrErr, err := d.ReadPointerFlag()
	isOk := okOrErr == 0
	if isOk {
		cLog(Yellow, "DisputesOutput is ok")

		if do.Ok == nil {
			do.Ok = new(DisputesOutputData)
		}
		if err = do.Ok.Decode(d); err != nil {
			return err
		}

		return nil
	} else {
		cLog(Yellow, "DisputesOutput is err")
		cLog(Yellow, "Decoding DisputesErrorCode")

		errByte, err := d.ReadErrorByte()
		if err != nil {
			return err
		}

		do.Err = (*DisputesErrorCode)(&errByte)

		cLog(Yellow, fmt.Sprintf("DisputesErrorCode: %d", *do.Err))
	}

	return nil
}

// DisputesState
func (ds *DisputesState) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding DisputesState")
	var err error

	if err = ds.Psi.Decode(d); err != nil {
		return nil
	}

	if err = ds.Rho.Decode(d); err != nil {
		return nil
	}

	if err = ds.Tau.Decode(d); err != nil {
		return nil
	}

	if err = ds.Kappa.Decode(d); err != nil {
		return nil
	}

	if err = ds.Lambda.Decode(d); err != nil {
		return nil
	}

	return nil
}

// DisputesTestCase
func (dtc *DisputesTestCase) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding DisputesTestCase")
	var err error

	if err = dtc.Input.Decode(d); err != nil {
		return nil
	}

	if err = dtc.PreState.Decode(d); err != nil {
		return nil
	}

	if err = dtc.Output.Decode(d); err != nil {
		return nil
	}

	if err = dtc.PostState.Decode(d); err != nil {
		return nil
	}

	return nil
}
