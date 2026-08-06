package PVM

import (
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func encodeUintVariable(val uint64) []byte {
	if val < 0x80 {
		return []byte{byte(val)}
	}
	buf := make([]byte, 9)
	buf[0] = 0xFF
	for i := range 8 {
		buf[1+i] = byte(val >> (8 * i))
	}
	return buf
}

func buildTestBlob(t *testing.T, instBytes []byte, instrBoundaries []int) []byte {
	t.Helper()
	instSize := len(instBytes)
	bitmaskByteCount := instSize / 8
	if instSize%8 > 0 {
		bitmaskByteCount++
	}
	bitmaskData := make([]byte, bitmaskByteCount)
	for _, idx := range instrBoundaries {
		bitmaskData[idx/8] |= 1 << (idx % 8)
	}
	var blob []byte
	blob = append(blob, encodeUintVariable(0)...)
	blob = append(blob, 0)
	blob = append(blob, encodeUintVariable(uint64(instSize))...)
	blob = append(blob, instBytes...)
	blob = append(blob, bitmaskData...)
	return blob
}

func TestGasChargedForIntegratedResume(t *testing.T) {
	prog := decodedGasTestProgram(t,
		ProgramCode{10, 0, 1, 0},
		Bitmask{0x03, 0x00, 0x01, 0x03},
	)
	if gasChargedForIntegratedResume(&prog, 0, true) {
		t.Fatal("block entry with stored flag should clear")
	}
	if !gasChargedForIntegratedResume(&prog, 2, true) {
		t.Fatal("mid-block resume should keep flag")
	}
	if gasChargedForIntegratedResume(&prog, 2, false) {
		t.Fatal("stored false should stay false")
	}
	if gasChargedForIntegratedResume(nil, 0, true) {
		t.Fatal("nil program should clear flag")
	}
}

func TestInvokeInnerTrapDeductsBlockGas(t *testing.T) {
	blob := buildTestBlob(t, []byte{0}, []int{0})

	outerMem := &Memory{Pages: map[uint32]*Page{
		0: {Value: make([]byte, ZP), Access: MemoryReadWrite},
	}}

	innerBudget := Gas(1000)
	var w Registers
	scratch := make([]byte, 112)
	binary.LittleEndian.PutUint64(scratch[0:8], uint64(innerBudget))
	for i := uint64(1); i < 14; i++ {
		binary.LittleEndian.PutUint64(scratch[8*i:8*(i+1)], w[i-1])
	}
	outerMem.Write(0, scratch)

	outerGas := addGas(HostGasInvoke, innerBudget+1000)
	regs := Registers{}
	regs[7] = 0 // n
	regs[8] = 0 // o

	vm := &VMState{Registers: &regs, Gas: &outerGas, Mem: NewPagedGuestMemory(outerMem)}
	addition := HostCallArgs{
		RefineArgs: RefineArgs{
			IntegratedPVMMap: IntegratedPVMMap{
				0: {
					ProgramCode: ProgramCode(blob),
					Memory:      Memory{},
					PC:          0,
					GasCharged:  false,
				},
			},
		},
	}

	out := invoke(OmegaInput{VM: vm, Addition: addition})
	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want continue", out.ExitReason)
	}
	if regs[7] != INNERPANIC {
		t.Fatalf("reg7 = %d, want INNERPANIC(%d)", regs[7], INNERPANIC)
	}

	prog, reason := DeBlobProgramCode(blob, 0)
	if reason != ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", reason)
	}
	blockGas := GasCostForBlock(&prog, 0)
	wantInnerRemain := innerBudget - blockGas

	var gRPrime uint64
	decoder := types.NewDecoder()
	if err := decoder.Decode(outerMem.Read(0, 8), &gRPrime); err != nil {
		t.Fatalf("decode g_R': %v", err)
	}
	if Gas(gRPrime) != wantInnerRemain {
		t.Fatalf("inner g_R' = %d, want %d (budget %d - block gas %d)", gRPrime, wantInnerRemain, innerBudget, blockGas)
	}
}

func TestDeBlobProgramCodeEntryPC(t *testing.T) {
	blob := buildTestBlob(t, []byte{0}, []int{0})
	if _, got := DeBlobProgramCode(blob, 0); got != ExitContinue {
		t.Fatalf("valid blob: got %v", got)
	}
	if _, got := DeBlobProgramCode(blob, 1); got != ExitPanic {
		t.Fatalf("invalid entry PC: got %v, want panic", got)
	}
}

func TestMachineStoresDecodedProgram(t *testing.T) {
	blob := buildTestBlob(t, []byte{0}, []int{0})
	outerMem := &Memory{Pages: map[uint32]*Page{
		0: {Value: make([]byte, ZP), Access: MemoryReadWrite},
	}}
	outerMem.Write(0, blob)

	regs := Registers{}
	regs[7] = 0 // po
	regs[8] = uint64(len(blob))
	regs[9] = 0 // i

	gas := Gas(HostGasMachineConst + MemGas(HostGasMachineOctets, uint64(len(blob))) + 1000)
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(outerMem)}
	addition := HostCallArgs{RefineArgs: RefineArgs{IntegratedPVMMap: IntegratedPVMMap{}}}

	out := machine(OmegaInput{VM: vm, Addition: addition})
	if out.ExitReason != ExitContinue {
		t.Fatalf("machine exit = %v", out.ExitReason)
	}
	n := regs[7]
	entry, ok := out.Addition.IntegratedPVMMap[n]
	if !ok {
		t.Fatal("machine did not register integrated VM")
	}
	if entry.Program == nil {
		t.Fatal("machine should store decoded Program")
	}
	if len(entry.Program.Instrs) == 0 {
		t.Fatal("decoded program should have InstrMeta")
	}
	if len(entry.Program.BlockAt) == 0 {
		t.Fatal("decoded program should have pre-decoded blocks")
	}
}

func TestIntegratedProgramForInvokeUsesCachedDecode(t *testing.T) {
	blob := buildTestBlob(t, []byte{0}, []int{0})
	prog, reason := DeBlobProgramCode(blob, 0)
	if reason != ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", reason)
	}

	integrated := IntegratedPVMType{
		ProgramCode: ProgramCode(blob),
		Program:     &prog,
		PC:          0,
	}
	got, reason := IntegratedProgramForInvoke(integrated)
	if reason != ExitContinue {
		t.Fatalf("invoke resolve: %v", reason)
	}
	if got != &prog {
		t.Fatal("should reuse cached program pointer")
	}

	integrated.PC = 1
	if _, reason := IntegratedProgramForInvoke(integrated); reason != ExitPanic {
		t.Fatalf("invalid entry PC: got %v, want panic", reason)
	}
}
