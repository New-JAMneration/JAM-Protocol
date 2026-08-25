//go:build linux && amd64 && cgo

package recompiler

import (
	"encoding/binary"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

func trapThenJumpBlob() []byte {
	// trap; jump to PC 0 (offset -1)
	return buildBlobExact([]byte{0, 40, 0xFF}, []int{0, 1})
}

func TestDirectLinkUsesRel32WhenTargetAlreadyCompiled(t *testing.T) {
	prog := mustDeblob(t, trapThenJumpBlob())
	if prog.Instrs[1].Opcode != 40 || PVM.ProgramCounter(prog.Instrs[1].Imm[0]) != 0 {
		t.Fatalf("jump instr=%+v, want opcode 40 target 0", prog.Instrs[1])
	}

	ctx, compiler := compilerWithProgram(t, &prog)
	trap, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("compile trap: %v", err)
	}

	before := jm.directLinkRel32.Load()
	jump, err := compiler.CompileBasicBlock(1)
	if err != nil {
		t.Fatalf("compile jump: %v", err)
	}
	if got := jm.directLinkRel32.Load() - before; got != 1 {
		t.Fatalf("directLinkRel32 delta=%d, want 1", got)
	}

	native := compiler.ctx.executableMem.rwMem[jump.NativeOffset : jump.NativeOffset+jump.NativeSize]
	if !rel32LandsOn(native, jump.NativeOffset, trap.NativeOffset) {
		t.Fatalf("jump block has no JMP rel32 to trap NativeOffset=%d (jump@%d size=%d)",
			trap.NativeOffset, jump.NativeOffset, jump.NativeSize)
	}

	ctx.WriteGas(machineGas)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, jump); got != PVM.ExitPanic {
		t.Fatalf("linked jump→trap exit=%v, want PANIC", got)
	}
}

func TestUncompiledTargetStaysDispatchAndIsLargerThanRel32(t *testing.T) {
	prog := mustDeblob(t, trapThenJumpBlob())

	_, dispatchCompiler := compilerWithProgram(t, &prog)
	dispatchJump, err := dispatchCompiler.CompileBasicBlock(1)
	if err != nil {
		t.Fatalf("compile jump without target: %v", err)
	}

	_, relCompiler := compilerWithProgram(t, &prog)
	if _, err := relCompiler.CompileBasicBlock(0); err != nil {
		t.Fatalf("compile trap: %v", err)
	}
	relJump, err := relCompiler.CompileBasicBlock(1)
	if err != nil {
		t.Fatalf("compile jump with target: %v", err)
	}

	if relJump.NativeSize >= dispatchJump.NativeSize {
		t.Fatalf("rel32 jump size=%d, dispatch jump size=%d; rel32 should be smaller",
			relJump.NativeSize, dispatchJump.NativeSize)
	}
}

func TestFallthroughDirectLinkUsesRel32(t *testing.T) {
	// fallthrough; trap
	prog := mustDeblob(t, buildBlobExact([]byte{1, 0}, []int{0, 1}))
	ctx, compiler := compilerWithProgram(t, &prog)
	trap, err := compiler.CompileBasicBlock(1)
	if err != nil {
		t.Fatalf("compile trap: %v", err)
	}
	before := jm.directLinkRel32.Load()
	head, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("compile fallthrough: %v", err)
	}
	if jm.directLinkRel32.Load()-before != 1 {
		t.Fatalf("fallthrough should emit one rel32 link")
	}
	native := compiler.ctx.executableMem.rwMem[head.NativeOffset : head.NativeOffset+head.NativeSize]
	if !rel32LandsOn(native, head.NativeOffset, trap.NativeOffset) {
		t.Fatal("fallthrough epilogue has no JMP rel32 to trap block")
	}

	ctx.WriteGas(machineGas)
	ctx.WriteGasCharged(false)
	if got := ExecuteBlock(ctx, head); got != PVM.ExitPanic {
		t.Fatalf("fallthrough→trap exit=%v, want PANIC", got)
	}
}

func BenchmarkCompileDirectRel32Link(b *testing.B) {
	b.ReportAllocs()
	prog, reason := PVM.DeBlobProgramCode(trapThenJumpBlob(), 0)
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

	for b.Loop() {
		b.StopTimer()
		if err := em.Reset(); err != nil {
			b.Fatal(err)
		}
		cache := NewCodeCache()
		cache.BindExecutableMemory(em)
		compiler := NewCompiler(&prog, ctx, cache)
		if _, err := compiler.CompileBasicBlock(0); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		jump, err := compiler.CompileBasicBlock(1)
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		b.ReportMetric(float64(jump.NativeSize), "native-bytes")
	}
}

func rel32LandsOn(code []byte, codeNativeOffset, wantNativeOffset int) bool {
	for i := 0; i+5 <= len(code); i++ {
		if code[i] != 0xE9 {
			continue
		}
		rel := int32(binary.LittleEndian.Uint32(code[i+1 : i+5]))
		land := codeNativeOffset + i + 5 + int(rel)
		if land == wantNativeOffset {
			return true
		}
	}
	return false
}
