package utilities

import (
	"bytes"
	"sort"

	jamtypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// Serializable is the interface that all types that can be serialized must implement.
type Serializable interface {
	Serialize() jamtypes.ByteSequence
}

type Comparable interface {
	Less(other interface{}) bool
}

// Empty represents E(∅) = []
type Empty struct{}

func (e Empty) Serialize() jamtypes.ByteSequence {
	return jamtypes.ByteSequence{}
}

func (e Empty) Less(other interface{}) bool {
	return true
}

// StringOctets treats a string as an octet sequence
type StringOctets string

func (s StringOctets) Serialize() jamtypes.ByteSequence {
	return jamtypes.ByteSequence(s)
}
func (s StringOctets) Less(other interface{}) bool {
	if otherKey, ok := other.(StringOctets); ok {
		return s < otherKey
	}
	return false
}

// Tuple (or a sequence) E({a,b,...}) = E(a)||E(b)||...
// We can represent this as a slice of Serializable.
type SerializableSequence []Serializable

func (t SerializableSequence) Serialize() jamtypes.ByteSequence {
	var result jamtypes.ByteSequence
	for _, elem := range t {
		result = append(result, elem.Serialize()...)
	}
	return result
}

// Here we define wrapper types that hold a jam_types value and implement Serializable.
// For example, a U64Wrapper that holds a jam_types.U64 and provides Serialize().

type U8Wrapper struct {
	Value jamtypes.U8
}

type U16Wrapper struct {
	Value jamtypes.U16
}

type U32Wrapper struct {
	Value jamtypes.U32
}

type U64Wrapper struct {
	Value jamtypes.U64
}

type ByteSequenceWrapper struct {
	Value jamtypes.ByteSequence
}

type ByteArray32Wrapper struct {
	Value jamtypes.ByteArray32
}

// SerializeFixedLength corresponds to E_l in the given specification (C.5).
// It serializes a non-negative integer x into exactly l octets in little-endian order.
// If l=0, returns an empty slice.
func SerializeFixedLength[T jamtypes.U32 | jamtypes.U64](x T, l T) jamtypes.ByteSequence {
	if l == 0 {
		return []byte{}
	}
	out := make([]byte, l)
	for i := T(0); i < l; i++ {
		out[i] = byte(x & 0xFF)
		x >>= 8
	}
	return out
}

// SerializeGeneral corresponds to E in the given specification (C.6).
// It serializes an integer x (0 <= x < 2^64) into a variable number of octets as described.
func SerializeU64(x jamtypes.U64) jamtypes.ByteSequence {
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

			return append([]byte{prefix}, SerializeFixedLength(remainder, jamtypes.U64(l))...)
		}
	}

	// If no suitable l found:
	// E(x) = [2^8 - 1] || E_8(x) = [255] || SerializeFixedLength(x,8)
	return append([]byte{0xFF}, SerializeFixedLength(x, 8)...)
}

func (w U8Wrapper) Serialize() jamtypes.ByteSequence {
	return SerializeU64(jamtypes.U64(w.Value))
}

func (w U8Wrapper) Less(other interface{}) bool {
	if otherKey, ok := other.(U8Wrapper); ok {
		return w.Value < otherKey.Value
	}
	return false
}

func (w U16Wrapper) Serialize() jamtypes.ByteSequence {
	return SerializeU64(jamtypes.U64(w.Value))
}

func (w U16Wrapper) Less(other interface{}) bool {
	if otherKey, ok := other.(U16Wrapper); ok {
		return w.Value < otherKey.Value
	}
	return false
}

func (w U32Wrapper) Serialize() jamtypes.ByteSequence {
	return SerializeU64(jamtypes.U64(w.Value))
}

func (w U32Wrapper) Less(other interface{}) bool {
	if otherKey, ok := other.(U32Wrapper); ok {
		return w.Value < otherKey.Value
	}
	return false
}

