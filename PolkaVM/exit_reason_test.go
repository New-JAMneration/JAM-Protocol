// test.go
package PolkaVM

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

func TestExitReasonsTypes(t *testing.T) {
	tests := []struct {
		reason ExitReasonTypes
		want   string
	}{
		{CONTINUE, "Continue (▸)"},
		{HALT, "Regular halt (∎)"},
		{PANIC, "Panic (☇)"},
		{OUT_OF_GAS, "Out-Of-Gas (∞)"},
		{PAGE_FAULT, "Page fault (F)"},
		{HOST_CALL, "Host-Call identifier (̵h)"},
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
		{1, 0, 10, 0, 0, "Page fault (F) at RAM address: 0"}, // PAGE_FAULT
		{1, 0, 0, 0, 1, "Out-Of-Gas (∞)"},                    // OUT_OF_GAS
		{1, 0, 10, 0, 1, "Host-Call identifier (̵h): 1"},     // HOST_CALL
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
	newPc, newGas, newReg, newMem, epsilon := psi1(1, 2, 3, pc, gas, reg, mem)

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

func psi1(c, k, j, pc, gas, reg, mem int) (int, int, int, int, error) {
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
	testCases := []struct {
		name               string
		target             uint32
		condition          bool
		basicBlocks        []uint32
		expectedExitReason ExitReasonTypes
		expectedPC         uint32
	}{
		{
			"BranchNotTaken",
			8,
			false,
			[]uint32{0, 4, 8},
			CONTINUE,
			1,
		},
		{
			"BranchTaken",
			8,
			true,
			[]uint32{0, 4, 8},
			CONTINUE,
			8,
		},
		{
			"BranchToInvalidTarget",
			10,
			true,
			[]uint32{0, 4, 8},
			PANIC,
			1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exitReason, newPC := Branch(tc.target, tc.condition, tc.basicBlocks)
			if exitReason != tc.expectedExitReason {
				t.Errorf("Expected exit reason %v, but got %v", tc.expectedExitReason, exitReason)
			}
			if newPC != tc.expectedPC {
				t.Errorf("Expected PC %d, but got %d", tc.expectedPC, newPC)
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
			4,
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
			"DjumpToZero",
			0,
			[]uint32{4, 8, 12},
			[]uint32{0, 4, 8},
			PANIC,
			0,
		},
		{
			"DjumpToSpecialCase",
			math.MaxUint32 - ZZ + 1,
			[]uint32{4, 8, 12},
			[]uint32{0, 4, 8},
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
