package asm

// ---------------------------------------------------------------------------
// Stack operations
// ---------------------------------------------------------------------------

// Push emits PUSH r64. Encoding: [REX?] 50+rd
func (a *Assembler) Push(reg Register) {
	if reg.IsExtended() {
		a.buf.Emit(rexByte(false, false, false, true))
	}
	a.buf.Emit(0x50 + reg.Lo3())
}

// Pop emits POP r64. Encoding: [REX?] 58+rd
func (a *Assembler) Pop(reg Register) {
	if reg.IsExtended() {
		a.buf.Emit(rexByte(false, false, false, true))
	}
	a.buf.Emit(0x58 + reg.Lo3())
}

// ---------------------------------------------------------------------------
// MOV — data movement
// ---------------------------------------------------------------------------

// MovRegToReg emits MOV r64, r64. Encoding: REX.W 89 /r
func (a *Assembler) MovRegToReg(dst, src Register) {
	emitREX(a.buf, true, src, dst)
	a.buf.Emit(0x89, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// MovMemToReg emits MOV r64, [base+disp]. Encoding: REX.W 8B /r
func (a *Assembler) MovMemToReg(dst, base Register, disp int32) {
	emitREX(a.buf, true, dst, base)
	a.buf.Emit(0x8B)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// MovRegToMem emits MOV [base+disp], r64. Encoding: REX.W 89 /r
func (a *Assembler) MovRegToMem(base Register, disp int32, src Register) {
	emitREX(a.buf, true, src, base)
	a.buf.Emit(0x89)
	emitMemOp(a.buf, src.Lo3(), base, disp)
}

// MovMemImm32 emits MOV qword [base+disp], imm32 (sign-extended). Encoding: REX.W C7 /0 id
func (a *Assembler) MovMemImm32(base Register, disp int32, imm int32) {
	emitREX(a.buf, true, Register(0), base)
	a.buf.Emit(0xC7)
	emitMemOp(a.buf, 0, base, disp)
	a.buf.EmitInt32LE(imm)
}

// MovMemImm32_32 emits MOV dword [base+disp], imm32 (32-bit store). Encoding: [REX?] C7 /0 id
func (a *Assembler) MovMemImm32_32(base Register, disp int32, imm int32) {
	if base.IsExtended() {
		a.buf.Emit(rexByte(false, false, false, true))
	}
	a.buf.Emit(0xC7)
	emitMemOp(a.buf, 0, base, disp)
	a.buf.EmitInt32LE(imm)
}

// MovImm64ToReg emits MOV r64, imm64 (movabs). Encoding: REX.W B8+rd io
func (a *Assembler) MovImm64ToReg(dst Register, imm uint64) {
	a.buf.Emit(rexByte(true, false, false, dst.IsExtended()))
	a.buf.Emit(0xB8 + dst.Lo3())
	a.buf.EmitUint64LE(imm)
}

// MovImm32ToReg emits MOV r64, imm32 (sign-extended to 64-bit). Encoding: REX.W C7 /0 mod=11 id
func (a *Assembler) MovImm32ToReg(dst Register, imm int32) {
	emitREX(a.buf, true, Register(0), dst)
	a.buf.Emit(0xC7, modRM(0x03, 0, dst.Lo3()))
	a.buf.EmitInt32LE(imm)
}

// ZeroExtend32 emits MOV r32d, r32d which zeros the upper 32 bits. Encoding: [REX?] 89 /r mod=11 (no REX.W)
func (a *Assembler) ZeroExtend32(reg Register) {
	emitREX(a.buf, false, reg, reg)
	a.buf.Emit(0x89, modRM(0x03, reg.Lo3(), reg.Lo3()))
}

// MovEcxEcx is an alias for ZeroExtend32.
func (a *Assembler) MovEcxEcx(reg Register) { a.ZeroExtend32(reg) }

// ---------------------------------------------------------------------------
// MOVZX / MOVSX — register-to-register extensions
// ---------------------------------------------------------------------------

// MovFromByteToRegZx emits MOVZX r64, r8. Encoding: REX.W 0F B6 /r mod=11
func (a *Assembler) MovFromByteToRegZx(dst, src Register) {
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x0F, 0xB6, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// MovFromByteToRegSx emits MOVSX r64, r8. Encoding: REX.W 0F BE /r mod=11
func (a *Assembler) MovFromByteToRegSx(dst, src Register) {
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x0F, 0xBE, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// MovFromWordToRegZx emits MOVZX r64, r16. Encoding: REX.W 0F B7 /r mod=11
func (a *Assembler) MovFromWordToRegZx(dst, src Register) {
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x0F, 0xB7, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// MovFromWordToRegSx emits MOVSX r64, r16. Encoding: REX.W 0F BF /r mod=11
func (a *Assembler) MovFromWordToRegSx(dst, src Register) {
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x0F, 0xBF, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// MovFromDwordToRegSx emits MOVSXD r64, r32. Encoding: REX.W 63 /r mod=11
func (a *Assembler) MovFromDwordToRegSx(dst, src Register) {
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x63, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// ---------------------------------------------------------------------------
// LEA
// ---------------------------------------------------------------------------

// LeaRegMem emits LEA r64, [base+disp]. Encoding: REX.W 8D /r
func (a *Assembler) LeaRegMem(dst, base Register, disp int32) {
	emitREX(a.buf, true, dst, base)
	a.buf.Emit(0x8D)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// LeaRIPRel emits LEA r64, [RIP+rel32]. Encoding: REX.W 8D /r (mod=00, rm=101)
func (a *Assembler) LeaRIPRel(dst Register, label string) {
	a.buf.Emit(rexByte(true, dst.IsExtended(), false, false))
	a.buf.Emit(0x8D, modRM(0x00, dst.Lo3(), 0x05))
	a.buf.UseLabel32(label)
}

// ---------------------------------------------------------------------------
// JMP / CALL / RET / NOP
// ---------------------------------------------------------------------------

// Jmp emits JMP rel32 (near jump to label). Encoding: E9 cd
func (a *Assembler) Jmp(label string) {
	a.buf.Emit(0xE9)
	a.buf.UseLabel32(label)
}

// JmpReg emits JMP r64 (indirect). Encoding: [REX?] FF /4 mod=11
func (a *Assembler) JmpReg(reg Register) {
	if reg.IsExtended() {
		a.buf.Emit(rexByte(false, false, false, true))
	}
	a.buf.Emit(0xFF, modRM(0x03, 4, reg.Lo3()))
}

// JmpMem emits JMP qword [base+disp] (indirect via memory). Encoding: [REX?] FF /4
func (a *Assembler) JmpMem(base Register, disp int32) {
	if base.IsExtended() {
		a.buf.Emit(rexByte(false, false, false, true))
	}
	a.buf.Emit(0xFF)
	emitMemOp(a.buf, 4, base, disp)
}

// Call emits CALL rel32 (near call to label). Encoding: E8 cd
func (a *Assembler) Call(label string) {
	a.buf.Emit(0xE8)
	a.buf.UseLabel32(label)
}

// CallReg emits CALL r64 (indirect). Encoding: [REX?] FF /2 mod=11
func (a *Assembler) CallReg(reg Register) {
	if reg.IsExtended() {
		a.buf.Emit(rexByte(false, false, false, true))
	}
	a.buf.Emit(0xFF, modRM(0x03, 2, reg.Lo3()))
}

// Ret emits RET. Encoding: C3
func (a *Assembler) Ret() { a.buf.Emit(0xC3) }

// Nop emits a single-byte NOP. Encoding: 90
func (a *Assembler) Nop() { a.buf.Emit(0x90) }

// ---------------------------------------------------------------------------
// Jcc / Label
// ---------------------------------------------------------------------------

// Jcc emits Jcc rel32 (conditional jump to label). Encoding: 0F 8x cd
func (a *Assembler) Jcc(cc ConditionCode, label string) {
	a.buf.Emit(0x0F, 0x80+byte(cc))
	a.buf.UseLabel32(label)
}

// BindLabel binds a label at the current code position.
func (a *Assembler) BindLabel(name string) error {
	return a.buf.BindLabel(name)
}

// ---------------------------------------------------------------------------
// ADD
// ---------------------------------------------------------------------------

// AddRegReg emits ADD r64, r64. Encoding: REX.W 01 /r
func (a *Assembler) AddRegReg(dst, src Register) {
	emitREX(a.buf, true, src, dst)
	a.buf.Emit(0x01, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// AddRegMem emits ADD r64, [base+disp]. Encoding: REX.W 03 /r
func (a *Assembler) AddRegMem(dst, base Register, disp int32) {
	emitREX(a.buf, true, dst, base)
	a.buf.Emit(0x03)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// AddRegImm32 emits ADD r64, imm32. Encoding: REX.W 81 /0 id (or 83 /0 ib for imm8)
func (a *Assembler) AddRegImm32(reg Register, imm int32) {
	a.emitALURegImm(0, reg, imm)
}

// Add32RegReg emits ADD r32, r32. Encoding: [REX?] 01 /r (no REX.W)
func (a *Assembler) Add32RegReg(dst, src Register) {
	emitREX(a.buf, false, src, dst)
	a.buf.Emit(0x01, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// ---------------------------------------------------------------------------
// SUB
// ---------------------------------------------------------------------------

// SubRegReg emits SUB r64, r64. Encoding: REX.W 29 /r
func (a *Assembler) SubRegReg(dst, src Register) {
	emitREX(a.buf, true, src, dst)
	a.buf.Emit(0x29, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// SubRegImm32 emits SUB r64, imm32. Encoding: REX.W 81 /5 id (or 83 /5 ib for imm8)
func (a *Assembler) SubRegImm32(reg Register, imm int32) {
	a.emitALURegImm(5, reg, imm)
}

// Sub32RegReg emits SUB r32, r32. Encoding: [REX?] 29 /r (no REX.W)
func (a *Assembler) Sub32RegReg(dst, src Register) {
	emitREX(a.buf, false, src, dst)
	a.buf.Emit(0x29, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// SubMemImm32 emits SUB qword [base+disp], imm32. Encoding: REX.W 81 /5 id (or 83 /5 ib)
func (a *Assembler) SubMemImm32(base Register, disp int32, imm int32) {
	emitREX(a.buf, true, Register(5), base)
	if fitsInt8(imm) {
		a.buf.Emit(0x83)
		emitMemOp(a.buf, 5, base, disp)
		a.buf.Emit(byte(int8(imm)))
	} else {
		a.buf.Emit(0x81)
		emitMemOp(a.buf, 5, base, disp)
		a.buf.EmitInt32LE(imm)
	}
}

// ---------------------------------------------------------------------------
// IMUL
// ---------------------------------------------------------------------------

// ImulRegReg emits IMUL r64, r64 (two-operand). Encoding: REX.W 0F AF /r
func (a *Assembler) ImulRegReg(dst, src Register) {
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x0F, 0xAF, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// ImulRegImm32 emits IMUL r64, r64, imm. Encoding: REX.W 69 /r id (or 6B /r ib for imm8)
func (a *Assembler) ImulRegImm32(dst, src Register, imm int32) {
	emitREX(a.buf, true, dst, src)
	if fitsInt8(imm) {
		a.buf.Emit(0x6B, modRM(0x03, dst.Lo3(), src.Lo3()))
		a.buf.Emit(byte(int8(imm)))
	} else {
		a.buf.Emit(0x69, modRM(0x03, dst.Lo3(), src.Lo3()))
		a.buf.EmitInt32LE(imm)
	}
}

// Imul32RegReg emits IMUL r32, r32 (two-operand, 32-bit). Encoding: [REX?] 0F AF /r
func (a *Assembler) Imul32RegReg(dst, src Register) {
	emitREX(a.buf, false, dst, src)
	a.buf.Emit(0x0F, 0xAF, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// ---------------------------------------------------------------------------
// DIV / IDIV
// ---------------------------------------------------------------------------

// Div emits DIV r64 (unsigned divide RDX:RAX by r64). Encoding: REX.W F7 /6
func (a *Assembler) Div(reg Register) {
	emitREX(a.buf, true, Register(6), reg)
	a.buf.Emit(0xF7, modRM(0x03, 6, reg.Lo3()))
}

// Idiv emits IDIV r64 (signed divide RDX:RAX by r64). Encoding: REX.W F7 /7
func (a *Assembler) Idiv(reg Register) {
	emitREX(a.buf, true, Register(7), reg)
	a.buf.Emit(0xF7, modRM(0x03, 7, reg.Lo3()))
}

// Div32 emits DIV r32 (32-bit unsigned divide). Encoding: [REX?] F7 /6
func (a *Assembler) Div32(reg Register) {
	emitREX(a.buf, false, Register(6), reg)
	a.buf.Emit(0xF7, modRM(0x03, 6, reg.Lo3()))
}

// Idiv32 emits IDIV r32 (32-bit signed divide). Encoding: [REX?] F7 /7
func (a *Assembler) Idiv32(reg Register) {
	emitREX(a.buf, false, Register(7), reg)
	a.buf.Emit(0xF7, modRM(0x03, 7, reg.Lo3()))
}

// ---------------------------------------------------------------------------
// MUL / IMUL — upper half (one-operand form, result in RDX:RAX)
// ---------------------------------------------------------------------------

// MulHigh emits MUL r64 (unsigned, RDX:RAX = RAX * r64). Encoding: REX.W F7 /4
func (a *Assembler) MulHigh(reg Register) {
	emitREX(a.buf, true, Register(4), reg)
	a.buf.Emit(0xF7, modRM(0x03, 4, reg.Lo3()))
}

// IMulHigh emits IMUL r64 (one-operand signed, RDX:RAX = RAX * r64). Encoding: REX.W F7 /5
func (a *Assembler) IMulHigh(reg Register) {
	emitREX(a.buf, true, Register(5), reg)
	a.buf.Emit(0xF7, modRM(0x03, 5, reg.Lo3()))
}

// ---------------------------------------------------------------------------
// NEG / CQO
// ---------------------------------------------------------------------------

// Neg emits NEG r64 (two's complement negate). Encoding: REX.W F7 /3
func (a *Assembler) Neg(reg Register) {
	emitREX(a.buf, true, Register(3), reg)
	a.buf.Emit(0xF7, modRM(0x03, 3, reg.Lo3()))
}

// Cqo emits CQO (sign-extend RAX into RDX:RAX). Encoding: REX.W 99
func (a *Assembler) Cqo() {
	a.buf.Emit(rexByte(true, false, false, false), 0x99)
}

// ---------------------------------------------------------------------------
// CMP / TEST
// ---------------------------------------------------------------------------

// CmpRegReg emits CMP r64, r64. Encoding: REX.W 39 /r
func (a *Assembler) CmpRegReg(dst, src Register) {
	emitREX(a.buf, true, src, dst)
	a.buf.Emit(0x39, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// CmpRegImm32 emits CMP r64, imm32. Encoding: REX.W 81 /7 id (or 83 /7 ib)
func (a *Assembler) CmpRegImm32(reg Register, imm int32) {
	a.emitALURegImm(7, reg, imm)
}

// CmpReg32Imm32 emits CMP r/m32, imm32 (32-bit). Encoding: [REX?] 81 /7 id (or 83 /7 ib)
func (a *Assembler) CmpReg32Imm32(reg Register, imm int32) {
	emitREX(a.buf, false, Register(7), reg)
	if fitsInt8(imm) {
		a.buf.Emit(0x83, modRM(0x03, 7, reg.Lo3()))
		a.buf.Emit(byte(int8(imm)))
	} else {
		a.buf.Emit(0x81, modRM(0x03, 7, reg.Lo3()))
		a.buf.EmitInt32LE(imm)
	}
}

// TestRegReg emits TEST r64, r64. Encoding: REX.W 85 /r
func (a *Assembler) TestRegReg(dst, src Register) {
	emitREX(a.buf, true, src, dst)
	a.buf.Emit(0x85, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// ---------------------------------------------------------------------------
// SetCC / CMOVcc
// ---------------------------------------------------------------------------

// SetCC emits SETcc r8 (set byte on condition). Encoding: [REX?] 0F 9x /0 mod=11
func (a *Assembler) SetCC(cc ConditionCode, reg Register) {
	if needRexForLow8Reg(reg) {
		a.buf.Emit(rexByte(false, false, false, reg.IsExtended()))
	}
	a.buf.Emit(0x0F, 0x90+byte(cc), modRM(0x03, 0, reg.Lo3()))
}

// CMovCC emits CMOVcc r64, r64. Encoding: REX.W 0F 4x /r
func (a *Assembler) CMovCC(cc ConditionCode, dst, src Register) {
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x0F, 0x40+byte(cc), modRM(0x03, dst.Lo3(), src.Lo3()))
}

// ---------------------------------------------------------------------------
// XOR (self-zero idiom)
// ---------------------------------------------------------------------------

// Xor emits XOR r64, r64 (self) to zero the register. Encoding: REX.W 31 /r
func (a *Assembler) Xor(reg Register) {
	emitREX(a.buf, true, reg, reg)
	a.buf.Emit(0x31, modRM(0x03, reg.Lo3(), reg.Lo3()))
}

// ---------------------------------------------------------------------------
// Bitwise: AND / OR / XOR / NOT
// ---------------------------------------------------------------------------

// AndRegReg emits AND r64, r64. Encoding: REX.W 21 /r
func (a *Assembler) AndRegReg(dst, src Register) {
	emitREX(a.buf, true, src, dst)
	a.buf.Emit(0x21, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// AndRegImm32 emits AND r64, imm32. Encoding: REX.W 81 /4 id (or 83 /4 ib)
func (a *Assembler) AndRegImm32(reg Register, imm int32) {
	a.emitALURegImm(4, reg, imm)
}

// OrRegReg emits OR r64, r64. Encoding: REX.W 09 /r
func (a *Assembler) OrRegReg(dst, src Register) {
	emitREX(a.buf, true, src, dst)
	a.buf.Emit(0x09, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// OrRegMem emits OR r64, [base+disp]. Encoding: REX.W 0B /r
func (a *Assembler) OrRegMem(dst, base Register, disp int32) {
	emitREX(a.buf, true, dst, base)
	a.buf.Emit(0x0B)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// OrRegImm32 emits OR r64, imm32. Encoding: REX.W 81 /1 id (or 83 /1 ib)
func (a *Assembler) OrRegImm32(reg Register, imm int32) {
	a.emitALURegImm(1, reg, imm)
}

// XorRegReg emits XOR r64, r64. Encoding: REX.W 31 /r
func (a *Assembler) XorRegReg(dst, src Register) {
	emitREX(a.buf, true, src, dst)
	a.buf.Emit(0x31, modRM(0x03, src.Lo3(), dst.Lo3()))
}

// XorRegImm32 emits XOR r64, imm32. Encoding: REX.W 81 /6 id (or 83 /6 ib)
func (a *Assembler) XorRegImm32(reg Register, imm int32) {
	a.emitALURegImm(6, reg, imm)
}

// Not emits NOT r64. Encoding: REX.W F7 /2
func (a *Assembler) Not(reg Register) {
	emitREX(a.buf, true, Register(2), reg)
	a.buf.Emit(0xF7, modRM(0x03, 2, reg.Lo3()))
}

// ---------------------------------------------------------------------------
// Shifts / Rotates — 64-bit
// ---------------------------------------------------------------------------

// ShlRegCL emits SHL r64, CL. Encoding: REX.W D3 /4
func (a *Assembler) ShlRegCL(reg Register) { a.emitShiftCL(true, 4, reg) }

// ShlRegImm emits SHL r64, imm8. Encoding: REX.W C1 /4 ib (or D1 /4 for imm=1)
func (a *Assembler) ShlRegImm(reg Register, imm uint8) { a.emitShiftImm(true, 4, reg, imm) }

// ShrRegCL emits SHR r64, CL. Encoding: REX.W D3 /5
func (a *Assembler) ShrRegCL(reg Register) { a.emitShiftCL(true, 5, reg) }

// ShrRegImm emits SHR r64, imm8. Encoding: REX.W C1 /5 ib (or D1 /5 for imm=1)
func (a *Assembler) ShrRegImm(reg Register, imm uint8) { a.emitShiftImm(true, 5, reg, imm) }

// SarRegCL emits SAR r64, CL. Encoding: REX.W D3 /7
func (a *Assembler) SarRegCL(reg Register) { a.emitShiftCL(true, 7, reg) }

// SarRegImm emits SAR r64, imm8. Encoding: REX.W C1 /7 ib (or D1 /7 for imm=1)
func (a *Assembler) SarRegImm(reg Register, imm uint8) { a.emitShiftImm(true, 7, reg, imm) }

// RolRegCL emits ROL r64, CL. Encoding: REX.W D3 /0
func (a *Assembler) RolRegCL(reg Register) { a.emitShiftCL(true, 0, reg) }

// RolRegImm emits ROL r64, imm8. Encoding: REX.W C1 /0 ib (or D1 /0 for imm=1)
func (a *Assembler) RolRegImm(reg Register, imm uint8) { a.emitShiftImm(true, 0, reg, imm) }

// RorRegCL emits ROR r64, CL. Encoding: REX.W D3 /1
func (a *Assembler) RorRegCL(reg Register) { a.emitShiftCL(true, 1, reg) }

// RorRegImm emits ROR r64, imm8. Encoding: REX.W C1 /1 ib (or D1 /1 for imm=1)
func (a *Assembler) RorRegImm(reg Register, imm uint8) { a.emitShiftImm(true, 1, reg, imm) }

// ---------------------------------------------------------------------------
// Shifts / Rotates — 32-bit
// ---------------------------------------------------------------------------

// Shl32RegCL emits SHL r32, CL. Encoding: [REX?] D3 /4
func (a *Assembler) Shl32RegCL(reg Register) { a.emitShiftCL(false, 4, reg) }

// Shr32RegImm emits SHR r32, imm8. Encoding: [REX?] C1 /5 ib (or D1 /5 for imm=1)
func (a *Assembler) Shr32RegImm(reg Register, imm uint8) { a.emitShiftImm(false, 5, reg, imm) }

// Sar32RegCL emits SAR r32, CL. Encoding: [REX?] D3 /7
func (a *Assembler) Sar32RegCL(reg Register) { a.emitShiftCL(false, 7, reg) }

// Rol32RegImm emits ROL r32, imm8. Encoding: [REX?] C1 /0 ib (or D1 /0 for imm=1)
func (a *Assembler) Rol32RegImm(reg Register, imm uint8) { a.emitShiftImm(false, 0, reg, imm) }

// Ror32RegCL emits ROR r32, CL. Encoding: [REX?] D3 /1
func (a *Assembler) Ror32RegCL(reg Register) { a.emitShiftCL(false, 1, reg) }

// Shl32RegImm emits SHL r32, imm8. Encoding: [REX?] C1 /4 ib
func (a *Assembler) Shl32RegImm(reg Register, imm uint8) { a.emitShiftImm(false, 4, reg, imm) }

// Shr32RegCL emits SHR r32, CL. Encoding: [REX?] D3 /5
func (a *Assembler) Shr32RegCL(reg Register) { a.emitShiftCL(false, 5, reg) }

// Sar32RegImm emits SAR r32, imm8. Encoding: [REX?] C1 /7 ib
func (a *Assembler) Sar32RegImm(reg Register, imm uint8) { a.emitShiftImm(false, 7, reg, imm) }

// Ror32RegImm emits ROR r32, imm8. Encoding: [REX?] C1 /1 ib
func (a *Assembler) Ror32RegImm(reg Register, imm uint8) { a.emitShiftImm(false, 1, reg, imm) }

// Rol32RegCL emits ROL r32, CL. Encoding: [REX?] D3 /0
func (a *Assembler) Rol32RegCL(reg Register) { a.emitShiftCL(false, 0, reg) }

// Add32RegImm32 emits ADD r32, imm32. Encoding: [REX?] 81 /0 id (or 83 /0 ib)
func (a *Assembler) Add32RegImm32(reg Register, imm int32) {
	emitREX(a.buf, false, Register(0), reg)
	if fitsInt8(imm) {
		a.buf.Emit(0x83, modRM(0x03, 0, reg.Lo3()))
		a.buf.Emit(byte(int8(imm)))
	} else {
		a.buf.Emit(0x81, modRM(0x03, 0, reg.Lo3()))
		a.buf.EmitInt32LE(imm)
	}
}

// ---------------------------------------------------------------------------
// Bit counting: POPCNT / LZCNT / TZCNT — 64-bit
// ---------------------------------------------------------------------------

// Popcnt emits POPCNT r64, r64. Encoding: F3 REX.W 0F B8 /r
func (a *Assembler) Popcnt(dst, src Register) {
	a.buf.Emit(0xF3)
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x0F, 0xB8, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// Lzcnt emits LZCNT r64, r64. Encoding: F3 REX.W 0F BD /r
func (a *Assembler) Lzcnt(dst, src Register) {
	a.buf.Emit(0xF3)
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x0F, 0xBD, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// Tzcnt emits TZCNT r64, r64. Encoding: F3 REX.W 0F BC /r
func (a *Assembler) Tzcnt(dst, src Register) {
	a.buf.Emit(0xF3)
	emitREX(a.buf, true, dst, src)
	a.buf.Emit(0x0F, 0xBC, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// ---------------------------------------------------------------------------
// Bit counting: POPCNT / LZCNT / TZCNT — 32-bit
// ---------------------------------------------------------------------------

// Popcnt32 emits POPCNT r32, r32. Encoding: F3 [REX?] 0F B8 /r
func (a *Assembler) Popcnt32(dst, src Register) {
	a.buf.Emit(0xF3)
	emitREX(a.buf, false, dst, src)
	a.buf.Emit(0x0F, 0xB8, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// Lzcnt32 emits LZCNT r32, r32. Encoding: F3 [REX?] 0F BD /r
func (a *Assembler) Lzcnt32(dst, src Register) {
	a.buf.Emit(0xF3)
	emitREX(a.buf, false, dst, src)
	a.buf.Emit(0x0F, 0xBD, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// Tzcnt32 emits TZCNT r32, r32. Encoding: F3 [REX?] 0F BC /r
func (a *Assembler) Tzcnt32(dst, src Register) {
	a.buf.Emit(0xF3)
	emitREX(a.buf, false, dst, src)
	a.buf.Emit(0x0F, 0xBC, modRM(0x03, dst.Lo3(), src.Lo3()))
}

// ---------------------------------------------------------------------------
// BSWAP
// ---------------------------------------------------------------------------

// Bswap emits BSWAP r64. Encoding: REX.W 0F C8+rd
func (a *Assembler) Bswap(reg Register) {
	a.buf.Emit(rexByte(true, false, false, reg.IsExtended()))
	a.buf.Emit(0x0F, 0xC8+reg.Lo3())
}

// Bswap32 emits BSWAP r32. Encoding: [REX?] 0F C8+rd
func (a *Assembler) Bswap32(reg Register) {
	if reg.IsExtended() {
		a.buf.Emit(rexByte(false, false, false, true))
	}
	a.buf.Emit(0x0F, 0xC8+reg.Lo3())
}

// ---------------------------------------------------------------------------
// Memory Load — sized
// ---------------------------------------------------------------------------

// LoadByte emits MOVZX r32, BYTE [base+disp] (zero-extends to 64-bit).
// Encoding: [REX?] 0F B6 /r (no REX.W; 32-bit dest auto-zero-extends)
func (a *Assembler) LoadByte(dst, base Register, disp int32) {
	emitREX(a.buf, false, dst, base)
	a.buf.Emit(0x0F, 0xB6)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// LoadByteFromMem is an alias for LoadByte.
func (a *Assembler) LoadByteFromMem(dst, base Register, disp int32) { a.LoadByte(dst, base, disp) }

// LoadSignedByte emits MOVSX r64, BYTE [base+disp] (sign-extends to 64-bit).
// Encoding: REX.W 0F BE /r
func (a *Assembler) LoadSignedByte(dst, base Register, disp int32) {
	emitREX(a.buf, true, dst, base)
	a.buf.Emit(0x0F, 0xBE)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// LoadWord emits MOVZX r32, WORD [base+disp] (zero-extends to 64-bit).
// Encoding: [REX?] 0F B7 /r
func (a *Assembler) LoadWord(dst, base Register, disp int32) {
	emitREX(a.buf, false, dst, base)
	a.buf.Emit(0x0F, 0xB7)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// LoadSignedWord emits MOVSX r64, WORD [base+disp] (sign-extends to 64-bit).
// Encoding: REX.W 0F BF /r
func (a *Assembler) LoadSignedWord(dst, base Register, disp int32) {
	emitREX(a.buf, true, dst, base)
	a.buf.Emit(0x0F, 0xBF)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// LoadDword emits MOV r32, DWORD [base+disp] (zero-extends to 64-bit).
// Encoding: [REX?] 8B /r (no REX.W)
func (a *Assembler) LoadDword(dst, base Register, disp int32) {
	emitREX(a.buf, false, dst, base)
	a.buf.Emit(0x8B)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// LoadSignedDword emits MOVSXD r64, DWORD [base+disp] (sign-extends to 64-bit).
// Encoding: REX.W 63 /r
func (a *Assembler) LoadSignedDword(dst, base Register, disp int32) {
	emitREX(a.buf, true, dst, base)
	a.buf.Emit(0x63)
	emitMemOp(a.buf, dst.Lo3(), base, disp)
}

// LoadQword emits MOV r64, QWORD [base+disp]. Same as MovMemToReg.
func (a *Assembler) LoadQword(dst, base Register, disp int32) { a.MovMemToReg(dst, base, disp) }

// ---------------------------------------------------------------------------
// Memory Store — sized
// ---------------------------------------------------------------------------

// StoreByte emits MOV BYTE [base+disp], r8. Encoding: [REX?] 88 /r
func (a *Assembler) StoreByte(base Register, disp int32, src Register) {
	if needRexForLow8Reg(src) || base.IsExtended() {
		a.buf.Emit(rexByte(false, src.IsExtended(), false, base.IsExtended()))
	}
	a.buf.Emit(0x88)
	emitMemOp(a.buf, src.Lo3(), base, disp)
}

// StoreByteRegToMem is an alias for StoreByte.
func (a *Assembler) StoreByteRegToMem(base Register, disp int32, src Register) {
	a.StoreByte(base, disp, src)
}

// StoreWord emits MOV WORD [base+disp], r16. Encoding: 66 [REX?] 89 /r
func (a *Assembler) StoreWord(base Register, disp int32, src Register) {
	a.buf.Emit(0x66)
	emitREX(a.buf, false, src, base)
	a.buf.Emit(0x89)
	emitMemOp(a.buf, src.Lo3(), base, disp)
}

// StoreDword emits MOV DWORD [base+disp], r32. Encoding: [REX?] 89 /r (no REX.W)
func (a *Assembler) StoreDword(base Register, disp int32, src Register) {
	emitREX(a.buf, false, src, base)
	a.buf.Emit(0x89)
	emitMemOp(a.buf, src.Lo3(), base, disp)
}

// StoreQword emits MOV QWORD [base+disp], r64. Same as MovRegToMem.
func (a *Assembler) StoreQword(base Register, disp int32, src Register) {
	a.MovRegToMem(base, disp, src)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// emitALURegImm emits a 64-bit ALU reg,imm32 instruction.
// opExt is the /r extension: ADD=0, OR=1, AND=4, SUB=5, XOR=6, CMP=7.
func (a *Assembler) emitALURegImm(opExt uint8, reg Register, imm int32) {
	emitREX(a.buf, true, Register(opExt), reg)
	if fitsInt8(imm) {
		a.buf.Emit(0x83, modRM(0x03, opExt, reg.Lo3()))
		a.buf.Emit(byte(int8(imm)))
	} else {
		a.buf.Emit(0x81, modRM(0x03, opExt, reg.Lo3()))
		a.buf.EmitInt32LE(imm)
	}
}

// emitShiftCL emits a shift/rotate by CL.
// w64 selects 64-bit (REX.W D3) vs 32-bit ([REX?] D3).
func (a *Assembler) emitShiftCL(w64 bool, opExt uint8, reg Register) {
	emitREX(a.buf, w64, Register(opExt), reg)
	a.buf.Emit(0xD3, modRM(0x03, opExt, reg.Lo3()))
}

// emitShiftImm emits a shift/rotate by imm8.
// imm=1 uses the compact D1 form; imm>1 uses C1+ib.
func (a *Assembler) emitShiftImm(w64 bool, opExt uint8, reg Register, imm uint8) {
	emitREX(a.buf, w64, Register(opExt), reg)
	if imm == 1 {
		a.buf.Emit(0xD1, modRM(0x03, opExt, reg.Lo3()))
	} else {
		a.buf.Emit(0xC1, modRM(0x03, opExt, reg.Lo3()))
		a.buf.Emit(imm)
	}
}
