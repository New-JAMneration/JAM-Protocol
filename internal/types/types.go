package types

// Reminder: When using jam_types, check if a Validate function exists.
// If a Validate function is available, remember to use it.
// If the desired Validate function is not found, please implement one yourself. :)
// version = 0.5.3
import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

// Simple

type U8 uint8

type U16 uint16

type U32 uint32

type (
	U64          uint64
	ByteSequence []byte
	ByteArray32  [32]byte
	ByteString   string
)

type BitSequence []bool

// Crypto

type BandersnatchPublic [32]byte

type Ed25519Public [32]byte

type BlsPublic [144]byte

type BandersnatchVrfSignature [96]byte

type BandersnatchRingVrfSignature [784]byte

type Ed25519Signature [64]byte

type BandersnatchRingCommitment [144]byte

// Application Specific Core

type OpaqueHash ByteArray32

type (
	TimeSlot       U32
	ValidatorIndex U16
	CoreIndex      U16
)

type (
	HeaderHash      OpaqueHash
	StateRoot       OpaqueHash
	BeefyRoot       OpaqueHash
	WorkPackageHash OpaqueHash
	WorkReportHash  OpaqueHash
	ExportsRoot     OpaqueHash
	ErasureRoot     OpaqueHash
)

type (
	ErrorCode U8
	Gas       U64
)

func (e *ErrorCode) Error() string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *e)
}

type (
	Entropy       OpaqueHash
	EntropyBuffer [4]Entropy
)

type ValidatorMetadata [128]byte

type Validator struct {
	Bandersnatch BandersnatchPublic `json:"bandersnatch,omitempty"`
	Ed25519      Ed25519Public      `json:"ed25519,omitempty"`
	Bls          BlsPublic          `json:"bls,omitempty"`
	Metadata     ValidatorMetadata  `json:"metadata,omitempty"`
}

type ValidatorsData []Validator

func (v ValidatorsData) Validate() error {
	if len(v) != ValidatorsCount {
		return fmt.Errorf("ValidatorsData must have exactly %v ValidatorData entries, got %v", ValidatorsCount, len(v))
	}
	return nil
}

// Service

type (
	ServiceId     U32
	ServiceIdList []ServiceId
)

// ServiceInfo is part of (9.3) ServiceAccount and (9.8) ServiceAccountDerivatives
// GP 0.6.7
type ServiceInfo struct {
	DepositOffset        U64        `json:"deposit_offset,omitempty"`         // a_f
	CodeHash             OpaqueHash `json:"code_hash,omitempty"`              // a_c
	Balance              U64        `json:"balance,omitempty"`                // a_b
	MinItemGas           Gas        `json:"min_item_gas,omitempty"`           // a_g
	MinMemoGas           Gas        `json:"min_memo_gas,omitempty"`           // a_m
	CreationSlot         TimeSlot   `json:"creation_slot,omitempty"`          // a_r
	LastAccumulationSlot TimeSlot   `json:"last_accumulation_slot,omitempty"` // a_a
	ParentService        ServiceId  `json:"parent_service,omitempty"`         // a_p
	Bytes                U64        `json:"bytes,omitempty"`                  // a_o
	Items                U32        `json:"items,omitempty"`                  // a_i
}

type ServiceAccountDerivatives struct {
	Items      U32 `json:"items,omitempty"` // a_i
	Bytes      U64 `json:"bytes,omitempty"` // a_o
	Minbalance U64 // a_t
}

type MetaCode struct {
	Metadata ByteSequence
	Code     ByteSequence
}

// Availability Assignments

type AvailabilityAssignment struct {
	Report  WorkReport `json:"report"`
	Timeout TimeSlot   `json:"timeout,omitempty"`
}

func (a AvailabilityAssignment) Validate() error {
	if err := a.Report.Validate(); err != nil {
		return fmt.Errorf("AvailabilityAssignment Report validation failed: %v", err)
	}

	if a.Timeout == 0 {
		return errors.New("AvailabilityAssignment Timeout cannot be 0")
	}

	return nil
}

type AvailabilityAssignmentsItem *AvailabilityAssignment

type AvailabilityAssignments []AvailabilityAssignmentsItem

func (assignments AvailabilityAssignments) Validate() error {
	if len(assignments) != CoresCount {
		return fmt.Errorf("AvailabilityAssignments must have exactly %v items, but got %v", CoresCount, len(assignments))
	}
	return nil
}

// v0.6.3 (11.4) Refine Context $\mathbb{X}$
type RefineContext struct {
	Anchor           HeaderHash   `json:"anchor,omitempty"`             // anchor header hash
	StateRoot        StateRoot    `json:"state_root,omitempty"`         // posterior state root
	BeefyRoot        BeefyRoot    `json:"beefy_root,omitempty"`         // posterior beefy root
	LookupAnchor     HeaderHash   `json:"lookup_anchor,omitempty"`      // lookup anchor header hash
	LookupAnchorSlot TimeSlot     `json:"lookup_anchor_slot,omitempty"` // lookup anchor time slot
	Prerequisites    []OpaqueHash `json:"prerequisites,omitempty"`      // hash of prerequisite work packages
}

func (r *RefineContext) ScaleDecode(data []byte) error {
	_, err := scale.Decode("refinecontext", data, r)
	if err != nil {
		return err
	}

	return nil
}

func (r *RefineContext) ScaleEncode() ([]byte, error) {
	return scale.Encode("refinecontext", r)
}

// --- Chapter 8. Authorization ---

// Used in v0.6.3 (14.2)
type Authorizer struct {
	CodeHash OpaqueHash   `json:"code_hash,omitempty"` // authorization code hash
	Params   ByteSequence `json:"params,omitempty"`    // parameterization blob, the term is updated to auth config in 0.6.5
}

type AuthorizerHash OpaqueHash

type AuthPool []AuthorizerHash

// (8.1) AuthPool and AuthQueue
func (a AuthPool) Validate() error {
	if len(a) > AuthPoolMaxSize {
		return fmt.Errorf("AuthPool exceeds max-auth-pool-size limit of %v", AuthPoolMaxSize)
	}
	return nil
}

