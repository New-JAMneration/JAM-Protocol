package types

import (
	. "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type DisputeErrorCode ErrorCode

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
