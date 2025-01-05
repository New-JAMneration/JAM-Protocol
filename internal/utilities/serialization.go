package utilities

import (
	"bytes"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Serializable is the interface that all types that can be serialized must implement.
type Serializable interface {
	Serialize() types.ByteSequence
}

type Comparable interface {
	Less(other interface{}) bool
}

// Empty represents E(∅) = []
type Empty struct{}

func (e Empty) Serialize() types.ByteSequence {
	return types.ByteSequence{}
}

func (e Empty) Less(other interface{}) bool {
	return true
}

// StringOctets treats a string as an octet sequence
type StringOctets string

func (s StringOctets) Serialize() types.ByteSequence {
	return types.ByteSequence(s)
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

func (t SerializableSequence) Serialize() types.ByteSequence {
	var result types.ByteSequence
	for _, elem := range t {
		result = append(result, elem.Serialize()...)
	}
	return result
}

// Here we define wrapper types that hold a jam_types value and implement Serializable.
// For example, a U64Wrapper that holds a jam_types.U64 and provides Serialize().

type U8Wrapper struct {
	Value types.U8
}

type U16Wrapper struct {
	Value types.U16
}

type U32Wrapper struct {
	Value types.U32
}

type U64Wrapper struct {
	Value types.U64
}

type ByteSequenceWrapper struct {
	Value types.ByteSequence
}

type ByteArray32Wrapper struct {
	Value types.ByteArray32
}

// SerializeFixedLength corresponds to E_l in the given specification (C.5).
// It serializes a non-negative integer x into exactly l octets in little-endian order.
// If l=0, returns an empty slice.
func SerializeFixedLength[T types.U32 | types.U64](x T, l T) types.ByteSequence {
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
func SerializeU64(x types.U64) types.ByteSequence {
	// If x = 0: E(x) = [0]
	if x == 0 {
		return []byte{0}
	}

	// Attempt to find l in [1..8] such that 2^(7*l) ≤ x < 2^(7*(l+1))
	for l := 0; l <= 7; l++ {
		l64 := uint(l)
		lowerBound := types.U64(1) << (7 * l64)       // 2^(7*l)
		upperBound := types.U64(1) << (7 * (l64 + 1)) // 2^(7*(l+1))
		if x >= lowerBound && x < upperBound {
			// Found suitable l.
			power8l := types.U64(1) << (8 * l64)
			remainder := x % power8l
			floor := x / power8l

			// prefix = 2^8 - 2^(8-l) + floor(x / 2^(8*l))
			prefix := byte((256 - (1 << (8 - l64))) + floor)

			return append([]byte{prefix}, SerializeFixedLength(remainder, types.U64(l))...)
		}
	}

	// If no suitable l found:
	// E(x) = [2^8 - 1] || E_8(x) = [255] || SerializeFixedLength(x,8)
	return append([]byte{0xFF}, SerializeFixedLength(x, 8)...)
}

func (w U8Wrapper) Serialize() types.ByteSequence {
	return SerializeU64(types.U64(w.Value))
}

func (w U8Wrapper) Less(other interface{}) bool {
	if otherKey, ok := other.(U8Wrapper); ok {
		return w.Value < otherKey.Value
	}
	return false
}

func (w U16Wrapper) Serialize() types.ByteSequence {
	return SerializeU64(types.U64(w.Value))
}

func (w U16Wrapper) Less(other interface{}) bool {
	if otherKey, ok := other.(U16Wrapper); ok {
		return w.Value < otherKey.Value
	}
	return false
}

func (w U32Wrapper) Serialize() types.ByteSequence {
	return SerializeU64(types.U64(w.Value))
}

func (w U32Wrapper) Less(other interface{}) bool {
	if otherKey, ok := other.(U32Wrapper); ok {
		return w.Value < otherKey.Value
	}
	return false
}

func (w U64Wrapper) Serialize() types.ByteSequence {
	return SerializeU64(types.U64(w.Value))
}

func (w U64Wrapper) Less(other interface{}) bool {
	if otherKey, ok := other.(U64Wrapper); ok {
		return w.Value < otherKey.Value
	}
	return false
}

func (w ByteSequenceWrapper) Serialize() types.ByteSequence {
	// E(x∈Y) = x directly
	return w.Value
}

func (w ByteArray32Wrapper) Serialize() types.ByteSequence {
	// Fixed length octet sequence
	return types.ByteSequence(w.Value[:])
}

// helper functions that directly take the jam_types values
// and return a Serializable wrapper. This makes usage easier.
func WrapU8(v types.U8) Serializable {
	return U8Wrapper{Value: v}
}

func WrapU16(v types.U16) Serializable {
	return U16Wrapper{Value: v}
}

func WrapU32(v types.U32) Serializable {
	return U32Wrapper{Value: v}
}

func WrapU64(v types.U64) Serializable {
	return U64Wrapper{Value: v}
}

func WrapByteSequence(v types.ByteSequence) Serializable {
	return ByteSequenceWrapper{v}
}
func WrapByteArray32(v types.ByteArray32) Serializable {
	return ByteArray32Wrapper{v}
}

// New Wrapper types for the new types defined in the JAM protocol
type OpaqueHashWrapper struct {
	Value types.OpaqueHash
}

func (w OpaqueHashWrapper) Serialize() types.ByteSequence {
	// Fixed length octet sequence
	return types.ByteSequence(w.Value[:])
}

func WrapOpaqueHash(v types.OpaqueHash) Serializable {
	return OpaqueHashWrapper{Value: v}
}

// ---------------------------------------------
// C.1.5 Bit Sequence Encoding
// E(b∈B): pack bits into bytes (LSB-first). If variable length, prefix bit length.
// BitSequence represents a sequence of bits
type BitSequenceWrapper struct {
	Bits             types.BitSequence
	IsVariableLength bool
}

func (b BitSequenceWrapper) Serialize() types.ByteSequence {
	if len(b.Bits) == 0 {
		return types.ByteSequence{}
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
		bitLen := WrapU64(types.U64(len(b.Bits))).Serialize()
		return append(bitLen, res...)
	}

	return types.ByteSequence(res)
}

// C.1.6. Dictionary Encoding.
type MapWarpper struct {
	Value map[Comparable]Serializable
}

func (m *MapWarpper) Serialize() types.ByteSequence {
	// Handle empty dictionary
	if len(m.Value) == 0 {
		return types.ByteSequence{}
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

func (s *SetWarpper) Serialize() types.ByteSequence {
	// Handle empty dictionary
	if len(s.Value) == 0 {
		return types.ByteSequence{}
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
