package telemetry

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Fixed-width integer roundtrips
// ---------------------------------------------------------------------------

func TestEncodeU8_Roundtrip(t *testing.T) {
	for _, v := range []uint8{0, 1, 0x7F, 0x80, 0xFF} {
		got := EncodeU8(v)
		if len(got) != 1 || got[0] != v {
			t.Fatalf("EncodeU8(%d): got %v", v, got)
		}
		dec, err := NewDecoder(got).ReadU8()
		if err != nil || dec != v {
			t.Fatalf("ReadU8(%v): %d, %v", got, dec, err)
		}
	}
}

func TestEncodeU16_Roundtrip(t *testing.T) {
	for _, v := range []uint16{0, 1, 0xFF, 0x100, 0x1234, 0xFFFF} {
		got := EncodeU16(v)
		if len(got) != 2 {
			t.Fatalf("EncodeU16(%d): want 2 bytes, got %v", v, got)
		}
		// LE check
		if got[0] != byte(v) || got[1] != byte(v>>8) {
			t.Fatalf("EncodeU16(%d): expected LE, got %v", v, got)
		}
		dec, err := NewDecoder(got).ReadU16()
		if err != nil || dec != v {
			t.Fatalf("ReadU16: %d, %v", dec, err)
		}
	}
}

func TestEncodeU32_Roundtrip(t *testing.T) {
	for _, v := range []uint32{0, 1, 0xFFFF, 0x10000, 0x12345678, 0xFFFFFFFF} {
		got := EncodeU32(v)
		if len(got) != 4 {
			t.Fatalf("EncodeU32(%d): want 4 bytes, got %v", v, got)
		}
		dec, err := NewDecoder(got).ReadU32()
		if err != nil || dec != v {
			t.Fatalf("ReadU32: %d, %v", dec, err)
		}
	}
}

