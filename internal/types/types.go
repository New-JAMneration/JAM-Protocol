package types

// Reminder: When using jam_types, check if a Validate function exists.
// If a Validate function is available, remember to use it.
// If the desired Validate function is not found, please implement one yourself. :)
// version = 0.5.3
import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

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
		return fmt.Errorf("ValidatorsData must have exactly %d ValidatorData entries, got %d", ValidatorsCount, len(v))
	}
	return nil
}

// Service

type ServiceId U32

// ServiceInfo is part of (9.3) ServiceAccount and (9.8) ServiceAccountDerivatives
type ServiceInfo struct {
	CodeHash   OpaqueHash `json:"code_hash,omitempty"`    // a_c
	Balance    U64        `json:"balance,omitempty"`      // a_b
	MinItemGas Gas        `json:"min_item_gas,omitempty"` // a_g
	MinMemoGas Gas        `json:"min_memo_gas,omitempty"` // a_m
	Bytes      U64        `json:"bytes,omitempty"`        // a_o
	Items      U32        `json:"items,omitempty"`        // a_i
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
		return fmt.Errorf("AvailabilityAssignments must have exactly %d items, but got %d", CoresCount, len(assignments))
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
	Params   ByteSequence `json:"params,omitempty"`    // parameterization blob
}

type AuthorizerHash OpaqueHash

type AuthPool []AuthorizerHash

// (8.1) AuthPool and AuthQueue
func (a AuthPool) Validate() error {
	if len(a) > AuthPoolMaxSize {
		return fmt.Errorf("AuthPool exceeds max-auth-pool-size limit of %d", AuthPoolMaxSize)
	}
	return nil
}

type AuthPools []AuthPool

func (a AuthPools) Validate() error {
	if len(a) != CoresCount {
		return fmt.Errorf("AuthPools exceeds max-auth-pools limit of %d", CoresCount)
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
	if len(a) != AuthQueueSize {
		return fmt.Errorf("AuthQueue exceeds max-auth-queue-size limit of %d", AuthQueueSize)
	}
	return nil
}

type AuthQueues []AuthQueue

func (a AuthQueues) Validate() error {
	if len(a) != CoresCount {
		return fmt.Errorf("AuthQueues exceeds max-auth-queues limit of %d", CoresCount)
	}

	for _, pool := range a {
		err := pool.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// --- v0.6.3 Chapter 14.3. Packages and Items ---

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
	Authorization ByteSequence  `json:"authorization,omitempty"`  // authorization token
	AuthCodeHost  ServiceId     `json:"auth_code_host,omitempty"` // host service index
	Authorizer    Authorizer    `json:"authorizer"`
	Context       RefineContext `json:"context"`
	Items         []WorkItem    `json:"items,omitempty"`
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
		return fmt.Errorf("WorkPackage must have items between 1 and %d, but got %d", MaximumWorkItems, len(wp.Items))
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
		return fmt.Errorf("total size exceeds %d bytes", MaxTotalSize)
	}

	// import/export segment count check （14.4)
	if totalImportSegments+totalExportSegments > MaxSegments {
		return fmt.Errorf("total import and export segments exceed %d", MaxSegments)
	}

	// extrinsics count check (14.4)
	if totalExtrinsics > MaxExtrinsics {
		return fmt.Errorf("total extrinsics exceed %d", MaxExtrinsics)
	}

	// gas limit check (14.7)
	var totalRefineGas, totalAccumulateGas Gas
	for _, item := range wp.Items {
		totalRefineGas += item.RefineGasLimit
		totalAccumulateGas += item.AccumulateGasLimit
	}

	if totalRefineGas > MaxRefineGas {
		return fmt.Errorf("refine gas limit exceeds %d", MaxRefineGas)
	}
	if totalAccumulateGas > MaxAccumulateGas {
		return fmt.Errorf("accumulate gas limit exceeds %d", MaxAccumulateGas)
	}

	return nil
}

// --- v0.6.3 Chapter 11.1.4. Work Result ---

// v0.6.3 (11.7) WorkExecResultType
type WorkExecResultType string

const (
	WorkExecResultOk           WorkExecResultType = "ok"
	WorkExecResultOutOfGas                        = "out-of-gas"
	WorkExecResultPanic                           = "panic"
	WorkExecResultBadExports                      = "bad-exports"
	WorkExecResultBadCode                         = "bad-code"
	WorkExecResultCodeOversize                    = "code-oversize"
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
	GasUsed        U64 `json:"gas_used,omitempty"`        // u
	Imports        U16 `json:"imports,omitempty"`         // i
	ExtrinsicCount U16 `json:"extrinsic_count,omitempty"` // x
	ExtrinsicSize  U32 `json:"extrinsic_size,omitempty"`  // z
	Exports        U16 `json:"exports,omitempty"`         // e
}

// v0.6.4 (11.6) WorkResult $\mathbb{L}$
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
	Hash         WorkPackageHash `json:"hash,omitempty"`
	Length       U32             `json:"length,omitempty"`
	ErasureRoot  ErasureRoot     `json:"erasure_root,omitempty"`
	ExportsRoot  ExportsRoot     `json:"exports_root,omitempty"`
	ExportsCount U16             `json:"exports_count,omitempty"`
}

