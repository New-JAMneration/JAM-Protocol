package ce

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/bits"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// DecodePayload decodes a CE request body when the protocol ID is carried separately (stream kind byte).
func (h *DefaultCERequestHandler) DecodePayload(req CERequestID, data []byte) (interface{}, error) {
	switch req {
	case BlockRequest:
		return decodeBlockRequest(data)
	case StateRequest:
		return decodeStateRequest(data)
	case CE131SafroleTicketDistribution, CE132SafroleTicketDistribution:
		return decodeSafroleTicketDistribution(data)
	case WorkReportRequest:
		return decodeWorkReportRequest(data)
	case AuditShardReqeust:
		return decodeAuditShardRequest(data)
	case SegmentShardRequest:
		return decodeSegmentShardRequest(data)
	case SegmentShardRequestWithJustification:
		return decodeSegmentShardRequestWithJustification(data)
	case BundleRequest:
		return decodeBundleRequest(data)
	case WorkPackageSubmission:
		return decodeWorkPackageSubmission(data)
	case WorkPackageSharing:
		return decodeWorkPackageSharing(data)
	case WorkReportDistribution:
		return decodeWorkReportDistribution(data)
	case ShardDistribution:
		return nil, fmt.Errorf("shard distribution decode requires stream framing")
	case AssuranceDistribution:
		return decodeAssuranceDistribution(data)
	case PreimageAnnouncement:
		return decodePreimageAnnouncement(data)
	case PreimageRequest:
		return decodePreimageRequest(data)
	case AuditAnnouncement:
		return decodeAuditAnnouncement(data)
	case JudgmentPublication:
		return decodeJudgmentPublication(data)
	default:
		return nil, fmt.Errorf("unknown request type: %d", req)
	}
}

func splitLeadingRequestID(data []byte) (CERequestID, []byte, error) {
	if len(data) == 0 {
		return 0, nil, fmt.Errorf("empty CE request payload")
	}
	dec := types.NewDecoder()
	id, err := dec.DecodeUint(data)
	if err != nil {
		return 0, nil, fmt.Errorf("decode request type: %w", err)
	}
	consumed := compactEncodedLength(data)
	if consumed > len(data) {
		return 0, nil, fmt.Errorf("truncated request type")
	}
	reqID := CERequestID(id)
	if !isKnownCERequestID(reqID) {
		return 0, nil, fmt.Errorf("unknown request type: %d", reqID)
	}
	return reqID, data[consumed:], nil
}

func compactEncodedLength(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	prefix := data[0]
	if prefix < 0x80 {
		return 1
	}
	if prefix == 0xFF {
		return 9
	}
	return 1 + int(bits.LeadingZeros8(^prefix))
}

func isKnownCERequestID(id CERequestID) bool {
	switch id {
	case BlockRequest, StateRequest,
		CE131SafroleTicketDistribution, CE132SafroleTicketDistribution,
		WorkPackageSubmission, WorkPackageSharing, WorkReportDistribution,
		WorkReportRequest, ShardDistribution, AuditShardReqeust,
		SegmentShardRequest, SegmentShardRequestWithJustification,
		AssuranceDistribution, PreimageAnnouncement, PreimageRequest,
		AuditAnnouncement, JudgmentPublication, BundleRequest:
		return true
	default:
		return false
	}
}

func prependRequestID(req CERequestID, payload []byte) ([]byte, error) {
	prefix, err := types.NewEncoder().EncodeUint(uint64(req))
	if err != nil {
		return nil, err
	}
	return append(prefix, payload...), nil
}

func decodeBlockRequest(data []byte) (*CE128Payload, error) {
	if len(data) < CE128MinRequestSize {
		return nil, fmt.Errorf("block request too short: got %d want >= %d", len(data), CE128MinRequestSize)
	}
	var req CE128Payload
	copy(req.HeaderHash[:], data[:HashSize])
	req.Direction = data[HashSize]
	req.MaxBlocks = binary.LittleEndian.Uint32(data[HashSize+1 : HashSize+1+U32Size])
	return &req, nil
}

func decodeStateRequest(data []byte) (*CE129Payload, error) {
	if len(data) < CE129RequestSize {
		return nil, fmt.Errorf("state request too short: got %d want %d", len(data), CE129RequestSize)
	}
	var req CE129Payload
	copy(req.HeaderHash[:], data[:HashSize])
	copy(req.KeyStart[:], data[HashSize:HashSize+StateKeySize])
	copy(req.KeyEnd[:], data[HashSize+StateKeySize:HashSize+StateKeySize*2])
	req.MaxSize = binary.LittleEndian.Uint32(data[CE129RequestSize-U32Size : CE129RequestSize])
	return &req, nil
}