func (a *AuthPool) RemovePairedValue(h OpaqueHash) {
	result := (*a)[:0]
	for _, v := range *a {
		if !bytes.Equal(v[:], h[:]) {
			result = append(result, v)
		}
	}
	*a = result
}

type AuthPools []AuthPool

func (a AuthPools) Validate() error {
	if len(a) != CoresCount {
		return fmt.Errorf("AuthPools exceeds max-auth-pools limit of %v", CoresCount)
	}

	for _, pool := range a {
		err := pool.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

type AuthQueue []AuthorizerHash

func (a AuthQueue) Validate() error {
	// (8.1) φ ∈ ⟦⟦H⟧_Q⟧_C
	if len(a) != AuthQueueSize {
		return fmt.Errorf("length of authQueue %v exceeds max-auth-queue-size limit of %v", len(a), AuthQueueSize)
	}
	return nil
}

type AuthQueues []AuthQueue

func (a AuthQueues) Validate() error {
	// (8.1) φ ∈ ⟦⟦H⟧_Q⟧_C
	if len(a) != CoresCount {
		return fmt.Errorf("length of authQueues %v exceeds cores limit of %v", len(a), CoresCount)
	}

	for _, queue := range a {
		err := queue.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// --- v0.6.3 Chapter 14.3. Packages and Items ---

type (
	ExportSegment       [SegmentSize]byte
	ExportSegmentMatrix [][]ExportSegment
	OpaqueHashMatrix    [][]OpaqueHash
)

type ImportSpec struct {
	TreeRoot OpaqueHash `json:"tree_root,omitempty"` // hash of segment root or work package
	Index    U16        `json:"index,omitempty"`     // index of prior exported segments
}

type ExtrinsicSpec struct {
	Hash OpaqueHash `json:"hash,omitempty"`
	Len  U32        `json:"len,omitempty"`
}

// v0.6.3 (14.3) Work Item $\mathbb{I}$
type WorkItem struct {
	Service            ServiceId       `json:"service,omitempty"`              // service index $s$
	CodeHash           OpaqueHash      `json:"code_hash,omitempty"`            // code hash of the service $c$
	Payload            ByteSequence    `json:"payload,omitempty"`              // payload blob $\mathbf{y}$
	RefineGasLimit     Gas             `json:"refine_gas_limit,omitempty"`     // refinement gas limit $g$
	AccumulateGasLimit Gas             `json:"accumulate_gas_limit,omitempty"` // accumulatation gas limit $a$
	ExportCount        U16             `json:"export_count,omitempty"`         // number of exported data segments $e$
	ImportSegments     []ImportSpec    `json:"import_segments,omitempty"`      // sequence of imported data segments $\mathbf{i}$
	Extrinsic          []ExtrinsicSpec `json:"extrinsic,omitempty"`            // sequence of blob hashes and lengths $\mathbf{x}$
}

func (w *WorkItem) ScaleDecode(data []byte) error {
	_, err := scale.Decode("workitem", data, w)
	if err != nil {
		return err
	}

	return nil
}

func (w *WorkItem) ScaleEncode() ([]byte, error) {
	return scale.Encode("workitem", w)
}

// v0.6.3 (14.2) Work Package
type WorkPackage struct {
	Authorization ByteSequence  `json:"authorization,omitempty"`  // authorization token j
	AuthCodeHost  ServiceId     `json:"auth_code_host,omitempty"` // host service index h
	Authorizer    Authorizer    `json:"authorizer"`               // u, f
	Context       RefineContext `json:"context"`                  // c
	Items         []WorkItem    `json:"items,omitempty"`          // w
}

func (w *WorkPackage) ScaleDecode(data []byte) error {
	_, err := scale.Decode("workpackage", data, w)
	if err != nil {
		return err
	}

	return nil
}

func (w *WorkPackage) ScaleEncode() ([]byte, error) {
	return scale.Encode("workpackage", w)
}

func (wp *WorkPackage) Validate() error {
	// v0.6.3 (14.2)
	if len(wp.Items) < 1 || len(wp.Items) > MaximumWorkItems {
		return fmt.Errorf("WorkPackage must have items between 1 and %v, but got %v", MaximumWorkItems, len(wp.Items))
	}

	totalSize := len(wp.Authorization) + len(wp.Authorizer.Params)
	totalImportSegments := 0
	totalExportSegments := 0
	totalExtrinsics := 0

	for _, item := range wp.Items {
		totalSize += len(item.Payload)

		totalImportSegments += len(item.ImportSegments)
		totalSize += len(item.ImportSegments) * SegmentSize

		for _, extrinsic := range item.Extrinsic {
			totalSize += int(extrinsic.Len)
		}

		totalExportSegments += int(item.ExportCount)

		totalExtrinsics += len(item.Extrinsic)
	}

	// total size check (14.5)
	if totalSize > MaxTotalSize {
		return fmt.Errorf("total size exceeds %v bytes", MaxTotalSize)
	}

	// import segment count check （14.4)
	if totalImportSegments > MaxImportCount {
		return fmt.Errorf("total import segments exceed %d", MaxImportCount)
	}

	// export segment count check (14.4)
	if totalExportSegments > MaxExportCount {
		return fmt.Errorf("total export segments exceed %d", MaxExportCount)
	}

	// extrinsics count check (14.4)
	if totalExtrinsics > MaxExtrinsics {
		return fmt.Errorf("total extrinsics exceed %v", MaxExtrinsics)
	}

	// gas limit check (14.7)
	var totalRefineGas, totalAccumulateGas Gas
	for _, item := range wp.Items {
		totalRefineGas += item.RefineGasLimit
		totalAccumulateGas += item.AccumulateGasLimit
	}

	if totalRefineGas > MaxRefineGas {
		return fmt.Errorf("refine gas limit exceeds %s", fmt.Sprintf("%d", uint64(MaxRefineGas)))
	}
	if totalAccumulateGas > MaxAccumulateGas {
		return fmt.Errorf("accumulate gas limit exceeds %s", fmt.Sprintf("%d", MaxAccumulateGas))
	}

	return nil
}

// --- v0.6.3 Chapter 11.1.4. Work Result ---

// v0.6.3 (11.7) WorkExecResultType
type WorkExecResultType string

const (
	WorkExecResultOk             WorkExecResultType = "ok"
	WorkExecResultOutOfGas                          = "out-of-gas"
	WorkExecResultPanic                             = "panic"
	WorkExecResultBadExports                        = "bad-exports"
	WorkExecResultReportOversize                    = "report-oversize"
	WorkExecResultBadCode                           = "bad-code"
	WorkExecResultCodeOversize                      = "code-oversize"
)

type WorkExecResult map[WorkExecResultType][]byte

func GetWorkExecResult(resultType WorkExecResultType, data []byte) WorkExecResult {
	if resultType == WorkExecResultOk {
		return map[WorkExecResultType][]byte{
			resultType: data,
		}
	}

	return map[WorkExecResultType][]byte{
		resultType: nil,
	}
}

type RefineLoad struct {
	GasUsed        Gas `json:"gas_used,omitempty"`        // u
	Imports        U16 `json:"imports,omitempty"`         // i
	ExtrinsicCount U16 `json:"extrinsic_count,omitempty"` // x
	ExtrinsicSize  U32 `json:"extrinsic_size,omitempty"`  // z
	Exports        U16 `json:"exports,omitempty"`         // e
}

// v0.6.4 (11.6) WorkResult $\mathbb{L}$
// v0.7.0 (11.6) Work Digest
type WorkResult struct {
	ServiceId     ServiceId      `json:"service_id,omitempty"`     // s
	CodeHash      OpaqueHash     `json:"code_hash,omitempty"`      // c
	PayloadHash   OpaqueHash     `json:"payload_hash,omitempty"`   // y
	AccumulateGas Gas            `json:"accumulate_gas,omitempty"` // g
	Result        WorkExecResult `json:"result,omitempty"`         // $\mathbf{d}$
	RefineLoad    RefineLoad     `json:"refine_load,omitempty"`    // ASN.1 specific field
}

func (w *WorkResult) Validate() error {
	return nil
}

func (w *WorkResult) ScaleDecode(data []byte) error {
	_, err := scale.Decode("workresult", data, w)
	if err != nil {
		return err
	}

	if err := w.Validate(); err != nil {
		return err
	}

	return nil
}

func (w *WorkResult) ScaleEncode() ([]byte, error) {
	return scale.Encode("workresult", w)
}

// --- v0.6.3 Chapter 11.1.1. Work Report ---

// v0.6.3 (11.5) Availability specifications $\mathbb{S}$
type WorkPackageSpec struct {
	Hash         WorkPackageHash `json:"hash,omitempty"`          // p
	Length       U32             `json:"length,omitempty"`        // l
	ErasureRoot  ErasureRoot     `json:"erasure_root,omitempty"`  // u
	ExportsRoot  ExportsRoot     `json:"exports_root,omitempty"`  // e
	ExportsCount U16             `json:"exports_count,omitempty"` // n
}

type SegmentRootLookupItem struct {
	WorkPackageHash WorkPackageHash `json:"work_package_hash,omitempty"`
	SegmentTreeRoot OpaqueHash      `json:"segment_tree_root,omitempty"`
}

type SegmentRootLookup []SegmentRootLookupItem // segment-tree-root

// v0.6.4 (11.2) WorkReport $\mathbb{W}$
type WorkReport struct {
	PackageSpec       WorkPackageSpec   `json:"package_spec"`                  // \mathbf{s}
	Context           RefineContext     `json:"context"`                       // \mathbf{c}
	CoreIndex         CoreIndex         `json:"core_index,omitempty"`          // c
	AuthorizerHash    OpaqueHash        `json:"authorizer_hash,omitempty"`     // a
	AuthOutput        ByteSequence      `json:"auth_output,omitempty"`         // \mathbf{t}
	SegmentRootLookup SegmentRootLookup `json:"segment_root_lookup,omitempty"` // \mathbf{l}
	Results           []WorkResult      `json:"results,omitempty"`             // \mathbf{d}
	AuthGasUsed       Gas               `json:"auth_gas_used,omitempty"`       // g
}

func (w *WorkReport) Validate() error {
	if len(w.Results) < 1 || len(w.Results) > MaximumWorkItems {
		return fmt.Errorf("WorkReport Results must have items between 1 and %v, but got %v", MaximumWorkItems, len(w.Results))
	}

	return nil
}

// ValidateLookupDictAndPrerequisites checks the number of SegmentRootLookup and Prerequisites < J | Eq. 11.3
func (w *WorkReport) ValidateLookupDictAndPrerequisites() error {
	if len(w.SegmentRootLookup)+len(w.Context.Prerequisites) > MaximumDependencyItems {
		// return fmt.Errorf("SegmentRootLookup and Prerequisites must have a total at most %d, but got %d", MaximumDependencyItems, len(w.SegmentRootLookup)+len(w.Context.Prerequisites))
		return fmt.Errorf("too_many_dependencies")
	}
	return nil
}

// ValidateOutputSize checks the total size of the output | Eq. 11.8
func (w *WorkReport) ValidateOutputSize() error {
	totalSize := len(w.AuthOutput)
	for _, result := range w.Results {
		for _, outputs := range result.Result {
			totalSize += len(outputs)
		}
	}

	if totalSize > WorkReportOutputBlobsMaximumSize {
		return fmt.Errorf("work_report_too_big")
	}
	return nil
}

func (w *WorkReport) ScaleDecode(data []byte) error {
	_, err := scale.Decode("workreport", data, w)
	if err != nil {
		return err
	}

	if err := w.Validate(); err != nil {
		return err
	}

	return nil
}

func (w *WorkReport) ScaleEncode() ([]byte, error) {
	return scale.Encode("workreport", w)
}

// Block History   ??

type MmrPeak *OpaqueHash

// Beefy Belt (7.3) GP 0.6.7
type Mmr struct {
	Peaks []MmrPeak `json:"peaks,omitempty"`
}

type ReportedWorkPackage struct {
	Hash        WorkReportHash `json:"hash,omitempty"`
	ExportsRoot ExportsRoot    `json:"exports_root,omitempty"`
}

type BlockInfo struct {
	HeaderHash HeaderHash            `json:"header_hash,omitempty"`
	BeefyRoot  OpaqueHash            `json:"beefy_root,omitempty"`
	StateRoot  StateRoot             `json:"state_root,omitempty"`
	Reported   []ReportedWorkPackage `json:"reported,omitempty"`
}

// (7.2) GP 0.6.7
type BlocksHistory []BlockInfo

func (b BlocksHistory) Validate() error {
	if len(b) > MaxBlocksHistory {
		return fmt.Errorf("BlocksHistory exceeds max-blocks-history limit of %v", MaxBlocksHistory)
	}
	return nil
}

// (7.1) GP 0.6.7
type RecentBlocks struct {
	History BlocksHistory `json:"history,omitempty"`
	Mmr     Mmr           `json:"mmr,omitempty"`
}

// Statistics

type ValidatorActivityRecord struct {
	Blocks        U32 `json:"blocks,omitempty"`
	Tickets       U32 `json:"tickets,omitempty"`
	PreImages     U32 `json:"pre_images,omitempty"`
	PreImagesSize U32 `json:"pre_images_size,omitempty"`
	Guarantees    U32 `json:"guarantees,omitempty"`
	Assurances    U32 `json:"assurances,omitempty"`
}

type ValidatorsStatistics []ValidatorActivityRecord

func (a ValidatorsStatistics) Validate() error {
	if len(a) != ValidatorsCount {
		return fmt.Errorf("ActivityRecords must have %v activity record", ValidatorsCount)
	}
	return nil
}

// v0.6.4 (13.6)
type CoreActivityRecord struct {
	DALoad         U32 `json:"da_load,omitempty"`
	Popularity     U16 `json:"popularity,omitempty"`
	Imports        U16 `json:"imports,omitempty"`
	Exports        U16 `json:"exports,omitempty"`
	ExtrinsicSize  U32 `json:"extrinsic_size,omitempty"`
	ExtrinsicCount U16 `json:"extrinsic_count,omitempty"`
	BundleSize     U32 `json:"bundle_size,omitempty"`
	GasUsed        Gas `json:"gas_used,omitempty"`
}

type CoresStatistics []CoreActivityRecord

func (c CoresStatistics) Validate() error {
	if len(c) != CoresCount {
		return fmt.Errorf("CoresStatisitics must have %v core activity record", CoresCount)
	}
	return nil
}

// v0.6.4 (13.7)
type ServiceActivityRecord struct {
	ProvidedCount      U16 `json:"provided_count,omitempty"`
	ProvidedSize       U32 `json:"provided_size,omitempty"`
	RefinementCount    U32 `json:"refinement_count,omitempty"`
	RefinementGasUsed  Gas `json:"refinement_gas_used,omitempty"`
	Imports            U32 `json:"imports,omitempty"`
	Exports            U32 `json:"exports,omitempty"`
	ExtrinsicSize      U32 `json:"extrinsic_size,omitempty"`
	ExtrinsicCount     U32 `json:"extrinsic_count,omitempty"`
	AccumulateCount    U32 `json:"accumulate_count,omitempty"`
	AccumulateGasUsed  Gas `json:"accumulate_gas_used,omitempty"`
	OnTransfersCount   U32 `json:"on_transfers_count,omitempty"`
	OnTransfersGasUsed Gas `json:"on_transfers_gas_used,omitempty"`
}

type ServicesStatistics map[ServiceId]ServiceActivityRecord

// v0.6.4 (13.1)
type Statistics struct {
	ValsCurr ValidatorsStatistics `json:"vals-curr,omitempty"`
	ValsLast ValidatorsStatistics `json:"vals-last,omitempty"`
	Cores    CoresStatistics      `json:"cores,omitempty"`
	Services ServicesStatistics   `json:"services,omitempty"`
}

// Tickets   (6.5)   or  6.7.  ?

type (
	TicketId      OpaqueHash
	TicketAttempt U8
)

type TicketEnvelope struct {
	Attempt   TicketAttempt                `json:"attempt,omitempty"`
	Signature BandersnatchRingVrfSignature `json:"signature,omitempty"`
}

type TicketBody struct {
	Id      TicketId      `json:"id,omitempty"`
	Attempt TicketAttempt `json:"attempt,omitempty"`
}

// (6.5)
type TicketsAccumulator []TicketBody

func (t TicketsAccumulator) Validate() error {
	if len(t) > EpochLength {
		return fmt.Errorf("TicketsAccumulator exceeds maximum size of %v", EpochLength)
	}
	return nil
}

type TicketsOrKeys struct {
	Tickets []TicketBody         `json:"tickets,omitempty"`
	Keys    []BandersnatchPublic `json:"keys,omitempty"`
}

func (t TicketsOrKeys) Validate() error {
	if len(t.Tickets) > 0 && len(t.Tickets) != EpochLength {
		return fmt.Errorf("TicketsOrKeys Tickets must have size %v", EpochLength)
	}

	if len(t.Keys) > 0 && len(t.Keys) != EpochLength {
		return fmt.Errorf("TicketsOrKeys Keys must have size %v", EpochLength)
	}
	return nil
}

// (6.29)
type TicketsExtrinsic []TicketEnvelope

func (t *TicketsExtrinsic) Validate() error {
	if len(*t) > MaxTicketsPerBlock {
		return fmt.Errorf("TicketsExtrinsic exceeds maximum size of %v", MaxTicketsPerBlock)
	}

	return nil
}

func (t *TicketsExtrinsic) ScaleDecode(data []byte) error {
	_, err := scale.Decode("ticketsextrinsic", data, t)
	if err != nil {
		return err
	}

	if err := t.Validate(); err != nil {
		return err
	}

	return nil
}

func (t *TicketsExtrinsic) ScaleEncode() ([]byte, error) {
	return scale.Encode("ticketsextrinsic", t)
}

// 10. Disputes

type Judgement struct {
	Vote      bool             `json:"vote,omitempty"`
	Index     ValidatorIndex   `json:"index,omitempty"`
	Signature Ed25519Signature `json:"signature,omitempty"`
}

// (10.2)   E_D ≡ (v,c,f)
// v = Verdict
type Verdict struct {
	Target OpaqueHash  `json:"target,omitempty"`
	Age    U32         `json:"age,omitempty"`
	Votes  []Judgement `json:"votes,omitempty"`
}

func (v Verdict) Validate() error {
	if len(v.Votes) != ValidatorsSuperMajority {
		return fmt.Errorf("verdict Votes must have size %v", ValidatorsSuperMajority)
	}
	return nil
}

// c = Culprit
type Culprit struct {
	Target    WorkReportHash   `json:"target,omitempty"`
	Key       Ed25519Public    `json:"key,omitempty"`
	Signature Ed25519Signature `json:"signature,omitempty"`
}

// f = Fault
type Fault struct {
	Target    WorkReportHash   `json:"target,omitempty"`
	Vote      bool             `json:"vote,omitempty"`
	Key       Ed25519Public    `json:"key,omitempty"`
	Signature Ed25519Signature `json:"signature,omitempty"`
}

// (10.1)
type DisputesRecords struct {
	Good      []WorkReportHash `json:"good,omitempty"`      // Good verdicts (psi_g)
	Bad       []WorkReportHash `json:"bad,omitempty"`       // Bad verdicts (psi_b)
	Wonky     []WorkReportHash `json:"wonky,omitempty"`     // Wonky verdicts (psi_w)
	Offenders []Ed25519Public  `json:"offenders,omitempty"` // Offenders (psi_o)
}

// 10.2. (10.2)
type DisputesExtrinsic struct {
	Verdicts []Verdict `json:"verdicts,omitempty"`
	Culprits []Culprit `json:"culprits,omitempty"`
	Faults   []Fault   `json:"faults,omitempty"`
}

func (d *DisputesExtrinsic) Validate() error {
	for _, verdict := range d.Verdicts {
		if err := verdict.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (d *DisputesExtrinsic) ScaleDecode(data []byte) error {
	_, err := scale.Decode("disputesextrinsic", data, d)
	if err != nil {
		return err
	}

	if err := d.Validate(); err != nil {
		return err
	}

	return nil
}

func (d *DisputesExtrinsic) ScaleEncode() ([]byte, error) {
	return scale.Encode("disputesextrinsic", d)
}

// Preimages

type Preimage struct {
	Requester ServiceId    `json:"requester,omitempty"`
	Blob      ByteSequence `json:"blob,omitempty"`
}

func (p *Preimage) Validate() error {
	return nil
}

// (12.28) (12.29)
type PreimagesExtrinsic []Preimage

func (p *PreimagesExtrinsic) Validate() error {
	for _, preimage := range *p {
		if err := preimage.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Len returns the length of the Preimages slice.
func (p *PreimagesExtrinsic) Len() int {
	return len(*p)
}

// Less returns true if the Preimage at index i is less than the Preimage at
// index j.
func (p *PreimagesExtrinsic) Less(i, j int) bool {
	if (*p)[i].Requester == (*p)[j].Requester {
		return bytes.Compare((*p)[i].Blob, (*p)[j].Blob) < 0
	}
	return (*p)[i].Requester < (*p)[j].Requester
}

// Swap swaps the Preimages at index i and j.
func (p *PreimagesExtrinsic) Swap(i, j int) {
	(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
}

// Sort sorts the Preimages slice.
func (p *PreimagesExtrinsic) Sort() {
	sort.Sort(p)
}

func (p *PreimagesExtrinsic) ScaleDecode(data []byte) error {
	_, err := scale.Decode("preimagesextrinsic", data, p)
	if err != nil {
		return err
	}

	if err := p.Validate(); err != nil {
		return err
	}

	return nil
}

func (p *PreimagesExtrinsic) ScaleEncode() ([]byte, error) {
	return scale.Encode("preimagesextrinsic", p)
}

// 11.2.1 Assurances
type AvailAssurance struct {
	Anchor         OpaqueHash       `json:"anchor,omitempty"`
	Bitfield       Bitfield         `json:"bitfield,omitempty"`
	ValidatorIndex ValidatorIndex   `json:"validator_index,omitempty"`
	Signature      Ed25519Signature `json:"signature,omitempty"`
}

func (a AvailAssurance) Validate() error {
	return a.Bitfield.Validate()
}

type Bitfield []byte

func MakeBitfieldFromHexString(hexStr string) (Bitfield, error) {
	if !strings.HasPrefix(hexStr, "0x") {
		return Bitfield{}, errors.New("hex string for bitfield must have the 0x prefix")
	}

	bytes, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		return Bitfield{}, err
	}

	return MakeBitfieldFromByteSlice(bytes)
}

func MakeBitfieldFromByteSlice(bytes []byte) (Bitfield, error) {
	if len(bytes) != AvailBitfieldBytes {
		return Bitfield{}, fmt.Errorf("Bitfield must have size %v bytes", AvailBitfieldBytes)
	}

	bitfield := make(Bitfield, CoresCount)
	for i := range bitfield {
		bitfield[i] = (bytes[i/8] >> (i % 8)) & 0x01
	}

	return bitfield, nil
}

// panics on error
// this method should only be used by test code
func MustMakeBitfieldFromHexString(hexStr string) Bitfield {
	bitfield, err := MakeBitfieldFromHexString(hexStr)
	if err != nil {
		panic(err)
	}
	return bitfield
}

func (bf Bitfield) ToOctetSlice() []byte {
	bytes := make([]byte, AvailBitfieldBytes)

	for i, b := range bf {
		bytes[i/8] |= b << (i % 8)
	}

	return bytes
}

// returns either 1 or 0
func (bf Bitfield) GetBit(index int) byte {
	return bf[index]
}

func (bf Bitfield) Validate() error {
	if len(bf) != CoresCount {
		return fmt.Errorf("Bitfield must have size %v", CoresCount)
	}

	return nil
}

// (11.10)
type AssurancesExtrinsic []AvailAssurance

func (a *AssurancesExtrinsic) Validate() error {
	if len(*a) > ValidatorsCount {
		return fmt.Errorf("AssurancesExtrinsic exceeds maximum size of %v validators", ValidatorsCount)
	}
	return nil
}

func (a *AssurancesExtrinsic) ScaleDecode(data []byte) error {
	_, err := scale.Decode("assurancesextrinsic", data, a)
	if err != nil {
		return err
	}

	if err := a.Validate(); err != nil {
		return err
	}

	return nil
}

func (a *AssurancesExtrinsic) ScaleEncode() ([]byte, error) {
	return scale.Encode("assurancesextrinsic", a)
}

// Guarantees

type ValidatorSignature struct {
	ValidatorIndex ValidatorIndex   `json:"validator_index,omitempty"`
	Signature      Ed25519Signature `json:"signature,omitempty"`
}

func (v ValidatorSignature) Validate() error {
	if int(v.ValidatorIndex) >= ValidatorsCount {
		// return fmt.Errorf("ValidatorIndex %v must be less than %v", v.ValidatorIndex, ValidatorsCount)
		return fmt.Errorf("bad_validator_index")
	}
	return nil
}

// (11.23)  Work Report Guarantee
type ReportGuarantee struct {
	Report     WorkReport           `json:"report"`               // g_w
	Slot       TimeSlot             `json:"slot,omitempty"`       // g_t
	Signatures []ValidatorSignature `json:"signatures,omitempty"` // g_a
}

func (r *ReportGuarantee) Validate() error {
	if err := r.Report.Validate(); err != nil {
		log.Println("report validation failed: %w", err)
	}

	if len(r.Signatures) < 2 {
		return errors.New("insufficient_guarantees")
	}

	if len(r.Signatures) > 3 {
		log.Println("too_many_guarantees")
	}

	for _, sig := range r.Signatures {
		if err := sig.Validate(); err != nil {
			return err
		}
	}

	for _, sig := range r.Signatures {
		if err := sig.Validate(); err != nil {
			// bad_validator_index : validator index is too big
			return err
		}
	}

	err := r.Report.ValidateLookupDictAndPrerequisites()
	if err != nil {
		// "too_many_dependencies"
		return err
	}

	err = r.Report.ValidateOutputSize()
	if err != nil {
		// "too_big_work_report_output"
		return err
	}

	return nil
}

type GuaranteesExtrinsic []ReportGuarantee

// (11.23)
func (g *GuaranteesExtrinsic) Validate() error {
	if len(*g) > CoresCount {
		return fmt.Errorf("Len of guaranteesExtrinsic %v exceeds maximum size of %v cores", len(*g), CoresCount)
	}
	for i, report := range *g {
		if err := report.Validate(); err != nil {
			return fmt.Errorf("eg's report[%v] validation failed: %w", i, err)
		}
	}
	return nil
}

func (g *GuaranteesExtrinsic) ScaleDecode(data []byte) error {
	_, err := scale.Decode("guaranteesextrinsic", data, g)
	if err != nil {
		return err
	}

	if err := g.Validate(); err != nil {
		return err
	}

	return nil
}

func (g *GuaranteesExtrinsic) ScaleEncode() ([]byte, error) {
	return scale.Encode("guaranteesextrinsic", g)
}

// Header
// (6.27)
type EpochMarkValidatorKeys struct {
	Bandersnatch BandersnatchPublic
	Ed25519      Ed25519Public
}

type EpochMark struct {
	Entropy        Entropy                  `json:"entropy,omitempty"`
	TicketsEntropy Entropy                  `json:"tickets_entropy,omitempty"`
	Validators     []EpochMarkValidatorKeys `json:"validators,omitempty"`
}

func (e EpochMark) Validate() error {
	if len(e.Validators) != ValidatorsCount {
		return fmt.Errorf("EpochMark Validators exceeds maximum size of %v", ValidatorsCount)
	}

	return nil
}

type TicketsMark []TicketBody

func (t TicketsMark) Validate() error {
	if len(t) != EpochLength {
		return fmt.Errorf("TicketsMark must have exactly %v tickets", EpochLength)
	}
	return nil
}

type OffendersMark []Ed25519Public

// (5.1)
type Header struct {
	Parent          HeaderHash               `json:"parent,omitempty"`            // H_p
	ParentStateRoot StateRoot                `json:"parent_state_root,omitempty"` // H_r
	ExtrinsicHash   OpaqueHash               `json:"extrinsic_hash,omitempty"`    // H_x
	Slot            TimeSlot                 `json:"slot,omitempty"`              // H_t
	EpochMark       *EpochMark               `json:"epoch_mark,omitempty"`        // H_e
	TicketsMark     *TicketsMark             `json:"tickets_mark,omitempty"`      // H_w
	OffendersMark   OffendersMark            `json:"offenders_mark,omitempty"`    // H_o
	AuthorIndex     ValidatorIndex           `json:"author_index,omitempty"`      // H_i
	EntropySource   BandersnatchVrfSignature `json:"entropy_source,omitempty"`    // H_v
	Seal            BandersnatchVrfSignature `json:"seal,omitempty"`              // H_s
}

func (h *Header) Validate() error {
	if h.EpochMark != nil {
		if err := h.EpochMark.Validate(); err != nil {
			return err
		}
	}

	if h.TicketsMark != nil {
		if err := h.TicketsMark.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (h *Header) ScaleDecode(data []byte) error {
	_, err := scale.Decode("header", data, h)
	if err != nil {
		return err
	}

	if err := h.Validate(); err != nil {
		return err
	}

	return nil
}

func (h *Header) ScaleEncode() ([]byte, error) {
	return scale.Encode("header", h)
}

// Block

type Extrinsic struct {
	Tickets    TicketsExtrinsic    `json:"tickets,omitempty"`
	Preimages  PreimagesExtrinsic  `json:"preimages"`
	Guarantees GuaranteesExtrinsic `json:"guarantees"`
	Assurances AssurancesExtrinsic `json:"assurances,omitempty"`
	Disputes   DisputesExtrinsic   `json:"disputes"`
}

func (e *Extrinsic) Validate() error {
	if err := e.Tickets.Validate(); err != nil {
		return err
	}

	if err := e.Preimages.Validate(); err != nil {
		return err
	}

	if err := e.Guarantees.Validate(); err != nil {
		return err
	}

	if err := e.Assurances.Validate(); err != nil {
		return err
	}

	if err := e.Disputes.Validate(); err != nil {
		return err
	}

	return nil
}

func (e *Extrinsic) ScaleDecode(data []byte) error {
	_, err := scale.Decode("extrinsic", data, e)
	if err != nil {
		return err
	}

	if err := e.Validate(); err != nil {
		return err
	}

	return nil
}

func (e *Extrinsic) ScaleEncode() ([]byte, error) {
	return scale.Encode("extrinsic", e)
}

type Block struct {
	Header    Header    `json:"header"`
	Extrinsic Extrinsic `json:"extrinsic"`
}

func (b *Block) Validate() error {
	if err := b.Header.Validate(); err != nil {
		return err
	}

	return nil
}

func (b *Block) ScaleDecode(data []byte) error {
	_, err := scale.Decode("block", data, b)
	if err != nil {
		return err
	}

	if err := b.Validate(); err != nil {
		return err
	}

	return nil
}

func (b *Block) ScaleEncode() ([]byte, error) {
	return scale.Encode("block", b)
}

// Safrole

// type State struct {
// 	Tau           TimeSlot                   `json:"tau"`            // Most recent block's timeslot
// 	Eta           EntropyBuffer              `json:"eta"`            // Entropy accumulator and epochal randomness
// 	Lambda        ValidatorsData             `json:"lambda"`         // Validator keys and metadata which were active in the prior epoch
// 	Kappa         ValidatorsData             `json:"kappa"`          // Validator keys and metadata currently active
// 	GammaK        ValidatorsData             `json:"gamma_k"`        // Validator keys for the following epoch
// 	Iota          ValidatorsData             `json:"iota"`           // Validator keys and metadata to be drawn from next
// 	GammaA        TicketsAccumulator         `json:"gamma_a"`        // Sealing-key contest ticket accumulator
// 	GammaS        TicketsOrKeys              `json:"gamma_s"`        // Sealing-key series of the current epoch
// 	GammaZ        BandersnatchRingCommitment `json:"gamma_z"`        // Bandersnatch ring commitment
// 	PostOffenders []Ed25519Public            `json:"post_offendors"` // Posterior offenders sequence
// }

type Input struct {
	Slot      TimeSlot         `json:"slot"`      // Current slot
	Entropy   Entropy          `json:"entropy"`   // Per block entropy (originated from block entropy source VRF)
	Extrinsic TicketsExtrinsic `json:"extrinsic"` // Safrole extrinsic
}

type OutputData struct {
	EpochMark   *EpochMark   `json:"epoch_mark,omitempty"`   // New epoch marker (optional).
	TicketsMark *TicketsMark `json:"tickets_mark,omitempty"` // Winning tickets marker (optional).
}

type Output struct {
	Ok  *OutputData `json:"ok,omitempty"`
	Err *ErrorCode  `json:"err,omitempty"`
}

type TestCase struct {
	Input     Input  `json:"input"`
	PreState  State  `json:"pre_state"`
	Output    Output `json:"output"`
	PostState State  `json:"post_state"`
}

type InputWrapper[T any] struct {
	Input T
}

func ParseData[t any](fileName string) (InputWrapper[t], error) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", fileName, err)
		return InputWrapper[t]{}, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %s: %v\n", fileName, err)
		return InputWrapper[t]{}, err
	}
	var wrapper InputWrapper[t]
	err = json.Unmarshal(bytes, &wrapper)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON:: %v\n", err)
	}
	return wrapper, nil
}

func DecodeJSONByte(input []byte) []byte {
	toJSON, _ := json.Marshal(input)
	out := string(toJSON)[1 : len(string(toJSON))-1]
	return hexToBytes(out)
}

func hexToBytes(hexString string) []byte {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		fmt.Printf("failed to decode hex string: %v", err)
	}
	return bytes
}

func parseFixedByteArray(data []byte, expectedLen int) ([]byte, error) {
	// Peek at first byte to see if it starts with '[' or '"'
	if len(data) > 0 && data[0] == '[' {
		arr, err := parseNormalByteArray(data, expectedLen)
		if err != nil {
			return nil, err
		}
		return arr, nil
	}

	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return nil, err
	}

	if len(hexStr) < 2 || hexStr[:2] != "0x" {
		return nil, fmt.Errorf("invalid hex format: %s", hexStr)
	}

	decoded, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		return nil, err
	}

	if len(decoded) != expectedLen {
		return nil, fmt.Errorf("invalid length: expected %v bytes, got %v", expectedLen, len(decoded))
	}

	return decoded, nil
}

func parseNormalByteArray(data []byte, size int) ([]byte, error) {
	// Peek at first byte to see if it starts with '[' or '"'
	if len(data) > 0 && data[0] == '[' {
		// Data is an array like [0,255,34,...]
		var arr []byte
		if err := json.Unmarshal(data, &arr); err != nil {
			return arr, err
		}
		if len(arr) != size {
			return nil, fmt.Errorf("invalid length: expected 32 bytes, got %v", len(arr))
		}
		return arr, nil
	}

	return nil, fmt.Errorf("invalid format for parseNormalByteArray")
}

func (o *OpaqueHash) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 32)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

func (o *HeaderHash) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 32)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

func (o *Ed25519Signature) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 64)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

func (o *BandersnatchPublic) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 32)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

