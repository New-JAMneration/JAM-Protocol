package PolkaVM

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

// Instruction tables

// result of "ζı" should be a opcode
type opcode int

// define "ζ"
var zeta = map[opcode]string{
	// Ins w/o Arg
	0: "trap",
	1: "fallthrough",
	// Ins w/ Arg of One Imm
	10: "ecalli",
	// Ins w/ Arg of One Reg and One Extended Width Imm
	20: "load_imm_64",
	// Ins w/ Arg of Two Imm
	30: "store_imm_u8",
	31: "store_imm_u16",
	32: "store_imm_u32",
	33: "store_imm_u64",
	// Ins w/ Arg of One Offset
	40: "jump",
	// Ins w/ Arg of One Reg & One Imm
	50: "jump_ind",
	51: "load_imm",
	52: "load_u8",
	53: "load_i8",
	54: "load_u16",
	55: "load_i16",
	56: "load_u32",
	57: "load_i32",
	58: "load_u64",
	59: "store_u8",
	60: "store_u16",
	61: "store_u32",
	62: "store_u64",
	// Ins w/ Arg of One Reg & Two Imm
	70: "store_imm_ind_u8",
	71: "store_imm_ind_u16",
	72: "store_imm_ind_u32",
	73: "store_imm_ind_u64",
	// Ins w/ Arg of One Reg & One Imm & One Offset
	80: "load_imm_jump",
	81: "branch_eq_imm",
	82: "branch_ne_imm",
	83: "branch_lt_u_imm",
	84: "branch_le_u_imm",
	85: "branch_ge_u_imm",
	86: "branch_gt_u_imm",
	87: "branch_lt_s_imm",
	88: "branch_le_s_imm",
	89: "branch_ge_s_imm",
	90: "branch_gt_s_imm",
	// Ins w/ Arg of Two Reg
	100: "move_reg",
	101: "sbrk",
	102: "count_set_bits_64",
	103: "count_set_bits_32",
	104: "leading_zero_bits_64",
	105: "leading_zero_bits_32",
	106: "trailing_zero_bits_64",
	107: "trailing_zero_bits_32",
	108: "sign_extend_8",
	109: "sign_extend_16",
	110: "zero_extend_16",
	111: "reverse_bytes",
	// Ins w/ Arg of Two Reg & One Imm
	120: "store_ind_u8",
	121: "store_ind_u16",
	122: "store_ind_u32",
	123: "store_ind_u64",
	124: "load_ind_u8",
	125: "load_ind_i8",
	126: "load_ind_u16",
	127: "load_ind_i16",
	128: "load_ind_u32",
	129: "load_ind_i32",
	130: "load_ind_u64",
	131: "add_imm_32",
	132: "and_imm",
	133: "xor_imm",
	134: "or_imm",
	135: "mul_imm_32",
	136: "set_lt_u_imm",
	137: "set_lt_s_imm",
	138: "shlo_l_imm_32",
	139: "shlo_r_imm_32",
	140: "shar_r_imm_32",
	141: "neg_add_imm_32",
	142: "set_gt_u_imm",
	143: "set_gt_s_imm",
	144: "shlo_l_imm_alt_32",
	145: "shlo_r_imm_alt_32",
	146: "shar_r_imm_alt_32",
	147: "cmov_iz_imm",
	148: "cmov_nz_imm",
	149: "add_imm_64",
	150: "mul_imm_64",
	151: "shlo_l_imm_64",
	152: "shlo_r_imm_64",
	153: "shar_r_imm_64",
	154: "neg_add_imm_64",
	155: "shlo_l_imm_alt_64",
	156: "shlo_r_imm_alt_64",
	157: "shar_r_imm_alt_64",
	158: "rot_r_64_imm",
	159: "rot_r_64_imm_alt",
	160: "rot_r_32_imm",
	161: "rot_r_32_imm_alt",
	// Ins w/ Arg of Two Reg & One Offset
	170: "branch_eq",
	171: "branch_ne",
	172: "branch_lt_u",
	173: "branch_lt_s",
	174: "branch_ge_u",
	175: "branch_ge_s",
	// Ins w/ Arg of Two Reg & Two Imm
	180: "load_imm_jump_ind",
	// Ins w/ Arg of Three Reg
	190: "add_32",
	191: "sub_32",
	192: "mul_32",
	193: "div_u_32",
	194: "div_s_32",
	195: "rem_u_32",
	196: "rem_s_32",
	197: "shlo_l_32",
	198: "shlo_r_32",
	199: "shar_r_32",
	200: "add_64",
	201: "sub_64",
	202: "mul_64",
	203: "div_u_64",
	204: "div_s_64",
	205: "rem_u_64",
	206: "rem_s_64",
	207: "shlo_l_64",
	208: "shlo_r_64",
	209: "shar_r_64",
	210: "and",
	211: "xor",
	212: "or",
	213: "mul_upper_s_s",
	214: "mul_upper_u_u",
	215: "mul_upper_s_u",
	216: "set_lt_u",
	217: "set_lt_s",
	218: "cmov_iz",
	219: "cmov_nz",
	220: "rot_l_64",
	221: "rot_l_32",
	222: "rot_r_64",
	223: "rot_r_32",
	224: "and_inv",
	225: "or_inv",
	226: "xnor",
	227: "max",
	228: "max_u",
	229: "min",
	230: "min_u",
}