type SegmentRootLookupItem struct {
	WorkPackageHash WorkPackageHash `json:"work_package_hash,omitempty"`
	SegmentTreeRoot OpaqueHash      `json:"segment_tree_root,omitempty"`
}

type SegmentRootLookup []SegmentRootLookupItem // segment-tree-root

// v0.6.4 (11.2) WorkReport $\mathbb{W}$
type WorkReport struct {
	PackageSpec       WorkPackageSpec   `json:"package_spec"`                  // s
	Context           RefineContext     `json:"context"`                       // x
	CoreIndex         CoreIndex         `json:"core_index,omitempty"`          // c
	AuthorizerHash    OpaqueHash        `json:"authorizer_hash,omitempty"`     // a
	AuthOutput        ByteSequence      `json:"auth_output,omitempty"`         // \mathbf{o}
	SegmentRootLookup SegmentRootLookup `json:"segment_root_lookup,omitempty"` // \mathbf{r}
	Results           []WorkResult      `json:"results,omitempty"`             // \mathbf{l}
	AuthGasUsed       U64               `json:"auth_gas_used,omitempty"`       // g
}

func (w *WorkReport) Validate() error {
	if len(w.Results) < 1 || len(w.Results) > MaximumWorkItems {
		return fmt.Errorf("WorkReport Results must have items between 1 and %d, but got %d", MaximumWorkItems, len(w.Results))
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

type Mmr struct {
	Peaks []MmrPeak `json:"peaks,omitempty"`
}

type ReportedWorkPackage struct {
	Hash        WorkReportHash `json:"hash,omitempty"`
	ExportsRoot ExportsRoot    `json:"exports_root,omitempty"`
}

type BlockInfo struct {
	HeaderHash HeaderHash            `json:"header_hash,omitempty"`
	Mmr        Mmr                   `json:"mmr"`
	StateRoot  StateRoot             `json:"state_root,omitempty"`
	Reported   []ReportedWorkPackage `json:"reported,omitempty"`
}

type BlocksHistory []BlockInfo

func (b BlocksHistory) Validate() error {
	if len(b) > MaxBlocksHistory {
		return fmt.Errorf("BlocksHistory exceeds max-blocks-history limit of %d", MaxBlocksHistory)
	}
	return nil
}

// Statistics

type ActivityRecord struct {
	Blocks        U32 `json:"blocks,omitempty"`
	Tickets       U32 `json:"tickets,omitempty"`
	PreImages     U32 `json:"pre_images,omitempty"`
	PreImagesSize U32 `json:"pre_images_size,omitempty"`
	Guarantees    U32 `json:"guarantees,omitempty"`
	Assurances    U32 `json:"assurances,omitempty"`
}

type ActivityRecords []ActivityRecord

func (a ActivityRecords) Validate() error {
	if len(a) != ValidatorsCount {
		return fmt.Errorf("ActivityRecords must have %d activity record", ValidatorsCount)
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
	GasUsed        U64 `json:"gas_used,omitempty"`
}

type CoresStatistics []CoreActivityRecord

func (c CoresStatistics) Validate() error {
	if len(c) != CoresCount {
		return fmt.Errorf("CoresStatisitics must have %d core activity record", CoresCount)
	}
	return nil
}

// v0.6.4 (13.7)
type ServiceActivityRecord struct {
	ProvidedCount      U16 `json:"provided_count,omitempty"`
	ProvidedSize       U32 `json:"provided_size,omitempty"`
	RefinementCount    U32 `json:"refinement_count,omitempty"`
	RefinementGasUsed  U64 `json:"refinement_gas_used,omitempty"`
	Imports            U32 `json:"imports,omitempty"`
	Exports            U32 `json:"exports,omitempty"`
	ExtrinsicSize      U32 `json:"extrinsic_size,omitempty"`
	ExtrinsicCount     U32 `json:"extrinsic_count,omitempty"`
	AccumulateCount    U32 `json:"accumulate_count,omitempty"`
	AccumulateGasUsed  U64 `json:"accumulate_gas_used,omitempty"`
	OnTransfersCount   U32 `json:"on_transfers_count,omitempty"`
	OnTransfersGasUsed U64 `json:"on_transfers_gas_used,omitempty"`
}

type ServicesStatistics map[ServiceId]ServiceActivityRecord

// v0.6.4 (13.1)
type Statistics struct {
	ValsCurrent ActivityRecords    `json:"vals-current,omitempty"`
	ValsLast    ActivityRecords    `json:"vals-last,omitempty"`
	Cores       CoresStatistics    `json:"cores,omitempty"`
	Services    ServicesStatistics `json:"services,omitempty"`
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
		return fmt.Errorf("TicketsAccumulator exceeds maximum size of %d", EpochLength)
	}
	return nil
}

type TicketsOrKeys struct {
	Tickets []TicketBody         `json:"tickets,omitempty"`
	Keys    []BandersnatchPublic `json:"keys,omitempty"`
}

func (t TicketsOrKeys) Validate() error {
	if len(t.Tickets) > 0 && len(t.Tickets) != EpochLength {
		return fmt.Errorf("TicketsOrKeys Tickets must have size %d", EpochLength)
	}

	if len(t.Keys) > 0 && len(t.Keys) != EpochLength {
		return fmt.Errorf("TicketsOrKeys Keys must have size %d", EpochLength)
	}
	return nil
}

// (6.29)
type TicketsExtrinsic []TicketEnvelope

func (t *TicketsExtrinsic) Validate() error {
	if len(*t) > MaxTicketsPerBlock {
		return fmt.Errorf("TicketsExtrinsic exceeds maximum size of %d", MaxTicketsPerBlock)
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
		return fmt.Errorf("verdict Votes must have size %d", ValidatorsSuperMajority)
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
	Bitfield       []byte           `json:"bitfield,omitempty"`
	ValidatorIndex ValidatorIndex   `json:"validator_index,omitempty"`
	Signature      Ed25519Signature `json:"signature,omitempty"`
}

func (a AvailAssurance) Validate() error {
	if len(a.Bitfield) != AvailBitfieldBytes {
		return fmt.Errorf("AvailAssurance Bitfield must have size %d", AvailBitfieldBytes)
	}
	return nil
}

// (11.10)
type AssurancesExtrinsic []AvailAssurance

func (a *AssurancesExtrinsic) Validate() error {
	if len(*a) > ValidatorsCount {
		return fmt.Errorf("AssurancesExtrinsic exceeds maximum size of %d validators", ValidatorsCount)
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
		return fmt.Errorf("ValidatorIndex %d must be less than %d", v.ValidatorIndex, ValidatorsCount)
	}
	return nil
}

// (11.23)  Work Report Guarantee
type ReportGuarantee struct {
	Report     WorkReport           `json:"report"`
	Slot       TimeSlot             `json:"slot,omitempty"`
	Signatures []ValidatorSignature `json:"signatures,omitempty"`
}

func (r *ReportGuarantee) Validate() error {
	if len(r.Signatures) != 2 && len(r.Signatures) != 3 {
		return errors.New("signatures length must be between 2 and 3")
	}
	for i, sig := range r.Signatures {
		if err := sig.Validate(); err != nil {
			return fmt.Errorf("signature %d validation failed: %w", i, err)
		}
	}
	return nil
}

type GuaranteesExtrinsic []ReportGuarantee

// (11.23)
func (g *GuaranteesExtrinsic) Validate() error {
	if len(*g) > CoresCount {
		return fmt.Errorf("GuaranteesExtrinsic exceeds maximum size of %d cores", CoresCount)
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
		return fmt.Errorf("EpochMark Validators exceeds maximum size of %d", ValidatorsCount)
	}

	return nil
}

type TicketsMark []TicketBody

func (t TicketsMark) Validate() error {
	if len(t) != EpochLength {
		return fmt.Errorf("TicketsMark must have exactly %d tickets", EpochLength)
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

// Recent History
// \mathbf{C} in GP from type B (12.15)
type BeefyCommitmentOutput []AccumulationOutput // TODO: How to check unique

// Instant-used struct
type AccumulationOutput struct {
	Serviceid  ServiceId
	Commitment OpaqueHash
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
		return nil, fmt.Errorf("invalid length: expected %d bytes, got %d", expectedLen, len(decoded))
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
			return nil, fmt.Errorf("invalid length: expected 32 bytes, got %d", len(arr))
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

type Privileges struct {
	Bless       ServiceId           `json:"chi_m"` // Manager
	Assign      ServiceId           `json:"chi_a"` // AlterPhi
	Designate   ServiceId           `json:"chi_v"` // AlterIota
	AlwaysAccum AlwaysAccumulateMap `json:"chi_g"` // AutoAccumulateGasLimits
}

type AccumulateRoot OpaqueHash

// (12.13)
type PartialStateSet struct {
	ServiceAccounts ServiceAccountState
	ValidatorKeys   ValidatorsData
	Authorizers     AuthQueues
	Privileges      Privileges
}

// (12.18)
type Operand struct {
	Hash           WorkPackageHash
	ExportsRoot    ExportsRoot
	AuthorizerHash OpaqueHash
	AuthOutput     ByteSequence
	PayloadHash    OpaqueHash
	Result         WorkExecResult
}

// (12.15) U
type ServiceGasUsedList []ServiceGasUsed

type ServiceGasUsed struct {
	ServiceId ServiceId
	Gas       Gas
}

type CoreIndexReport struct {
	CoreID CoreIndex
	Report WorkReport
}
