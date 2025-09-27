package PVM

import (
	"errors"
	"fmt"
	"math/bits"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"golang.org/x/exp/constraints"
)

// Instruction tables

// result of "ζı" should be a opcode
type opcode uint64

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

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// input: instructionCode, programCounter, skipLength, registers, memory
var execInstructions = [231]func([]byte, ProgramCounter, ProgramCounter, Registers, Memory, JumpTable, Bitmask) (error, ProgramCounter, Registers, Memory){
	// A.5.1 Instructiopns without Arguments
	0: instTrap,
	1: instFallthrough,
	// A.5.2 Instructions with Arguments of One Immediate
	10: instEcalli,
	// A.5.3 Instructions with Arguments of One Register & One Extended With Immediate
	20: instLoadImm64, // passed testvector
	// A.5.4 Instructions with Arguments of Two Immediates
	30: instStoreImmU8,
	31: instStoreImmU16,
	32: instStoreImmU32,
	33: instStoreImmU64,
	// A.5.5 Instructions with Arguments of One Offset
	40: instJump,
	// A.5.6 Instructions with Arguments of One Register & One Immediate
	50: instJumpInd,
	51: instLoadImm,
	52: instLoadU8,
	53: instLoadI8,
	54: instLoadU16,
	55: instLoadI16,
	56: instLoadU32,
	57: instLoadI32,
	58: instLoadU64,
	59: instStoreU8,
	60: instStoreU16,
	61: instStoreU32,
	62: instStoreU64,
	// A.5.7 Instructions with Arguments of One Register & Two Immediates
	70: instStoreImmIndU8,
	71: instStoreImmIndU16,
	72: instStoreImmIndU32,
	73: instStoreImmIndU64,
	// A.5.8 Instructions without Arguments of One Register, One Immediate and One Offset
	80: instImmediateBranch,
	81: instImmediateBranch,
	82: instImmediateBranch,
	83: instImmediateBranch,
	84: instImmediateBranch,
	85: instImmediateBranch,
	86: instImmediateBranch,
	87: instImmediateBranch,
	88: instImmediateBranch,
	89: instImmediateBranch,
	90: instImmediateBranch,
	// A.5.9 Instructions with arguments of Two Registers
	100: instMoveReg, // passed testvector
	101: instSbrk,
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
	120: instStoreIndU8,
	121: instStoreIndU16,
	122: instStoreIndU32,
	123: instStoreIndU64,
	124: instLoadIndU8,
	125: instLoadIndI8,
	126: instLoadIndU16,
	127: instLoadIndI16,
	128: instLoadIndU32,
	129: instLoadIndI32,
	130: instLoadIndU64,
	131: instAddImm32,
	132: instAndImm,
	133: instXORImm,
	134: instORImm,
	135: instMulImm32,
	136: instSetLtUImm,
	137: instSetLtSImm,
	138: instShloLImm32,
	139: instShloRImm32,
	140: instSharRImm32,
	141: instNegAddImm32,
	142: instSetGtUImm,
	143: instSetGtSImm,
	144: instShloLImmAlt32,
	145: instShloRImmAlt32,
	146: instSharRImmAlt32,
	147: instCmovIzImm,
	148: instCmovNzImm,
	149: instAddImm64,
	150: instMulImm64,
	151: instShloLImm64,
	152: instShloRImm64,
	153: instSharRImm64,
	154: instNegAddImm64,
	155: instShloLImmAlt64,
	156: instShloRImmAlt64,
	157: instSharRImmAlt64,
	158: instRotR64Imm,
	159: instRotR64ImmAlt,
	160: instRotR32Imm,
	161: instRotR32ImmAlt,
	170: instBranch,
	171: instBranch,
	172: instBranch,
	173: instBranch,
	174: instBranch,
	175: instBranch,
	// A.5.12 Instructions  with Arguments of Two Registers and Two Immediates
	180: instLoadImmJumpInd,
	// A.5.13 Instructions with Arguments of Three Registers
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
	213: instMulUpperSS,
	214: instMulUpperUU,
	215: instMulUpperSU,
	216: instSetLtU,
	217: instSetLtS,
	218: instCmovIz,
	219: instCmovNz,
	220: instRotL64,
	221: instRotL32,
	222: instRotR64,
	223: instRotR32,
	224: instAndInv,
	225: instOrInv,
	226: instXnor,
	227: instMax,
	228: instMaxU,
	229: instMin,
	230: instMinU,
	// register more instructions here
}

// opcode 0
func instTrap(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	logger.Debugf("[%d]: pc: %d, %s", instrCount, pc, zeta[opcode(instructionCode[pc])])
	return PVMExitTuple(PANIC, nil), pc, reg, mem
}

// opcode 1
func instFallthrough(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	logger.Debugf("[%d]: pc: %d, %s", instrCount, pc, zeta[opcode(instructionCode[pc])])
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 10
func instEcalli(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	lX := min(4, int(skipLength))

	// zeta_{iota+1,...,lX}
	instLength := instructionCode[pc+1 : pc+ProgramCounter(lX)+1]
	x, err := utils.DeserializeFixedLength(instLength, types.U64(lX))
	if err != nil {
		logger.Errorf("instEcalli deserialization error: %v", err)
		return err, pc, reg, mem
	}
	nuX, err := SignExtend(lX, uint64(x))
	if err != nil {
		logger.Errorf("instEcalli signExtend error: %v", err)
		return err, pc, reg, mem
	}

	nuX = uint64(uint32(nuX))

	logger.Debugf("[%d]: pc: %d, %s %d", instrCount, pc, zeta[opcode(instructionCode[pc])], nuX)
	return PVMExitTuple(HOST_CALL, nuX), pc, reg, mem
}

// opcode 20
func instLoadImm64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA := min(12, (int(instructionCode[pc+1]) % 16))
	// zeta_{iota+2,...,+8}
	instLength := instructionCode[pc+2 : pc+10]
	nuX, err := utils.DeserializeFixedLength(instLength, types.U64(8))
	if err != nil {
		logger.Errorf("insLoadImm64 deserialization raise error: %v", err)
	}
	reg[rA] = uint64(nuX)
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], formatInt(uint64(nuX)))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 30
func instStoreImmU8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	vx, vy, err := decodeTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreImmU8 decodeTwoImmediates error: %v", err)
		return err, pc, reg, mem
	}
	offset := 1
	vy = uint64(uint8(vy))
	exitReason := storeIntoMemory(mem, offset, uint32(vx), vy)
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vx), formatInt(vy))
	} else {
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vx))
	}
	return exitReason, pc, reg, mem
}

