package PolkaVM

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"golang.org/x/exp/constraints"
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
func smod[T constraints.Signed](a, b T) T {
	if b == 0 {
		return a
	}
	return T(a % b)
}

// input: instructionCode, programCounter, skipLength, registers, memory
var execInstructions = [230]func([]byte, ProgramCounter, ProgramCounter, Registers, Memory, JumpTable, Bitmask) (error, ProgramCounter, Gas, Registers, Memory){
	0:   instTrap,
	1:   instFallthrough,
	10:  instEcalli,
	20:  instLoadImm64, // passed testvector
	40:  instJump,
	50:  instJumpInd,
	80:  instImmediateBranch,
	81:  instImmediateBranch,
	82:  instImmediateBranch,
	83:  instImmediateBranch,
	84:  instImmediateBranch,
	85:  instImmediateBranch,
	86:  instImmediateBranch,
	87:  instImmediateBranch,
	88:  instImmediateBranch,
	89:  instImmediateBranch,
	90:  instImmediateBranch,
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
	170: instBranch,
	171: instBranch,
	172: instBranch,
	173: instBranch,
	174: instBranch,
	175: instBranch,
	180: instLoadImmJumpInd,
	190: instAdd32,
	191: instSub32,
	192: instMul32,
	193: instDivU32,
	194: instDivS32,
	195: instRemU32,
	196: instRemS32,
	197: instShloL32,
	198: instShloR32,
	199: instSharR32,
	200: instAdd64,   // passed testvector
	201: instSub64,   // passed testvector
	202: instMul64,   // passed testvector
	203: instDivU64,  // passed testvector
	204: instDivS64,  // passed testvector
	205: instRemU64,  // passed testvector
	206: instRemS64,  // passed testvector
	207: instShloL64, // passed testvector
	208: instShloR64, // passed testvector
	209: instSharR64, // passed testvector
	210: instAnd,     // passed testvector
	211: instXor,     // passed testvector
	212: instOr,      // passed testvector
	// register more instructions here
}

