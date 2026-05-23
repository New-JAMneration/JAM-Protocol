// ServiceAccount globalKV accessors and counter maintenance.
//
// globalKV unifies service storage (a_s) and preimage meta (a_l) entries,
// keyed by StateKey. PreimageLookup (a_p) is handled elsewhere and is not
// merged into globalKV.
//
// Design notes (GP §9.6 / §9.8):
//   - Insert/Delete keep totalNumberOfItems / totalNumberOfOctets (a_i / a_o)
//     in sync with the underlying map.
//   - All counter math goes through safemath; the methods are atomic:
//     intermediate values are kept in local variables and only assigned back
//     to the struct once every step has succeeded. On overflow nothing is
//     mutated.
//   - For storage entries the globalKV value is the raw byte slice provided
//     by the caller. For preimage-meta entries the value is the JAM
//     encoding of TimeSlotSet (matching the on-the-wire format produced by
//     the existing EncodeDelta4KeyVal helper).
package types

import (
	"bytes"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/safemath"
	"golang.org/x/crypto/blake2b"
)

// buildStorageStateKey mirrors merklization.NewStorageStateKey but lives in
// the types package so that callers like unmarshal_json.go can populate
// globalKV without creating a circular import.
//
// GP eq. (D.2): C(s, E4(2^32 - 1) ⌢ k) — prefix is 0xFFFFFFFF.
func buildStorageStateKey(serviceID ServiceID, rawKey ByteSequence) StateKey {
	preimage := make([]byte, 4+len(rawKey))
	preimage[0], preimage[1], preimage[2], preimage[3] = 0xFF, 0xFF, 0xFF, 0xFF
	copy(preimage[4:], rawKey)
	return interleaveServiceIDIntoHash(serviceID, preimage)
}

// buildPreimageMetaStateKey mirrors merklization.NewPreimageMetaStateKey for
// the same reason as buildStorageStateKey.
//
// GP eq. (D.2): C(s, E4(l) ⌢ h)
func buildPreimageMetaStateKey(serviceID ServiceID, hash OpaqueHash, length U32) StateKey {
	preimage := make([]byte, 4+len(hash))
	v := uint32(length)
	preimage[0] = byte(v)
	preimage[1] = byte(v >> 8)
	preimage[2] = byte(v >> 16)
	preimage[3] = byte(v >> 24)
	copy(preimage[4:], hash[:])
	return interleaveServiceIDIntoHash(serviceID, preimage)
}

// interleaveServiceIDIntoHash performs the StateKey-type-3 layout from
// merklization.ServiceWrapper.StateKeyConstruct:
//
//	[n0, h0, n1, h1, n2, h2, n3, h3, h4, h5, ..., h26]
//
// where n = encode_4(serviceID) and h = Blake2b(preimage)[:27].
func interleaveServiceIDIntoHash(serviceID ServiceID, preimage []byte) StateKey {
	digest := blake2b.Sum256(preimage)
	h := digest[:27]

	var n [4]byte
	v := uint32(serviceID)
	n[0] = byte(v)
	n[1] = byte(v >> 8)
	n[2] = byte(v >> 16)
	n[3] = byte(v >> 24)

	var out StateKey
	for i := 0; i <= 3; i++ {
		out[2*i] = n[i]
		out[2*i+1] = h[i]
	}
	for i := 4; i <= 26; i++ {
		out[i+4] = h[i]
	}
	return out
}

// NewServiceAccount is declared in state.go.

// GetGlobalKVItems returns the underlying globalKV map. Primarily used by
// serialization.
func (sa *ServiceAccount) GetGlobalKVItems() map[StateKey][]byte {
	return sa.globalKV
}

// SetGlobalKVItems replaces the underlying globalKV map. Primarily used by
// deserialization; counters are NOT updated here (the deserialization path
// initialises them directly from the wire format via SetTotalNumberOf*).
func (sa *ServiceAccount) SetGlobalKVItems(globalKV map[StateKey][]byte) {
	sa.globalKV = globalKV
}

// GetTotalNumberOfItems returns a_i (the items counter).
func (sa *ServiceAccount) GetTotalNumberOfItems() uint32 {
	return sa.totalNumberOfItems
}

// SetTotalNumberOfItems overrides a_i without any consistency check. Use only
// in the deserialization path.
func (sa *ServiceAccount) SetTotalNumberOfItems(n uint32) {
	sa.totalNumberOfItems = n
}

// GetTotalNumberOfOctets returns a_o (the octets counter).
func (sa *ServiceAccount) GetTotalNumberOfOctets() uint64 {
	return sa.totalNumberOfOctets
}

