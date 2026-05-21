//go:build linux && amd64

package recompiler

import (
	"runtime"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
	x86_signal_linux "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/x86signal"
	"golang.org/x/sys/unix"
)

func init() {
	x86_signal_linux.SetupSignalHandler()
}

func executeBlockTest(t *testing.T, ctx *JITContext, block *CompiledBlock) PVM.ExitReason {
	t.Helper()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	em := ctx.executableMem

	trampolineAsm := asm.NewAssembler()
	EmitEntryTrampoline(trampolineAsm)
	trampolineCode, err := trampolineAsm.Finalize()
	if err != nil {
		t.Fatalf("finalize entry trampoline: %v", err)
	}
	trampolineOff, err := em.Write(trampolineCode)
	if err != nil {
		t.Fatalf("write entry trampoline: %v", err)
	}

	trampolineAddr := em.GetPtr(trampolineOff)

	codeStart := em.GetPtr(0)
	codeEnd := codeStart + uintptr(em.Used())
	x86_signal_linux.SetFaultWindow(ctx.GuestBase(), codeStart, codeEnd)
	defer x86_signal_linux.ClearFaultWindow()
	callNative(ctx.GuestBase(), block.NativeAddr, trampolineAddr)

	return ctx.ReadExitReason()
}

func setupTestContext(t *testing.T) (*JITContext, func()) {
	t.Helper()

	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}

	em, err := NewExecutableMemory(0)
	if err != nil {
		ctx.Close()
		t.Fatalf("NewExecutableMemory: %v", err)
	}
	ctx.SetExecutableMemory(em)
	ctx.heapLimit = GuestMemorySize

	cleanup := func() {
		em.Close()
		ctx.Close()
	}
	return ctx, cleanup
}

func compileTestBlock(t *testing.T, ctx *JITContext, instBytes []byte, boundaries []int) *CompiledBlock {
	t.Helper()

	blob := buildBlobExact(instBytes, boundaries)
	prog, exitReason := PVM.DeBlobProgramCode(blob)
	if exitReason != PVM.ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", exitReason)
	}

	cache := NewCodeCache()
	compiler := newTestCompiler(&prog, ctx, cache)
	block, err := compiler.compileBasicBlock(0)
	if err != nil {
		t.Fatalf("CompileBasicBlock: %v", err)
	}
	return block
}

// testCompiler is a minimal in-package compiler used only for signal handler tests.
// It avoids importing the jit package (which would be circular).
type testCompiler struct {
	program *PVM.Program
	cache   *CodeCache
	ctx     *JITContext
}

func newTestCompiler(program *PVM.Program, ctx *JITContext, cache *CodeCache) *testCompiler {
	return &testCompiler{program: program, cache: cache, ctx: ctx}
}

func (tc *testCompiler) compileBasicBlock(startPC PVM.ProgramCounter) (*CompiledBlock, error) {
	a := asm.NewAssembler()

	blockMeta := tc.program.LookupBlock(startPC)
	if blockMeta == nil {
		idx := tc.program.InstrIdxAt[startPC]
		if idx < 0 {
			return nil, nil
		}
		startIdx := int(idx)
		endIdx := startIdx
		for endIdx < len(tc.program.Instrs) {
			if PVM.IsBlockTerminator(tc.program.Instrs[endIdx].Opcode) {
				endIdx++
				break
			}
			endIdx++
		}
		blockMeta = &PVM.BlockMeta{
			StartPC:    startPC,
			EndPC:      tc.program.Instrs[endIdx-1].PC,
			InstrStart: startIdx,
			InstrEnd:   endIdx,
			GasCost:    PVM.Gas(endIdx - startIdx),
		}
	}

	instrs := tc.program.Instrs[blockMeta.InstrStart:blockMeta.InstrEnd]
	lastInstr := &instrs[len(instrs)-1]
	fallthroughPC := lastInstr.PC + PVM.ProgramCounter(lastInstr.SkipLen) + 1

	for i := range instrs {
		instr := &instrs[i]
		a.MovMemToReg(RegScratch, RegGuestBase, -int32(OffsetGas))
		a.TestRegReg(RegScratch, RegScratch)
		a.SubMemImm32(RegGuestBase, -int32(OffsetGas), 1)
		emitInstrForTest(a, instr)
	}

	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(fallthroughPC))
	a.MovMemImm32(RegGuestBase, -int32(OffsetExitReason), 0)
	a.Jmp("exit_trampoline")
	EmitExitTrampoline(a)

	code, err := a.Finalize()
	if err != nil {
		return nil, err
	}

	em := tc.ctx.executableMem
	offset, err := em.Write(code)
	if err != nil {
		return nil, err
	}

	return &CompiledBlock{
		PVMStartPC: blockMeta.StartPC,
		PVMEndPC:   fallthroughPC,
		NativeAddr: em.GetPtr(offset),
		NativeSize: len(code),
	}, nil
}

