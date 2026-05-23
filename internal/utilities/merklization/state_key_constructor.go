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

// StateServiceWrapper is a wrapper for U8 and ServiceID(U32)
type StateServiceWrapper struct {
	StateIndex   types.U8
	ServiceIndex types.ServiceID
}

// ServiceWrapper is a wrapper for ServiceID(U32) and a byte array of length 27
type ServiceWrapper struct {
	ServiceIndex types.ServiceID
	h            types.ByteSequence
}

func encodeServiceID(serviceID types.ServiceID) []byte {
	encoder := types.GetEncoder()
	defer types.PutEncoder(encoder)
	encoded, _ := encoder.EncodeUintWithLength(uint64(serviceID), 4)
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
	n := encodeServiceID(w.ServiceIndex)

	for i := 0; i <= 3; i++ {
		output[2*i+1] = n[i]
	}
	return output
}

// StateKeyConstruct returns a OpaqueHash
func (w ServiceWrapper) StateKeyConstruct() (output types.StateKey) {
	// [n_0, h_0, n_1, h_1, n_2, h_2, n_3, h_3, h_4, h_5,...,h_26] where n = encode_4(service_id)

	// Encode the service index
	n := encodeServiceID(w.ServiceIndex)

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

// StateKey constructor helpers.
//
// Wrappers over ServiceWrapper.StateKeyConstruct() that take business-level
// inputs (raw key bytes / preimage hash / preimage length) and produce the
// 31-byte StateKey used by globalKV. Corresponds to GP eq. (D.2).
//
// The signatures return (StateKey, error) even though the current computation
// cannot fail; this keeps the API forward-compatible with future JAM-encoder
// based implementations and aligns with KV-store conventions.

// NewStorageStateKey builds the StateKey for a storage (delta2) entry.
// GP eq. (D.2): C(s, E4(2^32 - 1) ⌢ k)
func NewStorageStateKey(serviceID types.ServiceID, rawKey types.ByteSequence) (types.StateKey, error) {
	h := make(types.ByteSequence, delta2PrefixLen+len(rawKey))
	copy(h[:delta2PrefixLen], delta2Prefix)
	copy(h[delta2PrefixLen:], rawKey)
	wrapper := ServiceWrapper{ServiceIndex: serviceID, h: h}
	return wrapper.StateKeyConstruct(), nil
}

// NewPreimageLookupStateKey builds the StateKey for a preimage lookup (delta3) entry.
// GP eq. (D.2): C(s, E4(2^32 - 2) ⌢ h)
//
// NOTE: PreimageLookup is kept as an independent map; this helper is used by
// deserialization (to identify delta3 entries) and tests only.
func NewPreimageLookupStateKey(serviceID types.ServiceID, preimageHash types.OpaqueHash) (types.StateKey, error) {
	h := make(types.ByteSequence, delta3PrefixLen+len(preimageHash))
	copy(h[:delta3PrefixLen], delta3Prefix)
	copy(h[delta3PrefixLen:], preimageHash[:])
	wrapper := ServiceWrapper{ServiceIndex: serviceID, h: h}
	return wrapper.StateKeyConstruct(), nil
}

// NewPreimageMetaStateKey builds the StateKey for a preimage meta / lookup-dict (delta4) entry.
// GP eq. (D.2): C(s, E4(l) ⌢ h)
func NewPreimageMetaStateKey(serviceID types.ServiceID, preimageHash types.OpaqueHash, length types.U32) (types.StateKey, error) {
	h := make(types.ByteSequence, uint32EncodedLen+len(preimageHash))
	v := uint32(length)
	h[0] = byte(v)
	h[1] = byte(v >> 8)
	h[2] = byte(v >> 16)
	h[3] = byte(v >> 24)
	copy(h[uint32EncodedLen:], preimageHash[:])
	wrapper := ServiceWrapper{ServiceIndex: serviceID, h: h}
	return wrapper.StateKeyConstruct(), nil
}
