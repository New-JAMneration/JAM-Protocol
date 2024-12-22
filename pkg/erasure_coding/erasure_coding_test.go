package erasurecoding

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/rand"
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

func TestReconstructFixedData(t *testing.T) {
	// Open json test data file
	file, err := os.Open("test_json_input.json")
	if err != nil {
		t.Fatalf("Cannot open JSON file: %v\n", err)
		return
	}
	defer file.Close()

	var input TestInput
	if err := json.NewDecoder(file).Decode(&input); err != nil {
		t.Fatalf("Cannot parse JSON file: %v\n", err)
		return
	}

	// Decode original data
	originalData, err := hex.DecodeString(input.Data)
	if err != nil {
		t.Fatalf("Cannot decode original data: %v\n", err)
		return
	}

	shards := make([]Shard, len(input.Segment.Segments[0].SegmentEC))
	for i, hexStr := range input.Segment.Segments[0].SegmentEC {
		data, err := hex.DecodeString(hexStr)
		if err != nil {
			t.Fatalf("Cannot decode segment %d: %v\n", i, err)
			return
		}
		shards[i] = Shard{Index: i, Data: data}
	}

	// Initialize ErasureCoding
	ec, err := NewErasureCoding()
	if err != nil {
		t.Fatalf("Initialization failed: %v\n", err)
		return
	}

	// Reconstruct data
	reconstructedData, err := ec.Decode(shards)
	if err != nil {
		t.Fatalf("Data Reconstruction failed: %v", err)
	}

	// Check the first 342*2 bytes of the reconstructed data
	if !bytes.Equal(reconstructedData[:342*2], originalData[:342*2]) {
		t.Fatalf("Reconstructed data does not match original data:\nGot: %x\nWant: %x", reconstructedData[:342*2], originalData[:342*2])
	}
	t.Logf("Reconstructed data successfully")
}

func TestReconstructWithMinimumShards(t *testing.T) {
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
	ec, err := NewErasureCoding()
	if err != nil {
		t.Fatalf("Initialization failed: %v", err)
	}

	// Encode data
	shards, err := ec.Encode(data)
	if err != nil {
		t.Fatalf("編碼失敗: %v", err)
	}

	// Simulate data loss
	totalShards := len(shards)
	lostShards := rand.Perm(totalShards)[:681]
	for _, i := range lostShards {
		shards[i].Data = nil
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
