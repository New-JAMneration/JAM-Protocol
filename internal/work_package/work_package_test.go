package work_package

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestPadToMultiple(t *testing.T) {
	tests := []struct {
		input []byte
		n     int
	}{
		{[]byte{}, 4},
		{[]byte{1}, 4},
		{[]byte{1, 2, 3, 4}, 4},
		{[]byte{1, 2, 3, 4, 5}, 4},
		{[]byte{1, 2, 3, 4, 5, 6}, types.SegmentSize},
	}

	for _, tt := range tests {
		result := PadToMultiple(tt.input, tt.n)
		if len(result)%tt.n != 0 {
			t.Errorf("Expected length to be multiple of %d, got %d", tt.n, len(result))
		}
	}
}

func TestTranspose(t *testing.T) {
	input := [][]int{
		{1, 2, 3},
		{4, 5, 6},
	}
	expected := [][]int{
		{1, 4},
		{2, 5},
		{3, 6},
	}
	result := Transpose(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestTransposeTripleByte(t *testing.T) {
	input := [][][]byte{
		{
			[]byte("a1"),
			[]byte("a2"),
		},
		{
			[]byte("b1"),
			[]byte("b2"),
		},
	}

	expected := [][][]byte{
		{
			[]byte("a1"),
			[]byte("b1"),
		},
		{
			[]byte("a2"),
			[]byte("b2"),
		},
	}

	result := Transpose(input)

	if len(result) != len(expected) {
		t.Fatalf("Expected %d rows, got %d", len(expected), len(result))
	}
	for i := range result {
		for j := range result[i] {
			if !bytes.Equal(result[i][j], expected[i][j]) {
				t.Errorf("Mismatch at [%d][%d]: expected %s, got %s", i, j, expected[i][j], result[i][j])
			}
		}
	}
}

func TestExtractExtrinsics(t *testing.T) {
	specs := []types.ExtrinsicSpec{
		{
			Hash: hash.Blake2bHash([]byte("abcde")),
			Len:  5,
		},
		{
			Hash: hash.Blake2bHash([]byte("12345")),
			Len:  5,
		},
	}

	hash1 := hash.Blake2bHash([]byte("abcde"))
	hash2 := hash.Blake2bHash([]byte("12345"))

	expected := PVM.ExtrinsicDataMap{
		hash1: []byte("abcde"),
		hash2: []byte("12345"),
	}

	data := append([]byte("abcde"), []byte("12345")...)

	result, err := ExtractExtrinsics(data, specs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(expected) {
		t.Errorf("expected %d extrinsics, got %d", len(expected), len(result))
	}

	for k, v := range expected {
		got, ok := result[k]
		if !ok {
			t.Errorf("missing key %x in result", k)
			continue
		}
		if !bytes.Equal(v, got) {
			t.Errorf("mismatch for key %x: expected %x, got %x", k, v, got)
		}
	}
}

func TestPagedProofs(t *testing.T) {
	segment := make([]byte, 4104)
	copy(segment, []byte("segment-data"))
	exportSegments := []types.ExportSegment{
		types.ExportSegment(segment),
		types.ExportSegment(segment),
	}

	proofs, err := PagedProofs(exportSegments)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(proofs) == 0 {
		t.Errorf("expected at least one paged proof, got 0")
	}

	for i, proof := range proofs {
		if len(proof)%types.SegmentSize != 0 {
			t.Errorf("paged proof at index %d is not padded to segment size", i)
		}
	}
	// TODO: check if the proofs are correct
}

func TestBuildWorkPackageBundle(t *testing.T) {
	extrinsicHash1 := hash.Blake2bHash([]byte("abcde"))
	extrinsicHash2 := hash.Blake2bHash([]byte("12345"))
	wp := &types.WorkPackage{
		Authorization:    types.ByteSequence{0x01, 0x02, 0x03},
		AuthCodeHost:     types.ServiceId(1),
		AuthCodeHash:     types.OpaqueHash{0x04, 0x05, 0x06},
		AuthorizerConfig: types.ByteSequence{0x07, 0x08, 0x09},
		Context: types.RefineContext{
			Anchor:           types.HeaderHash{0x0A, 0x0B, 0x0C},
			StateRoot:        types.StateRoot{0x0D, 0x0E, 0x0F},
			BeefyRoot:        types.BeefyRoot{0x10, 0x11, 0x12},
			LookupAnchor:     types.HeaderHash{0x13, 0x14, 0x15},
			LookupAnchorSlot: types.TimeSlot(12345),
			Prerequisites:    nil,
		},
		Items: []types.WorkItem{
			{
				Service:            types.ServiceId(1),
				CodeHash:           types.OpaqueHash{0x16, 0x17, 0x18},
				Payload:            types.ByteSequence{0x19, 0x1A, 0x1B},
				RefineGasLimit:     types.Gas(1000),
				AccumulateGasLimit: types.Gas(2000),
				ExportCount:        types.U16(3),
				ImportSegments: []types.ImportSpec{
					{TreeRoot: types.OpaqueHash{0x1C, 0x1D, 0x1E}, Index: types.U16(1)},
				},
				Extrinsic: []types.ExtrinsicSpec{
					{Hash: extrinsicHash1, Len: 5},
					{Hash: extrinsicHash2, Len: 5},
				},
			},
		},
	}

	extrinsicMap := PVM.ExtrinsicDataMap{
		extrinsicHash1: []byte("abcde"),
		extrinsicHash2: []byte("12345"),
	}

	segment := make([]byte, 4104)
	copy(segment, []byte("segment1"))
	importSegments := types.ExportSegmentMatrix{
		{
			types.ExportSegment(segment),
		},
	}
	importProofs := types.OpaqueHashMatrix{
		{
			[32]byte{1},
		},
	}

	bundle, err := buildWorkPackageBundle(wp, extrinsicMap, importSegments, importProofs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bundle) == 0 {
		t.Errorf("expected non-empty bundle")
	}

	redisBackend, err := blockchain.GetRedisBackend()
	if err != nil {
		t.Fatalf("Failed to get redis backend: %v", err)
	}
	hashSegmentMap, err := redisBackend.GetHashSegmentMap()
	if err != nil {
		t.Fatalf("Failed to get hash segment map: %v", err)
	}
	var decoded types.WorkPackageBundle
	decoder := types.NewDecoder()
	decoder.SetHashSegmentMap(hashSegmentMap)
	err = decoder.Decode(bundle, &decoded)
	if err != nil {
		t.Fatalf("failed to decode bundle: %v", err)
	}
	if !reflect.DeepEqual(decoded.Package, *wp) {
		t.Errorf("decoded work package does not match original")
	}
	for i, extrinsic := range decoded.Extrinsics {
		if !bytes.Equal(extrinsic, extrinsicMap[wp.Items[0].Extrinsic[i].Hash]) {
			t.Errorf("decoded extrinsic data at index %d does not match original", i)
		}
	}
	if !reflect.DeepEqual(decoded.ImportSegments, importSegments) {
		t.Errorf("decoded import segments do not match original")
	}
	if !reflect.DeepEqual(decoded.ImportProofs, importProofs) {
		t.Errorf("decoded import proofs do not match original")
	}
}
