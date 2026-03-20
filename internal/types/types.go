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
	"maps"
	"sort"
	"strings"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

// ============================================================================
// SIMPLE TYPES
// Basic integer and byte array types used throughout the protocol.
// These provide type-safe wrappers around primitive values.
// ============================================================================

type (
	U8  uint8  // Unsigned 8-bit integer (0-255)
	U16 uint16 // Unsigned 16-bit integer (0-65535)
	U32 uint32 // Unsigned 32-bit integer (0-4294967295)
	U64 uint64 // Unsigned 64-bit integer (0-18446744073709551615)
)

type (
	ByteSequence []byte         // Variable-length byte sequence
	ByteArray32  [HashSize]byte // Fixed-size 32-byte array, commonly used for hashes
)

type BitSequence []bool

// ============================================================================
// CRYPTOGRAPHIC TYPES
// Public keys and signatures used for various cryptographic operations.
// ============================================================================

type BandersnatchPublic [32]byte

type Ed25519Public [32]byte

type BlsPublic [144]byte

type BandersnatchVrfSignature [BandersnatchSigSize]byte

type BandersnatchRingVrfSignature [784]byte

type Ed25519Signature [Ed25519SigSize]byte

type BandersnatchRingCommitment [144]byte

// ============================================================================
// APPLICATION SPECIFIC CORE TYPES
// Core types that define the fundamental structures of the JAM protocol,
// including time representation, validator identification, and various
// hash types used throughout the system.
// ============================================================================

// Generic 32-byte hash
type OpaqueHash ByteArray32

type (
	TimeSlot       U32 // Protocol time unit
	ValidatorIndex U16 // Index identifying a validator in the current set
	CoreIndex      U16 // Index identifying a core in the system
)

type (
	HeaderHash      OpaqueHash // Hash of a block header
	StateRoot       OpaqueHash // Hash of the state root
	BeefyRoot       OpaqueHash // BEEFY consensus root hash
	WorkPackageHash OpaqueHash // Hash of a work package
	WorkReportHash  OpaqueHash // Hash of a work report
	ExportsRoot     OpaqueHash // Root hash of exported data
	ErasureRoot     OpaqueHash // Root hash of erasure-coded data
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
	Entropy       OpaqueHash // Randomness seed used for various protocol operations
	EntropyBuffer [4]Entropy // Buffer of entropy values
)

// Additional validator-specific metadata
type ValidatorMetadata [128]byte

// Complete set of data associated with a validator
type Validator struct {
	Bandersnatch BandersnatchPublic `json:"bandersnatch,omitempty"`
	Ed25519      Ed25519Public      `json:"ed25519,omitempty"`
	Bls          BlsPublic          `json:"bls,omitempty"`
	Metadata     ValidatorMetadata  `json:"metadata,omitempty"`
}

// Complete set of validator data for all validators
type ValidatorsData []Validator

func (v *ValidatorsData) Validate() error {
	if len(*v) != ValidatorsCount {
		return fmt.Errorf("ValidatorsData must have exactly %v ValidatorData entries, got %v", ValidatorsCount, len(*v))
	}
	return nil
}

// ============================================================================
// SERVICE TYPES
// Types related to services.
// ============================================================================

type (
	ServiceID     U32 // Unique identifier for a service
	ServiceIDList []ServiceID
)

// Information about a deployed service
// ServiceInfo is part of GP §9.3 and GP §9.8
type ServiceInfo struct {
	Version              U8         `json:"version"`                // version: Service information version
	CodeHash             OpaqueHash `json:"code_hash"`              // a_c: Hash of the service's code
	Balance              U64        `json:"balance"`                // a_b: Current balance of the service
	MinItemGas           Gas        `json:"min_item_gas"`           // a_g: Minimum gas required for processing an item
	MinMemoGas           Gas        `json:"min_memo_gas"`           // a_m: Minimum gas required for processing a memo
	Bytes                U64        `json:"bytes"`                  // a_o: Total bytes stored by the service
	DepositOffset        U64        `json:"deposit_offset"`         // a_f: Offset of storage footprint only above which a minimum deposit is needed.
	Items                U32        `json:"items"`                  // a_i: Number of items stored by the service
	CreationSlot         TimeSlot   `json:"creation_slot"`          // a_r: Creation time slot
	LastAccumulationSlot TimeSlot   `json:"last_accumulation_slot"` // a_a: Most recent accumulation time slot
	ParentService        ServiceID  `json:"parent_service"`         // a_p: Parent service identifier
}

// GP §9.8
type ServiceAccountDerivatives struct {
	Items      U32 `json:"items"` // a_i: Number of items stored by the service
	Bytes      U64 `json:"bytes"` // a_o: Total bytes stored by the service
	Minbalance U64 // a_t: Threshold balance of the service in terms of storage footprint
}

type MetaCode struct {
	Metadata ByteSequence
	Code     ByteSequence
}

// ============================================================================
// AVAILABILITY ASSIGNMENTS
// Types related to the assignment of work reports to cores and tracking their
// availability status.
// ============================================================================

// GP §11.1, $\rho$
type (
	// Assignment of a work report with reported timeslot
	AvailabilityAssignment struct {
		Report       WorkReport `json:"report"`  // $\mathbf{r}$: The work report being assigned
		AssignedSlot TimeSlot   `json:"timeout"` // $t$: Reported timeslot of the work report
	}

	// Optional availability assignment (Some/None pattern)
	AvailabilityAssignmentsItem *AvailabilityAssignment

	// Assignments for all cores in the system
	AvailabilityAssignments []AvailabilityAssignmentsItem
)

func (a *AvailabilityAssignment) Validate() error {
	if err := a.Report.Validate(); err != nil {
		return fmt.Errorf("AvailabilityAssignment Report validation failed: %v", err)
	}

	if a.AssignedSlot == 0 {
		return errors.New("AvailabilityAssignment AssignedSlot cannot be 0")
	}

	return nil
}

func (assignments *AvailabilityAssignments) Validate() error {
	if len(*assignments) != CoresCount {
		return fmt.Errorf("AvailabilityAssignments length %d is not equal to CoresCount %d", len(*assignments), CoresCount)
	}
	return nil
}

