package utilities

import (
	"bytes"
	"encoding/hex"
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
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
	// Using jamtestnet test data safrole/blocks/425530_001.json
	blockHeader := jamTypes.Header{
		Parent: jamTypes.HeaderHash{
			0x92, 0x4d, 0x48, 0xaa, 0x6e, 0x10, 0x0a, 0x08,
			0x58, 0x37, 0x2e, 0x1e, 0xc9, 0x8b, 0xdc, 0xd4,
			0x26, 0xd5, 0x66, 0xd9, 0x8c, 0x5e, 0xe2, 0xa8,
			0x3d, 0xc2, 0xe2, 0x4e, 0xe4, 0x43, 0xf5, 0xbc,
		},
		ParentStateRoot: jamTypes.StateRoot{
			0xe0, 0x6d, 0xd5, 0xcd, 0x5c, 0x94, 0x93, 0xf2,
			0xcb, 0x42, 0xf5, 0xa5, 0x08, 0x77, 0x60, 0x1b,
			0xac, 0xfc, 0x71, 0xe2, 0xd3, 0xd5, 0xaa, 0xa4,
			0xf6, 0x7a, 0xb7, 0x53, 0x06, 0x7f, 0x18, 0xe0,
		},
		ExtrinsicHash: jamTypes.OpaqueHash{
			0x8f, 0xc8, 0x20, 0xed, 0x70, 0x63, 0x83, 0x1d,
			0xe1, 0xc4, 0x45, 0xdf, 0xe2, 0x68, 0xf6, 0x5c,
			0x11, 0x17, 0xa0, 0x0c, 0x30, 0x97, 0xb4, 0xcf,
			0x1e, 0x77, 0x76, 0x70, 0xd5, 0xfe, 0x97, 0xaa,
		},
		Slot:          5106361,
		EpochMark:     nil,
		TicketsMark:   nil,
		OffendersMark: []jamTypes.Ed25519Public{},
		AuthorIndex:   0,
		EntropySource: jamTypes.BandersnatchVrfSignature{
			0x4c, 0x99, 0x68, 0xe1, 0x33, 0xa4, 0x8e, 0x16,
			0xc7, 0xed, 0xeb, 0x1d, 0xf2, 0x82, 0xde, 0x7d,
			0x68, 0x70, 0x5c, 0x99, 0xb2, 0xc1, 0x87, 0x28,
			0xd0, 0x58, 0x93, 0xe6, 0x4c, 0x4c, 0x7d, 0x11,
			0x90, 0x88, 0xca, 0x51, 0xbd, 0x71, 0x26, 0x15,
			0xeb, 0x5c, 0x16, 0x8d, 0x47, 0xeb, 0xf8, 0x0c,
			0x2e, 0x9b, 0xc0, 0x1b, 0x93, 0x42, 0xa9, 0xf0,
			0xe0, 0x77, 0x75, 0xbe, 0xb6, 0x1f, 0x9e, 0x1b,
			0x6f, 0x4b, 0xd0, 0xf5, 0x6c, 0xf6, 0x9d, 0x8e,
			0x1c, 0xc3, 0x7f, 0x1f, 0xdf, 0x36, 0xf6, 0xd7,
			0xd1, 0x3d, 0x74, 0xfb, 0xd5, 0xd7, 0x82, 0x06,
			0xdc, 0xde, 0x95, 0xae, 0x7c, 0xe2, 0x9c, 0x1b,
		},
		Seal: jamTypes.BandersnatchVrfSignature{
			0x29, 0x0c, 0xf0, 0x94, 0x4d, 0x46, 0xcc, 0xf7,
			0xb7, 0xc6, 0xe9, 0xdb, 0xe5, 0xeb, 0x2c, 0x9e,
			0x4f, 0xef, 0x2f, 0x2c, 0x85, 0xf0, 0x50, 0xb1,
			0xb4, 0x05, 0x71, 0xed, 0x3a, 0x2c, 0x38, 0x12,
			0xa1, 0x35, 0x7f, 0xdb, 0x12, 0x86, 0xca, 0xde,
			0x6e, 0x90, 0x9b, 0xce, 0x81, 0x1b, 0x33, 0x40,
			0x83, 0x26, 0xfa, 0xf0, 0xb6, 0x51, 0x19, 0x38,
			0x6d, 0x5f, 0x3e, 0x7e, 0x48, 0xcd, 0xa4, 0x05,
			0x01, 0x1c, 0x9d, 0x50, 0x9a, 0x5a, 0x90, 0x3a,
			0x69, 0x1b, 0x15, 0xcb, 0x21, 0x6f, 0x0a, 0x70,
			0xc1, 0x29, 0x81, 0x52, 0x33, 0x97, 0xc3, 0xa4,
			0x95, 0xa3, 0x6d, 0xd1, 0xe9, 0x56, 0xe2, 0x05,
		},
	}
	// Expected output, safrole/blocks/425530_002.json parent header hash
	expectedHex := "e546d686892908dd69ec67dd7f8dfa5f6169b408e4b2044f231c8c7df40ef23a"
	result := HeaderSerialization(blockHeader)
	resultHash := hash.Blake2bHash(result)
	if hex.EncodeToString(resultHash[:]) != expectedHex {
		t.Errorf("Expected: %v, got: %v", expectedHex, hex.EncodeToString(resultHash[:]))
	}
}
