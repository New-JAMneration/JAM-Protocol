package PVM

import (
	"math/bits"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
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

// input: interpreter, programCounter, skipLength
var execInstructions = [231]func(*Interpreter, ProgramCounter, ProgramCounter) (ExitReason, ProgramCounter){
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
	200: instAdd64,
	201: instSub64,
	202: instMul64,
	203: instDivU64,
	204: instDivS64,
	205: instRemU64,
	206: instRemS64,
	207: instShloL64,
	208: instShloR64,
	209: instSharR64,
	210: instAnd,
	211: instXor,
	212: instOr,
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
func instTrap(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	pvmLogger.Debugf("[%d]: pc: %d, %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])])
	return ExitPanic, pc
}

// opcode 1
func instFallthrough(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	pvmLogger.Debugf("[%d]: pc: %d, %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])])
	return ExitContinue, pc
}

// opcode 10
func instEcalli(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	lX := min(4, int(skipLength))

	// zeta_{iota+1,...,lX}
	instLength := interp.Program.InstructionData[pc+1 : pc+ProgramCounter(lX)+1]
	x, err := utils.DeserializeFixedLength(types.ByteSequence(instLength), types.U64(lX))
	if err != nil {
		pvmLogger.Errorf("instEcalli deserialization error: %v", err)
		return ExitPanic, pc
	}
	nuX, err := SignExtend(uint8(lX), uint64(x))
	if err != nil {
		pvmLogger.Errorf("instEcalli signExtend error: %v", err)
		return ExitPanic, pc
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s %d", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], nuX)
	return ExitHostCall | ExitReason(nuX), pc
}

// opcode 20
func instLoadImm64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA := min(12, (int(interp.Program.InstructionData[pc+1]) % 16))
	// zeta_{iota+2,...,+8}
	instLength := interp.Program.InstructionData[pc+2 : pc+10]
	nuX, err := utils.DeserializeFixedLength(types.ByteSequence(instLength), types.U64(8))
	if err != nil {
		pvmLogger.Errorf("insLoadImm64 deserialization raise error: %v", err)
	}
	interp.Registers[rA] = uint64(nuX)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], formatInt(uint64(nuX)))
	return ExitContinue, pc
}

// opcode 30
func instStoreImmU8(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	vx, vy, err := decodeTwoImmediates(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreImmU8 decodeTwoImmediates error: %v", err)
		return ExitPanic, pc
	}
	offset := 1
	vy = uint64(uint8(vy))
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(vx), vy)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vx), formatInt(vy))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vx))
	}
	return exitReason, pc
}

// opcode 31
func instStoreImmU16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	vx, vy, err := decodeTwoImmediates(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreImmU16 decodeTwoImmediates error: %v", err)
		return ExitPanic, pc
	}
	offset := 2
	vy = uint64(uint16(vy))
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(vx), vy)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vx), formatInt(vy))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vx))
	}
	return exitReason, pc
}

// opcode 32
func instStoreImmU32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	vx, vy, err := decodeTwoImmediates(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreImmU32 decodeTwoImmediates error: %v", err)
		return ExitPanic, pc
	}
	offset := 4
	vy = uint64(uint32(vy))
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(vx), vy)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vx), formatInt(vy))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vx))
	}
	return exitReason, pc
}

// opcode 33
func instStoreImmU64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	vx, vy, err := decodeTwoImmediates(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreImmU64 decodeTwoImmediates error: %v", err)
		return ExitPanic, pc
	}
	offset := 8
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(vx), vy)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vx), formatInt(vy))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vx))
	}
	return exitReason, pc
}

// opcode 40
func instJump(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	vX, err := decodeOneOffset(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instJump decodeOneOffset error: %v", err)
		return ExitPanic, pc
	}

	reason, newPC := branch(pc, vX, true, interp.Program.Bitmasks, interp.Program.InstructionData)

	if reason != ExitContinue {
		pvmLogger.Debugf("[%d]: pc: %d, %s %d panic", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], newPC)
		return reason, pc
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s %d", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], newPC)
	return reason, newPC
}

// opcode 50
func instJumpInd(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instJumpInd decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	dest := uint32(interp.Registers[rA] + vX)
	reason, newPC := djump(pc, dest, interp.Program.JumpTable, interp.Program.Bitmasks)
	switch reason {
	case ExitPanic:
		pvmLogger.Debugf("[%d]: pc: %d, %s %d panic, %s = %s, vX = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			newPC, RegName[rA], formatInt(interp.Registers[rA]), formatInt(vX))
		return reason, pc
	case ExitHalt:
		pvmLogger.Debugf("[%d]: pc: %d, %s %d HALT, %s = %s, vX = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			newPC, RegName[rA], formatInt(interp.Registers[rA]), formatInt(vX))
		return reason, pc
	default: // continue
		pvmLogger.Debugf("[%d]: pc: %d, %s %d, %s = %s, vX = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			newPC, RegName[rA], formatInt(interp.Registers[rA]), formatInt(vX))
		return reason, newPC
	}
}

// opcode 51
func instLoadImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadImm decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}
	interp.Registers[rA] = uint64(vX)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 52
func instLoadU8(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadU8 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}
	offset := uint32(1)
	memVal, exitReason := loadFromMemory(interp.Memory, offset, uint32(vX))
	if exitReason != ExitContinue {
		return ExitContinue, pc
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(memVal))
	return ExitContinue, pc
}

// opcode 53
func instLoadI8(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadI8 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 1
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	if exitReason != ExitContinue {
		return exitReason, pc
	}

	extend, err := SignExtend(uint8(offset), memVal)
	if err != nil {
		pvmLogger.Errorf("instLoadI8 SignExtend error: %v", err)
		return ExitPanic, pc
	}

	interp.Registers[rA] = extend
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(memVal))
	return ExitContinue, pc
}