// (A.32 v0.6.2) smod function
func smod(a int, b int) int {
	if b == 0 {
		return a
	}

	mod := abs(a) % abs(b)

	if a < 0 {
		return -mod
	}
	return mod
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getRegModIndex(instructionCode []byte, pc ProgramCounter) uint64 {
	return uint64(min(12, (int(instructionCode[pc+1]) % 16)))
}
func getRegFloorIndex(instructionCode []byte, pc ProgramCounter) uint64 {
	return uint64(min(12, (int(instructionCode[pc+1]) >> 4)))
}

func getInstArgOfTwoReg(instructionCode []byte, pc ProgramCounter, reg Registers) (rD uint64, rA uint64, newReg Registers) {
	rD = getRegModIndex(instructionCode, pc)
	newReg[rD] = reg[rD]
	rA = getRegFloorIndex(instructionCode, pc)
	newReg[rA] = reg[rA]

	return rD, rA, newReg
}

// input: instructionCode, programCounter, skipLength, registers, memory
var execInstructions = [230]func([]byte, ProgramCounter, ProgramCounter, Registers, Memory, JumpTable, []bool) (error, ProgramCounter, Gas, Registers, Memory){
	0:  instTrap,
	1:  instFallthrough,
	10: instEcalli,
	20: instLoadImm64, // passed testvector
	// register more instructions here
	100: instMoveReg, // passed testvector
	102: instCountSetBits64,
	103: instCountSetBits32,
	104: instLeadingZeroBits64,
	105: instLeadingZeroBits32,
	106: instTrailZeroBits64,
	107: instTrailZeroBits32,
	108: instSignExtend8,
	109: instSignExtend16,
	110: instZeroExtend16,
	111: instReverseBytes,
	// register more instructions here
}

func instTrap(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}
func instFallthrough(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}
func instEcalli(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)

	lX := min(4, int(skipLength))

	// zeta_{iota+1,...,lX}
	instLength := instructionCode[pc+1 : pc+ProgramCounter(lX)+1]
	x, err := utils.DeserializeFixedLength(instLength, types.U64(lX))
	if err != nil {
		log.Println("insEcalli deserialization raise error:", err)
	}
	nuX, err := SignExtend(lX, uint64(x))
	if err != nil {
		log.Println("insEcalli sign extension raise error:", err)
	}
	return PVMExitTuple(HOST_CALL, nuX), pc, gasDelta, reg, mem
}

