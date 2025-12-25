package PVM

import (
	"errors"
	"fmt"
)

type JumpTable struct {
	Data   []byte // j
	Length uint32 // z
	Size   uint32 // |j|
}

// 0x01 bit stores whether an index is the start of a instruction
// 0x02 bit stores whether an index is the start of a basic block
type Bitmask []byte

func MakeBitMasks(instruction []byte, bitmaskData []byte) (Bitmask, error) {
	instSize := len(instruction)
	bitmaskSize := instSize / 8
	if instSize%8 > 0 {
		bitmaskSize++
	}

	if len(bitmaskData) != int(bitmaskSize) {
		return nil, fmt.Errorf("bitmask has incorrect size: expected %d, got %d", bitmaskSize, len(bitmaskData))
	}

	bitmask := make(Bitmask, instSize)
	prev := 0
	for i := range instSize {
		if bitmaskData[i/8]&(1<<(i%8)) > 0 {
			bitmask[i] = 0x01

			if i == 0 || isBasicBlockTerminationInstruction(instruction[prev]) {
				bitmask[i] |= 0x02
			}

			prev = i
		}
	}

	return bitmask, nil
}

// returns false if the address is invalid
// this is technically the wrong, but it makes life simple.
func (bitmask Bitmask) IsStartOfInstruction(addr int) bool {
	if addr < 0 || addr >= len(bitmask) {
		return false
	}

	return bitmask[addr] > 0
}

// checks both start of a basic block + start of an instruction
// returns false if the address is invalid
func (bitmask Bitmask) IsStartOfBasicBlock(addr ProgramCounter) bool {
	if addr >= ProgramCounter(len(bitmask)) {
		return false
	}

	return bitmask[addr] == 0x03
}

func isBasicBlockTerminationInstruction(opcode byte) bool {
	switch opcode {
	case 0, 1, 40, 50, 180:
		return true
	}

	if (opcode >= 80 && opcode <= 90) || (opcode >= 170 && opcode <= 175) {
		return true
	}

	return false
}

// type BasicBlock [][]byte // each sequence is a instruction
type Program struct {
	InstructionData []byte    // c , includes opcodes & instruction variables
	Bitmasks        Bitmask   // k
	JumpTable       JumpTable // j, z, |j|
}

func (p *Program) RunHostCallFunc(operationType OperationType) Omega {
	return HostCallFunctions[operationType]
}

// DeBlobProgramCode deblob code, jump table, bitmask | A.2
func DeBlobProgramCode(data []byte) (_ Program, exitReason error) {
	// E_(|j|) : size of jumpTable
	jumpTableSize, data, err := ReadUintVariable(data)
	if err != nil {
		return Program{}, fmt.Errorf("jumpTableSize ReadUintVariable error: %w", err)
	}

	// E_1(z) : length of jumpTableLength
	jumpTableLength, data, err := ReadUintFixed(data, 1)
	if err != nil {
		return Program{}, fmt.Errorf("jumpTableLength ReadUintFixed error: %w", err)
	}
	// E_(|c|) : size of instructions
	instSize, data, err := ReadUintVariable(data)
	if err != nil {
		return Program{}, fmt.Errorf("instSize ReadUintVariable error: %w", err)
	}

	if jumpTableLength*jumpTableSize >= 1<<32 {
		return Program{}, fmt.Errorf("jump table size %d bits exceed litmit of 32 bits", jumpTableLength*jumpTableSize)
		// panic("the jump table's size is supposed to be at most 32 bits")
	}

	// E_z(j) = jumpTableSize * jumpTableLength = E_(|j|) * E_1(z)
	jumpTableData, data, err := ReadBytes(data, jumpTableLength*jumpTableSize)
	if err != nil {
		return Program{}, fmt.Errorf("jumpTableData ReadBytes error: %w", err)
	}

	instructions := data[:instSize]
	bitmaskData := data[instSize:]
	bitmask, err := MakeBitMasks(instructions, bitmaskData)
	if err != nil {
		// A.2 if bitmasks cannot fit instructions, return panic
		return Program{}, PVMExitTuple(PANIC, nil)
	}

	return Program{
		JumpTable: JumpTable{
			Data:   jumpTableData,           // j
			Length: uint32(jumpTableLength), // z
			Size:   uint32(jumpTableSize),   // |j|
		},
		Bitmasks:        bitmask,      // k
		InstructionData: instructions, // c
	}, PVMExitTuple(CONTINUE, nil)
}

// skip computes the distance to the next opcode  A.3
func skip(pc int, bitmask Bitmask) uint32 {
	j := 1
	for ; pc+j < len(bitmask); j++ {
		if bitmask.IsStartOfInstruction(j + pc) {
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

// GP 0.6.7 , for checking basic block first opcode validity
func (code ProgramCode) isOpcodeValid(pc ProgramCounter) bool {
	if _, opcodeExists := zeta[opcode(code[pc])]; opcodeExists {
		return true
	}

	return false
}

// GP 0.6.7 formula A.19
func (code ProgramCode) isOpcode(pc ProgramCounter) opcode {
	if _, opcodeExists := zeta[opcode(code[pc])]; opcodeExists {
		return opcode(code[pc])
	}
	return 0
}
