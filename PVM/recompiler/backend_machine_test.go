//go:build linux && amd64 && cgo

package recompiler

import (
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"golang.org/x/sys/unix"
)

const machineGas PVM.Gas = 10_000

type machineResult struct {
	reason     PVM.ExitReason
	pc         PVM.ProgramCounter
	gas        PVM.Gas
	gasCharged bool
	regs       PVM.Registers
}

func TestInterpreterRecompilerMachineHaltTrap(t *testing.T) {
	blob := buildBlobExact([]byte{0}, []int{0}) // trap
	interp, rec := runBothBackends(t, blob, PVM.Registers{}, nil, machineGas)
	if interp.reason != PVM.ExitPanic || rec.reason != PVM.ExitPanic {
		t.Fatalf("trap: interp=%v rec=%v, want PANIC", interp.reason, rec.reason)
	}
	if interp.gas != rec.gas {
		t.Fatalf("gas: interp=%d rec=%d", interp.gas, rec.gas)
	}
}

func TestInterpreterRecompilerJumpIndHaltSentinel(t *testing.T) {
	// unlikely; jump_ind r0+0 — jump_ind at PC 1 so interpreter's HALT PC
	// zeroing (BlockBasedInvokeDecodedBlocks returns 0) is distinguishable
	// from the recompiler storing instr.PC.
	blob := buildBlobExact([]byte{2, 50, 0}, []int{0, 1})
	var regs PVM.Registers
	regs[0] = 0xffff0000
	interp, rec := runBothBackends(t, blob, regs, nil, machineGas)
	if interp.reason != PVM.ExitHalt || rec.reason != PVM.ExitHalt {
		t.Fatalf("halt: interp=%v rec=%v, want HALT", interp.reason, rec.reason)
	}
	const jumpIndPC PVM.ProgramCounter = 1
	if rec.pc != jumpIndPC {
		t.Fatalf("recompiler HALT pc=%d, want jump_ind pc=%d", rec.pc, jumpIndPC)
	}
	// Interpreter invoke loop returns 0 on HALT/PANIC; that is not the
	// PAGE_FAULT ExitPC bug. Recompiler must still report instr.PC.
	if interp.pc != 0 {
		t.Fatalf("interpreter HALT pc=%d, want 0 (invoke loop zeroes HALT pc)", interp.pc)
	}
}

func TestInterpreterRecompilerSamePageFaultPayload(t *testing.T) {
	// store_u8 r0, 0x10000; trap — above ZZ, unmapped.
	blob := buildBlobExact([]byte{59, 0, 0x00, 0x00, 0x01, 0}, []int{0, 5})
	interp, rec := runBothBackends(t, blob, PVM.Registers{}, nil, machineGas)
	if interp.reason.GetReasonType() != PVM.PAGE_FAULT || rec.reason.GetReasonType() != PVM.PAGE_FAULT {
		t.Fatalf("want PAGE_FAULT: interp=%v rec=%v", interp.reason, rec.reason)
	}
	const wantAddr uint32 = 0x10000
	if got := interp.reason.GetPageFaultAddress(); got != wantAddr {
		t.Fatalf("interpreter fault addr=0x%x, want 0x%x", got, wantAddr)
	}
	if got := rec.reason.GetPageFaultAddress(); got != wantAddr {
		t.Fatalf("recompiler fault addr=0x%x, want 0x%x", got, wantAddr)
	}
	if interp.pc != 0 {
		t.Fatalf("interpreter fault pc=%d, want 0 (faulting store)", interp.pc)
	}
	if rec.pc != 0 {
		t.Fatalf("recompiler fault pc=%d, want 0 (block start == faulting store)", rec.pc)
	}
}

func TestPageFaultExitPCSecondInstrCharacterization(t *testing.T) {
	// unlikely; store_u8 r0, 0x10000; trap — store is not the first instruction.
	blob := buildBlobExact([]byte{2, 59, 0, 0x00, 0x00, 0x01, 0}, []int{0, 1, 6})
	interp, rec := runBothBackends(t, blob, PVM.Registers{}, nil, machineGas)
	if interp.reason.GetReasonType() != PVM.PAGE_FAULT || rec.reason.GetReasonType() != PVM.PAGE_FAULT {
		t.Fatalf("want PAGE_FAULT: interp=%v rec=%v", interp.reason, rec.reason)
	}
	const storePC PVM.ProgramCounter = 1
	if interp.pc != storePC {
		t.Fatalf("interpreter pc=%d, want faulting store %d", interp.pc, storePC)
	}
	// Known mismatch: handler does not write ExitPC, so recompiler keeps block start.
	if rec.pc != 0 {
		t.Fatalf("characterization: recompiler pc=%d, want block start 0 (not yet instr.PC)", rec.pc)
	}
	if rec.pc == interp.pc {
		t.Fatal("characterization stale: backends now agree on PAGE_FAULT ExitPC; promote to a passing gate")
	}
}

