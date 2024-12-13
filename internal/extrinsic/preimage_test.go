package extrinsic

import (
	"testing"
)

func TestInit(t *testing.T) {
	var preimages Preimages

	if len(preimages) != 0 {
		t.Fatalf("Expected 0 preimages, got %d", len(preimages))
	}
}

func TestAdd(t *testing.T) {
	var preimages Preimages

	// Add a new Preimage
	preimage1 := Preimage{
		Requester: 16909060,
		Blob:      "0x81095e6122e3bc9d961e00014a7fc833",
	}

	preimages.Add(preimage1)
}

func TestRemoveDuplicatesEmpty(t *testing.T) {
	var preimages Preimages

	preimages.RemoveDuplicates()

	if len(preimages) != 0 {
		t.Fatalf("Expected 0 preimages, got %d", len(preimages))
	}
}

func TestRemoveDuplicates(t *testing.T) {
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

	// Add a new Preimage
	preimage2 := Preimage{
		Requester: 16909061,
		Blob:      "0xd257bc7d93a55be3561d720d40a6a342",
	}
	preimages.Add(preimage2)

	if len(preimages) != 2 {
		t.Fatalf("Expected 2 preimages, got %d", len(preimages))
	}
}

func TestSort(t *testing.T) {
	preimages := Preimages{
		{Requester: 16909060, Blob: "0x81095e6122e3bc9d961e00014a7fc833"},
		{Requester: 16909062, Blob: "0x38db056c7c3065fadb630ce6ccbc7385"},
		{Requester: 16909061, Blob: "0xd257bc7d93a55be3561d720d40a6a342"},
	}

	preimages.Sort()

	expectedRequesters := []uint32{16909060, 16909061, 16909062}
	for i, preimage := range preimages {
		if preimage.Requester != expectedRequesters[i] {
			t.Errorf("Expected requester %d, got %d", expectedRequesters[i], preimage.Requester)
		}
	}
}