func (o *Ed25519Public) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 32)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

func (o *BlsPublic) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 144)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

func (o *BandersnatchVrfSignature) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 96)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

func (o *BandersnatchRingVrfSignature) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 784)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

func (o *BandersnatchRingCommitment) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 144)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

func (o *ValidatorMetadata) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 128)
	if err != nil {
		return err
	}
	copy(o[:], decoded)
	return nil
}

// (12.14) deferred transfer
type DeferredTransfer struct {
	SenderID   ServiceId `json:"senderid"`
	ReceiverID ServiceId `json:"receiverid"`
	Balance    U64       `json:"balance"`
	Memo       [128]byte `json:"memo"`
	GasLimit   Gas       `json:"gas"`
}

type DeferredTransfers []DeferredTransfer

// --------------------------------------------
// -- Accumulation
// --------------------------------------------

type ReadyRecord struct {
	Report       WorkReport
	Dependencies []WorkPackageHash
}

type (
	ReadyQueueItem       []ReadyRecord
	ReadyQueue           []ReadyQueueItem // SEQUENCE (SIZE(epoch-length)) OF ReadyQueueItem
	AccumulatedQueueItem []WorkPackageHash
	AccumulatedQueue     []AccumulatedQueueItem // SEQUENCE (SIZE(epoch-length)) OF AccumulatedQueueItem
)

