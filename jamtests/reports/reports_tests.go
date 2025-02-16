package jamtests

import (
	"encoding/json"
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type ReportsTestCase struct {
	Input     ReportsInput  `json:"input"`
	PreState  ReportsState  `json:"pre_state"`
	Output    ReportsOutput `json:"output"`
	PostState ReportsState  `json:"post_state"`
}

type ServiceItem struct {
	ServiceId   types.ServiceId   `json:"service_id"`
	ServiceInfo types.ServiceInfo `json:"service_info"`
}

type Services []ServiceItem

type ReportsInput struct {
	Guarantees types.GuaranteesExtrinsic `json:"guarantees"`
	Slot       types.TimeSlot            `json:"slot"`
}

type ReportedPackage struct {
	WorkPackageHash types.WorkPackageHash `json:"work_package_hash"`
	SegmentTreeRoot types.OpaqueHash      `json:"segment_tree_root"`
}

type ReportsOutputData struct {
	Reported  []ReportedPackage     `json:"reported"`
	Reporters []types.Ed25519Public `json:"reporters"`
}

type ReportsOutput struct {
	Ok  *ReportsOutputData `json:"ok,omitempty"`
	Err *ReportsErrorCode  `json:"err,omitempty"`
}

type Account struct {
	Service types.ServiceInfo `json:"service"`
}

type AccountsMapEntry struct {
	Id   types.ServiceId `json:"id"`
	Info Account         `json:"data"`
}

type ReportsState struct {
	// [ρ‡] Intermediate pending reports after that any work report judged as
	// uncertain or invalid has been removed from it (ϱ†), and the availability
	// assurances are processed. Mutated to ϱ'.
	AvailAssignments types.AvailabilityAssignments `json:"avail_assignments"`

	// [κ'] Posterior active validators.
	CurrValidators types.ValidatorsData `json:"curr_validators"`

	// [λ'] Posterior previous validators.
	PrevValidators types.ValidatorsData `json:"prev_validators"`

	// [η'] Posterior entropy buffer.
	Entropy types.EntropyBuffer `json:"entropy"`

	//  [ψ'_o] Posterior offenders.
	Offenders []types.Ed25519Public `json:"offenders,omitempty"` // Offenders (psi_o)

	// [β] Recent blocks.
	RecentBlocks types.BlocksHistory `json:"recent_blocks"`

	// [α] Authorization pools.
	AuthPools types.AuthPools `json:"auth_pools"`

	// [δ] Relevant services account data. Refer to T(σ) in GP Appendix D.
	Accounts []AccountsMapEntry `json:"accounts"`
}

type ReportsErrorCode types.ErrorCode

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

var reportsErrorMap = map[string]ReportsErrorCode{
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

func (e *ReportsErrorCode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if val, ok := reportsErrorMap[str]; ok {
			*e = val
			return nil
		}
		return errors.New("invalid error code name: " + str)
	}
	return errors.New("invalid error code format, expected string")
}
