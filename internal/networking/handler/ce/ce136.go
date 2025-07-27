package ce

import (
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// WorkReportLookupFunc defines a function to look up a WorkReport by hash
// This allows injection of a test map or real store
// Returns (workReport, found)
type WorkReportLookupFunc func(hash types.WorkReportHash) (*types.WorkReport, bool)

// HandleWorkReportRequest implements CE 136: Auditor -> Auditor work-report request
// Reads a 32-byte work-report hash and 'FIN', looks up the work-report, writes the encoded work-report and 'FIN'.
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