type AlwaysAccumulateMap map[ServiceId]Gas

// jam-types.asn AlwaysAccumulateMapEntry
type AlwaysAccumulateMapDTO struct {
	ServiceId ServiceId `json:"id"`
	Gas       Gas       `json:"gas"`
}

type Privileges struct {
	Bless       ServiceId           `json:"bless"`      // Manager
	Assign      ServiceIdList       `json:"assign"`     // AlterPhi
	Designate   ServiceId           `json:"designate"`  // AlterIota
	AlwaysAccum AlwaysAccumulateMap `json:"always_acc"` // AutoAccumulateGasLimits
}

type AccumulateRoot OpaqueHash

// (12.13)
type PartialStateSet struct {
	ServiceAccounts ServiceAccountState // d
	ValidatorKeys   ValidatorsData      // i
	Authorizers     AuthQueues          // q
	Bless           ServiceId           // m
	Assign          ServiceIdList       // a
	Designate       ServiceId           // v
	AlwaysAccum     AlwaysAccumulateMap // z
}

// (12.18 pre-0.6.5)
// (12.19 0.6.5)
type Operand struct {
	Hash           WorkPackageHash // h
	ExportsRoot    ExportsRoot     // e
	AuthorizerHash OpaqueHash      // a
	PayloadHash    OpaqueHash      // y
	GasLimit       Gas             // g   0.6.5
	Result         WorkExecResult  // d
	AuthOutput     ByteSequence    // o
}

