package PolkaVM

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

// type BasicBlock [][]byte // each sequence is a instruction
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
	// fmt.Println(data[:16])
	jumpTableSize, data, err := ReadUintVariable(data)
	if err != nil {
		return
	}
	// fmt.Println(data[:16])
	// E_1(z) : length of jumpTableLength
	jumpTableLength, data, err := ReadUintFixed(data, 1)
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

	// data = data[:jumpTableLength*jumpTableSize]
	// bitmasks
	instructions := data[:instSize]
	bitmaskRaw := data[instSize:]

	// A.2 if bitmasks cannot fit instructions, return panic
	bitmaskSize := instSize / 8
	if instSize%8 > 0 {
		bitmaskSize++
	}
	if len(bitmaskRaw) != int(bitmaskSize) {
		return ProgramBlob{}, PANIC
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

// skip computes the distance to the next opcode  A.3
func skip(i int, bitmask []byte) uint32 {
	j := 1
	for ; j < len(bitmask); j++ {
		if bitmask[j+i] == byte(1) {
			break
		}
	}
	return uint32(min(24, j))
}

func inBasicBlock(data []byte, bitmask []byte, n int) bool {
	if data[n-1] != byte(0) {
		return false
	}

	if bitmask[n] != byte(1) {
		return false
	}

	if _, exists := Zeta[opcode(data[n])]; !exists {
		return false
	}

	return true
}

func ReadUintVariable(data []byte) (uint64, []byte, error) {
	if len(data) < 1 {
		return 0, data, fmt.Errorf("not enough data to read a uint")
	}

	firstByte := data[0]
	data = data[1:]

	valueMask, bytesToRead, err := decodeUintFirstByte(firstByte)
	if err != nil {
		return 0, data, err
	}
	fmt.Println("byteToRead : ", bytesToRead)
	fmt.Println("valueMask : ", valueMask)
	valueFromFirstByte := uint64(firstByte & valueMask)
	fmt.Println("valueFromFirstByte : ", valueFromFirstByte)
	valueFromRemainingBytes, data, err := ReadUintFixed(data, bytesToRead)
	if err != nil {
		return 0, data, err
	}

	return valueFromFirstByte<<(8*bytesToRead) | valueFromRemainingBytes, data[bytesToRead:], nil
}

// return (mask, bytes to read)
func decodeUintFirstByte(firstByte byte) (byte, int, error) {
	leadingBits := []byte{
		0x80, 0x40, 0x20, 0x10, 0x08, 0x04, 0x02, 0x01,
	}

	lengthMask := byte(0)
	for index, leadingBit := range leadingBits {
		// first N + 1 bits are N bits of 1 followed by 1 bit of 0
		// e.g. N = 0 and the first bit is 0
		//      N = 3 and the first 4 bits are 1110
		if firstByte&(lengthMask|leadingBit) == lengthMask {
			return 0xff - (lengthMask | leadingBit), index, nil
		}

		lengthMask |= leadingBit
	}

	return 0, 8, nil
}