// opcode 54
func instLoadU16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadU16 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 2
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(vX))
	if exitReason != ExitContinue {
		return ExitContinue, pc
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(memVal))
	return ExitContinue, pc
}

// opcode 55
func instLoadI16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadI16 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}
	offset := 2
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	extend, err := SignExtend(uint8(offset), memVal)
	if err != nil {
		pvmLogger.Errorf("instLoadI16 signExtend error: %v", err)
		return ExitPanic, pc
	}
	interp.Registers[rA] = extend
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(extend))
	return ExitContinue, pc
}

// opcode 56
func instLoadU32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadU32 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 4
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(memVal))
	return ExitContinue, pc
}

// opcode 57
func instLoadI32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadI32 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 4
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	extend, err := SignExtend(uint8(offset), memVal)
	if err != nil {
		pvmLogger.Errorf("instLoadI32 signExtend error: %v", err)
		return ExitPanic, pc
	}
	interp.Registers[rA] = extend
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(extend))
	return ExitContinue, pc
}

// opcode 58
func instLoadU64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadU64 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 8
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(memVal))
	return ExitContinue, pc
}

// opcode 59
func instStoreU8(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreU8 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 1
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(vX), uint64(uint8(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vX), RegName[rA], formatInt(uint64(uint8(interp.Registers[rA]))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vX))
	}

	return exitReason, pc
}

// opcode 60
func instStoreU16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreU16 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 2
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(vX), uint64(uint16(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vX), RegName[rA], formatInt(uint64(uint16(interp.Registers[rA]))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vX))
	}
	return exitReason, pc
}

// opcode 61
func instStoreU32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreU32 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 4
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(vX), uint64(uint32(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vX), RegName[rA], formatInt(uint64(uint32(interp.Registers[rA]))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vX))
	}
	return exitReason, pc
}

// opcode 62
func instStoreU64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, err := decodeOneRegisterAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreU64 decodeOneRegisterAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 8
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(vX), interp.Registers[rA])
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vX), RegName[rA], formatInt(interp.Registers[rA]))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], uint32(vX))
	}
	return exitReason, pc
}

// opcode 70
func instStoreImmIndU8(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, vY, err := decodeOneRegisterAndTwoImmediates(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreImmIndU8 decodeOneRegisterAndTwoImmediates error: %v", err)
		return ExitHalt, pc
	}

	offset := 1
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(interp.Registers[rA]+vX), uint64(uint8(vY)))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX), formatInt(uint64(uint8(vY))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX))
	}
	return exitReason, pc
}

// opcode 71
func instStoreImmIndU16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, vY, err := decodeOneRegisterAndTwoImmediates(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreImmIndU16 decodeOneRegisterAndTwoImmediates error: %v", err)
		return ExitHalt, pc
	}

	offset := 2
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(interp.Registers[rA]+vX), uint64(uint16(vY)))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX), formatInt(uint64(uint16(vY))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX))
	}
	return exitReason, pc
}

// opcode 72
func instStoreImmIndU32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, vY, err := decodeOneRegisterAndTwoImmediates(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreImmIndU32 decodeOneRegisterAndTwoImmediates error: %v", err)
		return ExitHalt, pc
	}

	offset := 4
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(interp.Registers[rA]+vX), uint64(uint32(vY)))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX), formatInt(uint64(uint32(vY))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s= 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX))
	}
	return exitReason, pc
}

// opcode 73
func instStoreImmIndU64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, vY, err := decodeOneRegisterAndTwoImmediates(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreImmIndU64 decodeOneRegisterAndTwoImmediates error: %v", err)
		return ExitHalt, pc
	}

	offset := 8
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(interp.Registers[rA]+vX), vY)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX), formatInt(vY))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX))
	}
	return exitReason, pc
}