// (12.15) U
type ServiceGasUsedList []ServiceGasUsed

type ServiceGasUsed struct {
	ServiceId ServiceId
	Gas       Gas
}

type AccumulatedServiceHash struct {
	ServiceId ServiceId
	Hash      OpaqueHash // AccumulationOutput
}

// v0.6.7 (7.4)
// TODO: rename LastAccOut to Theta, and Theta to Vartheta
type LastAccOut []AccumulatedServiceHash

// (12.15) B
// INFO:
// - We define (12.15) AccumulatedServiceOutput as a map of AccumulatedServiceHash.
// - We convert the AccumulatedServiceOutput to LastAccOut (a slice of AccumulatedServiceHash) for (7.4) Theta
type AccumulatedServiceOutput map[AccumulatedServiceHash]bool

// (12.23)
type GasAndNumAccumulatedReports struct {
	Gas                   Gas
	NumAccumulatedReports U64
}

// (12.23)
// I: accumulation statistics
// dictionary<serviceId, (gas used, the number of work-reports accumulated)>
type AccumulationStatistics map[ServiceId]GasAndNumAccumulatedReports

// (12.29)
type NumDeferredTransfersAndTotalGasUsed struct {
	NumDeferredTransfers U64
	TotalGasUsed         Gas
}

// (12.29)
// X: deferred-transfers statistics
// dictionary<destination service index, (the number of deferred-transfers, total gas used)>
type DeferredTransfersStatistics map[ServiceId]NumDeferredTransfersAndTotalGasUsed

