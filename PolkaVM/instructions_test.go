package PolkaVM

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
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

// Test pvm instructions testvector here
const PVM_TEST_VECTORS_DIR = "../pkg/test_data/instruction/pvm/"

func LoadInstructionTestCase(filename string) (InstructionTestCase, error) {
	file, err := os.Open(filename)
	if err != nil {
		return InstructionTestCase{}, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return InstructionTestCase{}, err
	}

	var testCase InstructionTestCase
	err = json.Unmarshal(byteValue, &testCase)
	if err != nil {
		return InstructionTestCase{}, err
	}

	return testCase, nil
}

func GetTestJsonFiles(dir string) []string {
	jsonFiles := []string{}

	f, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		return nil
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			jsonFiles = append(jsonFiles, file.Name())
		}
	}

	return jsonFiles
}

// TODO: MemoryChunks and PageMaps
func TestInstruction(t *testing.T) {
	dir := filepath.Join(PVM_TEST_VECTORS_DIR, "programs")
	jsonFiles := GetTestJsonFiles(dir)
	if len(jsonFiles) == 0 {
		t.Fatal("No test JSON files found")
	}

	for _, file := range jsonFiles {
		// Control specific test case
		if file != "inst_load_imm_64.json" {
			continue
		}
		t.Run(file, func(t *testing.T) {
			filename := filepath.Join(dir, file)
			testCase, err := LoadInstructionTestCase(filename)
			if err != nil {
				t.Fatalf("Error loading test case %s: %v", file, err)
			}

			ourStatus, pc, gas, reg, _ := SingleStepInvoke(
				testCase.ProgramBlob,
				testCase.InitialProgramCounter,
				testCase.InitialGas,
				testCase.InitialRegisters,
				Memory{},
			)

			if pc != testCase.ExpectedProgramCounter {
				t.Errorf("expected PC %d, got %d", testCase.ExpectedProgramCounter, pc)
			}
			if gas != testCase.ExpectedGas {
				t.Errorf("expected gas %d, got %d", testCase.ExpectedGas, gas)
			}
			if !reflect.DeepEqual(reg, testCase.ExpectedRegisters) {
				t.Errorf("expected registers %v, got %v", testCase.ExpectedRegisters, reg)
			}
			if ourStatus.Error() != testCase.ExpectedStatus {
				t.Errorf("expected status %v, got %v", testCase.ExpectedStatus, ourStatus.Error())
			}
		})
	}
}