// opcode in [80, 90]
func instImmediateBranch(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, vX, vY, err := decodeOneRegisterOneImmediateAndOneOffset(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instImmediateBranch decodeOneRegisterOneImmediateAndOneOffset error: %v", err)
		return ExitHalt, pc
	}
	branchCondition := false

	switch interp.Program.InstructionData[pc] {
	case 80:
		interp.Registers[rA] = vX
		branchCondition = true
	case 81:
		branchCondition = interp.Registers[rA] == vX
	case 82:
		branchCondition = interp.Registers[rA] != vX
	case 83:
		branchCondition = interp.Registers[rA] < vX
	case 84:
		branchCondition = interp.Registers[rA] <= vX
	case 85:
		branchCondition = interp.Registers[rA] >= vX
	case 86:
		branchCondition = interp.Registers[rA] > vX
	case 87:
		branchCondition = int64(interp.Registers[rA]) < int64(vX)
	case 88:
		branchCondition = int64(interp.Registers[rA]) <= int64(vX)
	case 89:
		branchCondition = int64(interp.Registers[rA]) >= int64(vX)
	case 90:
		branchCondition = int64(interp.Registers[rA]) > int64(vX)
	default:
		pvmLogger.Fatalf("instImmediateBranch is supposed to be called with opcode in [80, 90]")
	}

	reason, newPC := branch(pc, vY, branchCondition, interp.Program.Bitmasks, interp.Program.InstructionData)
	if reason != ExitContinue {
		pvmLogger.Debugf("[%d]: pc: %d, %s panic", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])])
		return reason, pc
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s branch(%d, %s=%s, vX=%s) = %t",
		interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], vY, RegName[rA], formatInt(interp.Registers[rA]), formatInt(vX), branchCondition)
	return reason, newPC
}

// opcode 100
func instMoveReg(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMoveReg decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA]
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], RegName[rA], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 101
func instSbrk(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMoveReg decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}

	// this reivision is according to jam-test-vector traces: Note on SBRK
	if interp.Registers[rA] == 0 {
		interp.Registers[rD] = interp.Memory.heapPointer
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rD], formatInt(interp.Registers[rD]))
		return ExitContinue, pc
	}

	nextPageBoundary := P(int(interp.Memory.heapPointer))
	newHeapPointer := interp.Memory.heapPointer + interp.Registers[rA]

	if newHeapPointer > uint64(nextPageBoundary) {
		finalBoundary := P(int(newHeapPointer))

		// allocated memeory access
		allocateMemorySegment(interp.Memory, uint32(interp.Memory.heapPointer), uint32(finalBoundary), nil, MemoryReadWrite)
	}

	interp.Memory.heapPointer = newHeapPointer
	interp.Registers[rD] = newHeapPointer

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s + %s = %s + %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rD], RegName[rA], formatInt(interp.Registers[rD]), formatInt(interp.Registers[rA]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 102
func instCountSetBits64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instCountSetBits64 decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := interp.Registers[rA]
	bitslice, err := UnsignedToBits(regA, 8)
	if err != nil {
		pvmLogger.Errorf("insCountSetBits64 UnsignedToBits error: %v", err)
	}
	var sum uint64 = 0
	for i := 0; i < 64; i++ {
		if bitslice[i] {
			sum++
		}
	}
	interp.Registers[rD] = sum
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 103
func instCountSetBits32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instCountSetBits32 decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := interp.Registers[rA]
	bitslice, err := UnsignedToBits((regA % (1 << 32)), 4)
	if err != nil {
		pvmLogger.Errorf("instCountSetBits32 UnsignedToBits error: %v", err)
	}
	var sum uint64 = 0
	for i := 0; i < 32; i++ {
		if bitslice[i] {
			sum++
		}
	}
	interp.Registers[rD] = sum
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 104
func instLeadingZeroBits64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instLeadingZeroBits64 decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := interp.Registers[rA]
	bitslice, err := UnsignedToBits(regA, 8)
	if err != nil {
		pvmLogger.Errorf("instLeadingZeroBits64 UnsignedToBits error: %v", err)
	}
	var n uint64 = 0
	for i := 0; i < 64; i++ {
		if bitslice[i] {
			break
		}
		n++
	}
	interp.Registers[rD] = n
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 105
func instLeadingZeroBits32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instLeadingZeroBits32 decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := interp.Registers[rA]
	bitslice, err := UnsignedToBits((regA % (1 << 32)), 4)
	if err != nil {
		pvmLogger.Errorf("instLeadingZeroBits32 UnsignedToBits error: %v", err)
	}
	var n uint64 = 0
	for i := 0; i < 32; i++ {
		if bitslice[i] {
			break
		}
		n++
	}
	interp.Registers[rD] = n
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 106
func instTrailZeroBits64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instTrailZeroBits64 decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := interp.Registers[rA]
	bitslice, err := UnsignedToBits(regA, 8)
	if err != nil {
		pvmLogger.Errorf("instTrailZeroBits64 UnsignedToBits error: %v", err)
	}
	var n uint64 = 0
	for i := 63; i >= 0; i-- {
		if bitslice[i] {
			break
		}
		n++
	}
	interp.Registers[rD] = n
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 107
func instTrailZeroBits32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instTrailZeroBits32 decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := interp.Registers[rA]
	bitslice, err := UnsignedToBits((regA % (1 << 32)), 4)
	if err != nil {
		pvmLogger.Errorf("instTrailZeroBits32 UnsignedToBits error: %v", err)
	}
	var n uint64 = 0
	for i := 31; i >= 0; i-- {
		if bitslice[i] {
			break
		}
		n++
	}
	interp.Registers[rD] = n
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 108
func instSignExtend8(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instSignExtend8 decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := interp.Registers[rA]
	signedInt := int8(regA)
	unsignedInt := uint64(signedInt)

	interp.Registers[rD] = unsignedInt
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 109
func instSignExtend16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instSignExtend16 decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := interp.Registers[rA]
	signedInt := int16(regA)
	unsignedInt := uint64(signedInt)

	interp.Registers[rD] = unsignedInt
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 110
func instZeroExtend16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instZeroExtend16 decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := interp.Registers[rA]
	interp.Registers[rD] = regA % (1 << 16)
	pvmLogger.Debugf("[%d]: pc: %d, %s , %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 111
func instReverseBytes(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rD, rA, err := decodeTwoRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instReverseBytes decodeTwoRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	regA := types.U64(interp.Registers[rA])
	bytes := utils.SerializeFixedLength(regA, types.U64(8))
	var reversedBytes uint64 = 0
	for i := uint8(0); i < 8; i++ {
		reversedBytes = (reversedBytes << 8) | uint64(bytes[i])
	}
	interp.Registers[rD] = reversedBytes
	pvmLogger.Debugf("[%d]: pc: %d, %s , %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rD], formatInt(reversedBytes))
	return ExitContinue, pc
}

// opcode 120
func instStoreIndU8(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreIndU8 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 1
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(interp.Registers[rB]+vX), uint64(uint8(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX), formatInt(uint64(uint8(interp.Registers[rA]))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+0x%x = 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rB], uint32(vX), uint32(interp.Registers[rB]+vX))
	}
	return exitReason, pc
}

// opcode 121
func instStoreIndU16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreIndU16 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 2
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(interp.Registers[rB]+vX), uint64(uint16(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rB], formatInt(uint32(vX)), formatInt(uint32(interp.Registers[rB]+vX)), formatInt(int64(uint16(interp.Registers[rA]))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rB], formatInt(uint32(vX)), formatInt(uint32(interp.Registers[rB]+vX)))
	}
	return exitReason, pc
}

// opcode 122
func instStoreIndU32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreIndU32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 4
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(interp.Registers[rB]+vX), uint64(uint32(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s ,[ %s+%s = 0x%x ] = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX), formatInt(uint64(uint32(interp.Registers[rA]))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s , page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX))
	}
	return exitReason, pc
}

