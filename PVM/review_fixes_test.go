package PVM

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestDeblobRejectsOpenFinalBlock(t *testing.T) {
	// Minimal blob: empty jump table, 3-byte load_imm (non-terminator), bitmask.
	open := []byte{0, 0, 3, 51, 0x00, 0, 1}
	if _, got := DeBlobProgramCode(open, 0); got != ExitPanic {
		t.Fatalf("open final block via deblob: got %v, want panic", got)
	}
	// Same fixture is acceptable for gas-model decode.
	if _, got := deblobProgramForGasModel(open); got != ExitContinue {
		t.Fatalf("gas-model decode of open block: got %v, want continue", got)
	}

	// Valid: trap terminator.
	ok := []byte{0, 0, 1, 0, 1}
	if _, got := DeBlobProgramCode(ok, 0); got != ExitContinue {
		t.Fatalf("terminator final block: got %v, want continue", got)
	}
}

func TestBranchCyclesOutOfRangeTargetIsTrap(t *testing.T) {
	prog := decodedGasTestProgram(t,
		ProgramCode{
			170, 0x01, 100, 0, 0, 0, // branch_eq target PC 100 (past end)
			0, // trap fallthrough at PC 6
		},
		Bitmask{0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
	)
	cost := InstructionCost(&prog, &prog.Instrs[0])
	if cost.Cycles != 1 {
		t.Fatalf("out-of-range branch target cycles = %d, want 1", cost.Cycles)
	}
}

func TestSelfLoopJumpRetakesTerminator(t *testing.T) {
	// jump to PC 0 (self); fallthrough unreachable.
	prog := decodedGasTestProgram(t,
		ProgramCode{40, 0, 0, 0, 0}, // jump imm=0
		Bitmask{0x03, 0x00, 0x00, 0x00, 0x00},
	)
	gas := Gas(10_000)
	interp := &Interpreter{Program: &prog, Gas: gas, GasCharged: false}
	exit, pc := interp.BlockBasedInvokeDecodedBlocks(0)
	// Should OOG eventually while looping, with PC at the jump.
	if exit.GetReasonType() != OUT_OF_GAS {
		t.Fatalf("exit = %v, want OOG (self-loop must re-enter block)", exit)
	}
	if pc != 0 {
		t.Fatalf("OOG pc = %d, want 0 (self-loop)", pc)
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

func TestHeapMaxPagesReservesMajorZone(t *testing.T) {
	stackStart := uint64(1<<32 - 2*ZZ - ZI)
	mem := &Memory{heapLimit: stackStart}
	got := NewPagedGuestMemory(mem).HeapMaxPages()
	want := (stackStart - uint64(ZZ)) / uint64(ZP)
	if got != want {
		t.Fatalf("HeapMaxPages = %d, want %d", got, want)
	}
}

func TestMachineCapacityFullBeforeMemory(t *testing.T) {
	regs := Registers{}
	regs[7] = 0
	regs[8] = 1
	regs[9] = 0
	gas := Gas(HostGasMachineConst + 10_000)
	m := IntegratedPVMMap{}
	for i := uint64(0); i < 63; i++ {
		m[i] = IntegratedPVMType{}
	}
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(&Memory{Pages: map[uint32]*Page{}})}
	out := machine(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			RefineArgs: RefineArgs{IntegratedPVMMap: m},
		},
	})
	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want continue", out.ExitReason)
	}
	if regs[7] != FULL {
		t.Fatalf("reg7 = %d, want FULL(%d)", regs[7], FULL)
	}
}

func TestFetchConstantsOmitRemovedV080Fields(t *testing.T) {
	types.SetTinyMode()
	t.Cleanup(types.SetTinyMode)

	data := getFetchConstantsData()
	// v0.8.0 fetch(0) dropped N, V, W_E, W_P (2+2+4+4 = 12 bytes vs 0.7.2).
	// EncodeMany of remaining fields: count fixed-width scalars.
	// Spot-check: after L (u32) comes O (u16), not N (u16 tickets).
	// Layout through L: 8*3 + 2 + 4*2 + 8*4 + 2*4 + 4 = 24+2+8+32+8+4 = 78, then O at offset 78.
	if len(data) < 80 {
		t.Fatalf("constants too short: %d", len(data))
	}
	// O = AuthPoolMaxSize as u16 LE at offset after L.
	off := 8 + 8 + 8 + 2 + 4 + 4 + 8 + 8 + 8 + 8 + 2 + 2 + 2 + 2 + 4 // through L
	gotO := uint16(data[off]) | uint16(data[off+1])<<8
	if gotO != uint16(types.AuthPoolMaxSize) {
		t.Fatalf("after L expected O=%d, got %d (N/V still present?)", types.AuthPoolMaxSize, gotO)
	}
}
