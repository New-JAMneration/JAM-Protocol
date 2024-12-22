package utilities

import (
	"errors"
	"fmt"

	jamtypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// DeserializeFixedLength corresponds to the inverse of SerializeFixedLength
// It deserializes a fixed-length byte sequence back to an integer
func DeserializeFixedLength[T jamtypes.U32 | jamtypes.U64](data jamtypes.ByteSequence, l T) (T, error) {
	if l == T(0) {
		return T(0), nil
	}

	if len(data) != int(l) {
		return T(0), fmt.Errorf("invalid data length: expected %d, got %d", l, len(data))
	}

	var result T
	for i := T(0); i < l; i++ {
		result |= T(data[i]) << (8 * i)
	}

	return result, nil
}

// DeserializeU64 corresponds to the inverse of SerializeU64
// It deserializes a variable-length encoded integer back to a U64
func DeserializeU64(data jamtypes.ByteSequence) (jamtypes.U64, error) {
	if len(data) == 0 {
		return 0, errors.New("no data to deserialize U64")
	}
	prefix := data[0]
	data = data[1:]

	// If x = 0: E(x) = [0]
	if prefix == 0 {
		return 0, nil
	}

	// If prefix = 0xFF: E(x) = [255] || E_8(x)
	if prefix == 0xFF {
		if len(data) < 8 {
			return 0, errors.New("not enough data for 8-byte U64")
		}
		var x jamtypes.U64
		for i := 0; i < 8; i++ {
			x |= jamtypes.U64(data[i]) << (8 * i)
		}
		return x, nil
	}

	// Otherwise, attempt to find the correct l by checking each candidate l
	// and verifying that decoded x fits the expected range.
	var bestX jamtypes.U64
	found := false
	for tryL := 1; tryL <= 8; tryL++ {
		base := 256 - (1 << (8 - tryL))
		if int(prefix) >= base {
			// floorVal = prefix - base
			floorVal := jamtypes.U64(int(prefix) - base)

			if len(data) < tryL {
				// Not enough data for remainder, try next l
				continue
			}

			remainderData := data[:tryL]
			var remainder jamtypes.U64
			for i := 0; i < tryL; i++ {
				remainder |= jamtypes.U64(remainderData[i]) << (8 * i)
			}

			// x = floorVal*(2^(8*l)) + remainder
			power8l := jamtypes.U64(1) << (8 * tryL)
			x := floorVal*power8l + remainder

			// Check range to confirm l:
			// 2^(7*l) ≤ x < 2^(7*(l+1))
			lowerBound := jamtypes.U64(1) << (7 * tryL)
			upperBound := jamtypes.U64(1) << (7 * (tryL + 1))

			if x >= lowerBound && x < upperBound {
				bestX = x
				data = data[tryL:] // consume remainder bytes
				found = true
				break
			}
		}
	}

	if !found {
		return 0, errors.New("invalid U64 encoding")
	}

	return bestX, nil
}

func DeserializeU8Wrapper(data jamtypes.ByteSequence) (U8Wrapper, error) {
	val, err := DeserializeU64(data)
	if err != nil {
		return U8Wrapper{}, err
	}
	if val > 0xFF {
		return U8Wrapper{}, errors.New("U8 overflow")
	}
	return U8Wrapper{Value: jamtypes.U8(val)}, nil
}

// U16
func DeserializeU16Wrapper(data jamtypes.ByteSequence) (U16Wrapper, error) {
	val, err := DeserializeU64(data)
	if err != nil {
		return U16Wrapper{}, err
	}
	if val > 0xFFFF {
		return U16Wrapper{}, errors.New("U16 overflow")
	}
	return U16Wrapper{Value: jamtypes.U16(val)}, nil
}

// U32
func DeserializeU32Wrapper(data jamtypes.ByteSequence) (U32Wrapper, error) {
	val, err := DeserializeU64(data)
	if err != nil {
		return U32Wrapper{}, err
	}
	if val > 0xFFFFFFFF {
		return U32Wrapper{}, errors.New("U32 overflow")
	}
	return U32Wrapper{Value: jamtypes.U32(val)}, nil
}

// U64
func DeserializeU64Wrapper(data jamtypes.ByteSequence) (U64Wrapper, error) {
	val, err := DeserializeU64(data)
	if err != nil {
		return U64Wrapper{}, err
	}
	return U64Wrapper{Value: jamtypes.U64(val)}, nil
}
