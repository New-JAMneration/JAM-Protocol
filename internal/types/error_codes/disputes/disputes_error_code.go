package types

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

const (
	AlreadyJudged             types.ErrorCode = iota // 0
	BadVoteSplit                                     // 1
	VerdictsNotSortedUnique                          // 2
	JudgementsNotSortedUnique                        // 3
	CulpritsNotSortedUnique                          // 4
	FaultsNotSortedUnique                            // 5
	NotEnoughCulprits                                // 6 — unused since GP v0.8.0 dropped the ">= 2 culprits per bad verdict" rule; ordinal kept until the official v0.8.0 error enum lands (#1017)
	NotEnoughFaults                                  // 7
	CulpritsVerdictNotBad                            // 8
	FaultVerdictWrong                                // 9
	OffenderAlreadyReported                          // 10
	BadJudgementAge                                  // 11
	BadValidatorIndex                                // 12
	BadSignature                                     // 13
	BadGuarantorKey                                  // 14
	BadAuditorKey                                    // 15
	// GP v0.8.0 eq:disputesextrinsics sequence caps. Ordinals are provisional
	// until the official v0.8.0 error enum ships (#1017 / #1012).
	TooManyVerdicts // 16
	TooManyOffenses // 17
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
	"too_many_verdicts":            TooManyVerdicts,
	"too_many_offenses":            TooManyOffenses,
}

// This map provides human-readable messages following the fuzz-proto examples
var DisputesErrorCodeMessages = map[types.ErrorCode]string{
	AlreadyJudged:             "already judged",
	BadVoteSplit:              "bad vote split",
	VerdictsNotSortedUnique:   "verdicts not sorted and unique",
	JudgementsNotSortedUnique: "judgements not sorted and unique",
	CulpritsNotSortedUnique:   "culprits not sorted and unique",
	FaultsNotSortedUnique:     "faults not sorted and unique",
	NotEnoughCulprits:         "not enough culprits",
	NotEnoughFaults:           "not enough faults",
	CulpritsVerdictNotBad:     "culprits' verdict not 'bad'",
	FaultVerdictWrong:         "fault verdict wrong",
	OffenderAlreadyReported:   "offender already reported",
	BadJudgementAge:           "bad judgement age",
	BadValidatorIndex:         "bad validator index",
	BadSignature:              "bad signature",
	BadGuarantorKey:           "bad guarantor key",
	TooManyVerdicts:           "too many verdicts",
	TooManyOffenses:           "too many offenses",
}
