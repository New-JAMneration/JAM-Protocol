package ce

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// HandleWorkReportDistribution implements CE 135: distribution of a fully guaranteed work-report
// Accepts any io.ReadWriteCloser for testability, and allows injection of a keypair for signing
func HandleWorkReportDistribution(
	blockchain blockchain.Blockchain,
	stream io.ReadWriteCloser,
	keypair keystore.KeyPair,
) error {
	// Read 4-byte length prefix
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(stream, lenBuf); err != nil {
		return fmt.Errorf("failed to read length prefix: %w", err)
	}
	guaranteeLen := binary.LittleEndian.Uint32(lenBuf)

	// Read the encoded ReportGuarantee
	data := make([]byte, guaranteeLen)
	if _, err := io.ReadFull(stream, data); err != nil {
		return fmt.Errorf("failed to read guarantee: %w", err)
	}

	// Decode ReportGuarantee
	var guarantee types.ReportGuarantee
	decoder := types.NewDecoder()
	if err := decoder.Decode(data, &guarantee); err != nil {
		return fmt.Errorf("failed to decode ReportGuarantee: %w", err)
	}

	// Determine validator sets
	currentVals := store.GetInstance().GetPosteriorStates().GetKappa()
	nextVals := store.GetInstance().GetPosteriorStates().GetGammaK()
	_ = currentVals
	_ = nextVals

	// Write FIN to stream to indicate done
	if _, err := stream.Write([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN: %w", err)
	}

	// Expect FIN in response (simulate: read 3 bytes)
	finResp := make([]byte, 3)
	if _, err := io.ReadFull(stream, finResp); err != nil {
		return fmt.Errorf("failed to read FIN response: %w", err)
	}
	if string(finResp) != "FIN" {
		return fmt.Errorf("expected FIN response, got %q", finResp)
	}

	return stream.Close()
}
