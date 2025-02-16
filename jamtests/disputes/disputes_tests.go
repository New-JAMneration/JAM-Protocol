package jamtests

import (
	"encoding/json"
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

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
)

var disputeErrorMap = map[string]DisputeErrorCode{
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

func (e *DisputeErrorCode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if val, ok := disputeErrorMap[str]; ok {
			*e = val
			return nil
		}
		return errors.New("invalid error code name: " + str)
	}
	return errors.New("invalid error code format, expected string")
}
