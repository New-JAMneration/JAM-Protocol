package types

import (
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
