package jamtests

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type DisputeTestCase struct {
	Input     DisputeInput  `json:"input"`
	PreState  DisputeState  `json:"pre_state"`
	Output    DisputeOutput `json:"output"`
	PostState DisputeState  `json:"post_state"`
}

type DisputeInput struct {
	disputes types.DisputesExtrinsic `json:"disputes"`
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
