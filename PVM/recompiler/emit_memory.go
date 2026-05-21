//go:build linux && amd64

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// emitMemoryBoundsCheck emits a software bounds check (Phase 4 temporary).
// If the address in addrReg < ZZ (0x10000), jump to panic_exit.
func emitMemoryBoundsCheck(a *asm.Assembler, addrReg asm.Register, panicLabel string) {
	a.CmpRegImm32(addrReg, int32(PVM.ZZ))
	a.Jcc(asm.CondB, panicLabel)
}

// emitPanicExitAt emits an inline panic exit at the given label.
func emitPanicExitAt(a *asm.Assembler, label string, pc PVM.ProgramCounter) {
	_ = a.BindLabel(label)
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitPanic))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(pc))
	a.Jmp("exit_trampoline")
}

// ---- 4.5.2 Store instructions ----

// opcode 30-33: store_imm — store vY truncated to memory at address vX
func (c *Compiler) emitStoreImm(a *asm.Assembler, instr *PVM.InstrMeta, size int) error {
	vX, vY := twoImmFromMeta(instr)
	return c.emitStoreImmGeneric(a, instr.PC, uint32(vX), vY, size)
}

// emitStoreImmGeneric stores a compile-time-known value to a compile-time-known address.
// For 8-byte stores, temporarily spills PVM T0 to the control region as a second scratch.
func (c *Compiler) emitStoreImmGeneric(a *asm.Assembler, pc PVM.ProgramCounter, addr uint32, val uint64, size int) error {
	panicLabel := fmt.Sprintf("panic_%d", pc)
	doneLabel := fmt.Sprintf("done_%d", pc)

	if addr < 0x80000000 {
		a.MovImm32ToReg(RegScratch, int32(addr))
	} else {
		a.MovImm64ToReg(RegScratch, uint64(addr))
	}
	emitMemoryBoundsCheck(a, RegScratch, panicLabel)
	a.AddRegReg(RegScratch, RegGuestBase)

	if size < 8 {
		switch size {
		case 1:
			a.Buffer().Emit(0xC6)
			a.Buffer().Emit(modRMByte(0x00, 0, RegScratch.Lo3()))
			a.Buffer().Emit(byte(val))
		case 2:
			a.Buffer().Emit(0x66, 0xC7)
			a.Buffer().Emit(modRMByte(0x00, 0, RegScratch.Lo3()))
			a.Buffer().Emit(byte(val), byte(val>>8))
		case 4:
			a.Buffer().Emit(0xC7)
			a.Buffer().Emit(modRMByte(0x00, 0, RegScratch.Lo3()))
			a.Buffer().EmitInt32LE(int32(uint32(val)))
		}
	} else {
		spillReg := PVMToX86[2]
		a.MovRegToMem(RegGuestBase, regOffset(2), spillReg)
		a.MovImm64ToReg(spillReg, val)
		a.MovRegToMem(RegScratch, 0, spillReg)
		a.MovMemToReg(spillReg, RegGuestBase, regOffset(2))
	}

	a.Jmp(doneLabel)
	emitPanicExitAt(a, panicLabel, pc)
	_ = a.BindLabel(doneLabel)

	emitRecordMemAccessImmVal(a, addr, val)
	return nil
}

// opcode 59-62: store — store truncated Reg[rA] to memory at address vX
func (c *Compiler) emitStore(a *asm.Assembler, instr *PVM.InstrMeta, size int) error {
	pc := instr.PC
	xReg, vX := oneRegImmFromMeta(instr)

	panicLabel := fmt.Sprintf("panic_%d", pc)
	doneLabel := fmt.Sprintf("done_%d", pc)

	emitDirectMemAddr(a, vX)

	emitMemoryBoundsCheck(a, RegScratch, panicLabel)
	a.AddRegReg(RegScratch, RegGuestBase)

	switch size {
	case 1:
		a.StoreByte(RegScratch, 0, xReg)
	case 2:
		a.StoreWord(RegScratch, 0, xReg)
	case 4:
		a.StoreDword(RegScratch, 0, xReg)
	case 8:
		a.StoreQword(RegScratch, 0, xReg)
	}

	emitRecordMemAccessImm(a, uint32(vX), xReg)

	a.Jmp(doneLabel)
	emitPanicExitAt(a, panicLabel, pc)
	_ = a.BindLabel(doneLabel)
	return nil
}

// opcode 70-73: store_imm_ind — store vY to memory at Reg[rA]+vX
func (c *Compiler) emitStoreImmInd(a *asm.Assembler, instr *PVM.InstrMeta, size int) error {
	pc := instr.PC
	xReg, vX, vY := oneRegTwoImmFromMeta(instr)

	panicLabel := fmt.Sprintf("panic_%d", pc)
	doneLabel := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, xReg)
	emitAddUint64ToReg(a, RegScratch, vX)
	a.ZeroExtend32(RegScratch)

	emitRecordGuestAddrFromScratch(a)

	emitMemoryBoundsCheck(a, RegScratch, panicLabel)
	a.AddRegReg(RegScratch, RegGuestBase)

	if size < 8 {
		switch size {
		case 1:
			a.Buffer().Emit(0xC6)
			a.Buffer().Emit(modRMByte(0x00, 0, RegScratch.Lo3()))
			a.Buffer().Emit(byte(vY))
		case 2:
			a.Buffer().Emit(0x66, 0xC7)
			a.Buffer().Emit(modRMByte(0x00, 0, RegScratch.Lo3()))
			a.Buffer().Emit(byte(vY), byte(vY>>8))
		case 4:
			a.Buffer().Emit(0xC7)
			a.Buffer().Emit(modRMByte(0x00, 0, RegScratch.Lo3()))
			a.Buffer().EmitInt32LE(int32(uint32(vY)))
		}
	} else {
		spillReg := PVMToX86[2]
		a.MovRegToMem(RegGuestBase, regOffset(2), spillReg)
		a.MovImm64ToReg(spillReg, vY)
		a.MovRegToMem(RegScratch, 0, spillReg)
		a.MovMemToReg(spillReg, RegGuestBase, regOffset(2))
	}

	emitRecordMemValImm(a, vY)

	a.Jmp(doneLabel)
	emitPanicExitAt(a, panicLabel, pc)
	_ = a.BindLabel(doneLabel)
	return nil
}