// opcode 31
func instStoreImmU16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	vx, vy, err := decodeTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreImmU16 decodeTwoImmediates error: %v", err)
		return err, pc, reg, mem
	}
	offset := 2
	vy = uint64(uint16(vy))
	exitReason := storeIntoMemory(mem, offset, uint32(vx), vy)
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vx), formatInt(vy))
	} else {
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vx))
	}
	return exitReason, pc, reg, mem
}

// opcode 32
func instStoreImmU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	vx, vy, err := decodeTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreImmU32 decodeTwoImmediates error: %v", err)
		return err, pc, reg, mem
	}
	offset := 4
	vy = uint64(uint32(vy))
	exitReason := storeIntoMemory(mem, offset, uint32(vx), vy)
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vx), formatInt(vy))
	} else {
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vx))
	}
	return exitReason, pc, reg, mem
}

// opcode 33
func instStoreImmU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	vx, vy, err := decodeTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreImmU64 decodeTwoImmediates error: %v", err)
		return err, pc, reg, mem
	}
	offset := 8
	exitReason := storeIntoMemory(mem, offset, uint32(vx), vy)
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vx), formatInt(vy))
	} else {
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vx))
	}
	return exitReason, pc, reg, mem
}

// opcode 40
func instJump(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	vX, err := decodeOneOffset(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instJump decodeOneOffset error: %v", err)
		return err, pc, reg, mem
	}

	reason, newPC := branch(pc, vX, true, bitmask, instructionCode)

	if reason != CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s %d panic", instrCount, pc, zeta[opcode(instructionCode[pc])], newPC)
		return PVMExitTuple(reason, nil), pc, reg, mem
	}
	logger.Debugf("[%d]: pc: %d, %s %d", instrCount, pc, zeta[opcode(instructionCode[pc])], newPC)
	return PVMExitTuple(reason, nil), newPC, reg, mem
}

// opcode 50
func instJumpInd(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instJumpInd decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	dest := uint32(reg[rA] + vX)
	reason, newPC := djump(pc, dest, jumpTable, bitmask)
	if reason != CONTINUE { // newPc = dest
		logger.Debugf("[%d]: pc: %d, %s %d panic, %s = %s, vX = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			newPC, RegName[rA], formatInt(reg[rA]), formatInt(vX))
		return PVMExitTuple(reason, nil), pc, reg, mem
	}
	logger.Debugf("[%d]: pc: %d, %s %d, %s = %s, vX = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		newPC, RegName[rA], formatInt(reg[rA]), formatInt(vX))
	return PVMExitTuple(reason, nil), newPC, reg, mem
}

// opcode 51
func instLoadImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadImm decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}
	reg[rA] = uint64(vX)
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 52
func instLoadU8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadU8 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}
	offset := uint32(1)
	memVal, exitReason := loadFromMemory(mem, offset, uint32(uint8(vX)))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
		} else {
			logger.Errorf("instLoadU8 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	reg[rA] = memVal
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(memVal))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 53
func instLoadI8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadI8 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 1
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(int8(vX)))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
		} else {
			logger.Errorf("instLoadI8 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}
	extend, err := SignExtend(offset, memVal)
	if err != nil {
		logger.Errorf("instLoadI8 SignExtend error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = extend
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(memVal))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 54
func instLoadU16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadU16 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 2
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(uint16(vX)))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
		} else {
			logger.Errorf("instLoadU16 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}
	reg[rA] = memVal
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(memVal))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 55
func instLoadI16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadI16 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 2
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(int16(vX)))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
		} else {
			logger.Errorf("instLoadI16 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}
	extend, err := SignExtend(offset, memVal)
	if err != nil {
		logger.Errorf("instLoadI16 signExtend error: %v", err)
		return exitReason, pc, reg, mem
	}
	reg[rA] = extend
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(extend))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 56
func instLoadU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadU32 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 4
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
		} else {
			logger.Errorf("instLoadU32 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	reg[rA] = memVal
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(memVal))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 57
func instLoadI32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadI32 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 4
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
		} else {
			logger.Errorf("instLoadI32 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	extend, err := SignExtend(offset, memVal)
	if err != nil {
		logger.Errorf("instLoadI32 signExtend error: %v", err)
	}
	reg[rA] = extend
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(extend))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 58
func instLoadU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadU64 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 8
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
		} else {
			logger.Errorf("instLoadU64 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	reg[rA] = memVal
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(memVal))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 59
func instStoreU8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreU8 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 1
	exitReason := storeIntoMemory(mem, offset, uint32(vX), uint64(uint8(reg[rA])))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ 0x%x ] = %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX), RegName[rA], formatInt(uint64(uint8(reg[rA]))))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
	}

	return exitReason, pc, reg, mem
}

// opcode 60
func instStoreU16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreU16 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 2
	exitReason := storeIntoMemory(mem, offset, uint32(vX), uint64(uint16(reg[rA])))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ 0x%x ] = %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX), RegName[rA], formatInt(uint64(uint16(reg[rA]))))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
	}
	return exitReason, pc, reg, mem
}

// opcode 61
func instStoreU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreU32 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 4
	exitReason := storeIntoMemory(mem, offset, uint32(vX), uint64(uint32(reg[rA])))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ 0x%x ] = %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX), RegName[rA], formatInt(uint64(uint32(reg[rA]))))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
	}
	return exitReason, pc, reg, mem
}

// opcode 62
func instStoreU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreU64 decodeOneRegisterAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 8
	exitReason := storeIntoMemory(mem, offset, uint32(vX), reg[rA])
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ 0x%x ] = %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX), RegName[rA], formatInt(reg[rA]))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], uint32(vX))
	}
	return exitReason, pc, reg, mem
}

// opcode 70
func instStoreImmIndU8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, vY, err := decodeOneRegisterAndTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreImmIndU8 decodeOneRegisterAndTwoImmediates error: %v", err)
		return err, pc, reg, mem
	}

	offset := 1
	exitReason := storeIntoMemory(mem, offset, uint32(reg[rA]+vX), uint64(uint8(vY)))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ %s+%s = 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], formatInt(uint32(vX)), uint32(reg[rA]+vX), formatInt(uint64(uint8(vY))))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], formatInt(uint32(vX)), uint32(reg[rA]+vX))
	}
	return exitReason, pc, reg, mem
}

