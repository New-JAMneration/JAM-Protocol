package utilities

import (
	"bytes"
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func TestWorkPackageSpecSerialization(t *testing.T) {
	var work_package_spec jamTypes.WorkPackageSpec
	result := WorkPackageSpecSerialization(work_package_spec)
	expectedOutput := make([]byte, 102)
	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("WorkPackageSpecSerialization() = %v, want %v", len(result), expectedOutput)
	}
}

func TestRefineContextSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultRefineContext := jamTypes.RefineContext{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 133)

	result := RefineContextSerialization(defaultRefineContext)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestWorkResultSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultWorkResult := jamTypes.WorkResult{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 77)

	result := WorkResultSerialization(defaultWorkResult)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestWorkReportSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultWorkReport := jamTypes.WorkReport{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 271)

	result := WorkReportSerialization(defaultWorkReport)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicGuaranteeSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultGuaranteesExtrinsic := jamTypes.GuaranteesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicGuaranteeSerialization(defaultGuaranteesExtrinsic)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicPreimageSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultExtrinsicPreimage := jamTypes.PreimagesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicPreimageSerialization(defaultExtrinsicPreimage)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicTicketSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultTicketsExtrinsic := jamTypes.TicketsExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicTicketSerialization(defaultTicketsExtrinsic)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicDisputeSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultDisputesExtrinsic := jamTypes.DisputesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 3)

	result := ExtrinsicDisputeSerialization(defaultDisputesExtrinsic)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicAssuranceSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultAssurancesExtrinsic := jamTypes.AssurancesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicAssuranceSerialization(defaultAssurancesExtrinsic)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestBlockSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultBlock := jamTypes.Block{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := jamTypes.ByteSequence{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 1, 0, 0, 0, 0, 0, 0, 0, 255, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	result := BlockSerialization(defaultBlock)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", len(result), len(expectedOutput))
	}
}

func TestHeaderSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultHeader := jamTypes.Header{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := jamTypes.ByteSequence{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 1, 0, 0, 0, 0, 0, 0, 0, 255, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	result := HeaderSerialization(defaultHeader)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}
