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
	errCode := ValidateHeaderSeal(trace.Block.Header, &State)
	if errCode != nil {
		t.Fatalf("ValidateHeaderSeal failed: %v", errCode)
	}
}

// v0.7.0
func TestValidateSealUsingTraceFile(t *testing.T) {
	t.Skip("Skipping outdated 0.7.0 fuzz reports")
	runSealValidationTraceTest(t, "1758621879", "00000347.bin")
	runSealValidationTraceTest(t, "1758621879", "00000348.bin")
	runSealValidationTraceTest(t, "1758622313", "00000012.bin")
	runSealValidationTraceTest(t, "1757092821", "00000157.bin")
}

func TestSignHeaderEntropy_RoundTrip(t *testing.T) {
	validators, err := LoadTinyValidatorsData()
	if err != nil {
		t.Fatalf("LoadTinyValidatorsData: %v", err)
	}
	validator := validators[0]
	sk, err := LookupBandersnatchSecretSeed(validator.Bandersnatch)
	if err != nil {
		t.Fatalf("LookupBandersnatchSecretSeed: %v", err)
	}

	sealBytes, err := vrf.IETFSign(sk, []byte("test-seal-context"), nil)
	if err != nil {
		t.Fatalf("IETFSign seal: %v", err)
	}
	if len(sealBytes) != types.BandersnatchSigSize {
		t.Fatalf("seal length: got %d want %d", len(sealBytes), types.BandersnatchSigSize)
	}

	var seal types.BandersnatchVrfSignature
	copy(seal[:], sealBytes)

	hv, err := SignHeaderEntropy(sk, seal)
	if err != nil {
		t.Fatalf("SignHeaderEntropy: %v", err)
	}

	header := types.Header{
		Seal:          seal,
		EntropySource: hv,
		AuthorIndex:   0,
	}
	state := &types.State{
		Kappa: types.ValidatorsData{{Bandersnatch: validator.Bandersnatch}},
	}
	if errCode := ValidateHeaderEntropy(header, state); errCode != nil {
		t.Fatalf("ValidateHeaderEntropy failed: %v", *errCode)
	}
}
