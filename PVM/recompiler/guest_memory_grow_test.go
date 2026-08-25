//go:build linux && amd64 && cgo

package recompiler

import (
	"errors"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

func TestGrowHeapToExpandsContiguousPages(t *testing.T) {
	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	defer ctx.Close()

	const startPages uint64 = 16 // ZZ / ZP
	ctx.WriteHeapPointer(startPages * PVM.ZP)
	ctx.heapLimit = (startPages + 64) * PVM.ZP

	mem := ctx.GuestMemory()
	if err := mem.GrowHeapTo(startPages + 8); err != nil {
		t.Fatalf("GrowHeapTo: %v", err)
	}
	if got := mem.HeapPages(); got != startPages+8 {
		t.Fatalf("HeapPages=%d, want %d", got, startPages+8)
	}
	start := startPages * PVM.ZP
	if !mem.IsWriteable(start, PVM.ZP*8) {
		t.Fatal("new heap pages should be writeable")
	}
}

func TestGrowHeapToOnePageIsOneSyscall(t *testing.T) {
	testGrowHeapToSyscallCount(t, 1)
}

func TestGrowHeapToEightPagesIsOneSyscall(t *testing.T) {
	testGrowHeapToSyscallCount(t, 8)
}

func testGrowHeapToSyscallCount(t *testing.T, pages uint64) {
	t.Helper()
	orig := guestMprotect
	t.Cleanup(func() { guestMprotect = orig })

	var calls int
	var lastLen int
	guestMprotect = func(b []byte, prot int) error {
		calls++
		lastLen = len(b)
		return orig(b, prot)
	}

	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	defer ctx.Close()

	const startPages uint64 = 16
	ctx.WriteHeapPointer(startPages * PVM.ZP)
	ctx.heapLimit = (startPages + 64) * PVM.ZP
	mem := ctx.GuestMemory()
	if err := mem.GrowHeapTo(startPages + pages); err != nil {
		t.Fatalf("GrowHeapTo: %v", err)
	}
	if calls != 1 {
		t.Fatalf("mprotect calls=%d, want 1 for %d pages", calls, pages)
	}
	wantLen := int(pages * PVM.ZP)
	if lastLen != wantLen {
		t.Fatalf("mprotect length=%d, want %d", lastLen, wantLen)
	}
}

func TestGrowHeapToProtectFailureLeavesHeapAndPagesUnchanged(t *testing.T) {
	orig := guestMprotect
	t.Cleanup(func() { guestMprotect = orig })
	injected := errors.New("injected protect failure")
	guestMprotect = func([]byte, int) error { return injected }

	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	defer ctx.Close()

	const startPages uint64 = 16
	startHP := startPages * PVM.ZP
	ctx.WriteHeapPointer(startHP)
	ctx.heapLimit = (startPages + 64) * PVM.ZP
	mem := ctx.GuestMemory()
	if err := mem.GrowHeapTo(startPages + 8); err == nil {
		t.Fatal("GrowHeapTo: want injected protect failure")
	} else if !errors.Is(err, injected) {
		t.Fatalf("GrowHeapTo: %v, want %v", err, injected)
	}
	if got := mem.HeapPages(); got != startPages {
		t.Fatalf("HeapPages=%d, want unchanged %d", got, startPages)
	}
	if mem.IsWriteable(startHP, PVM.ZP*8) {
		t.Fatal("failed grow must not mark new pages writeable")
	}
}

func TestGrowHeapToZeroGrowthIsNoop(t *testing.T) {
	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	defer ctx.Close()

	const pages uint64 = 16
	ctx.WriteHeapPointer(pages * PVM.ZP)
	ctx.heapLimit = (pages + 8) * PVM.ZP
	mem := ctx.GuestMemory()
	orig := guestMprotect
	t.Cleanup(func() { guestMprotect = orig })
	var calls int
	guestMprotect = func(b []byte, prot int) error {
		calls++
		return orig(b, prot)
	}
	if err := mem.GrowHeapTo(pages); err != nil {
		t.Fatalf("GrowHeapTo same size: %v", err)
	}
	if calls != 0 {
		t.Fatalf("zero growth mprotect calls=%d, want 0", calls)
	}
	if got := mem.HeapPages(); got != pages {
		t.Fatalf("HeapPages=%d, want %d", got, pages)
	}
}

func BenchmarkGrowHeapTo(b *testing.B) {
	b.ReportAllocs()
	for _, n := range []uint64{1, 8} {
		b.Run(itoaPages(n), func(b *testing.B) {
			for b.Loop() {
				b.StopTimer()
				ctx, err := NewJITContext()
				if err != nil {
					b.Fatal(err)
				}
				const startPages uint64 = 16
				ctx.WriteHeapPointer(startPages * PVM.ZP)
				ctx.heapLimit = (startPages + 64) * PVM.ZP
				mem := ctx.GuestMemory()
				b.StartTimer()
				if err := mem.GrowHeapTo(startPages + n); err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				_ = ctx.Close()
			}
		})
	}
}

func itoaPages(n uint64) string {
	if n == 1 {
		return "1page"
	}
	return "8pages"
}
