// test.go
package PolkaVM

import (
	"errors"
	"fmt"
	"reflect"
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

func TestParseMemoryAccessError(t *testing.T) {
	testCases := []struct {
		name               string
		invalidAddresses   []uint64
		expectedExitReason ExitReasonTypes
		expectedError      error
	}{
		{
			name:               "NoError",
			invalidAddresses:   []uint64{}, // Changed to a slice
			expectedExitReason: CONTINUE,
			expectedError:      nil,
		},
		{
			"PageFaultError1",
			[]uint64{0x1000000, 0x2000000, 0x3000000},
			PAGE_FAULT,
			PVMExitTuple(PAGE_FAULT, 0x1000000/ZP),
		},
		{
			"LowAddressAccessError",
			[]uint64{0x100, 0x2000},
			PANIC,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exitReason, err := ParseMemoryAccessError(tc.invalidAddresses)
			if exitReason != tc.expectedExitReason {
				t.Errorf("Expected exit reason %v, but got %v", tc.expectedExitReason, exitReason)
			}
			if err != nil && tc.expectedError != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error message %q, but got %q", tc.expectedError, err)
			}
		})
	}
}

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
			name:           "all valid",
			readAddresses:  []uint64{1, 2, 3},
			writeAddresses: []uint64{3, 4, 5},
			readable:       map[int]bool{1: true, 2: true, 3: true, 4: true},
			writeable:      map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true},
			want:           []uint64{},
		},
		{
			name:           "one read address invalid",
			readAddresses:  []uint64{1, 2, 4},
			writeAddresses: []uint64{3, 4, 5},
			readable:       map[int]bool{1: true, 2: true, 3: true},
			writeable:      map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true},
			want:           []uint64{4},
		},
		{
			name:           "one write address invalid",
			readAddresses:  []uint64{1, 2, 3},
			writeAddresses: []uint64{3, 4, 6},
			readable:       map[int]bool{1: true, 2: true, 3: true, 4: true},
			writeable:      map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true},
			want:           []uint64{6},
		},
		{
			name:           "multiple invalid addresses",
			readAddresses:  []uint64{1, 2, 4},
			writeAddresses: []uint64{3, 4, 6},
			readable:       map[int]bool{1: true, 2: true, 3: true},
			writeable:      map[int]bool{1: true, 2: true, 3: true, 5: true},
			want:           []uint64{4, 6},
		},
		{
			name:           "empty input",
			readAddresses:  []uint64{},
			writeAddresses: []uint64{},
			readable:       map[int]bool{},
			writeable:      map[int]bool{},
			want:           []uint64{},
		},
		{
			name:           "duplicate invalid addresses",
			readAddresses:  []uint64{1, 2, 4, 4},
			writeAddresses: []uint64{3, 4, 6, 6},
			readable:       map[int]bool{1: true, 2: true, 3: true},
			writeable:      map[int]bool{1: true, 2: true, 3: true, 5: true},
			want:           []uint64{4, 6},
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
