package types

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestEncodeHeader(t *testing.T) {
	header := Header{
		Slot: 99,
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&header)
	if err != nil {
		t.Errorf("Error encoding Header: %v", err)
	}

	decoder := NewDecoder()
	expectedHeader := Header{}
	err = decoder.Decode(encoded, &expectedHeader)
	if err != nil {
		t.Errorf("Error decoding Header: %v", err)
	}

	// compare the two headers
	if !reflect.DeepEqual(header, expectedHeader) {
		t.Errorf("Headers do not match")
	}
}

func TestEncodeBlock(t *testing.T) {
	block := Block{
		Header: Header{
			Slot: 99,
		},
		Extrinsic: Extrinsic{},
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&block)
	if err != nil {
		t.Errorf("Error encoding Block: %v", err)
	}

	decoder := NewDecoder()
	expectedBlock := Block{}
	err = decoder.Decode(encoded, &expectedBlock)
	if err != nil {
		t.Errorf("Error decoding Block: %v", err)
	}

	// compare the two blocks
	if !reflect.DeepEqual(block, expectedBlock) {
		t.Errorf("Blocks do not match")
	}
}

func TestEncodeU64(t *testing.T) {
	u64 := U64(10000000000)

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&u64)
	if err != nil {
		t.Errorf("Error encoding U64: %v", err)
	}

	expected := []byte{0, 228, 11, 84, 2, 0, 0, 0}

	if !reflect.DeepEqual(encoded, expected) {
		t.Errorf("Encoded U64 does not match expected")
	}
}

func TestEncodeTimeSlot(t *testing.T) {
	timeSlot := TimeSlot(970)

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&timeSlot)
	if err != nil {
		t.Errorf("Error encoding TimeSlot: %v", err)
	}

	expected := []byte{202, 3, 0, 0}

	if !reflect.DeepEqual(encoded, expected) {
		t.Errorf("Encoded TimeSlot does not match expected")
	}
}

func TestEncodeHash(t *testing.T) {
	hexString := "0xbd87fb6de829abf2bb25a15b82618432c94e82848d9dd204f5d775d4b880ae0d"
	bytes := hexToBytes(hexString)
	hash := OpaqueHash(bytes)

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&hash)
	if err != nil {
		t.Errorf("Error encoding Hash: %v", err)
	}

	expected := []byte{189, 135, 251, 109, 232, 41, 171, 242, 187, 37, 161, 91, 130, 97, 132, 50, 201, 78, 130, 132, 141, 157, 210, 4, 245, 215, 117, 212, 184, 128, 174, 13}

	if !reflect.DeepEqual(encoded, expected) {
		t.Errorf("Encoded Hash does not match expected")
	}
}

