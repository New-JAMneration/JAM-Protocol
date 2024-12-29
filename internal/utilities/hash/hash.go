package hash

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

// Blake2bHash generates a hash using the BLAKE2b-256 algorithm.
// The output is a 32-byte hash. (jam_types.OpaqueHash)
func Blake2bHash(input jamTypes.ByteSequence) jamTypes.OpaqueHash {
	hash := blake2b.Sum256(input)
	return jamTypes.OpaqueHash(hash[:])
}

// Blake2bHashPartial generates a hash using the BLAKE2b-256 algorithm for certain x octects.
// The output is a x-byte hash. (jam_types.ByteSequence)
func Blake2bHashPartial(input jamTypes.ByteSequence, x uint) jamTypes.ByteSequence {
	hash := blake2b.Sum256(input)
	return jamTypes.ByteSequence(hash[:x])
}

// KeccakHash generates a hash using the Keccak-256 algorithm.
// The output is a 32-byte hash. (jam_types.OpaqueHash)
func KeccakHash(input jamTypes.ByteSequence) jamTypes.OpaqueHash {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(input)

	res := hash.Sum(nil)

	return jamTypes.OpaqueHash(res[:])
}
