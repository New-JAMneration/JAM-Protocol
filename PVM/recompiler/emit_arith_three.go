//go:build linux && amd64

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// ---- 4.7.1 32-bit arithmetic ----

// opcode 190: add_32 — Reg[rD] = sext_4(uint32(Reg[rA] + Reg[rB]))
func (c *Compiler) emitAdd32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.Add32RegReg(RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	emitSignExt32(a, dReg)
	return nil
}

// opcode 191: sub_32 — Reg[rD] = sext_4(uint32(Reg[rA]) - uint32(Reg[rB]))
func (c *Compiler) emitSub32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.Sub32RegReg(RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	emitSignExt32(a, dReg)
	return nil
}

// opcode 192: mul_32 — Reg[rD] = sext_4(uint32(Reg[rA] * Reg[rB]))
func (c *Compiler) emitMul32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.Imul32RegReg(RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	emitSignExt32(a, dReg)
	return nil
}

// opcode 193: div_u_32 — Reg[rD] = sext_4(uint32(Reg[rA]) /u uint32(Reg[rB])); div0 → 2^64-1
func (c *Compiler) emitDivU32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	dReg, aReg, bReg := threeRegFromMeta(instr)
	divByZero := fmt.Sprintf("divz_%d", pc)
	done := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	emitZeroExt32(a, RegScratch)
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondEQ, divByZero)

	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(asm.RAX, aReg)
	emitZeroExt32(a, asm.RAX)
	a.Xor(asm.RDX)
	a.Div32(RegScratch)
	emitRestoreDiv32Result(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(divByZero)
	a.MovImm64ToReg(dReg, ^uint64(0)) // div0 → 2^64-1 (same as interpreter)
	a.Jmp(done)

	_ = a.BindLabel(done)
	return nil
}

// opcode 194: div_s_32 — Reg[rD] = int32(Reg[rA]) / int32(Reg[rB]); div0 → 2^64-1
func (c *Compiler) emitDivS32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	dReg, aReg, bReg := threeRegFromMeta(instr)
	divByZero := fmt.Sprintf("divz_%d", pc)
	overflow := fmt.Sprintf("ovf_%d", pc)
	doDiv := fmt.Sprintf("do_div_%d", pc)
	done := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondEQ, divByZero)

	a.CmpReg32Imm32(RegScratch, -1)
	a.Jcc(asm.CondEQ, overflow)

	_ = a.BindLabel(doDiv)
	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(asm.RAX, aReg)
	a.MovFromDwordToRegSx(asm.RAX, asm.RAX)
	a.Cqo()
	a.MovFromDwordToRegSx(RegScratch, RegScratch)
	a.Idiv(RegScratch)
	emitRestoreDiv32Result(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(overflow)
	a.MovRegToReg(RegScratch, aReg)
	a.CmpReg32Imm32(RegScratch, -2147483648)
	a.Jcc(asm.CondNE, doDiv)
	a.MovImm32ToReg(dReg, -2147483648)
	emitSignExt32(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(divByZero)
	a.MovImm64ToReg(dReg, ^uint64(0))
	a.Jmp(done)

	_ = a.BindLabel(done)
	return nil
}

// opcode 195: rem_u_32 — Reg[rD] = sext_4(uint32(Reg[rA]) %u uint32(Reg[rB])); div0 → Reg[rA]
func (c *Compiler) emitRemU32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	dReg, aReg, bReg := threeRegFromMeta(instr)
	divByZero := fmt.Sprintf("divz_%d", pc)
	done := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	emitZeroExt32(a, RegScratch)
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondEQ, divByZero)

	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(asm.RAX, aReg)
	emitZeroExt32(a, asm.RAX)
	a.Xor(asm.RDX)
	a.Div32(RegScratch)
	emitRestoreDiv32Remainder(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(divByZero)
	a.MovRegToReg(dReg, aReg)
	emitSignExt32(a, dReg)

	_ = a.BindLabel(done)
	return nil
}

