package utilities

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// U8SignatureWrapper is a wrapper for U8 and U32
type U8SignatureWrapper struct {
	Value     jamTypes.U8
	Signature jamTypes.U32
}

// U32HashWrapper is a wrapper for U32 and ByteSequence
type U32HashWrapper struct {
	Value jamTypes.U32
	Hash  jamTypes.ByteSequence
}

// D.1 State-Key-Construction

// StateKeyConstruct returns a StringOctets
func (w U8Wrapper) StateKeyConstruct() (output jamTypes.ByteArray32) {
	output[0] = byte(w.Value)
	return output
}

// StateKeyConstruct returns a StringOctets
func (w U8SignatureWrapper) StateKeyConstruct() (output jamTypes.ByteArray32) {
	output[0] = byte(w.Value)
	Serialized := SerializeFixedLength(w.Signature, 4)
	for i := 0; i < 4; i++ {
		output[2*i+1] = Serialized[i]
	}
	return output
}

// StateKeyConstruct returns a StringOctets
func (w U32HashWrapper) StateKeyConstruct() (output jamTypes.ByteArray32) {
	Serialized := SerializeFixedLength(w.Value, 4)
	for i := 0; i < 4; i++ {
		output[2*i] = Serialized[i]
		output[2*i+1] = w.Hash[i]
	}
	for i := 4; i < 28; i++ {
		output[i+4] = w.Hash[i]
	}
	return output
}