func emitInstrForTest(a *asm.Assembler, instr *PVM.InstrMeta) {
	switch instr.Opcode {
	case 0: // trap
		a.MovImm64ToReg(RegScratch, uint64(PVM.ExitPanic))
		a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
		a.Jmp("exit_trampoline")
	case 59: // store_u8
		xReg := PVMReg(instr.Dst)
		a.MovRegToReg(RegScratch, xReg)
		a.MovEcxEcx(RegScratch)
		a.AddRegReg(RegScratch, RegGuestBase)
		a.StoreByteRegToMem(RegScratch, 0, xReg)
	case 52: // load_u8
		xReg := PVMReg(instr.Dst)
		a.MovRegToReg(RegScratch, xReg)
		a.MovEcxEcx(RegScratch)
		a.AddRegReg(RegScratch, RegGuestBase)
		a.LoadByteFromMem(xReg, RegScratch, 0)
	}
}

func buildStoreU8Reg(rA uint8) []byte {
	return []byte{59, rA | (rA << 4)}
}

func TestSignalHandler_WriteToProtNone(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	inst := buildStoreU8Reg(0)
	instBytes := append(inst, 0x00)
	boundaries := []int{0, len(inst)}

	block := compileTestBlock(t, ctx, instBytes, boundaries)

	regs := PVM.Registers{}
	regs[0] = 0x30000
	ctx.WriteRegisters(regs)
	ctx.WriteGas(1000)
	ctx.WriteExitReason(0)

	exitReason := executeBlockTest(t, ctx, block)

	if exitReason.GetReasonType() != PVM.PAGE_FAULT {
		t.Fatalf("expected PAGE_FAULT, got %v", exitReason)
	}
	faultAddr := exitReason.GetPageFaultAddress()
	if faultAddr != 0x30000 {
		t.Errorf("fault address: got 0x%08x, want 0x00030000", faultAddr)
	}
}

func TestSignalHandler_WriteToReadOnly(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	roStart := uint32(PVM.ZZ)
	roEnd := roStart + uint32(PVM.ZP)
	content := make([]byte, PVM.ZP)
	if err := ctx.mapSegment(roStart, roEnd, content, unix.PROT_READ); err != nil {
		t.Fatalf("mapSegment: %v", err)
	}

	inst := buildStoreU8Reg(0)
	instBytes := append(inst, 0x00)
	boundaries := []int{0, len(inst)}

	block := compileTestBlock(t, ctx, instBytes, boundaries)

	regs := PVM.Registers{}
	regs[0] = uint64(roStart)
	ctx.WriteRegisters(regs)
	ctx.WriteGas(1000)
	ctx.WriteExitReason(0)

	exitReason := executeBlockTest(t, ctx, block)

	if exitReason.GetReasonType() != PVM.PAGE_FAULT {
		t.Fatalf("expected PAGE_FAULT for write to R/O, got %v", exitReason)
	}
}

func TestSignalHandler_LegitRW(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	rwStart := uint32(2 * PVM.ZZ)
	rwEnd := rwStart + uint32(PVM.ZP)
	if err := ctx.mapSegment(rwStart, rwEnd, nil, unix.PROT_READ|unix.PROT_WRITE); err != nil {
		t.Fatalf("mapSegment: %v", err)
	}

	inst := buildStoreU8Reg(0)
	instBytes := append(inst, 0x00)
	boundaries := []int{0, len(inst)}

	block := compileTestBlock(t, ctx, instBytes, boundaries)

	regs := PVM.Registers{}
	regs[0] = uint64(rwStart) + 10
	ctx.WriteRegisters(regs)
	ctx.WriteGas(1000)
	ctx.WriteExitReason(0)

	exitReason := executeBlockTest(t, ctx, block)

	if exitReason.GetReasonType() == PVM.PAGE_FAULT {
		t.Fatalf("unexpected PAGE_FAULT on RW region")
	}
	if exitReason.GetReasonType() != PVM.PANIC {
		t.Fatalf("expected PANIC from trap, got %v", exitReason)
	}

	written := ctx.guestMem[rwStart+10]
	wantByte := byte(uint64(rwStart) + 10)
	if written != wantByte {
		t.Errorf("memory[0x%x] = 0x%02x, want 0x%02x", rwStart+10, written, wantByte)
	}
}

