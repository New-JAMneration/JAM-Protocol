package ce

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type CERequestID uint8

const (
	BlockRequest CERequestID = 128 + iota
	StateRequest
	SafroleTicketDistribution
	WorkPackageSubmission
	WorkPackageSharing
	WorkReportDistribution
	WorkReportRequest
	ShardDistribution
	AuditShardReqeust
	SegmentShardRequest
	AssuranceDistribution
	PreimageAnnouncement
	PreimageRequest
	AuditAnnouncement
	JudgmentPublication
)

type CERequestHandler interface {
	Encode(req CERequestID, message interface{}) ([]byte, error)
	Decode(data []byte) (CERequestID, interface{}, error)
}

type DefaultCERequestHandler struct {
	encoder *types.Encoder
}

func NewDefaultCERequestHandler() *DefaultCERequestHandler {
	return &DefaultCERequestHandler{
		encoder: types.NewEncoder(),
	}
}

func (h *DefaultCERequestHandler) Encode(req CERequestID, message interface{}) ([]byte, error) {
	h.encoder = types.NewEncoder()

	if err := h.encoder.EncodeInteger(uint64(req)); err != nil {
		return nil, fmt.Errorf("failed to encode request type: %w", err)
	}

	switch req {
	case BlockRequest:
		return h.encodeBlockRequest(message)
	case StateRequest:
		return h.encodeStateRequest(message)
	case SafroleTicketDistribution:
		// TODO: Implement SafroleTicketDistribution encoding
		return nil, fmt.Errorf("SafroleTicketDistribution encoding not implemented yet")
	case WorkPackageSubmission:
		// TODO: Implement WorkPackageSubmission encoding
		return nil, fmt.Errorf("WorkPackageSubmission encoding not implemented yet")
	case WorkPackageSharing:
		// TODO: Implement WorkPackageSharing encoding
		return nil, fmt.Errorf("WorkPackageSharing encoding not implemented yet")
	case WorkReportDistribution:
		// TODO: Implement WorkReportDistribution encoding
		return nil, fmt.Errorf("WorkReportDistribution encoding not implemented yet")
	case WorkReportRequest:
		// TODO: Implement WorkReportRequest encoding
		return nil, fmt.Errorf("WorkReportRequest encoding not implemented yet")
	case ShardDistribution:
		// TODO: Implement ShardDistribution encoding
		return nil, fmt.Errorf("ShardDistribution encoding not implemented yet")
	case AuditShardReqeust:
		// TODO: Implement AuditShardReqeust encoding
		return nil, fmt.Errorf("AuditShardReqeust encoding not implemented yet")
	case SegmentShardRequest:
		// TODO: Implement SegmentShardRequest encoding
		return nil, fmt.Errorf("SegmentShardRequest encoding not implemented yet")
	case AssuranceDistribution:
		// TODO: Implement AssuranceDistribution encoding
		return nil, fmt.Errorf("AssuranceDistribution encoding not implemented yet")
	case PreimageAnnouncement:
		// TODO: Implement PreimageAnnouncement encoding
		return nil, fmt.Errorf("PreimageAnnouncement encoding not implemented yet")
	case PreimageRequest:
		// TODO: Implement PreimageRequest encoding
		return nil, fmt.Errorf("PreimageRequest encoding not implemented yet")
	case AuditAnnouncement:
		// TODO: Implement AuditAnnouncement encoding
		return nil, fmt.Errorf("AuditAnnouncement encoding not implemented yet")
	case JudgmentPublication:
		// TODO: Implement JudgmentPublication encoding
		return nil, fmt.Errorf("JudgmentPublication encoding not implemented yet")
	default:
		return nil, fmt.Errorf("unknown request type: %d", req)
	}
}

func (h *DefaultCERequestHandler) encodeBlockRequest(message interface{}) ([]byte, error) {
	blockReq, ok := message.(*CE128Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for BlockRequest: %T", message)
	}

	h.encoder = types.NewEncoder()

	// Encode HeaderHash
	if err := blockReq.HeaderHash.Encode(h.encoder); err != nil {
		return nil, fmt.Errorf("failed to encode HeaderHash: %w", err)
	}

	// Encode Direction as U8
	if err := h.encoder.WriteByte(blockReq.Direction); err != nil {
		return nil, fmt.Errorf("failed to encode Direction: %w", err)
	}

	// Encode MaxBlocks as U32
	maxBlocks := types.U32(blockReq.MaxBlocks)
	if err := maxBlocks.Encode(h.encoder); err != nil {
		return nil, fmt.Errorf("failed to encode MaxBlocks: %w", err)
	}

	// Use the encoder's Encode method to get the bytes
	encoded, err := h.encoder.Encode(blockReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode CE128Payload: %w", err)
	}
	return encoded, nil
}

func (h *DefaultCERequestHandler) encodeStateRequest(message interface{}) ([]byte, error) {
	stateReq, ok := message.(*CE129Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for StateRequest: %T", message)
	}

	// Create a new encoder for this request
	encoder := types.NewEncoder()

	// Helper to write raw bytes to the encoder
	writeRaw := func(b []byte) error {
		for _, v := range b {
			if err := encoder.WriteByte(v); err != nil {
				return err
			}
		}
		return nil
	}

	// Encode HeaderHash (32 bytes)
	if err := writeRaw(stateReq.HeaderHash[:]); err != nil {
		return nil, fmt.Errorf("failed to encode HeaderHash: %w", err)
	}
	// Encode KeyStart (31 bytes)
	if err := writeRaw(stateReq.KeyStart[:]); err != nil {
		return nil, fmt.Errorf("failed to encode KeyStart: %w", err)
	}
	// Encode KeyEnd (31 bytes)
	if err := writeRaw(stateReq.KeyEnd[:]); err != nil {
		return nil, fmt.Errorf("failed to encode KeyEnd: %w", err)
	}
	// Encode MaxSize (U32, 4 bytes little-endian)
	maxSize := stateReq.MaxSize
	maxSizeBytes := []byte{
		byte(maxSize),
		byte(maxSize >> 8),
		byte(maxSize >> 16),
		byte(maxSize >> 24),
	}
	if err := writeRaw(maxSizeBytes); err != nil {
		return nil, fmt.Errorf("failed to encode MaxSize: %w", err)
	}

	// Since we can't access the buffer directly, let's just return the manually constructed bytes
	result := make([]byte, 0, 98)
	result = append(result, stateReq.HeaderHash[:]...)
	result = append(result, stateReq.KeyStart[:]...)
	result = append(result, stateReq.KeyEnd[:]...)
	result = append(result, maxSizeBytes...)

	return result, nil
}

func (h *DefaultCERequestHandler) Decode(data []byte) (CERequestID, interface{}, error) {
	// TODO: Implement decoding logic
	// This would involve:
	// 1. Reading the request type from the beginning of the data
	// 2. Based on the request type, decoding the appropriate message structure
	// 3. Returning the decoded request type and message

	return BlockRequest, nil, fmt.Errorf("decoding not implemented yet")
}