// opcode 123
func instStoreIndU64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instStoreIndU64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 8
	exitReason := storeIntoMemory(interp.Memory, offset, uint32(interp.Registers[rB]+vX), uint64(interp.Registers[rA]))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s ,[ %s+%s = 0x%x...+%d ] = %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX), offset, RegName[rA], formatInt(uint64(interp.Registers[rA])))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s , page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX))
	}

	return exitReason, pc
}

// opcode 124
func instLoadIndU8(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadIndU8 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 1
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA],
		RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 125
func instLoadIndI8(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadIndI8 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 1
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	interp.Registers[rA] = uint64(int8(memVal))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 126
func instLoadIndU16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadIndU16 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 2
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 127
func instLoadIndI16(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadIndI16 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 2
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	interp.Registers[rA] = uint64(int16(memVal))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 128
func instLoadIndU32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadIndU32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 4
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 129
func instLoadIndI32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadIndI32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 4
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	interp.Registers[rA] = uint64(int32(memVal))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 130
func instLoadIndU64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadIndU64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	offset := 8
	memVal, exitReason := loadFromMemory(interp.Memory, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, pc
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 131
func instAddImm32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instAddImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	val, err := SignExtend(4, uint64(uint32(interp.Registers[rB]+vX)))
	if err != nil {
		pvmLogger.Errorf("instAddImm32 SignExtend error: %v", err)
	}
	interp.Registers[rA] = val
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 132
func instAndImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instAndImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = interp.Registers[rB] & vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s & %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 133
func instXORImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instXORImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = interp.Registers[rB] ^ vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s ^ %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 134
func instORImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instORImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = interp.Registers[rB] | vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s | %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 135
func instMulImm32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instMulImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	val, err := SignExtend(4, uint64(uint32(interp.Registers[rB]*vX)))
	if err != nil {
		pvmLogger.Errorf("instMulImm32 signExtend error: %v", err)
		return ExitHalt, pc
	}
	interp.Registers[rA] = val
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 136
func instSetLtUImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instSetLtUImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	if interp.Registers[rB] < vX {
		interp.Registers[rA] = 1
	} else {
		interp.Registers[rA] = 0
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s < %s) = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 137
func instSetLtSImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instSetLtSImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	if int64(interp.Registers[rB]) < int64(vX) {
		interp.Registers[rA] = 1
	} else {
		interp.Registers[rA] = 0
	}

	return ExitContinue, pc
}

// opcode 138
func instShloLImm32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instShloLImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	vX = vX & 31 // % 32
	imm, err := SignExtend(4, uint64(uint32(interp.Registers[rB]<<vX)))
	if err != nil {
		pvmLogger.Errorf("instShloLImm32 SignExtend error: %v", err)
		return ExitHalt, pc
	}
	interp.Registers[rA] = imm

	return ExitContinue, pc
}

// opcode 139
func instShloRImm32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instShloRImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	vX = vX & 31 // % 32
	imm, err := SignExtend(4, uint64(uint32(interp.Registers[rB])>>vX))
	if err != nil {
		pvmLogger.Errorf("instShloRImm32 signExtend error: %v", err)
		return ExitPanic, pc
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 140
func instSharRImm32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instSharRImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	vX = vX & 31 // % 32
	interp.Registers[rA] = uint64(int32(interp.Registers[rB]) >> vX)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = 0x%x", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 141
func instNegAddImm32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instNegAddImm32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	imm, err := SignExtend(4, uint64(uint32(vX+(1<<32)-interp.Registers[rB])))
	if err != nil {
		pvmLogger.Errorf("instNegAddImm32 signExtend: %v", err)
		return ExitHalt, pc
	}
	interp.Registers[rA] = uint64(imm)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (0x%x + (1<<32) - %s) = (0x%x + (1<<32) - %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], vX, RegName[rB], vX, formatInt(interp.Registers[rB]), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 142
func instSetGtUImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instSetGtUImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	if interp.Registers[rB] > vX {
		interp.Registers[rA] = 1
	} else {
		interp.Registers[rA] = 0
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s > %s) = (%s > %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 143
func instSetGtSImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instSetGtSImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	if int64(interp.Registers[rB]) > int64(vX) {
		interp.Registers[rA] = 1
	} else {
		interp.Registers[rA] = 0
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s > %s) = (0x%x > %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(int64(vX)), formatInt(interp.Registers[rB]), formatInt(int64(vX)), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 144
func instShloLImmAlt32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instShloLImmAlt32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	imm, err := SignExtend(4, uint64(uint32(vX<<(interp.Registers[rB]&31))))
	if err != nil {
		pvmLogger.Errorf("instShloLImmAlt32 signExtend error: %v", err)
		return ExitHalt, pc
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s) ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 145
func instShloRImmAlt32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instShloRImmAlt32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	imm, err := SignExtend(4, uint64(uint32(vX)>>(interp.Registers[rB]&31)))
	if err != nil {
		pvmLogger.Errorf("instShloRImmAlt32 signExtend error: %v", err)
		return ExitHalt, pc
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]&31), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 146
func instSharRImmAlt32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instSharRImmAlt32 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	imm := uint64(int32(uint32(vX)) >> (interp.Registers[rB] & 31))
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> 0x%x) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(uint32(vX)), RegName[rB], formatInt(uint32(vX)), interp.Registers[rB], formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 147
func instCmovIzImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instCmovIzImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	if interp.Registers[rB] == 0 {
		interp.Registers[rA] = vX
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s (%s == 0)", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rA], formatInt(vX), RegName[rB])
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s (%s != 0)", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rA], formatInt(interp.Registers[rA]), RegName[rB])
	}

	return ExitContinue, pc
}

// opcode 148
func instCmovNzImm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instCmovNzImm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	if interp.Registers[rB] != 0 {
		interp.Registers[rA] = vX
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s (%s != 0)", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rA], formatInt(vX), RegName[rB])
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s (%s == 0)", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rA], formatInt(vX), RegName[rB])
	}

	return ExitContinue, pc
}