// opcode 71
func instStoreImmIndU16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, vY, err := decodeOneRegisterAndTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreImmIndU16 decodeOneRegisterAndTwoImmediates error: %v", err)
		return err, pc, reg, mem
	}

	offset := 2
	exitReason := storeIntoMemory(mem, offset, uint32(reg[rA]+vX), uint64(uint16(vY)))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ %s+%s = 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], formatInt(uint32(vX)), uint32(reg[rA]+vX), formatInt(uint64(uint16(vY))))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], formatInt(uint32(vX)), uint32(reg[rA]+vX))
	}
	return exitReason, pc, reg, mem
}

// opcode 72
func instStoreImmIndU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, vY, err := decodeOneRegisterAndTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreImmIndU32 decodeOneRegisterAndTwoImmediates error: %v", err)
		return err, pc, reg, mem
	}

	offset := 4
	exitReason := storeIntoMemory(mem, offset, uint32(reg[rA]+vX), uint64(uint32(vY)))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ %s+%s = 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], formatInt(uint32(vX)), uint32(reg[rA]+vX), formatInt(uint64(uint32(vY))))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s= 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], formatInt(uint32(vX)), uint32(reg[rA]+vX))
	}
	return exitReason, pc, reg, mem
}

// opcode 73
func instStoreImmIndU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, vY, err := decodeOneRegisterAndTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreImmIndU64 decodeOneRegisterAndTwoImmediates error: %v", err)
		return err, pc, reg, mem
	}

	offset := 8
	exitReason := storeIntoMemory(mem, offset, uint32(reg[rA]+vX), vY)
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ %s+%s = 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], formatInt(uint32(vX)), uint32(reg[rA]+vX), formatInt(vY))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], formatInt(uint32(vX)), uint32(reg[rA]+vX))
	}
	return exitReason, pc, reg, mem
}

// opcode in [80, 90]
func instImmediateBranch(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, vX, vY, err := decodeOneRegisterOneImmediateAndOneOffset(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instImmediateBranch decodeOneRegisterOneImmediateAndOneOffset error: %v", err)
		return err, pc, reg, mem
	}
	var logMsg string
	branchCondition := false

	switch instructionCode[pc] {
	case 80:
		reg[rA] = vX
		branchCondition = true
		logMsg = fmt.Sprintf("branch(%d, T), %s = %d, %s = %s", vY, RegName[rA], reg[rA], RegName[rA], formatInt(reg[rA]))
	case 81:
		branchCondition = reg[rA] == vX
		logMsg = fmt.Sprintf("branch(%d, (%s=%s == %s) = %t)", vY, RegName[rA], formatInt(reg[rA]), formatInt(vX), branchCondition)
	case 82:
		branchCondition = reg[rA] != vX
		logMsg = fmt.Sprintf("branch(%d, (%s=%s != %s) = %t)", vY, RegName[rA], formatInt(reg[rA]), formatInt(vX), branchCondition)
	case 83:
		branchCondition = reg[rA] < vX
		logMsg = fmt.Sprintf("branch(%d, (%s=%s < %s) = %t)", vY, RegName[rA], formatInt(reg[rA]), formatInt(vX), branchCondition)
	case 84:
		branchCondition = reg[rA] <= vX
		logMsg = fmt.Sprintf("branch(%d, (%s=%s <= %s) = %t)", vY, RegName[rA], formatInt(reg[rA]), formatInt(vX), branchCondition)
	case 85:
		branchCondition = reg[rA] >= vX
		logMsg = fmt.Sprintf("branch(%d, (%s=%s >= %s) = %t)", vY, RegName[rA], formatInt(reg[rA]), formatInt(vX), branchCondition)
	case 86:
		branchCondition = reg[rA] > vX
		logMsg = fmt.Sprintf("branch(%d, (%s=%s > %s) = %t)", vY, RegName[rA], formatInt(reg[rA]), formatInt(vX), branchCondition)
	case 87:
		branchCondition = int64(reg[rA]) < int64(vX)
		logMsg = fmt.Sprintf("branch(%d, (%s=%s < %s) = %t)", vY, RegName[rA], formatInt(int64(reg[rA])), formatInt(int64(vX)), branchCondition)
	case 88:
		branchCondition = int64(reg[rA]) <= int64(vX)
		logMsg = fmt.Sprintf("branch(%d, (%s=%s <= %s) = %t)", vY, RegName[rA], formatInt(int64(reg[rA])), formatInt(int64(vX)), branchCondition)
	case 89:
		branchCondition = int64(reg[rA]) >= int64(vX)
		logMsg = fmt.Sprintf("branch(%d, (%s=%s >= %s) = %t)", vY, RegName[rA], formatInt(int64(reg[rA])), formatInt(int64(vX)), branchCondition)
	case 90:
		branchCondition = int64(reg[rA]) > int64(vX)
		logMsg = fmt.Sprintf("branch(%d, (%s=%s > %s) = %t)", vY, RegName[rA], formatInt(int64(reg[rA])), formatInt(int64(vX)), branchCondition)
	default:
		logger.Fatalf("instImmediateBranch is supposed to be called with opcode in [80, 90]")
	}

	reason, newPC := branch(pc, vY, branchCondition, bitmask, instructionCode)
	if reason != CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s panic", instrCount, pc, zeta[opcode(instructionCode[pc])])
		return PVMExitTuple(reason, nil), pc, reg, mem
	}

	logger.Debugf("[%d]: pc: %d, %s %s", instrCount, pc, zeta[opcode(instructionCode[pc])], logMsg)
	return PVMExitTuple(reason, nil), newPC, reg, mem
}

