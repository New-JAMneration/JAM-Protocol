// test.go
package PolkaVM

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestExitReasonsTypes(t *testing.T) {
	tests := []struct {
		reason ExitReasonTypes
		want   string
	}{
		{CONTINUE, "continue"},
		{HALT, "halt"},
		{PANIC, "panic"},
		{OUT_OF_GAS, "out-of-gas"},
		{PAGE_FAULT, "page-fault"},
		{HOST_CALL, "host-call identifier"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.reason), func(t *testing.T) {
			if got := exitMessages[tt.reason]; got != tt.want {
				t.Errorf("exitMessages[%v] = %v, want %v", tt.reason, got, tt.want)
			}
		})
	}
}

func TestInvocation(t *testing.T) {
	// mockHC := "mockHostCall"

	tests := []struct {
		p, pc, gas, reg, mem int
		want                 string
	}{
		// {1, 0, 10, 0, 1, "Continue (▸)"},                     // CONTINUE
		{1, 0, 10, 0, 0, "page-fault at RAM address: 0"}, // PAGE_FAULT
		{1, 0, 0, 0, 1, "out-of-gas"},                    // OUT_OF_GAS
		{1, 0, 10, 0, 1, "host-call identifier: 1"},      // HOST_CALL
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("p=%d, pc=%d, gas=%d, reg=%d, mem=%d", tt.p, tt.pc, tt.gas, tt.reg, tt.mem), func(t *testing.T) {
			_, _, _, _, got := exPsi(tt.p, tt.pc, tt.gas, tt.reg, tt.mem)
			gotStr := got.Error()
			if gotStr != tt.want {
				t.Errorf("ExPsi(%d, %d, %d, %d, %d) = %v, want %v", tt.p, tt.pc, tt.gas, tt.reg, tt.mem, got, tt.want)
			}
		})
	}
}

// test all exit types of invocation function
func exPsi(p, pc, gas, reg, mem int) (int, int, int, int, error) {
	// call Psi1 to renew states first
	newPc, newGas, newReg, newMem, epsilon := psi1(pc, gas, reg, mem)

	if errors.Is(epsilon, PVMExitTuple(CONTINUE, nil)) {
		return exPsi(p, newPc, newGas, newReg, newMem)
	} else if errors.Is(epsilon, PVMExitTuple(OUT_OF_GAS, nil)) || errors.Is(epsilon, PVMExitTuple(HALT, nil)) {
		return newPc, newGas, newReg, newMem, epsilon
	} else if errors.Is(epsilon, PVMExitTuple(PAGE_FAULT, newMem)) { // test page fault
		return newPc, newGas, newReg, newMem, epsilon
	} else if errors.Is(epsilon, PVMExitTuple(HOST_CALL, newMem)) { // test host call
		return newPc, newGas, newReg, newMem, epsilon
	} else {
		return newPc, newGas, newReg, newMem, epsilon
	}
}

func psi1(pc, gas, reg, mem int) (int, int, int, int, error) {
	var (
		newPc  = pc + 1 // + skip()
		newGas = gas - 10
		newReg = reg
		newMem = mem
	)

	if newGas < 0 {
		return newPc, newGas, newReg, newMem, PVMExitTuple(OUT_OF_GAS, nil)
	}
	if newMem == 0 { // mock "0" for page fault
		return newPc, newGas, newReg, newMem, PVMExitTuple(PAGE_FAULT, uint64(newMem))
	}
	// return newPc, newGas, newReg, newMem, PVMExitTuple(CONTINUE, nil)
	return newPc, newGas, newReg, newMem, PVMExitTuple(HOST_CALL, uint64(newMem))
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
		wantExit    ExitReasonTypes
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
		expectedExitReason ExitReasonTypes
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
