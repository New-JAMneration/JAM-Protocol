package ce

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/hdevalence/ed25519consensus"
)

// Field sizes and offsets within the CE145 judgment header (103 bytes total).
const (
	ce145HeaderSize   = U32Size + U16Size + 1 + HashSize + types.Ed25519SigSize // 103
	ce145OffEpoch     = 0
	ce145OffValidator = ce145OffEpoch + U32Size        // 4
	ce145OffValidity  = ce145OffValidator + U16Size    // 6
	ce145OffHash      = ce145OffValidity + 1           // 7
	ce145OffSig       = ce145OffHash + HashSize        // 39
	ce145SigEntrySize = U16Size + types.Ed25519SigSize // 66 (per guarantee entry)
)

// ce145Stream supports JAMNP message framing (ReadMessage / WriteMessage).
type ce145Stream interface {
	io.ReadWriteCloser
	ReadMessage() ([]byte, error)
	WriteMessage(payload []byte) error
}

// CE145Guarantee is the optional second message sent with an invalid judgment.
// Wire: Slot ++ len++[ValidatorIndex ++ Ed25519Signature]  (2 or 3 entries only)
type CE145Guarantee struct {
	Slot       types.TimeSlot
	Signatures []types.ValidatorSignature
}

// Validate checks count range [GuaranteeMinCount, GuaranteeMaxCount] and each
// ValidatorIndex via types.ValidateGuaranteeSignatures.
func (g *CE145Guarantee) Validate() error {
	return types.ValidateGuaranteeSignatures(g.Signatures)
}

// CE145Payload carries all data for a CE 145 judgment publication.
type CE145Payload struct {
	EpochIndex     types.U32
	ValidatorIndex types.ValidatorIndex
	Validity       uint8 // 0 = Invalid, 1 = Valid
	WorkReportHash types.WorkReportHash
	Signature      types.Ed25519Signature
	Guarantee      *CE145Guarantee // non-nil iff Validity == 0
}

// ── Public handlers ───────────────────────────────────────────────────────────

// HandleJudgmentAnnouncement_Send sends CE145 over a stream.
// The stream kind byte (145) is written raw; both messages use WriteMessage framing.
func HandleJudgmentAnnouncement_Send(stream ce145Stream, payload *CE145Payload) error {
	if payload == nil {
		return fmt.Errorf("nil payload")
	}
	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	// Stream kind byte — raw, not framed.
	if _, err := stream.Write([]byte{145}); err != nil {
		return fmt.Errorf("failed to write protocol ID: %w", err)
	}

	firstMsg, err := encodeJudgmentHeader(payload)
	if err != nil {
		return fmt.Errorf("failed to encode judgment header: %w", err)
	}
	if err := stream.WriteMessage(firstMsg); err != nil {
		return fmt.Errorf("failed to write judgment header: %w", err)
	}

	// For invalid judgments, encode and write the second message (guarantee).
	if payload.Validity == 0 {
		if payload.Guarantee == nil {
			return fmt.Errorf("guarantee required for invalid judgment")
		}
		guaranteeBytes, err := encodeGuarantee(payload.Guarantee)
		if err != nil {
			return fmt.Errorf("failed to encode guarantee: %w", err)
		}
		if err := stream.WriteMessage(guaranteeBytes); err != nil {
			return fmt.Errorf("failed to write guarantee: %w", err)
		}
	}

	if err := expectRemoteFIN(stream); err != nil {
		return err
	}
	return stream.Close()
}

