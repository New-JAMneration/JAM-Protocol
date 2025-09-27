package erasurecoding

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testVector struct {
	Data   string   `json:"data"`
	Shards []string `json:"shards"`
}

func hexToBytes(s string) []byte {
	s = trim0x(s)
	b, _ := hex.DecodeString(s)
	return b
}

func trim0x(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s[2:]
	}
	return s
}

func TestEncodeAll(t *testing.T) {
	type modeConfig struct {
		dir         string
		dataShards  int
		totalShards int
	}

	modes := []modeConfig{
		{
			dir:         "tiny_test",
			dataShards:  2,
			totalShards: 6,
		},
		{
			dir:         "full_test",
			dataShards:  342,
			totalShards: 1023,
		},
	}

	for _, mode := range modes {
		files, err := os.ReadDir(mode.dir)
		if err != nil {
			t.Fatalf("read dir %s: %v", mode.dir, err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			testName := fmt.Sprintf("%s/%s", mode.dir, file.Name())

			t.Run(testName, func(t *testing.T) {
				path := filepath.Join(mode.dir, file.Name())
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read file %s: %v", path, err)
				}

				var vec testVector
				if err := json.Unmarshal(data, &vec); err != nil {
					t.Fatalf("unmarshal %s: %v", path, err)
				}

				inputData := hexToBytes(vec.Data)
				parityShards := mode.totalShards - mode.dataShards
				shards, err := EncodeDataShards(inputData, mode.dataShards, parityShards)
				if err != nil {
					t.Fatalf("encode %s: %v", path, err)
				}

				if len(shards) != mode.totalShards {
					t.Fatalf("expected %d shards, got %d", mode.totalShards, len(shards))
				}

				t.Logf("Testing %s: data=%d bytes", path, len(inputData))
				for i, shard := range shards {
					got := hex.EncodeToString(shard)
					want := strings.TrimPrefix(vec.Shards[i], "0x")
					if got != strings.ToLower(want) {
						t.Errorf("shard[%d] mismatch:\n  got:  %s\n  want: %s", i, got, want)
					}
				}
			})
		}
	}
}

func TestDecodeAll(t *testing.T) {
	type modeConfig struct {
		dir         string
		dataShards  int
		totalShards int
	}

	modes := []modeConfig{
		{
			dir:         "tiny_test",
			dataShards:  2,
			totalShards: 6,
		},
		{
			dir:         "full_test",
			dataShards:  342,
			totalShards: 1023,
		},
	}

	for _, mode := range modes {
		files, err := os.ReadDir(mode.dir)
		if err != nil {
			t.Fatalf("read dir %s: %v", mode.dir, err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			testName := fmt.Sprintf("%s/%s", mode.dir, file.Name())

			t.Run(testName, func(t *testing.T) {
				path := filepath.Join(mode.dir, file.Name())
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read file %s: %v", path, err)
				}

				var vec testVector
				if err := json.Unmarshal(data, &vec); err != nil {
					t.Fatalf("unmarshal %s: %v", path, err)
				}

				inputData := hexToBytes(vec.Data)

				parityShards := mode.totalShards - mode.dataShards

				if len(vec.Shards) != mode.totalShards {
					t.Fatalf("expected %d shards, got %d", mode.totalShards, len(vec.Shards))
				}

				var allShards [][]byte
				for _, h := range vec.Shards {
					h = strings.TrimPrefix(h, "0x")
					b, err := hex.DecodeString(h)
					if err != nil {
						t.Fatalf("invalid shard hex: %v", err)
					}
					allShards = append(allShards, b)
				}

				shardSize := len(allShards[0])
				chunkCount := shardSize / 2
				if chunkCount == 0 {
					chunkCount = 1
				}
				t.Logf("Loaded %d shards, shardSize=%d bytes, chunkCount=%d", len(allShards), shardSize, chunkCount)

				// Select random index
				selectedIdx := rand.Perm(mode.totalShards)[:mode.dataShards]
				t.Logf("Selected indices: %v", selectedIdx)

				// Flatten shards
				flatten := make([]byte, 0, len(selectedIdx)*shardSize)
				for _, idx := range selectedIdx {
					flatten = append(flatten, allShards[idx]...)
				}

				// Decode
				recovered, err := DecodeShards(flatten, selectedIdx, mode.dataShards, parityShards, shardSize)
				if err != nil {
					t.Fatalf("Decode failed: %v", err)
				}

				if len(recovered) < len(inputData) {
					t.Fatalf("decoded data too short: got %d, want %d", len(recovered), len(inputData))
				}

				if !bytes.Equal(recovered[:len(inputData)], inputData) {
					t.Errorf("decoded data mismatch:\n got:  %x\n want: %x", recovered[:len(inputData)], inputData)
				}
			})
		}
	}
}

