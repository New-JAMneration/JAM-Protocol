package extrinsic

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestNewPreimageController(t *testing.T) {
	preimageController := NewPreimageController()

	if len(preimageController.Preimages) != 0 {
		t.Fatalf("Expected 0 preimages, got %d", len(preimageController.Preimages))
	}
}

func TestAddNewPreimage(t *testing.T) {
	preimageController := NewPreimageController()

	// Add a new Preimage
	preimage1 := types.Preimage{
		Requester: 16909060,
		Blob:      []byte("0x81095e6122e3bc9d961e00014a7fc833"),
	}

	preimageController.Add(preimage1)

	if len(preimageController.Preimages) != 1 {
		t.Fatalf("Expected 1 preimage, got %d", len(preimageController.Preimages))
	}
}

func TestSetPreimagesEmpty(t *testing.T) {
	preimageController := NewPreimageController()

	// Initialize the Preimages slice without using the Add function
	preimages := []types.Preimage{}
	preimageController.Set(preimages)

	if len(preimageController.Preimages) != 0 {
		t.Fatalf("Expected 0 preimages, got %d", len(preimageController.Preimages))
	}
}

func TestSetPreimages(t *testing.T) {
	preimageController := NewPreimageController()

	// Initialize the Preimages slice without using the Add function
	preimages := []types.Preimage{
		{Requester: 16909060, Blob: []byte("0x81095e6122e3bc9d961e00014a7fc833")},
		{Requester: 16909061, Blob: []byte("0xd257bc7d93a55be3561d720d40a6a342")},
		{Requester: 16909062, Blob: []byte("0x38db056c7c3065fadb630ce6ccbc7385")},
	}
	preimageController.Set(preimages)

	if len(preimageController.Preimages) != 3 {
		t.Fatalf("Expected 3 preimages, got %d", len(preimageController.Preimages))
	}
}

func TestAddMultiplePreimages(t *testing.T) {
	preimageController := NewPreimageController()

	// Add a new Preimage
	preimage1 := types.Preimage{
		Requester: 16909060,
		Blob:      []byte("0x81095e6122e3bc9d961e00014a7fc833"),
	}
	preimageController.Add(preimage1)

	// Add a new Preimage
	preimage2 := types.Preimage{
		Requester: 16909061,
		Blob:      []byte("0xd257bc7d93a55be3561d720d40a6a342"),
	}
	preimageController.Add(preimage2)

	if len(preimageController.Preimages) != 2 {
		t.Fatalf("Expected 2 preimages, got %d", len(preimageController.Preimages))
	}
}

func TestAddDuplicatePreimages(t *testing.T) {
	preimageController := NewPreimageController()

	// Add a new Preimage
	preimage1 := types.Preimage{
		Requester: 16909060,
		Blob:      []byte("0x81095e6122e3bc9d961e00014a7fc833"),
	}
	preimageController.Add(preimage1)

	// Add a duplicate Preimage
	duplicatePreimage := types.Preimage{
		Requester: 16909060,
		Blob:      []byte("0x81095e6122e3bc9d961e00014a7fc833"),
	}
	preimageController.Add(duplicatePreimage)

	if len(preimageController.Preimages) != 1 {
		t.Fatalf("Expected 1 preimage, got %d", len(preimageController.Preimages))
	}
}