func decodeSafroleTicketDistribution(data []byte) (*CE131Payload, error) {
	if len(data) < CE131PayloadSize {
		return nil, fmt.Errorf("safrole ticket distribution too short: got %d want %d", len(data), CE131PayloadSize)
	}
	var payload CE131Payload
	payload.EpochIndex = binary.LittleEndian.Uint32(data[:U32Size])
	payload.Attempt = data[U32Size]
	copy(payload.Proof[:], data[U32Size+1:CE131PayloadSize])
	return &payload, nil
}

func decodeWorkReportRequest(data []byte) (*CE136Payload, error) {
	if len(data) < HashSize {
		return nil, fmt.Errorf("work report request too short: got %d want %d", len(data), HashSize)
	}
	var payload CE136Payload
	copy(payload.WorkReportHash[:], data[:HashSize])
	return &payload, nil
}

func decodeAuditShardRequest(data []byte) (*CE138Payload, error) {
	if len(data) < CE138RequestSize {
		return nil, fmt.Errorf("audit shard request too short: got %d want %d", len(data), CE138RequestSize)
	}
	return &CE138Payload{
		ErasureRoot: bytes.Clone(data[:HashSize]),
		ShardIndex:  uint32(binary.LittleEndian.Uint16(data[HashSize:CE138RequestSize])),
	}, nil
}

func decodeSegmentShardRequest(data []byte) (*CE139Payload, error) {
	return decodeSegmentShardRequestPayload(data)
}

func decodeSegmentShardRequestWithJustification(data []byte) (*CE140Payload, error) {
	base, err := decodeSegmentShardRequestPayload(data)
	if err != nil {
		return nil, err
	}
	return &CE140Payload{
		ErasureRoot:    base.ErasureRoot,
		ShardIndex:     base.ShardIndex,
		SegmentIndices: base.SegmentIndices,
	}, nil
}

func decodeSegmentShardRequestPayload(data []byte) (*CE139Payload, error) {
	if len(data) < CE139140MinRequestSize {
		return nil, fmt.Errorf("segment shard request too short: got %d want >= %d", len(data), CE139140MinRequestSize)
	}
	payload := &CE139Payload{
		ErasureRoot: bytes.Clone(data[:HashSize]),
		ShardIndex:  uint32(binary.LittleEndian.Uint16(data[HashSize : HashSize+U16Size])),
	}
	count := binary.LittleEndian.Uint16(data[HashSize+U16Size : HashSize+U16Size*2])
	offset := HashSize + U16Size*2
	payload.SegmentIndices = make([]uint16, 0, count)
	for i := uint16(0); i < count; i++ {
		if offset+U16Size > len(data) {
			return nil, fmt.Errorf("segment shard request truncated at index %d", i)
		}
		payload.SegmentIndices = append(payload.SegmentIndices, binary.LittleEndian.Uint16(data[offset:offset+U16Size]))
		offset += U16Size
	}
	return payload, nil
}

func decodeBundleRequest(data []byte) (*CE147Payload, error) {
	if len(data) != CE147RequestSize {
		return nil, fmt.Errorf("bundle request: expected %d bytes, got %d", CE147RequestSize, len(data))
	}
	return &CE147Payload{ErasureRoot: bytes.Clone(data)}, nil
}

func decodeWorkPackageSubmission(data []byte) (*CE133WorkPackageSubmission, error) {
	if len(data) < U16Size {
		return nil, fmt.Errorf("work package submission too short")
	}
	coreIndex := types.CoreIndex(binary.LittleEndian.Uint16(data[:U16Size]))
	return &CE133WorkPackageSubmission{
		CoreIndex:   coreIndex,
		WorkPackage: bytes.Clone(data[U16Size:]),
	}, nil
}

func decodeWorkPackageSharing(data []byte) (interface{}, error) {
	return nil, fmt.Errorf("work package sharing decode requires stream framing")
}

func decodeWorkReportDistribution(data []byte) (interface{}, error) {
	return nil, fmt.Errorf("work report distribution decode requires stream framing")
}

func decodeAssuranceDistribution(data []byte) (interface{}, error) {
	var payload CE141Payload
	if err := payload.Decode(data); err != nil {
		return nil, err
	}
	return &payload, nil
}

func decodePreimageAnnouncement(data []byte) (interface{}, error) {
	var payload CE142Payload
	if err := payload.Decode(data); err != nil {
		return nil, err
	}
	return &payload, nil
}

func decodePreimageRequest(data []byte) (interface{}, error) {
	var payload CE143Payload
	if err := payload.Decode(data); err != nil {
		return nil, err
	}
	return &payload, nil
}

func decodeAuditAnnouncement(data []byte) (interface{}, error) {
	var payload CE144Payload
	if err := payload.Decode(data); err != nil {
		return nil, err
	}
	return &payload, nil
}

func decodeJudgmentPublication(data []byte) (interface{}, error) {
	var payload CE145Payload
	if err := payload.Decode(data); err != nil {
		return nil, err
	}
	return &payload, nil
}