func TestEncodeCustomContent(t *testing.T) {
	// graypaper (B.10) $\mathcal{E}(s, \eta^{'}_{0}, \mathbf{H_t})$
	// s: service index
	// η': entropy
	// H_t: time slot index

	serviceId := ServiceId(1)
	entropy := Entropy{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	timeSlot := TimeSlot(100)

	output := []byte{}
	encoder := NewEncoder()

	// Encode serviceId
	encoded, err := encoder.Encode(&serviceId)
	if err != nil {
		t.Errorf("Error encoding ServiceId: %v", err)
	}
	output = append(output, encoded...)

	// Encode entropy
	encoded, err = encoder.Encode(&entropy)
	if err != nil {
		t.Errorf("Error encoding Entropy: %v", err)
	}
	output = append(output, encoded...)

	// Encode timeSlot
	encoded, err = encoder.Encode(&timeSlot)
	if err != nil {
		t.Errorf("Error encoding TimeSlot: %v", err)
	}
	output = append(output, encoded...)

	expected := []byte{
		1, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 100, 0, 0, 0,
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Encoded CustomContent does not match expected")
	}
}

func TestEncodeUintWithLength(t *testing.T) {
	encoder := NewEncoder()
	encoded, err := encoder.EncodeUintWithLength(100, 3)
	if err != nil {
		t.Errorf("Error encoding UintWithLength: %v", err)
	}

	expected := []byte{100, 0, 0}

	if !reflect.DeepEqual(encoded, expected) {
		t.Errorf("Encoded UintWithLength does not match expected")
	}
}

func TestEncodeMetaCode(t *testing.T) {
	testMetaCode := MetaCode{
		Metadata: ByteSequence{0x01, 0x02, 0x03},
		Code:     ByteSequence{0x04, 0x05, 0x06},
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&testMetaCode)
	if err != nil {
		t.Errorf("Error encoding MetaCode: %v", err)
	}

	expected := []byte{3, 1, 2, 3, 4, 5, 6}

	if !reflect.DeepEqual(encoded, expected) {
		t.Errorf("Encoded MetaCode does not match expected")
	}
}

func TestEncodeWorkExecResult(t *testing.T) {
	// test "ok" type of WorkExecResult
	okResult := WorkExecResult{
		"ok": []byte{1, 2, 3, 4, 5},
	}

	// create encoder
	encoder := NewEncoder()
	encoded, err := encoder.Encode(&okResult)
	if err != nil {
		t.Errorf("Error encoding WorkExecResult: %v", err)
	}

	// print encoded result for debugging
	t.Logf("Encoded WorkExecResult (ok): %v", encoded)
	t.Logf("Hex: %x", encoded)

	// test if first byte is 0 (corresponding to "ok" type)
	if len(encoded) < 1 || encoded[0] != 0 {
		t.Errorf("Expected first byte to be 0 for 'ok' type, got: %v", encoded[0])
	}

	// decode check
	decoder := NewDecoder()
	decodedResult := WorkExecResult{}
	err = decoder.Decode(encoded, &decodedResult)
	if err != nil {
		t.Errorf("Error decoding WorkExecResult: %v", err)
	}

	// check decoded result
	if !reflect.DeepEqual(okResult, decodedResult) {
		t.Errorf("Decoded result doesn't match original\nExpected: %v\nGot: %v", okResult, decodedResult)
	}
}

func TestEncodeSliceHashStruct(t *testing.T) {
	testSliceHash := SliceHash{
		A: []OpaqueHash{
			{0x01, 0x02, 0x03},
			{0x04, 0x05, 0x06},
			{0x07, 0x08, 0x09},
		},
		B: []OpaqueHash{
			{0x0A, 0x0B, 0x0C},
			{0x0D, 0x0E, 0x0F},
			{0x10, 0x11, 0x12},
		},
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&testSliceHash)
	if err != nil {
		t.Errorf("Error encoding SliceHash: %v", err)
	}

	fmt.Println(encoded)
}

func TestExtrinsicData_Encode(t *testing.T) {
	data := ExtrinsicData([]byte("abcde"))

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&data)
	if err != nil {
		t.Errorf("Error encoding ExtrinsicData: %v", err)
	}
	expected := []byte{5, 97, 98, 99, 100, 101}
	if !reflect.DeepEqual(encoded, expected) {
		t.Errorf("Encoded ExtrinsicData does not match expected")
	}
}

func TestExtrinsicDataList_Encode(t *testing.T) {
	list := ExtrinsicDataList{
		[]byte("abc"),
		[]byte("xyz"),
	}
	encoder := NewEncoder()
	encoded, err := encoder.Encode(&list)
	if err != nil {
		t.Errorf("Error encoding ExtrinsicDataList: %v", err)
	}
	expected := []byte{2, 3, 97, 98, 99, 3, 120, 121, 122}
	if !reflect.DeepEqual(encoded, expected) {
		t.Errorf("Encoded ExtrinsicDataList does not match expected")
	}
}

