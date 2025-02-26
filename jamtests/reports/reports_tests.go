package jamtests

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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

// unmarshal json ReportsState
func (e *ReportsState) UnmarshalJSON(data []byte) error {
	var err error

	var state struct {
		AvailAssignments types.AvailabilityAssignments `json:"avail_assignments"`
		CurrValidators   types.ValidatorsData          `json:"curr_validators"`
		PrevValidators   types.ValidatorsData          `json:"prev_validators"`
		Entropy          types.EntropyBuffer           `json:"entropy"`
		Offenders        []types.Ed25519Public         `json:"offenders,omitempty"`
		RecentBlocks     types.BlocksHistory           `json:"recent_blocks"`
		AuthPools        types.AuthPools               `json:"auth_pools"`
		Accounts         []AccountsMapEntry            `json:"accounts"`
	}

	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}

	if len(state.AvailAssignments) != 0 {
		e.AvailAssignments = state.AvailAssignments
	}

	e.CurrValidators = state.CurrValidators
	e.PrevValidators = state.PrevValidators
	e.Entropy = state.Entropy

	if len(state.Offenders) != 0 {
		e.Offenders = state.Offenders
	}

	e.RecentBlocks = state.RecentBlocks
	e.AuthPools = state.AuthPools

	if len(state.Accounts) != 0 {
		e.Accounts = state.Accounts
	}

	return nil
}

// ReportsInput
func (r *ReportsInput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding ReportsInput")
	var err error

	if err = r.Guarantees.Decode(d); err != nil {
		return nil
	}

	if err = r.Slot.Decode(d); err != nil {
		return nil
	}

	return nil
}

// ReportedPackage
func (r *ReportedPackage) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding ReportedPackage")

	var err error

	if err = r.WorkPackageHash.Decode(d); err != nil {
		return err
	}

	if err = r.SegmentTreeRoot.Decode(d); err != nil {
		return err
	}

	return nil
}

// ReportsOutputData
func (r *ReportsOutputData) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding ReportsOutputData")

	var err error

	reportedLength, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if reportedLength != 0 {
		r.Reported = make([]ReportedPackage, reportedLength)
		for i := range r.Reported {
			if err = r.Reported[i].Decode(d); err != nil {
				return err
			}
		}
	}

	reportersLength, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if reportersLength != 0 {
		r.Reporters = make([]types.Ed25519Public, reportersLength)
		for i := range r.Reporters {
			if err = r.Reporters[i].Decode(d); err != nil {
				return err
			}
		}
	}

	return nil
}

// ReportsOutput
func (r *ReportsOutput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding ReportsOutput")

	var err error

	okOrErr, err := d.ReadPointerFlag()
	if err != nil {
		return err
	}

	isOk := okOrErr == 0
	if isOk {
		cLog(Yellow, "ReportsOutput is ok")

		if r.Ok == nil {
			r.Ok = &ReportsOutputData{}
		}

		if err = r.Ok.Decode(d); err != nil {
			return err
		}
		return nil
	} else {
		cLog(Yellow, "ReportsOutput is err")
		cLog(Yellow, "Decoding ReportsErrorCode")

		// Read a byte as error code
		errByte, err := d.ReadErrorByte()
		if err != nil {
			return err
		}

		r.Err = (*ReportsErrorCode)(&errByte)

		cLog(Yellow, fmt.Sprintf("ReportsErrorCode: %v", *r.Err))
	}

	return nil
}

// Account
func (a *Account) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding Account")
	var err error

	if err = a.Service.Decode(d); err != nil {
		return nil
	}

	return nil
}

// AccountsMapEntry
func (a *AccountsMapEntry) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AccountsMapEntry")
	var err error

	if err = a.Id.Decode(d); err != nil {
		return nil
	}

	if err = a.Info.Decode(d); err != nil {
		return nil
	}

	return nil
}

// ReportsState
func (r *ReportsState) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding ReportsState")
	var err error

	if err = r.AvailAssignments.Decode(d); err != nil {
		return nil
	}

	if err = r.CurrValidators.Decode(d); err != nil {
		return nil
	}

	if err = r.PrevValidators.Decode(d); err != nil {
		return nil
	}

	if err = r.Entropy.Decode(d); err != nil {
		return nil
	}

	offendersLength, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if offendersLength != 0 {
		r.Offenders = make([]types.Ed25519Public, offendersLength)
		for i := range r.Offenders {
			if err = r.Offenders[i].Decode(d); err != nil {
				return err
			}
		}
	}

	if err = r.RecentBlocks.Decode(d); err != nil {
		return nil
	}

	if err = r.AuthPools.Decode(d); err != nil {
		return nil
	}

	accountLength, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if accountLength != 0 {
		r.Accounts = make([]AccountsMapEntry, accountLength)
		for i := range r.Accounts {
			if err = r.Accounts[i].Decode(d); err != nil {
				return err
			}
		}
	}

	return nil
}

