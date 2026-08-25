//go:build linux && amd64 && cgo

package recompiler

import (
	"encoding/binary"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

func TestSharedExitTrampolineIsOutsideBlocksAndReused(t *testing.T) {
	prog := mustDeblob(t, trapThenJumpBlob())
	ctx, compiler := compilerWithProgram(t, &prog)

	trap, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("compile trap: %v", err)
	}
	if !ctx.exitTrampolineReady {
		t.Fatal("first compile should emit a shared exit trampoline")
	}
	exitOff := ctx.exitTrampolineOffset
	if inNativeRange(exitOff, trap) {
		t.Fatalf("exit trampoline offset %d is inside trap block [%d,%d)",
			exitOff, trap.NativeOffset, trap.NativeOffset+trap.NativeSize)
	}

	jump, err := compiler.CompileBasicBlock(1)
	if err != nil {
		t.Fatalf("compile jump: %v", err)
	}
	if ctx.exitTrampolineOffset != exitOff {
		t.Fatalf("second compile re-emitted exit trampoline: %d → %d",
			exitOff, ctx.exitTrampolineOffset)
	}
	if inNativeRange(exitOff, jump) {
		t.Fatalf("exit trampoline offset %d is inside jump block [%d,%d)",
			exitOff, jump.NativeOffset, jump.NativeOffset+jump.NativeSize)
	}

	trapNative := compiler.ctx.executableMem.rwMem[trap.NativeOffset : trap.NativeOffset+trap.NativeSize]
	if !rel32LandsOn(trapNative, trap.NativeOffset, exitOff) {
		t.Fatal("trap block has no JMP rel32 to the shared exit trampoline")
	}

	ctx.WriteGas(machineGas)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, trap); got != PVM.ExitPanic {
		t.Fatalf("trap exit=%v, want PANIC", got)
	}
}

func TestSharedExitTrampolineReemitsAfterEmReset(t *testing.T) {
	prog := mustDeblob(t, buildBlobExact([]byte{0}, []int{0}))
	ctx, compiler := compilerWithProgram(t, &prog)
	if _, err := compiler.CompileBasicBlock(0); err != nil {
		t.Fatalf("first compile: %v", err)
	}
	em := ctx.executableMem
	if err := em.Reset(); err != nil {
		t.Fatal(err)
	}
	cache := NewCodeCache()
	cache.BindExecutableMemory(em)
	compiler = NewCompiler(&prog, ctx, cache)
	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("compile after Reset: %v", err)
	}
	if !ctx.exitTrampolineReady {
		t.Fatal("compile after Reset should emit a new exit trampoline")
	}
	if inNativeRange(ctx.exitTrampolineOffset, block) {
		t.Fatalf("re-emitted trampoline %d is inside block [%d,%d)",
			ctx.exitTrampolineOffset, block.NativeOffset, block.NativeOffset+block.NativeSize)
	}
	ctx.WriteGas(machineGas)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, block); got != PVM.ExitPanic {
		t.Fatalf("after Reset exit=%v, want PANIC", got)
	}
}

func TestSharedExitTrampolinePreemittedByBuildCompiledProgram(t *testing.T) {
	prog := mustDeblob(t, buildBlobExact([]byte{0}, []int{0}))
	cp, err := buildCompiledProgram(&prog)
	if err != nil {
		t.Fatalf("buildCompiledProgram: %v", err)
	}
	t.Cleanup(func() { cp.close() })
	if cp.exitTramp == 0 {
		t.Fatal("cached artefact must pre-emit the shared exit trampoline")
	}

	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	t.Cleanup(func() { _ = ctx.Close() })
	cp.bindContext(ctx)
	if !ctx.exitTrampolineReady || ctx.exitTrampolineOffset != cp.exitTrampOffset {
		t.Fatalf("bindContext exit ready=%v offset=%d, want artefact offset=%d",
			ctx.exitTrampolineReady, ctx.exitTrampolineOffset, cp.exitTrampOffset)
	}

	compiler := NewCompiler(&prog, ctx, cp.cache)
	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("CompileBasicBlock: %v", err)
	}
	if ctx.exitTrampolineOffset != cp.exitTrampOffset {
		t.Fatalf("compile re-emitted exit trampoline: artefact=%d ctx=%d",
			cp.exitTrampOffset, ctx.exitTrampolineOffset)
	}
	if inNativeRange(cp.exitTrampOffset, block) {
		t.Fatalf("pre-emitted trampoline %d is inside block [%d,%d)",
			cp.exitTrampOffset, block.NativeOffset, block.NativeOffset+block.NativeSize)
	}

	native := ctx.executableMem.rwMem[block.NativeOffset : block.NativeOffset+block.NativeSize]
	if !rel32LandsOn(native, block.NativeOffset, cp.exitTrampOffset) {
		t.Fatal("block has no JMP rel32 to the pre-emitted exit trampoline")
	}

	ctx.WriteGas(machineGas)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, block); got != PVM.ExitPanic {
		t.Fatalf("exit=%v, want PANIC", got)
	}
}

