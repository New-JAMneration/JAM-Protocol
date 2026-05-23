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
)

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
// NOTE: this method intentionally preserves the historical behaviour of the
// standalone `CalcThresholdBalance` helper: when the storage sum is greater
// than or equal to a_f it returns the storage sum itself (instead of
// `storage - a_f`). That is a pre-existing bug tracked separately and will
// be fixed AFTER the global-KV refactor and the three fuzz suites have been
// re-validated, to keep the refactor diff easy to audit.
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

	if storage < aF {
		return 0, nil
	}
	// Pre-existing bug: should be `storage - aF` per GP §9.8. Intentionally
	// preserved during this refactor; see method-level comment.
	return U64(storage), nil
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
