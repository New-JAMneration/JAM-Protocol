package PVM

import (
	"math"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestAddGasAndUnitGasCost(t *testing.T) {
	if got := addGas(HostGasLookupConst, MemGas(HostGasLookupOctets, 0)); got != HostGasLookupConst {
		t.Fatalf("lookup length 0: got %d want %d", got, HostGasLookupConst)
	}
	if got := addGas(HostGasLookupConst, MemGas(HostGasLookupOctets, 1024)); got != HostGasLookupConst+HostGasLookupOctets {
		t.Fatalf("lookup length 1024: got %d want %d", got, HostGasLookupConst+HostGasLookupOctets)
	}
	if got := addGas(1, math.MaxInt64); got != math.MaxInt64 {
		t.Fatalf("addGas saturate: got %d", got)
	}
	if got := unitGasCost(HostGasBlessConst, HostGasBlessItem, 0); got != HostGasBlessConst {
		t.Fatalf("bless n=0: got %d", got)
	}
	if got := unitGasCost(HostGasBlessConst, HostGasBlessItem, 3); got != HostGasBlessConst+3*HostGasBlessItem {
		t.Fatalf("bless n=3: got %d", got)
	}
	if got := unitGasCost(1, math.MaxInt64, math.MaxUint64); got != math.MaxInt64 {
		t.Fatalf("unitGasCost saturate: got %d", got)
	}
}

func TestPagesGasCost(t *testing.T) {
	tests := []struct {
		name string
		r, c uint64
		want Gas
	}{
		{"free", 0, 2, HostGasPagesFreeConst + 2*HostGasPagesFreePage},
		{"alloc r=1", 1, 2, HostGasPagesAllocConst + 2*HostGasPagesAllocPage},
		{"alloc r=2", 2, 0, HostGasPagesAllocConst},
		{"setmode r=3", 3, 4, HostGasPagesSetModeConst + 4*HostGasPagesSetModePage},
		{"setmode r=4", 4, 1, HostGasPagesSetModeConst + HostGasPagesSetModePage},
		{"invalid", 5, 10, HostGasPagesInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pagesGasCost(tt.r, tt.c); got != tt.want {
				t.Fatalf("pagesGasCost(%d,%d)=%d want %d", tt.r, tt.c, got, tt.want)
			}
		})
	}
}

func TestIsBlobLength(t *testing.T) {
	if !isBlobLength(0) || !isBlobLength((1<<32)-1) {
		t.Fatal("expected in-range lengths to be valid")
	}
	if isBlobLength(1<<32) || isBlobLength(math.MaxUint64) {
		t.Fatal("expected z >= 2^32 to be invalid")
	}
}

func TestIsValidValidatorCount(t *testing.T) {
	types.SetTinyMode()
	t.Cleanup(types.SetTinyMode)

	if !isValidValidatorCount(6) {
		t.Fatal("tiny mode should accept z=6")
	}
	for _, z := range []uint64{0, 3, 9, 7, 5} {
		if isValidValidatorCount(z) {
			t.Fatalf("tiny mode should reject z=%d", z)
		}
	}

	types.SetFullMode()
	if !isValidValidatorCount(6) || !isValidValidatorCount(1023) || !isValidValidatorCount(9) {
		t.Fatal("full mode should accept 6, 9, 1023")
	}
	if isValidValidatorCount(1026) || isValidValidatorCount(1) {
		t.Fatal("full mode should reject out-of-range counts")
	}
	types.SetTinyMode()
}

func TestLookupLinearGasOOG(t *testing.T) {
	regs := Registers{}
	regs[11] = 1024                // z → MemGas(LookupOctets, 1024) > 0
	gas := Gas(HostGasLookupConst) // const only → OOG after linear term
	state := types.ServiceAccountState{}
	acct := types.ServiceAccount{}
	sid := types.ServiceID(1)
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(&Memory{})}
	out := lookup(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceID:           &sid,
				ServiceAccount:      &acct,
				ServiceAccountState: &state,
			},
		},
	})
	if out.ExitReason != ExitOOG {
		t.Fatalf("lookup OOG: got %v want %v", out.ExitReason, ExitOOG)
	}
}

func TestQueryBlobLengthHUH(t *testing.T) {
	regs := Registers{}
	regs[8] = 1 << 32 // z ∉ bloblength
	gas := Gas(HostGasQuery + 100)
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(&Memory{Pages: map[uint32]*Page{}})}
	out := query(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			AccumulateArgs: AccumulateArgs{
				ResultContextX: ResultContext{
					PartialState: types.PartialStateSet{
						ServiceAccounts: types.ServiceAccountState{},
					},
				},
			},
		},
	})
	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want continue", out.ExitReason)
	}
	if regs[7] != HUH {
		t.Fatalf("reg7 = %d, want HUH(%d)", regs[7], HUH)
	}
}

