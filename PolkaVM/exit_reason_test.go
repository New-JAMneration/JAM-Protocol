// test.go
package PolkaVM

import (
	"errors"
	"fmt"
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
			_, _, _, _, got := ExPsi(tt.p, tt.pc, tt.gas, tt.reg, tt.mem)
			gotStr := got.Error()
			if gotStr != tt.want {
				t.Errorf("ExPsi(%d, %d, %d, %d, %d) = %v, want %v", tt.p, tt.pc, tt.gas, tt.reg, tt.mem, got, tt.want)
			}
		})
	}
}

// test all exit types of invocation function
func ExPsi(p, pc, gas, reg, mem int) (int, int, int, int, error) {
	// call Psi1 to renew states first
	newPc, newGas, newReg, newMem, epsilon := Psi1(1, 2, 3, pc, gas, reg, mem)

	if errors.Is(epsilon, PVMExitTuple(CONTINUE, nil)) {
		return ExPsi(p, newPc, newGas, newReg, newMem)
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

func Psi1(c, k, j, pc, gas, reg, mem int) (int, int, int, int, error) {
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
