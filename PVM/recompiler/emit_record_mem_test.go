//go:build linux && amd64 && cgo

package recompiler

import (
	"bytes"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

func TestStoreEmitMemTraceControlStoresMatchBuildTag(t *testing.T) {
	const guestAddr uint32 = 0x10000
	native := compileStoreU8Native(t)
	pat := memTraceAddrStoreEncoding(t, guestAddr)
	found := bytes.Contains(native, pat)
	if found != memTraceControlStoresExpected {
		t.Fatalf(
			"mem-trace store to R15-%d present=%v, want %v (native=%dB)",
			OffsetMemAccessAddr,
			found,
			memTraceControlStoresExpected,
			len(native),
		)
	}
}

func compileStoreU8Native(t *testing.T) []byte {
	t.Helper()
	blob := buildBlobExact([]byte{59, 0, 0x00, 0x00, 0x01, 0}, []int{0, 5})
	prog, reason := PVM.DeBlobProgramCode(blob, 0)
	if reason != PVM.ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", reason)
	}

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

	compiler := NewCompiler(&prog, ctx, NewCodeCache())
	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("CompileBasicBlock: %v", err)
	}
	end := block.NativeOffset + block.NativeSize
	return append([]byte(nil), em.rwMem[block.NativeOffset:end]...)
}

func memTraceAddrStoreEncoding(t *testing.T, guestAddr uint32) []byte {
	t.Helper()
	a := asm.NewAssembler()
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetMemAccessAddr), int32(guestAddr))
	code, err := a.Finalize()
	if err != nil {
		t.Fatalf("Finalize mem-trace encoding: %v", err)
	}
	if len(code) == 0 {
		t.Fatal("mem-trace address-store encoding is empty")
	}
	return code
}