// HandleJudgmentAnnouncement handles incoming CE145 judgment publication.
//
// Protocol CE145: Auditor → Validator
//
//	--> Epoch Index ++ Validator Index ++ Validity ++ Work-Report Hash ++ Ed25519 Signature
//	--> Guarantee [present iff Validity == 0]
//	--> FIN
//	<-- FIN
func HandleJudgmentAnnouncement(bc blockchain.Blockchain, stream ce145Stream) error {
	// Message 1: fixed-size judgment header.
	firstMsg, err := stream.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read first message: %w", err)
	}
	if len(firstMsg) != ce145HeaderSize {
		return fmt.Errorf("invalid first message size: expected %d, got %d", ce145HeaderSize, len(firstMsg))
	}

	epochIndex := types.U32(binary.LittleEndian.Uint32(firstMsg[ce145OffEpoch:]))
	validatorIndex := types.ValidatorIndex(binary.LittleEndian.Uint16(firstMsg[ce145OffValidator:]))
	validity := firstMsg[ce145OffValidity]
	var workReportHash types.WorkReportHash
	copy(workReportHash[:], firstMsg[ce145OffHash:ce145OffSig])
	var signature types.Ed25519Signature
	copy(signature[:], firstMsg[ce145OffSig:ce145HeaderSize])

	// Message 2: guarantee (mandatory when validity == 0).
	var guarantee *CE145Guarantee
	if validity == 0 {
		guaranteeMsg, err := stream.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read guarantee message: %w", err)
		}
		if guarantee, err = decodeGuaranteeBytes(guaranteeMsg); err != nil {
			return fmt.Errorf("failed to decode guarantee: %w", err)
		}
	}

	if err := expectRemoteFIN(stream); err != nil {
		return err
	}
	if err := validateJudgmentAnnouncement(epochIndex, validatorIndex, validity, workReportHash, signature, guarantee); err != nil {
		return fmt.Errorf("invalid judgment announcement: %w", err)
	}
	if err := storeJudgmentAnnouncement(bc, epochIndex, validatorIndex, validity, workReportHash, signature, guarantee); err != nil {
		return fmt.Errorf("failed to store judgment announcement: %w", err)
	}
	return stream.Close()
}

// ── Codec helpers ─────────────────────────────────────────────────────────────

func encodeJudgmentHeader(p *CE145Payload) ([]byte, error) {
	buf := make([]byte, ce145HeaderSize)
	binary.LittleEndian.PutUint32(buf[ce145OffEpoch:], uint32(p.EpochIndex))
	binary.LittleEndian.PutUint16(buf[ce145OffValidator:], uint16(p.ValidatorIndex))
	buf[ce145OffValidity] = p.Validity
	copy(buf[ce145OffHash:ce145OffSig], p.WorkReportHash[:])
	copy(buf[ce145OffSig:ce145HeaderSize], p.Signature[:])
	return buf, nil
}

func encodeGuarantee(g *CE145Guarantee) ([]byte, error) {
	if g == nil {
		return nil, fmt.Errorf("nil guarantee")
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	count := len(g.Signatures)
	countBytes, err := types.NewEncoder().EncodeUint(uint64(count))
	if err != nil {
		return nil, fmt.Errorf("failed to encode count: %w", err)
	}
	buf := make([]byte, 0, U32Size+len(countBytes)+count*ce145SigEntrySize)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(g.Slot))
	buf = append(buf, countBytes...)
	for _, s := range g.Signatures {
		buf = binary.LittleEndian.AppendUint16(buf, uint16(s.ValidatorIndex))
		buf = append(buf, s.Signature[:]...)
	}
	return buf, nil
}

// decodeGuaranteeBytes parses the wire bytes of a guarantee message.
// Used by both HandleJudgmentAnnouncement (stream path) and CE145Payload.Decode (persistence).
func decodeGuaranteeBytes(data []byte) (*CE145Guarantee, error) {
	if len(data) < U32Size+1 { // Slot(4) + at least 1 byte for compact count
		return nil, fmt.Errorf("incomplete guarantee data: %d bytes", len(data))
	}
	slot := types.TimeSlot(binary.LittleEndian.Uint32(data[:U32Size]))
	rest := data[U32Size:]

	count, n, err := decodeCompactUint(rest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode count: %w", err)
	}
	rest = rest[n:]

	// Check count before allocating to guard against malicious input.
	if count < types.GuaranteeMinCount || count > types.GuaranteeMaxCount {
		return nil, fmt.Errorf("guarantee signature count %d out of range [%d, %d]",
			count, types.GuaranteeMinCount, types.GuaranteeMaxCount)
	}

	sigs := make([]types.ValidatorSignature, count)
	for i := uint64(0); i < count; i++ {
		if len(rest) < ce145SigEntrySize {
			return nil, fmt.Errorf("insufficient data for guarantee signature %d", i)
		}
		var sig types.Ed25519Signature
		copy(sig[:], rest[U16Size:ce145SigEntrySize])
		sigs[i] = types.ValidatorSignature{
			ValidatorIndex: types.ValidatorIndex(binary.LittleEndian.Uint16(rest[:U16Size])),
			Signature:      sig,
		}
		rest = rest[ce145SigEntrySize:]
	}

	g := &CE145Guarantee{Slot: slot, Signatures: sigs}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	return g, nil
}

// ── Validation ────────────────────────────────────────────────────────────────

