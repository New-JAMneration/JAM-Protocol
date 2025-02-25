// nolint:ST1003
package PolkaVM

import (
	"io"
	"os"
	"reflect"
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
	filenames := []string{
		"test-file/jam-bootstrap-service.pvm",
	}

	for _, filename := range filenames {
		data, err := ReadFile(filename)
		if err != nil {
			t.Errorf("Error reading %s: %v", filename, err)
			continue
		}

		programCode, _, _, err := SingleInitializer(data, []byte{})
		if err != nil {
			t.Errorf("Error parsing %s: %v", filename, err)
		}

		// exitReason will not be used in this test
		programBlob, _ := DeBlobProgramCode(programCode)

		// the expected is stick to pvm debugger and only get the program counter < 40 instructions
		expected := [][]byte{
			{0x28, 0x67, 0x17, 0x00, 0x00},
			{0x28, 0xc3, 0x1f, 0x00, 0x00},
			{0x28, 0x6d, 0x41},
			{0x95, 0x11, 0xa0, 0xfe},
			{0x7b, 0x10, 0x58, 0x01},
			{0x7b, 0x15, 0x50, 0x01},
			{0x7b, 0x16, 0x48, 0x01},
			{0x64, 0x96},
			{0x7b, 0x18, 0x18},
			{0x82, 0x8a, 0x08},
			{0x14, 0x08, 0xf1, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0},
		}

		for pc, j := 0, 0; pc < 40; j++ {
			l := skip(pc, programBlob.Bitmasks)
			if !reflect.DeepEqual(expected[j], programBlob.InstructionData[pc:pc+int(l)+1]) {
				t.Errorf("Expected %v, but got %v", expected[j], programBlob.InstructionData[pc:pc+int(l)+1])
			}
			pc = pc + 1 + int(l)
			if pc > 40 {
				break
			}
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
