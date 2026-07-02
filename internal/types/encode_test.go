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

	serviceID := ServiceID(1)
	entropy := Entropy{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	timeSlot := TimeSlot(100)

	output := []byte{}
	encoder := NewEncoder()

	// Encode serviceID
	encoded, err := encoder.Encode(&serviceID)
	if err != nil {
		t.Errorf("Error encoding ServiceID: %v", err)
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
		Type: WorkExecResultOk,
		Data: []byte{1, 2, 3, 4, 5},
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
	var decodedResult WorkExecResult
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

// GP v0.8.0 eq:avspec: WorkPackageSpec gains an ErasureShards (u16) field
// between ErasureRoot and ExportsRoot. This guards the wire layout:
// round-trips, asserts the field survives, and pins the byte order so the
// 2-byte shard count sits immediately after the 32-byte erasure root.
func TestEncodeWorkPackageSpec_ErasureShards(t *testing.T) {
	spec := WorkPackageSpec{
		Hash:          WorkPackageHash{0x11},
		Length:        0x04030201,
		ErasureRoot:   ErasureRoot{0x22},
		ErasureShards: 0x0A0B,
		ExportsRoot:   ExportsRoot{0x33},
		ExportsCount:  0x0C0D,
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&spec)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Layout: hash(32) ++ length(4) ++ erasureRoot(32) ++ erasureShards(2)
	// ++ exportsRoot(32) ++ exportsCount(2) = 104 bytes.
	const want = 32 + 4 + 32 + 2 + 32 + 2
	if len(encoded) != want {
		t.Fatalf("encoded length = %d, want %d", len(encoded), want)
	}
	// ErasureShards (0x0A0B little-endian) must follow the erasure root.
	off := 32 + 4 + 32
	if encoded[off] != 0x0B || encoded[off+1] != 0x0A {
		t.Errorf("erasure_shards bytes = %x %x at offset %d, want 0b 0a",
			encoded[off], encoded[off+1], off)
	}

	decoder := NewDecoder()
	var got WorkPackageSpec
	if err := decoder.Decode(encoded, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(spec, got) {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", got, spec)
	}
}

// TestEncodeRefineContext_V080Fields covers the GP v0.8.0 eq:workcontext
// additions: anchor_slot (u32) right after the anchor hash, and
// lookup_anchor_state_root (32 bytes) right after lookup_anchor_slot.
func TestEncodeRefineContext_V080Fields(t *testing.T) {
	ctx := RefineContext{
		Anchor:                HeaderHash{0x11},
		AnchorSlot:            0x0A0B0C0D,
		StateRoot:             StateRoot{0x22},
		BeefyRoot:             BeefyRoot{0x33},
		LookupAnchor:          HeaderHash{0x44},
		LookupAnchorSlot:      0x01020304,
		LookupAnchorStateRoot: StateRoot{0x55},
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&ctx)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Layout: anchor(32) ++ anchorSlot(4) ++ stateRoot(32) ++ beefyRoot(32)
	// ++ lookupAnchor(32) ++ lookupAnchorSlot(4) ++ lookupAnchorStateRoot(32)
	// ++ prerequisites length prefix(1, value 0) = 169 bytes.
	const want = 32 + 4 + 32 + 32 + 32 + 4 + 32 + 1
	if len(encoded) != want {
		t.Fatalf("encoded length = %d, want %d", len(encoded), want)
	}
	// anchor_slot (0x0A0B0C0D little-endian) must follow the anchor hash.
	if off := 32; encoded[off] != 0x0D || encoded[off+1] != 0x0C ||
		encoded[off+2] != 0x0B || encoded[off+3] != 0x0A {
		t.Errorf("anchor_slot bytes = % x at offset %d, want 0d 0c 0b 0a",
			encoded[off:off+4], off)
	}
	// lookup_anchor_state_root (first byte 0x55) must follow lookup_anchor_slot.
	if off := 32 + 4 + 32 + 32 + 32 + 4; encoded[off] != 0x55 {
		t.Errorf("lookup_anchor_state_root[0] = %x at offset %d, want 55",
			encoded[off], off)
	}

	decoder := NewDecoder()
	var got RefineContext
	if err := decoder.Decode(encoded, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(ctx, got) {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", got, ctx)
	}
}

// TestEncodeBlockInfo_V080Timeslot covers the GP v0.8.0 eq:recenthistoryspec /
// C(3) addition: a 4-byte timeslot between the state root and the reported
// work packages.
func TestEncodeBlockInfo_V080Timeslot(t *testing.T) {
	info := BlockInfo{
		HeaderHash: HeaderHash{0x11},
		BeefyRoot:  OpaqueHash{0x22},
		StateRoot:  StateRoot{0x33},
		Timeslot:   0x0A0B0C0D,
		Reported: []ReportedWorkPackage{
			{Hash: WorkReportHash{0x44}, ExportsRoot: ExportsRoot{0x55}},
		},
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&info)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Layout: headerHash(32) ++ beefyRoot(32) ++ stateRoot(32) ++ timeslot(4)
	// ++ reported length prefix(1, value 1) ++ reported[0](32+32) = 165 bytes.
	const want = 32 + 32 + 32 + 4 + 1 + 64
	if len(encoded) != want {
		t.Fatalf("encoded length = %d, want %d", len(encoded), want)
	}
	// timeslot (0x0A0B0C0D little-endian) must follow the state root.
	if off := 32 + 32 + 32; encoded[off] != 0x0D || encoded[off+1] != 0x0C ||
		encoded[off+2] != 0x0B || encoded[off+3] != 0x0A {
		t.Errorf("timeslot bytes = % x at offset %d, want 0d 0c 0b 0a",
			encoded[off:off+4], off)
	}

	decoder := NewDecoder()
	var got BlockInfo
	if err := decoder.Decode(encoded, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(info, got) {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", got, info)
	}
}

// TestEncodeEpochMark_V080LengthPrefix covers the GP v0.8.0 encodeepochmark
// change: the validator-key sequence is length-prefixed (var{k}); v0.7.x
// emitted it fixed-length.
func TestEncodeEpochMark_V080LengthPrefix(t *testing.T) {
	mark := EpochMark{
		Entropy:        Entropy{0x11},
		TicketsEntropy: Entropy{0x22},
		Validators:     make([]EpochMarkValidatorKeys, ValidatorsCount),
	}
	for i := range mark.Validators {
		mark.Validators[i].Bandersnatch = BandersnatchPublic{byte(i + 1)}
		mark.Validators[i].Ed25519 = Ed25519Public{byte(i + 1)}
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&mark)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Layout: entropy(32) ++ ticketsEntropy(32) ++ length prefix(1, value
	// ValidatorsCount) ++ ValidatorsCount * (bandersnatch 32 + ed25519 32).
	want := 32 + 32 + 1 + ValidatorsCount*64
	if len(encoded) != want {
		t.Fatalf("encoded length = %d, want %d", len(encoded), want)
	}
	// The length prefix must follow the two entropies.
	if off := 32 + 32; encoded[off] != byte(ValidatorsCount) {
		t.Errorf("validators length prefix = %d at offset %d, want %d",
			encoded[off], off, ValidatorsCount)
	}

	decoder := NewDecoder()
	var got EpochMark
	if err := decoder.Decode(encoded, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(mark, got) {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", got, mark)
	}
}

// TestEncodeVerdict_V080LengthPrefix covers the GP v0.8.0 encodedisputes
// change: a verdict's judgment sequence is length-prefixed (var{...});
// v0.7.x emitted a fixed ValidatorsSuperMajority entries.
func TestEncodeVerdict_V080LengthPrefix(t *testing.T) {
	verdict := Verdict{
		Target: WorkReportHash{0x11},
		Age:    0x0A0B0C0D,
		Votes:  make([]Judgement, ValidatorsSuperMajority),
	}
	for i := range verdict.Votes {
		verdict.Votes[i] = Judgement{
			Vote:      i%2 == 0,
			Index:     ValidatorIndex(i),
			Signature: Ed25519Signature{byte(i + 1)},
		}
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&verdict)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Layout: target(32) ++ age(4) ++ length prefix(1, value
	// ValidatorsSuperMajority) ++ votes * (vote 1 + index 2 + signature 64).
	want := 32 + 4 + 1 + ValidatorsSuperMajority*(1+2+64)
	if len(encoded) != want {
		t.Fatalf("encoded length = %d, want %d", len(encoded), want)
	}
	// The length prefix must follow target ++ age.
	if off := 32 + 4; encoded[off] != byte(ValidatorsSuperMajority) {
		t.Errorf("votes length prefix = %d at offset %d, want %d",
			encoded[off], off, ValidatorsSuperMajority)
	}

	decoder := NewDecoder()
	var got Verdict
	if err := decoder.Decode(encoded, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(verdict, got) {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", got, verdict)
	}
}
