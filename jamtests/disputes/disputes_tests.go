package jamtests

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/google/go-cmp/cmp"
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

type DisputeTestCase struct {
	Input     DisputeInput  `json:"input"`
	PreState  DisputeState  `json:"pre_state"`
	Output    DisputeOutput `json:"output"`
	PostState DisputeState  `json:"post_state"`
}

type DisputeInput struct {
	Disputes types.DisputesExtrinsic `json:"disputes"`
}

type DisputeOutputData struct {
	OffendersMark types.OffendersMark `json:"offenders_mark"`
}

type DisputeOutput struct {
	Ok  *DisputeOutputData `json:"ok,omitempty"`
	Err *DisputeErrorCode  `json:"err,omitempty"`
}

type DisputeState struct {
	Psi    types.DisputesRecords         `json:"psi"`
	Rho    types.AvailabilityAssignments `json:"rho"`
	Tau    types.TimeSlot                `json:"tau"`
	Kappa  types.ValidatorsData          `json:"kappa"`
	Lambda types.ValidatorsData          `json:"lambda"`
}

type DisputeErrorCode types.ErrorCode

func (d *DisputeErrorCode) Error() string {
	if d == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *d)
}

const (
	AlreadyJudged             DisputeErrorCode = iota // 0
	BadVoteSplit                                      // 1
	VerdictsNotSortedUnique                           // 2
	JudgementsNotSortedUnique                         // 3
	CulpritsNotSortedUnique                           // 4
	FaultsNotSortedUnique                             // 5
	NotEnoughCulprits                                 // 6
	NotEnoughFaults                                   // 7
	CulpritsVerdictNotBad                             // 8
	FaultVerdictWrong                                 // 9
	OffenderAlreadyReported                           // 10
	BadJudgementAge                                   // 11
	BadValidatorIndex                                 // 12
	BadSignature                                      // 13
	BadGuarantorKey                                   // 14
	BadAuditorKey                                     // 15
)

var DisputeErrorMap = map[string]DisputeErrorCode{
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
	"bad_guarantor_key":            BadGuarantorKey,
	"bad_auditor_key":              BadAuditorKey,
}

func (e *DisputeErrorCode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if val, ok := DisputeErrorMap[str]; ok {
			*e = val
			return nil
		}
		return errors.New("invalid error code name: " + str)
	}
	return errors.New("invalid error code format, expected string")
}

// DisputesInput
func (di *DisputeInput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding DisputesInput")
	var err error

	if err = di.Disputes.Decode(d); err != nil {
		return nil
	}

	return nil
}

