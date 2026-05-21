//go:build linux && amd64

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// ---- 4.6.1 32-bit arithmetic ----

// opcode 131: add_imm_32 — Reg[rA] = sext_4(uint32(Reg[rB] + vX))
func (c *Compiler) emitAddImm32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	if fitsInt32(vX) {
		a.Add32RegImm32(aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.Add32RegReg(aReg, RegScratch)
	}
	emitSignExt32(a, aReg)
	return nil
}

// opcode 135: mul_imm_32 — Reg[rA] = sext_4(uint32(Reg[rB] * vX))
func (c *Compiler) emitMulImm32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovRegToReg(aReg, bReg)
	emitZeroExt32(a, aReg)
	if fitsInt32(vX) {
		a.ImulRegImm32(aReg, aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		emitZeroExt32(a, RegScratch)
		a.ImulRegReg(aReg, RegScratch)
	}
	emitSignExt32(a, aReg)
	return nil
}

// opcode 141: neg_add_imm_32 — Reg[rA] = sext_4(uint32(vX - Reg[rB]))
func (c *Compiler) emitNegAddImm32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovImm64ToReg(RegScratch, vX)
	a.Sub32RegReg(RegScratch, bReg)
	a.MovRegToReg(aReg, RegScratch)
	emitSignExt32(a, aReg)
	return nil
}

// ---- 4.6.2 64-bit arithmetic ----

// opcode 149: add_imm_64 — Reg[rA] = Reg[rB] + vX
func (c *Compiler) emitAddImm64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	if fitsInt32(vX) {
		a.AddRegImm32(aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.AddRegReg(aReg, RegScratch)
	}
	return nil
}

// opcode 150: mul_imm_64 — Reg[rA] = Reg[rB] * vX
func (c *Compiler) emitMulImm64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if fitsInt32(vX) {
		a.ImulRegImm32(aReg, bReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		if aReg != bReg {
			a.MovRegToReg(aReg, bReg)
		}
		a.ImulRegReg(aReg, RegScratch)
	}
	return nil
}

// opcode 154: neg_add_imm_64 — Reg[rA] = vX - Reg[rB]
func (c *Compiler) emitNegAddImm64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovImm64ToReg(RegScratch, vX)
	a.SubRegReg(RegScratch, bReg)
	a.MovRegToReg(aReg, RegScratch)
	return nil
}

// ---- 4.6.3 Shifts ----

// opcode 138: shlo_l_imm_32 — Reg[rA] = sext_4(uint32(Reg[rB] << (vX & 31)))
func (c *Compiler) emitShloLImm32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	a.Shl32RegImm(aReg, uint8(vX)&31)
	emitSignExt32(a, aReg)
	return nil
}

// opcode 139: shlo_r_imm_32 — Reg[rA] = sext_4(uint32(Reg[rB]) >> (vX & 31))
func (c *Compiler) emitShloRImm32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	a.Shr32RegImm(aReg, uint8(vX)&31)
	emitSignExt32(a, aReg)
	return nil
}

// opcode 140: shar_r_imm_32 — Reg[rA] = sign_extend(int32(Reg[rB]) >> (vX & 31))
func (c *Compiler) emitSharRImm32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	a.Sar32RegImm(aReg, uint8(vX)&31)
	emitSignExt32(a, aReg)
	return nil
}

// opcode 144: shlo_l_imm_alt_32 — Reg[rA] = sext_4(uint32(vX << (Reg[rB] & 31)))
func (c *Compiler) emitShloLImmAlt32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.AndRegImm32(RegScratch, 31)
	// CL = low byte of RCX = RegScratch
	a.MovImm64ToReg(aReg, vX)
	a.Shl32RegCL(aReg)
	emitSignExt32(a, aReg)
	return nil
}

// opcode 145: shlo_r_imm_alt_32 — Reg[rA] = sext_4(uint32(vX) >> (Reg[rB] & 31))
func (c *Compiler) emitShloRImmAlt32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.AndRegImm32(RegScratch, 31)
	a.MovImm64ToReg(aReg, vX)
	a.Shr32RegCL(aReg)
	emitSignExt32(a, aReg)
	return nil
}

// opcode 146: shar_r_imm_alt_32 — Reg[rA] = sign_extend(int32(vX) >> (Reg[rB] & 31))
func (c *Compiler) emitSharRImmAlt32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.AndRegImm32(RegScratch, 31)
	a.MovImm32ToReg(aReg, int32(int64(vX)))
	a.Sar32RegCL(aReg)
	emitSignExt32(a, aReg)
	return nil
}

// opcode 151: shlo_l_imm_64 — Reg[rA] = Reg[rB] << (vX & 63)
func (c *Compiler) emitShloLImm64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	a.ShlRegImm(aReg, uint8(vX)&63)
	return nil
}

// opcode 152: shlo_r_imm_64 — Reg[rA] = Reg[rB] >> (vX & 63)
func (c *Compiler) emitShloRImm64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	a.ShrRegImm(aReg, uint8(vX)&63)
	return nil
}

// opcode 153: shar_r_imm_64 — Reg[rA] = int64(Reg[rB]) >> (vX & 63)
func (c *Compiler) emitSharRImm64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	a.SarRegImm(aReg, uint8(vX)&63)
	return nil
}

// opcode 155: shlo_l_imm_alt_64 — Reg[rA] = vX << (Reg[rB] & 63)
func (c *Compiler) emitShloLImmAlt64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.AndRegImm32(RegScratch, 63)
	a.MovImm64ToReg(aReg, vX)
	a.ShlRegCL(aReg)
	return nil
}