func TestEncodeU64_Roundtrip(t *testing.T) {
	for _, v := range []uint64{0, 1, 0xFFFFFFFF, 0x100000000, 0x123456789ABCDEF0, ^uint64(0)} {
		got := EncodeU64(v)
		if len(got) != 8 {
			t.Fatalf("EncodeU64(%d): want 8 bytes, got %v", v, got)
		}
		dec, err := NewDecoder(got).ReadU64()
		if err != nil || dec != v {
			t.Fatalf("ReadU64: %d, %v", dec, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Bool
// ---------------------------------------------------------------------------

func TestEncodeBool_Roundtrip(t *testing.T) {
	for _, v := range []bool{false, true} {
		got := EncodeBool(v)
		if len(got) != 1 {
			t.Fatalf("EncodeBool(%v): %v", v, got)
		}
		dec, err := NewDecoder(got).ReadBool()
		if err != nil || dec != v {
			t.Fatalf("ReadBool: %v, %v", dec, err)
		}
	}
}

func TestReadBool_Invalid(t *testing.T) {
	for _, b := range []byte{2, 0x7F, 0xFF} {
		if _, err := NewDecoder([]byte{b}).ReadBool(); err == nil {
			t.Fatalf("ReadBool(0x%02x): expected error, got nil", b)
		}
	}
}

// ---------------------------------------------------------------------------
// Natural number (variable-length)
//
// The encoding crosses 8 length classes: l=0 (just the prefix byte 0..127),
// l=1, l=2, ..., l=7, then the 0xFF + 8-byte fallback for v >= 2^56.
// ---------------------------------------------------------------------------

func TestEncodeNatural_Roundtrip(t *testing.T) {
	cases := []uint64{
		0, 1, 0x7F, // l=0 boundary
		0x80, 0x3FFF, // l=1 range
		0x4000, 0x1FFFFF, // l=2 range
		0x200000, 0x0FFFFFFF, // l=3 range
		0x10000000, 0x07FFFFFFFF, // l=4 range
		0x0800000000, 0x03FFFFFFFFFF, // l=5 range
		0x040000000000, 0x01FFFFFFFFFFFF, // l=6 range
		0x02000000000000, 0x00FFFFFFFFFFFFFF, // l=7 range
		0x0100000000000000, ^uint64(0), // 0xFF fallback
	}
	for _, v := range cases {
		enc := EncodeNatural(v)
		dec, err := NewDecoder(enc).ReadNatural()
		if err != nil {
			t.Fatalf("ReadNatural(%x): %v (encoded as %x)", v, err, enc)
		}
		if dec != v {
			t.Fatalf("EncodeNatural(%x) -> %x -> ReadNatural %x", v, enc, dec)
		}
	}
}

func TestEncodeNatural_LengthClasses(t *testing.T) {
	// Check the encoded byte count for each length class (l = leading zeros
	// to add after the prefix byte).
	tests := []struct {
		v     uint64
		bytes int
	}{
		{0, 1},                  // l=0
		{0x7F, 1},               // l=0 max
		{0x80, 2},               // l=1 min
		{0x3FFF, 2},             // l=1 max
		{0x4000, 3},             // l=2 min
		{0x0100000000000000, 9}, // 0xFF fallback (1 + 8)
		{^uint64(0), 9},         // 0xFF fallback max
	}
	for _, tc := range tests {
		got := EncodeNatural(tc.v)
		if len(got) != tc.bytes {
			t.Errorf("EncodeNatural(%x): want %d bytes, got %d (%x)",
				tc.v, tc.bytes, len(got), got)
		}
	}
}

// ---------------------------------------------------------------------------
// Bytes (len++[u8])
// ---------------------------------------------------------------------------

func TestEncodeBytes_Roundtrip(t *testing.T) {
	cases := [][]byte{
		nil,
		{},
		{0x00},
		{0x01, 0x02, 0x03},
		bytes.Repeat([]byte{0xAB}, 200),
		bytes.Repeat([]byte{0x55}, 20000),
	}
	for _, in := range cases {
		enc := EncodeBytes(in)
		dec, err := NewDecoder(enc).ReadBytes()
		if err != nil {
			t.Fatalf("ReadBytes len=%d: %v", len(in), err)
		}
		if !bytes.Equal(dec, in) {
			t.Fatalf("ReadBytes mismatch len=%d", len(in))
		}
	}
}

// ---------------------------------------------------------------------------
// String<N>
// ---------------------------------------------------------------------------

func TestEncodeString_Roundtrip(t *testing.T) {
	cases := []string{
		"",
		"hello",
		"日本語", // multi-byte UTF-8
		"🚀💥🌈", // 4-byte UTF-8
		strings.Repeat("a", 100),
	}
	for _, s := range cases {
		enc, err := EncodeString(s, 0) // no cap
		if err != nil {
			t.Fatalf("EncodeString(%q): %v", s, err)
		}
		dec, err := NewDecoder(enc).ReadString(0)
		if err != nil {
			t.Fatalf("ReadString(%q): %v", s, err)
		}
		if dec != s {
			t.Fatalf("ReadString: got %q want %q", dec, s)
		}
	}
}

func TestEncodeString_LengthCap(t *testing.T) {
	if _, err := EncodeString(strings.Repeat("x", 33), 32); err == nil {
		t.Fatal("expected length-cap error for 33 chars in String<32>")
	}
	if _, err := EncodeString("ok", 32); err != nil {
		t.Fatalf("unexpected error for 2 chars in String<32>: %v", err)
	}
}

func TestEncodeString_InvalidUTF8(t *testing.T) {
	// Build raw bytes that are not valid UTF-8.
	bad := string([]byte{0xFF, 0xFE, 0xFD})
	if _, err := EncodeString(bad, 0); err == nil {
		t.Fatal("expected UTF-8 error for invalid bytes")
	}
}

func TestReadString_DecodeLengthCap(t *testing.T) {
	enc, err := EncodeString(strings.Repeat("x", 50), 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewDecoder(enc).ReadString(32); err == nil {
		t.Fatal("expected decode length-cap error")
	}
}

func TestReadString_DecodeInvalidUTF8(t *testing.T) {
	// Build valid len++ wrapping invalid UTF-8 bytes.
	bad := []byte{0xFF, 0xFE, 0xFD}
	enc := EncodeBytes(bad)
	if _, err := NewDecoder(enc).ReadString(0); err == nil {
		t.Fatal("expected decode UTF-8 error")
	}
}

func TestEncodeReason_LengthCap(t *testing.T) {
	if _, err := EncodeReason(strings.Repeat("y", int(ReasonMaxLen)+1)); err == nil {
		t.Fatal("expected Reason length-cap error")
	}
	if _, err := EncodeReason(strings.Repeat("y", int(ReasonMaxLen))); err != nil {
		t.Fatalf("Reason at max length should be ok: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Option
// ---------------------------------------------------------------------------

func TestEncodeOptionRaw_None(t *testing.T) {
	enc := EncodeOptionRaw(nil, false)
	if !bytes.Equal(enc, []byte{0}) {
		t.Fatalf("None: got %v", enc)
	}
	d := NewDecoder(enc)
	present, err := d.ReadOptionPresent()
	if err != nil || present {
		t.Fatalf("ReadOptionPresent: present=%v err=%v", present, err)
	}
	if !d.Done() {
		t.Fatalf("Decoder should be done, %d remaining", d.Remaining())
	}
}

func TestEncodeOptionRaw_Some(t *testing.T) {
	inner := EncodeU32(0xDEADBEEF)
	enc := EncodeOptionRaw(inner, true)
	if enc[0] != 1 {
		t.Fatalf("Some: got prefix %x", enc[0])
	}
	d := NewDecoder(enc)
	present, err := d.ReadOptionPresent()
	if err != nil || !present {
		t.Fatalf("ReadOptionPresent: present=%v err=%v", present, err)
	}
	v, err := d.ReadU32()
	if err != nil || v != 0xDEADBEEF {
		t.Fatalf("ReadU32 inner: %x %v", v, err)
	}
}

// ---------------------------------------------------------------------------
// Decoder boundary errors
// ---------------------------------------------------------------------------

func TestDecoder_ShortReads(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*Decoder) error
		buf  []byte
	}{
		{"u8 empty", func(d *Decoder) error { _, err := d.ReadU8(); return err }, nil},
		{"u16 short", func(d *Decoder) error { _, err := d.ReadU16(); return err }, []byte{0x01}},
		{"u32 short", func(d *Decoder) error { _, err := d.ReadU32(); return err }, []byte{0x01, 0x02, 0x03}},
		{"u64 short", func(d *Decoder) error { _, err := d.ReadU64(); return err }, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}},
		{"natural short remainder", func(d *Decoder) error { _, err := d.ReadNatural(); return err }, []byte{0x80}}, // l=1 prefix without remainder
		{"bytesN short", func(d *Decoder) error { _, err := d.ReadBytesN(5); return err }, []byte{0x01, 0x02}},
	}
	for _, tc := range tests {
		err := tc.fn(NewDecoder(tc.buf))
		if err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
			continue
		}
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Logf("%s: error = %v (not ErrUnexpectedEOF, but acceptable if message clearly indicates short read)", tc.name, err)
		}
	}
}

func TestDecoder_PosTracking(t *testing.T) {
	enc := append(EncodeU8(0xAB), EncodeU32(0xDEADBEEF)...)
	enc = append(enc, EncodeBool(true)...)
	d := NewDecoder(enc)

	if d.Pos() != 0 {
		t.Fatalf("initial pos: %d", d.Pos())
	}
	if _, err := d.ReadU8(); err != nil {
		t.Fatal(err)
	}
	if d.Pos() != 1 {
		t.Fatalf("after u8 pos: %d", d.Pos())
	}
	if _, err := d.ReadU32(); err != nil {
		t.Fatal(err)
	}
	if d.Pos() != 5 {
		t.Fatalf("after u32 pos: %d", d.Pos())
	}
	if _, err := d.ReadBool(); err != nil {
		t.Fatal(err)
	}
	if !d.Done() {
		t.Fatalf("expected Done, %d remaining", d.Remaining())
	}
}

// ---------------------------------------------------------------------------
// Disabled client basic invariants
// ---------------------------------------------------------------------------

func TestDisabledClient_AllReturnInvalidID(t *testing.T) {
	c := NewDisabled()
	if c.Enabled() {
		t.Fatal("disabled client should report Enabled() == false")
	}
	if id := c.Emit(42, []byte{1}); id != InvalidID {
		t.Fatalf("Emit on disabled: got %d, want InvalidID", id)
	}
	if id := c.EmitLazy(42, func() []byte { t.Fatal("builder must not be called"); return nil }); id != InvalidID {
		t.Fatalf("EmitLazy on disabled: got %d, want InvalidID", id)
	}
	if id := c.EmitFollowup(42, 7, []byte{1}); id != InvalidID {
		t.Fatalf("EmitFollowup on disabled: got %d, want InvalidID", id)
	}
	if id := c.EmitFollowupLazy(42, 7, func() []byte { t.Fatal("builder must not be called"); return nil }); id != InvalidID {
		t.Fatalf("EmitFollowupLazy on disabled: got %d, want InvalidID", id)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close on disabled: %v", err)
	}
	// Idempotent
	if err := c.Close(); err != nil {
		t.Fatalf("second Close on disabled: %v", err)
	}
}
