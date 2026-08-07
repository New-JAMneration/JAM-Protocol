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

			if i == 0 || IsBlockTerminator(instruction[prev]) {
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

// type BasicBlock [][]byte // each sequence is a instruction
type Program struct {
	InstructionData ProgramCode // c , includes opcodes & instruction variables
	Bitmasks        Bitmask     // k
	JumpTable       JumpTable   // j, z, |j|

	Instrs     []InstrMeta  // pre-decoded instruction metadata (flat array)
	BlockAt    []*BlockMeta // PC-indexed: BlockAt[pc] non-nil if pc starts a basic block
	InstrIdxAt []int32      // PC-indexed: InstrIdxAt[pc] = index into Instrs[], -1 if not an instruction start
}

// DeBlobProgramCode is deblob(pvm_blob, ι): it decodes the blob and
// validates both the program as a whole and pc as an entry point for the
// instruction counter, returning ExitPanic where the Gray Paper yields an error.
func DeBlobProgramCode(data []byte, pc uint64) (Program, ExitReason) {
	prog, exitReason := deblobValidatedProgram(data)
	if exitReason != ExitContinue {
		return Program{}, exitReason
	}

	// 𝔳_inst(c, k, ι) (A.2)
	if !prog.ValidInstructionAt(pc) {
		pvmLogger.Errorf("instruction counter %d is not a valid entry point", pc)
		return Program{}, ExitPanic
	}

	return prog, ExitContinue
}

// deblobValidatedProgram is the entry-point-independent half of deblob: it
// already stored a decoded program, only 𝔳_inst(c, k, ι) is rechecked; otherwise
// falls back to full deblob(p, ι) (e.g. tests that omit Program).
func IntegratedProgramForInvoke(integrated IntegratedPVMType) (*Program, ExitReason) {
	if integrated.Program != nil {
		return validEntry(integrated.Program, uint64(integrated.PC))
	}
	p, reason := DeBlobProgramCode(integrated.ProgramCode, uint64(integrated.PC))
	if reason != ExitContinue {
		return nil, reason
	}
	return &p, ExitContinue
}

// deblobValidatedProgram is the entry-point-independent half of deblob: it
// decodes the blob and checks 𝔳_blob(c, k, 0). Callers that cache a decoded
// program across invocations (GetOrDeblobProgram) share this result and apply
// the per-entry 𝔳_inst check themselves.
func deblobValidatedProgram(data []byte) (Program, ExitReason) {
	return decodeProgramBlob(data, true)
}

// deblobProgramForGasModel decodes A.9 vector fixtures that may omit a final
// terminator. Production deblobValidatedProgram still enforces 𝔳_blob.
func deblobProgramForGasModel(data []byte) (Program, ExitReason) {
	return decodeProgramBlob(data, false)
}

func decodeProgramBlob(data []byte, requireFinalTerminator bool) (_ Program, _ ExitReason) {
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
		return Program{}, ExitPanic
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
	}

	// E_z(j) = jumpTableSize * jumpTableLength = E_(|j|) * E_1(z)
	jumpTableData, data, err := ReadBytes(data, jumpTableLength*jumpTableSize)
	if err != nil {
		pvmLogger.Errorf("jumpTableData ReadBytes error: %v", err)
		return Program{}, ExitPanic
	}

	if instSize > uint64(len(data)) {
		pvmLogger.Errorf("instruction size %d exceeds remaining blob length %d", instSize, len(data))
		return Program{}, ExitPanic
	}

	instructions := data[:instSize]
	bitmaskData := data[instSize:]
	bitmask, exitReason := MakeBitMasks(instructions, bitmaskData)
	if exitReason == ExitPanic {
		// A.2 if bitmasks cannot fit instructions, return panic
		return Program{}, ExitPanic
	}

	prog := Program{
		JumpTable: JumpTable{
			Data:   jumpTableData,           // j
			Length: uint32(jumpTableLength), // z
			Size:   uint32(jumpTableSize),   // |j|
		},
		Bitmasks:        bitmask,      // k
		InstructionData: instructions, // c
	}

	if exitReason := prog.preDecodeBlocks(); exitReason != ExitContinue {
		return Program{}, exitReason
	}
	// A.2 𝔳_blob: final instruction must be a basic-block terminator.
	if requireFinalTerminator && !prog.finalInstructionIsTerminator() {
		return Program{}, ExitPanic
	}

	return prog, ExitContinue
}

// ValidInstructionAt is 𝔳_inst(c, k, ι): whether the instruction counter
// may enter this program at pc.
//
// Unlike 𝔳_blob, this depends on the entry point rather than the program, so it
// is re-checked on every entry instead of being carried by a cached *Program.
// Three lookups make that free, and the entry point is not always trustworthy:
// the machine and invoke host-calls take it from a guest register.
func (p *Program) ValidInstructionAt(pc uint64) bool {
	return validInst(p.InstructionData, p.Bitmasks, pc)
}

// validInst is 𝔳_inst(c, k, ι) (A.2)
func validInst(c ProgramCode, k Bitmask, i uint64) bool {
	return len(k) == len(c) &&
		i < uint64(len(k)) &&
		k.IsStartOfInstruction(int(i)) &&
		IsValidOpcode(c[i])
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

	if !IsValidOpcode(data[n]) {
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
	return IsValidOpcode(code[pc])
}

// GP 0.6.7 formula A.19
// Kept for SingleStepStateTransition (GP 0.7.2 path); deprecated for v0.8.0 and later.
func (code ProgramCode) isOpcode(pc ProgramCounter) opcode {
	if IsValidOpcode(code[pc]) {
		return opcode(code[pc])
	}
	return 0
}