// Context for the refinement process when executing a work package
// GP §11.4, $\mathbb{C}$
type RefineContext struct {
	Anchor           HeaderHash   `json:"anchor"`             // $a$: Anchor block hash
	StateRoot        StateRoot    `json:"state_root"`         // $s$: Posterior state root at the anchor
	BeefyRoot        BeefyRoot    `json:"beefy_root"`         // $b$: Posterior BEEFY consensus root ( accumulation output log super-peak )
	LookupAnchor     HeaderHash   `json:"lookup_anchor"`      // $l$: Block hash for preimage lookups
	LookupAnchorSlot TimeSlot     `json:"lookup_anchor_slot"` // $t$: Time slot of the lookup anchor
	Prerequisites    []OpaqueHash `json:"prerequisites"`      // $\mathbf{p}$: Hashes of prerequisite work packages
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

// ============================================================================
// AUTHORIZATION
// Types related to the authorization of work packages.
// ============================================================================

// Entity that can authorize work packages
// GP §14.2
type Authorizer struct {
	CodeHash OpaqueHash   `json:"code_hash"` // code-hash: Hash of the authorizer's code
	Params   ByteSequence `json:"params"`    // params: Parameters for the authorizer
}

// Hash of encoded Authorizer
// GP §8.1
type AuthorizerHash OpaqueHash

// Pool of authorizer hashes for a core
// GP §8.1
type AuthPool []AuthorizerHash

func (a *AuthPool) Validate() error {
	if len(*a) > AuthPoolMaxSize {
		return fmt.Errorf("AuthPool length %d is greater than AuthPoolMaxSize %d", len(*a), AuthPoolMaxSize)
	}
	return nil
}

func (a *AuthPool) RemoveLeftMostPairedValue(h OpaqueHash) {
	result := (*a)[:0]
	removed := false
	for _, v := range *a {
		if removed || !bytes.Equal(v[:], h[:]) {
			result = append(result, v)
		} else {
			removed = true
		}
	}
	*a = result
}

// Pools of authorizers for all cores
// GP §8.1
type AuthPools []AuthPool

func (a *AuthPools) Validate() error {
	if len(*a) != CoresCount {
		return fmt.Errorf("AuthPools length %d is not equal to CoresCount %d", len(*a), CoresCount)
	}

	for _, pool := range *a {
		err := pool.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// Queue of authorizer hashes for a core
// GP §8.1
type AuthQueue []AuthorizerHash

func (a *AuthQueue) Validate() error {
	// (8.1) φ ∈ ⟦⟦H⟧_Q⟧_C
	if len(*a) != AuthQueueSize {
		return fmt.Errorf("AuthQueue length %d is not equal to AuthQueueSize %d", len(*a), AuthQueueSize)
	}
	return nil
}

// Queues of authorizers for all cores
// GP §8.1
type AuthQueues []AuthQueue

func (a *AuthQueues) Validate() error {
	// (8.1) φ ∈ ⟦⟦H⟧_Q⟧_C
	if len(*a) != CoresCount {
		return fmt.Errorf("AuthQueues length %d is not equal to CoresCount %d", len(*a), CoresCount)
	}

	for _, queue := range *a {
		err := queue.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

/*
	--- Work Packages and Items ---
	Types defining work packages, which are sequences of Work Items, the units
	of computation that can be submitted to the JAM protocol.
*/

type (
	ExportSegment       [SegmentSize]byte // segment of data exported from a work item
	ExportSegmentMatrix [][]ExportSegment // matrix of exported segments
	OpaqueHashMatrix    [][]OpaqueHash    // matrix of opaque hashes
)

// Specification for importing a segment
// GP §14.3
type ImportSpec struct {
	TreeRoot OpaqueHash `json:"tree_root"` // tree-root: Root hash of the segment tree
	Index    U16        `json:"index"`     // index: Index of the segment
}

// Specification for an extrinsic
// GP §14.3
type ExtrinsicSpec struct {
	Hash OpaqueHash `json:"hash"` // hash: Hash of the extrinsic
	Len  U32        `json:"len"`  // len: Length of the extrinsic in bytes
}

// Individual work item within a package
// GP §14.3, $\mathbb{I}$
type WorkItem struct {
	Service            ServiceID       `json:"service"`              // $s$: Service ID that will process this item
	CodeHash           OpaqueHash      `json:"code_hash"`            // $c$: Hash of the code to execute
	RefineGasLimit     Gas             `json:"refine_gas_limit"`     // $g$: Gas limit for refinement phase
	AccumulateGasLimit Gas             `json:"accumulate_gas_limit"` // $a$: Gas limit for accumulation phase
	ExportCount        U16             `json:"export_count"`         // $e$: Number of exported segments
	Payload            ByteSequence    `json:"payload"`              // $\mathbf{y}$: Input payload for the work item
	ImportSegments     []ImportSpec    `json:"import_segments"`      // $\mathbf{i}$: Segments to import
	Extrinsic          []ExtrinsicSpec `json:"extrinsic"`            // $\mathbf{x}$: Extrinsics to include
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

// Complete work package containing multiple work items
// GP §14.2, $\mathbb{P}$
type WorkPackage struct {
	AuthCodeHost     ServiceID     `json:"auth_code_host"`    // $h$: Service ID hosting the authorization code
	AuthCodeHash     OpaqueHash    `json:"auth_code_hash"`    // $u$: Hash of the authorizer's code
	Context          RefineContext `json:"context"`           // $c$: Refinement context
	Authorization    ByteSequence  `json:"authorization"`     // $j$: Authorization data
	AuthorizerConfig ByteSequence  `json:"authorizer_config"` // $f$: Parameters for the authorizer
	Items            []WorkItem    `json:"items"`             // $w$: Work items to process (1-16)
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
	// (14.2) Validate the number of work items (1-16)
	if len(wp.Items) < 1 || len(wp.Items) > MaximumWorkItems {
		return fmt.Errorf("WorkPackage items length %d is not between 1 and %d", len(wp.Items), MaximumWorkItems)
	}

	totalSize := len(wp.Authorization) + len(wp.AuthorizerConfig)
	totalImportSegments := 0
	totalExportSegments := 0
	totalExtrinsics := 0

	for _, item := range wp.Items {
		totalSize += len(item.Payload)

		totalImportSegments += len(item.ImportSegments)
		totalSize += len(item.ImportSegments) * SegmentFootprint

		for _, extrinsic := range item.Extrinsic {
			totalSize += int(extrinsic.Len)
		}

		totalExportSegments += int(item.ExportCount)

		totalExtrinsics += len(item.Extrinsic)
	}

	// total size check (14.5)
	if totalSize > MaxTotalSize {
		return fmt.Errorf("total size %d is greater than MaxTotalSize %d", totalSize, MaxTotalSize)
	}

	// import segment count check （14.4)
	if totalImportSegments > MaxImportCount {
		return fmt.Errorf("total import segments %d is greater than MaxImportCount %d", totalImportSegments, MaxImportCount)
	}

	// export segment count check (14.4)
	if totalExportSegments > MaxExportCount {
		return fmt.Errorf("total export segments %d is greater than MaxExportCount %d", totalExportSegments, MaxExportCount)
	}

	// extrinsics count check (14.4)
	if totalExtrinsics > MaxExtrinsics {
		return fmt.Errorf("total extrinsics %d is greater than MaxExtrinsics %d", totalExtrinsics, MaxExtrinsics)
	}

	// gas limit check (14.7)
	var totalRefineGas, totalAccumulateGas Gas
	for _, item := range wp.Items {
		totalRefineGas += item.RefineGasLimit
		totalAccumulateGas += item.AccumulateGasLimit
	}

	if totalRefineGas > Gas(MaxRefineGas) {
		return fmt.Errorf("refine gas limit %d is greater than MaxRefineGas %d", totalRefineGas, MaxRefineGas)
	}
	if totalAccumulateGas > MaxAccumulateGas {
		return fmt.Errorf("accumulate gas limit %d is greater than MaxAccumulateGas %d", totalAccumulateGas, MaxAccumulateGas)
	}

	return nil
}

/*
	--- Work Report ---
	Types defining work reports, which contain the results of executing work packages.
*/

// Result of executing a work item
// GP §11.7
type WorkExecResultType string

const (
	WorkExecResultOk             WorkExecResultType = "ok"
	WorkExecResultOutOfGas                          = "out-of-gas"
	WorkExecResultPanic                             = "panic"
	WorkExecResultBadExports                        = "bad-exports"
	WorkExecResultReportOversize                    = "output-oversize"
	WorkExecResultBadCode                           = "bad-code"
	WorkExecResultCodeOversize                      = "code-oversize"
)

type WorkExecResult struct {
	Type WorkExecResultType
	Data []byte // only meaningful when Type == WorkExecResultOk
}

func GetWorkExecResult(resultType WorkExecResultType, data []byte) WorkExecResult {
	if resultType == WorkExecResultOk {
		return WorkExecResult{Type: resultType, Data: data}
	}

	return WorkExecResult{Type: resultType, Data: nil}
}

// Resource usage during refinement for a core
// GP §11.6, part of $\mathbb{D}$
type RefineLoad struct {
	GasUsed        Gas `json:"gas_used"`        // $u$: Gas used during refinement
	Imports        U16 `json:"imports"`         // $i$: Number of import segments from D3L processed
	ExtrinsicCount U16 `json:"extrinsic_count"` // $x$: Number of extrinsics processed
	ExtrinsicSize  U32 `json:"extrinsic_size"`  // $z$: Total size of extrinsics in bytes
	Exports        U16 `json:"exports"`         // $e$: Number of export segments generated into D3L
}

// Result(digest) of executing a single work item
// GP §11.6, $\mathbb{D}$
type WorkResult struct {
	ServiceID     ServiceID      `json:"service_id"`     // $s$: Service ID that processed this item
	CodeHash      OpaqueHash     `json:"code_hash"`      // $c$: Hash of the code that was executed
	PayloadHash   OpaqueHash     `json:"payload_hash"`   // $y$: Hash of the input payload
	AccumulateGas Gas            `json:"accumulate_gas"` // $g$: Gas used during accumulation
	Result        WorkExecResult `json:"result"`         // $\mathbf{l}$: Execution result
	RefineLoad    RefineLoad     `json:"refine_load"`    // ASN.1 specific field: Resource usage during refinement
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

// Availability specification of a work package
// GP §11.5, $\mathbb{Y}$
type WorkPackageSpec struct {
	Hash         WorkPackageHash `json:"hash"`          // $p$: Hash of the work package
	Length       U32             `json:"length"`        // $l$: Length of the work package in bytes
	ErasureRoot  ErasureRoot     `json:"erasure_root"`  // $u$: Root hash of erasure-coded data
	ExportsRoot  ExportsRoot     `json:"exports_root"`  // $e$: Root hash of exported data
	ExportsCount U16             `json:"exports_count"` // $n$: Number of exports
}

// Mapping between work package hash and segment tree root
type SegmentRootLookupItem struct {
	WorkPackageHash WorkPackageHash `json:"work_package_hash"` // work-package-hash: Hash of the work package
	SegmentTreeRoot OpaqueHash      `json:"segment_tree_root"` // segment-tree-root: Root hash of the segment tree
}

// Segment root lookups map
type SegmentRootLookup []SegmentRootLookupItem // segment-tree-root

// Complete report of work package execution
// GP §11.2, $\mathbb{R}$
type WorkReport struct {
	PackageSpec       WorkPackageSpec   `json:"package_spec"`        // $\mathbf{s}$: Specification of the work package
	Context           RefineContext     `json:"context"`             // $\mathbf{c}$: Refinement context
	CoreIndex         CoreIndex         `json:"core_index"`          // $c$: Index of the core that executed this work
	AuthorizerHash    OpaqueHash        `json:"authorizer_hash"`     // $a$: Hash of the authorizer
	AuthGasUsed       Gas               `json:"auth_gas_used"`       // $g$: Gas used during authorization
	AuthOutput        ByteSequence      `json:"auth_output"`         // $\mathbf{t}$: Output from the authorization process
	SegmentRootLookup SegmentRootLookup `json:"segment_root_lookup"` // $\mathbf{l}$: Segment root lookups
	Results           []WorkResult      `json:"results"`             // $\mathbf{d}$: Results of executing each work item (1-16)
}

func (w *WorkReport) Validate() error {
	if len(w.Results) < 1 {
		logger.Warnf("WorkReport Results must have at least 1 item, but got %d", len(w.Results))
		return errors.New("missing_work_results")
	}

	if len(w.Results) > MaximumWorkItems {
		logger.Warnf("WorkReport Results must have at most %d items, but got %d", MaximumWorkItems, len(w.Results))
		return errors.New("too_many_work_results")
	}

	return nil
}

// ValidateLookupDictAndPrerequisites checks the number of SegmentRootLookup and Prerequisites < J
// GP §11.3
func (w *WorkReport) ValidateLookupDictAndPrerequisites() error {
	if len(w.SegmentRootLookup)+len(w.Context.Prerequisites) > MaximumDependencyItems {
		logger.Warnf("SegmentRootLookup and Prerequisites must have a total at most %d, but got %d", MaximumDependencyItems, len(w.SegmentRootLookup)+len(w.Context.Prerequisites))
		return errors.New("too_many_dependencies")
	}
	return nil
}

// ValidateOutputSize checks the total size of the output
// GP §11.8
func (w *WorkReport) ValidateOutputSize() error {
	totalSize := len(w.AuthOutput)
	for _, result := range w.Results {
		// only compute $\mathcal{B}$ => ok
		if result.Result.Type == WorkExecResultOk {
			totalSize += len(result.Result.Data)
		}
	}

	if totalSize > WorkReportOutputBlobsMaximumSize {
		logger.Warnf("total size %d is greater than WorkReportOutputBlobsMaximumSize %d", totalSize, WorkReportOutputBlobsMaximumSize)
		return errors.New("work_report_too_big")
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

// ============================================================================
// BLOCK HISTORY
// Types for tracking the history of blocks and their associated work reports.
// ============================================================================

// Optional Merkle Mountain Range peak
type MmrPeak *OpaqueHash

// Merkle Mountain Range structure (Beefy Belt)
// GP §7.3
type Mmr struct {
	Peaks []MmrPeak `json:"peaks"` // peaks: Sequence of MMR peaks
}

// Information about a work package that was reported
type ReportedWorkPackage struct {
	Hash        WorkReportHash `json:"hash"`         // hash: Hash of the work report
	ExportsRoot ExportsRoot    `json:"exports_root"` // exports-root: Root hash of exported data
}

// Information about a block
type BlockInfo struct {
	HeaderHash HeaderHash            `json:"header_hash"` // $h$: Hash of the block header
	BeefyRoot  OpaqueHash            `json:"beefy_root"`  // $b$: Merkle Mountain Range root
	StateRoot  StateRoot             `json:"state_root"`  // $s$: Posterior state root
	Reported   []ReportedWorkPackage `json:"reported"`    // $\mathbf{p}$: Work packages reported in this block (...Cores)
}

func (b *BlockInfo) Validate() error {
	if len(b.Reported) > CoresCount {
		return fmt.Errorf("BlockInfo Reported length %d is greater than CoresCount %d", len(b.Reported), CoresCount)
	}
	return nil
}

// History of recent blocks
// GP §7.2
type BlocksHistory []BlockInfo

func (b *BlocksHistory) Validate() error {
	if len(*b) > MaxBlocksHistory {
		return fmt.Errorf("BlocksHistory length %d is greater than MaxBlocksHistory %d", len(*b), MaxBlocksHistory)
	}
	for _, blockInfo := range *b {
		if err := blockInfo.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Imported Blocks Information
// GP §7.1
type RecentBlocks struct {
	History BlocksHistory `json:"history"` // history: Recent blocks history
	Mmr     Mmr           `json:"mmr"`     // mmr: MMR
}

// ============================================================================
// STATISTICS
// Types for tracking various statistics about validators, cores, and services.
// ============================================================================

// Record of a validator's activity (per epoch)
// GP §13.1, $\pi_V$, $\pi_L$
type ValidatorActivityRecord struct {
	Blocks        U32 `json:"blocks"`          // $b$: Number of blocks produced
	Tickets       U32 `json:"tickets"`         // $t$: Number of Safrole tickets consumed
	PreImages     U32 `json:"pre_images"`      // $p$: Number of pre-images provided
	PreImagesSize U32 `json:"pre_images_size"` // $d$: Total size of provided pre-images in bytes
	Guarantees    U32 `json:"guarantees"`      // $g$: Number of guarantees provided
	Assurances    U32 `json:"assurances"`      // $a$: Number of assurances provided
}

// Statistics for all validators
type ValidatorsStatistics []ValidatorActivityRecord

func (a *ValidatorsStatistics) Validate() error {
	if len(*a) != ValidatorsCount {
		return fmt.Errorf("ActivityRecords length %d is not equal to ValidatorsCount %d", len(*a), ValidatorsCount)
	}
	return nil
}

// Record of a per-block core's activity
// GP §13.6, $\pi_C$
type CoreActivityRecord struct {
	DALoad         U32 `json:"da_load"`         // $d$: Total bytes written in the Data Availability (DA) layer, includes work bundle, extrinsic, imports and exported segments
	Popularity     U16 `json:"popularity"`      // $p$: Number of validators which formed super-majority for assurance
	Imports        U16 `json:"imports"`         // $i$: Number of segments imported from DA for block processing
	ExtrinsicCount U16 `json:"extrinsic_count"` // $x$: Total number of extrinsics for reported work
	ExtrinsicSize  U32 `json:"extrinsic_size"`  // $z$: Total size of extrinsics for reported work
	Exports        U16 `json:"exports"`         // $e$: Number of segments exported to DA during block processing
	BundleSize     U32 `json:"bundle_size"`     // $l$: Serialized work bundle size written to DA
	GasUsed        Gas `json:"gas_used"`        // $u$: Total gas consumed during block processing (includes refinements and authorizations)
}

// Statistics for all cores
type CoresStatistics []CoreActivityRecord

func (c *CoresStatistics) Validate() error {
	if len(*c) != CoresCount {
		return fmt.Errorf("CoresStatisitics length %d is not equal to CoresCount %d", len(*c), CoresCount)
	}
	return nil
}

// Record of a service's activity
// GP §13.7, $\pi_S$
type ServiceActivityRecord struct {
	ProvidedCount     U16 `json:"provided_count"`      // $p$: Number of preimages provided to this service
	ProvidedSize      U32 `json:"provided_size"`       // $p$: Total size of preimages provided to this service
	RefinementCount   U32 `json:"refinement_count"`    // $r$: Number of work-items refined
	RefinementGasUsed Gas `json:"refinement_gas_used"` // $r$: Amount of gas used for refinement
	Imports           U32 `json:"imports"`             // $i$: Number of segments imported from the DL
	ExtrinsicCount    U32 `json:"extrinsic_count"`     // $x$: Total number of extrinsics used
	ExtrinsicSize     U32 `json:"extrinsic_size"`      // $z$: Total size of extrinsics used
	Exports           U32 `json:"exports"`             // $e$: Number of segments exported into the DL
	AccumulateCount   U32 `json:"accumulate_count"`    // $a$: Number of work-items accumulated
	AccumulateGasUsed Gas `json:"accumulate_gas_used"` // $a$: Amount of gas used for accumulation
}

// Statistics for all services
// Map entry for service statistics
type ServicesStatistics map[ServiceID]ServiceActivityRecord

// Complete statistics for the system
// GP §13.1, $\pi$
type Statistics struct {
	ValsCurr ValidatorsStatistics `json:"vals-curr,omitempty"` // $\pi_V$: Current validator statistics
	ValsLast ValidatorsStatistics `json:"vals-last,omitempty"` // $\pi_L$: Previous validator statistics
	Cores    CoresStatistics      `json:"cores,omitempty"`     // $\pi_C$: Core statistics
	Services ServicesStatistics   `json:"services,omitempty"`  // $\pi_S$: Service statistics
}

// ============================================================================
// TICKETS
// Types related to the Safrole ticket system used for validator selection.
// ============================================================================

type (
	TicketID      OpaqueHash // Unique identifier for a ticket
	TicketAttempt U64        // Attempt counter for ticket submissions
)

// Envelope containing a ticket submission
type TicketEnvelope struct {
	Attempt   TicketAttempt                `json:"attempt"`   // attempt: Attempt number
	Signature BandersnatchRingVrfSignature `json:"signature"` // signature: Ring VRF signature (VRF output maps to a TicketID)
}

// Body of a ticket
// GP §6.6, $\mathbb{T}$
type TicketBody struct {
	ID      TicketID      `json:"id"`      // $y$: Ticket identifier
	Attempt TicketAttempt `json:"attempt"` // $e$: Attempt number
}

// Accumulator for tickets within an epoch
// GP §6.5, $\gamma_A$
type TicketsAccumulator []TicketBody

func (t *TicketsAccumulator) Validate() error {
	if len(*t) > EpochLength {
		return fmt.Errorf("TicketsAccumulator length %d is greater than EpochLength %d", len(*t), EpochLength)
	}
	return nil
}

// Either tickets or public keys
// GP §6.5, $\gamma_S$
type TicketsOrKeys struct {
	Tickets []TicketBody         `json:"tickets,omitempty"` // $\mathbb{T}$: Sequence of ticket bodies
	Keys    []BandersnatchPublic `json:"keys,omitempty"`    // $\tilde{\mathbb{H}}$: Sequence of public keys
}

func (t *TicketsOrKeys) Validate() error {
	if len(t.Tickets) > 0 && len(t.Tickets) != EpochLength {
		return fmt.Errorf("TicketsOrKeys Tickets length %d is not equal to EpochLength %d", len(t.Tickets), EpochLength)
	}

	if len(t.Keys) > 0 && len(t.Keys) != EpochLength {
		return fmt.Errorf("TicketsOrKeys Keys length %d is not equal to EpochLength %d", len(t.Keys), EpochLength)
	}
	if len(t.Tickets) == 0 && len(t.Keys) == 0 {
		return fmt.Errorf("TicketsOrKeys Tickets and Keys are empty")
	}
	return nil
}

// Extrinsic containing ticket submissions
// GP §6.29, $\mathbf{E}_T$
type TicketsExtrinsic []TicketEnvelope

func (t *TicketsExtrinsic) Validate() error {
	if len(*t) > MaxTicketsPerBlock {
		return fmt.Errorf("TicketsExtrinsic length %d is greater than MaxTicketsPerBlock %d", len(*t), MaxTicketsPerBlock)
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

// ============================================================================
// DISPUTES
// Types related to the dispute resolution system.
// ============================================================================

// A single validator's judgment on a disputed work report
type Judgement struct {
	Vote      bool             `json:"vote"`      // $v$: True for valid, False for invalid
	Index     ValidatorIndex   `json:"index"`     // $i$: Index of the validator making the judgment
	Signature Ed25519Signature `json:"signature"` // $s$: Signature of the validator
}

// A complete verdict on a disputed work report
// GP §10.2, $\mathbf{E}_V$
type Verdict struct {
	Target WorkReportHash `json:"target"` // $r$: Hash of the disputed work report
	Age    U32            `json:"age"`    // $a$: Epoch index of the prior state or one less depending on the key set used to sign the votes
	Votes  []Judgement    `json:"votes"`  // $\mathbf{j}$: Judgments from a super-majority of validators
}

func (v *Verdict) Validate() error {
	if len(v.Votes) != ValidatorsSuperMajority {
		return fmt.Errorf("verdict Votes length %d is not equal to ValidatorsSuperMajority %d", len(v.Votes), ValidatorsSuperMajority)
	}
	return nil
}

// Information about a validator who submitted an invalid report
// GP §10.2, $\mathbf{E}_C$
type Culprit struct {
	Target    WorkReportHash   `json:"target"`    // $r$: Hash of the disputed work report
	Key       Ed25519Public    `json:"key"`       // $f$: Public key of the culprit
	Signature Ed25519Signature `json:"signature"` // $s$: Signature proving culpability
}

// Information about a validator who made an incorrect judgment
// GP §10.2, $\mathbf{E}_F$
type Fault struct {
	Target    WorkReportHash   `json:"target"`    // $r$: Hash of the disputed work report
	Vote      bool             `json:"vote"`      // $v$: The incorrect vote (True/False)
	Key       Ed25519Public    `json:"key"`       // $f$: Public key of the validator
	Signature Ed25519Signature `json:"signature"` // $s$: Signature proving the fault
}

// Records of dispute outcomes
// GP §10.1
type DisputesRecords struct {
	Good      []WorkReportHash `json:"good"`      // $psi_g$: Reports deemed valid
	Bad       []WorkReportHash `json:"bad"`       // $psi_b$: Reports deemed invalid
	Wonky     []WorkReportHash `json:"wonky"`     // $psi_w$: Reports with conflicting judgments
	Offenders []Ed25519Public  `json:"offenders"` // $psi_o$: Validators who submitted invalid reports
}

// Extrinsic containing dispute information
// GP §10.2, $E_D$ ≡ (v,c,f)
type DisputesExtrinsic struct {
	Verdicts []Verdict `json:"verdicts"` // verdicts: Verdicts on disputed items
	Culprits []Culprit `json:"culprits"` // culprits: Information about culprits
	Faults   []Fault   `json:"faults"`   // faults: Information about faulty judgments
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

// ============================================================================
// PREIMAGES
// Types related to preimages blobs provided to services.
// ============================================================================

// A preimage with its requester
// GP §12.35, $\mathbf{E}_P$
type Preimage struct {
	Requester ServiceID    `json:"requester"` // $s$: ID of the service requesting the preimage
	Blob      ByteSequence `json:"blob"`      // $\mathbf{d}$: The preimage data
}

func (p *Preimage) Validate() error {
	return nil
}

// Extrinsic containing preimages
// GP §12.35, $\mathbf{E}_P$
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

// ============================================================================
// ASSURANCES
// Types related to data availability assurances provided by validators.
// ============================================================================

// Assurance of data availability from a validator
// GP §11.2.1
type AvailAssurance struct {
	Anchor         HeaderHash       `json:"anchor"`          // $s$: The block to which this availability attestation corresponds
	Bitfield       Bitfield         `json:"bitfield"`        // $f$: The bitfield of whether any given core's reported package at anchor block is helped made available by this validator.
	ValidatorIndex ValidatorIndex   `json:"validator_index"` // $v$: Index of the validator providing the assurance
	Signature      Ed25519Signature `json:"signature"`       // $s$: Signature of the validator proving the assurance
}

func (a *AvailAssurance) Validate() error {
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
		return Bitfield{}, fmt.Errorf("Bitfield length %d is not equal to AvailBitfieldBytes %d", len(bytes), AvailBitfieldBytes)
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

func (bf *Bitfield) Validate() error {
	if len(*bf) != CoresCount {
		return fmt.Errorf("Bitfield length %d is not equal to CoresCount %d", len(*bf), CoresCount)
	}

	return nil
}

// Extrinsic containing assurances from validators
// GP §11.10, $\mathbf{E}_A$
type AssurancesExtrinsic []AvailAssurance

func (a *AssurancesExtrinsic) Validate() error {
	if len(*a) > ValidatorsCount {
		return fmt.Errorf("AssurancesExtrinsic length %d is greater than ValidatorsCount %d", len(*a), ValidatorsCount)
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

// ============================================================================
// GUARANTEES
// Types related to guarantees provided by validators for work reports.
// ============================================================================

// Credential: Signature from a validator
// GP §11.25 $a$
type ValidatorSignature struct {
	ValidatorIndex ValidatorIndex   `json:"validator_index"` // $v$: Index of the validator
	Signature      Ed25519Signature `json:"signature"`       // $s$: Signature of the validator
}

func (v *ValidatorSignature) Validate() error {
	if int(v.ValidatorIndex) >= ValidatorsCount {
		// return fmt.Errorf("ValidatorIndex %v must be less than %v", v.ValidatorIndex, ValidatorsCount)
		return errors.New("bad_validator_index")
	}
	return nil
}

func ValidateGuaranteeSignatures(sigs []ValidatorSignature) error {
	count := len(sigs)
	if count < GuaranteeMinCount {
		return errors.New("insufficient_guarantees")
	}
	if count > GuaranteeMaxCount {
		return errors.New("too_many_guarantees")
	}
	for i := range sigs {
		if err := sigs[i].Validate(); err != nil {
			return fmt.Errorf("guarantee signature[%d]: %w", i, err)
		}
	}
	return nil
}

// Guarantee for a work report
// GP §11.23
type ReportGuarantee struct {
	Report     WorkReport           `json:"report"`     // $\mathbf{r}$: The work report being guaranteed
	Slot       TimeSlot             `json:"slot"`       // $t$: Time slot following production of the report
	Signatures []ValidatorSignature `json:"signatures"` // $a$: Signatures from validators
}

func (r *ReportGuarantee) Validate() error {
	if err := r.Report.Validate(); err != nil {
		return err
	}

	if len(r.Signatures) < GuaranteeMinCount {
		return errors.New("insufficient_guarantees")
	}

	if len(r.Signatures) > GuaranteeMaxCount {
		logger.Warn("too_many_guarantees")
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

// Extrinsic containing guarantees for work reports
// GP §11.23
type GuaranteesExtrinsic []ReportGuarantee

// (11.23)
func (g *GuaranteesExtrinsic) Validate() error {
	if len(*g) > CoresCount {
		return fmt.Errorf("GuaranteesExtrinsic length %d is greater than CoresCount %d", len(*g), CoresCount)
	}
	for i, report := range *g {
		if err := report.Validate(); err != nil {
			return fmt.Errorf("GuaranteesExtrinsic report[%d] validation failed: %w", i, err)
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

// ============================================================================
// ACCUMULATION
// Types related to the accumulation of work reports and tracking of
// dependencies between work packages.
// ============================================================================

// Record of a work report that is ready for processing
type ReadyRecord struct {
	Report       WorkReport        // $\mathbb{R}$: The work report
	Dependencies []WorkPackageHash // ${\mathbb{H}}$: Dependencies required by this report
}

// A group of records that became ready at a specific slot
type ReadyQueueItem []ReadyRecord

// A fixed-size FIFO queue of ready record groups (one per slot), with maximum size equal to the epoch length
// GP §12.3, $w$
type ReadyQueue []ReadyQueueItem

func (r *ReadyQueue) Validate() error {
	if len(*r) != EpochLength {
		return fmt.Errorf("ReadyQueue length %d is not equal to EpochLength %d", len(*r), EpochLength)
	}
	return nil
}

// A group of accumulated work packages for a specific slot
type AccumulatedQueueItem []WorkPackageHash

// A fixed-size FIFO queue of accumulated work package groups (one per slot), with maximum size equal to the epoch length
// GP §12.1, $\xi$
type AccumulatedQueue []AccumulatedQueueItem

func (a *AccumulatedQueue) Validate() error {
	if len(*a) != EpochLength {
		return fmt.Errorf("AccumulatedQueue length %d is not equal to EpochLength %d", len(*a), EpochLength)
	}
	return nil
}

// Map of services that are always accumulated
type AlwaysAccumulateMap map[ServiceID]Gas

// Entry in the always-accumulate map
type AlwaysAccumulateMapDTO struct {
	ServiceID ServiceID `json:"id"`  // id: Service ID
	Gas       Gas       `json:"gas"` // gas: Gas limit for accumulation
}

// Special privileges for certain services
// GP §9.9~9.11, $\chi$
type Privileges struct {
	Bless       ServiceID           `json:"chi_m"` // bless: Service ID with blessing privileges
	Designate   ServiceID           `json:"chi_v"` // designate: Service ID with designation privileges
	CreateAcct  ServiceID           `json:"chi_r"` // register: Service ID with registrar privileges
	Assign      ServiceIDList       `json:"chi_a"` // assign: Service ID with assignment privileges
	AlwaysAccum AlwaysAccumulateMap `json:"chi_g"` // always-acc: Services that are always accumulated
}

// (12.16) \mathbb{S}, states that are needed and mutable by the accumulation process
type PartialStateSet struct {
	ServiceAccounts ServiceAccountState // $\mathbf{d}$: Service accounts
	ValidatorKeys   ValidatorsData      // $\mathbf{i}$: Upcoming validator keys $\iota$
	Authorizers     AuthQueues          // $\mathbf{q}$: The queues of authorizers $\varphi$
	// Privileges state $\chi$
	Bless       ServiceID           // $m$: Bless
	Assign      ServiceIDList       // $\mathbf{a}$: Assign
	Designate   ServiceID           // $v$: Designate
	CreateAcct  ServiceID           // $r$: Create account
	AlwaysAccum AlwaysAccumulateMap // $\mathbf{z}$: Always accumulate
}

func (origin *PartialStateSet) DeepCopy() PartialStateSet {
	// ServiceAccountState
	// Pre-allocate capacity for service accounts map
	copiedServiceAccounts := make(ServiceAccountState, len(origin.ServiceAccounts))
	for serviceID, originAccount := range origin.ServiceAccounts {
		var copiedAccount ServiceAccount
		copiedAccount.ServiceInfo = originAccount.ServiceInfo
		// use shallow copy for PreimageLookup since host call only delete preimage item, no modification
		copiedAccount.PreimageLookup = maps.Clone(originAccount.PreimageLookup)
		copiedAccount.LookupDict = make(LookupMetaMapEntry)
		for k, v := range originAccount.LookupDict {
			copiedAccount.LookupDict[k] = make(TimeSlotSet, len(v))
			copy(copiedAccount.LookupDict[k], v)
		}
		copiedAccount.StorageDict = make(Storage, len(originAccount.StorageDict))
		for storageKey, storageVal := range originAccount.StorageDict {
			copiedStorage := make(ByteSequence, len(storageVal))
			copy(copiedStorage, storageVal)
			copiedAccount.StorageDict[storageKey] = copiedStorage
		}

		copiedServiceAccounts[serviceID] = copiedAccount
	}

	// ValidatorsData
	copiedValidators := make(ValidatorsData, len(origin.ValidatorKeys))
	copy(copiedValidators, origin.ValidatorKeys)
	/*
		for k, v := range origin.ValidatorKeys {
			// array(validator element: BandersnatchPublic, Ed25519Public, BlsPublic, ValidatorMetadata) is value type
			copiedValidators[k] = v
		}
	*/
	// AuthQueues
	copiedAuthorizers := make(AuthQueues, len(origin.Authorizers))
	for authorizerIdx, authorizerValue := range origin.Authorizers {
		copiedAuthorizers[authorizerIdx] = make(AuthQueue, len(authorizerValue))
		copy(copiedAuthorizers[authorizerIdx], authorizerValue)
		/*
			for queueIdx, queueValue := range authorizerValue {
				copiedAuthorizers[authorizerIdx][queueIdx] = queueValue
			}
		*/
	}

	// Assign
	copiedAssign := make(ServiceIDList, len(origin.Assign))
	copy(copiedAssign, origin.Assign)
	/*
		for idx, serviceID := range copiedAssign {
			copiedAssign[idx] = serviceID
		}
	*/
	// AlwaysAccum
	copiedAlwaysAccum := maps.Clone(origin.AlwaysAccum)

	return PartialStateSet{
		ServiceAccounts: copiedServiceAccounts,
		ValidatorKeys:   copiedValidators,
		Authorizers:     copiedAuthorizers,
		Bless:           origin.Bless,
		Assign:          copiedAssign,
		Designate:       origin.Designate,
		CreateAcct:      origin.CreateAcct,
		AlwaysAccum:     copiedAlwaysAccum,
	}
}

// Operand tuple: salient information from digests and reports for accumulation
// GP §12.13, $\mathbb{U}$
type Operand struct {
	Hash           WorkPackageHash // $p$
	ExportsRoot    ExportsRoot     // $e$
	AuthorizerHash OpaqueHash      // $a$
	PayloadHash    OpaqueHash      // $y$
	GasLimit       Gas             // $g$
	Result         WorkExecResult  // $\mathbf{l}$
	AuthOutput     ByteSequence    // $\mathbf{t}$
}

// Deferred transfer
// GP §12.14, $\mathbb{X}$
type DeferredTransfer struct {
	SenderID   ServiceID              // $s$
	ReceiverID ServiceID              // $d$
	Balance    U64                    // $a$
	Memo       [TransferMemoSize]byte // $m$
	GasLimit   Gas                    // $g$
}

// Accumulation input: operand or deferred transfer
// GP §12.15, $\mathbb{I}$
type OperandOrDeferredTransfer struct {
	Operand          *Operand          // $\mathbb{U}$
	DeferredTransfer *DeferredTransfer // $\mathbb{X}$
}

// (12.17) $U$, list of service gas used
type ServiceGasUsedList []ServiceGasUsed

type ServiceGasUsed struct {
	ServiceID ServiceID
	Gas       Gas
}

// Service-indexed commitments to accumulation output
// GP §12.15, $B$
type AccumulatedServiceHash struct {
	ServiceID ServiceID
	Hash      OpaqueHash // AccumulationOutput
}

// v0.6.7 (7.4)
type LastAccOut []AccumulatedServiceHash

// (12.15) B
// INFO:
// - We define (12.15) AccumulatedServiceOutput as a map of AccumulatedServiceHash.
// - We convert the AccumulatedServiceOutput to LastAccOut (a slice of AccumulatedServiceHash) for (7.4) Theta
type AccumulatedServiceOutput map[AccumulatedServiceHash]bool

// (12.28)
type GasAndNumAccumulatedReports struct {
	Gas                   Gas // $\mathbf{u}$: gas used by the service
	NumAccumulatedReports U64 // $n$: number of work-reports accumulated by the service
}

// GP §12.26, $\mathbf{S}$: accumulation statistics
// dictionary<ServiceID, (gas used, the number of work-reports accumulated)>
type AccumulationStatistics map[ServiceID]GasAndNumAccumulatedReports

// ============================================================================
// BLOCK
// Types defining the block structure, including the header and all extrinsics.
// ============================================================================

// Validator keys included in an epoch announcement
// GP §6.27
type EpochMarkValidatorKeys struct {
	Bandersnatch BandersnatchPublic // $k_b$: Bandersnatch public key
	Ed25519      Ed25519Public      // $k_e$: Ed25519 public key
}

// Mark containing the next epoch parameters
// GP §6.27, $\mathbf{H}_E$
type EpochMark struct {
	Entropy        Entropy                  `json:"entropy"`         // $\eta_0$: Entropy for the epoch
	TicketsEntropy Entropy                  `json:"tickets_entropy"` // $\eta_1$: Entropy used to build the epoch's tickets
	Validators     []EpochMarkValidatorKeys `json:"validators"`      // Public keys of validators
}

func (e *EpochMark) Validate() error {
	if len(e.Validators) != ValidatorsCount {
		return fmt.Errorf("EpochMark Validators length %d is not equal to ValidatorsCount %d", len(e.Validators), ValidatorsCount)
	}

	return nil
}

// Mark containing the next epoch tickets
// GP §6.28, $\mathbf{H}_W$
type TicketsMark []TicketBody

func (t *TicketsMark) Validate() error {
	if len(*t) > 0 && len(*t) != EpochLength {
		return fmt.Errorf("TicketsMark length %d is not equal to EpochLength %d", len(*t), EpochLength)
	}
	return nil
}

// Mark containing offenders
// GP §10.20, $\mathbf{H}_O$
type OffendersMark []Ed25519Public

// Block header
// GP §5.1
type Header struct {
	Parent          HeaderHash               `json:"parent"`                 // $H_p$: Hash of the parent block header
	ParentStateRoot StateRoot                `json:"parent_state_root"`      // $H_r$: State root associated to the parent block
	ExtrinsicHash   OpaqueHash               `json:"extrinsic_hash"`         // $H_x$: Hash of the extrinsic data
	Slot            TimeSlot                 `json:"slot"`                   // $H_t$: Time slot of this block
	EpochMark       *EpochMark               `json:"epoch_mark,omitempty"`   // $H_e$: Mark for epoch transition
	TicketsMark     *TicketsMark             `json:"tickets_mark,omitempty"` // $H_w$: Mark for tickets
	AuthorIndex     ValidatorIndex           `json:"author_index"`           // $H_i$: Index of the validator who authored this block
	EntropySource   BandersnatchVrfSignature `json:"entropy_source"`         // $H_v$: Source of entropy for this block
	OffendersMark   OffendersMark            `json:"offenders_mark"`         // $H_o$: Mark for offenders
	Seal            BandersnatchVrfSignature `json:"seal"`                   // $H_s$: Seal signature for this block
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

// Collection of all extrinsics in a block
type Extrinsic struct {
	Tickets    TicketsExtrinsic    `json:"tickets"`    // tickets: Ticket submissions
	Preimages  PreimagesExtrinsic  `json:"preimages"`  // preimages: Preimage submissions
	Guarantees GuaranteesExtrinsic `json:"guarantees"` // guarantees: Work report guarantees
	Assurances AssurancesExtrinsic `json:"assurances"` // assurances: Data availability assurances
	Disputes   DisputesExtrinsic   `json:"disputes"`   // disputes: Dispute resolutions
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

// Complete block structure
type Block struct {
	Header    Header    `json:"header"`    // header: Block header
	Extrinsic Extrinsic `json:"extrinsic"` // extrinsic: Block extrinsics
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
	ServiceID ServiceID
	Blob      []byte
}
type ServiceBlobs []ServiceBlob

// For state serialization, merklization, and reading trace test cases
type StateKey [31]byte

// BoundaryNode represents a node in the state trie for merklization boundary reporting.
type BoundaryNode struct {
	Key    StateKey
	Hash   [32]byte
	Parent *StateKey
	IsLeaf bool
}

type StateKeyVal struct {
	Key   StateKey
	Value ByteSequence
}

type StateKeyVals []StateKeyVal

func (origin *StateKeyVals) DeepCopy() StateKeyVals {
	copiedStateKeyVals := make(StateKeyVals, len(*origin))
	// copy(copiedStateKeyVals, *origin)
	for i := range *origin {
		copiedStateKeyVals[i].Key = (*origin)[i].Key

		copiedValue := make(ByteSequence, len((*origin)[i].Value))
		copy(copiedValue, (*origin)[i].Value)
		copiedStateKeyVals[i].Value = copiedValue
	}

	return copiedStateKeyVals
}

type StateKeyValDiff struct {
	Key           StateKey
	ExpectedValue ByteSequence
	ActualValue   ByteSequence
}

func Some[T any](v T) *T {
	return &v
}

type HashSegmentMap map[OpaqueHash]OpaqueHash

type AncestryItem struct {
	Slot       TimeSlot
	HeaderHash HeaderHash
}

type Ancestry []AncestryItem

type (
	ProtocolParameters struct {
		BI U64 // B_I
		BL U64 // B_L
		BS U64 // B_S
		C  U16 // C
		D  U32 // D
		E  U32 // E
		GA U64 // G_A
		GI U64 // G_I
		GR U64 // G_R
		GT U64 // G_T
		H  U16 // H
		I  U16 // I
		J  U16 // J
		K  U16 // K
		L  U32 // L
		N  U16 // N
		O  U16 // O
		P  U16 // P
		Q  U16 // Q
		R  U16 // R
		T  U16 // T
		U  U16 // U
		V  U16 // V
		WA U32 // W_A
		WB U32 // W_B
		WC U32 // W_C
		WE U32 // W_E
		WM U32 // W_M
		WP U32 // W_P
		WR U32 // W_R
		WT U32 // W_T
		WX U32 // W_X
		Y  U32 // Y
	}
)

// ProtocolParamSnapshot and SnapshotProtocolParams() are for log and debugging purpose
type ProtocolParamSnapshot struct {
	// will not change by chainspec now
	BI uint64 `json:"B_I"` // AdditionalMinBalancePerItem
	BL uint64 `json:"B_L"` // AdditionalMinBalancePerOctet
	BS uint64 `json:"B_S"` // BasicMinBalance
	GA uint64 `json:"G_A"` // MaxAccumulateGas
	GI uint64 `json:"G_I"` // IsAuthorizedGas
	H  uint16 `json:"H"`   // MaxBlocksHistory
	I  uint16 `json:"I"`   // MaximumWorkItems
	J  uint16 `json:"J"`   // MaximumDependencyItems
	O  uint16 `json:"O"`   // AuthPoolMaxSize
	P  uint16 `json:"P"`   // SlotPeriod
	Q  uint16 `json:"Q"`   // AuthQueueSize
	T  uint16 `json:"T"`   // MaxExtrinsics
	U  uint16 `json:"U"`   // WorkReportTimeout
	WA uint32 `json:"W_A"` // MaxIsAuthorizedCodeSize
	WB uint32 `json:"W_B"` // MaxTotalSize
	WC uint32 `json:"W_C"` // MaxServiceCodeSize
	WM uint32 `json:"W_M"` // MaxImportCount
	WR uint32 `json:"W_R"` // WorkReportOutputBlobsMaximumSize
	WT uint32 `json:"W_T"` // TransferMemoSize
	WX uint32 `json:"W_X"` // MaxExportCount

	// will change by chainspec now
	C  uint16 `json:"C"`   // CoresCount
	D  uint32 `json:"D"`   // UnreferencedPreimageTimeslots
	E  uint32 `json:"E"`   // EpochLength
	GR uint64 `json:"G_R"` // MaxRefineGas
	GT uint64 `json:"G_T"` // TotalGas
	K  uint16 `json:"K"`   // MaxTicketsPerBlock
	L  uint32 `json:"L"`   // MaxLookupAge
	N  uint16 `json:"N"`   // TicketsPerValidator
	R  uint16 `json:"R"`   // RotationPeriod
	V  uint16 `json:"V"`   // ValidatorsCount
	WE uint32 `json:"W_E"` // ECBasicSize
	WP uint32 `json:"W_P"` // ECPiecesPerSegment
	Y  uint32 `json:"Y"`   // SlotSubmissionEnd
}

func SnapshotProtocolParams() ProtocolParamSnapshot {
	return ProtocolParamSnapshot{
		// const
		BI: uint64(AdditionalMinBalancePerItem),
		BL: uint64(AdditionalMinBalancePerOctet),
		BS: uint64(BasicMinBalance),
		GA: uint64(MaxAccumulateGas),
		GI: uint64(IsAuthorizedGas),
		H:  uint16(MaxBlocksHistory),
		I:  uint16(MaximumWorkItems),
		J:  uint16(MaximumDependencyItems),
		O:  uint16(AuthPoolMaxSize),
		P:  uint16(SlotPeriod),
		Q:  uint16(AuthQueueSize),
		T:  uint16(MaxExtrinsics),
		U:  uint16(WorkReportTimeout),
		WA: uint32(MaxIsAuthorizedCodeSize),
		WB: uint32(MaxTotalSize),
		WC: uint32(MaxServiceCodeSize),
		WM: uint32(MaxImportCount),
		WR: uint32(WorkReportOutputBlobsMaximumSize),
		WT: uint32(TransferMemoSize),
		WX: uint32(MaxExportCount),

		// var
		C:  uint16(CoresCount),
		D:  uint32(UnreferencedPreimageTimeslots),
		E:  uint32(EpochLength),
		GR: uint64(MaxRefineGas),
		GT: uint64(TotalGas),
		K:  uint16(MaxTicketsPerBlock),
		L:  uint32(MaxLookupAge),
		N:  uint16(TicketsPerValidator),
		R:  uint16(RotationPeriod),
		V:  uint16(ValidatorsCount),
		WE: uint32(ECBasicSize),
		WP: uint32(ECPiecesPerSegment),
		Y:  uint32(SlotSubmissionEnd),
	}
}

func (s ProtocolParamSnapshot) JSON() string {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":"marshal snapshot: %v"}`, err)
	}
	return string(b)
}
