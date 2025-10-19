package ce

import (
	"encoding/binary"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type CE135Payload struct {
	CoreIndex uint32
	Report    types.WorkReport
	Signature types.Ed25519Signature
}

// HandleWorkReportDistribution_Validator implements distribution of a fully guaranteed work-report
// Role: Guarantor -> Validator
//
// [TODO-Validation]
// 1. Check valid slot (too far in past/future)
// [TODO]
// 1. Get all current validators.
// 2. Distributed to them.
func HandleWorkReportDistribution_Validator(
	blockchain blockchain.Blockchain,
	stream *quic.Stream,
	keypair keystore.KeyPair,
) error {
	lenBuf := make([]byte, 4)
	if _, err := stream.Read(lenBuf); err != nil {
		return fmt.Errorf("failed to read length prefix: %w", err)
	}
	guaranteeLen := binary.LittleEndian.Uint32(lenBuf)

	data := make([]byte, guaranteeLen)
	if _, err := stream.Read(data); err != nil {
		return fmt.Errorf("failed to read guarantee: %w", err)
	}

	var guarantee types.ReportGuarantee
	decoder := types.NewDecoder()
	if err := decoder.Decode(data, &guarantee); err != nil {
		return fmt.Errorf("failed to decode ReportGuarantee: %w", err)
	}

	currentVals := store.GetInstance().GetPosteriorStates().GetKappa()
	nextVals := store.GetInstance().GetPosteriorStates().GetGammaK()
	_ = currentVals
	_ = nextVals

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN: %w", err)
	}

	finResp := make([]byte, 3)
	if err := stream.ReadFull(finResp); err != nil {
		return fmt.Errorf("failed to read FIN response: %w", err)
	}
	if string(finResp) != "FIN" {
		return fmt.Errorf("expected FIN response, got %q", finResp)
	}

	return stream.Close()
}

func HandleWorkReportDistribution_Guarantor(
	blockchain blockchain.Blockchain,
	stream *quic.Stream,
	keypair keystore.KeyPair,
) error {
	lenBuf := make([]byte, 4)
	if err := stream.WriteMessage(lenBuf); err != nil {
		return fmt.Errorf("failed to read length prefix: %w", err)
	}
	guaranteeLen := binary.LittleEndian.Uint32(lenBuf)

	data := make([]byte, guaranteeLen)
	if err := stream.WriteMessage(data); err != nil {
		return fmt.Errorf("failed to read guarantee: %w", err)
	}

	var guarantee types.ReportGuarantee
	decoder := types.NewDecoder()
	if err := decoder.Decode(data, &guarantee); err != nil {
		return fmt.Errorf("failed to decode ReportGuarantee: %w", err)
	}

	currentVals := store.GetInstance().GetPosteriorStates().GetKappa()
	nextVals := store.GetInstance().GetPosteriorStates().GetGammaK()
	_ = currentVals
	_ = nextVals

	if err := stream.WriteMessage([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN: %w", err)
	}

	finResp := make([]byte, 3)
	if err := stream.ReadFull(finResp); err != nil {
		return fmt.Errorf("failed to read FIN response: %w", err)
	}
	if string(finResp) != "FIN" {
		return fmt.Errorf("expected FIN response, got %q", finResp)
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