func TestCrossPageStoreCharacterization(t *testing.T) {
	// store_u64 r0 at 0x1FFFC (last 4 bytes of page 31, first 4 of page 32).
	blob := buildBlobExact([]byte{62, 0, 0xFC, 0xFF, 0x01, 0}, []int{0, 5})
	const addr uint32 = 0x1FFFC
	page31 := addr / PVM.ZP
	setup := func(_ *testing.T, interpMem *PVM.Memory, ctx *JITContext) {
		if interpMem != nil {
			interpMem.Pages = map[uint32]*PVM.Page{
				page31: {Value: make([]byte, PVM.ZP), Access: PVM.MemoryReadWrite},
			}
		}
		if ctx != nil {
			if err := ctx.SetPageAccess(page31, unix.PROT_READ|unix.PROT_WRITE); err != nil {
				t.Fatalf("mprotect page %d: %v", page31, err)
			}
		}
	}
	var regs PVM.Registers
	regs[0] = 0x0102030405060708
	interp, rec := runBothBackends(t, blob, regs, setup, machineGas)
	if interp.reason.GetReasonType() != PVM.PAGE_FAULT {
		t.Fatalf("interpreter want PAGE_FAULT, got %v", interp.reason)
	}
	if got := interp.reason.GetPageFaultAddress(); got != addr {
		t.Fatalf("interpreter payload=0x%x, want start addr 0x%x", got, addr)
	}
	// Recompiler may report the second page via si_addr; record it, do not require equality.
	if rec.reason.GetReasonType() != PVM.PAGE_FAULT {
		t.Fatalf("recompiler want PAGE_FAULT, got %v", rec.reason)
	}
	t.Logf("cross-page characterization: interp_addr=0x%x rec_addr=0x%x interp_pc=%d rec_pc=%d",
		interp.reason.GetPageFaultAddress(), rec.reason.GetPageFaultAddress(), interp.pc, rec.pc)
}

func TestInterpreterRecompilerOOG(t *testing.T) {
	blob := buildBlobExact([]byte{2, 1}, []int{0, 1}) // unlikely; fallthrough
	interp, rec := runBothBackends(t, blob, PVM.Registers{}, nil, 0)
	if interp.reason != PVM.ExitOOG || rec.reason != PVM.ExitOOG {
		t.Fatalf("OOG: interp=%v rec=%v", interp.reason, rec.reason)
	}
	if interp.gasCharged || rec.gasCharged {
		t.Fatal("GasCharged must stay false when the block charge fails")
	}
	if interp.pc != rec.pc {
		t.Fatalf("OOG pc: interp=%d rec=%d", interp.pc, rec.pc)
	}
}

func TestInterpreterRecompilerEcalliSuffix(t *testing.T) {
	// ecalli 0; load_imm r0, 42; trap
	blob := buildBlobExact([]byte{10, 0, 51, 0, 42, 0}, []int{0, 2, 5})
	prog := mustDeblob(t, blob)

	interpRes, interp := runInterpreterUntil(t, &prog, PVM.Registers{}, nil, machineGas, 1)
	if interpRes.reason.GetReasonType() != PVM.HOST_CALL {
		t.Fatalf("interpreter first exit=%v, want HOST_CALL", interpRes.reason)
	}
	if !interpRes.gasCharged {
		t.Fatal("interpreter ecalli must preserve GasCharged")
	}

	recRes, rec := runRecompilerUntil(t, &prog, PVM.Registers{}, nil, machineGas, 1)
	if recRes.reason.GetReasonType() != PVM.HOST_CALL {
		t.Fatalf("recompiler first exit=%v, want HOST_CALL", recRes.reason)
	}
	if !recRes.gasCharged {
		t.Fatal("recompiler ecalli must preserve GasCharged")
	}
	if interpRes.pc != recRes.pc {
		t.Fatalf("host-call resume pc: interp=%d rec=%d", interpRes.pc, recRes.pc)
	}
	if interpRes.gas != recRes.gas {
		t.Fatalf("gas after ecalli: interp=%d rec=%d", interpRes.gas, recRes.gas)
	}

	interp2 := invokeInterpreter(t, interp, interpRes.pc)
	rec2 := invokeRecompiler(t, rec, recRes.pc)
	if interp2.reason != PVM.ExitPanic || rec2.reason != PVM.ExitPanic {
		t.Fatalf("suffix trap: interp=%v rec=%v", interp2.reason, rec2.reason)
	}
	if interp2.gas != rec2.gas {
		t.Fatalf("suffix must not recharge: interp gas=%d rec gas=%d", interp2.gas, rec2.gas)
	}
}

