//go:build linux && amd64 && cgo

package recompiler

import (
	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// opcode 100: move_reg — Reg[rA] = Reg[rB]
func (c *Compiler) emitMoveReg(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	if dReg != aReg {
		a.MovRegToReg(dReg, aReg)
	}
	return nil
}

// GP 0.8.0: sbrk removed; heap growth is now via grow_heap host call (B.5)

// opcode 101: count_set_bits_64 — Reg[rA] = popcnt(Reg[rB])
func (c *Compiler) emitCountSetBits64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Popcnt(dReg, aReg)
	return nil
}

// opcode 102: count_set_bits_32 — Reg[rA] = popcnt(uint32(Reg[rB]))
func (c *Compiler) emitCountSetBits32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Popcnt32(dReg, aReg)
	return nil
}

// opcode 103: leading_zero_bits_64 — Reg[rA] = lzcnt(Reg[rB])
func (c *Compiler) emitLeadingZeroBits64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Lzcnt(dReg, aReg)
	return nil
}

// opcode 104: leading_zero_bits_32 — Reg[rA] = lzcnt(uint32(Reg[rB]))
func (c *Compiler) emitLeadingZeroBits32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Lzcnt32(dReg, aReg)
	return nil
}

// opcode 105: trailing_zero_bits_64 — Reg[rA] = tzcnt(Reg[rB])
func (c *Compiler) emitTrailingZeroBits64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Tzcnt(dReg, aReg)
	return nil
}

// opcode 106: trailing_zero_bits_32 — Reg[rA] = tzcnt(uint32(Reg[rB]))
func (c *Compiler) emitTrailingZeroBits32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Tzcnt32(dReg, aReg)
	return nil
}

// opcode 107: sign_extend_8 — Reg[rA] = sign_extend_8_to_64(Reg[rB])
func (c *Compiler) emitSignExtend8(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.MovFromByteToRegSx(dReg, aReg)
	return nil
}

// opcode 108: sign_extend_16 — Reg[rA] = sign_extend_16_to_64(Reg[rB])
func (c *Compiler) emitSignExtend16(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.MovFromWordToRegSx(dReg, aReg)
	return nil
}

// opcode 109: zero_extend_16 — Reg[rA] = zero_extend_16_to_64(Reg[rB])
func (c *Compiler) emitZeroExtend16(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.MovFromWordToRegZx(dReg, aReg)
	return nil
}

// opcode 110: reverse_bytes — Reg[rA] = bswap64(Reg[rB])
func (c *Compiler) emitReverseBytes(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	if dReg != aReg {
		a.MovRegToReg(dReg, aReg)
	}
	a.Bswap(dReg)
	return nil
}