func TestDecodeOnly(t *testing.T) {
	filename := "tiny_test/ec-10000.json"

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	var vec testVector
	if err := json.Unmarshal(data, &vec); err != nil {
		t.Fatal(err)
	}

	inputData := hexToBytes(vec.Data)
	shardHexes := vec.Shards

	dataShards := 2
	parityShards := 4
	totalShards := dataShards + parityShards

	if len(shardHexes) != totalShards {
		t.Fatalf("expected %d shards, got %d", totalShards, len(shardHexes))
	}

	var allShards [][]byte
	for _, h := range shardHexes {
		h = strings.TrimPrefix(h, "0x")
		b, err := hex.DecodeString(h)
		if err != nil {
			t.Fatalf("invalid shard hex: %v", err)
		}
		allShards = append(allShards, b)
	}

	shardSize := len(allShards[0])
	chunkCount := shardSize / 2
	if chunkCount == 0 {
		chunkCount = 1
	}
	t.Logf("Loaded %d shards, shardSize=%d bytes, chunkCount=%d", len(allShards), shardSize, chunkCount)

	// Random select shards
	selectedIdx := rand.Perm(totalShards)[:dataShards]
	t.Logf("Selected indices: %v", selectedIdx)

	// Flatten selected shards
	flatten := make([]byte, 0, dataShards*shardSize)
	for _, idx := range selectedIdx {
		flatten = append(flatten, allShards[idx]...)
	}

	// Decode
	recovered, err := DecodeShards(flatten, selectedIdx, dataShards, parityShards, shardSize)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(recovered) < len(inputData) {
		t.Fatalf("decoded data too short: got %d, want %d", len(recovered), len(inputData))
	}

	if !bytes.Equal(recovered[:len(inputData)], inputData) {
		t.Errorf("decoded data mismatch:\n got:  %x\n want: %x", recovered[:len(inputData)], inputData)
	}
}

func TestEncodeThenDecode(t *testing.T) {
	inputData := []byte("Hello, this is some test data to be erasure-coded and recovered!")

	// tiny
	// dataShards := 2
	// parityShards := 4
	// full
	dataShards := 342
	parityShards := 681
	totalShards := dataShards + parityShards

	// Encode
	encodedShards, err := EncodeDataShards(inputData, dataShards, parityShards)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(encodedShards) != totalShards {
		t.Fatalf("expected %d shards, got %d", totalShards, len(encodedShards))
	}

	shardSize := len(encodedShards[0])
	t.Logf("Encoded %d shards, shardSize=%d bytes", totalShards, shardSize)

	// Randomly pick dataShards shards to simulate loss
	selectedIdx := rand.Perm(totalShards)[:dataShards]
	t.Logf("Selected indices for recovery: %v", selectedIdx)

	// Flatten selected shards
	flatten := make([]byte, 0, dataShards*shardSize)
	for _, idx := range selectedIdx {
		flatten = append(flatten, encodedShards[idx]...)
	}

	// Decode
	recovered, err := DecodeShards(flatten, selectedIdx, dataShards, parityShards, shardSize)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Validate recovered data matches original input
	if len(recovered) < len(inputData) {
		t.Fatalf("decoded data too short: got %d, want %d", len(recovered), len(inputData))
	}
	if !bytes.Equal(recovered[:len(inputData)], inputData) {
		t.Errorf("decoded data mismatch:\n got:  %x\n want: %x", recovered[:len(inputData)], inputData)
	}
}