// opcode 149
func instAddImm64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instAddImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = interp.Registers[rB] + vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s)  = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 150
func instMulImm64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instMulImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = interp.Registers[rB] * vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = (%s • %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 151
func instShloLImm64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instShloLImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	imm, err := SignExtend(8, interp.Registers[rB]<<(vX&63))
	if err != nil {
		pvmLogger.Errorf("instShloLImm64 signExtend error: %v", err)
		return ExitHalt, pc
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX&63), formatInt(interp.Registers[rB]), formatInt(vX&63), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 152
func instShloRImm64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instShloRImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	imm, err := SignExtend(8, interp.Registers[rB]>>(vX&63))
	if err != nil {
		pvmLogger.Errorf("instShloRImm64 signExtend error: %v", err)
		return ExitHalt, pc
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX&63), formatInt(interp.Registers[rB]), formatInt(vX&63), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 153
func instSharRImm64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instSharRImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = uint64(int64(interp.Registers[rB]) >> (vX & 63))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], RegName[rB], formatInt(vX&63), formatInt(interp.Registers[rB]), formatInt(vX&63), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 154
func instNegAddImm64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instNegAddImm64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = vX - interp.Registers[rB]
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + (1<<64) - %s) = (%s + (1<<64) - %s) = %s ", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 155
func instShloLImmAlt64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instShloLImmAlt64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = vX << (interp.Registers[rB] & 63)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]&63), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 156
func instShloRImmAlt64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instShloRImmAlt64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = vX >> (interp.Registers[rB] & 63)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s)", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]&63), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 157
func instSharRImmAlt64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instSharRImmAlt64 decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	interp.Registers[rA] = uint64(int64(vX) >> (interp.Registers[rB] & 63))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s)", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(int64(vX)), RegName[rB], formatInt(int64(vX)), formatInt(interp.Registers[rB]&63), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 158
func instRotR64Imm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instRotR64Imm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	// rotate right
	interp.Registers[rA] = bits.RotateLeft64(interp.Registers[rB], -int(vX))
	// interp.Registers[rA] = (interp.Registers[rB] >> vX) | (interp.Registers[rB] << (64 - vX))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 159
