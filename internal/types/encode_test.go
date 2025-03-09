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
