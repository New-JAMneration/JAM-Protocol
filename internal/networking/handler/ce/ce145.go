package ce

import (
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Role: [Auditor -> Validator]
func HandleJudgmentAnnouncement_Validator(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// Epoch Index ++ Validator Index ++ Validity ++ Work-Report Hash ++ Ed25519 Signature
	// =  4 + 2 + 1 + 32 + 64 = 103 bytes
	payloadSize := 4 + 2 + 1 + 32 + 64
	payload := make([]byte, payloadSize)
	if err := stream.ReadFull(payload); err != nil {
		return fmt.Errorf("failed to read judgment payload: %w", err)
	}

	epochIndex := types.U32(binary.LittleEndian.Uint32(payload[:4]))
	validatorIndex := types.ValidatorIndex(binary.LittleEndian.Uint16(payload[4:6]))
	validity := payload[6]
	workReportHash := types.WorkReportHash{}
	copy(workReportHash[:], payload[7:39])
	signature := types.Ed25519Signature{}
	copy(signature[:], payload[39:103])

	if err := validateJudgmentAnnouncement(epochIndex, validatorIndex, validity, workReportHash, signature); err != nil {
		return fmt.Errorf("invalid judgment announcement: %w", err)
	} else if err := storeJudgmentAnnouncement(epochIndex, validatorIndex, validity, workReportHash, signature); err != nil {
		return fmt.Errorf("failed to store judgment announcement: %w", err)
	} else if err := stream.WriteMessage([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN response: %w", err)
	}

	finBuf := make([]byte, 3)
	if err := stream.ReadFull(finBuf); err != nil {
		return fmt.Errorf("failed to read FIN: %w", err)
	} else if string(finBuf) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	return stream.Close()
}

// Role: [Auditor -> Validator]
func HandleJudgmentAnnouncement_Auditor(
	stream *quic.Stream,
	epochIndex types.U32,
	validatorIndex types.ValidatorIndex,
	validity uint8,
	workReportHash types.WorkReportHash,
	signature types.Ed25519Signature,
) error {
	// TODO: create (or mock) data for constructing judgment announcement
	payload, err := CreateJudgmentAnnouncement(epochIndex, validatorIndex, validity, workReportHash, signature)
	if err != nil {
		return fmt.Errorf("failed to create judgment announcement: %w", err)
	}

	if err := stream.WriteMessage(payload); err != nil {
		return fmt.Errorf("failed to send judgment announcement: %w", err)
	}

	if err := stream.WriteMessage([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to send FIN: %w", err)
	}

	finBuf := make([]byte, 3)
	if err := stream.ReadFull(finBuf); err != nil {
		return fmt.Errorf("failed to read FIN: %w", err)
	} else if string(finBuf) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	return stream.Close()
}

// TODO: add other judgment announcement validation here.
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

	validators := store.GetInstance().GetPriorStates().GetKappa()
	if int(validatorIndex) < 0 || int(validatorIndex) >= len(validators) {
		return fmt.Errorf("bad_validator_key")
	}

	pub := validators[validatorIndex].Ed25519[:]
	if !ed25519.Verify(pub, msg, signature[:]) {
		return fmt.Errorf("bad_signature")
	}

	return nil
}

func storeJudgmentAnnouncement(epochIndex types.U32, validatorIndex types.ValidatorIndex, validity uint8, workReportHash types.WorkReportHash, signature types.Ed25519Signature) error {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return fmt.Errorf("failed to get Redis backend: %w", err)
	}

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

	workReportHashHex := hex.EncodeToString(workReportHash[:])
	key := fmt.Sprintf("judgment:%s:%d:%d", workReportHashHex, epochIndex, validatorIndex)

	client := redisBackend.GetClient()
	err = client.Put(key, encodedJudgment)
	if err != nil {
		return fmt.Errorf("failed to store judgment in Redis: %w", err)
	}

	workReportKey := fmt.Sprintf("work_report_judgments:%s", workReportHashHex)
	err = client.SAdd(workReportKey, encodedJudgment)
	if err != nil {
		return fmt.Errorf("failed to add judgment to work report set: %w", err)
	}

	epochKey := fmt.Sprintf("epoch_judgments:%d", epochIndex)
	err = client.SAdd(epochKey, encodedJudgment)
	if err != nil {
		return fmt.Errorf("failed to add judgment to epoch set: %w", err)
	}

	validatorKey := fmt.Sprintf("validator_judgments:%d", validatorIndex)
	err = client.SAdd(validatorKey, encodedJudgment)
	if err != nil {
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

func GetJudgment(workReportHash types.WorkReportHash, epochIndex types.U32, validatorIndex types.ValidatorIndex) (*CE145Payload, error) {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis backend: %w", err)
	}

	workReportHashHex := hex.EncodeToString(workReportHash[:])
	key := fmt.Sprintf("judgment:%s:%d:%d", workReportHashHex, epochIndex, validatorIndex)

	client := redisBackend.GetClient()
	encodedJudgment, err := client.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get judgment from Redis: %w", err)
	}

	if encodedJudgment == nil {
		return nil, fmt.Errorf("judgment not found for work report: %x, epoch: %d, validator: %d", workReportHash, epochIndex, validatorIndex)
	}

	judgmentData := &CE145Payload{}
	err = judgmentData.Decode(encodedJudgment)
	if err != nil {
		return nil, fmt.Errorf("failed to decode judgment data: %w", err)
	}

	return judgmentData, nil
}

func GetAllJudgmentsForWorkReport(workReportHash types.WorkReportHash) ([]*CE145Payload, error) {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis backend: %w", err)
	}

	workReportHashHex := hex.EncodeToString(workReportHash[:])
	workReportKey := fmt.Sprintf("work_report_judgments:%s", workReportHashHex)

	client := redisBackend.GetClient()
	encodedJudgments, err := client.SMembers(workReportKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get judgments set from Redis: %w", err)
	}

	var judgments []*CE145Payload
	for _, encodedJudgment := range encodedJudgments {
		judgmentData := &CE145Payload{}
		err := judgmentData.Decode(encodedJudgment)
		if err != nil {
			return nil, fmt.Errorf("failed to decode judgment data: %w", err)
		}
		judgments = append(judgments, judgmentData)
	}

	return judgments, nil
}

func GetAllJudgmentsForEpoch(epochIndex types.U32) ([]*CE145Payload, error) {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis backend: %w", err)
	}

	epochKey := fmt.Sprintf("epoch_judgments:%d", epochIndex)

	client := redisBackend.GetClient()
	encodedJudgments, err := client.SMembers(epochKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get judgments set from Redis: %w", err)
	}

	var judgments []*CE145Payload
	for _, encodedJudgment := range encodedJudgments {
		judgmentData := &CE145Payload{}
		err := judgmentData.Decode(encodedJudgment)
		if err != nil {
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