func (w U64Wrapper) Serialize() jamtypes.ByteSequence {
	return SerializeU64(jamtypes.U64(w.Value))
}

func (w U64Wrapper) Less(other interface{}) bool {
	if otherKey, ok := other.(U64Wrapper); ok {
		return w.Value < otherKey.Value
	}
	return false
}

func (w ByteSequenceWrapper) Serialize() jamtypes.ByteSequence {
	// E(x∈Y) = x directly
	return w.Value
}

func (w ByteArray32Wrapper) Serialize() jamtypes.ByteSequence {
	// Fixed length octet sequence
	return w.Value[:]
}

// helper functions that directly take the jam_types values
// and return a Serializable wrapper. This makes usage easier.
func WrapU8(v jamtypes.U8) Serializable {
	return U8Wrapper{Value: v}
}

func WrapU16(v jamtypes.U16) Serializable {
	return U16Wrapper{Value: v}
}

func WrapU32(v jamtypes.U32) Serializable {
	return U32Wrapper{Value: v}
}

func WrapU64(v jamtypes.U64) Serializable {
	return U64Wrapper{Value: v}
}

func WrapByteSequence(v jamtypes.ByteSequence) Serializable {
	return ByteSequenceWrapper{v}
}
func WrapByteArray32(v jamtypes.ByteArray32) Serializable {
	return ByteArray32Wrapper{v}
}

// ---------------------------------------------
// C.1.5 Bit Sequence Encoding
// E(b∈B): pack bits into bytes (LSB-first). If variable length, prefix bit length.
// BitSequence represents a sequence of bits
type BitSequenceWrapper struct {
	Bits             jamtypes.BitSequence
	IsVariableLength bool
}

func (b BitSequenceWrapper) Serialize() jamtypes.ByteSequence {
	if len(b.Bits) == 0 {
		return jamtypes.ByteSequence{}
	}

	if len(b.Bits) == 0 {
		return []byte{}
	}
	var buf bytes.Buffer
	for i := 0; i < len(b.Bits); i += 8 {
		var octet byte
		chunk := b.Bits[i:]
		limit := 8
		if len(chunk) < limit {
			limit = len(chunk)
		}
		for j := 0; j < limit; j++ {
			if chunk[j] {
				octet |= (1 << j)
			}
		}
		buf.WriteByte(octet)
	}
	res := buf.Bytes()

	// In the case of a variable length sequence, then the length is prefixed as in the general case.
	if b.IsVariableLength {
		bitLen := WrapU64(jamtypes.U64(len(b.Bits))).Serialize()
		return append(bitLen, res...)
	}

	return jamtypes.ByteSequence(res)
}

// C.1.6. Dictionary Encoding.
type MapWarpper struct {
	Value map[Comparable]Serializable
}

func (m *MapWarpper) Serialize() jamtypes.ByteSequence {
	// Handle empty dictionary
	if len(m.Value) == 0 {
		return jamtypes.ByteSequence{}
	}

	// Extract and sort keys
	keys := make([]Comparable, 0, len(m.Value))
	for k := range m.Value {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Less(keys[j])
	})

	// Put the (key, value) pair into an array
	seq := []Serializable{}
	for _, key := range keys {
		seq = append(seq, SerializableSequence{key.(Serializable), m.Value[key]})
	}

	d := Discriminator{Value: seq}
	return d.Serialize()
}

// C.1.7. Set Encoding.
type SetWarpper struct {
	Value []Comparable
}

func (s *SetWarpper) Serialize() jamtypes.ByteSequence {
	// Handle empty dictionary
	if len(s.Value) == 0 {
		return jamtypes.ByteSequence{}
	}

	sort.Slice(s.Value, func(i, j int) bool {
		return s.Value[i].Less(s.Value[j])
	})

	seq := SerializableSequence{}
	for _, value := range s.Value {
		v, ok := value.(Serializable)
		if ok {
			seq = append(seq, v)
		}
	}
	return seq.Serialize()
}
