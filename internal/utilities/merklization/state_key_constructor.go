package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// StateKeyConstruct is an interface
type StateKeyConstruct interface {
	StateKeyConstruct() (output types.StateKey)
}

// StateWrapper is a wrapper for U8
type StateWrapper struct {
	StateIndex types.U8
}

// StateServiceWrapper is a wrapper for U8 and ServiceId(U32)
type StateServiceWrapper struct {
	StateIndex   types.U8
	ServiceIndex types.ServiceId
}

// ServiceWrapper is a wrapper for ServiceId(U32) and a byte array of length 27
type ServiceWrapper struct {
	ServiceIndex types.ServiceId
	h            types.ByteSequence
}

func encodeServiceId(serviceId types.ServiceId) []byte {
	encoder := types.NewEncoder()
	encoded, _ := encoder.EncodeUintWithLength(uint64(serviceId), 4)

	return encoded
}

// D.1 State-Key-Construction
func (s StateWrapper) StateKeyConstruct() (output types.StateKey) {
	// [i, 0, 0,...]
	output[0] = byte(s.StateIndex)
	return output
}

func (w StateServiceWrapper) StateKeyConstruct() (output types.StateKey) {
	// [i, n_0, 0, n_1, 0, n2, 0, n3, 0, 0,...] where n = encode_4(service_id)
	output[0] = byte(w.StateIndex)

	// Encode the service index
	n := encodeServiceId(w.ServiceIndex)

	for i := 0; i <= 3; i++ {
		output[2*i+1] = n[i]
	}
	return output
}

// StateKeyConstruct returns a OpaqueHash
func (w ServiceWrapper) StateKeyConstruct() (output types.StateKey) {
	// [n_0, h_0, n_1, h_1, n_2, h_2, n_3, h_3, h_4, h_5,...,h_26] where n = encode_4(service_id)

	// Encode the service index
	n := encodeServiceId(w.ServiceIndex)

	a := hash.Blake2bHashPartial(w.h[:], 27)

	for i := 0; i <= 3; i++ {
		output[2*i] = n[i]
		output[2*i+1] = a[i]
	}

	for i := 4; i <= 26; i++ {
		output[i+4] = a[i]
	}

	return output
}
