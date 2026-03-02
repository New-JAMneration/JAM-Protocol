package ce

import (
	"crypto/ed25519"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandleWorkReportDistribution_Basic(t *testing.T) {
	// Create a minimal ReportGuarantee (spec: Work-Report ++ Slot ++ len++[ValidatorIndex ++ Ed25519Signature])
	wp := types.WorkReport{}
	guarantee := types.ReportGuarantee{
		Report:     wp,
		Slot:       42,
		Signatures: []types.ValidatorSignature{{ValidatorIndex: 0}},
	}

	encoder := types.NewEncoder()
	data, err := encoder.Encode(&guarantee)
	if err != nil {
		t.Fatalf("failed to encode ReportGuarantee: %v", err)
	}

	// Send only ReportGuarantee bytes, then stream close (FIN)
	stream := newMockStream(data)

	_, priv, _ := ed25519.GenerateKey(nil)
	keypair, _ := keystore.FromEd25519PrivateKey(priv)

	err = HandleWorkReportDistribution(nil, stream, keypair)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
}