func instRotR64ImmAlt(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instRotR64ImmAlt decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	// rotate right
	interp.Registers[rB] &= 63 // % 64
	interp.Registers[rA] = bits.RotateLeft64(vX, -int(interp.Registers[rB]))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 160
func instRotR32Imm(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instRotR32Imm decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	// rotate right
	imm := bits.RotateLeft32(uint32(interp.Registers[rB]), -int(vX))

	val, err := SignExtend(4, uint64(imm))
	if err != nil {
		pvmLogger.Errorf("instRotR32Imm signExtend error: %v", err)
		return ExitPanic, pc
	}
	interp.Registers[rA] = val
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 161
func instRotR32ImmAlt(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instRotR32ImmAlt decodeTwoRegistersAndOneImmediate error: %v", err)
		return ExitHalt, pc
	}

	// rotate right
	imm := bits.RotateLeft32(uint32(vX), -int(interp.Registers[rB]))

	val, err := SignExtend(4, uint64(imm))
	if err != nil {
		pvmLogger.Errorf("instRotR32ImmAlt signExtend error: %v", err)
		return ExitPanic, pc
	}
	interp.Registers[rA] = val
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rA], formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode in [170, 175]
func instBranch(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, err := decodeTwoRegistersAndOneOffset(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		return ExitHalt, pc
	}
	var op string
	branchCondition := false
	switch interp.Program.InstructionData[pc] {
	case 170:
		branchCondition = interp.Registers[rA] == interp.Registers[rB]
		op = "=="
	case 171:
		branchCondition = interp.Registers[rA] != interp.Registers[rB]
		op = "!="
	case 172:
		branchCondition = interp.Registers[rA] < interp.Registers[rB]
		op = "<"
	case 173:
		branchCondition = int64(interp.Registers[rA]) < int64(interp.Registers[rB])
		op = "<(signed)"
	case 174:
		branchCondition = interp.Registers[rA] >= interp.Registers[rB]
		op = ">="
	case 175:
		branchCondition = int64(interp.Registers[rA]) >= int64(interp.Registers[rB])
		op = ">=(signed)"
	default:
		pvmLogger.Fatalf("instBranch is supposed to be called with opcode in [170, 175]")
	}

	reason, newPC := branch(pc, vX, branchCondition, interp.Program.Bitmasks, interp.Program.InstructionData)
	if reason != ExitContinue {
		pvmLogger.Errorf("[%d]: pc: %d, %s panic", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])])
		return ExitReason(reason), pc
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s branch(%d, %s=%s %s %s=%s) = %t",
		interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], vX, RegName[rA], formatInt(interp.Registers[rA]), op, RegName[rB], formatInt(interp.Registers[rB]), branchCondition)
	return reason, newPC
}

// opcode 180
func instLoadImmJumpInd(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, vX, vY, err := decodeTwoRegistersAndTwoImmediates(interp.Program.InstructionData, pc, skipLength)
	if err != nil {
		pvmLogger.Errorf("instLoadImmJumpInd decodeTwoRegistersAndTwoImmediates error: %v", err)
		return ExitPanic, pc
	}
	// per https://github.com/koute/jamtestvectors/blob/master_pvm_initial/pvm/TESTCASES.md#inst_load_imm_and_jump_indirect_invalid_djump_to_zero_different_regs_without_offset_nok
	// the register update should take place even if the jump panics
	dest := uint32(interp.Registers[rB] + vY)
	reason, newPC := djump(pc, dest, interp.Program.JumpTable, interp.Program.Bitmasks)

	interp.Registers[rA] = vX
	switch reason {
	case ExitPanic:
		pvmLogger.Debugf("[%d]: pc: %d PANIC, %s, %v", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], reason)
		return reason, pc
	case ExitHalt:
		pvmLogger.Debugf("[%d]: pc: %d HALT, %s, %v", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])], reason)
		return reason, pc
	default:
		pvmLogger.Debugf("[%d]: pc: %d, %s, (%s + %s) = (%s + %s) mod (1<<32) = %s)", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rB], formatInt(vY), formatInt(interp.Registers[rB]), formatInt(vY), formatInt(dest))
		return reason, newPC
	}
}

// opcode 190
func instAdd32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instAdd32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD], err = SignExtend(4, uint64(uint32(interp.Registers[rA]+interp.Registers[rB])))
	if err != nil {
		pvmLogger.Errorf("instAdd32 signExtend error: %v", err)
		return ExitHalt, pc
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s) = u32(%s + %s)  = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 191
func instSub32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instSub32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	// bMod32 := uint32(interp.Registers[rB])
	interp.Registers[rD], err = SignExtend(4, uint64(uint32(interp.Registers[rA])-uint32(interp.Registers[rB])))
	if err != nil {
		pvmLogger.Errorf("instSub32 signExtend error: %v", err)
		return ExitHalt, pc
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = u32(%s) - u32(%s) = u32(%s) - u32(%s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(uint32(interp.Registers[rA])), formatInt(uint32(interp.Registers[rB])), formatInt(interp.Registers[rA]))
	return ExitContinue, pc
}

