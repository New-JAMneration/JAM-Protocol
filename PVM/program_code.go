package PVM

import (
	"encoding/binary"
	"math/bits"
)

type JumpTable struct {
	Data   []byte // j
	Length uint32 // z
	Size   uint32 // |j|
}

// 0x01 bit stores whether an index is the start of a instruction
// 0x02 bit stores whether an index is the start of a basic block
type Bitmask []byte

func MakeBitMasks(instruction []byte, bitmaskData []byte) (Bitmask, ExitReason) {
	instSize := len(instruction)
	bitmaskSize := instSize / 8
	if instSize%8 > 0 {
		bitmaskSize++
	}

	if len(bitmaskData) != int(bitmaskSize) {
		pvmLogger.Errorf("bitmask has incorrect size: expected %d, got %d", bitmaskSize, len(bitmaskData))
		return nil, ExitPanic
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

	return bitmask, ExitContinue
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
	InstructionData ProgramCode // c , includes opcodes & instruction variables
	Bitmasks        Bitmask     // k
	JumpTable       JumpTable   // j, z, |j|
	InstrCount      uint64
}

// DeBlobProgramCode deblob code, jump table, bitmask | A.2
func DeBlobProgramCode(data []byte) (_ Program, _ ExitReason) {
	// E_(|j|) : size of jumpTable
	jumpTableSize, dataUsed, exitReason := ReadUintVariable(data)
	if exitReason != ExitContinue {
		pvmLogger.Errorf("jumpTableSize ReadUintVariable error")
		return Program{}, ExitPanic
	}
	data = data[dataUsed:]

	// E_1(z) : length of jumpTableLength
	jumpTableLength, exitReason := decodeUintFixedLength(data, 1)
	if exitReason != ExitContinue {
		pvmLogger.Errorf("jumpTableLength decodeUintFixedLength error")
	}
	data = data[1:]

	// E_(|c|) : size of instructions
	instSize, dataUsed, exitReason := ReadUintVariable(data)
	if exitReason != ExitContinue {
		pvmLogger.Errorf("instSize ReadUintVariable error")
		return Program{}, ExitPanic
	}
	data = data[dataUsed:]

	if jumpTableLength*jumpTableSize >= 1<<32 {
		pvmLogger.Errorf("jump table size %d bits exceed litmit of 32 bits", jumpTableLength*jumpTableSize)
		return Program{}, ExitPanic
		// panic("the jump table's size is supposed to be at most 32 bits")
	}

	// E_z(j) = jumpTableSize * jumpTableLength = E_(|j|) * E_1(z)
	jumpTableData, data, err := ReadBytes(data, jumpTableLength*jumpTableSize)
	if err != nil {
		pvmLogger.Errorf("jumpTableData ReadBytes error: %v", err)
		return Program{}, ExitPanic
	}

	instructions := data[:instSize]
	bitmaskData := data[instSize:]
	bitmask, exitReason := MakeBitMasks(instructions, bitmaskData)
	if exitReason == ExitPanic {
		// A.2 if bitmasks cannot fit instructions, return panic
		return Program{}, ExitPanic
	}

	return Program{
		JumpTable: JumpTable{
			Data:   jumpTableData,           // j
			Length: uint32(jumpTableLength), // z
			Size:   uint32(jumpTableSize),   // |j|
		},
		Bitmasks:        bitmask,      // k
		InstructionData: instructions, // c
	}, ExitContinue
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

func ReadUintVariable(data []byte) (uint64, int, ExitReason) {
	if len(data) < 1 {
		pvmLogger.Errorf("readUintVariable failed: no data to deserialize U64")
		return 0, 0, ExitPanic
	}
	prefix := data[0]
	if prefix < 0x80 {
		return uint64(prefix), 1, ExitContinue
	}

	if prefix == 0xFF {
		if len(data) < 9 {
			pvmLogger.Errorf("readUintVariable: not enough data for 8-byte payload")
			return 0, 0, ExitPanic
		}

		return binary.LittleEndian.Uint64(data[1:9]), 9, ExitContinue
	}

	l := bits.LeadingZeros8(^prefix)
	needed := l + 1
	if len(data) < needed {
		pvmLogger.Errorf("readUintVariable failed:not enough data for 8-byte U64")
		return 0, 0, ExitPanic
	}

	base := 0xFF - (uint8(1) << (8 - uint(l))) + 1
	floorVal := uint64(prefix - base)

	var x uint64
	switch l {
	case 1:
		x = (floorVal << 8) | uint64(data[1])
	case 2:
		x = (floorVal << 16) | uint64(binary.LittleEndian.Uint16(data[1:3]))
	case 3:
		if len(data) >= 5 {
			x = (floorVal << 24) | (uint64(binary.LittleEndian.Uint32(data[1:5])) & 0xFFFFFF)
		} else {
			x = (floorVal << 24) | uint64(data[1]) | uint64(data[2])<<8 | uint64(data[3])<<16
		}
	case 4:
		x = (floorVal << 32) | uint64(binary.LittleEndian.Uint32(data[1:5]))
	default:
		remainder := uint64(0)
		for i := range l {
			remainder |= uint64(data[i+1]) << (8 * uint(i))
		}
		x = (floorVal << (8 * uint(l))) | remainder
	}

	if x < (uint64(1) << (7 * uint(l))) {
		pvmLogger.Errorf("readUintVariable: invalid encoding")
		return 0, 0, ExitPanic
	}
	return x, needed, ExitContinue
}

func decodeUintFixedLength(data []byte, l int) (uint64, ExitReason) {
	if len(data) < l {
		pvmLogger.Errorf("not enough data to read a uint: got %d", len(data))
		return 0, ExitPanic
	}
	switch l {
	case 1:
		return uint64(data[0]), ExitContinue
	case 2:
		return uint64(binary.LittleEndian.Uint16(data[0:2])), ExitContinue
	case 3:
		if len(data) >= 4 {
			return uint64(binary.LittleEndian.Uint32(data[0:4])) & 0x00FFFFFF, ExitContinue
		}
		return decodeRemainderManual(data, l), ExitContinue
	case 4:
		return uint64(binary.LittleEndian.Uint32(data[0:4])), ExitContinue
	case 5, 6, 7:
		return decodeRemainderLong(data, l), ExitContinue
	default:
		pvmLogger.Errorf("invalid number of octets to read: got %d", l)
		return 0, ExitPanic
	}
}

func decodeRemainderLong(data []byte, l int) uint64 {
	if len(data) >= 9 {
		mask := (uint64(1) << (uint(l) * 8)) - 1
		return binary.LittleEndian.Uint64(data[1:9]) & mask
	}
	return decodeRemainderManual(data, l)
}

func decodeRemainderManual(data []byte, l int) uint64 {
	var remainder uint64
	for i := 0; i < l && i+1 < len(data); i++ {
		remainder |= uint64(data[i+1]) << (8 * i)
	}
	return remainder
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