// opcode 20
func instLoadImm64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)

	rA := min(12, (int(instructionCode[pc+1]) % 16))
	// zeta_{iota+2,...,+8}
	instLength := instructionCode[pc+2 : pc+10]
	nuX, err := utils.DeserializeFixedLength(instLength, types.U64(8))
	if err != nil {
		log.Println("insLoadImm64 deserialization raise error:", err)
	}
	reg[rA] = uint64(nuX)

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 100
func instMoveReg(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	newReg[rD] = reg[rA]

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 102
func instCountSetBits64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits(regA, 8)
	if err != nil {
		log.Println("insCountSetBits64 raise error:", err)
	}
	var sum uint64 = 0
	for i := 0; i < 64; i++ {
		if bitslice[i] {
			sum++
		}
	}
	newReg[rD] = sum

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 103
func instCountSetBits32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits((regA % (1 << 32)), 4)
	if err != nil {
		log.Println("instCountSetBits32 raise error:", err)
	}
	var sum uint64 = 0
	for i := 0; i < 32; i++ {
		if bitslice[i] {
			sum++
		}
	}
	newReg[rD] = sum

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 104
func instLeadingZeroBits64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits(regA, 8)
	if err != nil {
		log.Println("instLeadingZeroBits64 raise error:", err)
	}
	var n uint64 = 0
	for i := 0; i < 64; i++ {
		n++
		if bitslice[i] {
			break
		}
	}
	newReg[rD] = n

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 105
func instLeadingZeroBits32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits((regA % (1 << 32)), 4)
	if err != nil {
		log.Println("instLeadingZeroBits32 raise error:", err)
	}
	var n uint64 = 0
	for i := 0; i < 32; i++ {
		n++
		if bitslice[i] {
			break
		}
	}
	newReg[rD] = n

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 106
func instTrailZeroBits64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits(regA, 8)
	if err != nil {
		log.Println("instTrailZeroBits64 raise error:", err)
	}
	var n uint64 = 0
	for i := 63; i >= 0; i-- {
		n++
		if bitslice[i] {
			break
		}
	}
	newReg[rD] = n

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 107
func instTrailZeroBits32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits((regA % (1 << 32)), 4)
	if err != nil {
		log.Println("instTrailZeroBits32 raise error:", err)
	}
	var n uint64 = 0
	for i := 31; i >= 0; i-- {
		n++
		if bitslice[i] {
			break
		}
	}
	newReg[rD] = n

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 108
func instSignExtend8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := reg[rA]
	signedInt, err := UnsignedToSigned((regA % (1 << 8)), 1)
	if err != nil {
		log.Println("instSignExtend8 UnsignedToSigned raise error:", err)
	}
	unsignedInt, err := SignedToUnsigned(signedInt, 8)
	if err != nil {
		log.Println("instSignExtend8 SignedToUnsigned raise error:", err)
	}
	newReg[rD] = unsignedInt

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 109
func instSignExtend16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := reg[rA]
	signedInt, err := UnsignedToSigned((regA % (1 << 16)), 2)
	if err != nil {
		log.Println("instSignExtend16 UnsignedToSigned raise error:", err)
	}
	unsignedInt, err := SignedToUnsigned(signedInt, 8)
	if err != nil {
		log.Println("instSignExtend16 SignedToUnsigned raise error:", err)
	}
	newReg[rD] = unsignedInt

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 110
func instZeroExtend16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := reg[rA]
	newReg[rD] = regA % (1 << 16)

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

// opcode 111
func instReverseBytes(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, newReg := getInstArgOfTwoReg(instructionCode, pc, reg)

	// mutation
	regA := types.U64(reg[rA])
	bytes := utils.SerializeFixedLength(regA, types.U64(8))
	var reversedBytes uint64 = 0
	for i := uint8(0); i < 8; i++ {
		reversedBytes = (reversedBytes << 8) | uint64(bytes[i])
	}
	newReg[rD] = reversedBytes

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, newReg, mem
}

func instJump(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	offset, err := decodeOneOffset(instructionCode, pc, skipLength)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	reason, pc := branch(pc+ProgramCounter(offset), true, bitmask)

	return PVMExitTuple(reason, nil), pc, Gas(2), reg, mem
}

func instJumpInd(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instLoadImmJump(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchEqImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchNeImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchLtUImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchLeUImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchGeUImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchGtUImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchLtSImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchLeSImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchGeSImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchGtSImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchEq(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchNe(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchLtU(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchLtS(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchGeU(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instBranchGeS(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}

func instLoadImmJumpInd(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	panic("not implemented")
}
