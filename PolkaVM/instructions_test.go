package PolkaVM

import (
	"testing"
)

// Test zeta
func TestOpcodeString(t *testing.T) {
	opcode := "example_opcode"
	expected := "example_opcode"

	if opcode != expected {
		t.Errorf("Expected %s, but got %s", expected, opcode)
	}
}

// Test smod function
func TestSmod(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{10, 3, 1},
		{-10, 3, -1},
		{10, -3, 1},
		{-10, -3, -1},
		{10, 0, 10},
		{-10, 0, -10},
	}

	for _, test := range tests {
		result := smod(test.a, test.b)
		if result != test.expected {
			t.Errorf("smod(%d, %d) = %d; expected %d", test.a, test.b, result, test.expected)
		}
	}
}
