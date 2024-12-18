package utilities

import (
	"testing"

	jamtypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func TestDeserializeFixedLength(t *testing.T) {
	tests := []jamtypes.U32{1, 127, 128, 255}
	for _, tt := range tests {
		ser := SerializeFixedLength(tt, jamtypes.U32(2))
		data := jamtypes.ByteSequence(ser)
		deser, err := DeserializeFixedLength(data, jamtypes.U32(2))
		if err != nil {
			t.Fatalf("DeserializeU8Wrapper failed: %v", err)
		}
		if deser != tt {
			t.Fatalf("U8 roundtrip failed. Expected %d, got %d", tt, deser)
		}
	}
}

func TestU8Serialization(t *testing.T) {
	tests := []jamtypes.U8{1, 127, 128, 255}
	for _, tt := range tests {
		orig := U8Wrapper{Value: tt}
		ser := orig.Serialize()
		data := jamtypes.ByteSequence(ser) // make a copy we can consume
		deser, err := DeserializeU8Wrapper(data)
		if err != nil {
			t.Fatalf("DeserializeU8Wrapper failed: %v", err)
		}
		if deser.Value != tt {
			t.Fatalf("U8 failed. Expected %d, got %d", tt, deser.Value)
		}
	}
}

func TestU16Serialization(t *testing.T) {
	tests := []jamtypes.U16{0, 1, 255, 256, 16639}
	for _, tt := range tests {
		orig := U16Wrapper{Value: tt}
		ser := orig.Serialize()
		data := jamtypes.ByteSequence(ser)
		deser, err := DeserializeU16Wrapper(data)
		if err != nil {
			t.Fatalf("DeserializeU16Wrapper failed: %v", err)
		}
		if deser.Value != tt {
			t.Fatalf("U16 roundtrip failed. Expected %d got %d", tt, deser.Value)
		}
	}
}

func TestU32Serialization(t *testing.T) {
	tests := []jamtypes.U32{0, 1, 65535, 65536, 4294967295}
	for _, tt := range tests {
		orig := U32Wrapper{Value: tt}
		ser := orig.Serialize()
		data := jamtypes.ByteSequence(ser)
		deser, err := DeserializeU32Wrapper(data)
		if err != nil {
			t.Fatalf("DeserializeU32Wrapper failed: %v", err)
		}
		if deser.Value != tt {
			t.Fatalf("U32 failed. Expected %d got %d", tt, deser.Value)
		}
	}
}

func TestU64Serialization(t *testing.T) {
	tests := []jamtypes.U64{0, 1, 255, 256, 65535, 65536, 4294967295, 4294967296, 18446744073709551615}
	for _, tt := range tests {
		orig := U64Wrapper{Value: tt}
		ser := orig.Serialize()
		data := jamtypes.ByteSequence(ser)
		deser, err := DeserializeU64Wrapper(data)
		if err != nil {
			t.Fatalf("DeserializeU64Wrapper failed: %v", err)
		}
		if deser.Value != tt {
			t.Fatalf("U64 failed. Expected %d got %d", tt, deser.Value)
		}
	}
}
