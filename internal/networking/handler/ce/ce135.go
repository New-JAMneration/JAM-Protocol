package ce

import (
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// HandleWorkReportDistribution implements CE 135: distribution of a fully guaranteed work-report.
// Spec format: Work-Report ++ Slot ++ len++[ValidatorIndex ++ Ed25519Signature] then FIN.
func HandleWorkReportDistribution(
	_ blockchain.Blockchain,
	stream io.ReadWriteCloser,
	keypair keystore.KeyPair,
) error {
	data, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("failed to read work-report distribution: %w", err)
	}

	var guarantee types.ReportGuarantee
	decoder := types.NewDecoder()
	if err := decoder.Decode(data, &guarantee); err != nil {
		return fmt.Errorf("failed to decode ReportGuarantee: %w", err)
	}

	_ = keypair
	_ = guarantee

	if err := expectRemoteFIN(stream); err != nil {
		return err
	}
	return stream.Close()
}

func (h *DefaultCERequestHandler) encodeWorkReportDistribution(message interface{}) ([]byte, error) {
	payload, ok := message.(*CE135Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for WorkReportDistribution: %T", message)
	}

	if payload == nil {
		return nil, fmt.Errorf("nil payload for WorkReportDistribution")
	}

	guarantee := types.ReportGuarantee{
		Report:     payload.Report,
		Slot:       payload.Slot,
		Signatures: payload.Signatures,
	}
	if guarantee.Signatures == nil {
		guarantee.Signatures = []types.ValidatorSignature{}
	}

	encoder := types.NewEncoder()
	return encoder.Encode(&guarantee)
}