// opcode 100
func instMoveReg(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMoveReg decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA]
	logger.Debugf("[%d]: pc: %d, %s, %s = %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], RegName[rA], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 101
func instSbrk(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMoveReg decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}

	// this reivision is according to jam-test-vector traces: Note on SBRK
	if reg[rA] == 0 {
		reg[rD] = mem.heapPointer
		logger.Debugf("[%d]: pc: %d, %s, %s = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rD], formatInt(reg[rD]))
		return PVMExitTuple(CONTINUE, nil), pc, reg, mem
	}

	nextPageBoundary := P(int(mem.heapPointer))
	newHeapPointer := mem.heapPointer + reg[rA]

	if newHeapPointer > uint64(nextPageBoundary) {
		finalBoundary := P(int(newHeapPointer))

		// allocated memeory access
		allocateMemorySegment(&mem, uint32(mem.heapPointer), uint32(finalBoundary), nil, MemoryReadWrite)
	}

	mem.heapPointer = newHeapPointer
	reg[rD] = newHeapPointer

	logger.Debugf("[%d]: pc: %d, %s, %s = %s + %s = %s + %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rD], RegName[rA], formatInt(reg[rD]), formatInt(reg[rA]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 102
func instCountSetBits64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instCountSetBits64 decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits(regA, 8)
	if err != nil {
		logger.Errorf("insCountSetBits64 UnsignedToBits error: %v", err)
	}
	var sum uint64 = 0
	for i := 0; i < 64; i++ {
		if bitslice[i] {
			sum++
		}
	}
	reg[rD] = sum
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 103
func instCountSetBits32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instCountSetBits32 decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits((regA % (1 << 32)), 4)
	if err != nil {
		logger.Errorf("instCountSetBits32 UnsignedToBits error: %v", err)
	}
	var sum uint64 = 0
	for i := 0; i < 32; i++ {
		if bitslice[i] {
			sum++
		}
	}
	reg[rD] = sum
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 104
func instLeadingZeroBits64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instLeadingZeroBits64 decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits(regA, 8)
	if err != nil {
		logger.Errorf("instLeadingZeroBits64 UnsignedToBits error: %v", err)
	}
	var n uint64 = 0
	for i := 0; i < 64; i++ {
		if bitslice[i] {
			break
		}
		n++
	}
	reg[rD] = n
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 105
func instLeadingZeroBits32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instLeadingZeroBits32 decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits((regA % (1 << 32)), 4)
	if err != nil {
		logger.Errorf("instLeadingZeroBits32 UnsignedToBits error: %v", err)
	}
	var n uint64 = 0
	for i := 0; i < 32; i++ {
		if bitslice[i] {
			break
		}
		n++
	}
	reg[rD] = n
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 106
func instTrailZeroBits64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instTrailZeroBits64 decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits(regA, 8)
	if err != nil {
		logger.Errorf("instTrailZeroBits64 UnsignedToBits error: %v", err)
	}
	var n uint64 = 0
	for i := 63; i >= 0; i-- {
		if bitslice[i] {
			break
		}
		n++
	}
	reg[rD] = n
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 107
func instTrailZeroBits32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instTrailZeroBits32 decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := reg[rA]
	bitslice, err := UnsignedToBits((regA % (1 << 32)), 4)
	if err != nil {
		logger.Errorf("instTrailZeroBits32 UnsignedToBits error: %v", err)
	}
	var n uint64 = 0
	for i := 31; i >= 0; i-- {
		if bitslice[i] {
			break
		}
		n++
	}
	reg[rD] = n
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 108
func instSignExtend8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instSignExtend8 decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := reg[rA]
	signedInt := int8(regA)
	unsignedInt := uint64(signedInt)

	reg[rD] = unsignedInt
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 109
func instSignExtend16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instSignExtend16 decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := reg[rA]
	signedInt := int16(regA)
	unsignedInt := uint64(signedInt)

	reg[rD] = unsignedInt
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 110
func instZeroExtend16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instZeroExtend16 decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := reg[rA]
	reg[rD] = regA % (1 << 16)
	logger.Debugf("[%d]: pc: %d, %s , %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 111
func instReverseBytes(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rD, rA, err := decodeTwoRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instReverseBytes decodeTwoRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	regA := types.U64(reg[rA])
	bytes := utils.SerializeFixedLength(regA, types.U64(8))
	var reversedBytes uint64 = 0
	for i := uint8(0); i < 8; i++ {
		reversedBytes = (reversedBytes << 8) | uint64(bytes[i])
	}
	reg[rD] = reversedBytes
	logger.Debugf("[%d]: pc: %d, %s , %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rD], formatInt(reversedBytes))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 120
func instStoreIndU8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreIndU8 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 1
	exitReason := storeIntoMemory(mem, offset, uint32(reg[rB]+vX), uint64(uint8(reg[rA])))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ %s+%s = 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(reg[rB]+vX), formatInt(uint64(uint8(reg[rA]))))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+0x%x = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rB], uint32(vX), uint32(reg[rB]+vX))
	}
	return exitReason, pc, reg, mem
}

// opcode 121
func instStoreIndU16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreIndU16 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 2
	exitReason := storeIntoMemory(mem, offset, uint32(reg[rB]+vX), uint64(uint16(reg[rA])))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, mem[ %s+%s = 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rB], formatInt(uint32(vX)), formatInt(uint32(reg[rB]+vX)), formatInt(int64(uint16(reg[rA]))))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rB], formatInt(uint32(vX)), formatInt(uint32(reg[rB]+vX)))
	}
	return exitReason, pc, reg, mem
}

// opcode 122
func instStoreIndU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreIndU32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 4
	exitReason := storeIntoMemory(mem, offset, uint32(reg[rB]+vX), uint64(uint32(reg[rA])))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s , mem[ %s+%s = 0x%x ] = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(reg[rB]+vX), formatInt(uint64(uint32(reg[rA]))))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s , page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(reg[rB]+vX))
	}
	return exitReason, pc, reg, mem
}

// opcode 123
func instStoreIndU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instStoreIndU64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 8
	exitReason := storeIntoMemory(mem, offset, uint32(reg[rB]+vX), uint64(reg[rA]))
	if exitReason.(*PVMExitReason).Reason == CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s , mem[ %s+%s = 0x%x...+%d ] = %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(reg[rB]+vX), offset, RegName[rA], formatInt(uint64(reg[rA])))
	} else { // page fault error
		logger.Debugf("[%d]: pc: %d, %s , page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(reg[rB]+vX))
	}

	return exitReason, pc, reg, mem
}