func validateJudgmentAnnouncement(
	epochIndex types.U32,
	validatorIndex types.ValidatorIndex,
	validity uint8,
	workReportHash types.WorkReportHash,
	signature types.Ed25519Signature,
	guarantee *CE145Guarantee,
) error {
	if validity != 0 && validity != 1 {
		return fmt.Errorf("invalid validity value: %d (must be 0 or 1)", validity)
	}
	if validity == 0 && guarantee == nil {
		return fmt.Errorf("guarantee is required for invalid judgments")
	}
	if validity == 1 && guarantee != nil {
		return fmt.Errorf("guarantee must not be present for valid judgments")
	}
	if guarantee != nil {
		if err := guarantee.Validate(); err != nil {
			return err
		}
	}

	var msg []byte
	if validity == 1 {
		msg = []byte(types.JamValid)
	} else {
		msg = []byte(types.JamInvalid)
	}
	msg = append(msg, workReportHash[:]...)

	validators := blockchain.GetInstance().GetPriorStates().GetKappa()
	if len(validators) == 0 {
		return nil
	}
	if int(validatorIndex) >= len(validators) {
		return fmt.Errorf("validator index out of range: index=%d validators=%d", validatorIndex, len(validators))
	}
	pub := validators[validatorIndex].Ed25519[:]
	if !bytes.Equal(pub, make([]byte, len(pub))) {
		if !ed25519consensus.Verify(pub, msg, signature[:]) {
			return errors.New("bad_signature")
		}
	}
	return nil
}

// ── Payload methods ───────────────────────────────────────────────────────────

// Validate checks validity flag. Invalid judgments (Validity==0) must carry a non-nil Guarantee.
func (p *CE145Payload) Validate() error {
	if p.Validity != 0 && p.Validity != 1 {
		return fmt.Errorf("invalid validity value: %d (must be 0 or 1)", p.Validity)
	}
	if p.Validity == 0 && p.Guarantee == nil {
		return fmt.Errorf("guarantee is required for invalid judgments")
	}
	if p.Validity == 1 && p.Guarantee != nil {
		return fmt.Errorf("guarantee must not be present for valid judgments")
	}
	if p.Guarantee != nil {
		return p.Guarantee.Validate()
	}
	return nil
}

// Encode serialises the judgment header; invalid judgments (Validity==0) also append guarantee bytes.
func (p *CE145Payload) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	header, err := encodeJudgmentHeader(p)
	if err != nil {
		return nil, err
	}
	if p.Validity == 0 {
		g, err := encodeGuarantee(p.Guarantee)
		if err != nil {
			return nil, fmt.Errorf("failed to encode guarantee: %w", err)
		}
		return append(header, g...), nil
	}
	return header, nil
}

// Decode deserialises a CE145Payload.  The fixed header is always ce145HeaderSize bytes;
// invalid judgments (Validity==0) must be followed by guarantee bytes.
func (p *CE145Payload) Decode(data []byte) error {
	if len(data) < ce145HeaderSize {
		return fmt.Errorf("data too short: expected at least %d, got %d", ce145HeaderSize, len(data))
	}
	p.EpochIndex = types.U32(binary.LittleEndian.Uint32(data[ce145OffEpoch:]))
	p.ValidatorIndex = types.ValidatorIndex(binary.LittleEndian.Uint16(data[ce145OffValidator:]))
	p.Validity = data[ce145OffValidity]
	copy(p.WorkReportHash[:], data[ce145OffHash:ce145OffSig])
	copy(p.Signature[:], data[ce145OffSig:ce145HeaderSize])

	if p.Validity == 0 {
		if len(data) <= ce145HeaderSize {
			return fmt.Errorf("invalid judgment must include guarantee bytes")
		}
		g, err := decodeGuaranteeBytes(data[ce145HeaderSize:])
		if err != nil {
			return fmt.Errorf("failed to decode guarantee: %w", err)
		}
		p.Guarantee = g
	}
	return p.Validate()
}

func (p *CE145Payload) IsValid() bool   { return p.Validity == 1 }
func (p *CE145Payload) IsInvalid() bool { return p.Validity == 0 }

// ── Storage & retrieval ───────────────────────────────────────────────────────

