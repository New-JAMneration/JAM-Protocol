package types

import (
	. "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type ReportsErrorCode ErrorCode

const (
	BadCoreIndex                ReportsErrorCode = iota // 0
	FutureReportSlot                                    // 1
	ReportEpochBeforeLast                               // 2
	InsufficientGuarantees                              // 3
	OutOfOrderGuarantee                                 // 4
	NotSortedOrUniqueGuarantors                         // 5
	WrongAssignment                                     // 6
	CoreEngaged                                         // 7
	AnchorNotRecent                                     // 8
	BadServiceId                                        // 9
	BadCodeHash                                         // 10
	DependencyMissing                                   // 11
	DuplicatePackage                                    // 12
	BadStateRoot                                        // 13
	BadBeefyMmrRoot                                     // 14
	CoreUnauthorized                                    // 15
	BadValidatorIndex                                   // 16
	WorkReportGasTooHigh                                // 17
	ServiceItemGasTooLow                                // 18
	TooManyDependencies                                 // 19
	SegmentRootLookupInvalid                            // 20
	BadSignature                                        // 21
	WorkReportTooBig                                    // 22
)
