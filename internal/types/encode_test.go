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