// opcode 192
func instMul32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMul32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD], err = SignExtend(4, uint64(uint32(interp.Registers[rA]*interp.Registers[rB])))
	if err != nil {
		pvmLogger.Errorf("instMul32 signExtend error: %v", err)
		return ExitHalt, pc
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = (%s • %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 193
func instDivU32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instDivU32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	bMod32 := uint32(interp.Registers[rB])
	aMod32 := uint32(interp.Registers[rA])

	if bMod32 == 0 {
		interp.Registers[rD] = ^uint64(0) // 2^64 - 1
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rD], formatInt(interp.Registers[rD]))
	} else {
		interp.Registers[rD], err = SignExtend(4, uint64(aMod32/bMod32))
		if err != nil {
			pvmLogger.Errorf("instDivU32 signExtend error: %v", err)
			return ExitHalt, pc
		}
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s / %s) = (%s / %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	}

	return ExitContinue, pc
}

// opcode 194
func instDivS32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instDivS32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	a := int64(int32(interp.Registers[rA]))
	b := int64(int32(interp.Registers[rB]))

	if b == 0 {
		interp.Registers[rD] = ^uint64(0) // 2^64 - 1
	} else if a == int64(-1<<31) && b == -1 {
		interp.Registers[rD] = uint64(a)
	} else {
		interp.Registers[rD] = uint64(a / b)
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = 0x%x", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 195
func instRemU32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instRemU32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	bMod32 := uint32(interp.Registers[rB])
	aMod32 := uint32(interp.Registers[rA])

	if bMod32 == 0 {
		interp.Registers[rD], err = SignExtend(4, uint64(aMod32))
	} else {
		interp.Registers[rD], err = SignExtend(4, uint64(aMod32%bMod32))
	}
	if err != nil {
		pvmLogger.Errorf("instRemU32 signExtend error: %v", err)
		return ExitHalt, pc
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 196
func instRemS32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instRemS32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}

	a := int64(int32(interp.Registers[rA]))
	b := int64(int32(interp.Registers[rB]))

	if a == int64(-1<<31) && b == -1 {
		interp.Registers[rD] = 0
	} else {
		interp.Registers[rD] = uint64((smod(a, b)))
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 197
func instShloL32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instShloL32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	shift := interp.Registers[rB] % 32
	interp.Registers[rD], err = SignExtend(4, uint64(uint32(interp.Registers[rA]<<shift)))
	if err != nil {
		pvmLogger.Errorf("instShloL32 signExtend error: %v", err)
		return ExitHalt, pc
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 198
func instShloR32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instShloR32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}

	modA := uint32(interp.Registers[rA])
	shift := interp.Registers[rB] % 32
	interp.Registers[rD], err = SignExtend(4, uint64(modA>>shift))
	if err != nil {
		pvmLogger.Errorf("instShloR32 signExtend error: %v", err)
		return ExitHalt, pc
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 199
func instSharR32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instSharR32 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}

	signedA := int32(interp.Registers[rA])

	shift := interp.Registers[rB] % 32
	interp.Registers[rD] = uint64(signedA >> shift)

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(signedA), formatInt(shift), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 200
func instAdd64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instAdd64 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] + interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s) = (%s + %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 201
func instSub64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instSub64 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] + (^interp.Registers[rB] + 1)

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s - %s) = (%s - %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 202
func instMul64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMul64 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] * interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = (%s • %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 203
func instDivU64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instDivU64 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	if interp.Registers[rB] == 0 {
		interp.Registers[rD] = ^uint64(0) // 2^64 - 1
	} else {
		interp.Registers[rD] = interp.Registers[rA] / interp.Registers[rB]
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 204
func instDivS64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instDivS64 decodeThreeRegisters error: %v", err)
		return ExitHalt, pc
	}
	// mutation
	if interp.Registers[rB] == 0 {
		interp.Registers[rD] = ^uint64(0) // 2^64 - 1
	} else if int64(interp.Registers[rA]) == -(1<<63) && int64(interp.Registers[rB]) == -1 {
		interp.Registers[rD] = interp.Registers[rA]
	} else {
		interp.Registers[rD] = uint64((int64(interp.Registers[rA]) / int64(interp.Registers[rB])))
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 205
func instRemU64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instRemU64 decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	if interp.Registers[rB] == 0 {
		interp.Registers[rD] = interp.Registers[rA]
	} else {
		interp.Registers[rD] = interp.Registers[rA] % interp.Registers[rB]
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 206
func instRemS64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instRemS64 decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	if int64(interp.Registers[rA]) == -(1<<63) && int64(interp.Registers[rB]) == -1 {
		interp.Registers[rD] = 0
	} else {
		interp.Registers[rD] = uint64(smod(int64(interp.Registers[rA]), int64(interp.Registers[rB])))
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 207
func instShloL64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instShloL64 decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] << (interp.Registers[rB] % 64)

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]%64), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 208
func instShloR64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instShloR64 decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] >> (interp.Registers[rB] % 64)

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]%64), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 209
func instSharR64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instSharR64 decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = uint64(int64(interp.Registers[rA]) >> (interp.Registers[rB] % 64))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(int64(interp.Registers[rA])), formatInt(interp.Registers[rB]%64), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 210
func instAnd(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instAnd decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] & interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s & %s) = (%s & %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 211
func instXor(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instXor decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] ^ interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s ^ %s) = (%s ^ %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 212
func instOr(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instOr decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] | interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s | %s) = (%s | %s) = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 213
func instMulUpperSS(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMulUpperSS decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	signedA := int64(interp.Registers[rA])
	signedB := int64(interp.Registers[rB])

	hi, _ := bits.Mul64(uint64(abs(signedA)), uint64(abs(signedB)))

	if (signedA < 0) == (signedB < 0) {
		interp.Registers[rD] = hi
	} else {
		interp.Registers[rD] = uint64(-int64(hi))
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 214
func instMulUpperUU(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMulUpperUU decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	hi, _ := bits.Mul64(interp.Registers[rA], interp.Registers[rB])
	interp.Registers[rD] = hi

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 215
func instMulUpperSU(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMulUpperSU decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	signedA := int64(interp.Registers[rA])
	hi, lo := bits.Mul64(uint64(abs(signedA)), interp.Registers[rB])

	if signedA < 0 {
		hi = -hi
		if lo != 0 { // 2's complement, borrow 1 from hi
			hi--
		}
		interp.Registers[rD] = hi

	} else {
		interp.Registers[rD] = hi
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 216
func instSetLtU(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instSetLtU decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	if interp.Registers[rA] < interp.Registers[rB] {
		interp.Registers[rD] = 1
	} else {
		interp.Registers[rD] = 0
	}

	return ExitContinue, pc
}

// opcode 217
func instSetLtS(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instSetLts decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	if int64(interp.Registers[rA]) < int64(interp.Registers[rB]) {
		interp.Registers[rD] = 1
	} else {
		interp.Registers[rD] = 0
	}

	return ExitContinue, pc
}

// opcode 218
func instCmovIz(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instCmovIz decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	if interp.Registers[rB] == 0 {
		interp.Registers[rD] = interp.Registers[rA]
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rD], RegName[rA], formatInt(interp.Registers[rD]))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rD], formatInt(interp.Registers[rD]))
	}

	return ExitContinue, pc
}

// opcode 219
func instCmovNz(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instCmovNz decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	if interp.Registers[rB] != 0 {
		interp.Registers[rD] = interp.Registers[rA]
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rD], RegName[rA], formatInt(interp.Registers[rD]))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
			RegName[rD], formatInt(interp.Registers[rD]))
	}

	return ExitContinue, pc
}

