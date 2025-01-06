package hash

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

// Blake2bHash generates a hash using the BLAKE2b-256 algorithm.
// The output is a 32-byte hash. (jam_types.OpaqueHash)
func Blake2bHash(input types.ByteSequence) types.OpaqueHash {
	hash := blake2b.Sum256(input)
	return types.OpaqueHash(hash[:])
}

// Blake2bHashPartial generates a hash using the BLAKE2b-256 algorithm for certain x octects.
// The output is a x-byte hash. (jam_types.ByteSequence)
func Blake2bHashPartial(input types.ByteSequence, x uint) types.ByteSequence {
	hash := blake2b.Sum256(input)
	return types.ByteSequence(hash[:x])
}

// KeccakHash generates a hash using the Keccak-256 algorithm.
// The output is a 32-byte hash. (jam_types.OpaqueHash)
func KeccakHash(input types.ByteSequence) types.OpaqueHash {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(input)

	res := hash.Sum(nil)

	return types.OpaqueHash(res[:])
}
