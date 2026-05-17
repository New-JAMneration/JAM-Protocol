package up

import (
	"encoding/binary"
	"fmt"
	"math/bits"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// StreamKindUP0 is JAMNP-S UP 0 (block announcement).
const StreamKindUP0 byte = 0

const finalWireSize = 32 + 4 // HeaderHash ++ Slot (u32 LE)

// BlockRef is a header hash and slot (Final or Leaf on the wire).
type BlockRef struct {
	Hash types.HeaderHash
	Slot types.TimeSlot
}

// Handshake is Final ++ len++[Leaf] per JAMNP-S UP 0.
type Handshake struct {
	Final  BlockRef
	Leaves []BlockRef
}

// Announcement is Header ++ Final per JAMNP-S UP 0.
type Announcement struct {
	Header types.Header
	Final  BlockRef
}

// EncodeHandshake encodes a UP 0 handshake message (not including JAMNP message framing).
func EncodeHandshake(h Handshake) ([]byte, error) {
	enc := types.NewEncoder()
	out := make([]byte, 0, finalWireSize+8+len(h.Leaves)*finalWireSize)
	out = append(out, encodeBlockRef(h.Final)...)

	lenPrefix, err := enc.EncodeUint(uint64(len(h.Leaves)))
	if err != nil {
		return nil, fmt.Errorf("encode leaf count: %w", err)
	}
	out = append(out, lenPrefix...)
	for _, leaf := range h.Leaves {
		out = append(out, encodeBlockRef(leaf)...)
	}
	return out, nil
}

// DecodeHandshake decodes a UP 0 handshake message.
func DecodeHandshake(data []byte) (Handshake, error) {
	if len(data) < finalWireSize {
		return Handshake{}, fmt.Errorf("handshake too short: %d", len(data))
	}
	var h Handshake
	off := 0
	h.Final, off = decodeBlockRef(data, off)

	count, consumed, err := decodeCompactUint(data[off:])
	if err != nil {
		return Handshake{}, fmt.Errorf("decode leaf count: %w", err)
	}
	off += consumed

	if count > uint64(len(data)-off)/finalWireSize+1 {
		return Handshake{}, fmt.Errorf("invalid leaf count %d", count)
	}
	h.Leaves = make([]BlockRef, 0, count)
	for i := uint64(0); i < count; i++ {
		if off+finalWireSize > len(data) {
			return Handshake{}, fmt.Errorf("truncated leaf %d", i)
		}
		var leaf BlockRef
		leaf, off = decodeBlockRef(data, off)
		h.Leaves = append(h.Leaves, leaf)
	}
	if off != len(data) {
		return Handshake{}, fmt.Errorf("trailing handshake bytes: %d", len(data)-off)
	}
	return h, nil
}

// EncodeAnnouncement encodes a UP 0 announcement message.
func EncodeAnnouncement(a Announcement) ([]byte, error) {
	enc := types.NewEncoder()
	headerBytes, err := enc.Encode(&a.Header)
	if err != nil {
		return nil, fmt.Errorf("encode header: %w", err)
	}
	out := make([]byte, 0, len(headerBytes)+finalWireSize)
	out = append(out, headerBytes...)
	out = append(out, encodeBlockRef(a.Final)...)
	return out, nil
}

// DecodeAnnouncement decodes a UP 0 announcement message.
func DecodeAnnouncement(data []byte) (Announcement, error) {
	if len(data) < finalWireSize {
		return Announcement{}, fmt.Errorf("announcement too short: %d", len(data))
	}
	var a Announcement
	dec := types.NewDecoder()
	headerEnd := len(data) - finalWireSize
	if err := dec.Decode(data[:headerEnd], &a.Header); err != nil {
		return Announcement{}, fmt.Errorf("decode header: %w", err)
	}
	_, _ = decodeBlockRef(data, headerEnd)
	a.Final, _ = decodeBlockRef(data, headerEnd)
	return a, nil
}

func encodeBlockRef(r BlockRef) []byte {
	out := make([]byte, finalWireSize)
	copy(out[:32], r.Hash[:])
	binary.LittleEndian.PutUint32(out[32:], uint32(r.Slot))
	return out
}

func decodeBlockRef(data []byte, off int) (BlockRef, int) {
	var r BlockRef
	copy(r.Hash[:], data[off:off+32])
	r.Slot = types.TimeSlot(binary.LittleEndian.Uint32(data[off+32:]))
	return r, off + finalWireSize
}

func decodeCompactUint(data []byte) (uint64, int, error) {
	if len(data) < 1 {
		return 0, 0, fmt.Errorf("no data for compact uint")
	}
	l := bits.LeadingZeros8(^data[0])
	needed := l + 1
	if len(data) < needed {
		return 0, 0, fmt.Errorf("insufficient compact uint")
	}
	v, err := types.NewDecoder().DecodeUint(data[:needed])
	return v, needed, err
}
