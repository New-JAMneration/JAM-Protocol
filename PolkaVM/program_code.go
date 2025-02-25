package PolkaVM

import (
	"errors"
)

// type BasicBlock [][]byte // each sequence is a instruction
type ProgramBlob struct {
	InstructionData []byte   // c , includes opcodes & instruction variables
	Bitmasks        []bool   // k
	JumpTables      []uint64 // j
	JumpTableLength uint64
}

// DeBlobProgramCode deblob code, jump table, bitmask | A.2
func DeBlobProgramCode(data []byte) (_ ProgramBlob, exitReason error) {
	// E_(|j|) : size of jumpTable
	jumpTableSize, data, err := ReadUintVariable(data)
	if err != nil {
		return
	}
	// E_1(z) : length of jumpTableLength
	jumpTableLength, data, err := ReadUintFixed(data, 1)
	if err != nil {
		return
	}
	// E_(|c|) : size of instructions
	instSize, data, err := ReadUintVariable(data)
	// E_z(j) = jumpTableSize * jumpTableLength = E_(|j|) * E_1(z)
	jumpTables := make([]uint64, jumpTableSize)
	for i := 0; i < int(jumpTableSize); i++ {
		tmp, _, err := ReadUintFixed(data, int(jumpTableLength))
		if err != nil {
			return
		}
		data = data[jumpTableLength:]
		jumpTables[i] = uint64(tmp)
	}

	instructions := data[:instSize]
	bitmaskRaw := data[instSize:]

	// A.2 if bitmasks cannot fit instructions, return panic
	bitmaskSize := instSize / 8
	if instSize%8 > 0 {
		bitmaskSize++
	}
	if len(bitmaskRaw) != int(bitmaskSize) {
		return ProgramBlob{}, PVMExitTuple(PANIC, nil)
	}

	bitmask := make([]bool, instSize)
	for i := range instSize {
		bitmask[i] = bitmaskRaw[i/8]&(1<<(i%8)) > 0
	}
	return ProgramBlob{
		JumpTables:      jumpTables,   // j
		Bitmasks:        bitmask,      // k
		InstructionData: instructions, // c
	}, PVMExitTuple(CONTINUE, nil)
}

// skip computes the distance to the next opcode  A.3
func skip(i int, bitmask []bool) uint32 {
	j := 1
	for ; j < len(bitmask); j++ {
		if bitmask[j+i] {
			break
		}
	}
	return uint32(min(24, j-1))
}

func inBasicBlock(data []byte, bitmask []byte, n int) bool {
	if data[n-1] != byte(0) {
		return false
	}

	if bitmask[n] != byte(1) {
		return false
	}

	if _, exists := zeta[opcode(data[n])]; !exists {
		return false
	}

	return true
}

func ReadUintVariable(data []byte) (uint64, []byte, error) {
	if len(data) < 1 {
		return 0, data, errors.New("not enough data to read a uint")
	}

	firstByte := data[0]
	data = data[1:]

	valueMask, bytesToRead, err := decodeUintFirstByte(firstByte)
	if err != nil {
		return 0, data, err
	}
	valueFromFirstByte := uint64(firstByte & valueMask)
	valueFromRemainingBytes, data, err := ReadUintFixed(data, bytesToRead)
	if err != nil {
		return 0, data, err
	}

	return valueFromFirstByte<<(8*bytesToRead) | valueFromRemainingBytes, data, nil
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
