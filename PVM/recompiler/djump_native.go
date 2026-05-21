//go:build linux && amd64

package recompiler

import (
	"fmt"
	"unsafe"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// djumpSupport holds read-only jump-table/bitmask rodata and a PC→native dispatch
// table updated as blocks are compiled. Native djump resolves jump-table addresses
// without exiting to Go; on dispatch miss it falls back to emitExitToPC.
type djumpSupport struct {
	tableLen   uint32
	tableSize  uint32
	maxAddr    uint32
	bitmaskLen uint32
	dispatch   []uintptr
}

func (c *Compiler) ensureDjumpSupport() error {
	if c.djump != nil {
		return nil
	}

	jt := c.program.JumpTable
	bm := c.program.Bitmasks
	if c.ctx.executableMem == nil {
		return fmt.Errorf("executable memory not initialized")
	}

	blob := make([]byte, len(jt.Data)+len(bm))
	copy(blob, jt.Data)
	copy(blob[len(jt.Data):], bm)

	em := c.ctx.executableMem
	offset, err := em.Write(blob)
	if err != nil {
		return fmt.Errorf("write djump rodata: %w", err)
	}

	dispatch := make([]uintptr, len(bm))
	c.djump = &djumpSupport{
		tableLen:   jt.Length,
		tableSize:  jt.Size,
		maxAddr:    jt.Size * PVM.ZA,
		bitmaskLen: uint32(len(bm)),
		dispatch:   dispatch,
	}

	tableAddr := em.GetPtr(offset)
	bitmaskAddr := em.GetPtr(offset + len(jt.Data))
	dispatchBase := uintptr(unsafe.Pointer(&dispatch[0]))
	c.ctx.setDjumpPointers(tableAddr, bitmaskAddr, dispatchBase)
	return nil
}

func (ctx *JITContext) setDjumpPointers(table, bitmask, dispatch uintptr) {
	*(*uintptr)(ctx.controlPtr(OffsetDjumpTable)) = table
	*(*uintptr)(ctx.controlPtr(OffsetDjumpBitmask)) = bitmask
	*(*uintptr)(ctx.controlPtr(OffsetDjumpDispatch)) = dispatch
}

func (c *Compiler) registerDispatch(block *CompiledBlock) {
	if c.djump == nil {
		return
	}
	pc := int(block.PVMStartPC)
	if pc >= 0 && pc < len(c.djump.dispatch) {
		c.djump.dispatch[pc] = block.NativeAddr
	}
}

func (c *Compiler) emitDjumpPanic(a *asm.Assembler, instrPC PVM.ProgramCounter) {
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitPanic))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(instrPC))
	a.Jmp("exit_trampoline")
}

func (c *Compiler) emitDjumpMiss(a *asm.Assembler, destReg asm.Register) {
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitPC), destReg)
	a.MovMemImm32(RegGuestBase, -int32(OffsetExitReason), 0)
	a.Jmp("exit_trampoline")
}

func (c *Compiler) emitLoadJumpEntry(a *asm.Assembler, tablePtr asm.Register, entryLen uint32) {
	// ReadUintFixed little-endian, 1..8 octets (Graypaper E_z(j)).
	switch entryLen {
	case 1:
		a.LoadByte(RegScratch, tablePtr, 0)
	case 2:
		a.LoadWord(RegScratch, tablePtr, 0)
	case 4:
		a.LoadDword(RegScratch, tablePtr, 0)
	case 8:
		a.LoadQword(RegScratch, tablePtr, 0)
	case 3, 5, 6, 7:
		// tablePtr is usually RegScratch; keep it at a fixed stack slot while accumulating.
		const ptrStackSlot = int32(8) // one accum push sits above the saved ptr
		a.Push(tablePtr)
		a.MovMemToReg(RegScratch, asm.RSP, 0)
		a.LoadByte(RegScratch, RegScratch, 0)
		for i := 1; i < int(entryLen); i++ {
			a.Push(RegScratch)
			a.MovMemToReg(RegScratch, asm.RSP, ptrStackSlot)
			a.AddRegImm32(RegScratch, int32(i))
			a.LoadByte(RegScratch, RegScratch, 0)
			a.ShlRegImm(RegScratch, uint8(i*8))
			a.OrRegMem(RegScratch, asm.RSP, 0)
			a.AddRegImm32(asm.RSP, 8)
		}
		a.AddRegImm32(asm.RSP, 8)
	default:
		panic(fmt.Sprintf("unsupported jump table entry length %d", entryLen))
	}
}

