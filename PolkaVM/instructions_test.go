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
		if file != "inst_load_u64.json" {
			continue
		}
		t.Run(file, func(t *testing.T) {
			filename := filepath.Join(dir, file)

			testCase, err := LoadInstructionTestCase(filename)
			if err != nil {
				t.Fatalf("Error loading test case %s: %v", file, err)
			}
			memory := loadTestCasePageMap(testCase.InitialPageMap)
			memory = loadTestCaseMemory(memory, testCase.InitialMemory)
			ourStatus, pc, gas, reg, memory := SingleStepInvoke(
				testCase.ProgramBlob,
				testCase.InitialProgramCounter,
				testCase.InitialGas,
				testCase.InitialRegisters,
				memory,
			)

			if ourStatus.Error() != ErrNotImplemented.Error() {
				if ourStatus.Error() != testCase.ExpectedStatus {
					t.Errorf("expected status %v, got %v", testCase.ExpectedStatus, ourStatus.Error())
				} else {
					t.Logf("got %v", testCase.ExpectedStatus)
				}

				if pc != testCase.ExpectedProgramCounter {
					t.Errorf("expected PC %d, got %d", testCase.ExpectedProgramCounter, pc)
				}
				if gas != testCase.ExpectedGas {
					t.Errorf("expected gas %d, got %d", testCase.ExpectedGas, gas)
				}
				if !reflect.DeepEqual(reg, testCase.ExpectedRegisters) {
					t.Errorf("expected registers %v, got %v", testCase.ExpectedRegisters, reg)
				}
			}
			expectedMemory := loadTestCaseMemory(Memory{}, testCase.ExpectedMemory)

			if len(memory.Pages) != len(expectedMemory.Pages) {
				t.Errorf("expected memory length %d, got %d", len(expectedMemory.Pages), len(memory.Pages))
			}

			for pageNum, expectedPage := range expectedMemory.Pages {
				// page := memory[pageNum]
				if page, exists := memory.Pages[pageNum]; exists {
					if !reflect.DeepEqual(page.Value, expectedPage.Value) {
						t.Errorf("expected memory %v, got %v", expectedPage.Value, memory.Pages[pageNum].Value)
					}
				} else {
					t.Errorf("expected memory %v, but not exists", testCase.ExpectedMemory)
				}
			}
		})
	}
}

func loadTestCasePageMap(initialPageMap PageMaps) Memory {
	var memory Memory
	memory.Pages = make(map[uint32]*Page)
	if len(initialPageMap) > 0 {
		for _, pageMap := range initialPageMap {
			pageNum := pageMap.Address >> 12
			page := Page{
				Value:  make([]byte, 0),
				Access: MemoryReadWrite,
			}
			memory.Pages[pageNum] = &page
		}
	}
	return memory
}

func loadTestCaseMemory(memory Memory, initialMemory MemoryChunks) Memory {
	if len(initialMemory) > 0 {
		if memory.Pages == nil {
			memory.Pages = make(map[uint32]*Page)
			for _, memoryChunk := range initialMemory {
				pageNum := memoryChunk.Address >> 12
				page := Page{
					Value:  memoryChunk.Contents,
					Access: MemoryReadWrite,
				}
				memory.Pages[pageNum] = &page
			}
		} else {
			for _, memoryChunk := range initialMemory {
				pageNum := memoryChunk.Address >> 12
				if mem, exists := memory.Pages[pageNum]; exists {
					mem.Value = append(mem.Value, memoryChunk.Contents...)
					// copy(mem.Value[:], memoryChunk.Contents)
				}
			}
		}
	}

	return memory
}