func TestTrapAndOOGPadsJumpToSharedExitTrampoline(t *testing.T) {
	prog := mustDeblob(t, buildBlobExact([]byte{0}, []int{0}))
	ctx, compiler := compilerWithProgram(t, &prog)
	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("CompileBasicBlock: %v", err)
	}
	exitOff := ctx.exitTrampolineOffset
	native := ctx.executableMem.rwMem[block.NativeOffset : block.NativeOffset+block.NativeSize]
	if n := rel32LandingCount(native, block.NativeOffset, exitOff); n < 2 {
		t.Fatalf("JMP rel32 to shared trampoline count=%d, want ≥2 (trap + OOG pad)", n)
	}

	ctx.WriteGas(0)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, block); got != PVM.ExitOOG {
		t.Fatalf("gas=0 exit=%v, want OOG via shared trampoline", got)
	}

	ctx.WriteGas(machineGas)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, block); got != PVM.ExitPanic {
		t.Fatalf("gas=%d exit=%v, want PANIC via shared trampoline", machineGas, got)
	}
}

func TestSetExecutableMemoryClearsSharedExitTrampoline(t *testing.T) {
	prog := mustDeblob(t, buildBlobExact([]byte{0}, []int{0}))
	ctx, compiler := compilerWithProgram(t, &prog)
	if _, err := compiler.CompileBasicBlock(0); err != nil {
		t.Fatalf("first compile: %v", err)
	}
	if !ctx.exitTrampolineReady {
		t.Fatal("first compile should install an exit trampoline on ctx")
	}

	em2, err := NewExecutableMemory(0)
	if err != nil {
		t.Fatalf("NewExecutableMemory: %v", err)
	}
	t.Cleanup(func() { _ = em2.Close() })
	ctx.SetExecutableMemory(em2)
	if ctx.exitTrampolineReady || ctx.exitTrampolineAddr != 0 {
		t.Fatalf("SetExecutableMemory must clear exit trampoline (ready=%v addr=%#x)",
			ctx.exitTrampolineReady, ctx.exitTrampolineAddr)
	}

	cache := NewCodeCache()
	cache.BindExecutableMemory(em2)
	compiler = NewCompiler(&prog, ctx, cache)
	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("compile on new arena: %v", err)
	}
	if !ctx.exitTrampolineReady {
		t.Fatal("compile on a new arena should emit a new exit trampoline")
	}
	if inNativeRange(ctx.exitTrampolineOffset, block) {
		t.Fatalf("new trampoline %d is inside block [%d,%d)",
			ctx.exitTrampolineOffset, block.NativeOffset, block.NativeOffset+block.NativeSize)
	}
	ctx.WriteGas(machineGas)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, block); got != PVM.ExitPanic {
		t.Fatalf("new arena exit=%v, want PANIC", got)
	}
}

func inNativeRange(offset int, block *CompiledBlock) bool {
	return offset >= block.NativeOffset && offset < block.NativeOffset+block.NativeSize
}

func rel32LandingCount(code []byte, codeNativeOffset, wantNativeOffset int) int {
	n := 0
	for i := 0; i+5 <= len(code); i++ {
		if code[i] != 0xE9 {
			continue
		}
		rel := int32(binary.LittleEndian.Uint32(code[i+1 : i+5]))
		if codeNativeOffset+i+5+int(rel) == wantNativeOffset {
			n++
		}
	}
	return n
}