func TestInterpreterRecompilerStoreThenTrap(t *testing.T) {
	blob := buildBlobExact([]byte{59, 0, 0x00, 0x00, 0x01, 0}, []int{0, 5})
	const addr uint32 = 0x10000
	page := addr / PVM.ZP
	setup := func(_ *testing.T, interpMem *PVM.Memory, ctx *JITContext) {
		if interpMem != nil {
			interpMem.Pages = map[uint32]*PVM.Page{
				page: {Value: make([]byte, PVM.ZP), Access: PVM.MemoryReadWrite},
			}
		}
		if ctx != nil {
			if err := ctx.SetPageAccess(page, unix.PROT_READ|unix.PROT_WRITE); err != nil {
				t.Fatalf("mprotect: %v", err)
			}
		}
	}
	var regs PVM.Registers
	regs[0] = 0xab
	prog := mustDeblob(t, blob)
	interpRes, interp := runInterpreterUntil(t, &prog, regs, setup, machineGas, 1)
	recRes, rec := runRecompilerUntil(t, &prog, regs, setup, machineGas, 1)
	if interpRes.reason != PVM.ExitPanic || recRes.reason != PVM.ExitPanic {
		t.Fatalf("want trap PANIC: interp=%v rec=%v", interpRes.reason, recRes.reason)
	}
	if got := interp.Memory.Pages[page].Value[addr%PVM.ZP]; got != 0xab {
		t.Fatalf("interpreter mem[0x%x]=0x%x, want 0xab", addr, got)
	}
	if got := rec.ctx.guestMem[addr]; got != 0xab {
		t.Fatalf("recompiler mem[0x%x]=0x%x, want 0xab", addr, got)
	}
}

func TestInterpreterRecompilerBackEdgeOOG(t *testing.T) {
	// jump to PC 0 (self-loop).
	blob := buildBlobExact([]byte{40, 0}, []int{0})
	interp, rec := runBothBackends(t, blob, PVM.Registers{}, nil, 20)
	if interp.reason.GetReasonType() != PVM.OUT_OF_GAS || rec.reason.GetReasonType() != PVM.OUT_OF_GAS {
		t.Fatalf("self-loop want OOG: interp=%v rec=%v", interp.reason, rec.reason)
	}
	if interp.pc != 0 || rec.pc != 0 {
		t.Fatalf("OOG pc: interp=%d rec=%d, want 0", interp.pc, rec.pc)
	}
}

func mustDeblob(t *testing.T, blob []byte) PVM.Program {
	t.Helper()
	prog, reason := PVM.DeBlobProgramCode(blob, 0)
	if reason != PVM.ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", reason)
	}
	return prog
}

type memSetup func(t *testing.T, interpMem *PVM.Memory, ctx *JITContext)

func runBothBackends(t *testing.T, blob []byte, regs PVM.Registers, setup memSetup, gas PVM.Gas) (machineResult, machineResult) {
	t.Helper()
	prog := mustDeblob(t, blob)
	interp, _ := runInterpreterUntil(t, &prog, regs, setup, gas, 1)
	rec, _ := runRecompilerUntil(t, &prog, regs, setup, gas, 1)
	return interp, rec
}

func runInterpreterUntil(t *testing.T, prog *PVM.Program, regs PVM.Registers, setup memSetup, gas PVM.Gas, _ int) (machineResult, *PVM.Interpreter) {
	t.Helper()
	mem := &PVM.Memory{Pages: map[uint32]*PVM.Page{}}
	interp := PVM.NewInterpreter(prog, regs, mem, gas)
	if setup != nil {
		setup(t, mem, nil)
	}
	return invokeInterpreter(t, interp, 0), interp
}

func runRecompilerUntil(t *testing.T, prog *PVM.Program, regs PVM.Registers, setup memSetup, gas PVM.Gas, _ int) (machineResult, *Recompiler) {
	t.Helper()
	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	t.Cleanup(func() { _ = ctx.Close() })

	em, err := NewExecutableMemory(0)
	if err != nil {
		t.Fatalf("NewExecutableMemory: %v", err)
	}
	t.Cleanup(func() { _ = em.Close() })
	ctx.SetExecutableMemory(em)
	ctx.heapLimit = GuestMemorySize
	if setup != nil {
		setup(t, nil, ctx)
	}

	rec := NewRecompiler(prog, ctx)
	ctx.WriteRegisters(regs)
	ctx.WriteGas(gas)
	ctx.WriteGasCharged(false)
	return invokeRecompiler(t, rec, 0), rec
}

func invokeInterpreter(t *testing.T, interp *PVM.Interpreter, pc PVM.ProgramCounter) machineResult {
	t.Helper()
	reason, outPC := interp.BlockBasedInvokeDecodedBlocks(pc)
	return machineResult{
		reason:     reason,
		pc:         outPC,
		gas:        interp.Gas,
		gasCharged: interp.GasCharged,
		regs:       interp.Registers,
	}
}

func invokeRecompiler(t *testing.T, rec *Recompiler, pc PVM.ProgramCounter) machineResult {
	t.Helper()
	reason, outPC := rec.BlockBasedInvoke(pc)
	return machineResult{
		reason:     reason,
		pc:         outPC,
		gas:        rec.ctx.ReadGas(),
		gasCharged: rec.ctx.ReadGasCharged(),
		regs:       rec.ctx.ReadRegisters(),
	}
}
