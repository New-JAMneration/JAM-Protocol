// nolint:ST1003
package PolkaVM

import (
	"fmt"
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
			// log.Fatalf("Error reading %s: %v", filename, err)
			continue
		}

		// TODO : registers & memory initialization
		programCode, _, _, err := SingleInitializer(data, []byte{})
		fmt.Println(programCode[:10])
		if err != nil {
			t.Errorf("Error parsing %s: %v", filename, err)
			// log.Fatalf("Error parsing %s: %v", filename, err)
		}

		// exitReason will not be used in this test
		programBlob, _ := DeBlobProgramCode(programCode)

		fmt.Println("JumpTableLength : ", programBlob.JumpTableLength)
		fmt.Println("InstructionData : ", programBlob.InstructionData[:10])
		fmt.Println(len(programBlob.Bitmasks) == len(programBlob.InstructionData))
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
