package types

import (
	"bytes"
	"fmt"
	"math/bits"
	"sync"
)

type Encoder struct {
	buf            *bytes.Buffer
	HashSegmentMap HashSegmentMap
}

// encoderPool is a sync.Pool that caches Encoder instances to reduce
// memory allocations. Since Encoder.Encode already calls buf.Reset(),
// pooled encoders are safe to reuse without additional cleanup.
var encoderPool = sync.Pool{
	New: func() any {
		return &Encoder{
			buf:            new(bytes.Buffer),
			HashSegmentMap: nil,
		}
	},
}

// NewEncoder creates a new Encoder instance.
// Deprecated: prefer GetEncoder/PutEncoder for pooled allocation,
// or use HashEncode for the common encode-then-hash pattern.
func NewEncoder() *Encoder {
	cLog(Cyan, "Creating new encoder")
	return &Encoder{
		buf:            new(bytes.Buffer),
		HashSegmentMap: nil,
	}
}

// GetEncoder retrieves a pooled Encoder instance. Callers must return
// the encoder via PutEncoder when done to avoid leaking pool objects.
func GetEncoder() *Encoder {
	return encoderPool.Get().(*Encoder)
}

// PutEncoder returns an Encoder to the pool for reuse. The encoder's
// HashSegmentMap is cleared before pooling to avoid retaining references.
func PutEncoder(e *Encoder) {
	e.HashSegmentMap = nil
	encoderPool.Put(e)
}

func (e *Encoder) Encode(v interface{}) ([]byte, error) {
	cLog(Cyan, "Encoding")
	e.buf.Reset()

	if err := e.encodeStruct(v); err != nil {
		return nil, err
	}

	return bytes.Clone(e.buf.Bytes()), nil
}

func (e *Encoder) EncodeMany(vs ...any) ([]byte, error) {
	cLog(Cyan, "Encoding")
	e.buf.Reset()

	for _, v := range vs {
		if err := e.encodeStruct(v); err != nil {
			return nil, err
		}
	}

	return bytes.Clone(e.buf.Bytes()), nil
}

type Encodable interface {
	Encode(e *Encoder) error
}

func (e *Encoder) encodeStruct(v interface{}) error {
	if encodable, ok := v.(Encodable); ok {
		return encodable.Encode(e)
	}

	return fmt.Errorf("type %T does not implement Encodable", v)
}

// EncodeUintWithLength
func (e *Encoder) EncodeUintWithLength(value uint64, l int) ([]byte, error) {
	if l == 0 {
		return []byte{}, nil
	}

	out := make([]byte, l)
	for i := 0; i < l; i++ {
		out[i] = byte(value & 0xFF)
		value >>= 8
	}

	return out, nil
}

// EncodeUint
func (e *Encoder) EncodeUint(value uint64) ([]byte, error) {
	// If x = 0: E(x) = [0]
	if value == 0 {
		return []byte{0}, nil
	}

	if value < 0x80 {
		return []byte{byte(value)}, nil
	}

	if value >= (uint64(1) << 56) {
		remainderBytes, err := e.EncodeUintWithLength(value, 8)
		if err != nil {
			return nil, err
		}
		return append([]byte{0xFF}, remainderBytes...), nil
	}

	k := bits.Len64(value) - 1
	l := k / 7
	l64 := uint(l)

	power8l := uint64(1) << (8 * l64)
	remainder := value % power8l
	floor := value / power8l

	// prefix = 2^8 - 2^(8-l) + floor(x / 2^(8*l))
	prefix := byte((256 - (1 << (8 - l64))) + floor)

	remainderBytes, err := e.EncodeUintWithLength(remainder, int(l))
	if err != nil {
		return nil, err
	}

	return append([]byte{prefix}, remainderBytes...), nil
}

func (e *Encoder) EncodeInteger(value uint64) error {
	cLog(Cyan, "Encoding Integer")
	encoded, err := e.EncodeUint(value)
	if err != nil {
		return err
	}

	_, err = e.buf.Write(encoded)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("Encoded Integer: %v", encoded))

	return nil
}

func (e *Encoder) EncodeLength(length uint64) error {
	cLog(Cyan, "Encoding Length")
	encodedLength, err := e.EncodeUint(length)
	if err != nil {
		return err
	}

	_, err = e.buf.Write(encodedLength)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("Length: %v", encodedLength))

	return nil
}

// Write a byte
func (e *Encoder) WriteByte(b byte) error {
	cLog(Cyan, "Writing Byte")
	return e.buf.WriteByte(b)
}

func (e *Encoder) SetHashSegmentMap(hashSegmentMap HashSegmentMap) {
	e.HashSegmentMap = hashSegmentMap
}
