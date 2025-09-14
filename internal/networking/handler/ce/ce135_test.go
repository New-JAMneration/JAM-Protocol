package ce

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandleWorkReportDistribution_Basic(t *testing.T) {
	// Create a minimal ReportGuarantee
	wp := types.WorkReport{}
	guarantee := types.ReportGuarantee{
		Report:     wp,
		Slot:       42,
		Signatures: []types.ValidatorSignature{{ValidatorIndex: 0}},
	}

	// Encode the guarantee
	encoder := types.NewEncoder()
	data, err := encoder.Encode(&guarantee)
	if err != nil {
		t.Fatalf("failed to encode ReportGuarantee: %v", err)
	}

	// Prefix with 4-byte little-endian length
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(data)))
	input := append(lenBuf, data...)
	input = append(input, []byte("FIN")...) // FIN for handler to read

	// Prepare the mock stream with the encoded guarantee and FIN
	stream := newMockStream(input)

	// Generate Ed25519 keypair
	_, priv, _ := ed25519.GenerateKey(nil)
	keypair, _ := keystore.FromEd25519PrivateKey(priv)

	err = HandleWorkReportDistribution_Guarantor(nil, &quic.Stream{Stream: stream}, keypair)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	resp := stream.w.Bytes()
	if !bytes.Contains(resp, []byte("FIN")) {
		t.Errorf("expected handler to write FIN, got %x", resp)
	}
}
