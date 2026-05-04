package telemetry

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

// JIP-3 §"Core Type Definitions" common-type codec. JAM-standard:
// fixed-width ints are LE; varlen naturals use JAM's general integer
// encoding (also the length prefix for len++[T] and String<N>); bool is
// 1 byte; Option<T> is 1 byte (0x00 None / 0x01 Some) plus encoded T;
// String<N> is len++[u8], valid UTF-8, len ≤ N.
//
// Encoders return a fresh []byte. Decoders are methods on Decoder,
// which walks the input and advances on each successful read.

// ReasonMaxLen is the upper bound for the JIP-3 Reason type (String<128>).
const ReasonMaxLen uint32 = 128

// ---------------------------------------------------------------------------
// Encoders
// ---------------------------------------------------------------------------

// EncodeU8 encodes v as a single byte.
func EncodeU8(v uint8) []byte { return []byte{v} }

// EncodeU16 encodes v as 2 little-endian bytes.
func EncodeU16(v uint16) []byte {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	return b[:]
}

// EncodeU32 encodes v as 4 little-endian bytes.
func EncodeU32(v uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	return b[:]
}

// EncodeU64 encodes v as 8 little-endian bytes.
func EncodeU64(v uint64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	return b[:]
}

// EncodeBool encodes v as 0x00 (false) or 0x01 (true).
func EncodeBool(v bool) []byte {
	if v {
		return []byte{1}
	}
	return []byte{0}
}

// EncodeNatural encodes v as a JAM variable-length natural number. This is
// the same scheme used by internal/utilities.SerializeU64 and is the length
// prefix used by len++[T] sequences and String<N>.
func EncodeNatural(v uint64) []byte {
	if v == 0 {
		return []byte{0}
	}
	for l := uint(0); l <= 7; l++ {
		lower := uint64(1) << (7 * l)
		upper := uint64(1) << (7 * (l + 1))
		if v >= lower && v < upper {
			power := uint64(1) << (8 * l)
			floor := v / power
			remainder := v % power
			prefix := byte((256 - (1 << (8 - l))) + uint(floor))
			out := make([]byte, 1, 1+l)
			out[0] = prefix
			for i := uint(0); i < l; i++ {
				out = append(out, byte(remainder>>(8*i)))
			}
			return out
		}
	}
	// v >= 2^56: prefix 0xFF + 8-byte LE
	out := make([]byte, 9)
	out[0] = 0xFF
	binary.LittleEndian.PutUint64(out[1:], v)
	return out
}

// EncodeBytes encodes a byte slice with a natural-number length prefix.
// This is the underlying len++[u8] form used by String<N>.
func EncodeBytes(b []byte) []byte {
	out := EncodeNatural(uint64(len(b)))
	return append(out, b...)
}

// EncodeString encodes s as len++[u8] after validating UTF-8 and the
// maximum length. Pass 0 for maxLen to skip the length cap check.
func EncodeString(s string, maxLen uint32) ([]byte, error) {
	if !utf8.ValidString(s) {
		return nil, errors.New("telemetry: string is not valid UTF-8")
	}
	if maxLen != 0 && uint32(len(s)) > maxLen {
		return nil, fmt.Errorf("telemetry: string length %d exceeds max %d", len(s), maxLen)
	}
	return EncodeBytes([]byte(s)), nil
}

// EncodeReason encodes s as a Reason (String<128>).
func EncodeReason(s string) ([]byte, error) {
	return EncodeString(s, ReasonMaxLen)
}

// EncodeOptionRaw encodes the option discriminant followed by the already-
// encoded inner bytes when present. Pass nil inner for None.
//
// Use this when the inner value's encoder is type-specific and producing
// the bytes upfront is convenient. For nested generic options, callers
// can compose this manually.
func EncodeOptionRaw(inner []byte, present bool) []byte {
	if !present {
		return []byte{0}
	}
	out := make([]byte, 1, 1+len(inner))
	out[0] = 1
	return append(out, inner...)
}

// ---------------------------------------------------------------------------
// Decoder
// ---------------------------------------------------------------------------

// Decoder reads JIP-3-encoded values from a byte buffer, advancing its
// internal position on each successful read. Methods return an error when
// the buffer does not contain enough bytes for the requested type or when
// the on-wire encoding is malformed.
//
// The zero value is not usable; call NewDecoder.
type Decoder struct {
	data []byte
	pos  int
}

// NewDecoder returns a Decoder positioned at the start of data.
func NewDecoder(data []byte) *Decoder { return &Decoder{data: data} }

// Pos returns the number of bytes consumed so far.
func (d *Decoder) Pos() int { return d.pos }

// Remaining returns the number of bytes left to read.
func (d *Decoder) Remaining() int { return len(d.data) - d.pos }

// Done reports whether the entire buffer has been consumed.
func (d *Decoder) Done() bool { return d.pos >= len(d.data) }

func (d *Decoder) readByte() (byte, error) {
	if d.pos >= len(d.data) {
		return 0, io.ErrUnexpectedEOF
	}
	b := d.data[d.pos]
	d.pos++
	return b, nil
}