func TestBlessManagerOnly(t *testing.T) {
	types.SetTinyMode()
	t.Cleanup(types.SetTinyMode)

	regs := Registers{}
	// m,a,v,r,o,n — n=0 so always-accum payload is empty
	regs[7] = 1
	regs[8] = 16 * ZP
	regs[9] = 2
	regs[10] = 3
	regs[11] = 16*ZP + uint64(4*types.CoresCount)
	regs[12] = 0

	mem := &Memory{Pages: map[uint32]*Page{}}
	page := uint32(16)
	mem.Pages[page] = &Page{Value: make([]byte, ZP), Access: MemoryReadWrite}

	gas := Gas(HostGasBlessConst + 1000)
	caller := types.ServiceID(99) // not manager
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
	out := bless(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			AccumulateArgs: AccumulateArgs{
				ResultContextX: ResultContext{
					ServiceID: caller,
					PartialState: types.PartialStateSet{
						Bless: types.ServiceID(1), // manager
					},
				},
			},
		},
	})
	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want continue", out.ExitReason)
	}
	if regs[7] != HUH {
		t.Fatalf("reg7 = %d, want HUH(%d)", regs[7], HUH)
	}
}

func newGrowHeapTestMem() *Memory {
	return &Memory{
		Pages:       map[uint32]*Page{},
		heapPointer: 3 * ZZ, // h = 3*ZZ/ZP
		heapLimit:   1<<32 - 2*ZZ - ZI,
	}
}

func TestGrowHeapGasOutcomes(t *testing.T) {
	t.Run("noGrowthChargesConstEvenIfNegative", func(t *testing.T) {
		mem := newGrowHeapTestMem()
		h := NewPagedGuestMemory(mem).HeapPages()
		gas := HostGasGrowHeapConst / 2
		regs := Registers{}
		regs[7] = h // no growth
		vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
		out := growHeap(OmegaInput{VM: vm})
		if out.ExitReason != ExitContinue {
			t.Fatalf("exit = %v, want continue", out.ExitReason)
		}
		if gas != HostGasGrowHeapConst/2-HostGasGrowHeapConst {
			t.Fatalf("gas = %d, want %d", gas, HostGasGrowHeapConst/2-HostGasGrowHeapConst)
		}
		if regs[7] != h {
			t.Fatalf("reg7 = %d, want h=%d", regs[7], h)
		}
	})

	t.Run("growthOOGLeavesGasUnchanged", func(t *testing.T) {
		mem := newGrowHeapTestMem()
		h := NewPagedGuestMemory(mem).HeapPages()
		b := NewPagedGuestMemory(mem).HeapMaxPages()
		want := h + 1
		if want > b {
			t.Skip("heap already at max")
		}
		g := HostGasGrowHeapConst + HostGasGrowHeapPage
		gas := g - 1
		regs := Registers{}
		regs[7] = want
		vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
		out := growHeap(OmegaInput{VM: vm})
		if out.ExitReason != ExitOOG {
			t.Fatalf("exit = %v, want OOG", out.ExitReason)
		}
		if gas != g-1 {
			t.Fatalf("gas changed to %d, want unchanged %d", gas, g-1)
		}
		if regs[7] != h {
			t.Fatalf("reg7 = %d, want h=%d", regs[7], h)
		}
	})

	t.Run("growthSucceeds", func(t *testing.T) {
		mem := newGrowHeapTestMem()
		h := NewPagedGuestMemory(mem).HeapPages()
		b := NewPagedGuestMemory(mem).HeapMaxPages()
		want := h + 1
		if want > b {
			t.Skip("heap already at max")
		}
		g := HostGasGrowHeapConst + HostGasGrowHeapPage
		gas := g + 10
		regs := Registers{}
		regs[7] = want
		vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
		out := growHeap(OmegaInput{VM: vm})
		if out.ExitReason != ExitContinue {
			t.Fatalf("exit = %v, want continue", out.ExitReason)
		}
		if gas != 10 {
			t.Fatalf("gas = %d, want 10", gas)
		}
		if regs[7] != want {
			t.Fatalf("reg7 = %d, want %d", regs[7], want)
		}
	})
}

func TestFetchConstantsOmitRemovedV080Fields(t *testing.T) {
	types.SetTinyMode()
	t.Cleanup(types.SetTinyMode)

	data := getFetchConstantsData()
	// v0.8.0 fetch(0) dropped N, V, W_E, W_P (2+2+4+4 = 12 bytes vs 0.7.2).
	// Spot-check: after L (u32) comes O (u16), not N (u16 tickets).
	if len(data) < 80 {
		t.Fatalf("constants too short: %d", len(data))
	}
	off := 8 + 8 + 8 + 2 + 4 + 4 + 8 + 8 + 8 + 8 + 2 + 2 + 2 + 2 + 4 // through L
	gotO := uint16(data[off]) | uint16(data[off+1])<<8
	if gotO != uint16(types.AuthPoolMaxSize) {
		t.Fatalf("after L expected O=%d, got %d (N/V still present?)", types.AuthPoolMaxSize, gotO)
	}
}