// DisputesOutputData
func (dod *DisputeOutputData) Decode(d *types.Decoder) error {
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
func (do *DisputeOutput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding DisputesOutput")
	var err error

	okOrErr, err := d.ReadPointerFlag()
	if err != nil {
		return err
	}

	isOk := okOrErr == 0
	if isOk {
		cLog(Yellow, "DisputesOutput is ok")

		if do.Ok == nil {
			do.Ok = new(DisputeOutputData)
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

		do.Err = (*DisputeErrorCode)(&errByte)

		cLog(Yellow, fmt.Sprintf("DisputesErrorCode: %d", *do.Err))
	}

	return nil
}

// DisputesState
func (ds *DisputeState) Decode(d *types.Decoder) error {
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
func (dtc *DisputeTestCase) Decode(d *types.Decoder) error {
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

// Encode
type Encodable interface {
	Encode(e *types.Encoder) error
}

// DisputesInput
func (di *DisputeInput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding DisputesInput")
	var err error

	if err = di.Disputes.Encode(e); err != nil {
		return nil
	}

	return nil
}

// DisputesOutputData
func (dod *DisputeOutputData) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding DisputesOutputData")
	var err error

	if err = e.EncodeLength(uint64(len(dod.OffendersMark))); err != nil {
		return err
	}

	for i := range dod.OffendersMark {
		if err = dod.OffendersMark[i].Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// DisputesOutput
func (do *DisputeOutput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding DisputesOutput")
	var err error

	if do.Ok != nil && do.Err != nil {
		return errors.New("both ok and err are not nil")
	}

	if do.Ok == nil && do.Err == nil {
		return errors.New("both ok and err are nil")
	}

	if do.Ok != nil {
		cLog(Yellow, "DisputesOutput is ok")
		if err := e.WriteByte(0); err != nil {
			return err
		}

		// Encode DisputeOutputData
		if err = do.Ok.Encode(e); err != nil {
			return err
		}

		return nil
	}

	if do.Err != nil {
		cLog(Yellow, "DisputesOutput is err")
		if err := e.WriteByte(1); err != nil {
			return err
		}

		// Encode DisputeErrorCode
		if err = e.WriteByte(byte(*do.Err)); err != nil {
			return err
		}

		cLog(Yellow, fmt.Sprintf("DisputeErrorCode: %d", *do.Err))

		return nil
	}

	return nil
}

// DisputesState
func (ds *DisputeState) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding DisputesState")
	var err error

	if err = ds.Psi.Encode(e); err != nil {
		return nil
	}

	if err = ds.Rho.Encode(e); err != nil {
		return nil
	}

	if err = ds.Tau.Encode(e); err != nil {
		return nil
	}

	if err = ds.Kappa.Encode(e); err != nil {
		return nil
	}

	if err = ds.Lambda.Encode(e); err != nil {
		return nil
	}

	return nil
}

// DisputesTestCase
func (dtc *DisputeTestCase) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding DisputesTestCase")
	var err error

	if err = dtc.Input.Encode(e); err != nil {
		return nil
	}

	if err = dtc.PreState.Encode(e); err != nil {
		return nil
	}

	if err = dtc.Output.Encode(e); err != nil {
		return nil
	}

	if err = dtc.PostState.Encode(e); err != nil {
		return nil
	}

	return nil
}

func (d *DisputeOutput) IsError() bool {
	return d.Err != nil
}

func (d *DisputeTestCase) Dump() error {
	store.ResetInstance()
	storeInstance := store.GetInstance()

	storeInstance.GetPriorStates().SetPsi(d.PreState.Psi)
	storeInstance.GetPriorStates().SetRho(d.PreState.Rho)
	storeInstance.GetPriorStates().SetTau(d.PreState.Tau)
	storeInstance.GetPriorStates().SetKappa(d.PreState.Kappa)
	storeInstance.GetPriorStates().SetLambda(d.PreState.Lambda)

	// Add block with DisputesExtrinsic
	block := types.Block{
		Extrinsic: types.Extrinsic{
			Disputes: d.Input.Disputes,
		},
	}
	storeInstance.AddBlock(block)

	return nil
}

func (d *DisputeTestCase) GetPostState() interface{} {
	return d.PostState
}

func (d *DisputeTestCase) GetOutput() interface{} {
	return d.Output
}

func (d *DisputeTestCase) ExpectError() error {
	if d.Output.Err == nil {
		return nil
	}
	return d.Output.Err
}

func (d *DisputeTestCase) Validate() error {
	s := store.GetInstance()

	if !reflect.DeepEqual(s.GetPosteriorStates().GetPsi(), d.PostState.Psi) {
		diff := cmp.Diff(s.GetPosteriorStates().GetPsi(), d.PostState.Psi)
		fmt.Errorf("Psi does not match expected:\n%v,\nbut got %v\nDiff:\n%v", d.PostState.Psi, s.GetPosteriorStates().GetPsi(), diff)
	}
	if !reflect.DeepEqual(s.GetPosteriorStates().GetRho(), d.PostState.Rho) {
		diff := cmp.Diff(s.GetPosteriorStates().GetRho(), d.PostState.Rho)
		fmt.Errorf("Rho does not match expected:\n%v,\nbut got %v\nDiff:\n%v", d.PostState.Rho, s.GetPosteriorStates().GetRho(), diff)
	}
	if s.GetPosteriorStates().GetTau() != d.PostState.Tau {
		fmt.Errorf("Time slot does not match expected: %v, but got %v", d.PostState.Tau, s.GetPosteriorStates().GetTau())
	}

	if !reflect.DeepEqual(s.GetPosteriorStates().GetKappa(), d.PostState.Kappa) {
		diff := cmp.Diff(s.GetPosteriorStates().GetKappa(), d.PostState.Kappa)
		fmt.Errorf("Kappa does not match expected:\n%v,\nbut got %v\nDiff:\n%v", d.PostState.Kappa, s.GetPosteriorStates().GetKappa(), diff)
	}

	if !reflect.DeepEqual(s.GetPosteriorStates().GetLambda(), d.PostState.Lambda) {
		diff := cmp.Diff(s.GetPosteriorStates().GetLambda(), d.PostState.Lambda)
		fmt.Errorf("Lambda does not match expected:\n%v,\nbut got %v\nDiff:\n%v", d.PostState.Lambda, s.GetPosteriorStates().GetLambda(), diff)
	}

	return nil
}