// opcode 196: rem_s_32 — Reg[rD] = smod(int32(Reg[rA]), int32(Reg[rB]))
func (c *Compiler) emitRemS32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	dReg, aReg, bReg := threeRegFromMeta(instr)
	divByZero := fmt.Sprintf("divz_%d", pc)
	divNegOne := fmt.Sprintf("divm1_%d", pc)
	done := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondEQ, divByZero)

	a.CmpReg32Imm32(RegScratch, -1)
	a.Jcc(asm.CondEQ, divNegOne)

	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(asm.RAX, aReg)
	a.MovFromDwordToRegSx(asm.RAX, asm.RAX)
	a.Cqo()
	a.MovFromDwordToRegSx(RegScratch, RegScratch)
	a.Idiv(RegScratch)
	emitRestoreDiv32Remainder(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(divNegOne)
	a.MovImm32ToReg(dReg, 0)
	emitSignExt32(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(divByZero)
	a.MovRegToReg(dReg, aReg)
	emitSignExt32(a, dReg)

	_ = a.BindLabel(done)
	return nil
}

// opcode 197: shlo_l_32 — Reg[rD] = sext_4(uint32(Reg[rA]) << (Reg[rB] % 32))
func (c *Compiler) emitShloL32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	return c.emitShift32(a, instr, (*asm.Assembler).Shl32RegCL)
}

// opcode 198: shlo_r_32 — Reg[rD] = sext_4(uint32(Reg[rA]) >> (Reg[rB] % 32))
func (c *Compiler) emitShloR32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	return c.emitShift32(a, instr, (*asm.Assembler).Shr32RegCL)
}

// opcode 199: shar_r_32 — Reg[rD] = int32(Reg[rA]) >> (Reg[rB] % 32)
func (c *Compiler) emitSharR32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	return c.emitShift32(a, instr, (*asm.Assembler).Sar32RegCL)
}

// emitShift32 — shared JIT for opcodes 197–199 (32-bit shift by Reg[rB] in CL)
func (c *Compiler) emitShift32(a *asm.Assembler, instr *PVM.InstrMeta, shiftFn func(*asm.Assembler, asm.Register)) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.MovRegToReg(dReg, aReg)
	emitSignExt32(a, dReg)
	shiftFn(a, dReg)
	emitSignExt32(a, dReg)
	return nil
}

// ---- 4.7.2 64-bit arithmetic ----

// opcode 200: add_64 — Reg[rD] = Reg[rA] + Reg[rB]
func (c *Compiler) emitAdd64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.AddRegReg(RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// opcode 201: sub_64 — Reg[rD] = Reg[rA] - Reg[rB]
func (c *Compiler) emitSub64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.SubRegReg(RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// opcode 202: mul_64 — Reg[rD] = Reg[rA] * Reg[rB]
func (c *Compiler) emitMul64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.ImulRegReg(RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// opcode 203: div_u_64 — Reg[rD] = Reg[rA] /u Reg[rB]; div0 → 2^64-1
func (c *Compiler) emitDivU64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	dReg, aReg, bReg := threeRegFromMeta(instr)
	divByZero := fmt.Sprintf("divz_%d", pc)
	done := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondEQ, divByZero)

	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(asm.RAX, aReg)
	a.Xor(asm.RDX)
	a.Div(RegScratch)
	emitRestoreDiv64Result(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(divByZero)
	a.MovImm64ToReg(dReg, ^uint64(0))

	_ = a.BindLabel(done)
	return nil
}

// opcode 204: div_s_64 — Reg[rD] = int64(Reg[rA]) / int64(Reg[rB]); div0 → 2^64-1
func (c *Compiler) emitDivS64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	dReg, aReg, bReg := threeRegFromMeta(instr)
	divByZero := fmt.Sprintf("divz_%d", pc)
	divNegOne := fmt.Sprintf("divm1_%d", pc)
	normalDiv := fmt.Sprintf("div_%d", pc)
	done := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondEQ, divByZero)

	a.CmpRegImm32(RegScratch, -1)
	a.Jcc(asm.CondNE, normalDiv)

	// b == -1 && a == INT64_MIN → d = a (interpreter overflow case)
	a.MovImm64ToReg(RegScratch, uint64(1)<<63)
	a.CmpRegReg(aReg, RegScratch)
	a.Jcc(asm.CondNE, divNegOne)
	a.MovRegToReg(dReg, aReg)
	a.Jmp(done)

	// b == -1 && a != INT64_MIN → d = -a
	_ = a.BindLabel(divNegOne)
	a.MovRegToReg(dReg, aReg)
	a.Neg(dReg)
	a.Jmp(done)

	// default: signed divide
	_ = a.BindLabel(normalDiv)
	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(asm.RAX, aReg)
	a.Cqo()
	a.Idiv(RegScratch)
	emitRestoreDiv64Result(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(divByZero)
	a.MovImm64ToReg(dReg, ^uint64(0))

	_ = a.BindLabel(done)
	return nil
}

// opcode 205: rem_u_64 — Reg[rD] = Reg[rA] % Reg[rB]; div0 → Reg[rA]
func (c *Compiler) emitRemU64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	dReg, aReg, bReg := threeRegFromMeta(instr)
	divByZero := fmt.Sprintf("divz_%d", pc)
	done := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondEQ, divByZero)

	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(asm.RAX, aReg)
	a.Xor(asm.RDX)
	a.Div(RegScratch)
	emitRestoreDiv64Remainder(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(divByZero)
	a.MovRegToReg(dReg, aReg)

	_ = a.BindLabel(done)
	return nil
}

