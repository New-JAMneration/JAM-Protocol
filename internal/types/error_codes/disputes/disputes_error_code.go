package types

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

const (
	AlreadyJudged             types.ErrorCode = iota // 0
	BadVoteSplit                                     // 1
	VerdictsNotSortedUnique                          // 2
	JudgementsNotSortedUnique                        // 3
	CulpritsNotSortedUnique                          // 4
	FaultsNotSortedUnique                            // 5
	NotEnoughCulprits                                // 6
	NotEnoughFaults                                  // 7
	CulpritsVerdictNotBad                            // 8
	FaultVerdictWrong                                // 9
	OffenderAlreadyReported                          // 10
	BadJudgementAge                                  // 11
	BadValidatorIndex                                // 12
	BadSignature                                     // 13
	BadGuarantorKey                                  // 14
	BadAuditorKey                                    // 15
)

var DisputesErrorMap = map[string]types.ErrorCode{
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
