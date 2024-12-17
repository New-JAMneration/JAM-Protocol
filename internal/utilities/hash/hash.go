package hash

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"golang.org/x/crypto/blake2b"
	keecak "golang.org/x/crypto/sha3"
)

// blake2bHash generates a hash using the BLAKE2b algorithm.
// The output is a 32-byte hash. (jam_types.OpaqueHash)
func Blake2bHash(input []byte) jamTypes.OpaqueHash {
	hash := blake2b.Sum256(input)
	return jamTypes.OpaqueHash(hash[:])
}

// blake2bHashPartial generates a hash using the BLAKE2b algorithm for certain x octects.
// The output is a x-byte hash. (jam_types.ByteSequence)
func Blake2bHashPartial(input []byte, x int) jamTypes.ByteSequence {
	hash := blake2b.Sum256(input)
	return jamTypes.ByteSequence(hash[:x])
}

// KeccakHash generates a hash using the Keccak algorithm.
// The output is a 32-byte hash. (jam_types.OpaqueHash)
func KeccakHash(input []byte) jamTypes.OpaqueHash {
	hash := keecak.NewLegacyKeccak256()
	hash.Write(input)

	res := hash.Sum(nil)

	return jamTypes.OpaqueHash(res[:])
}
