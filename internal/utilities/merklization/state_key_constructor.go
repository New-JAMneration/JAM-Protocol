package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

// StateKeyConstruct is an interface
type StateKeyConstruct interface {
	StateKeyConstruct() (output types.OpaqueHash)
}

// StateWrapper is a wrapper for U8
type StateWrapper struct {
	StateIndex types.U8
}

// StateServiceWrapper is a wrapper for U8 and U32
type StateServiceWrapper struct {
	StateIndex   types.U8
	ServiceIndex types.U32
}

// ServiceWrapper is a wrapper for U32 and OpaqueHash
type ServiceWrapper struct {
	ServiceIndex types.U32
	Hash         types.OpaqueHash
}

// D.1 State-Key-Construction

// StateKeyConstruct returns a OpaqueHash
func (s StateWrapper) StateKeyConstruct() (output types.OpaqueHash) {
	output[0] = byte(s.StateIndex)
	return output
}

// StateKeyConstruct returns a OpaqueHash
func (w StateServiceWrapper) StateKeyConstruct() (output types.OpaqueHash) {
	output[0] = byte(w.StateIndex)
	Serialized := utilities.SerializeFixedLength(w.ServiceIndex, 4)
	for i := 0; i < 4; i++ {
		output[2*i+1] = Serialized[i]
	}
	return output
}

// StateKeyConstruct returns a OpaqueHash
func (w ServiceWrapper) StateKeyConstruct() (output types.OpaqueHash) {
	Serialized := utilities.SerializeFixedLength(w.ServiceIndex, 4)
	for i := 0; i < 4; i++ {
		output[2*i] = Serialized[i]
		output[2*i+1] = w.Hash[i]
	}
	for i := 4; i < 28; i++ {
		output[i+4] = w.Hash[i]
	}
	return output
}