func storeJudgmentAnnouncement(
	bc blockchain.Blockchain,
	epochIndex types.U32,
	validatorIndex types.ValidatorIndex,
	validity uint8,
	workReportHash types.WorkReportHash,
	signature types.Ed25519Signature,
	guarantee *CE145Guarantee,
) error {
	db := DB(bc)
	judgmentData := &CE145Payload{
		EpochIndex:     epochIndex,
		ValidatorIndex: validatorIndex,
		Validity:       validity,
		WorkReportHash: workReportHash,
		Signature:      signature,
		Guarantee:      guarantee,
	}
	encodedJudgment, err := judgmentData.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode judgment data: %w", err)
	}
	if err := PutKV(db, ceJudgmentKey(workReportHash, epochIndex, validatorIndex), encodedJudgment); err != nil {
		return fmt.Errorf("failed to store judgment: %w", err)
	}
	if err := SAdd(db, ceJudgmentWorkReportSetKey(workReportHash), encodedJudgment); err != nil {
		return fmt.Errorf("failed to add judgment to work report set: %w", err)
	}
	if err := SAdd(db, ceJudgmentEpochSetKey(epochIndex), encodedJudgment); err != nil {
		return fmt.Errorf("failed to add judgment to epoch set: %w", err)
	}
	if err := SAdd(db, ceJudgmentValidatorSetKey(validatorIndex), encodedJudgment); err != nil {
		return fmt.Errorf("failed to add judgment to validator set: %w", err)
	}
	return nil
}

func CreateJudgmentAnnouncement(
	epochIndex types.U32,
	validatorIndex types.ValidatorIndex,
	validity uint8,
	workReportHash types.WorkReportHash,
	signature types.Ed25519Signature,
) ([]byte, error) {
	payload := &CE145Payload{
		EpochIndex:     epochIndex,
		ValidatorIndex: validatorIndex,
		Validity:       validity,
		WorkReportHash: workReportHash,
		Signature:      signature,
	}
	return payload.Encode()
}

func GetJudgment(bc blockchain.Blockchain, workReportHash types.WorkReportHash, epochIndex types.U32, validatorIndex types.ValidatorIndex) (*CE145Payload, error) {
	db := DB(bc)
	encodedJudgment, err := GetKV(db, ceJudgmentKey(workReportHash, epochIndex, validatorIndex))
	if err != nil {
		return nil, fmt.Errorf("failed to get judgment: %w", err)
	}
	if encodedJudgment == nil {
		return nil, fmt.Errorf("judgment not found for work report: %x, epoch: %d, validator: %d", workReportHash, epochIndex, validatorIndex)
	}
	judgmentData := &CE145Payload{}
	if err := judgmentData.Decode(encodedJudgment); err != nil {
		return nil, fmt.Errorf("failed to decode judgment data: %w", err)
	}
	return judgmentData, nil
}

func GetAllJudgmentsForWorkReport(bc blockchain.Blockchain, workReportHash types.WorkReportHash) ([]*CE145Payload, error) {
	db := DB(bc)
	encodedJudgments, err := SMembers(db, ceJudgmentWorkReportSetKey(workReportHash))
	if err != nil {
		return nil, fmt.Errorf("failed to get judgments set: %w", err)
	}
	var judgments []*CE145Payload
	for _, encodedJudgment := range encodedJudgments {
		judgmentData := &CE145Payload{}
		if err := judgmentData.Decode(encodedJudgment); err != nil {
			return nil, fmt.Errorf("failed to decode judgment data: %w", err)
		}
		judgments = append(judgments, judgmentData)
	}
	return judgments, nil
}

func GetAllJudgmentsForEpoch(bc blockchain.Blockchain, epochIndex types.U32) ([]*CE145Payload, error) {
	db := DB(bc)
	encodedJudgments, err := SMembers(db, ceJudgmentEpochSetKey(epochIndex))
	if err != nil {
		return nil, fmt.Errorf("failed to get judgments set: %w", err)
	}
	var judgments []*CE145Payload
	for _, encodedJudgment := range encodedJudgments {
		judgmentData := &CE145Payload{}
		if err := judgmentData.Decode(encodedJudgment); err != nil {
			return nil, fmt.Errorf("failed to decode judgment data: %w", err)
		}
		judgments = append(judgments, judgmentData)
	}
	return judgments, nil
}

func (h *DefaultCERequestHandler) encodeJudgmentPublication(message interface{}) ([]byte, error) {
	judgment, ok := message.(*CE145Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for JudgmentPublication: %T", message)
	}
	if judgment == nil {
		return nil, fmt.Errorf("nil payload for JudgmentPublication")
	}
	if err := judgment.Validate(); err != nil {
		return nil, fmt.Errorf("invalid judgment payload: %w", err)
	}
	return judgment.Encode()
}
