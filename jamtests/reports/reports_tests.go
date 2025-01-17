package jamtests

import (
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
	Slot       types.TimeSlot            `json:"tau"`
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

type ReportsState struct {
	Alpha     types.AuthPools               `json:"alpha"`
	Beta      types.BlocksHistory           `json:"beta"`
	Delta     types.ServiceAccountState     `json:"accounts"`
	Eta       types.EntropyBuffer           `json:"eta"`
	Kappa     types.ValidatorsData          `json:"kappa"`
	Lambda    types.ValidatorsData          `json:"lambda"`
	Offenders []types.Ed25519Public         `json:"offenders,omitempty"` // Offenders (psi_o)
	Rho       types.AvailabilityAssignments `json:"rho"`
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
