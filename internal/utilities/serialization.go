package utilities

import (
	jamtypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// SerializeFixedLength corresponds to E_l in the given specification (C.5).
// It serializes a non-negative integer x into exactly l octets in little-endian order.
// If l=0, returns an empty slice.
func SerializeFixedLength(x jamtypes.U64, l int) jamtypes.ByteSequence {
	if l == 0 {
		return []byte{}
	}
	out := make([]byte, l)
	for i := 0; i < l; i++ {
		out[i] = byte(x & 0xFF)
		x >>= 8
	}
	return out
}

// SerializeGeneral corresponds to E in the given specification (C.6).
// It serializes an integer x (0 <= x < 2^64) into a variable number of octets as described.
func Serialize(x jamtypes.U64) jamtypes.ByteSequence {
	// If x = 0: E(x) = [0]
	if x == 0 {
		return []byte{0}
	}

	// Attempt to find l in [1..8] such that 2^(7*l) ≤ x < 2^(7*(l+1))
	for l := 1; l <= 8; l++ {
		l64 := uint(l)
		lowerBound := jamtypes.U64(1) << (7 * l64)       // 2^(7*l)
		upperBound := jamtypes.U64(1) << (7 * (l64 + 1)) // 2^(7*(l+1))
		if x >= lowerBound && x < upperBound {
			// Found suitable l.
			power8l := jamtypes.U64(1) << (8 * l64)
			remainder := x % power8l
			floor := x / power8l

			// prefix = 2^8 - 2^(8-l) + floor(x / 2^(8*l))
			prefix := byte((256 - (1 << (8 - l64))) + floor)

			return append([]byte{prefix}, SerializeFixedLength(remainder, l)...)
		}
	}

	// If no suitable l found:
	// E(x) = [2^8 - 1] || E_8(x) = [255] || SerializeFixedLength(x,8)
	return append([]byte{0xFF}, SerializeFixedLength(x, 8)...)
}
