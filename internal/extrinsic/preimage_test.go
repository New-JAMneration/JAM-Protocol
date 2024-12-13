package extrinsic

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func TestInit(t *testing.T) {
	var preimages jam_types.PreimagesExtrinsic

	if Len(preimages) != 0 {
		t.Fatalf("Expected 0 preimages, got %d", Len(preimages))
	}
}

func TestAdd(t *testing.T) {
	var preimages jam_types.PreimagesExtrinsic

	// Add a new Preimage
	preimage1 := jam_types.Preimage{
		Requester: 16909060,
		Blob:      []byte("0x81095e6122e3bc9d961e00014a7fc833"),
	}

	preimages = Add(preimages, preimage1)

	if Len(preimages) != 1 {
		t.Fatalf("Expected 1 preimage, got %d", Len(preimages))
	}
}

func TestRemoveDuplicatesEmpty(t *testing.T) {
	var preimages jam_types.PreimagesExtrinsic

	preimages = RemoveDuplicates(preimages)

	if Len(preimages) != 0 {
		t.Fatalf("Expected 0 preimages, got %d", Len(preimages))
	}
}

func TestRemoveDuplicates(t *testing.T) {
	var preimages jam_types.PreimagesExtrinsic

	// Add a new Preimage
	preimage1 := jam_types.Preimage{
		Requester: 16909060,
		Blob:      []byte("0x81095e6122e3bc9d961e00014a7fc833"),
	}
	preimages = Add(preimages, preimage1)

	// Add a duplicate Preimage
	duplicatePreimage := jam_types.Preimage{
		Requester: 16909060,
		Blob:      []byte("0x81095e6122e3bc9d961e00014a7fc833"),
	}
	preimages = Add(preimages, duplicatePreimage)

	// Add a new Preimage
	preimage2 := jam_types.Preimage{
		Requester: 16909061,
		Blob:      []byte("0xd257bc7d93a55be3561d720d40a6a342"),
	}
	preimages = Add(preimages, preimage2)

	if Len(preimages) != 2 {
		t.Fatalf("Expected 2 preimages, got %d", Len(preimages))
	}
}

func TestSort(t *testing.T) {
	preimages := jam_types.PreimagesExtrinsic{
		{Requester: 16909060, Blob: []byte("0x81095e6122e3bc9d961e00014a7fc833")},
		{Requester: 16909062, Blob: []byte("0x38db056c7c3065fadb630ce6ccbc7385")},
		{Requester: 16909061, Blob: []byte("0xd257bc7d93a55be3561d720d40a6a342")},
	}

	Sort(preimages)

	expectedRequesters := []jam_types.ServiceId{16909060, 16909061, 16909062}
	for i, preimage := range preimages {
		if preimage.Requester != expectedRequesters[i] {
			t.Errorf("Expected requester %d, got %d", expectedRequesters[i], preimage.Requester)
		}
	}
}
