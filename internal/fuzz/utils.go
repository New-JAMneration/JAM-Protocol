package fuzz

// All messages are encoded according to the JAM codec format.
// Prior to transmission, each encoded message is prefixed with its length,
// represented as a 32-bit little-endian integer.
func marshalUint32LE(value uint32) []byte {
	return []byte{
		byte(value),
		byte(value >> 8),
		byte(value >> 16),
		byte(value >> 24),
	}
}

func unmarshalUint32LE(data []byte) uint32 {
	return uint32(data[0]) |
		uint32(data[1])<<8 |
		uint32(data[2])<<16 |
		uint32(data[3])<<24
}

func marshalUint8(value uint8) []byte {
	return []byte{value}
}

func unmarshalUint8(data []byte) uint8 {
	return uint8(data[0])
}

// compactEncode encodes a uint64 value using JAM compact encoding.
// Reference: GP Appendix E - Compact Integer Encoding
func compactEncode(x uint64) []byte {
	if x == 0 {
		return []byte{0}
	}

	// find the smallest l such that 2^(7*l) <= x < 2^(7*(l+1))
	var l int
	found := false
	for i := 0; i < 8; i++ {
		lower := uint64(1) << (7 * i)
		upper := uint64(1) << (7 * (i + 1))
		if lower <= x && x < upper {
			l = i
			found = true
			break
		}
	}

	if found {
		// calculate the first byte: (2^8 - 2^(8-l)) + (x / 2^(8*l))
		prefix := (uint64(1) << 8) - (uint64(1) << (8 - l))
		prefix += x / (uint64(1) << (8 * l))
		firstByte := byte(prefix)

		result := []byte{firstByte}

		// calculate the remaining bytes: x % 2^(8*l)
		rem := x % (uint64(1) << (8 * l))
		for i := 0; i < l; i++ {
			result = append(result, byte(rem>>(8*i)))
		}
		return result
	}

	// if l >= 8, write 0xFF followed by 8 little-endian bytes of x
	result := []byte{0xFF}
	for i := 0; i < 8; i++ {
		result = append(result, byte(x>>(8*i)))
	}
	return result
}

// compactDecode decodes a JAM compact encoded value from data.
// Returns the decoded value and the number of bytes consumed.
func compactDecode(data []byte) (uint64, int) {
	if len(data) == 0 {
		return 0, 0
	}

	b := data[0]

	if b == 0 {
		return 0, 1
	}

	if b == 0xFF {
		// read the next 8 bytes as a little-endian uint64
		if len(data) < 9 {
			return 0, 0 // data too short
		}
		var v uint64
		for i := 0; i < 8; i++ {
			v |= uint64(data[1+i]) << (8 * i)
		}
		return v, 9
	}

	// find the position of the first 0 bit (from the highest bit)
	length := 0
	for i := 0; i < 8; i++ {
		if (b & (0b10000000 >> i)) == 0 {
			length = i
			break
		}
	}

	// calculate the total number of bytes needed
	totalBytes := 1 + length
	if len(data) < totalBytes {
		return 0, 0 // data too short
	}

	// calculate the rem part (the effective bits of the first byte)
	rem := int(b & ((1 << (7 - length)) - 1))

	if length == 0 {
		return uint64(rem), 1
	}

	// read the next bytes
	var v uint64
	for i := 0; i < length; i++ {
		v |= uint64(data[1+i]) << (8 * i)
	}
	v += uint64(rem) << (8 * length)

	return v, totalBytes
}
