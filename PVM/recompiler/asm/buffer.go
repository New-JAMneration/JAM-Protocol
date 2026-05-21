package asm

import (
	"encoding/binary"
	"fmt"
)

// CodeBuffer accumulates emitted machine code bytes and manages labels
// with forward-reference fixups.
type CodeBuffer struct {
	data   []byte
	labels map[string]int // label name → byte offset in data
	fixups []fixup
}

type fixup struct {
	label  string
	offset int // position in data where the rel value should be written
	size   int // 1 (rel8) or 4 (rel32)
}

func NewCodeBuffer() *CodeBuffer {
	return &CodeBuffer{
		data:   make([]byte, 0, 4096),
		labels: make(map[string]int),
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
	clear(b.labels)
	b.fixups = b.fixups[:0]
}

// BindLabel records the current buffer offset for the named label.
func (b *CodeBuffer) BindLabel(name string) error {
	if _, exists := b.labels[name]; exists {
		return fmt.Errorf("label %q already bound", name)
	}
	b.labels[name] = len(b.data)
	return nil
}

// UseLabel32 emits a 4-byte placeholder at the current position for a
// PC-relative reference to the named label, to be resolved later.
func (b *CodeBuffer) UseLabel32(name string) {
	b.fixups = append(b.fixups, fixup{
		label:  name,
		offset: len(b.data),
		size:   4,
	})
	b.EmitInt32LE(0) // placeholder
}

// ResolveFixups patches all forward/backward label references.
// PC-relative encoding: target - (fixupPos + fixupSize).
func (b *CodeBuffer) ResolveFixups() error {
	for _, f := range b.fixups {
		target, ok := b.labels[f.label]
		if !ok {
			return fmt.Errorf("unresolved label %q", f.label)
		}
		rel := target - (f.offset + f.size)
		switch f.size {
		case 4:
			if rel < -(1<<31) || rel >= (1<<31) {
				return fmt.Errorf("label %q: rel32 overflow (%d)", f.label, rel)
			}
			binary.LittleEndian.PutUint32(b.data[f.offset:], uint32(int32(rel)))
		case 1:
			if rel < -128 || rel > 127 {
				return fmt.Errorf("label %q: rel8 overflow (%d)", f.label, rel)
			}
			b.data[f.offset] = byte(int8(rel))
		default:
			return fmt.Errorf("label %q: unsupported fixup size %d", f.label, f.size)
		}
	}
	return nil
}

// PatchInt32At overwrites 4 bytes at the given offset (for manual patching).
func (b *CodeBuffer) PatchInt32At(offset int, v int32) {
	binary.LittleEndian.PutUint32(b.data[offset:], uint32(v))
}
