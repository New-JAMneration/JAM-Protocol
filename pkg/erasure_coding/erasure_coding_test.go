package erasurecoding

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/rand/v2"
	"os"
	"testing"
)

type TestInput struct {
	Data    string `json:"data"`
	Segment struct {
		Segments []struct {
			SegmentEC []string `json:"segment_ec"`
		} `json:"segments"`
	} `json:"segment"`
}

func TestErasureCoding(t *testing.T) {
	ec, err := NewErasureCoding(342, 1023, 6)
	if err != nil {
		t.Fatalf("failed to create erasure coding: %v", err)
	}

	// mock data
	data := make([]byte, 4104) // or multiple of 684
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}

	// Encode
	shards, err := ec.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(shards) != 1023 {
		t.Fatalf("unexpected shards count: got %d, want 1023", len(shards))
	}

	// take first 342 shards
	selectedShards := make([][]byte, 1023)
	for i := 0; i < 342; i++ {
		selectedShards[i] = shards[i]
	}

	// Decode
	recovered, err := ec.Decode(selectedShards)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !bytes.Equal(data, recovered[:len(data)]) {
		t.Fatal("recovered data does not match original")
	}

	t.Logf("Reconstructed data successfully")
}

func TestReconstructWithTestFile(t *testing.T) {
	// Open json test data file
	file, err := os.Open("test_json_input.json")
	if err != nil {
		t.Fatalf("Cannot open JSON file: %v", err)
	}
	defer file.Close()

	var input TestInput
	if err := json.NewDecoder(file).Decode(&input); err != nil {
		t.Fatalf("Cannot parse JSON file: %v", err)
	}

	// Decode original data
	data, err := hex.DecodeString(input.Data)
	if err != nil {
		t.Fatalf("Cannot decode original data: %v", err)
	}

	// Initialize erasure coding
	ec, err := NewErasureCoding(342, 1023, len(data)/684)
	if err != nil {
		t.Fatalf("Initialization failed: %v", err)
	}

	// Encode data
	shards, err := ec.Encode(data)
	if err != nil {
		t.Fatalf("Encoding fail: %v", err)
	}

	// Simulate data loss
	totalShards := len(shards)
	lostShards := rand.Perm(totalShards)[:681]
	for _, i := range lostShards {
		shards[i] = nil
	}

	// Reconstruct data
	reconstructedData, err := ec.Decode(shards)
	if err != nil {
		t.Fatalf("Reconstruction failed: %v", err)
	}

	// Verify the reconstructed data
	if !bytes.Equal(data, reconstructedData) {
		t.Fatalf("Reconstructed data does not match original data:\nGot: %x\nWant: %x", reconstructedData, data)
	}
	t.Logf("Reconstructed data successfully")
}
