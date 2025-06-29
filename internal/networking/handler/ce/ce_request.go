package ce

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type CERequestID uint8

const (
	BlockRequest CERequestID = iota
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
		// TODO: Implement StateRequest encoding
		return nil, fmt.Errorf("StateRequest encoding not implemented yet")
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
	if blockReq, ok := message.(*BlockRequestMessage); ok {
		return h.encodeBlockRequestStruct(blockReq)
	}

	return nil, fmt.Errorf("unsupported message type for BlockRequest: %T", message)
}

type BlockRequestMessage struct {
	HeaderHash types.HeaderHash // Reference block hash
	Direction  byte             // 0: Ascending exclusive, 1: Descending inclusive
	MaxBlocks  uint32           // Maximum number of blocks requested
}

func (br *BlockRequestMessage) Encode(e *types.Encoder) error {
	// Encode HeaderHash
	if err := br.HeaderHash.Encode(e); err != nil {
		return fmt.Errorf("failed to encode HeaderHash: %w", err)
	}

	// Encode Direction as U8
	if err := e.WriteByte(br.Direction); err != nil {
		return fmt.Errorf("failed to encode Direction: %w", err)
	}

	// Encode MaxBlocks as U32
	maxBlocks := types.U32(br.MaxBlocks)
	if err := maxBlocks.Encode(e); err != nil {
		return fmt.Errorf("failed to encode MaxBlocks: %w", err)
	}

	return nil
}

func (h *DefaultCERequestHandler) encodeBlockRequestStruct(blockReq *BlockRequestMessage) ([]byte, error) {
	h.encoder = types.NewEncoder()

	if err := blockReq.Encode(h.encoder); err != nil {
		return nil, fmt.Errorf("failed to encode BlockRequestMessage: %w", err)
	}

	return h.encoder.Encode(blockReq)
}

func (h *DefaultCERequestHandler) Decode(data []byte) (CERequestID, interface{}, error) {
	// TODO: Implement decoding logic
	// This would involve:
	// 1. Reading the request type from the beginning of the data
	// 2. Based on the request type, decoding the appropriate message structure
	// 3. Returning the decoded request type and message

	return BlockRequest, nil, fmt.Errorf("decoding not implemented yet")
}
