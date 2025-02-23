// nolint:ST1003
package PolkaVM

import (
	"io"
	"os"
	"testing"
)

func TestLoadPVMFile(t *testing.T) {
	filenames := []string{
		"test-file/jam-bootstrap-service.pvm",
	}

	for _, filename := range filenames {
		data, err := ReadFile(filename)
		if err != nil {
			t.Errorf("Error reading %s: %v", filename, err)
			continue
		}

		// TODO : registers & memory initialization
		programCode, _, _, err := SingleInitializer(data, []byte{})
		if err != nil {
			t.Errorf("Error parsing %s: %v", filename, err)
		}

		// exitReason will not be used in this test
		programBlob, _ := DeBlobProgramCode(programCode)

		expected := map[string]int{
			"InstructionDataSize": 53963,
			"JumpTableEntrySize":  2,
			"JumpTableSize":       961,
			"OpcodeBitMaskSize":   6746,
		}
		if len(programBlob.InstructionData) != expected["InstructionDataSize"] {
			t.Errorf("Expected %d, but got %d", expected["InstructionDataSize"], len(programBlob.InstructionData))
		}

		bitmaskSize := len(programBlob.Bitmasks) / 8
		if len(programBlob.Bitmasks)%8 > 0 {
			bitmaskSize++
		}
		if bitmaskSize != expected["OpcodeBitMaskSize"] {
			t.Errorf("Expected %d, but got %d", expected["OpcodeBitMaskSize"], bitmaskSize)
		}
		// only test instructions and bitmasks since if jump table size is not correct
		//     then instructions and bitmasks must be wrong
	}
}

func ReadFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func TestSkip(t *testing.T) {
	bitmask := []byte{1, 0, 0, 1, 0, 0, 0, 1, 0, 0, 1, 0, 1}
	skipIndex := []int{0, 3, 7, 10}
	expectedDistance := []int{3, 4, 3, 2}
	for i := 0; i < len(skipIndex); i++ {
		if skip(skipIndex[i], bitmask) != uint32(expectedDistance[i]) {
			t.Errorf("Expected %d, but got %d", expectedDistance[i], skip(skipIndex[i], bitmask))
		}
	}
}

func TestInBasicBlock(t *testing.T) {
	// true , false -> no such opcode defined , false -> bitmask do not fit the opcode
	data := [][]byte{
		{0, 40, 32, 31, 29},
		{0, 39, 1, 2, 3},
		{0, 40, 0, 0, 0},
	}
	bitmask := [][]byte{{0, 1, 0, 0, 0}, {0, 1, 0, 0, 0}, {0, 0, 1, 0, 0}}
	expected := []bool{true, false, false}
	for i := range len(data) {
		if inBasicBlock(data[i], bitmask[i], 1) != expected[i] {
			t.Errorf("Expected %t, but got %t", expected[i], inBasicBlock(data[i], bitmask[i], 1))
		}
	}
}