// opcode 124
func instLoadIndU8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadIndU8 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 1
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(reg[rB]+vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])],
				RegName[rB], formatInt(vX), uint32(reg[rB]+vX))
		} else {
			logger.Errorf("instLoadIndU8 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}
	reg[rA] = memVal
	logger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA],
		RegName[rB], formatInt(vX), uint32(reg[rB]+vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 125
func instLoadIndI8(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadIndI8 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 1
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(reg[rB]+vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rB], formatInt(vX), uint32(reg[rB]+vX))
		} else {
			logger.Errorf("instLoadIndI8 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	reg[rA] = uint64(int8(memVal))
	logger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], RegName[rB], formatInt(vX), uint32(reg[rB]+vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 126
func instLoadIndU16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadIndU16 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 2
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(reg[rB]+vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rB], formatInt(vX), uint32(reg[rB]+vX))
		} else {
			logger.Errorf("instLoadIndU16 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	reg[rA] = memVal
	logger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], RegName[rB], formatInt(vX), uint32(reg[rB]+vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 127
func instLoadIndI16(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadIndI16 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 2
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(reg[rB]+vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rB], formatInt(vX), uint32(reg[rB]+vX))
		} else {
			logger.Errorf("instLoadIndI16 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	reg[rA] = uint64(int16(memVal))
	logger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", pc, zeta[opcode(instructionCode[pc])], RegName[rA], RegName[rB], formatInt(vX), uint32(reg[rB]+vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 128
func instLoadIndU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadIndU32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 4
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(reg[rB]+vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rB], formatInt(vX), uint32(reg[rB]+vX))
		} else {
			logger.Errorf("instLoadIndU32 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	reg[rA] = memVal
	logger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])], RegName[rA], RegName[rB], formatInt(vX), uint32(reg[rB]+vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 129
func instLoadIndI32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadIndI32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 4
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(reg[rB]+vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])],
				RegName[rB], formatInt(vX), uint32(reg[rB]+vX))
		} else {
			logger.Errorf("instLoadIndI32 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	reg[rA] = uint64(int32(memVal))
	logger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), uint32(reg[rB]+vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 130
func instLoadIndU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadIndU64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	offset := 8
	memVal, exitReason := loadFromMemory(mem, uint32(offset), uint32(reg[rB]+vX))
	if exitReason != nil {
		var pvmExit *PVMExitReason
		if errors.As(exitReason, &pvmExit) {
			logger.Debugf("[%d]: pc: %d, %s page fault error at mem[ %s+%s = 0x%x ]", instrCount, pc, zeta[opcode(instructionCode[pc])],
				RegName[rB], formatInt(vX), uint32(reg[rB]+vX))
		} else {
			logger.Errorf("instLoadIndU64 loadFromMemory error: %v", err)
		}
		return exitReason, pc, reg, mem
	}

	reg[rA] = memVal
	logger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), uint32(reg[rB]+vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 131
func instAddImm32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instAddImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	val, err := SignExtend(4, uint64(uint32(reg[rB]+vX)))
	if err != nil {
		logger.Errorf("instAddImm32 SignExtend error: %v", err)
	}
	reg[rA] = val
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 132
func instAndImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instAndImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = reg[rB] & vX
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s & %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 133
func instXORImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instXORImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = reg[rB] ^ vX
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s ^ %s) = %s", pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 134
func instORImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instORImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = reg[rB] | vX
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s | %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], vX, reg[rA])
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 135
func instMulImm32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instMulImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	val, err := SignExtend(4, uint64(uint32(reg[rB]*vX)))
	if err != nil {
		logger.Errorf("instMulImm32 signExtend error: %v", err)
		return err, pc, reg, mem
	}
	reg[rA] = val
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 136
func instSetLtUImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instSetLtUImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	if reg[rB] < vX {
		reg[rA] = 1
	} else {
		reg[rA] = 0
	}
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s < %s) = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 137
func instSetLtSImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instSetLtSImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	if int64(reg[rB]) < int64(vX) {
		reg[rA] = 1
	} else {
		reg[rA] = 0
	}
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s < %s) = (%s < %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(int(reg[rB])), formatInt(int64(vX)), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 138
func instShloLImm32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instShloLImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	vX = vX & 31 // % 32
	imm, err := SignExtend(4, uint64(uint32(reg[rB]<<vX)))
	if err != nil {
		logger.Errorf("instShloLImm32 SignExtend error: %v", err)
		return err, pc, reg, mem
	}
	reg[rA] = imm
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rB]), formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 139
func instShloRImm32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instShloRImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	vX = vX & 31 // % 32
	imm, err := SignExtend(4, uint64(uint32(reg[rB])>>vX))
	if err != nil {
		logger.Errorf("instShloRImm32 signExtend error: %v", err)
		return PVMExitTuple(PANIC, nil), pc, reg, mem
	}
	reg[rA] = imm
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rB]), formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 140
func instSharRImm32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instSharRImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	vX = vX & 31 // % 32
	reg[rA] = uint64(int32(reg[rB]) >> vX)
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = 0x%x", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rB]), formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 141
func instNegAddImm32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instNegAddImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	imm, err := SignExtend(4, uint64(uint32(vX+(1<<32)-reg[rB])))
	if err != nil {
		logger.Errorf("instNegAddImm32 signExtend: %v", err)
		return err, pc, reg, mem
	}
	reg[rA] = uint64(imm)
	logger.Debugf("[%d]: pc: %d, %s, %s = (0x%x + (1<<32) - %s) = (0x%x + (1<<32) - %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], vX, RegName[rB], vX, formatInt(reg[rB]), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 142
func instSetGtUImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instSetGtUImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	if reg[rB] > vX {
		reg[rA] = 1
	} else {
		reg[rA] = 0
	}
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s > %s) = (%s > %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rB]), formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 143
func instSetGtSImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instSetGtSImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	if int64(reg[rB]) > int64(vX) {
		reg[rA] = 1
	} else {
		reg[rA] = 0
	}
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s > %s) = (0x%x > %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(int64(vX)), formatInt(reg[rB]), formatInt(int64(vX)), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 144
func instShloLImmAlt32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instShloLImmAlt32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	imm, err := SignExtend(4, uint64(uint32(vX<<(reg[rB]&31))))
	if err != nil {
		logger.Errorf("instShloLImmAlt32 signExtend error: %v", err)
		return err, pc, reg, mem
	}
	reg[rA] = imm
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s) ", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(reg[rB]), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 145
func instShloRImmAlt32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instShloRImmAlt32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	imm, err := SignExtend(4, uint64(uint32(vX)>>(reg[rB]&31)))
	if err != nil {
		logger.Errorf("instShloRImmAlt32 signExtend error: %v", err)
		return err, pc, reg, mem
	}
	reg[rA] = imm
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(reg[rB]&31), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 146
func instSharRImmAlt32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instSharRImmAlt32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	imm := uint64(int32(uint32(vX)) >> (reg[rB] & 31))
	reg[rA] = imm
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> 0x%x) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(uint32(vX)), RegName[rB], formatInt(uint32(vX)), reg[rB], formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 147
func instCmovIzImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instCmovIzImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	if reg[rB] == 0 {
		reg[rA] = vX
		logger.Debugf("[%d]: pc: %d, %s, %s = %s (%s == 0)", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rA], formatInt(vX), RegName[rB])
	} else {
		logger.Debugf("[%d]: pc: %d, %s, %s = %s (%s != 0)", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rA], formatInt(reg[rA]), RegName[rB])
	}

	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 148
