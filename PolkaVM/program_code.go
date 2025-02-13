package PolkaVM

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

/*
   instructions will be saved as block-based instructions

*/

type BasicBlock [][]byte // each sequence is a instruction

type ProgramBlob struct {
	InstructionData []byte   // c , includes opcodes & instruction variables
	Bitmasks        []byte   // k
	JumpTables      []uint64 // j
	JumpTableLength uint64
	// JumpTableSize   uint64
	// InstructionSize uint64
}

// DeBlobProgramCode deblob code, jump table, bitmask | A.2
func DeBlobProgramCode(data []byte) (_ ProgramBlob, exitReason ExitReasonTypes) {
	// E_(|j|) : size of jumpTable
	// will rewrite after the refactor of Deserialization is complete
	var jumpTableSize types.U64
	var err error
	for i := 0; i < 8; i++ {
		jumpTableSize, err = utilities.DeserializeU64(data[i:9])
		if err == nil {
			data = data[i:]
			break
		}
	}
	// E_1(z) : length of jumpTableLength
	var jumpTableLength types.U64
	jumpTableLength, err = utilities.DeserializeFixedLength(data[:1], types.U64(1))
	data = data[1:]
	if err != nil {
		return
	}
	// E_(|c|) : size of instructions
	// will rewrite after the refactor of Deserialization is complete
	var instSize types.U64
	for i := 0; i < 8; i++ {
		instSize, err = utilities.DeserializeU64(data[:i])
		if err == nil {
			data = data[i:]
			break
		}
	}
	// E_z(j) = jumpTableSize * jumpTableLength = E_(|j|) * E_1(z)
	jumpTables := make([]uint64, jumpTableSize)
	for i := 0; i < int(jumpTableSize); i++ {
		tmp, _ := utilities.DeserializeFixedLength(data[i*2:i*2+2], types.U64(jumpTableLength))
		jumpTables[i] = uint64(tmp)
	}

	data = data[:jumpTableLength*jumpTableSize]
	// bitmasks
	instructions := data[:instSize]
	bitmaskRaw := data[instSize:]

	// A.2 if bitmasks cannot fit instructions, return panic
	bitmaskSize := instSize / 8
	if instSize%8 > 0 {
		bitmaskSize++
	}
	if len(bitmaskRaw) != int(bitmaskSize) {
		// return PANIC
	}

	bitmask := make([]byte, instSize)
	for i := range instSize {
		bitmask[i] = bitmaskRaw[i/8] & (1 << (i % 8))
	}

	return ProgramBlob{
		JumpTables:      jumpTables,   // j
		Bitmasks:        bitmask,      // k
		InstructionData: instructions, // c
	}, CONTINUE
}

func ReadUintFixed(data []byte, numBytes int) (uint64, []byte, error) {
	if numBytes == 0 {
		return 0, data, nil
	}
	if numBytes > 8 || numBytes < 0 {
		return 0, data, fmt.Errorf("invalid number of octets to read")
	}
	if len(data) < numBytes {
		return 0, data, fmt.Errorf("not enough data to read a uint")
	}

	var result uint64
	for i := 0; i < numBytes; i++ {
		// little-endian
		result |= uint64(data[i]) << (8 * i)
	}

	return result, data[numBytes:], nil
}

func ReadBytes(data []byte, numBytes uint64) ([]byte, []byte, error) {
	if uint64(len(data)) < numBytes {
		return nil, data, fmt.Errorf("not enough data to read %d bytes", numBytes)
	}

	return data[:numBytes], data[numBytes:], nil
}

/*
type ProgramBlob struct {
	// JumpTableSize   uint64
	JumpTableLength uint64
	JumpTables      []uint64 // j
	InstructionSize uint64
	InstructionData []byte // c , includes opcodes & instruction variables
	Bitmasks        []byte // k
}
*/

/*
type StandardProgram struct {
	ROData      []byte
	RWData      []byte
	PaddingPage uint16
	Stack       uint32
	ProgramBlob []byte
}

func ReadFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
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
*/
