package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/statekey"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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

// StateKeyConstruct returns a StateKey using the type-3 interleave layout.
// Delegates to statekey.Interleave (single source of truth for type-3 key construction).
func (w ServiceWrapper) StateKeyConstruct() (output types.StateKey) {
	return types.StateKey(statekey.Interleave(uint32(w.ServiceIndex), w.h))
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
//
// Delegates to statekey.Storage (single source of truth for type-3 key construction).
func NewStorageStateKey(serviceID types.ServiceID, rawKey types.ByteSequence) (types.StateKey, error) {
	return types.StateKey(statekey.Storage(uint32(serviceID), rawKey)), nil
}

// NewPreimageLookupStateKey builds the StateKey for a preimage lookup (delta3) entry.
// GP eq. (D.2): C(s, E4(2^32 - 2) ⌢ h)
//
// NOTE: PreimageLookup is kept as an independent map; this helper is used by
// deserialization (to identify delta3 entries) and tests only.
//
// Delegates to statekey.PreimageLookup (single source of truth for type-3 key construction).
func NewPreimageLookupStateKey(serviceID types.ServiceID, preimageHash types.OpaqueHash) (types.StateKey, error) {
	return types.StateKey(statekey.PreimageLookup(uint32(serviceID), [32]byte(preimageHash))), nil
}

// NewPreimageMetaStateKey builds the StateKey for a preimage meta / lookup-dict (delta4) entry.
// GP eq. (D.2): C(s, E4(l) ⌢ h)
//
// Delegates to statekey.PreimageMeta (single source of truth for type-3 key construction).
func NewPreimageMetaStateKey(serviceID types.ServiceID, preimageHash types.OpaqueHash, length types.U32) (types.StateKey, error) {
	return types.StateKey(statekey.PreimageMeta(uint32(serviceID), [32]byte(preimageHash), uint32(length))), nil
}
