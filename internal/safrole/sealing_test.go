package safrole

import (
	"encoding/hex"
	"path/filepath"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	jamtests_trace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

func hexToByteArray32(hexString string) types.ByteArray32 {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		return types.ByteArray32{}
	}

	if len(bytes) != 32 {
		return types.ByteArray32{}
	}

	var result types.ByteArray32
	copy(result[:], bytes[:])

	return result
}

func runSealValidationTraceTest(t *testing.T, dir string, file string) {
	t.Helper()

	// Locate trace path
	tracePath := filepath.Join("..", "..",
		"pkg", "test_data", "jam-conformance",
		"fuzz-reports", "0.7.0", "traces",
		dir, file,
	)

	// --- Load Trace File ---
	trace := &jamtests_trace.TraceTestCase{}
	err := utilities.GetTestFromBin(tracePath, trace)
	if err != nil {
		t.Fatalf("Failed to load trace file %s: %v", tracePath, err)
	}

	// --- Parse State ---
	State, _, err := merklization.StateKeyValsToState(trace.PostState.KeyVals)
	if err != nil {
		t.Fatalf("Failed to parse PreState: %v", err)
	}

	// --- Validate Seal ---
	ring := []byte{}
	for _, validator := range State.Kappa {
		ring = append(ring, []byte(validator.Bandersnatch[:])...)
	}
	verifier, err := vrf.NewVerifier(ring, uint(len(State.Kappa)))
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}
	defer verifier.Free()
	errCode := ValidateHeaderSeal(verifier, trace.Block.Header, &State)
	if errCode != nil {
		t.Fatalf("ValidateHeaderSeal failed: %v", errCode)
	}
}

// v0.7.0
func TestValidateSealUsingTraceFile(t *testing.T) {
	runSealValidationTraceTest(t, "1758621879", "00000347.bin")
	runSealValidationTraceTest(t, "1758621879", "00000348.bin")
	runSealValidationTraceTest(t, "1758622313", "00000012.bin")
	runSealValidationTraceTest(t, "1757092821", "00000157.bin")
}
