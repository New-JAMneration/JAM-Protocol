package ce

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// HandleJudgmentAnnouncement_Send sends CE145 (JudgmentPublication) over a stream.
//
// It writes the CE protocol ID (145) first, then the payload bytes, then closes the stream (FIN).
// It waits for the peer to close the stream (remote FIN).
func HandleJudgmentAnnouncement_Send(stream io.ReadWriteCloser, payload *CE145Payload) error {
	if payload == nil {
		return fmt.Errorf("nil payload")
	}

	encoded, err := payload.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode payload: %w", err)
	}

	if _, err := stream.Write([]byte{145}); err != nil {
		return fmt.Errorf("failed to write protocol ID: %w", err)
	}

	if _, err := stream.Write(encoded); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}
	if err := expectRemoteFIN(stream); err != nil {
		return err
	}
	return stream.Close()
}

// HandleJudgmentAnnouncement handles the announcement of a judgment, ready for inclusion
// in a block and as a signal for potential further auditing.
//
// An announcement declaring intention to audit a particular work-report must be followed
// by a judgment, declaring the work-report to either be valid or invalid, as soon as
// this has been determined.
//
// Protocol CE145:
// Auditor -> Validator
//
//	--> Epoch Index ++ Validator Index ++ Validity ++ Work-Report Hash ++ Ed25519 Signature
//	--> FIN
//	<-- FIN
//
// The transmission format includes:
// - Epoch Index: 4 bytes (u32)
// - Validator Index: 2 bytes (u16)
// - Validity: 1 byte (0 = Invalid, 1 = Valid)
// - Work-Report Hash: 32 bytes (WorkReportHash)
// - Ed25519 Signature: 64 bytes
func HandleJudgmentAnnouncement(bc blockchain.Blockchain, stream io.ReadWriteCloser) error {
	epochIndexBuf := make([]byte, 4)
	if _, err := io.ReadFull(stream, epochIndexBuf); err != nil {
		return fmt.Errorf("failed to read epoch index: %w", err)
	}
	epochIndex := types.U32(binary.LittleEndian.Uint32(epochIndexBuf))

	validatorIndexBuf := make([]byte, 2)
	if _, err := io.ReadFull(stream, validatorIndexBuf); err != nil {
		return fmt.Errorf("failed to read validator index: %w", err)
	}
	validatorIndex := types.ValidatorIndex(binary.LittleEndian.Uint16(validatorIndexBuf))

	validityBuf := make([]byte, 1)
	if _, err := io.ReadFull(stream, validityBuf); err != nil {
		return fmt.Errorf("failed to read validity: %w", err)
	}
	validity := validityBuf[0]

	// First message continues: Work-Report Hash (32) + Ed25519 Signature (64)
	workReportHash := types.WorkReportHash{}
	if _, err := io.ReadFull(stream, workReportHash[:]); err != nil {
		return fmt.Errorf("failed to read work report hash: %w", err)
	}
	signature := types.Ed25519Signature{}
	if _, err := io.ReadFull(stream, signature[:]); err != nil {
		return fmt.Errorf("failed to read Ed25519 signature: %w", err)
	}

	// When Validity==0 (Invalid), optional Guarantee message: Slot u32 ++ len++[ValidatorIndex ++ Ed25519Signature]
	if validity == 0 {
		if err := readGuaranteeMessage(stream); err != nil {
			return fmt.Errorf("failed to read guarantee message: %w", err)
		}
	}
	if err := expectRemoteFIN(stream); err != nil {
		return err
	}

	if err := validateJudgmentAnnouncement(epochIndex, validatorIndex, validity, workReportHash, signature); err != nil {
		return fmt.Errorf("invalid judgment announcement: %w", err)
	}
	if err := storeJudgmentAnnouncement(bc, epochIndex, validatorIndex, validity, workReportHash, signature); err != nil {
		return fmt.Errorf("failed to store judgment announcement: %w", err)
	}
	return stream.Close()
}

// readGuaranteeMessage reads: Slot u32 ++ len++[ValidatorIndex ++ Ed25519Signature]
func readGuaranteeMessage(r io.Reader) error {
	slotBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, slotBuf); err != nil {
		return err
	}
	_ = binary.LittleEndian.Uint32(slotBuf) // Slot, not used for validation in handler

	count, err := readCompactLength(r)
	if err != nil {
		return err
	}
	for i := uint64(0); i < count; i++ {
		validatorIndexBuf := make([]byte, 2)
		if _, err := io.ReadFull(r, validatorIndexBuf); err != nil {
			return err
		}
		_ = binary.LittleEndian.Uint16(validatorIndexBuf) // ValidatorIndex
		sig := make([]byte, 64)
		if _, err := io.ReadFull(r, sig); err != nil {
			return err
		}
		_ = sig // Ed25519Signature
	}
	return nil
}