func TestExportSegment_Encode(t *testing.T) {
	var segment ExportSegment
	copy(segment[:], []byte("abcde"))

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&segment)
	if err != nil {
		t.Fatalf("Error encoding ExportSegment: %v", err)
	}

	if len(encoded) != 4104 {
		t.Fatalf("expected encoded length 4104, got %d", len(encoded))
	}

	if !bytes.Equal(encoded[:5], []byte("abcde")) {
		t.Errorf("expected first 5 bytes to be 'abcde', got %v", encoded[:5])
	}

	for i := 5; i < len(encoded); i++ {
		if encoded[i] != 0 {
			t.Errorf("expected zero padding at byte %d, got %x", i, encoded[i])
			break
		}
	}
}

func TestExportSegmentMatrix_Encode(t *testing.T) {
	var seg1, seg2, seg3 ExportSegment
	copy(seg1[:], []byte("test-segment-1"))
	copy(seg2[:], []byte("test-segment-2"))
	copy(seg3[:], []byte("test-segment-3"))

	matrix := ExportSegmentMatrix{
		{seg1, seg2}, // 2 segments in one row
		{seg3},       // 1 segment in second row
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&matrix)
	if err != nil {
		t.Fatalf("Error encoding ExportSegmentMatrix: %v", err)
	}

	expectedLength := 1 + // outer matrix length
		1 + // row[0] length
		1 + // row[1] length
		3*4104 // segment data

	if len(encoded) != expectedLength {
		t.Errorf("expected encoded length %d, got %d", expectedLength, len(encoded))
	}

	expected1 := make([]byte, 4104)
	copy(expected1, []byte("test-segment-1"))
	expected2 := make([]byte, 4104)
	copy(expected2, []byte("test-segment-2"))
	expected3 := make([]byte, 4104)
	copy(expected3, []byte("test-segment-3"))

	segmentStart := 1 + 1 // matrix len + row[0] len
	segment0 := encoded[segmentStart : segmentStart+4104]
	segment1 := encoded[segmentStart+4104 : segmentStart+2*4104]
	segment2 := encoded[segmentStart+2*4104+1 : segmentStart+3*4104+1] // row[1] len

	if !bytes.Equal(segment0, expected1) {
		t.Errorf("first segment mismatch")
		fmt.Printf("expected: %x, got: %x\n", expected1, segment0)
	}
	if !bytes.Equal(segment1, expected2) {
		t.Errorf("second segment mismatch")
	}
	if !bytes.Equal(segment2, expected3) {
		t.Errorf("third segment mismatch")
	}
}

func TestOpaqueHashMatrix_Encode(t *testing.T) {
	hash1 := OpaqueHash{1, 2, 3}
	hash2 := OpaqueHash{4, 5, 6}

	matrix := OpaqueHashMatrix{
		{hash1, hash2},
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&matrix)
	if err != nil {
		t.Fatalf("Error encoding OpaqueHashMatrix: %v", err)
	}

	offset := 0

	// matrix length
	if encoded[offset] != 1 {
		t.Errorf("expected matrix length = 1, got %d", encoded[offset])
	}
	offset += 1

	// row[0] length
	if encoded[offset] != 2 {
		t.Errorf("expected row[0] length = 2, got %d", encoded[offset])
	}
	offset += 1

	// row[0] hash length
	expected1 := hash1[:]
	actual1 := encoded[offset : offset+len(expected1)]
	if !bytes.Equal(actual1, expected1) {
		t.Errorf("first hash mismatch: expected %v, got %v", expected1, actual1)
	}
	offset += len(expected1)

	// row[1] hash length
	expected2 := hash2[:]
	actual2 := encoded[offset : offset+len(expected2)]
	if !bytes.Equal(actual2, expected2) {
		t.Errorf("second hash mismatch: expected %v, got %v", expected2, actual2)
	}
	offset += len(expected2)

	if offset != len(encoded) {
		t.Errorf("expected total length %d, got %d", offset, len(encoded))
	}
}