// opcode 156: shlo_r_imm_alt_64 — Reg[rA] = vX >> (Reg[rB] & 63)
func (c *Compiler) emitShloRImmAlt64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.AndRegImm32(RegScratch, 63)
	a.MovImm64ToReg(aReg, vX)
	a.ShrRegCL(aReg)
	return nil
}

// opcode 157: shar_r_imm_alt_64 — Reg[rA] = int64(vX) >> (Reg[rB] & 63)
func (c *Compiler) emitSharRImmAlt64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.AndRegImm32(RegScratch, 63)
	if fitsInt32(vX) {
		a.MovImm32ToReg(aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(aReg, vX)
	}
	a.SarRegCL(aReg)
	return nil
}

// ---- 4.6.4 Rotations ----

// opcode 158: rot_r_64_imm — Reg[rA] = ror64(Reg[rB], vX & 63)
func (c *Compiler) emitRotRImm64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	a.RorRegImm(aReg, uint8(vX)&63)
	return nil
}

// opcode 159: rot_r_64_imm_alt — Reg[rA] = ror64(vX, Reg[rB] & 63)
func (c *Compiler) emitRotRImmAlt64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.AndRegImm32(RegScratch, 63)
	a.MovImm64ToReg(aReg, vX)
	a.RorRegCL(aReg)
	return nil
}

// opcode 160: rot_r_32_imm — Reg[rA] = sext_4(ror32(uint32(Reg[rB]), vX & 31))
func (c *Compiler) emitRotRImm32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	a.Ror32RegImm(aReg, uint8(vX)&31)
	emitSignExt32(a, aReg)
	return nil
}

// opcode 161: rot_r_32_imm_alt — Reg[rA] = sext_4(ror32(uint32(vX), Reg[rB] & 31))
func (c *Compiler) emitRotRImmAlt32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	a.MovRegToReg(RegScratch, bReg)
	a.AndRegImm32(RegScratch, 31)
	a.MovImm64ToReg(aReg, vX)
	a.Ror32RegCL(aReg)
	emitSignExt32(a, aReg)
	return nil
}

// ---- 4.6.5 Logic ----

// opcode 132: and_imm — Reg[rA] = Reg[rB] & vX
func (c *Compiler) emitAndImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	if fitsInt32(vX) {
		a.AndRegImm32(aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.AndRegReg(aReg, RegScratch)
	}
	return nil
}

// opcode 133: xor_imm — Reg[rA] = Reg[rB] ^ vX
func (c *Compiler) emitXorImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	if fitsInt32(vX) {
		a.XorRegImm32(aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.XorRegReg(aReg, RegScratch)
	}
	return nil
}

// opcode 134: or_imm — Reg[rA] = Reg[rB] | vX
func (c *Compiler) emitOrImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if aReg != bReg {
		a.MovRegToReg(aReg, bReg)
	}
	if fitsInt32(vX) {
		a.OrRegImm32(aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.OrRegReg(aReg, RegScratch)
	}
	return nil
}

// ---- 4.6.6 Comparisons ----

// opcode 136: set_lt_u_imm — Reg[rA] = (Reg[rB] <u vX) ? 1 : 0
func (c *Compiler) emitSetLtUImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if fitsInt32(vX) {
		a.CmpRegImm32(bReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.CmpRegReg(bReg, RegScratch)
	}
	a.MovImm32ToReg(aReg, 0)
	a.SetCC(asm.CondB, aReg)
	return nil
}

// opcode 137: set_lt_s_imm — Reg[rA] = (int64(Reg[rB]) < int64(vX)) ? 1 : 0
func (c *Compiler) emitSetLtSImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if fitsInt32(vX) {
		a.CmpRegImm32(bReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.CmpRegReg(bReg, RegScratch)
	}
	a.MovImm32ToReg(aReg, 0)
	a.SetCC(asm.CondLT, aReg)
	return nil
}

// opcode 142: set_gt_u_imm — Reg[rA] = (Reg[rB] >u vX) ? 1 : 0
func (c *Compiler) emitSetGtUImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if fitsInt32(vX) {
		a.CmpRegImm32(bReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.CmpRegReg(bReg, RegScratch)
	}
	a.MovImm32ToReg(aReg, 0)
	a.SetCC(asm.CondA, aReg)
	return nil
}

// opcode 143: set_gt_s_imm — Reg[rA] = (int64(Reg[rB]) > int64(vX)) ? 1 : 0
func (c *Compiler) emitSetGtSImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	if fitsInt32(vX) {
		a.CmpRegImm32(bReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.CmpRegReg(bReg, RegScratch)
	}
	a.MovImm32ToReg(aReg, 0)
	a.SetCC(asm.CondGT, aReg)
	return nil
}

// ---- 4.6.7 Conditional moves ----

// opcode 147: cmov_iz_imm — if Reg[rB] == 0 then Reg[rA] = vX
func (c *Compiler) emitCmovIzImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	skipLabel := fmt.Sprintf("skip_cmov_%d", pc)
	a.TestRegReg(bReg, bReg)
	a.Jcc(asm.CondNE, skipLabel)
	if fitsInt32(vX) {
		a.MovImm32ToReg(aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(aReg, vX)
	}
	_ = a.BindLabel(skipLabel)
	return nil
}

// opcode 148: cmov_nz_imm — if Reg[rB] != 0 then Reg[rA] = vX
func (c *Compiler) emitCmovNzImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	aReg, bReg, vX := twoRegImmFromMeta(instr)
	skipLabel := fmt.Sprintf("skip_cmov_%d", pc)
	a.TestRegReg(bReg, bReg)
	a.Jcc(asm.CondEQ, skipLabel)
	if fitsInt32(vX) {
		a.MovImm32ToReg(aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(aReg, vX)
	}
	_ = a.BindLabel(skipLabel)
	return nil
}