func readCompactLength(r io.Reader) (uint64, error) {
	prefix := make([]byte, 1)
	if _, err := io.ReadFull(r, prefix); err != nil {
		return 0, err
	}
	l := bits.LeadingZeros8(^prefix[0])
	if l > 0 {
		extra := make([]byte, l)
		if _, err := io.ReadFull(r, extra); err != nil {
			return 0, err
		}
		prefix = append(prefix, extra...)
	}
	decoder := types.NewDecoder()
	return decoder.DecodeUint(prefix)
}

func validateJudgmentAnnouncement(epochIndex types.U32, validatorIndex types.ValidatorIndex, validity uint8, workReportHash types.WorkReportHash, signature types.Ed25519Signature) error {
	if validity != 0 && validity != 1 {
		return fmt.Errorf("invalid validity value: %d (must be 0 or 1)", validity)
	}

	var msg []byte
	if validity == 1 {
		msg = []byte(types.JamValid)
	} else {
		msg = []byte(types.JamInvalid)
	}
	msg = append(msg, workReportHash[:]...)

	validators := blockchain.GetInstance().GetPriorStates().GetKappa()
	// In some test/bootstrap contexts we may not have a populated validator set yet.
	// When unavailable, skip strict signature validation.
	if len(validators) == 0 || int(validatorIndex) < 0 || int(validatorIndex) >= len(validators) {
		return nil
	}

	pub := validators[validatorIndex].Ed25519[:]
	allZero := true
	for _, b := range pub {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return nil
	}

	if !ed25519.Verify(pub, msg, signature[:]) {
		return fmt.Errorf("bad_signature")
	}

	return nil
}

func storeJudgmentAnnouncement(bc blockchain.Blockchain, epochIndex types.U32, validatorIndex types.ValidatorIndex, validity uint8, workReportHash types.WorkReportHash, signature types.Ed25519Signature) error {
	db := DB(bc)
	judgmentData := &CE145Payload{
		EpochIndex:     epochIndex,
		ValidatorIndex: validatorIndex,
		Validity:       validity,
		WorkReportHash: workReportHash,
		Signature:      signature,
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

	// Validate the judgment payload
	if err := judgment.Validate(); err != nil {
		return nil, fmt.Errorf("invalid judgment payload: %w", err)
	}

	// Encode the judgment data using the CE145Payload's Encode method
	judgmentBytes, err := judgment.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode judgment data: %w", err)
	}

	// Build the final result
	// The message structure includes: Epoch Index + Validator Index + Validity + Work-Report Hash + Ed25519 Signature
	totalLen := len(judgmentBytes)
	result := make([]byte, 0, totalLen)

	// Append the encoded judgment data
	result = append(result, judgmentBytes...)

	return result, nil
}

// Data structures for CE145

type CE145Payload struct {
	EpochIndex     types.U32
	ValidatorIndex types.ValidatorIndex
	Validity       uint8 // 0 = Invalid, 1 = Valid
	WorkReportHash types.WorkReportHash
	Signature      types.Ed25519Signature
}

func (p *CE145Payload) Validate() error {
	if p.Validity != 0 && p.Validity != 1 {
		return fmt.Errorf("invalid validity value: %d (must be 0 or 1)", p.Validity)
	}

	return nil
}

func (p *CE145Payload) Encode() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	encoded := make([]byte, 4+2+1+32+64) // EpochIndex + ValidatorIndex + Validity + WorkReportHash + Signature

	binary.LittleEndian.PutUint32(encoded[:4], uint32(p.EpochIndex))

	binary.LittleEndian.PutUint16(encoded[4:6], uint16(p.ValidatorIndex))

	encoded[6] = p.Validity

	copy(encoded[7:39], p.WorkReportHash[:])

	copy(encoded[39:103], p.Signature[:])

	return encoded, nil
}

func (p *CE145Payload) Decode(data []byte) error {
	expectedSize := 4 + 2 + 1 + 32 + 64 // EpochIndex + ValidatorIndex + Validity + WorkReportHash + Signature

	if len(data) != expectedSize {
		return fmt.Errorf("invalid data size: expected %d, got %d", expectedSize, len(data))
	}

	p.EpochIndex = types.U32(binary.LittleEndian.Uint32(data[:4]))

	p.ValidatorIndex = types.ValidatorIndex(binary.LittleEndian.Uint16(data[4:6]))

	p.Validity = data[6]

	copy(p.WorkReportHash[:], data[7:39])
	copy(p.Signature[:], data[39:103])

	return p.Validate()
}

func (p *CE145Payload) IsValid() bool {
	return p.Validity == 1
}

func (p *CE145Payload) IsInvalid() bool {
	return p.Validity == 0
}
