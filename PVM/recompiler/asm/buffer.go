package asm

import (
	"encoding/binary"
	"fmt"
)

// Label is an integer handle for a jump target, allocated by NewLabel and
// resolved at Finalize. Handles are only meaningful within one Reset cycle.
type Label int32

const unboundLabel = int32(-1)

// CodeBuffer accumulates emitted machine code bytes and manages labels
// with forward-reference fixups.
type CodeBuffer struct {
	data     []byte
	labelPos []int32 // index = Label; value = byte offset in data; -1 = unbound
	fixups   []fixup
}

type fixup struct {
	label  Label
	offset int32 // position in data where the rel value should be written
	size   int8  // 1 (rel8) or 4 (rel32)
}

func NewCodeBuffer() *CodeBuffer {
	return &CodeBuffer{
		data: make([]byte, 0, 4096),
	}
}

func (b *CodeBuffer) Emit(bytes ...byte) {
	b.data = append(b.data, bytes...)
}

func (b *CodeBuffer) EmitUint32LE(v uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	b.data = append(b.data, buf[:]...)
}

func (b *CodeBuffer) EmitInt32LE(v int32) {
	b.EmitUint32LE(uint32(v))
}

func (b *CodeBuffer) EmitUint64LE(v uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], v)
	b.data = append(b.data, buf[:]...)
}

// Len returns the current number of emitted bytes.
func (b *CodeBuffer) Len() int { return len(b.data) }

// Bytes returns a copy of the emitted machine code.
func (b *CodeBuffer) Bytes() []byte {
	out := make([]byte, len(b.data))
	copy(out, b.data)
	return out
}

func (b *CodeBuffer) Reset() {
	b.data = b.data[:0]
	b.labelPos = b.labelPos[:0]
	b.fixups = b.fixups[:0]
}

// NewLabel allocates a fresh unbound label handle.
func (b *CodeBuffer) NewLabel() Label {
	b.labelPos = append(b.labelPos, unboundLabel)
	return Label(len(b.labelPos) - 1)
}

// BindLabel records the current buffer offset for the label.
func (b *CodeBuffer) BindLabel(l Label) error {
	if int(l) < 0 || int(l) >= len(b.labelPos) {
		return fmt.Errorf("label %d not allocated", l)
	}
	if b.labelPos[l] != unboundLabel {
		return fmt.Errorf("label %d already bound", l)
	}
	b.labelPos[l] = int32(len(b.data))
	return nil
}

// UseLabel32 emits a 4-byte placeholder at the current position for a
// PC-relative reference to the label, to be resolved later.
func (b *CodeBuffer) UseLabel32(l Label) {
	b.fixups = append(b.fixups, fixup{
		label:  l,
		offset: int32(len(b.data)),
		size:   4,
	})
	b.EmitInt32LE(0) // placeholder
}

// ResolveFixups patches all forward/backward label references.
// PC-relative encoding: target - (fixupPos + fixupSize).
func (b *CodeBuffer) ResolveFixups() error {
	for _, f := range b.fixups {
		if int(f.label) < 0 || int(f.label) >= len(b.labelPos) {
			return fmt.Errorf("label %d not allocated", f.label)
		}
		target := b.labelPos[f.label]
		if target == unboundLabel {
			return fmt.Errorf("unresolved label %d", f.label)
		}
		rel := int(target) - int(f.offset) - int(f.size)
		switch f.size {
		case 4:
			if rel < -(1<<31) || rel >= (1<<31) {
				return fmt.Errorf("label %d: rel32 overflow (%d)", f.label, rel)
			}
			binary.LittleEndian.PutUint32(b.data[f.offset:], uint32(int32(rel)))
		case 1:
			if rel < -128 || rel > 127 {
				return fmt.Errorf("label %d: rel8 overflow (%d)", f.label, rel)
			}
			b.data[f.offset] = byte(int8(rel))
		default:
			return fmt.Errorf("label %d: unsupported fixup size %d", f.label, f.size)
		}
	}
	return nil
}

// PatchInt32At overwrites 4 bytes at the given offset (for manual patching).
func (b *CodeBuffer) PatchInt32At(offset int, v int32) {
	binary.LittleEndian.PutUint32(b.data[offset:], uint32(v))
}