func instCmovNzImm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instCmovNzImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	if reg[rB] != 0 {
		reg[rA] = vX
		logger.Debugf("[%d]: pc: %d, %s, %s = %s (%s != 0)", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rA], formatInt(vX), RegName[rB])
	} else {
		logger.Debugf("[%d]: pc: %d, %s, %s = %s (%s == 0)", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rA], formatInt(vX), RegName[rB])
	}

	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 149
func instAddImm64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instAddImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = reg[rB] + vX
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s)  = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 150
func instMulImm64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instMulImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = reg[rB] * vX
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = (%s • %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(reg[rB]), formatInt(vX), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 151
func instShloLImm64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instShloLImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	imm, err := SignExtend(8, reg[rB]<<(vX&63))
	if err != nil {
		logger.Errorf("instShloLImm64 signExtend error: %v", err)
		return err, pc, reg, mem
	}
	reg[rA] = imm
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX&63), formatInt(reg[rB]), formatInt(vX&63), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 152
func instShloRImm64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instShloRImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	imm, err := SignExtend(8, reg[rB]>>(vX&63))
	if err != nil {
		logger.Errorf("instShloRImm64 signExtend error: %v", err)
		return err, pc, reg, mem
	}
	reg[rA] = imm
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX&63), formatInt(reg[rB]), formatInt(vX&63), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 153
func instSharRImm64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instSharRImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = uint64(int64(reg[rB]) >> (vX & 63))
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], RegName[rB], formatInt(vX&63), formatInt(reg[rB]), formatInt(vX&63), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 154
func instNegAddImm64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instNegAddImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = vX - reg[rB]
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s + (1<<64) - %s) = (%s + (1<<64) - %s) = %s ", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(reg[rB]), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 155
func instShloLImmAlt64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instShloLImmAlt64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = vX << (reg[rB] & 63)
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(reg[rB]&63), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 156
func instShloRImmAlt64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instShloRImmAlt64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = vX >> (reg[rB] & 63)
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s)", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(reg[rB]&63), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 157
func instSharRImmAlt64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instSharRImmAlt64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	reg[rA] = uint64(int64(vX) >> (reg[rB] & 63))
	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s)", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(int64(vX)), RegName[rB], formatInt(int64(vX)), formatInt(reg[rB]&63), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 158
func instRotR64Imm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instRotR64Imm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	// rotate right
	reg[rA] = bits.RotateLeft64(reg[rB], -int(vX&63))
	// reg[rA] = (reg[rB] >> vX) | (reg[rB] << (64 - vX))
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 159
func instRotR64ImmAlt(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instRotR64ImmAlt decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	// rotate right
	reg[rB] &= 63 // % 64
	reg[rA] = bits.RotateLeft64(vX, -int(reg[rB]&63))
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 160
func instRotR32Imm(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instRotR32Imm decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	// rotate right
	imm := bits.RotateLeft32(uint32(reg[rB]), -int(vX&31))

	val, err := SignExtend(4, uint64(imm))
	if err != nil {
		logger.Errorf("instRotR32Imm signExtend error: %v", err)
		return PVMExitTuple(PANIC, nil), pc, reg, mem
	}
	reg[rA] = val
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 161
func instRotR32ImmAlt(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instRotR32ImmAlt decodeTwoRegistersAndOneImmediate error: %v", err)
		return err, pc, reg, mem
	}

	// rotate right
	imm := bits.RotateLeft32(uint32(vX), -int(reg[rB]&31))

	val, err := SignExtend(4, uint64(imm))
	if err != nil {
		logger.Errorf("instRotR32ImmAlt signExtend error: %v", err)
		return PVMExitTuple(PANIC, nil), pc, reg, mem
	}
	reg[rA] = val
	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rA], formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode in [170, 175]
func instBranch(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, err := decodeTwoRegistersAndOneOffset(instructionCode, pc, skipLength)
	if err != nil {
		return err, pc, reg, mem
	}
	var logMsg string
	branchCondition := false
	// logMsg = fmt.Sprintf("branch(%d, (%s=%s != %s) = %t)", vY, RegName[rA], formatInt(reg[rA]), formatInt(vX), branchCondition)
	switch instructionCode[pc] {
	case 170:
		branchCondition = reg[rA] == reg[rB]
		logMsg = fmt.Sprintf("branch(%d, (%s=%s == %s=%s) = %t)",
			vX, RegName[rA], formatInt(reg[rA]), RegName[rB], formatInt(reg[rB]), branchCondition)
	case 171:
		branchCondition = reg[rA] != reg[rB]
		logMsg = fmt.Sprintf("branch(%d, (%s=%s != %s=%s) = %t)",
			vX, RegName[rA], formatInt(reg[rA]), RegName[rB], formatInt(reg[rB]), branchCondition)
	case 172:
		branchCondition = reg[rA] < reg[rB]
		logMsg = fmt.Sprintf("branch(%d, (%s=%s < %s=%s) = %t)",
			vX, RegName[rA], formatInt(reg[rA]), RegName[rB], formatInt(reg[rB]), branchCondition)
	case 173:
		branchCondition = int64(reg[rA]) < int64(reg[rB])
		logMsg = fmt.Sprintf("branch(%d, (%s=%s < %s=%s) = %t)",
			vX, RegName[rA], formatInt(int64(reg[rA])), RegName[rB], formatInt(int64(reg[rB])), branchCondition)
	case 174:
		branchCondition = reg[rA] >= reg[rB]
		logMsg = fmt.Sprintf("branch(%d, (%s=%s >= %s=%s) = %t)",
			vX, RegName[rA], formatInt(reg[rA]), RegName[rB], formatInt(reg[rB]), branchCondition)
	case 175:
		branchCondition = int64(reg[rA]) >= int64(reg[rB])
		logMsg = fmt.Sprintf("branch(%d, (%s= 0x%x >= %s=%s) = %t)",
			vX, RegName[rA], formatInt(int64(reg[rA])), RegName[rB], formatInt(int64(reg[rB])), branchCondition)
	default:
		logger.Fatalf("instBranch is supposed to be called with opcode in [170, 175]")
	}

	reason, newPC := branch(pc, vX, branchCondition, bitmask, instructionCode)
	if reason != CONTINUE {
		logger.Errorf("[%d]: pc: %d, %s panic", instrCount, pc, zeta[opcode(instructionCode[pc])])
		return PVMExitTuple(reason, nil), pc, reg, mem
	}
	logger.Debugf("[%d]: pc: %d, %s, %s", instrCount, pc, zeta[opcode(instructionCode[pc])], logMsg)
	return PVMExitTuple(reason, nil), newPC, reg, mem
}

// opcode 180
func instLoadImmJumpInd(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, vX, vY, err := decodeTwoRegistersAndTwoImmediates(instructionCode, pc, skipLength)
	if err != nil {
		logger.Errorf("instLoadImmJumpInd decodeTwoRegistersAndTwoImmediates error: %v", err)
		return err, pc, reg, mem
	}
	// per https://github.com/koute/jamtestvectors/blob/master_pvm_initial/pvm/TESTCASES.md#inst_load_imm_and_jump_indirect_invalid_djump_to_zero_different_regs_without_offset_nok
	// the register update should take place even if the jump panics
	dest := uint32(reg[rB] + vY)
	reason, newPC := djump(pc, dest, jumpTable, bitmask)

	reg[rA] = vX
	if reason != CONTINUE {
		logger.Debugf("[%d]: pc: %d, %s, %v", instrCount, pc, zeta[opcode(instructionCode[pc])], reason)
		return PVMExitTuple(reason, nil), pc, reg, mem
	}
	logger.Debugf("[%d]: pc: %d, %s, (%s + %s) = (%s + %s) mod (1<<32) = %s)", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rB], formatInt(vY), formatInt(reg[rB]), formatInt(vY), formatInt(dest))
	return PVMExitTuple(reason, nil), newPC, reg, mem
}