type AuditReport struct {
	CoreID      CoreIndex
	Report      WorkReport
	ValidatorID ValidatorIndex
	AuditResult bool
	Signature   Ed25519Signature
}

// (17.12)
type AssignmentMap map[WorkPackageHash][]ValidatorIndex

type AuditPool struct {
	mu   sync.RWMutex
	data map[WorkPackageHash][]AuditReport
}

type (
	ExtrinsicData     []byte
	ExtrinsicDataList []ExtrinsicData
	ExtrinsicDataMap  map[OpaqueHash]ExtrinsicData
)

type WorkPackageBundle struct {
	Package        WorkPackage
	Extrinsics     ExtrinsicDataList
	ImportSegments ExportSegmentMatrix
	ImportProofs   OpaqueHashMatrix
}

// v0.6.5
type ServiceBlob struct {
	ServiceID ServiceId
	Blob      []byte
}
type ServiceBlobs []ServiceBlob

// For state serialization, merklization, and reading trace test cases
type StateKey [31]byte

type StateKeyVal struct {
	Key   StateKey
	Value ByteSequence
}

type StateKeyVals []StateKeyVal

type StateKeyValDiff struct {
	Key           StateKey
	ExpectedValue ByteSequence
	ActualValue   ByteSequence
}

func Some[T any](v T) *T {
	return &v
}

type HashSegmentMap map[OpaqueHash]OpaqueHash
