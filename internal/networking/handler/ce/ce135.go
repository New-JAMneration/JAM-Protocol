package ce

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// HandleWorkReportDistribution implements CE 135: distribution of a fully guaranteed work-report
// Accepts any io.ReadWriteCloser for testability, and allows injection of a keypair for signing
func HandleWorkReportDistribution(
	_ blockchain.Blockchain,
	stream io.ReadWriteCloser,
	keypair keystore.KeyPair,
) error {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(stream, lenBuf); err != nil {
		return fmt.Errorf("failed to read length prefix: %w", err)
	}
	guaranteeLen := binary.LittleEndian.Uint32(lenBuf)

	data := make([]byte, guaranteeLen)
	if _, err := io.ReadFull(stream, data); err != nil {
		return fmt.Errorf("failed to read guarantee: %w", err)
	}

	var guarantee types.ReportGuarantee
	decoder := types.NewDecoder()
	if err := decoder.Decode(data, &guarantee); err != nil {
		return fmt.Errorf("failed to decode ReportGuarantee: %w", err)
	}

	// Determine validator sets
	cs := blockchain.GetInstance()
	currentVals := cs.GetPosteriorStates().GetKappa()
	nextVals := cs.GetPosteriorStates().GetGammaK()
	_ = currentVals
	_ = nextVals

	// Peer signals FIN by closing send half; we expect EOF after the message
	if err := expectRemoteFIN(stream); err != nil {
		return err
	}
	return stream.Close()
}

func (h *DefaultCERequestHandler) encodeWorkReportDistribution(message interface{}) ([]byte, error) {
	workReport, ok := message.(*CE135Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for WorkReportDistribution: %T", message)
	}

	if workReport == nil {
		return nil, fmt.Errorf("nil payload for WorkReportDistribution")
	}

	// Encode request type
	requestType := byte(WorkReportDistribution)

	// Encode CoreIndex (4 bytes little-endian)
	coreIndexBytes := encodeLE32(workReport.CoreIndex)

	// Get WorkReport bytes using ScaleEncode
	reportBytes, err := workReport.Report.ScaleEncode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode WorkReport: %w", err)
	}

	// Build the final result
	totalLen := 1 + 4 + len(reportBytes) + len(workReport.Signature) // 1 byte for request type + 4 bytes for CoreIndex + report bytes + signature bytes
	result := make([]byte, 0, totalLen)

	// Append all components
	result = append(result, requestType)
	result = append(result, coreIndexBytes...)
	result = append(result, reportBytes...)
	result = append(result, workReport.Signature[:]...)

	return result, nil
}