// opcode 220
func instRotL64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instRotL64 decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = bits.RotateLeft64(interp.Registers[rA], int(interp.Registers[rB]%64))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 221
func instRotL32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instRotL32 decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	rotated := uint64(bits.RotateLeft32(uint32(interp.Registers[rA]), int(interp.Registers[rB]%32)))
	extend, err := SignExtend(4, rotated)
	if err != nil {
		pvmLogger.Errorf("instRoTL32 signExtend error:%v", err)
		return ExitHalt, pc
	}
	interp.Registers[rD] = extend

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 222
func instRotR64(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instRotR64 decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = bits.RotateLeft64(interp.Registers[rA], -int(interp.Registers[rB]))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 223
func instRotR32(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instRotR32 decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	rotated := uint64(bits.RotateLeft32(uint32(interp.Registers[rA]), -int(interp.Registers[rB])))
	extend, err := SignExtend(4, rotated)
	if err != nil {
		pvmLogger.Errorf("instRotR32 signExtend error:%v", err)
		return ExitHalt, pc
	}
	interp.Registers[rD] = extend

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 224
func instAndInv(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instAndInv decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] & ^interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 225
func instOrInv(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instOrInv decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = interp.Registers[rA] | ^interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 226
func instXnor(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instXnor decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}
	// mutation
	interp.Registers[rD] = ^(interp.Registers[rA] ^ interp.Registers[rB])

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 227
func instMax(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMax decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}

	// mutation
	interp.Registers[rD] = uint64(max(int64(interp.Registers[rA]), int64(interp.Registers[rB])))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 228
func instMaxU(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMaxU decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}

	// mutation
	if interp.Registers[rA] > interp.Registers[rB] {
		interp.Registers[rD] = interp.Registers[rA]
	} else {
		interp.Registers[rD] = interp.Registers[rB]
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 229
func instMin(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf(" decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}

	// mutation
	interp.Registers[rD] = uint64(min(int64(interp.Registers[rA]), int64(interp.Registers[rB])))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}

// opcode 230
func instMinU(interp *Interpreter, pc ProgramCounter, skipLength ProgramCounter) (ExitReason, ProgramCounter) {
	rA, rB, rD, err := decodeThreeRegisters(interp.Program.InstructionData, pc)
	if err != nil {
		pvmLogger.Errorf("instMinU decodeThreeRegisters error:%v", err)
		return ExitHalt, pc
	}

	// mutation
	if interp.Registers[rA] < interp.Registers[rB] {
		interp.Registers[rD] = interp.Registers[rA]
	} else {
		interp.Registers[rD] = interp.Registers[rB]
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, pc, zeta[opcode(interp.Program.InstructionData[pc])],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, pc
}
