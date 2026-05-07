package telemetry

import "fmt"

// BlockOutline is the fixed-size block summary used by JIP-3 events 42
// (Authored), 43 (Importing), and 68 (Block sent / received).
//
// Wire layout (60 bytes, all fixed):
//
//	u32 LE  Size            (block size in bytes)
//	[32]B   HeaderHash
//	u32 LE  Tickets
//	u32 LE  Preimages
//	u32 LE  PreimagesBytes  (total size of preimages in bytes)
//	u32 LE  Guarantees
//	u32 LE  Assurances
//	u32 LE  DisputeVerdicts
type BlockOutline struct {
	Size            uint32
	HeaderHash      [32]byte
	Tickets         uint32
	Preimages       uint32
	PreimagesBytes  uint32
	Guarantees      uint32
	Assurances      uint32
	DisputeVerdicts uint32
}

// blockOutlineEncodedSize is the fixed wire size of an encoded BlockOutline.
const blockOutlineEncodedSize = 4 + 32 + 4*6

// Encode produces the wire bytes for o. Length is always
// blockOutlineEncodedSize.
func (o BlockOutline) Encode() []byte {
	out := make([]byte, 0, blockOutlineEncodedSize)
	out = append(out, EncodeU32(o.Size)...)
	out = append(out, o.HeaderHash[:]...)
	out = append(out, EncodeU32(o.Tickets)...)
	out = append(out, EncodeU32(o.Preimages)...)
	out = append(out, EncodeU32(o.PreimagesBytes)...)
	out = append(out, EncodeU32(o.Guarantees)...)
	out = append(out, EncodeU32(o.Assurances)...)
	out = append(out, EncodeU32(o.DisputeVerdicts)...)
	return out
}

// ReadBlockOutline reads a BlockOutline from the decoder.
func (d *Decoder) ReadBlockOutline() (BlockOutline, error) {
	var o BlockOutline
	var err error
	if o.Size, err = d.ReadU32(); err != nil {
		return BlockOutline{}, fmt.Errorf("BlockOutline.Size: %w", err)
	}
	hash, err := d.ReadBytesN(32)
	if err != nil {
		return BlockOutline{}, fmt.Errorf("BlockOutline.HeaderHash: %w", err)
	}
	copy(o.HeaderHash[:], hash)
	if o.Tickets, err = d.ReadU32(); err != nil {
		return BlockOutline{}, fmt.Errorf("BlockOutline.Tickets: %w", err)
	}
	if o.Preimages, err = d.ReadU32(); err != nil {
		return BlockOutline{}, fmt.Errorf("BlockOutline.Preimages: %w", err)
	}
	if o.PreimagesBytes, err = d.ReadU32(); err != nil {
		return BlockOutline{}, fmt.Errorf("BlockOutline.PreimagesBytes: %w", err)
	}
	if o.Guarantees, err = d.ReadU32(); err != nil {
		return BlockOutline{}, fmt.Errorf("BlockOutline.Guarantees: %w", err)
	}
	if o.Assurances, err = d.ReadU32(); err != nil {
		return BlockOutline{}, fmt.Errorf("BlockOutline.Assurances: %w", err)
	}
	if o.DisputeVerdicts, err = d.ReadU32(); err != nil {
		return BlockOutline{}, fmt.Errorf("BlockOutline.DisputeVerdicts: %w", err)
	}
	return o, nil
}
