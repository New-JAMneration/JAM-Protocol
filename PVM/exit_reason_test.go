// test.go
package PVM

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestExitReasonGetReasonType(t *testing.T) {
	tests := []struct {
		name     string
		reason   ExitReason
		expected ExitReasonType
	}{
		{"Continue", ExitContinue, CONTINUE},
		{"Halt", ExitHalt, HALT},
		{"Panic", ExitPanic, PANIC},
		{"OOG", ExitOOG, OUT_OF_GAS},
		{"PageFault", ExitPageFault | 0x12345678, PAGE_FAULT},
		{"HostCall", ExitHostCall | 0xFF, HOST_CALL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reason.GetReasonType(); got != tt.expected {
				t.Errorf("GetReasonType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExitReasonDataExtraction(t *testing.T) {
	t.Run("PageFaultAddress_Boundary", func(t *testing.T) {
		var addr uint32 = 0xABCDEFFF
		reason := ExitPageFault | ExitReason(addr)
		if got := reason.GetPageFaultAddress(); got != addr {
			t.Errorf("GetPageFaultAddress() = 0x%x, want 0x%x", got, addr)
		}

		if reason.GetReasonType() != PAGE_FAULT {
			t.Errorf("Type corrupted! got %v", reason.GetReasonType())
		}
	})

	t.Run("HostCallID_Boundary", func(t *testing.T) {
		var id uint8 = 255
		reason := ExitHostCall | ExitReason(id)
		if got := reason.GetHostCallID(); got != id {
			t.Errorf("GetHostCallID() = %d, want %d", got, id)
		}
	})
}

func TestExitReasonString(t *testing.T) {
	tests := []struct {
		reason   ExitReason
		expected string
	}{
		{ExitContinue, "Continue"},
		{ExitHalt, "Halt"},
		{ExitPageFault | 0xDEADBEEF, "Page fault at 0xdeadbeef"},
		{ExitHostCall | 42, "Host Call 42"},
		{ExitReason(99) << 56, "Unknown ExitReasonType(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.reason.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func psi1(pc, gas, reg, mem int) (int, int, int, int, ExitReason) {
	var (
		newPc  = pc + 1 // + skip()
		newGas = gas - 10
		newReg = reg
		newMem = mem
	)

	if newGas < 0 {
		return newPc, newGas, newReg, newMem, ExitOOG
	}
	if newMem == 0 { // mock "0" for page fault
		return newPc, newGas, newReg, newMem, ExitPageFault | ExitReason(newMem)
	}
	// return newPc, newGas, newReg, newMem, PVMExitTuple(CONTINUE, nil)
	return newPc, newGas, newReg, newMem, ExitHostCall | ExitReason(newMem)
}

// ---

func naiveGeneralFunction(con uint64) any {
	register := [13]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	newRegister := register
	newRegister[7] = con
	return newRegister
}

func TestOmega(t *testing.T) {
	constants := []struct {
		given uint64
		want  uint64
	}{
		{OK, 0},
		{HUH, ^uint64(8)},
		{LOW, ^uint64(7)},
		{CASH, ^uint64(6)},
		{CORE, ^uint64(5)},
		{FULL, ^uint64(4)},
		{WHO, ^uint64(3)},
		{OOB, ^uint64(2)},
		{WHAT, ^uint64(1)},
		{NONE, ^uint64(0)},
		{INNERHALT, 0},
		{INNERPANIC, 1},
		{INNERFAULT, 2},
		{INNERHOST, 3},
		{INNEROOG, 4},
	}
	for _, tt := range constants {
		t.Run(fmt.Sprintf("given=%d", tt.given), func(t *testing.T) {
			got := naiveGeneralFunction(tt.given)
			want := tt.want
			// test 7th bit in register
			// fmt.Printf("want: %v\n", want)
			if got.([13]uint64)[7] != want {
				t.Errorf("omega() = %v, want %v", got.([13]uint64)[7], want)
			}
		})
	}
}

func TestBranch(t *testing.T) {
	tests := []struct {
		name        string
		pc          ProgramCounter
		offset      uint32
		condition   bool
		basicBlocks []uint32
		wantExit    ExitReasonType
		wantPC      ProgramCounter
	}{
		{
			name:        "ContinueExecution",
			pc:          0,
			offset:      10,
			condition:   false,
			basicBlocks: []uint32{0, 10, 20, 30},
			wantExit:    CONTINUE,
			wantPC:      0,
		},
		{
			name:        "JumpToBasicBlock",
			pc:          0,
			offset:      10,
			condition:   true,
			basicBlocks: []uint32{0, 10, 20, 30},
			wantExit:    CONTINUE,
			wantPC:      10,
		},
		{
			name:        "InvalidTarget",
			pc:          0,
			offset:      10,
			condition:   true,
			basicBlocks: []uint32{0, 20, 30}, // 10 is not a basic block
			wantExit:    PANIC,
			wantPC:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitReason, newPC := Branch(tt.pc, tt.offset, tt.condition, tt.basicBlocks)
			if exitReason != tt.wantExit {
				t.Errorf("Branch() exitReason = %v, want %v", exitReason, tt.wantExit)
			}
			if newPC != tt.wantPC {
				t.Errorf("Branch() newPC = %v, want %v", newPC, tt.wantPC)
			}
		})
	}
}

func TestDjump(t *testing.T) {
	testCases := []struct {
		name               string
		target             uint32
		jumpTable          []uint32
		basicBlocks        []uint32
		expectedExitReason ExitReasonType
		expectedPC         uint32
	}{
		{
			"DjumpToValidTarget",
			2,
			[]uint32{4, 8, 12},
			[]uint32{0, 4, 8},
			CONTINUE,
			8,
		},
		{
			"DjumpToInvalidTarget",
			3,
			[]uint32{4, 8, 12},
			[]uint32{0, 4, 8},
			PANIC,
			3,
		},
		{
			"InvalidJump_TargetNotBasicBlock",
			4,
			[]uint32{10, 20, 30, 40},
			[]uint32{0, 10, 20, 30}, // 40 is not in basicBlocks
			CONTINUE,
			30,
		},
		{
			"SpecialCase_Halt",
			math.MaxUint32 - ZZ + 1,
			[]uint32{10, 20, 30, 40},
			[]uint32{0, 10, 20, 30, 40},
			HALT,
			math.MaxUint32 - ZZ + 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exitReason, newPC := Djump(tc.target, tc.jumpTable, tc.basicBlocks)
			if exitReason != tc.expectedExitReason {
				t.Errorf("Expected exit reason %v, but got %v", tc.expectedExitReason, exitReason)
			}
			if newPC != tc.expectedPC {
				t.Errorf("Expected PC %d, but got %d", tc.expectedPC, newPC)
			}
		})
	}
}

func TestGetInvalidAddress(t *testing.T) {
	tests := []struct {
		name           string
		readAddresses  []uint64
		writeAddresses []uint64
		readable       map[int]bool
		writeable      map[int]bool
		want           []uint64
	}{
		{
			name:           "case1",
			readAddresses:  []uint64{5000, 10005, 33300},
			writeAddresses: []uint64{},
			readable:       map[int]bool{1: true},
			writeable:      map[int]bool{},
			want:           []uint64{10005, 33300},
		},
		{
			name:           "case2",
			readAddresses:  []uint64{1, 2, 4},
			writeAddresses: []uint64{3, 4, 5},
			readable:       map[int]bool{},
			writeable:      map[int]bool{},
			want:           []uint64{1, 2, 3, 4, 5},
		},
		{
			name:           "case3",
			readAddresses:  []uint64{4800, 9600, 10800},
			writeAddresses: []uint64{},
			readable:       map[int]bool{2: true},
			writeable:      map[int]bool{},
			want:           []uint64{4800},
		},
		{
			name:           "case4",
			readAddresses:  []uint64{},
			writeAddresses: []uint64{4800, 5000, 9600, 10800},
			readable:       map[int]bool{},
			writeable:      map[int]bool{1: true},
			want:           []uint64{9600, 10800},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInvalidAddress(tt.readAddresses, tt.writeAddresses, tt.readable, tt.writeable); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInvalidAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
