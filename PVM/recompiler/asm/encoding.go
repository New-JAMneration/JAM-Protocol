package asm

// rexByte builds a REX prefix: 0100WRXB.
//   - w: 64-bit operand size
//   - r: extends ModR/M reg field (register in bits 5:3)
//   - x: extends SIB index field
//   - b: extends ModR/M rm field or SIB base field
func rexByte(w, r, x, b bool) byte {
	v := byte(0x40)
	if w {
		v |= 0x08
	}
	if r {
		v |= 0x04
	}
	if x {
		v |= 0x02
	}
	if b {
		v |= 0x01
	}
	return v
}

// emitREX emits a REX prefix if needed. w controls 64-bit operand size.
// reg is the register encoded in ModR/M.reg, rm is the register in ModR/M.rm.
func emitREX(buf *CodeBuffer, w bool, reg, rm Register) {
	if w || reg.IsExtended() || rm.IsExtended() {
		buf.Emit(rexByte(w, reg.IsExtended(), false, rm.IsExtended()))
	}
}

// emitREXSIB emits a REX prefix for instructions using SIB (base+index).
func emitREXSIB(buf *CodeBuffer, w bool, reg, index, base Register) {
	if w || reg.IsExtended() || index.IsExtended() || base.IsExtended() {
		buf.Emit(rexByte(w, reg.IsExtended(), index.IsExtended(), base.IsExtended()))
	}
}

// modRM builds a ModR/M byte: [mod:2][reg:3][rm:3].
func modRM(mod, reg, rm uint8) byte {
	return (mod << 6) | ((reg & 0x07) << 3) | (rm & 0x07)
}

// sib builds a SIB byte: [scale:2][index:3][base:3].
func sib(scale, index, base uint8) byte {
	return (scale << 6) | ((index & 0x07) << 3) | (base & 0x07)
}

// fitsInt8 reports whether v fits in a signed 8-bit immediate.
func fitsInt8(v int32) bool {
	return v >= -128 && v <= 127
}

// needRexForLow8Reg reports whether a REX prefix is required to encode the low
// 8 bits of reg in MOV r/m8,r8 (0x88) or MOV r8,r/m8 (0x8A). Without REX,
// ModR/M.reg values 4–7 mean AH/CH/DH/BH; with REX they mean SPL/BPL/SIL/DIL.
func needRexForLow8Reg(r Register) bool {
	return r.IsExtended() || r.Lo3() >= 4
}

// emitMemOp emits [ModR/M] (+ SIB if needed) (+ displacement) for [base + disp].
// regField is the value for the reg bits in ModR/M (register or opcode extension).
// Special-cases: RSP/R12 as base requires SIB; RBP/R13 as base with 0 disp requires disp8=0.
func emitMemOp(buf *CodeBuffer, regField uint8, base Register, disp int32) {
	rm := base.Lo3()

	needsSIB := rm == RSP.Lo3() // RSP(4) or R12 — ModR/M rm=100 signals SIB

	if disp == 0 && rm != RBP.Lo3() {
		// mod=00: [base] (no displacement)
		if needsSIB {
			buf.Emit(modRM(0x00, regField, 0x04))
			buf.Emit(sib(0x00, 0x04, base.Lo3())) // index=RSP(none), base=base
		} else {
			buf.Emit(modRM(0x00, regField, rm))
		}
	} else if fitsInt8(disp) {
		// mod=01: [base + disp8]
		if needsSIB {
			buf.Emit(modRM(0x01, regField, 0x04))
			buf.Emit(sib(0x00, 0x04, base.Lo3()))
		} else {
			buf.Emit(modRM(0x01, regField, rm))
		}
		buf.Emit(byte(int8(disp)))
	} else {
		// mod=10: [base + disp32]
		if needsSIB {
			buf.Emit(modRM(0x02, regField, 0x04))
			buf.Emit(sib(0x00, 0x04, base.Lo3()))
		} else {
			buf.Emit(modRM(0x02, regField, rm))
		}
		buf.EmitInt32LE(disp)
	}
}

// emitMemOpSIB emits [ModR/M + SIB + disp] for [base + index*1 + disp].
func emitMemOpSIB(buf *CodeBuffer, regField uint8, base, index Register, disp int32) {
	if disp == 0 && base.Lo3() != RBP.Lo3() {
		buf.Emit(modRM(0x00, regField, 0x04))
		buf.Emit(sib(0x00, index.Lo3(), base.Lo3()))
	} else if fitsInt8(disp) {
		buf.Emit(modRM(0x01, regField, 0x04))
		buf.Emit(sib(0x00, index.Lo3(), base.Lo3()))
		buf.Emit(byte(int8(disp)))
	} else {
		buf.Emit(modRM(0x02, regField, 0x04))
		buf.Emit(sib(0x00, index.Lo3(), base.Lo3()))
		buf.EmitInt32LE(disp)
	}
}