// opcode 0
func instTrap(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(1)
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 1
func instFallthrough(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 10
func instEcalli(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
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
func instLoadImm64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)

	rA := min(12, (int(instructionCode[pc+1]) % 16))
	// zeta_{iota+2,...,+8}
	instLength := instructionCode[pc+2 : pc+10]
	nuX, err := utils.DeserializeFixedLength(instLength, types.U64(8))
	if err != nil {
		log.Println("insLoadImm64 deserialization raise error:", err)
	}
	reg[rA] = uint64(nuX)

	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 40
func instJump(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	vX, err := decodeOneOffset(instructionCode, pc, skipLength)
	if err != nil {
		return PVMExitTuple(PANIC, nil), pc, Gas(0), reg, mem
	}

	reason, newPC := branch(pc, vX, true, bitmask)

	if reason != CONTINUE {
		return PVMExitTuple(reason, nil), pc, Gas(1), reg, mem
	}

	// TODO double-check gas expenditure
	return PVMExitTuple(reason, nil), newPC, Gas(2), reg, mem
}

// opcode 50
func instJumpInd(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		return PVMExitTuple(PANIC, nil), pc, Gas(0), reg, mem
	}

	dest := uint32(reg[rA] + vX)
	reason, newPC := djump(pc, dest, jumpTable, bitmask)
	if reason != CONTINUE {
		return PVMExitTuple(reason, nil), pc, Gas(1), reg, mem
	}

	return PVMExitTuple(reason, nil), newPC, Gas(2), reg, mem
}

// opcode in [80, 90]
func instImmediateBranch(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	rA, vX, vY, err := decodeOneRegisterOneImmediateAndOneOffset(instructionCode, pc, skipLength)
	if err != nil {
		return PVMExitTuple(PANIC, nil), pc, Gas(0), reg, mem
	}

	branchCondition := false

	switch instructionCode[pc] {
	case 80:
		reg[rA] = vX
		branchCondition = true
	case 81:
		branchCondition = reg[rA] == vX
	case 82:
		branchCondition = reg[rA] != vX
	case 83:
		branchCondition = reg[rA] < vX
	case 84:
		branchCondition = reg[rA] <= vX
	case 85:
		branchCondition = reg[rA] >= vX
	case 86:
		branchCondition = reg[rA] > vX
	case 87:
		branchCondition = int64(reg[rA]) < int64(vX)
	case 88:
		branchCondition = int64(reg[rA]) <= int64(vX)
	case 89:
		branchCondition = int64(reg[rA]) >= int64(vX)
	case 90:
		branchCondition = int64(reg[rA]) > int64(vX)
	default:
		panic("this function is supposed to be called with opcode in [80, 90]")
	}

	reason, newPC := branch(pc, vY, branchCondition, bitmask)
	if reason != CONTINUE {
		return PVMExitTuple(reason, nil), pc, Gas(1), reg, mem
	}

	return PVMExitTuple(reason, nil), newPC, Gas(1), reg, mem
}

// opcode 100
func instMoveReg(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = reg[rA]

	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 102
func instCountSetBits64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
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
	reg[rD] = sum

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 103
func instCountSetBits32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
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
	reg[rD] = sum

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 104
func instLeadingZeroBits64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
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
	reg[rD] = n

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 105
func instLeadingZeroBits32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
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
	reg[rD] = n

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 106
func instTrailZeroBits64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
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
	reg[rD] = n

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 107
func instTrailZeroBits32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
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
	reg[rD] = n

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 108
func instSignExtend8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
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
	reg[rD] = unsignedInt

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 109
func instSignExtend16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
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
	reg[rD] = unsignedInt

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 110
func instZeroExtend16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	regA := reg[rA]
	reg[rD] = regA % (1 << 16)

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 111
func instReverseBytes(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	regA := types.U64(reg[rA])
	bytes := utils.SerializeFixedLength(regA, types.U64(8))
	var reversedBytes uint64 = 0
	for i := uint8(0); i < 8; i++ {
		reversedBytes = (reversedBytes << 8) | uint64(bytes[i])
	}
	reg[rD] = reversedBytes

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode in [170, 175]
func instBranch(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneOffset(instructionCode, pc, skipLength)
	if err != nil {
		return PVMExitTuple(PANIC, nil), pc, Gas(0), reg, mem
	}

	branchCondition := false

	switch instructionCode[pc] {
	case 170:
		branchCondition = reg[rA] == reg[rB]
	case 171:
		branchCondition = reg[rA] != reg[rB]
	case 172:
		branchCondition = reg[rA] < reg[rB]
	case 173:
		branchCondition = int64(reg[rA]) < int64(reg[rB])
	case 174:
		branchCondition = reg[rA] >= reg[rB]
	case 175:
		branchCondition = int64(reg[rA]) >= int64(reg[rB])
	default:
		panic("this function is supposed to be called with opcode in [170, 175]")
	}

	reason, newPC := branch(pc, vX, branchCondition, bitmask)
	if reason != CONTINUE {
		return PVMExitTuple(reason, nil), pc, Gas(1), reg, mem
	}

	return PVMExitTuple(reason, nil), newPC, Gas(2), reg, mem
}

// opcode 180
func instLoadImmJumpInd(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	rA, rB, vX, vY, err := decodeTwoRegistersAndTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		return PVMExitTuple(PANIC, nil), pc, Gas(0), reg, mem
	}

	// per https://github.com/koute/jamtestvectors/blob/master_pvm_initial/pvm/TESTCASES.md#inst_load_imm_and_jump_indirect_invalid_djump_to_zero_different_regs_without_offset_nok
	// the register update should take place even if the jump panics
	reg[rA] = vX

	dest := uint32(reg[rB] + vY)
	reason, newPC := djump(pc, dest, jumpTable, bitmask)
	if reason != CONTINUE {
		return PVMExitTuple(reason, nil), pc, Gas(1), reg, mem
	}

	return PVMExitTuple(reason, nil), newPC, Gas(2), reg, mem
}

// opcode 190
func instAdd32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD], err = SignExtend(4, uint64(uint32(reg[rA]+reg[rB])))
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 191
func instSub32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	bMod32 := uint32(reg[rB])
	reg[rD], err = SignExtend(4, uint64(uint32(reg[rA]+^uint64(bMod32)+1)))

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 192
func instMul32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD], err = SignExtend(4, uint64(uint32(reg[rA]*reg[rB])))
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 193
func instDivU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	bMod32 := uint32(reg[rB])
	aMod32 := uint32(reg[rA])

	if bMod32 == 0 {
		reg[rD] = ^uint64(0) // 2^64 - 1
	} else {
		reg[rD], err = SignExtend(4, uint64(aMod32/bMod32))
		if err != nil {
			return err, pc, Gas(0), reg, mem
		}
	}

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 194
func instDivS32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	a, err := UnsignedToSigned(uint64(uint32(reg[rA])), 4)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	b, err := UnsignedToSigned(uint64(uint32(reg[rB])), 4)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	if b == 0 {
		reg[rD] = ^uint64(0) // 2^64 - 1
	} else if a == int64(-1<<31) && b == -1 {
		reg[rD] = uint64(a)
	} else {
		reg[rD] = uint64(a / b)
	}

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 195
func instRemU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	bMod32 := uint32(reg[rB])
	aMod32 := uint32(reg[rA])

	if bMod32 == 0 {
		reg[rD], err = SignExtend(4, uint64(aMod32))
	} else {
		reg[rD], err = SignExtend(4, uint64(aMod32%bMod32))
	}
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 196
func instRemS32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	a, err := UnsignedToSigned(uint64(uint32(reg[rA])), 4)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	b, err := UnsignedToSigned(uint64(uint32(reg[rB])), 4)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	if a == int64(-1<<31) && b == -1 {
		reg[rD] = 0
	} else {
		reg[rD], err = SignedToUnsigned(smod(a, b), 8)
		if err != nil {
			return err, pc, Gas(0), reg, mem
		}
	}

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 197
func instShloL32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	shift := reg[rB] % 32
	reg[rD], err = SignExtend(4, uint64(uint32(reg[rA]<<shift)))
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 198
func instShloR32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	modA := uint32(reg[rA])
	shift := reg[rB] % 32
	reg[rD], err = SignExtend(4, uint64(modA>>shift))
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 199
func instSharR32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask []bool) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	signedA, err := UnsignedToSigned(uint64(uint32(reg[rA])), 4)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	shift := reg[rB] % 32
	reg[rD], err = SignedToUnsigned(signedA/(1<<shift), 8)

	if err != nil {
		return err, pc, Gas(0), reg, mem
	}

	// TODO: Why panic?
	return PVMExitTuple(PANIC, nil), pc, gasDelta, reg, mem
}

