package types_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	jamtests_accmuluate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
)

func GetBytesFromFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return data, nil
}

func GetBlockFromJson(filename string) (*types.Block, error) {
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %v", err)
	}

	// Unmarshal the JSON data
	var block types.Block
	err = json.Unmarshal(data, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

// func TestDecodeBlock(t *testing.T) {
// 	blockBinFile := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.bin"

// 	// Get the bytes from the file
// 	data, err := GetBytesFromFile(blockBinFile)
// 	if err != nil {
// 		t.Errorf("Error getting bytes from file: %v", err)
// 	}

// 	// Create a new block (pointer)
// 	block := &types.Block{}

// 	// Decode the block from the binary file
// 	decoder := types.NewDecoder()
// 	err = decoder.Decode(data, block)
// 	if err != nil {
// 		t.Errorf("Error decoding block: %v", err)
// 	}

// 	// Check the decode block is same as expected block
// 	// Read json file and unmarshal it to get the expected block
// 	blockJsonFile := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.json"
// 	expectedBlock, err := GetBlockFromJson(blockJsonFile)
// 	if err != nil {
// 		t.Errorf("Error getting block from json file: %v", err)
// 	}

// 	// Compare the two blocks
// 	if !reflect.DeepEqual(block, expectedBlock) {
// 		t.Errorf("Blocks do not match")
// 	}
// }

func TestDecodeU64(t *testing.T) {
	data := []byte{0, 228, 11, 84, 2, 0, 0, 0}

	decoder := types.NewDecoder()
	u64 := types.U64(0)
	err := decoder.Decode(data, &u64)
	if err != nil {
		t.Errorf("Error decoding U64: %v", err)
	}

	expected := types.U64(10000000000)

	if u64 != expected {
		t.Errorf("Decoded U64 does not match expected")
	}
}

func TestDecodeServiceId(t *testing.T) {
	data := []byte{100, 0, 0, 0, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	decoder := types.NewDecoder()
	serviceId := types.ServiceId(0)
	err := decoder.Decode(data, &serviceId)
	if err != nil {
		t.Errorf("Error decoding ServiceId: %v", err)
	}

	expected := types.ServiceId(100)

	if serviceId != expected {
		t.Errorf("Decoded ServiceId does not match expected")
	}
}

func TestDecodeJamTestVectorsAccumulateFile(t *testing.T) {
	file := "../../pkg/test_data/jam-test-vectors/stf/accumulate/tiny/accumulate_ready_queued_reports-1.bin"

	data, err := GetBytesFromFile(file)
	if err != nil {
		t.Errorf("Error getting bytes from file: %v", err)
	}

	decoder := types.NewDecoder()
	accumulateTestCase := &jamtests_accmuluate.AccumulateTestCase{}
	err = decoder.Decode(data, accumulateTestCase)
	if err != nil {
		t.Errorf("Error decoding AccumulateTestCase: %v", err)
	}

	// You can access the fields of the AccumulateTestCase struct
	// e.g. accumulateTestCase.Input.Slot
}

func TestDecodeMetaCode(t *testing.T) {
	testCases := []struct {
		testMetaCodeEncoded []byte
		expectedMetaCode    types.MetaCode
	}{
		{
			testMetaCodeEncoded: []byte{3, 1, 2, 3, 4, 5, 6},
			expectedMetaCode: types.MetaCode{
				Metadata: types.ByteSequence{0x01, 0x02, 0x03},
				Code:     types.ByteSequence{0x04, 0x05, 0x06},
			},
		},
		{
			testMetaCodeEncoded: []byte{},
			expectedMetaCode:    types.MetaCode{},
		},
	}

	for _, tc := range testCases {
		metaCode := types.MetaCode{}
		decoder := types.NewDecoder()
		err := decoder.Decode(tc.testMetaCodeEncoded, &metaCode)
		if err != nil {
			t.Errorf("Error decoding MetaCode: %v", err)
		}

		if !reflect.DeepEqual(metaCode, tc.expectedMetaCode) {
			t.Errorf("Decoded MetaCode does not match expected")
		}
	}
}

func TestExtrinsicData_EncodeDecode(t *testing.T) {
	original := types.ExtrinsicData([]byte("abcde"))
	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(&original)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	var decoded types.ExtrinsicData
	decoder := types.NewDecoder()
	err = decoder.Decode(encoded, &decoded)
	if err != nil {
		t.Errorf("Error decoding ExtrinsicData: %v", err)
	}

	if !bytes.Equal(original, decoded) {
		t.Errorf("expected %v, got %v", original, decoded)
	}
}

func TestExtrinsicDataList_EncodeDecode(t *testing.T) {
	original := types.ExtrinsicDataList{
		[]byte("abc"),
		[]byte("12345"),
	}
	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(&original)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	var decoded types.ExtrinsicDataList
	decoder := types.NewDecoder()
	err = decoder.Decode(encoded, &decoded)
	if err != nil {
		t.Errorf("Error decoding ExtrinsicDataList: %v", err)
	}

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("decoded ExtrinsicDataList does not match original")
	}
}

func TestExportSegmentMatrix_EncodeDecode(t *testing.T) {
	var seg1, seg2, seg3 types.ExportSegment
	copy(seg1[:], []byte("seg1"))
	copy(seg2[:], []byte("seg2"))
	copy(seg3[:], []byte("seg3"))
	original := types.ExportSegmentMatrix{
		{seg1, seg2},
		{seg3},
	}
	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(&original)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	var decoded types.ExportSegmentMatrix
	decoder := types.NewDecoder()
	err = decoder.Decode(encoded, &decoded)
	if err != nil {
		t.Errorf("Error decoding ExportSegmentMatrix: %v", err)
	}

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("decoded ExportSegmentMatrix does not match original")
	}
}

