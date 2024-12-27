package safrole

import (
	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

var entropy OpaqueHash

// OutsideInSequencer re-order the slice of ticketsBodies as in GP Eq. 6.25
func OutsideInSequencer(t *TicketsBodies) []TicketBody {
	left := 0
	right := types.EpochLength - 1

	out := make([]TicketBody, types.EpochLength)
	for i := 0; i < types.EpochLength; i++ {
		if i%2 == 0 {
			out[i] = (*t)[left]
			left++
		} else {
			out[i] = (*t)[right]
			right--
		}
	}

	return out
}

// FallbackKeySequence implements the fallback key sequence in GP Eq. 6.26
func FallbackKeySequence(entropy types.OpaqueHash, validators types.ValidatorsData) []types.BandersnatchPublic {
	// require globa variable entropy
	// type EpochKeys []BandersnatchKey
	keys := make([]types.BandersnatchPublic, types.EpochLength)
	var i types.U32
	var epochLength types.U32 = types.U32(types.EpochLength)

	for i = 0; i < epochLength; i++ {
		// Get E_4(i)
		serial := utils.SerializeFixedLength(i, 4)
		// Concatenate  entropy with E_4(i)
		concatenation := append(entropy[:], serial...)
		// H4 : Keccak256(serializedBytes) -> See section 3.8 , take only the first 4 octets of the hash,
		hash := hash.Blake2bHashPartial(concatenation, 4)
		// E^(-1) deserialization
		validatorIndex, _ := utils.DeserializeFixedLength(types.ByteSequence(hash), types.U64(4))
		// validatorIndex : jamtypes.U64
		validatorIndex %= (types.U64(types.ValidatorsCount))
		// k[]_b : validatorData -> bandersnatch
		keys[i] = validators[validatorIndex].Bandersnatch
	}

	return keys
}