// opcode 190
func instAdd32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instAdd32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD], err = SignExtend(4, uint64(uint32(reg[rA]+reg[rB])))
	if err != nil {
		logger.Errorf("instAdd32 signExtend error: %v", err)
		return err, pc, reg, mem
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s) = u32(%s + %s)  = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 191
func instSub32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instSub32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	// bMod32 := uint32(reg[rB])
	// reg[rD], err = SignExtend(4, uint64(uint32(reg[rA]+^uint64(bMod32)+1)))
	reg[rD], err = SignExtend(4, uint64(uint32(reg[rA])-uint32(reg[rB])))
	if err != nil {
		logger.Errorf("instSub32 signExtend error: %v", err)
		return err, pc, reg, mem
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = u32(%s) - u32(%s) = u32(%s) - u32(%s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(uint32(reg[rA])), formatInt(uint32(reg[rB])), formatInt(reg[rA]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 192
func instMul32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMul32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD], err = SignExtend(4, uint64(uint32(reg[rA]*reg[rB])))
	if err != nil {
		logger.Errorf("instMul32 signExtend error: %v", err)
		return err, pc, reg, mem
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = (%s • %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 193
func instDivU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instDivU32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	bMod32 := uint32(reg[rB])
	aMod32 := uint32(reg[rA])

	if bMod32 == 0 {
		reg[rD] = ^uint64(0) // 2^64 - 1
		logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rD], formatInt(reg[rD]))
	} else {
		reg[rD], err = SignExtend(4, uint64(aMod32/bMod32))
		if err != nil {
			logger.Errorf("instDivU32 signExtend error: %v", err)
			return err, pc, reg, mem
		}
		logger.Debugf("[%d]: pc: %d, %s, %s = (%s / %s) = (%s / %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	}

	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 194
func instDivS32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instDivS32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	a := int64(int32(reg[rA]))
	b := int64(int32(reg[rB]))

	if b == 0 {
		reg[rD] = ^uint64(0) // 2^64 - 1
	} else if a == int64(-1<<31) && b == -1 {
		reg[rD] = uint64(a)
	} else {
		reg[rD] = uint64(a / b)
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = 0x%x", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 195
func instRemU32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instRemU32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	bMod32 := uint32(reg[rB])
	aMod32 := uint32(reg[rA])

	if bMod32 == 0 {
		reg[rD], err = SignExtend(4, uint64(aMod32))
	} else {
		reg[rD], err = SignExtend(4, uint64(aMod32%bMod32))
	}
	if err != nil {
		logger.Errorf("instRemU32 signExtend error: %v", err)
		return err, pc, reg, mem
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 196
func instRemS32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instRemS32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}

	a := int64(int32(reg[rA]))
	b := int64(int32(reg[rB]))

	if a == int64(-1<<31) && b == -1 {
		reg[rD] = 0
	} else {
		reg[rD] = uint64((smod(a, b)))
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 197
func instShloL32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instShloL32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	shift := reg[rB] % 32
	reg[rD], err = SignExtend(4, uint64(uint32(reg[rA]<<shift)))
	if err != nil {
		logger.Errorf("instShloL32 signExtend error: %v", err)
		return err, pc, reg, mem
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 198
func instShloR32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instShloR32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}

	modA := uint32(reg[rA])
	shift := reg[rB] % 32
	reg[rD], err = SignExtend(4, uint64(modA>>shift))
	if err != nil {
		logger.Errorf("instShloR32 signExtend error: %v", err)
		return err, pc, reg, mem
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 199
func instSharR32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instSharR32 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}

	signedA := int32(reg[rA])

	shift := reg[rB] % 32
	reg[rD] = uint64(signedA >> shift)

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(signedA), formatInt(shift), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 200
func instAdd64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instAdd64 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] + reg[rB]

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s) = (%s + %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 201
func instSub64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instSub64 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] + (^reg[rB] + 1)

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s - %s) = (%s - %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 202
func instMul64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMul64 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] * reg[rB]

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = (%s • %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 203
func instDivU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instDivU64 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	if reg[rB] == 0 {
		reg[rD] = ^uint64(0) // 2^64 - 1
	} else {
		reg[rD] = reg[rA] / reg[rB]
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 204
func instDivS64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instDivS64 decodeThreeRegisters error: %v", err)
		return err, pc, reg, mem
	}
	// mutation
	if reg[rB] == 0 {
		reg[rD] = ^uint64(0) // 2^64 - 1
	} else if int64(reg[rA]) == -(1<<63) && int64(reg[rB]) == -1 {
		reg[rD] = reg[rA]
	} else {
		reg[rD] = uint64((int64(reg[rA]) / int64(reg[rB])))
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 205
func instRemU64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instRemU64 decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	if reg[rB] == 0 {
		reg[rD] = reg[rA]
	} else {
		reg[rD] = reg[rA] % reg[rB]
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 206
func instRemS64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instRemS64 decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	if int64(reg[rA]) == -(1<<63) && int64(reg[rB]) == -1 {
		reg[rD] = 0
	} else {
		reg[rD] = uint64(smod(int64(reg[rA]), int64(reg[rB])))
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 207
func instShloL64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instShloL64 decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] << (reg[rB] % 64)

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = %s", pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rA]), formatInt(reg[rB]%64), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 208
func instShloR64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instShloR64 decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] >> (reg[rB] % 64)

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rA]), formatInt(reg[rB]%64), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 209
func instSharR64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instSharR64 decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = uint64(int64(reg[rA]) >> (reg[rB] % 64))

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(int64(reg[rA])), formatInt(reg[rB]%64), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 210
func instAnd(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instAnd decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] & reg[rB]

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s & %s) = (%s & %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 211
func instXor(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instXor decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] ^ reg[rB]

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s ^ %s) = (%s ^ %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 212
func instOr(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instOr decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] | reg[rB]

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s | %s) = (%s | %s) = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 213
func instMulUpperSS(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMulUpperSS decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	signedA := int64(reg[rA])
	signedB := int64(reg[rB])

	hi, _ := bits.Mul64(uint64(abs(signedA)), uint64(abs(signedB)))

	if (signedA < 0) == (signedB < 0) {
		reg[rD] = hi
	} else {
		reg[rD] = uint64(-int64(hi))
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 214
func instMulUpperUU(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMulUpperUU decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	hi, _ := bits.Mul64(reg[rA], reg[rB])
	reg[rD] = hi

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 215
func instMulUpperSU(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMulUpperSU decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	signedA := int64(reg[rA])
	hi, lo := bits.Mul64(uint64(abs(signedA)), reg[rB])

	if signedA < 0 {
		hi = -hi
		if lo != 0 { // 2's complement, borrow 1 from hi
			hi--
		}
		reg[rD] = hi

	} else {
		reg[rD] = hi
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 216
func instSetLtU(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instSetLtU decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	if reg[rA] < reg[rB] {
		reg[rD] = 1
	} else {
		reg[rD] = 0
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = (%s < %s) = (%s < %s) = %t", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(reg[rA]), formatInt(reg[rB]), formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 217
func instSetLtS(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instSetLts decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	if int64(reg[rA]) < int64(reg[rB]) {
		reg[rD] = 1
	} else {
		reg[rD] = 0
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = 0x%x, %s = 0x%x, %s = 0x%x", pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(int64(reg[rA])), formatInt(int64(reg[rB])), formatInt(int64(reg[rD])))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 218
func instCmovIz(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instCmovIz decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	if reg[rB] == 0 {
		reg[rD] = reg[rA]
		logger.Debugf("[%d]: pc: %d, %s, %s = %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rD], RegName[rA], formatInt(reg[rD]))
	} else {
		logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rD], formatInt(reg[rD]))
	}

	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 219
func instCmovNz(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instCmovNz decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	if reg[rB] != 0 {
		reg[rD] = reg[rA]
		logger.Debugf("[%d]: pc: %d, %s, %s = %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rD], RegName[rA], formatInt(reg[rD]))
	} else {
		logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
			RegName[rD], formatInt(reg[rD]))
	}

	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 220
func instRotL64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instRotL64 decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = bits.RotateLeft64(reg[rA], int(reg[rB]%64))

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 221
func instRotL32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instRotL32 decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	rotated := uint64(bits.RotateLeft32(uint32(reg[rA]), int(reg[rB]%32)))
	extend, err := SignExtend(4, rotated)
	if err != nil {
		logger.Errorf("instRoTL32 signExtend error:%v", err)
		return err, pc, reg, mem
	}
	reg[rD] = extend

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 222
func instRotR64(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instRotR64 decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = bits.RotateLeft64(reg[rA], -int(reg[rB]%64))

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 223
func instRotR32(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instRotR32 decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	rotated := uint64(bits.RotateLeft32(uint32(reg[rA]), -int(reg[rB]%32)))
	extend, err := SignExtend(4, rotated)
	if err != nil {
		logger.Errorf("instRotR32 signExtend error:%v", err)
		return err, pc, reg, mem
	}
	reg[rD] = extend

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 224
func instAndInv(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instAndInv decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] & ^reg[rB]

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 225
func instOrInv(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instOrInv decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = reg[rA] | ^reg[rB]

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 226
func instXnor(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instXnor decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}
	// mutation
	reg[rD] = ^(reg[rA] ^ reg[rB])

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 227
func instMax(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMax decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}

	// mutation
	reg[rD] = uint64(max(int64(reg[rA]), int64(reg[rB])))

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 228
func instMaxU(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMaxU decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}

	// mutation
	if reg[rA] > reg[rB] {
		reg[rD] = reg[rA]
	} else {
		reg[rD] = reg[rB]
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 229
func instMin(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf(" decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}

	// mutation
	reg[rD] = uint64(min(int64(reg[rA]), int64(reg[rB])))

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}

// opcode 230
func instMinU(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter, reg Registers, mem Memory, jumpTable JumpTable, bitmask Bitmask) (error, ProgramCounter, Registers, Memory) {
	rA, rB, rD, err := decodeThreeRegisters(instructionCode, pc)
	if err != nil {
		logger.Errorf("instMinU decodeThreeRegisters error:%v", err)
		return err, pc, reg, mem
	}

	// mutation
	if reg[rA] < reg[rB] {
		reg[rD] = reg[rA]
	} else {
		reg[rD] = reg[rB]
	}

	logger.Debugf("[%d]: pc: %d, %s, %s = %s", instrCount, pc, zeta[opcode(instructionCode[pc])],
		RegName[rD], formatInt(reg[rD]))
	return PVMExitTuple(CONTINUE, nil), pc, reg, mem
}