// opcode 200
func instAdd64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = reg[rA] + reg[rB]

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 201
func instSub64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = reg[rA] + (^reg[rB] + 1)

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 202
func instMul64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = reg[rA] * reg[rB]

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 203
func instDivU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	if reg[rB] == 0 {
		reg[rD] = ^uint64(0) // 2^64 - 1
	} else {
		reg[rD] = reg[rA] / reg[rB]
	}

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 204
func instDivS64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	if reg[rB] == 0 {
		reg[rD] = ^uint64(0) // 2^64 - 1
	} else if int64(reg[rA]) == -(1<<63) && int64(reg[rB]) == -1 {
		reg[rD] = reg[rA]
	} else {
		reg[rD] = uint64((int64(reg[rA]) / int64(reg[rB])))
	}

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 205
func instRemU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	if reg[rB] == 0 {
		reg[rD] = reg[rA]
	} else {
		reg[rD] = reg[rA] % reg[rB]
	}

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 206
func instRemS64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	if int64(reg[rA]) == -(1<<63) && int64(reg[rB]) == -1 {
		reg[rD] = 0
	} else {
		reg[rD] = uint64(smod(int64(reg[rA]), int64(reg[rB])))
	}

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 207
func instShloL64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = reg[rA] << (reg[rB] % 64)

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 208
func instShloR64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = reg[rA] >> (reg[rB] % 64)

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 209
func instSharR64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = uint64(int64(reg[rA]) >> (reg[rB] % 64))

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 210
func instAnd(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = reg[rA] & reg[rB]

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 211
func instXor(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = reg[rA] ^ reg[rB]

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}

// opcode 212
func instOr(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Gas, Registers, Memory) {
	gasDelta := Gas(2)
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		return err, pc, Gas(0), reg, mem
	}
	// mutation
	reg[rD] = reg[rA] | reg[rB]

	// TODO: Why panic?
	return PVMExitTuple(CONTINUE, nil), pc, gasDelta, reg, mem
}
