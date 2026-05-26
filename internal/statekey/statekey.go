// Package statekey provides canonical StateKey construction for GP D.1 type-3 keys.
//
// This package has zero internal dependencies (only golang.org/x/crypto/blake2b)
// so that both internal/types and internal/utilities/merklization can import it
// without creating a circular dependency.
package statekey

import "golang.org/x/crypto/blake2b"

const (
	// StoragePrefix is E4(2^32 - 1) = 0xFFFFFFFF, used for delta2 (storage) keys.
	storagePrefixB0 = 0xFF
	storagePrefixB1 = 0xFF
	storagePrefixB2 = 0xFF
	storagePrefixB3 = 0xFF

	// PreimageLookupPrefix is E4(2^32 - 2) = 0xFEFFFFFF, used for delta3 keys.
	preimageLookupPrefixB0 = 0xFE
	preimageLookupPrefixB1 = 0xFF
	preimageLookupPrefixB2 = 0xFF
	preimageLookupPrefixB3 = 0xFF
)

// Interleave constructs a 31-byte state key (GP D.1 type-3 layout) from a
// service ID and a variable-length preimage.
//
// Layout: [n0, h0, n1, h1, n2, h2, n3, h3, h4, h5, ..., h26]
// where n = encode_4(serviceID) (little-endian) and h = Blake2b(preimage)[:27].
func Interleave(serviceID uint32, preimage []byte) [31]byte {
	digest := blake2b.Sum256(preimage)

	var out [31]byte
	out[0] = byte(serviceID)
	out[1] = digest[0]
	out[2] = byte(serviceID >> 8)
	out[3] = digest[1]
	out[4] = byte(serviceID >> 16)
	out[5] = digest[2]
	out[6] = byte(serviceID >> 24)
	out[7] = digest[3]
	for i := 4; i < 27; i++ {
		out[i+4] = digest[i]
	}
	return out
}

// Storage builds the StateKey for a storage (delta2) entry.
// GP eq. (D.2): C(s, E4(2^32 - 1) ⌢ k)
func Storage(serviceID uint32, rawKey []byte) [31]byte {
	preimage := make([]byte, 4+len(rawKey))
	preimage[0] = storagePrefixB0
	preimage[1] = storagePrefixB1
	preimage[2] = storagePrefixB2
	preimage[3] = storagePrefixB3
	copy(preimage[4:], rawKey)
	return Interleave(serviceID, preimage)
}

// PreimageMeta builds the StateKey for a preimage meta / lookup-dict (delta4) entry.
// GP eq. (D.2): C(s, E4(l) ⌢ h)
func PreimageMeta(serviceID uint32, hash [32]byte, length uint32) [31]byte {
	preimage := make([]byte, 4+32)
	preimage[0] = byte(length)
	preimage[1] = byte(length >> 8)
	preimage[2] = byte(length >> 16)
	preimage[3] = byte(length >> 24)
	copy(preimage[4:], hash[:])
	return Interleave(serviceID, preimage)
}

// PreimageLookup builds the StateKey for a preimage lookup (delta3) entry.
// GP eq. (D.2): C(s, E4(2^32 - 2) ⌢ h)
func PreimageLookup(serviceID uint32, hash [32]byte) [31]byte {
	preimage := make([]byte, 4+32)
	preimage[0] = preimageLookupPrefixB0
	preimage[1] = preimageLookupPrefixB1
	preimage[2] = preimageLookupPrefixB2
	preimage[3] = preimageLookupPrefixB3
	copy(preimage[4:], hash[:])
	return Interleave(serviceID, preimage)
}