// opcode 206: rem_s_64 — Reg[rD] = smod(int64(Reg[rA]), int64(Reg[rB]))
func (c *Compiler) emitRemS64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	dReg, aReg, bReg := threeRegFromMeta(instr)
	divByZero := fmt.Sprintf("divz_%d", pc)
	divNegOne := fmt.Sprintf("divm1_%d", pc)
	done := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondEQ, divByZero)

	a.CmpRegImm32(RegScratch, -1)
	a.Jcc(asm.CondEQ, divNegOne)

	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(asm.RAX, aReg)
	a.Cqo()
	a.Idiv(RegScratch)
	emitRestoreDiv64Remainder(a, dReg)
	a.Jmp(done)

	_ = a.BindLabel(divNegOne)
	a.MovImm32ToReg(dReg, 0)
	a.Jmp(done)

	_ = a.BindLabel(divByZero)
	a.MovRegToReg(dReg, aReg)

	_ = a.BindLabel(done)
	return nil
}

// opcode 207: shlo_l_64 — Reg[rD] = Reg[rA] << (Reg[rB] % 64)
func (c *Compiler) emitShloL64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	return c.emitShift64(a, instr, (*asm.Assembler).ShlRegCL)
}

// opcode 208: shlo_r_64 — Reg[rD] = Reg[rA] >> (Reg[rB] % 64)
func (c *Compiler) emitShloR64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	return c.emitShift64(a, instr, (*asm.Assembler).ShrRegCL)
}

// opcode 209: shar_r_64 — Reg[rD] = int64(Reg[rA]) >> (Reg[rB] % 64)
func (c *Compiler) emitSharR64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	return c.emitShift64(a, instr, (*asm.Assembler).SarRegCL)
}

// emitShift64 — shared JIT for opcodes 207–209 (64-bit shift by Reg[rB] in CL)
func (c *Compiler) emitShift64(a *asm.Assembler, instr *PVM.InstrMeta, shiftFn func(*asm.Assembler, asm.Register)) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.MovRegToReg(dReg, aReg)
	shiftFn(a, dReg)
	return nil
}

// ---- 4.7.3 Bitwise ----

