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
	SegmentShardRequestWithJustification
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

func (h *DefaultCERequestHandler) writeBytes(encoder *types.Encoder, data []byte) error {
	for _, b := range data {
		if err := encoder.WriteByte(b); err != nil {
			return fmt.Errorf("failed to write byte: %w", err)
		}
	}
	return nil
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
		return h.encodeSafroleTicketDistribution(message)
	case WorkPackageSubmission:
		return h.encodeWorkPackageSubmission(message)
	case WorkPackageSharing:
		return h.encodeWorkPackageSharing(message)
	case WorkReportDistribution:
		return h.encodeWorkReportDistribution(message)
	case WorkReportRequest:
		return h.encodeWorkReportRequest(message)
	case ShardDistribution:
		return h.encodeShardDistribution(message)
	case AuditShardReqeust:
		return h.encodeAuditShardRequest(message)
	case SegmentShardRequest:
		return h.encodeSegmentShardRequest(message)
	case SegmentShardRequestWithJustification:
		return h.encodeSegmentShardRequestWithJustification(message)
	case AssuranceDistribution:
		return h.encodeAssuranceDistribution(message)
	case PreimageAnnouncement:
		return h.encodePreimageAnnouncement(message)
	case PreimageRequest:
		return h.encodePreimageRequest(message)
	case AuditAnnouncement:
		return h.encodeAuditAnnouncement(message)
	case JudgmentPublication:
		return h.encodeJudgmentPublication(message)
	default:
		return nil, fmt.Errorf("unknown request type: %d", req)
	}
}

func (h *DefaultCERequestHandler) Decode(data []byte) (CERequestID, interface{}, error) {
	// TODO: Implement decoding logic
	// This would involve:
	// 1. Reading the request type from the beginning of the data
	// 2. Based on the request type, decoding the appropriate message structure
	// 3. Returning the decoded request type and message

	return BlockRequest, nil, fmt.Errorf("decoding not implemented yet")
}
