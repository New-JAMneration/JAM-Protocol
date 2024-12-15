package extrinsic

import (
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func TestGuaranteeCredentialsInit(t *testing.T) {
	controller := NewGuaranteeCredentialsController()

	if controller.Len() != 0 {
		t.Fatalf("Expected 0 credentials, got %d", controller.Len())
	}
}

func TestGuaranteeCredentialsAdd(t *testing.T) {
	controller := NewGuaranteeCredentialsController()

	// Create a test credential
	credential := jamTypes.ValidatorSignature{
		ValidatorIndex: 0,
		Signature:      [64]byte{0x5f, 0x6e, 0x74, 0xd2, 0x04, 0xc2, 0x49, 0x0e, 0x71, 0xbe, 0x44, 0x51, 0x96, 0x3d, 0x7d, 0x7d, 0xa7, 0x97, 0xd4, 0xfd, 0x37, 0xd6, 0xe0, 0xbd, 0xa5, 0x69, 0x27, 0xd0, 0x2a, 0x33, 0x02, 0xca, 0x3b, 0x3a, 0x0e, 0x08, 0xc9, 0x61, 0xe7, 0x58, 0x0e, 0x97, 0xa0, 0xf0, 0x8c, 0x26, 0x9f, 0x54, 0x97, 0x28, 0xf5, 0x2d, 0x9c, 0x7d, 0xe3, 0xaf, 0xfe, 0x85, 0x0a, 0x03, 0x71, 0x38, 0x00, 0x12},
	}

	controller.Add(credential)

	if controller.Len() != 1 {
		t.Fatalf("Expected 1 credential, got %d", controller.Len())
	}

	// Verify the added credential's validator index
	if controller.Credentials[0].ValidatorIndex != 0 {
		t.Errorf("Expected validator index 0, got %d", controller.Credentials[0].ValidatorIndex)
	}

	// Verify the signature
	expectedSig := [64]byte{0x5f, 0x6e, 0x74, 0xd2, 0x04, 0xc2, 0x49, 0x0e, 0x71, 0xbe, 0x44, 0x51, 0x96, 0x3d, 0x7d, 0x7d, 0xa7, 0x97, 0xd4, 0xfd, 0x37, 0xd6, 0xe0, 0xbd, 0xa5, 0x69, 0x27, 0xd0, 0x2a, 0x33, 0x02, 0xca, 0x3b, 0x3a, 0x0e, 0x08, 0xc9, 0x61, 0xe7, 0x58, 0x0e, 0x97, 0xa0, 0xf0, 0x8c, 0x26, 0x9f, 0x54, 0x97, 0x28, 0xf5, 0x2d, 0x9c, 0x7d, 0xe3, 0xaf, 0xfe, 0x85, 0x0a, 0x03, 0x71, 0x38, 0x00, 0x12}
	if controller.Credentials[0].Signature != expectedSig {
		t.Error("Signature mismatch")
	}
}

func TestGuaranteeCredentialsRemoveDuplicatesEmpty(t *testing.T) {
	controller := NewGuaranteeCredentialsController()

	result := controller.RemoveDuplicates()

	if len(result) != 0 {
		t.Fatalf("Expected 0 credentials, got %d", len(result))
	}
}

func TestGuaranteeCredentialsRemoveDuplicates(t *testing.T) {
	controller := NewGuaranteeCredentialsController()

	// Add first credential
	credential1 := jamTypes.ValidatorSignature{
		ValidatorIndex: 0,
		Signature:      [64]byte{0x5f, 0x6e, 0x74, 0xd2, 0x04, 0xc2, 0x49, 0x0e, 0x71, 0xbe, 0x44, 0x51, 0x96, 0x3d, 0x7d, 0x7d},
	}
	controller.Add(credential1)

	// Add duplicate credential with same validator index but different signature
	duplicateCredential := jamTypes.ValidatorSignature{
		ValidatorIndex: 0,
		Signature:      [64]byte{0x3a, 0x68, 0x13, 0xf7, 0x69, 0x18, 0x95, 0xa4, 0x44, 0xd7, 0x2c, 0xad, 0x60, 0xe3, 0xd5, 0x4d},
	}
	controller.Add(duplicateCredential)

	// Add different credential
	credential2 := jamTypes.ValidatorSignature{
		ValidatorIndex: 1,
		Signature:      [64]byte{0xd3, 0x04, 0x51, 0x7e, 0x4d, 0xe8, 0x8a, 0x39, 0x9e, 0xd4, 0xd3, 0xfa, 0xa2, 0xfc, 0x86, 0xe3},
	}
	controller.Add(credential2)

	result := controller.RemoveDuplicates()

	if len(result) != 2 {
		t.Fatalf("Expected 2 credentials after removing duplicates, got %d", len(result))
	}

	// Verify the remaining credentials have different validator indices
	if result[0].ValidatorIndex == result[1].ValidatorIndex {
		t.Error("Duplicate validator indices found after removing duplicates")
	}
}

func TestGuaranteeCredentialsSort(t *testing.T) {
	controller := NewGuaranteeCredentialsController()

	// Add credentials in unsorted order
	credentials := []jamTypes.ValidatorSignature{
		{
			ValidatorIndex: 2,
			Signature:      [64]byte{0x02},
		},
		{
			ValidatorIndex: 0,
			Signature:      [64]byte{0x00},
		},
		{
			ValidatorIndex: 1,
			Signature:      [64]byte{0x01},
		},
	}

	for _, c := range credentials {
		controller.Add(c)
	}

	controller.Sort()

	// Verify sorting
	for i := 1; i < len(controller.Credentials); i++ {
		if controller.Credentials[i].ValidatorIndex < controller.Credentials[i-1].ValidatorIndex {
			t.Errorf("Credentials not properly sorted at index %d", i)
		}
	}

	// Verify the first credential has the smallest validator index
	if controller.Credentials[0].ValidatorIndex != 0 {
		t.Errorf("Expected first credential to have validator index 0, got %d", controller.Credentials[0].ValidatorIndex)
	}
}