// ReadU8 reads one byte.
func (d *Decoder) ReadU8() (uint8, error) {
	return d.readByte()
}

// ReadU16 reads 2 little-endian bytes.
func (d *Decoder) ReadU16() (uint16, error) {
	if d.Remaining() < 2 {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint16(d.data[d.pos:])
	d.pos += 2
	return v, nil
}

// ReadU32 reads 4 little-endian bytes.
func (d *Decoder) ReadU32() (uint32, error) {
	if d.Remaining() < 4 {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(d.data[d.pos:])
	d.pos += 4
	return v, nil
}

// ReadU64 reads 8 little-endian bytes.
func (d *Decoder) ReadU64() (uint64, error) {
	if d.Remaining() < 8 {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint64(d.data[d.pos:])
	d.pos += 8
	return v, nil
}

// ReadBool reads a 1-byte boolean. Any byte other than 0 or 1 is an error.
func (d *Decoder) ReadBool() (bool, error) {
	b, err := d.readByte()
	if err != nil {
		return false, err
	}
	switch b {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("telemetry: invalid bool byte 0x%02x", b)
	}
}

// ReadNatural reads a JAM variable-length natural number.
//
// Encoding layout, by leading byte:
//
//	0x00 .. 0x7F : l=0, value = prefix (0..127)
//	0x80 .. 0xBF : l=1, 1 remainder byte
//	0xC0 .. 0xDF : l=2, 2 remainder bytes
//	0xE0 .. 0xEF : l=3, 3 remainder bytes
//	0xF0 .. 0xF7 : l=4, 4 remainder bytes
//	0xF8 .. 0xFB : l=5, 5 remainder bytes
//	0xFC .. 0xFD : l=6, 6 remainder bytes
//	0xFE         : l=7, 7 remainder bytes
//	0xFF         : 8-byte LE fallback
func (d *Decoder) ReadNatural() (uint64, error) {
	prefix, err := d.readByte()
	if err != nil {
		return 0, err
	}
	// l = 0 covers prefix 0x00..0x7F (the full value is in the prefix byte).
	if prefix < 0x80 {
		return uint64(prefix), nil
	}
	if prefix == 0xFF {
		return d.ReadU64()
	}
	// l in 1..7. Range size is 1 << (7-l); base is 256 - (1 << (8-l)).
	for l := uint(1); l <= 7; l++ {
		base := 256 - (1 << (8 - l))
		size := 1 << (7 - l)
		if int(prefix) >= base && int(prefix) < base+size {
			floor := uint64(int(prefix) - base)
			if d.Remaining() < int(l) {
				return 0, io.ErrUnexpectedEOF
			}
			var remainder uint64
			for i := uint(0); i < l; i++ {
				remainder |= uint64(d.data[d.pos+int(i)]) << (8 * i)
			}
			d.pos += int(l)
			power := uint64(1) << (8 * l)
			v := floor*power + remainder
			lower := uint64(1) << (7 * l)
			upper := uint64(1) << (7 * (l + 1))
			if v < lower || v >= upper {
				return 0, fmt.Errorf("telemetry: natural value %d out of expected range for l=%d", v, l)
			}
			return v, nil
		}
	}
	return 0, fmt.Errorf("telemetry: invalid natural prefix 0x%02x", prefix)
}

// ReadBytesN reads exactly n raw bytes.
func (d *Decoder) ReadBytesN(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("telemetry: negative read length %d", n)
	}
	if d.Remaining() < n {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]byte, n)
	copy(out, d.data[d.pos:d.pos+n])
	d.pos += n
	return out, nil
}

// ReadBytes reads a len++[u8] sequence.
func (d *Decoder) ReadBytes() ([]byte, error) {
	n, err := d.ReadNatural()
	if err != nil {
		return nil, err
	}
	return d.ReadBytesN(int(n))
}

// ReadString reads a String<N>: len++[u8] with UTF-8 validation and
// optional length cap (pass 0 to skip the cap).
func (d *Decoder) ReadString(maxLen uint32) (string, error) {
	b, err := d.ReadBytes()
	if err != nil {
		return "", err
	}
	if maxLen != 0 && uint32(len(b)) > maxLen {
		return "", fmt.Errorf("telemetry: string length %d exceeds max %d", len(b), maxLen)
	}
	if !utf8.Valid(b) {
		return "", errors.New("telemetry: decoded string is not valid UTF-8")
	}
	return string(b), nil
}

// ReadReason reads a Reason (String<128>).
func (d *Decoder) ReadReason() (string, error) {
	return d.ReadString(ReasonMaxLen)
}

// ReadOptionPresent reads the option discriminant byte and returns true
// when an inner value follows, false for None. Caller is responsible for
// reading the inner value if present.
func (d *Decoder) ReadOptionPresent() (bool, error) {
	b, err := d.readByte()
	if err != nil {
		return false, err
	}
	switch b {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("telemetry: invalid option byte 0x%02x", b)
	}
}
