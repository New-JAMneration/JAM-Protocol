package extrinsic

import (
	"encoding/json"
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

func TestDeserializePreimagesEmpty(t *testing.T) {
	jsonData := `[]`
	var preimages Preimages

	err := json.Unmarshal([]byte(jsonData), &preimages)
	if err != nil {
		t.Fatalf("Error deserializing JSON: %v", err)
	}

	if len(preimages) != 0 {
		t.Fatalf("Expected 0 preimages, got %d", len(preimages))
	}
}

func TestDeserializePreimage(t *testing.T) {
	jsonData := `{"requester": 16909060, "blob": "0x81095e6122e3bc9d961e00014a7fc833"}`
	var preimage Preimage

	err := json.Unmarshal([]byte(jsonData), &preimage)
	if err != nil {
		t.Fatalf("Error deserializing JSON: %v", err)
	}

	if preimage.Requester != 16909060 {
		t.Errorf("Expected Requester to be 16909060, got %d", preimage.Requester)
	}

	if preimage.Blob != "0x81095e6122e3bc9d961e00014a7fc833" {
		t.Errorf("Expected Blob to be 0x81095e6122e3bc9d961e00014a7fc833, got %s", preimage.Blob)
	}
}

func TestDeserializePreimages(t *testing.T) {
	jsonData := `[
		{"requester": 16909060, "blob": "0x81095e6122e3bc9d961e00014a7fc833"},
		{"requester": 16909061, "blob": "0xd257bc7d93a55be3561d720d40a6a342"},
		{"requester": 16909062, "blob": "0x38db056c7c3065fadb630ce6ccbc7385"}
	]`
	var preimages Preimages

	err := json.Unmarshal([]byte(jsonData), &preimages)
	if err != nil {
		t.Fatalf("Error deserializing JSON: %v", err)
	}

	expectedRequesters := []uint32{16909060, 16909061, 16909062}
	for i, preimage := range preimages {
		if preimage.Requester != expectedRequesters[i] {
			t.Errorf("Expected Requester to be %d, got %d", expectedRequesters[i], preimage.Requester)
		}
	}
}

func TestSerializePreimagesEmpty(t *testing.T) {
	preimages := Preimages{}

	jsonData, err := json.Marshal(preimages)
	if err != nil {
		t.Fatalf("Error serializing JSON: %v", err)
	}

	expectedJSON := `[]`
	if string(jsonData) != expectedJSON {
		t.Errorf("Expected JSON to be %s, got %s", expectedJSON, string(jsonData))
	}
}

func TestSerializePreimage(t *testing.T) {
	preimage := Preimage{
		Requester: 16909060,
		Blob:      "0x81095e6122e3bc9d961e00014a7fc833",
	}

	jsonData, err := json.Marshal(preimage)
	if err != nil {
		t.Fatalf("Error serializing JSON: %v", err)
	}

	expectedJSON := `{"requester":16909060,"blob":"0x81095e6122e3bc9d961e00014a7fc833"}`
	if string(jsonData) != expectedJSON {
		t.Errorf("Expected JSON to be %s, got %s", expectedJSON, string(jsonData))
	}
}

func TestSerializePreimages(t *testing.T) {
	preimages := Preimages{
		{Requester: 16909060, Blob: "0x81095e6122e3bc9d961e00014a7fc833"},
		{Requester: 16909061, Blob: "0xd257bc7d93a55be3561d720d40a6a342"},
		{Requester: 16909062, Blob: "0x38db056c7c3065fadb630ce6ccbc7385"},
	}

	jsonData, err := json.Marshal(preimages)
	if err != nil {
		t.Fatalf("Error serializing JSON: %v", err)
	}

	expectedJSON := `[{"requester":16909060,"blob":"0x81095e6122e3bc9d961e00014a7fc833"},{"requester":16909061,"blob":"0xd257bc7d93a55be3561d720d40a6a342"},{"requester":16909062,"blob":"0x38db056c7c3065fadb630ce6ccbc7385"}]`
	if string(jsonData) != expectedJSON {
		t.Errorf("Expected JSON to be %s, got %s", expectedJSON, string(jsonData))
	}
}