// opcode 120-123: store_ind — store truncated Reg[rA] to memory at Reg[rB]+vX
func (c *Compiler) emitStoreInd(a *asm.Assembler, instr *PVM.InstrMeta, size int) error {
	pc := instr.PC
	aReg, bReg, vX := twoRegImmFromMeta(instr)

	panicLabel := fmt.Sprintf("panic_%d", pc)
	doneLabel := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	emitAddUint64ToReg(a, RegScratch, vX)
	a.ZeroExtend32(RegScratch)

	emitRecordGuestAddrFromScratch(a)

	emitMemoryBoundsCheck(a, RegScratch, panicLabel)
	a.AddRegReg(RegScratch, RegGuestBase)

	switch size {
	case 1:
		a.StoreByte(RegScratch, 0, aReg)
	case 2:
		a.StoreWord(RegScratch, 0, aReg)
	case 4:
		a.StoreDword(RegScratch, 0, aReg)
	case 8:
		a.StoreQword(RegScratch, 0, aReg)
	}

	emitRecordMemValFromReg(a, aReg)

	a.Jmp(doneLabel)
	emitPanicExitAt(a, panicLabel, pc)
	_ = a.BindLabel(doneLabel)
	return nil
}

// ---- 4.5.3 Load instructions ----

// opcode 52-58: load — load from memory at address vX into Reg[rA]
func (c *Compiler) emitLoad(a *asm.Assembler, instr *PVM.InstrMeta, size int, signed bool) error {
	pc := instr.PC
	xReg, vX := oneRegImmFromMeta(instr)

	panicLabel := fmt.Sprintf("panic_%d", pc)
	doneLabel := fmt.Sprintf("done_%d", pc)

	emitDirectMemAddr(a, vX)

	emitMemoryBoundsCheck(a, RegScratch, panicLabel)
	a.AddRegReg(RegScratch, RegGuestBase)

	emitMemLoadFromAddr(a, xReg, RegScratch, size, signed)

	emitRecordMemAccessImm(a, uint32(vX), xReg)

	a.Jmp(doneLabel)
	emitPanicExitAt(a, panicLabel, pc)
	_ = a.BindLabel(doneLabel)
	return nil
}

// opcode 124-130: load_ind — load from memory at Reg[rB]+vX into Reg[rA]
func (c *Compiler) emitLoadInd(a *asm.Assembler, instr *PVM.InstrMeta, size int, signed bool) error {
	pc := instr.PC
	aReg, bReg, vX := twoRegImmFromMeta(instr)

	panicLabel := fmt.Sprintf("panic_%d", pc)
	doneLabel := fmt.Sprintf("done_%d", pc)

	a.MovRegToReg(RegScratch, bReg)
	emitAddUint64ToReg(a, RegScratch, vX)
	a.ZeroExtend32(RegScratch)

	emitRecordGuestAddrFromScratch(a)

	emitMemoryBoundsCheck(a, RegScratch, panicLabel)
	a.AddRegReg(RegScratch, RegGuestBase)

	emitMemLoadFromAddr(a, aReg, RegScratch, size, signed)

	emitRecordMemValFromReg(a, aReg)

	a.Jmp(doneLabel)
	emitPanicExitAt(a, panicLabel, pc)
	_ = a.BindLabel(doneLabel)
	return nil
}

// emitMemLoadFromAddr loads a value from [addrReg] into dst with proper sign/zero extension.
func emitMemLoadFromAddr(a *asm.Assembler, dst, addrReg asm.Register, size int, signed bool) {
	switch size {
	case 1:
		if signed {
			a.LoadSignedByte(dst, addrReg, 0)
		} else {
			a.LoadByte(dst, addrReg, 0)
		}
	case 2:
		if signed {
			a.LoadSignedWord(dst, addrReg, 0)
		} else {
			a.LoadWord(dst, addrReg, 0)
		}
	case 4:
		if signed {
			a.LoadSignedDword(dst, addrReg, 0)
		} else {
			a.LoadDword(dst, addrReg, 0)
		}
	case 8:
		a.LoadQword(dst, addrReg, 0)
	}
}

func modRMByte(mod, reg, rm uint8) byte {
	return (mod << 6) | (reg << 3) | rm
}

// emitDirectMemAddr sets RegScratch to a guest memory pointer for direct load/store
// opcodes (52–62), where the address is uint32(vX) only.
func emitDirectMemAddr(a *asm.Assembler, vX uint64) {
	addr := uint32(vX)
	if addr < 0x80000000 {
		a.MovImm32ToReg(RegScratch, int32(addr))
	} else {
		a.MovImm64ToReg(RegScratch, uint64(addr))
	}
}