// ReportsTestCase
func (r *ReportsTestCase) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding ReportsTestCase")
	var err error

	if err = r.Input.Decode(d); err != nil {
		return err
	}

	if err = r.PreState.Decode(d); err != nil {
		return err
	}

	if err = r.Output.Decode(d); err != nil {
		return err
	}

	if err = r.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}

// Encode
type Encodable interface {
	Encode(e *types.Encoder) error
}

// ReportsInput
func (r *ReportsInput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding ReportsInput")
	var err error

	if err = r.Guarantees.Encode(e); err != nil {
		return nil
	}

	if err = r.Slot.Encode(e); err != nil {
		return nil
	}

	return nil
}

// ReportsOutput
func (r *ReportsOutput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding ReportsOutput")
	var err error

	if r.Ok != nil {
		cLog(Yellow, "ReportsOutput is ok")
		if err := e.WriteByte(0); err != nil {
			return err
		}

		// Encode ReportsOutputData
		if err = r.Ok.Encode(e); err != nil {
			return err
		}

		return nil
	} else {
		cLog(Yellow, "ReportsOutput is err")
		if err := e.WriteByte(1); err != nil {
			return err
		}

		// Encode ReportsErrorCode
		if err = e.WriteByte(byte(*r.Err)); err != nil {
			return err
		}

		cLog(Yellow, fmt.Sprintf("ReportsErrorCode: %v", *r.Err))
	}

	return nil
}

// ReportedPackage
func (r *ReportedPackage) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding ReportedPackage")
	var err error

	if err = r.WorkPackageHash.Encode(e); err != nil {
		return err
	}

	if err = r.SegmentTreeRoot.Encode(e); err != nil {
		return err
	}

	return nil
}

// ReportsOutputData
func (r *ReportsOutputData) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding ReportsOutputData")
	var err error

	if err = e.EncodeLength(uint64(len(r.Reported))); err != nil {
		return err
	}

	for i := range r.Reported {
		if err = r.Reported[i].Encode(e); err != nil {
			return err
		}
	}

	if err = e.EncodeLength(uint64(len(r.Reporters))); err != nil {
		return err
	}

	for i := range r.Reporters {
		if err = r.Reporters[i].Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Account
func (a *Account) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding Account")
	var err error

	if err = a.Service.Encode(e); err != nil {
		return nil
	}

	return nil
}

// AccountsMapEntry
func (a *AccountsMapEntry) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AccountsMapEntry")
	var err error

	if err = a.Id.Encode(e); err != nil {
		return err
	}

	if err = a.Info.Encode(e); err != nil {
		return err
	}

	return nil
}

// ReportsState
func (r *ReportsState) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding ReportsState")
	var err error

	if err = r.AvailAssignments.Encode(e); err != nil {
		return nil
	}

	if err = r.CurrValidators.Encode(e); err != nil {
		return nil
	}

	if err = r.PrevValidators.Encode(e); err != nil {
		return nil
	}

	if err = r.Entropy.Encode(e); err != nil {
		return nil
	}

	if err = e.EncodeLength(uint64(len(r.Offenders))); err != nil {
		return err
	}

	for i := range r.Offenders {
		if err = r.Offenders[i].Encode(e); err != nil {
			return err
		}
	}

	if err = r.RecentBlocks.Encode(e); err != nil {
		return nil
	}

	if err = r.AuthPools.Encode(e); err != nil {
		return nil
	}

	if err = e.EncodeLength(uint64(len(r.Accounts))); err != nil {
		return err
	}

	for i := range r.Accounts {
		if err = r.Accounts[i].Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// ReportsTestCase
func (r *ReportsTestCase) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding ReportsTestCase")
	var err error

	if err = r.Input.Encode(e); err != nil {
		return err
	}

	if err = r.PreState.Encode(e); err != nil {
		return err
	}

	if err = r.Output.Encode(e); err != nil {
		return err
	}

	if err = r.PostState.Encode(e); err != nil {
		return err
	}

	return nil
}
