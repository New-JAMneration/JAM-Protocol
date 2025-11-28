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