// emitDjumpNative resolves a jump-table address in targetReg (32-bit) and either
// jumps directly to a compiled block or exits CONTINUE with the resolved PC.
func (c *Compiler) emitDjumpNative(a *asm.Assembler, targetReg asm.Register, instrPC PVM.ProgramCounter) error {
	jt := c.program.JumpTable
	if jt.Length == 0 || jt.Size == 0 || len(jt.Data) == 0 {
		emitDjumpExit(a, targetReg)
		return nil
	}

	if err := c.ensureDjumpSupport(); err != nil {
		return err
	}

	meta := c.djump
	panicLabel := fmt.Sprintf("djump_panic_%d", instrPC)
	missLabel := fmt.Sprintf("djump_miss_%d", instrPC)

	if targetReg != RegScratch {
		a.MovRegToReg(RegScratch, targetReg)
	}
	a.Push(RegScratch) // preserve jump address for the whole resolve sequence

	// a == 0
	a.MovMemToReg(RegScratch, asm.RSP, 0)
	a.CmpReg32Imm32(RegScratch, 0)
	a.Jcc(asm.CondEQ, panicLabel)

	// misaligned (ZA == 2)
	a.MovMemToReg(RegScratch, asm.RSP, 0)
	a.AndRegImm32(RegScratch, 1)
	a.Jcc(asm.CondNE, panicLabel)

	// a > jumpTable.Size * ZA
	a.MovMemToReg(RegScratch, asm.RSP, 0)
	a.CmpReg32Imm32(RegScratch, int32(meta.maxAddr))
	a.Jcc(asm.CondA, panicLabel)

	// index = (a >> 1) - 1
	a.MovMemToReg(RegScratch, asm.RSP, 0)
	a.ShrRegImm(RegScratch, 1)
	a.SubRegImm32(RegScratch, 1)

	// index >= tableSize
	a.CmpReg32Imm32(RegScratch, int32(meta.tableSize))
	a.Jcc(asm.CondAE, panicLabel)

	// byte offset = index * tableLen
	switch meta.tableLen {
	case 1:
		// index is already the byte offset
	case 2:
		a.ShlRegImm(RegScratch, 1)
	case 4:
		a.ShlRegImm(RegScratch, 2)
	case 8:
		a.ShlRegImm(RegScratch, 3)
	default:
		a.ImulRegImm32(RegScratch, RegScratch, int32(meta.tableLen))
	}

	// dest PC = jumpTable[offset]
	a.Push(RegScratch)
	a.MovMemToReg(RegScratch, RegGuestBase, -int32(OffsetDjumpTable))
	a.AddRegMem(RegScratch, asm.RSP, 0)
	c.emitLoadJumpEntry(a, RegScratch, meta.tableLen)
	a.AddRegImm32(asm.RSP, 8)

	// dest >= len(bitmask)
	a.CmpReg32Imm32(RegScratch, int32(meta.bitmaskLen))
	a.Jcc(asm.CondAE, panicLabel)

	// bitmask[dest] == 0x03 (basic-block start)
	a.Push(RegScratch)
	a.MovMemToReg(RegScratch, RegGuestBase, -int32(OffsetDjumpBitmask))
	a.AddRegMem(RegScratch, asm.RSP, 0)
	a.LoadByte(RegScratch, RegScratch, 0)
	a.CmpReg32Imm32(RegScratch, 3)
	a.Pop(RegScratch)
	a.Jcc(asm.CondNE, panicLabel)

	// dispatch[dest]: direct native JMP on hit, emitExitToPC on miss
	a.Push(RegScratch) // dest PC
	a.ShlRegImm(RegScratch, 3)
	a.Push(RegScratch) // dest * 8

	a.MovMemToReg(RegScratch, RegGuestBase, -int32(OffsetDjumpDispatch))
	a.AddRegMem(RegScratch, asm.RSP, 0)
	a.LoadQword(RegScratch, RegScratch, 0)
	a.AddRegImm32(asm.RSP, 8)
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondEQ, missLabel)
	a.AddRegImm32(asm.RSP, 16) // drop saved jump addr + dest PC
	a.JmpReg(RegScratch)

	_ = a.BindLabel(missLabel)
	a.Pop(RegScratch)         // dest PC
	a.AddRegImm32(asm.RSP, 8) // drop saved jump addr
	c.emitDjumpMiss(a, RegScratch)

	_ = a.BindLabel(panicLabel)
	a.AddRegImm32(asm.RSP, 8) // drop saved jump addr
	c.emitDjumpPanic(a, instrPC)
	return nil
}
