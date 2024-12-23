package utilities

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func TestSerializeByteArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected jamTypes.ByteSequence
	}{
		{
			name:     "Empty input",
			input:    []byte{},
			expected: jamTypes.ByteSequence{}, // Adjust expected output based on WrapByteSequence and Serialize behavior
		},
		{
			name:     "Single byte input",
			input:    []byte{0x01},
			expected: jamTypes.ByteSequence{0x01}, // Adjust expected output based on WrapByteSequence and Serialize behavior
		},
		{
			name:     "Multiple byte input",
			input:    []byte{0x01, 0x02, 0x03},
			expected: jamTypes.ByteSequence{0x01, 0x02, 0x03}, // Adjust expected output based on WrapByteSequence and Serialize behavior
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeByteArray(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("SerializeByteArray(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSerializeU32(t *testing.T) {
	tests := []struct {
		name     string
		input    jam_types.U32
		expected jamTypes.ByteSequence
	}{
		{
			name:     "Zero",
			input:    jam_types.U32(0),
			expected: []byte{0x00}, // Adjust expected output based on SerializeU64 behavior
		},
		{
			name:     "Small value",
			input:    jam_types.U32(42),
			expected: []byte{255, 42, 0, 0, 0, 0, 0, 0, 0}, // Adjust expected output based on SerializeU64 behavior
		},
		{
			name:     "Large value",
			input:    jam_types.U32(0xffffffff),
			expected: []byte{240, 255, 255, 255, 255}, // Adjust expected output based on SerializeU64 behavior
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SerializeU32(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("SerializeU32(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWorkPackageSpecSerialization(t *testing.T) {
	var work_package_spec jam_types.WorkPackageSpec
	result := WorkPackageSpecSerialization(work_package_spec)
	expectedOutput := make([]byte, 102)
	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("WorkPackageSpecSerialization() = %v, want %v", len(result), expectedOutput)
	}
}

func TestRefineContextSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultRefineContext := jam_types.RefineContext{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 133)

	result := RefineContextSerialization(defaultRefineContext)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestWorkResultSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultWorkResult := jam_types.WorkResult{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 66)

	result := WorkResultSerialization(defaultWorkResult)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestWorkReportSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultWorkReport := jam_types.WorkReport{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 271)

	result := WorkReportSerialization(defaultWorkReport)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicGuaranteeSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultGuaranteesExtrinsic := jam_types.GuaranteesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicGuaranteeSerialization(defaultGuaranteesExtrinsic)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicPreimageSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultExtrinsicPreimage := jam_types.PreimagesExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicPreimageSerialization(defaultExtrinsicPreimage)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}

func TestExtrinsicTicketSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultTicketsExtrinsic := jam_types.TicketsExtrinsic{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 1)

	result := ExtrinsicTicketSerialization(defaultTicketsExtrinsic)

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