func TestAddUnsortedPreimages(t *testing.T) {
	preimageController := NewPreimageController()

	// Add a new Preimage
	preimage1 := types.Preimage{
		Requester: 16909060,
		Blob:      []byte("0x81095e6122e3bc9d961e00014a7fc833"),
	}
	preimageController.Add(preimage1)

	// Add a new Preimage
	preimage2 := types.Preimage{
		Requester: 16909062,
		Blob:      []byte("0x38db056c7c3065fadb630ce6ccbc7385"),
	}
	preimageController.Add(preimage2)

	// Add a new Preimage
	preimage3 := types.Preimage{
		Requester: 16909061,
		Blob:      []byte("0xd257bc7d93a55be3561d720d40a6a342"),
	}
	preimageController.Add(preimage3)

	if len(preimageController.Preimages) != 3 {
		t.Fatalf("Expected 3 preimages, got %d", len(preimageController.Preimages))
	}

	expectedRequesters := []types.ServiceId{16909060, 16909061, 16909062}
	for i, preimage := range preimageController.Preimages {
		if preimage.Requester != expectedRequesters[i] {
			t.Errorf("Expected requester %d, got %d", expectedRequesters[i], preimage.Requester)
		}
	}

	expectedBlobs := [][]byte{
		[]byte("0x81095e6122e3bc9d961e00014a7fc833"),
		[]byte("0xd257bc7d93a55be3561d720d40a6a342"),
		[]byte("0x38db056c7c3065fadb630ce6ccbc7385"),
	}
	for i, preimage := range preimageController.Preimages {
		if !bytes.Equal(preimage.Blob, expectedBlobs[i]) {
			t.Errorf("Expected blob %s, got %s", expectedBlobs[i], preimage.Blob)
		}
	}
}

func TestSort(t *testing.T) {
	preimageController := NewPreimageController()

	// Initialize the Preimages slice without using the Add function
	preimages := []types.Preimage{
		{Requester: 16909060, Blob: []byte("0x81095e6122e3bc9d961e00014a7fc833")},
		{Requester: 16909062, Blob: []byte("0x38db056c7c3065fadb630ce6ccbc7385")},
		{Requester: 16909061, Blob: []byte("0xd257bc7d93a55be3561d720d40a6a342")},
	}
	preimageController.Set(preimages)

	preimageController.Sort()

	expectedRequesters := []types.ServiceId{16909060, 16909061, 16909062}
	for i, preimage := range preimageController.Preimages {
		if preimage.Requester != expectedRequesters[i] {
			t.Errorf("Expected requester %d, got %d", expectedRequesters[i], preimage.Requester)
		}
	}

	expectedBlobs := [][]byte{
		[]byte("0x81095e6122e3bc9d961e00014a7fc833"),
		[]byte("0xd257bc7d93a55be3561d720d40a6a342"),
		[]byte("0x38db056c7c3065fadb630ce6ccbc7385"),
	}
	for i, preimage := range preimageController.Preimages {
		if !bytes.Equal(preimage.Blob, expectedBlobs[i]) {
			t.Errorf("Expected blob %s, got %s", expectedBlobs[i], preimage.Blob)
		}
	}
}

func TestRemoveDuplicatesEmpty(t *testing.T) {
	preimageController := NewPreimageController()

	preimageController.RemoveDuplicates()

	if len(preimageController.Preimages) != 0 {
		t.Fatalf("Expected 0 preimages, got %d", len(preimageController.Preimages))
	}
}

func TestRemoveDuplicates(t *testing.T) {
	preimageController := NewPreimageController()

	// Initialize the Preimages slice without using the Add function
	preimages := []types.Preimage{
		{Requester: 16909060, Blob: []byte("0x81095e6122e3bc9d961e00014a7fc833")},
		{Requester: 16909060, Blob: []byte("0x81095e6122e3bc9d961e00014a7fc833")},
		{Requester: 16909061, Blob: []byte("0xd257bc7d93a55be3561d720d40a6a342")},
	}
	preimageController.Set(preimages)

	preimageController.RemoveDuplicates()

	if len(preimageController.Preimages) != 2 {
		t.Fatalf("Expected 2 preimages, got %d", len(preimageController.Preimages))
	}

	expectedRequesters := []types.ServiceId{16909060, 16909061}
	for i, preimage := range preimageController.Preimages {
		if preimage.Requester != expectedRequesters[i] {
			t.Errorf("Expected requester %d, got %d", expectedRequesters[i], preimage.Requester)
		}
	}

	expectedBlobs := [][]byte{
		[]byte("0x81095e6122e3bc9d961e00014a7fc833"),
		[]byte("0xd257bc7d93a55be3561d720d40a6a342"),
	}
	for i, preimage := range preimageController.Preimages {
		if !bytes.Equal(preimage.Blob, expectedBlobs[i]) {
			t.Errorf("Expected blob %s, got %s", expectedBlobs[i], preimage.Blob)
		}
	}
}