// opcode 210: and — Reg[rD] = Reg[rA] & Reg[rB]
func (c *Compiler) emitAnd(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.AndRegReg(RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// opcode 211: xor — Reg[rD] = Reg[rA] ^ Reg[rB]
func (c *Compiler) emitXor(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.XorRegReg(RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// opcode 212: or — Reg[rD] = Reg[rA] | Reg[rB]
func (c *Compiler) emitOr(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.OrRegReg(RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// opcode 224: and_inv — Reg[rD] = Reg[rA] & ~Reg[rB]
func (c *Compiler) emitAndInv(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.Not(RegScratch)
	a.MovRegToReg(dReg, aReg)
	a.AndRegReg(dReg, RegScratch)
	return nil
}

// opcode 225: or_inv — Reg[rD] = Reg[rA] | ~Reg[rB]
func (c *Compiler) emitOrInv(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.Not(RegScratch)
	a.MovRegToReg(dReg, aReg)
	a.OrRegReg(dReg, RegScratch)
	return nil
}

// opcode 226: xnor — Reg[rD] = ~(Reg[rA] ^ Reg[rB])
func (c *Compiler) emitXnor(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, aReg)
	a.XorRegReg(RegScratch, bReg)
	a.Not(RegScratch)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// ---- 4.7.4 Multiplication upper half ----

// opcode 213: mul_upper_s_s — Reg[rD] = high 64 bits of int64(Reg[rA]) * int64(Reg[rB])
func (c *Compiler) emitMulUpperSS(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(RegScratch, bReg)
	a.MovRegToReg(asm.RAX, aReg)
	a.IMulHigh(RegScratch)
	a.MovRegToReg(dReg, asm.RDX)
	emitRestoreRAXRDXIfNeeded(a, dReg)
	return nil
}

// opcode 214: mul_upper_u_u — Reg[rD] = high 64 bits of Reg[rA] *u Reg[rB]
func (c *Compiler) emitMulUpperUU(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(RegScratch, bReg)
	a.MovRegToReg(asm.RAX, aReg)
	a.MulHigh(RegScratch)
	a.MovRegToReg(dReg, asm.RDX)
	emitRestoreRAXRDXIfNeeded(a, dReg)
	return nil
}

// opcode 215: mul_upper_s_u — Reg[rD] = high 64 bits of int64(Reg[rA]) *u Reg[rB]
func (c *Compiler) emitMulUpperSU(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	dReg, aReg, bReg := threeRegFromMeta(instr)
	positive := fmt.Sprintf("positive_%d", pc)
	done := fmt.Sprintf("done_%d", pc)

	emitSaveRAXRDX(a, dReg)
	a.MovRegToReg(RegScratch, bReg)
	a.TestRegReg(aReg, aReg)
	a.Jcc(asm.CondNS, positive)
	a.MovRegToReg(asm.RAX, aReg)
	a.MulHigh(RegScratch)
	a.MovRegToReg(dReg, asm.RDX)
	a.SubRegReg(dReg, RegScratch)
	a.Jmp(done)

	_ = a.BindLabel(positive)
	a.MovRegToReg(asm.RAX, aReg)
	a.MulHigh(RegScratch)
	a.MovRegToReg(dReg, asm.RDX)

	_ = a.BindLabel(done)
	emitRestoreRAXRDXIfNeeded(a, dReg)
	return nil
}

// ---- 4.7.5 Comparisons ----

// opcode 216: set_lt_u — Reg[rD] = (Reg[rA] <u Reg[rB]) ? 1 : 0
func (c *Compiler) emitSetLtU(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.CmpRegReg(aReg, bReg)
	a.MovImm32ToReg(dReg, 0)
	a.SetCC(asm.CondB, dReg)
	return nil
}

// opcode 217: set_lt_s — Reg[rD] = (int64(Reg[rA]) < int64(Reg[rB])) ? 1 : 0
func (c *Compiler) emitSetLtS(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.CmpRegReg(aReg, bReg)
	a.MovImm32ToReg(dReg, 0)
	a.SetCC(asm.CondLT, dReg)
	return nil
}

// opcode 218: cmov_iz — if Reg[rB] == 0 then Reg[rD] = Reg[rA]
func (c *Compiler) emitCmovIz(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.TestRegReg(bReg, bReg)
	a.CMovCC(asm.CondEQ, dReg, aReg)
	return nil
}

// opcode 219: cmov_nz — if Reg[rB] != 0 then Reg[rD] = Reg[rA]
func (c *Compiler) emitCmovNz(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.TestRegReg(bReg, bReg)
	a.CMovCC(asm.CondNE, dReg, aReg)
	return nil
}

// ---- 4.7.6 Rotations ----

// opcode 220: rot_l_64 — Reg[rD] = rol64(Reg[rA], Reg[rB] % 64)
func (c *Compiler) emitRotL64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.MovRegToReg(dReg, aReg)
	a.RolRegCL(dReg)
	return nil
}

// opcode 221: rot_l_32 — Reg[rD] = sext_4(rol32(uint32(Reg[rA]), Reg[rB] % 32))
func (c *Compiler) emitRotL32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.MovRegToReg(dReg, aReg)
	emitSignExt32(a, dReg)
	a.Rol32RegCL(dReg)
	emitSignExt32(a, dReg)
	return nil
}

// opcode 222: rot_r_64 — Reg[rD] = ror64(Reg[rA], Reg[rB] % 64)
func (c *Compiler) emitRotR64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.MovRegToReg(dReg, aReg)
	a.RorRegCL(dReg)
	return nil
}

// opcode 223: rot_r_32 — Reg[rD] = sext_4(ror32(uint32(Reg[rA]), Reg[rB] % 32))
func (c *Compiler) emitRotR32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.MovRegToReg(dReg, aReg)
	emitSignExt32(a, dReg)
	a.Ror32RegCL(dReg)
	emitSignExt32(a, dReg)
	return nil
}

// ---- 4.7.7 Min/Max ----

// opcode 227: max — Reg[rD] = max(int64(Reg[rA]), int64(Reg[rB]))
func (c *Compiler) emitMax(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.CmpRegReg(aReg, bReg)
	a.MovRegToReg(RegScratch, aReg)
	a.CMovCC(asm.CondLT, RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// opcode 228: max_u — Reg[rD] = max(Reg[rA], Reg[rB])
func (c *Compiler) emitMaxU(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.CmpRegReg(aReg, bReg)
	a.MovRegToReg(RegScratch, aReg)
	a.CMovCC(asm.CondB, RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// opcode 229: min — Reg[rD] = min(int64(Reg[rA]), int64(Reg[rB]))
func (c *Compiler) emitMin(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.CmpRegReg(aReg, bReg)
	a.MovRegToReg(RegScratch, aReg)
	a.CMovCC(asm.CondGT, RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// opcode 230: min_u — Reg[rD] = min(Reg[rA], Reg[rB])
func (c *Compiler) emitMinU(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg, bReg := threeRegFromMeta(instr)
	a.CmpRegReg(aReg, bReg)
	a.MovRegToReg(RegScratch, aReg)
	a.CMovCC(asm.CondA, RegScratch, bReg)
	a.MovRegToReg(dReg, RegScratch)
	return nil
}

// ---- Division helpers (used by div/rem opcodes when RAX/RDX are PVM registers) ----

// emitSaveRAXRDX spills native RAX/RDX to the control region when dst is not those registers.
func emitSaveRAXRDX(a *asm.Assembler, dstReg asm.Register) {
	if dstReg != asm.RAX {
		a.MovRegToMem(RegGuestBase, regOffset(0), asm.RAX)
	}
	if dstReg != asm.RDX {
		a.MovRegToMem(RegGuestBase, regOffset(1), asm.RDX)
	}
}

// emitRestoreDiv32Result moves EAX quotient into dst and restores spilled RAX/RDX.
func emitRestoreDiv32Result(a *asm.Assembler, dstReg asm.Register) {
	if dstReg != asm.RAX {
		a.MovRegToReg(dstReg, asm.RAX)
	}
	emitSignExt32(a, dstReg)
	emitRestoreRAXRDXIfNeeded(a, dstReg)
}

// emitRestoreDiv32Remainder moves EDX remainder into dst and restores spilled RAX/RDX.
func emitRestoreDiv32Remainder(a *asm.Assembler, dstReg asm.Register) {
	if dstReg != asm.RDX {
		a.MovRegToReg(dstReg, asm.RDX)
	}
	emitSignExt32(a, dstReg)
	emitRestoreRAXRDXIfNeeded(a, dstReg)
}

// emitRestoreDiv64Result moves RAX quotient into dst and restores spilled RAX/RDX.
func emitRestoreDiv64Result(a *asm.Assembler, dstReg asm.Register) {
	if dstReg != asm.RAX {
		a.MovRegToReg(dstReg, asm.RAX)
	}
	emitRestoreRAXRDXIfNeeded(a, dstReg)
}

// emitRestoreDiv64Remainder moves RDX remainder into dst and restores spilled RAX/RDX.
func emitRestoreDiv64Remainder(a *asm.Assembler, dstReg asm.Register) {
	if dstReg != asm.RDX {
		a.MovRegToReg(dstReg, asm.RDX)
	}
	emitRestoreRAXRDXIfNeeded(a, dstReg)
}

// emitRestoreRAXRDXIfNeeded reloads spilled native RAX/RDX after a div sequence.
func emitRestoreRAXRDXIfNeeded(a *asm.Assembler, dstReg asm.Register) {
	if dstReg != asm.RAX {
		a.MovMemToReg(asm.RAX, RegGuestBase, regOffset(0))
	}
	if dstReg != asm.RDX {
		a.MovMemToReg(asm.RDX, RegGuestBase, regOffset(1))
	}
}
