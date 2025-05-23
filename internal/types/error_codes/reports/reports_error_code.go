package types

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

const (
	BadCoreIndex                types.ErrorCode = iota // 0
	FutureReportSlot                                   // 1
	ReportEpochBeforeLast                              // 2
	InsufficientGuarantees                             // 3
	OutOfOrderGuarantee                                // 4
	NotSortedOrUniqueGuarantors                        // 5
	WrongAssignment                                    // 6
	CoreEngaged                                        // 7
	AnchorNotRecent                                    // 8
	BadServiceId                                       // 9
	BadCodeHash                                        // 10
	DependencyMissing                                  // 11
	DuplicatePackage                                   // 12
	BadStateRoot                                       // 13
	BadBeefyMmrRoot                                    // 14
	CoreUnauthorized                                   // 15
	BadValidatorIndex                                  // 16
	WorkReportGasTooHigh                               // 17
	ServiceItemGasTooLow                               // 18
	TooManyDependencies                                // 19
	SegmentRootLookupInvalid                           // 20
	BadSignature                                       // 21
	WorkReportTooBig                                   // 22
)

var ReportsErrorMap = map[string]types.ErrorCode{
	"bad_core_index":                  BadCoreIndex,
	"future_report_slot":              FutureReportSlot,
	"report_epoch_before_last":        ReportEpochBeforeLast,
	"insufficient_guarantees":         InsufficientGuarantees,
	"out_of_order_guarantee":          OutOfOrderGuarantee,
	"not_sorted_or_unique_guarantors": NotSortedOrUniqueGuarantors,
	"wrong_assignment":                WrongAssignment,
	"core_engaged":                    CoreEngaged,
	"anchor_not_recent":               AnchorNotRecent,
	"bad_service_id":                  BadServiceId,
	"bad_code_hash":                   BadCodeHash,
	"dependency_missing":              DependencyMissing,
	"duplicate_package":               DuplicatePackage,
	"bad_state_root":                  BadStateRoot,
	"bad_beefy_mmr_root":              BadBeefyMmrRoot,
	"core_unauthorized":               CoreUnauthorized,
	"bad_validator_index":             BadValidatorIndex,
	"work_report_gas_too_high":        WorkReportGasTooHigh,
	"service_item_gas_too_low":        ServiceItemGasTooLow,
	"too_many_dependencies":           TooManyDependencies,
	"segment_root_lookup_invalid":     SegmentRootLookupInvalid,
	"bad_signature":                   BadSignature,
	"work_report_too_big":             WorkReportTooBig,
}
