package ce

import (
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// WorkReportLookupFunc defines a function to look up a WorkReport by hash
// This allows injection of a test map or real store
// Returns (workReport, found)
type WorkReportLookupFunc func(hash types.WorkReportHash) (*types.WorkReport, bool)

// HandleWorkReportRequest implements CE 136: Auditor -> Auditor work-report request
// Reads a 32-byte work-report hash and FIN, looks up the work-report, writes the encoded work-report as a framed message and FIN.
func HandleWorkReportRequest(
	stream quic.MessageStream,
	lookup WorkReportLookupFunc,
) error {
	hashBuf := make([]byte, 32)
	if _, err := io.ReadFull(stream, hashBuf); err != nil {
		return fmt.Errorf("failed to read work-report hash: %w", err)
	}
	var hash types.WorkReportHash
	copy(hash[:], hashBuf)

	if err := expectRemoteFIN(stream); err != nil {
		return fmt.Errorf("expected FIN after hash: %w", err)
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

	if err := stream.WriteMessage(data); err != nil {
		return fmt.Errorf("failed to write work-report: %w", err)
	}
	// Send FIN by closing write half
	return stream.Close()
}

func (h *DefaultCERequestHandler) encodeWorkReportRequest(message interface{}) ([]byte, error) {
	workReportReq, ok := message.(*CE136Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for WorkReportRequest: %T", message)
	}
	return workReportReq.WorkReportHash[:], nil
}