// SetTotalNumberOfOctets overrides a_o without any consistency check. Use only
// in the deserialization path.
func (sa *ServiceAccount) SetTotalNumberOfOctets(n uint64) {
	sa.totalNumberOfOctets = n
}

// ----- Storage (a_s) -----

// GetStorage reads a storage value from globalKV.
func (sa *ServiceAccount) GetStorage(key StateKey) ([]byte, bool) {
	value, ok := sa.globalKV[key]
	return value, ok
}

// InsertStorage inserts or updates a storage entry.
//
// originalKeySize is the length of the caller's raw key (the StateKey is a
// Blake2b hash and cannot be reversed).
//
// Counter updates:
//   - New key:      items += 1, octets += 34 + originalKeySize + len(value).
//   - Existing key: items unchanged; octets first Sub(len(prevValue)) then
//     Add(len(newValue)) to avoid uint64 underflow when the new value is
//     smaller than the old one.
//
// The method does NOT treat an empty value as a delete; callers must invoke
// DeleteStorage explicitly. On overflow ErrOverflow is returned and neither
// the counters nor globalKV are mutated.
func (sa *ServiceAccount) InsertStorage(key StateKey, originalKeySize uint64, value []byte) error {
	if sa.globalKV == nil {
		sa.globalKV = make(map[StateKey][]byte)
	}

	newItems := sa.totalNumberOfItems
	newOctets := sa.totalNumberOfOctets

	if prevVal, exists := sa.globalKV[key]; !exists {
		var ok bool
		newItems, ok = safemath.Add(newItems, uint32(1))
		if !ok {
			return safemath.ErrOverflow
		}
		newOctets, ok = safemath.Add(newOctets, uint64(34))
		if !ok {
			return safemath.ErrOverflow
		}
		newOctets, ok = safemath.Add(newOctets, originalKeySize)
		if !ok {
			return safemath.ErrOverflow
		}
		newOctets, ok = safemath.Add(newOctets, uint64(len(value)))
		if !ok {
			return safemath.ErrOverflow
		}
	} else {
		var ok bool
		// Sub-then-Add to avoid uint64 underflow when the new value is
		// smaller than the previous one.
		newOctets, ok = safemath.Sub(newOctets, uint64(len(prevVal)))
		if !ok {
			return safemath.ErrOverflow
		}
		newOctets, ok = safemath.Add(newOctets, uint64(len(value)))
		if !ok {
			return safemath.ErrOverflow
		}
	}

	sa.globalKV[key] = value
	sa.totalNumberOfItems = newItems
	sa.totalNumberOfOctets = newOctets
	return nil
}

// DeleteStorage removes a storage entry and updates the counters.
//
// keyLen / valueLen must be supplied by the caller (the StateKey is a hash
// and cannot be reversed). The operation is idempotent: deleting a missing
// key returns nil and does not touch the counters.
func (sa *ServiceAccount) DeleteStorage(key StateKey, keyLen, valueLen uint64) error {
	if _, exists := sa.globalKV[key]; !exists {
		return nil
	}

	newItems, ok := safemath.Sub(sa.totalNumberOfItems, uint32(1))
	if !ok {
		return safemath.ErrOverflow
	}
	newOctets, ok := safemath.Sub(sa.totalNumberOfOctets, uint64(34))
	if !ok {
		return safemath.ErrOverflow
	}
	newOctets, ok = safemath.Sub(newOctets, keyLen)
	if !ok {
		return safemath.ErrOverflow
	}
	newOctets, ok = safemath.Sub(newOctets, valueLen)
	if !ok {
		return safemath.ErrOverflow
	}

	delete(sa.globalKV, key)
	sa.totalNumberOfItems = newItems
	sa.totalNumberOfOctets = newOctets
	return nil
}

// ----- PreimageMeta (a_l, the timeslot metadata previously stored in LookupDict) -----

// encodePreimageMetaValue serialises a TimeSlotSet using the same wire format
// as EncodeDelta4KeyVal (the existing delta4 value layout). Concretely:
//   [variable-length length prefix (JAM EncodeUint)] ++ [each timeslot as 4-byte LE]
//
// Keeping the same wire format is critical for state-root compatibility once
// we start writing TimeSlotSet values into globalKV.
func encodePreimageMetaValue(timeslots TimeSlotSet) ([]byte, error) {
	enc := GetEncoder()
	defer PutEncoder(enc)
	return enc.Encode(&timeslots)
}

