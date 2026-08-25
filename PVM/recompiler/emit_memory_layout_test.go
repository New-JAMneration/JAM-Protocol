//go:build linux && amd64 && cgo

package recompiler

import (
	"encoding/binary"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"golang.org/x/sys/unix"
)

func TestMemPanicPadsArePerInstructionAndAfterHotBody(t *testing.T) {
	// store_u8 r0, 0x10000; store_u8 r0, 0; trap
	blob := buildBlobExact(
		[]byte{59, 0, 0x00, 0x00, 0x01, 59, 0, 0, 0},
		[]int{0, 5, 8},
	)
	prog := mustDeblob(t, blob)
	if len(prog.Instrs) < 2 {
		t.Fatalf("want ≥2 instructions, got %d", len(prog.Instrs))
	}
	secondPC := prog.Instrs[1].PC

	block, native := compileNativeAt(t, &prog, 0)
	targets := jbRel32Targets(native)
	if len(targets) != 2 {
		t.Fatalf("JB rel32 count=%d, want 2 (one per store)", len(targets))
	}
	if targets[0] == targets[1] {
		t.Fatal("memory ops must not share a panic pad")
	}
	hotEnd := min(targets[0], targets[1])
	if hotEnd <= 0 {
		t.Fatalf("pad targets %v should be after the hot body", targets)
	}
	_ = block

	ctx, compiler := compilerWithProgram(t, &prog)
	page := uint32(0x10000 / PVM.ZP)
	if err := ctx.SetPageAccess(page, unix.PROT_READ|unix.PROT_WRITE); err != nil {
		t.Fatalf("mprotect: %v", err)
	}
	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("CompileBasicBlock: %v", err)
	}
	ctx.WriteGas(machineGas)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, block); got != PVM.ExitPanic {
		t.Fatalf("exit=%v, want PANIC (second store below ZZ)", got)
	}
	if got := ctx.ReadExitPC(); got != secondPC {
		t.Fatalf("ExitPC=%d, want second store PC %d", got, secondPC)
	}
}

func TestMemHappyPathFallthroughWritesBothStores(t *testing.T) {
	const addr0 uint32 = 0x10000
	const addr1 uint32 = 0x10010
	blob := buildBlobExact(
		[]byte{59, 0, 0x00, 0x00, 0x01, 59, 0, 0x10, 0x00, 0x01, 0},
		[]int{0, 5, 10},
	)
	prog := mustDeblob(t, blob)
	ctx, compiler := compilerWithProgram(t, &prog)
	page := addr0 / PVM.ZP
	if err := ctx.SetPageAccess(page, unix.PROT_READ|unix.PROT_WRITE); err != nil {
		t.Fatalf("mprotect: %v", err)
	}

	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("CompileBasicBlock: %v", err)
	}
	var regs PVM.Registers
	regs[0] = 0xab
	ctx.WriteRegisters(regs)
	ctx.WriteGas(machineGas)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, block); got != PVM.ExitPanic {
		t.Fatalf("exit=%v, want trap PANIC", got)
	}
	if ctx.guestMem[addr0] != 0xab {
		t.Fatalf("mem[0x%x]=0x%x, want 0xab", addr0, ctx.guestMem[addr0])
	}
	if ctx.guestMem[addr1] != 0xab {
		t.Fatalf("mem[0x%x]=0x%x, want 0xab (happy path must fall through)", addr1, ctx.guestMem[addr1])
	}
}

func BenchmarkCompileMemoryStores(b *testing.B) {
	b.ReportAllocs()
	blob := repeatedStoreTrapBlob(8)
	prog, reason := PVM.DeBlobProgramCode(blob, 0)
	if reason != PVM.ExitContinue {
		b.Fatalf("DeBlobProgramCode: %v", reason)
	}
	ctx, err := NewJITContext()
	if err != nil {
		b.Fatal(err)
	}
	defer ctx.Close()
	em, err := NewExecutableMemory(0)
	if err != nil {
		b.Fatal(err)
	}
	defer em.Close()
	ctx.SetExecutableMemory(em)

	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		if err := em.Reset(); err != nil {
			b.Fatal(err)
		}
		cache := NewCodeCache()
		cache.BindExecutableMemory(em)
		compiler := NewCompiler(&prog, ctx, cache)
		b.StartTimer()
		block, err := compiler.CompileBasicBlock(0)
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		b.ReportMetric(float64(block.NativeSize), "native-bytes")
	}
}

func repeatedStoreTrapBlob(stores int) []byte {
	var inst []byte
	var bounds []int
	for i := 0; i < stores; i++ {
		bounds = append(bounds, len(inst))
		addr := uint32(0x10000 + i)
		inst = append(inst, 59, 0, byte(addr), byte(addr>>8), byte(addr>>16))
	}
	bounds = append(bounds, len(inst))
	inst = append(inst, 0)
	return buildBlobExact(inst, bounds)
}

func compileNativeAt(t *testing.T, prog *PVM.Program, pc PVM.ProgramCounter) (*CompiledBlock, []byte) {
	t.Helper()
	_, compiler := compilerWithProgram(t, prog)
	block, err := compiler.CompileBasicBlock(pc)
	if err != nil {
		t.Fatalf("CompileBasicBlock(%d): %v", pc, err)
	}
	em := compiler.ctx.executableMem
	end := block.NativeOffset + block.NativeSize
	return block, append([]byte(nil), em.rwMem[block.NativeOffset:end]...)
}

func compilerWithProgram(t *testing.T, prog *PVM.Program) (*JITContext, *Compiler) {
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
	compiler := NewCompiler(prog, ctx, NewCodeCache())
	return ctx, compiler
}

func jbRel32Targets(code []byte) []int {
	const jb = 0x82 // Jcc CondB
	var targets []int
	for i := 0; i+6 <= len(code); i++ {
		if code[i] != 0x0F || code[i+1] != jb {
			continue
		}
		rel := int32(binary.LittleEndian.Uint32(code[i+2 : i+6]))
		targets = append(targets, i+6+int(rel))
	}
	return targets
}
