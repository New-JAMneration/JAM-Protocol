package utilities

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestWorkPackageSpecSerialization(t *testing.T) {
	var work_package_spec types.WorkPackageSpec
	result := WorkPackageSpecSerialization(work_package_spec)
	expectedOutput := make([]byte, 102)
	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("WorkPackageSpecSerialization() = %v, want %v", len(result), expectedOutput)
	}
}

func TestRefineContextSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultRefineContext := types.RefineContext{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 133)

	result := RefineContextSerialization(defaultRefineContext)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestWorkResultSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultWorkResult := types.WorkResult{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 76)

	result := WorkResultSerialization(defaultWorkResult)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestWorkReportSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultWorkReport := types.WorkReport{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 271)

	result := WorkReportSerialization(defaultWorkReport)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicGuaranteeSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultGuaranteesExtrinsic := types.GuaranteesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicGuaranteeSerialization(defaultGuaranteesExtrinsic)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicPreimageSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultExtrinsicPreimage := types.PreimagesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicPreimageSerialization(defaultExtrinsicPreimage)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicTicketSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultTicketsExtrinsic := types.TicketsExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicTicketSerialization(defaultTicketsExtrinsic)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicDisputeSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultDisputesExtrinsic := types.DisputesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 3)

	result := ExtrinsicDisputeSerialization(defaultDisputesExtrinsic)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicAssuranceSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultAssurancesExtrinsic := types.AssurancesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicAssuranceSerialization(defaultAssurancesExtrinsic)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestBlockSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultBlock := types.Block{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := types.ByteSequence{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 1, 0, 0, 0, 0, 0, 0, 0, 255, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	result := BlockSerialization(defaultBlock)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", len(result), len(expectedOutput))
	}
}

func TestHeaderSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultHeader := types.Header{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := types.ByteSequence{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 1, 0, 0, 0, 0, 0, 0, 0, 255, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	result := HeaderSerialization(defaultHeader)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}