func TestOpaqueHashMatrix_EncodeDecode(t *testing.T) {
	h1 := types.OpaqueHash{1, 2, 3}
	h2 := types.OpaqueHash{4, 5, 6}
	original := types.OpaqueHashMatrix{
		{h1, h2},
		{h1},
	}
	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(&original)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	var decoded types.OpaqueHashMatrix
	decoder := types.NewDecoder()
	err = decoder.Decode(encoded, &decoded)
	if err != nil {
		t.Errorf("Error decoding OpaqueHashMatrix: %v", err)
	}

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("decoded OpaqueHashMatrix does not match original")
	}
}

func TestWorkPackageBundle_EncodeDecode(t *testing.T) {
	h1 := types.OpaqueHash{1, 2, 3}
	seg := types.ExportSegment{}
	copy(seg[:], []byte("seg"))

	bundle := types.WorkPackageBundle{
		Package: types.WorkPackage{
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
						{Hash: hash.Blake2bHash([]byte("abc")), Len: 3},
						{Hash: hash.Blake2bHash([]byte("def")), Len: 3},
					},
				},
			},
		},
		Extrinsics: types.ExtrinsicDataList{
			[]byte("abc"),
		},
		ImportSegments: types.ExportSegmentMatrix{
			{seg},
		},
		ImportProofs: types.OpaqueHashMatrix{
			{h1},
		},
	}
	redisBackend, _ := blockchain.GetRedisBackend()
	encoder := types.NewEncoder()
	hashSegmentMap, err := redisBackend.GetHashSegmentMap()
	if err != nil {
		t.Fatalf("Failed to get hash segment map: %v", err)
	}
	encoder.SetHashSegmentMap(hashSegmentMap)
	encoded, err := encoder.Encode(&bundle)
	if err != nil {
		t.Fatalf("failed to encode WorkPackageBundle: %v", err)
	}
	var decoded types.WorkPackageBundle
	decoder := types.NewDecoder()
	decoder.SetHashSegmentMap(hashSegmentMap)
	err = decoder.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("failed to decode WorkPackageBundle: %v", err)
	}

	if !reflect.DeepEqual(bundle, decoded) {
		t.Errorf("decoded WorkPackageBundle does not match original")
	}
}

func TestWorkPackage_EncodeDecode(t *testing.T) {
	wp := types.WorkPackage{
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
					{Hash: hash.Blake2bHash([]byte("abc")), Len: 3},
					{Hash: hash.Blake2bHash([]byte("def")), Len: 3},
				},
			},
		},
	}
	redisBackend, _ := blockchain.GetRedisBackend()
	e := types.NewEncoder()
	hashSegmentMap, err := redisBackend.GetHashSegmentMap()
	if err != nil {
		t.Fatalf("Failed to get hash segment map: %v", err)
	}
	e.SetHashSegmentMap(hashSegmentMap)
	encoded, err := e.Encode(&wp)
	if err != nil {
		t.Fatalf("failed to encode WorkPackage: %v", err)
	}

	var decoded types.WorkPackage
	d := types.NewDecoder()
	d.SetHashSegmentMap(hashSegmentMap)
	err = d.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("failed to decode WorkPackage: %v", err)
	}
	if !reflect.DeepEqual(wp, decoded) {
		t.Errorf("decoded WorkPackage does not match original")
	}
}