// decodePreimageMetaValue is the inverse of encodePreimageMetaValue.
func decodePreimageMetaValue(data []byte) (TimeSlotSet, error) {
	dec := NewDecoder()
	var out TimeSlotSet
	if err := dec.Decode(bytes.Clone(data), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPreimageMeta reads and decodes a preimage-meta TimeSlotSet from globalKV.
// Any decoding failure is surfaced as (nil, false) so that the public API
// matches "key not present".
func (sa *ServiceAccount) GetPreimageMeta(key StateKey) (TimeSlotSet, bool) {
	data, ok := sa.globalKV[key]
	if !ok {
		return nil, false
	}
	out, err := decodePreimageMetaValue(data)
	if err != nil {
		return nil, false
	}
	return out, true
}

// InsertPreimageMeta inserts (or overwrites) a preimage-meta entry.
//
// The timeslots are JAM-encoded before being written into globalKV.
//
// Counter updates:
//   - New key:      items += 2, octets += 81 + length.
//   - Existing key: value is overwritten without touching the counters.
//
// length is the length of the original preimage blob (the z component of
// the LookupMetaMapkey / GP a_l index).
func (sa *ServiceAccount) InsertPreimageMeta(key StateKey, length uint64, timeslots TimeSlotSet) error {
	encoded, err := encodePreimageMetaValue(timeslots)
	if err != nil {
		return err
	}
	if sa.globalKV == nil {
		sa.globalKV = make(map[StateKey][]byte)
	}

	newItems := sa.totalNumberOfItems
	newOctets := sa.totalNumberOfOctets

	if _, exists := sa.globalKV[key]; !exists {
		var ok bool
		newItems, ok = safemath.Add(newItems, uint32(2))
		if !ok {
			return safemath.ErrOverflow
		}
		newOctets, ok = safemath.Add(newOctets, uint64(81))
		if !ok {
			return safemath.ErrOverflow
		}
		newOctets, ok = safemath.Add(newOctets, length)
		if !ok {
			return safemath.ErrOverflow
		}
	}

	sa.globalKV[key] = encoded
	sa.totalNumberOfItems = newItems
	sa.totalNumberOfOctets = newOctets
	return nil
}

// UpdatePreimageMeta overwrites the value of an existing preimage-meta entry
// without touching the counters.
//
// Unlike InsertPreimageMeta this method refuses to lazy-init globalKV and
// refuses to create a missing key — both of those are logical errors here.
// Existence is checked directly against the map (not via GetPreimageMeta) so
// that a decoding failure is not silently mistaken for a missing key.
func (sa *ServiceAccount) UpdatePreimageMeta(key StateKey, newValue TimeSlotSet) error {
	if sa.globalKV == nil {
		return fmt.Errorf("UpdatePreimageMeta: globalKV is nil")
	}
	if _, exists := sa.globalKV[key]; !exists {
		return fmt.Errorf("UpdatePreimageMeta: key does not exist")
	}
	encoded, err := encodePreimageMetaValue(newValue)
	if err != nil {
		return err
	}
	sa.globalKV[key] = encoded
	return nil
}

// DeletePreimageMeta removes a preimage-meta entry and updates the counters.
// Idempotent: deleting a missing key returns nil and does not touch the
// counters.
func (sa *ServiceAccount) DeletePreimageMeta(key StateKey, length uint64) error {
	if _, exists := sa.globalKV[key]; !exists {
		return nil
	}

	newItems, ok := safemath.Sub(sa.totalNumberOfItems, uint32(2))
	if !ok {
		return safemath.ErrOverflow
	}
	newOctets, ok := safemath.Sub(sa.totalNumberOfOctets, uint64(81))
	if !ok {
		return safemath.ErrOverflow
	}
	newOctets, ok = safemath.Sub(newOctets, length)
	if !ok {
		return safemath.ErrOverflow
	}

	delete(sa.globalKV, key)
	sa.totalNumberOfItems = newItems
	sa.totalNumberOfOctets = newOctets
	return nil
}

// ----- Threshold balance (a_t) -----

// ThresholdBalance returns the per-account minimum balance threshold a_t.
// GP §9.8: a_t ≡ max(0, B_S + B_I*a_i + B_L*a_o − a_f).
//
// All arithmetic goes through safemath so the operation cannot wrap silently.
func (sa *ServiceAccount) ThresholdBalance() (U64, error) {
	aI := uint64(sa.totalNumberOfItems)
	aO := uint64(sa.totalNumberOfOctets)
	aF := uint64(sa.ServiceInfo.DepositOffset)

	itemsContribution, ok := safemath.Mul(uint64(AdditionalMinBalancePerItem), aI)
	if !ok {
		return 0, safemath.ErrOverflow
	}
	octetsContribution, ok := safemath.Mul(uint64(AdditionalMinBalancePerOctet), aO)
	if !ok {
		return 0, safemath.ErrOverflow
	}

	storage, ok := safemath.Add(uint64(BasicMinBalance), itemsContribution)
	if !ok {
		return 0, safemath.ErrOverflow
	}
	storage, ok = safemath.Add(storage, octetsContribution)
	if !ok {
		return 0, safemath.ErrOverflow
	}

	// max(0, storage - aF). The storage < aF branch implements the max(0, _)
	// clamp; the storage >= aF branch must subtract the GratisStorageOffset.
	if storage < aF {
		return 0, nil
	}
	diff, ok := safemath.Sub(storage, aF)
	if !ok {
		return 0, safemath.ErrOverflow
	}
	return U64(diff), nil
}

// MigrateLegacyMapsToGlobalKV walks the deprecated StorageDict / LookupDict
// maps and re-installs each entry in globalKV via InsertStorage /
// InsertPreimageMeta. Counters are reset to zero before the walk so that
// the resulting (a_i, a_o) match what the Insert methods would have
// produced from scratch.
//
// Intended for test fixtures that still build a ServiceAccount with a
// struct literal seeding the legacy maps directly — calling this once
// after construction yields a ServiceAccount that behaves identically to
// one built incrementally through Insert*.
func (sa *ServiceAccount) MigrateLegacyMapsToGlobalKV(serviceID ServiceID) error {
	sa.globalKV = make(map[StateKey][]byte)
	sa.totalNumberOfItems = 0
	sa.totalNumberOfOctets = 0

	for lookupKey, timeslots := range sa.LookupDict {
		stateKey := buildPreimageMetaStateKey(serviceID, lookupKey.Hash, lookupKey.Length)
		if err := sa.InsertPreimageMeta(stateKey, uint64(lookupKey.Length), timeslots); err != nil {
			return fmt.Errorf("MigrateLegacyMapsToGlobalKV: InsertPreimageMeta: %w", err)
		}
	}
	for rawKey, value := range sa.StorageDict {
		stateKey := buildStorageStateKey(serviceID, ByteSequence(rawKey))
		if err := sa.InsertStorage(stateKey, uint64(len(rawKey)), value); err != nil {
			return fmt.Errorf("MigrateLegacyMapsToGlobalKV: InsertStorage: %w", err)
		}
	}
	return nil
}

// ----- Clone / DeepCopy -----

// Clone returns a deep copy of the ServiceAccount.
// `cloned := *sa` copies every scalar field, including unexported counters;
// the maps are then deep-cloned separately.
func (sa *ServiceAccount) Clone() ServiceAccount {
	cloned := *sa
	cloned.globalKV = cloneMapOfSlices(sa.globalKV)
	cloned.PreimageLookup = cloneMapOfSlices(sa.PreimageLookup)
	// Transitional: StorageDict / LookupDict are still dual-written.
	// These two map clones go away together with the fields in Step 8.
	cloned.StorageDict = cloneMapOfSlices(sa.StorageDict)
	cloned.LookupDict = cloneLookupDict(sa.LookupDict)
	return cloned
}

// Clone deep-copies the entire ServiceAccountState.
func (ss ServiceAccountState) Clone() ServiceAccountState {
	if ss == nil {
		return nil
	}
	cloned := make(ServiceAccountState, len(ss))
	for id, account := range ss {
		cloned[id] = account.Clone()
	}
	return cloned
}

// cloneMapOfSlices deep-copies a map whose values are slice-typed. The
// constraint allows named slice types (e.g. ByteSequence) thanks to the ~
// approximation operator.
func cloneMapOfSlices[M ~map[K]V, K comparable, V ~[]E, E any](src M) M {
	if src == nil {
		return nil
	}
	dst := make(M, len(src))
	for k, v := range src {
		cp := make(V, len(v))
		copy(cp, v)
		dst[k] = cp
	}
	return dst
}

// cloneLookupDict deep-copies a LookupMetaMapEntry value.
// Transitional helper; removed alongside the LookupDict field in Step 8.
func cloneLookupDict(src LookupMetaMapEntry) LookupMetaMapEntry {
	if src == nil {
		return nil
	}
	dst := make(LookupMetaMapEntry, len(src))
	for k, v := range src {
		cp := make(TimeSlotSet, len(v))
		copy(cp, v)
		dst[k] = cp
	}
	return dst
}
