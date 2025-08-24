package ce

import (
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// WorkReportLookupFunc defines a function to look up a WorkReport by hash
type WorkReportLookupFunc func(hash types.WorkReportHash) (*types.WorkReport, bool)

type CE136Payload struct {
	CoreIndex uint32
}

// HandleWorkReportRequest implements CE 136: Auditor -> Auditor work-report request
//
// [TODO]
// 1. Ensure auditors who pushing or forwarding a negative judgment be queried first for missing report.
func HandleWorkReportRequest(
	stream io.ReadWriteCloser,
	lookup WorkReportLookupFunc,
) error {
	hashBuf := make([]byte, 32)
	if _, err := io.ReadFull(stream, hashBuf); err != nil {
		return fmt.Errorf("failed to read work-report hash: %w", err)
	}
	var hash types.WorkReportHash
	copy(hash[:], hashBuf)

	finBuf := make([]byte, 3)
	if _, err := io.ReadFull(stream, finBuf); err != nil {
		return fmt.Errorf("failed to read FIN: %w", err)
	}
	if string(finBuf) != "FIN" {
		return fmt.Errorf("expected FIN, got %q", finBuf)
	}

	workReport, found := lookup(hash)
	if !found {
		return fmt.Errorf("work-report not found for hash %x", hash)
	}

	encoder := types.NewEncoder()
	data, err := encoder.Encode(workReport)
	if err != nil {
		return fmt.Errorf("failed to encode work-report: %w", err)
	}

	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to write work-report: %w", err)
	}

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN: %w", err)
	}

	return stream.Close()
}

func (h *DefaultCERequestHandler) encodeWorkReportRequest(message interface{}) ([]byte, error) {
	workReportReq, ok := message.(*CE136Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for WorkReportRequest: %T", message)
	}

	encoder := types.NewEncoder()

	if err := h.writeBytes(encoder, encodeLE32(workReportReq.CoreIndex)); err != nil {
		return nil, fmt.Errorf("failed to encode CoreIndex: %w", err)
	}

	result := make([]byte, 0, 4) // 4 bytes for CoreIndex
	result = append(result, encodeLE32(workReportReq.CoreIndex)...)
	return result, nil
}
