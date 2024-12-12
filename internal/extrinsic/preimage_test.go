package extrinsic

import (
	"testing"
)

func TestAdd(t *testing.T) {
	var preimages Preimages

	// Add a new Preimage
	preimage1 := Preimage{
		Requester: 16909060,
		Blob:      "0x81095e6122e3bc9d961e00014a7fc833",
	}

	preimages.Add(preimage1)

	if len(preimages) != 1 {
		t.Fatalf("Expected 1 preimage, got %d", len(preimages))
	}
}

func TestSort(t *testing.T) {
	var preimages Preimages

	// Add a new Preimage
	preimage1 := Preimage{
		Requester: 16909060,
		Blob:      "0x81095e6122e3bc9d961e00014a7fc833",
	}
	preimages.Add(preimage1)

	// Add a new Preimage
	preimage2 := Preimage{
		Requester: 16909062,
		Blob:      "0x38db056c7c3065fadb630ce6ccbc7385",
	}
	preimages.Add(preimage2)

	// Add a new Preimage
	preimage3 := Preimage{
		Requester: 16909061,
		Blob:      "0xd257bc7d93a55be3561d720d40a6a342",
	}
	preimages.Add(preimage3)

	expectedRequesters := []uint32{16909060, 16909061, 16909062}
	expectedBlobs := []string{
		"0x81095e6122e3bc9d961e00014a7fc833",
		"0xd257bc7d93a55be3561d720d40a6a342",
		"0x38db056c7c3065fadb630ce6ccbc7385",
	}

	if len(preimages) != len(expectedRequesters) {
		t.Fatalf("Expected %d preimages, got %d", len(expectedRequesters), len(preimages))
	}

	for i, preimage := range preimages {
		if preimage.Requester != expectedRequesters[i] {
			t.Errorf("Expected requester %d, got %d", expectedRequesters[i], preimage.Requester)
		}
		if preimage.Blob != expectedBlobs[i] {
			t.Errorf("Expected blob %s, got %s", expectedBlobs[i], preimage.Blob)
		}
	}
}

func TestAddDuplicate(t *testing.T) {
	var preimages Preimages

	// Add a new Preimage
	preimage1 := Preimage{
		Requester: 16909060,
		Blob:      "0x81095e6122e3bc9d961e00014a7fc833",
	}
	preimages.Add(preimage1)

	// Add a duplicate Preimage
	duplicatePreimage := Preimage{
		Requester: 16909060,
		Blob:      "0x81095e6122e3bc9d961e00014a7fc833",
	}
	preimages.Add(duplicatePreimage)

	if len(preimages) != 1 {
		t.Fatalf("Expected 1 preimage, got %d", len(preimages))
	}
}