func TestSignalHandler_ReservedZone(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	inst := buildStoreU8Reg(0)
	instBytes := append(inst, 0x00)
	boundaries := []int{0, len(inst)}

	block := compileTestBlock(t, ctx, instBytes, boundaries)

	regs := PVM.Registers{}
	regs[0] = 0x100
	ctx.WriteRegisters(regs)
	ctx.WriteGas(1000)
	ctx.WriteExitReason(0)

	exitReason := executeBlockTest(t, ctx, block)

	if exitReason.GetReasonType() != PVM.PANIC {
		t.Fatalf("expected PANIC for addr < ZZ, got %v", exitReason)
	}
}

func TestSignalHandler_RegisterPreservation(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	inst := buildStoreU8Reg(2)
	instBytes := append(inst, 0x00)
	boundaries := []int{0, len(inst)}

	block := compileTestBlock(t, ctx, instBytes, boundaries)

	initRegs := PVM.Registers{
		0:  0xAAAA_AAAA_AAAA_AAAA,
		1:  0xBBBB_BBBB_BBBB_BBBB,
		2:  0x40000,
		3:  0xDDDD_DDDD_DDDD_DDDD,
		4:  0xEEEE_EEEE_EEEE_EEEE,
		5:  0x1111_1111_1111_1111,
		6:  0x2222_2222_2222_2222,
		7:  0x3333_3333_3333_3333,
		8:  0x4444_4444_4444_4444,
		9:  0x5555_5555_5555_5555,
		10: 0x6666_6666_6666_6666,
		11: 0x7777_7777_7777_7777,
		12: 0x8888_8888_8888_8888,
	}
	ctx.WriteRegisters(initRegs)
	ctx.WriteGas(1000)
	ctx.WriteExitReason(0)

	exitReason := executeBlockTest(t, ctx, block)
	if exitReason.GetReasonType() != PVM.PAGE_FAULT {
		t.Fatalf("expected PAGE_FAULT, got %v", exitReason)
	}

	gotRegs := ctx.ReadRegisters()
	for i := range 13 {
		if gotRegs[i] != initRegs[i] {
			t.Errorf("Reg[%d]: got 0x%016x, want 0x%016x", i, gotRegs[i], initRegs[i])
		}
	}
}

func TestSbrk_ExpandHeap(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	rwStart := uint32(2 * PVM.ZZ)
	rwEnd := rwStart + uint32(PVM.ZP)
	if err := ctx.mapSegment(rwStart, rwEnd, nil, unix.PROT_READ|unix.PROT_WRITE); err != nil {
		t.Fatalf("mapSegment: %v", err)
	}
	ctx.WriteHeapPointer(uint64(rwEnd))

	amount := uint64(2 * PVM.ZP)
	regs := PVM.Registers{}
	regs[7] = amount
	ctx.WriteRegisters(regs)

	result := HandleSbrk(ctx, 0, 7)
	if result != PVM.ExitContinue {
		t.Fatalf("HandleSbrk: got %v, want ExitContinue", result)
	}

	gotHP := ctx.ReadHeapPointer()
	wantHP := uint64(rwEnd) + amount
	if gotHP != wantHP {
		t.Errorf("heap pointer: got 0x%x, want 0x%x", gotHP, wantHP)
	}

	gotRegs := ctx.ReadRegisters()
	if gotRegs[0] != wantHP {
		t.Errorf("Reg[rD]: got 0x%x, want 0x%x", gotRegs[0], wantHP)
	}

	newAddr := uint32(rwEnd) + uint32(PVM.ZP)
	ctx.guestMem[newAddr] = 0x42
	if ctx.guestMem[newAddr] != 0x42 {
		t.Error("new heap page not writable")
	}
}

func TestSbrk_ZeroAmount(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	hp := uint64(0x30000)
	ctx.WriteHeapPointer(hp)

	regs := PVM.Registers{}
	regs[7] = 0
	ctx.WriteRegisters(regs)

	result := HandleSbrk(ctx, 0, 7)
	if result != PVM.ExitContinue {
		t.Fatalf("HandleSbrk: got %v, want ExitContinue", result)
	}

	gotRegs := ctx.ReadRegisters()
	if gotRegs[0] != hp {
		t.Errorf("Reg[rD]: got 0x%x, want 0x%x", gotRegs[0], hp)
	}
}

func TestGoRuntimeCoexistence(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	instBytes := []byte{0}
	boundaries := []int{0}
	block := compileTestBlock(t, ctx, instBytes, boundaries)

	ctx.WriteRegisters(PVM.Registers{})
	ctx.WriteGas(1000)
	ctx.WriteExitReason(0)

	exitReason := executeBlockTest(t, ctx, block)
	if exitReason.GetReasonType() != PVM.PANIC {
		t.Fatalf("expected PANIC from trap, got %v", exitReason)
	}

	var slices [][]byte
	for i := 0; i < 1000; i++ {
		slices = append(slices, make([]byte, 1024))
	}
	_ = slices
}
