// Reminder: When using jam_types, check if a Validate function exists.
// If a Validate function is available, remember to use it.
// If the desired Validate function is not found, please implement one yourself. :)

package jam_types

import (
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

// Simple

type U8 uint8

type U16 uint16

type U32 uint32

type U64 uint64
type ByteSequence []byte
type ByteArray32 []byte

type BitSequence []bool

// Crypto

type BandersnatchPublic []byte

func (b BandersnatchPublic) Validate() error {
	if len(b) != 32 {
		return errors.New("BandersnatchPublic expected 32 bytes")
	}
	return nil
}

type Ed25519Public []byte

func (e Ed25519Public) Validate() error {
	if len(e) != 32 {
		return errors.New("Ed25519Public expected 32 bytes")
	}
	return nil
}

type BlsPublic []byte

func (e BlsPublic) Validate() error {
	if len(e) != 144 {
		return errors.New("BlsPublic expected 144 bytes")
	}
	return nil
}

type BandersnatchVrfSignature []byte

func (b BandersnatchVrfSignature) Validate() error {
	if len(b) != 96 {
		return errors.New("BandersnatchVrfSignature expected 96 bytes")
	}
	return nil
}

type BandersnatchRingVrfSignature []byte

func (b BandersnatchRingVrfSignature) Validate() error {
	if len(b) != 784 {
		return errors.New("BandersnatchRingVrfSignature expected 784 bytes")
	}
	return nil
}

type Ed25519Signature []byte

func (e Ed25519Signature) Validate() error {
	if len(e) != 64 {
		return errors.New("Ed25519Signature expected 64 bytes")
	}
	return nil
}

// Application Specific Core

type OpaqueHash ByteArray32

type TimeSlot U32
type ValidatorIndex U16
type CoreIndex U16

type HeaderHash OpaqueHash
type StateRoot OpaqueHash
type BeefyRoot OpaqueHash
type WorkPackageHash OpaqueHash
type WorkReportHash OpaqueHash
type ExportsRoot OpaqueHash
type ErasureRoot OpaqueHash

type Gas U64

type Entropy OpaqueHash
type EntropyBuffer []Entropy

func (e EntropyBuffer) Validate() error {
	if len(e) != 4 {
		return errors.New("EntropyBuffer must have exactly 4 Entropies")
	}
	return nil
}

type ValidatorMetadata []byte

func (v ValidatorMetadata) Validate() error {
	if len(v) != 128 {
		return errors.New("ValidatorMetadata expected 128 bytes")
	}
	return nil
}

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

type ServiceInfo struct {
	CodeHash   OpaqueHash `json:"code_hash,omitempty"`
	Balance    U64        `json:"balance,omitempty"`
	MinItemGas Gas        `json:"min_item_gas,omitempty"`
	MinMemoGas Gas        `json:"min_memo_gas,omitempty"`
	Bytes      U64        `json:"bytes,omitempty"`
	Items      U32        `json:"items,omitempty"`
}

// Availability Assignments

type AvailabilityAssignment struct {
	Report  WorkReport `json:"report"`
	Timeout uint32     `json:"timeout,omitempty"`
}

func (a AvailabilityAssignment) Validate() error {
	if err := a.Report.Validate(); err != nil {
		return fmt.Errorf("AvailabilityAssignment Report validation failed: %w", err)
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

// Refine Context

type RefineContext struct {
	Anchor           HeaderHash   `json:"anchor,omitempty"`
	StateRoot        StateRoot    `json:"state_root,omitempty"`
	BeefyRoot        BeefyRoot    `json:"beefy_root,omitempty"`
	LookupAnchor     HeaderHash   `json:"lookup_anchor,omitempty"`
	LookupAnchorSlot TimeSlot     `json:"lookup_anchor_slot,omitempty"`
	Prerequisites    []OpaqueHash `json:"prerequisites,omitempty"`
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

// Authorizations

type Authorizations struct {
	CodeHash  OpaqueHash   `json:"code_hash,omitempty"`
	Signature ByteSequence `json:"signature,omitempty"`
}

type AuthorizerHash OpaqueHash

type AuthPool []AuthorizerHash

func (a AuthPool) Validate() error {
	if len(a) > AuthPoolMaxSize {
		return fmt.Errorf("AuthPool exceeds max-auth-pool-size limit of %d", AuthPoolMaxSize)
	}
	return nil
}

type AuthPools []AuthPool

func (a AuthPools) Validate() error {
	if len(a) > CoresCount {
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
	if len(a) > AuthQueueSize {
		return fmt.Errorf("AuthQueue exceeds max-auth-queue-size limit of %d", AuthQueueSize)
	}
	return nil
}

type AuthQueues []AuthQueue

func (a AuthQueues) Validate() error {
	if len(a) > CoresCount {
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

// Work Package

type ImportSpec struct {
	TreeRoot OpaqueHash `json:"tree_root,omitempty"`
	Index    U16        `json:"index,omitempty"`
}
type ExtrinsicSpec struct {
	Hash OpaqueHash `json:"hash,omitempty"`
	Len  U32        `json:"len,omitempty"`
}

type Authorizer struct {
	CodeHash OpaqueHash   `json:"code_hash,omitempty"`
	Params   ByteSequence `json:"params,omitempty"`
}

type WorkItem struct {
	Service            ServiceId       `json:"service,omitempty"`
	CodeHash           OpaqueHash      `json:"code_hash,omitempty"`
	Payload            ByteSequence    `json:"payload,omitempty"`
	RefineGasLimit     Gas             `json:"refine_gas_limit,omitempty"`
	AccumulateGasLimit Gas             `json:"accumulate_gas_limit,omitempty"`
	ImportSegments     []ImportSpec    `json:"import_segments,omitempty"`
	Extrinsic          []ExtrinsicSpec `json:"extrinsic,omitempty"`
	ExportCount        U16             `json:"export_count,omitempty"`
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

type WorkPackage struct {
	Authorization ByteSequence  `json:"authorization,omitempty"`
	AuthCodeHost  ServiceId     `json:"auth_code_host,omitempty"`
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

func (w WorkPackage) Validate() error {
	if len(w.Items) < 1 || len(w.Items) > 4 {
		return fmt.Errorf("WorkPackage Items must have between 1 and 4 items, but got %d", len(w.Items))
	}
	return nil
}

// Work Report

type WorkExecResultType string

const (
	WorkExecResultOk           WorkExecResultType = "ok"
	WorkExecResultOutOfGas                        = "out-of-gas"
	WorkExecResultPanic                           = "panic"
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

type WorkResult struct {
	ServiceId     ServiceId      `json:"service_id,omitempty"`
	CodeHash      OpaqueHash     `json:"code_hash,omitempty"`
	PayloadHash   OpaqueHash     `json:"payload_hash,omitempty"`
	AccumulateGas Gas            `json:"accumulate_gas,omitempty"`
	Result        WorkExecResult `json:"result,omitempty"`
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

type WorkReport struct {
	PackageSpec       WorkPackageSpec   `json:"package_spec"`
	Context           RefineContext     `json:"context"`
	CoreIndex         CoreIndex         `json:"core_index,omitempty"`
	AuthorizerHash    OpaqueHash        `json:"authorizer_hash,omitempty"`
	AuthOutput        ByteSequence      `json:"auth_output,omitempty"`
	SegmentRootLookup SegmentRootLookup `json:"segment_root_lookup,omitempty"`
	Results           []WorkResult      `json:"results,omitempty"`
}

func (w WorkReport) Validate() error {
	if len(w.Results) < 1 || len(w.Results) > 4 {
		return fmt.Errorf("WorkReport Results must have between 1 and 4 items, but got %d", len(w.Results))
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

// Block History

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

// Tickets

type TicketId OpaqueHash
type TicketAttempt U8

type TicketEnvelope struct {
	Attempt   TicketAttempt                `json:"attempt,omitempty"`
	Signature BandersnatchRingVrfSignature `json:"signature,omitempty"`
}

type TicketBody struct {
	Id      TicketId      `json:"id,omitempty"`
	Attempt TicketAttempt `json:"attempt,omitempty"`
}

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

// Disputes

type Judgement struct {
	Vote      bool             `json:"vote,omitempty"`
	Index     ValidatorIndex   `json:"index,omitempty"`
	Signature Ed25519Signature `json:"signature,omitempty"`
}

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

type Culprit struct {
	Target    WorkReportHash   `json:"target,omitempty"`
	Key       Ed25519Public    `json:"key,omitempty"`
	Signature Ed25519Signature `json:"signature,omitempty"`
}

func (c *Culprit) Validate() error {
	if err := c.Key.Validate(); err != nil {
		return err
	}

	if err := c.Signature.Validate(); err != nil {
		return err
	}

	return nil
}

type Fault struct {
	Target    WorkReportHash   `json:"target,omitempty"`
	Vote      bool             `json:"vote,omitempty"`
	Key       Ed25519Public    `json:"key,omitempty"`
	Signature Ed25519Signature `json:"signature,omitempty"`
}

func (f *Fault) Validate() error {
	if err := f.Key.Validate(); err != nil {
		return err
	}

	if err := f.Signature.Validate(); err != nil {
		return err
	}

	return nil
}

type DisputesRecords struct {
	Good      []WorkReportHash `json:"good,omitempty"`      // Good verdicts (psi_g)
	Bad       []WorkReportHash `json:"bad,omitempty"`       // Bad verdicts (psi_b)
	Wonky     []WorkReportHash `json:"wonky,omitempty"`     // Wonky verdicts (psi_w)
	Offenders []Ed25519Public  `json:"offenders,omitempty"` // Offenders (psi_o)
}

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

	for _, culprit := range d.Culprits {
		if err := culprit.Validate(); err != nil {
			return err
		}
	}

	for _, fault := range d.Faults {
		if err := fault.Validate(); err != nil {
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

// Assurances

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

type EpochMark struct {
	Entropy        Entropy              `json:"entropy,omitempty"`
	TicketsEntropy Entropy              `json:"tickets_entropy,omitempty"`
	Validators     []BandersnatchPublic `json:"validators,omitempty"`
}

func (e EpochMark) Validate() error {
	if len(e.Validators) > ValidatorsCount {
		return fmt.Errorf("EpochMark Validators exceeds maximum size of %d", ValidatorsCount)
	}

	for _, validator := range e.Validators {
		if err := validator.Validate(); err != nil {
			return err
		}
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

type Header struct {
	Parent          HeaderHash               `json:"parent,omitempty"`
	ParentStateRoot StateRoot                `json:"parent_state_root,omitempty"`
	ExtrinsicHash   OpaqueHash               `json:"extrinsic_hash,omitempty"`
	Slot            TimeSlot                 `json:"slot,omitempty"`
	EpochMark       *EpochMark               `json:"epoch_mark,omitempty"`
	TicketsMark     *TicketsMark             `json:"tickets_mark,omitempty"`
	OffendersMark   OffendersMark            `json:"offenders_mark,omitempty"`
	AuthorIndex     ValidatorIndex           `json:"author_index,omitempty"`
	EntropySource   BandersnatchVrfSignature `json:"entropy_source,omitempty"`
	Seal            BandersnatchVrfSignature `json:"seal,omitempty"`
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

	for _, public := range h.OffendersMark {
		if err := public.Validate(); err != nil {
			return err
		}
	}

	if err := h.EntropySource.Validate(); err != nil {
		return err
	}

	if err := h.Seal.Validate(); err != nil {
		return err
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
