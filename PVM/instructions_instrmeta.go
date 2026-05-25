package PVM

import (
	"math/bits"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

var execInstructionsMeta = [231]func(*Interpreter, *InstrMeta) (ExitReason, ProgramCounter){
	// A.5.1 Instructiopns without Arguments
	0: instTrapMeta,
	1: instFallthroughMeta,
	// A.5.2 Instructions with Arguments of One Immediate
	10: instEcalliMeta,
	// A.5.3 Instructions with Arguments of One Register & One Extended With Immediate
	20: instLoadImm64Meta, // passed testvector
	// A.5.4 Instructions with Arguments of Two Immediates
	30: instStoreImmU8Meta,
	31: instStoreImmU16Meta,
	32: instStoreImmU32Meta,
	33: instStoreImmU64Meta,
	// A.5.5 Instructions with Arguments of One Offset
	40: instJumpMeta,
	// A.5.6 Instructions with Arguments of One Register & One Immediate
	50: instJumpIndMeta,
	51: instLoadImmMeta,
	52: instLoadU8Meta,
	53: instLoadI8Meta,
	54: instLoadU16Meta,
	55: instLoadI16Meta,
	56: instLoadU32Meta,
	57: instLoadI32Meta,
	58: instLoadU64Meta,
	59: instStoreU8Meta,
	60: instStoreU16Meta,
	61: instStoreU32Meta,
	62: instStoreU64Meta,
	// A.5.7 Instructions with Arguments of One Register & Two Immediates
	70: instStoreImmIndU8Meta,
	71: instStoreImmIndU16Meta,
	72: instStoreImmIndU32Meta,
	73: instStoreImmIndU64Meta,
	// A.5.8 Instructions without Arguments of One Register, One Immediate and One Offset
	80: instImmediateBranchMeta,
	81: instImmediateBranchMeta,
	82: instImmediateBranchMeta,
	83: instImmediateBranchMeta,
	84: instImmediateBranchMeta,
	85: instImmediateBranchMeta,
	86: instImmediateBranchMeta,
	87: instImmediateBranchMeta,
	88: instImmediateBranchMeta,
	89: instImmediateBranchMeta,
	90: instImmediateBranchMeta,
	// A.5.9 Instructions with arguments of Two Registers
	100: instMoveRegMeta, // passed testvector
	101: instSbrkMeta,
	102: instCountSetBits64Meta,
	103: instCountSetBits32Meta,
	104: instLeadingZeroBits64Meta,
	105: instLeadingZeroBits32Meta,
	106: instTrailZeroBits64Meta,
	107: instTrailZeroBits32Meta,
	108: instSignExtend8Meta,
	109: instSignExtend16Meta,
	110: instZeroExtend16Meta,
	111: instReverseBytesMeta,
	120: instStoreIndU8Meta,
	121: instStoreIndU16Meta,
	122: instStoreIndU32Meta,
	123: instStoreIndU64Meta,
	124: instLoadIndU8Meta,
	125: instLoadIndI8Meta,
	126: instLoadIndU16Meta,
	127: instLoadIndI16Meta,
	128: instLoadIndU32Meta,
	129: instLoadIndI32Meta,
	130: instLoadIndU64Meta,
	131: instAddImm32Meta,
	132: instAndImmMeta,
	133: instXORImmMeta,
	134: instORImmMeta,
	135: instMulImm32Meta,
	136: instSetLtUImmMeta,
	137: instSetLtSImmMeta,
	138: instShloLImm32Meta,
	139: instShloRImm32Meta,
	140: instSharRImm32Meta,
	141: instNegAddImm32Meta,
	142: instSetGtUImmMeta,
	143: instSetGtSImmMeta,
	144: instShloLImmAlt32Meta,
	145: instShloRImmAlt32Meta,
	146: instSharRImmAlt32Meta,
	147: instCmovIzImmMeta,
	148: instCmovNzImmMeta,
	149: instAddImm64Meta,
	150: instMulImm64Meta,
	151: instShloLImm64Meta,
	152: instShloRImm64Meta,
	153: instSharRImm64Meta,
	154: instNegAddImm64Meta,
	155: instShloLImmAlt64Meta,
	156: instShloRImmAlt64Meta,
	157: instSharRImmAlt64Meta,
	158: instRotR64ImmMeta,
	159: instRotR64ImmAltMeta,
	160: instRotR32ImmMeta,
	161: instRotR32ImmAltMeta,
	170: instBranchMeta,
	171: instBranchMeta,
	172: instBranchMeta,
	173: instBranchMeta,
	174: instBranchMeta,
	175: instBranchMeta,
	// A.5.12 Instructions  with Arguments of Two Registers and Two Immediates
	180: instLoadImmJumpIndMeta,
	// A.5.13 Instructions with Arguments of Three Registers
	190: instAdd32Meta,
	191: instSub32Meta,
	192: instMul32Meta,
	193: instDivU32Meta,
	194: instDivS32Meta,
	195: instRemU32Meta,
	196: instRemS32Meta,
	197: instShloL32Meta,
	198: instShloR32Meta,
	199: instSharR32Meta,
	200: instAdd64Meta,
	201: instSub64Meta,
	202: instMul64Meta,
	203: instDivU64Meta,
	204: instDivS64Meta,
	205: instRemU64Meta,
	206: instRemS64Meta,
	207: instShloL64Meta,
	208: instShloR64Meta,
	209: instSharR64Meta,
	210: instAndMeta,
	211: instXorMeta,
	212: instOrMeta,
	213: instMulUpperSSMeta,
	214: instMulUpperUUMeta,
	215: instMulUpperSUMeta,
	216: instSetLtUMeta,
	217: instSetLtSSMeta,
	218: instCmovIzMeta,
	219: instCmovNzMeta,
	220: instRotL64Meta,
	221: instRotL32Meta,
	222: instRotR64Meta,
	223: instRotR32Meta,
	224: instAndInvMeta,
	225: instOrInvMeta,
	226: instXnorMeta,
	227: instMaxMeta,
	228: instMaxUMeta,
	229: instMinMeta,
	230: instMinUMeta,
	// register more instructions here
}

// opcode 0
func instTrapMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	pvmLogger.Debugf("[%d]: pc: %d, %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)])
	return ExitPanic, instr.PC
}

// opcode 1
func instFallthroughMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	pvmLogger.Debugf("[%d]: pc: %d, %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)])
	return ExitContinue, instr.PC
}

// opcode 10
func instEcalliMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	nuX := instr.Imm[0]
	pvmLogger.Debugf("[%d]: pc: %d, %s %d", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], nuX)
	return ExitHostCall | ExitReason(nuX), instr.PC
}

// opcode 20
func instLoadImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	interp.Registers[instr.Dst] = instr.Imm[0]
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[instr.Dst], formatInt(instr.Imm[0]))
	return ExitContinue, instr.PC
}

// opcode 30
func instStoreImmU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vx, vy := instr.Imm[0], uint64(uint8(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 1, uint32(vx), vy)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vx), formatInt(vy))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vx))
	}
	return exitReason, instr.PC
}

// opcode 31
func instStoreImmU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vx, vy := instr.Imm[0], uint64(uint16(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 2, uint32(vx), vy)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vx), formatInt(vy))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vx))
	}
	return exitReason, instr.PC
}

// opcode 32
func instStoreImmU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vx, vy := instr.Imm[0], uint64(uint32(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 4, uint32(vx), vy)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vx), formatInt(vy))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vx))
	}
	return exitReason, instr.PC
}

// opcode 33
func instStoreImmU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vx, vy := instr.Imm[0], instr.Imm[1]
	exitReason := storeIntoMemory(interp, 8, uint32(vx), vy)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vx), formatInt(vy))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vx))
	}
	return exitReason, instr.PC
}

// opcode 40
func instJumpMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vX := ProgramCounter(instr.Imm[0])
	reason, newPC := branch(instr.PC, vX, true, interp.Program.Bitmasks, interp.Program.InstructionData)
	if reason != ExitContinue {
		pvmLogger.Debugf("[%d]: pc: %d, %s %d panic", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], newPC)
		return reason, instr.PC
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s %d", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], newPC)
	return reason, newPC
}

// opcode 50
func instJumpIndMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	vX := instr.Imm[0]
	dest := uint32(interp.Registers[rA] + vX)
	reason, newPC := djump(instr.PC, dest, interp.Program.JumpTable, interp.Program.Bitmasks)
	switch reason {
	case ExitPanic:
		pvmLogger.Debugf("[%d]: pc: %d, %s %d panic, %s = %s, vX = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			newPC, RegName[rA], formatInt(interp.Registers[rA]), formatInt(vX))
		return reason, instr.PC
	case ExitHalt:
		pvmLogger.Debugf("[%d]: pc: %d, %s %d HALT, %s = %s, vX = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			newPC, RegName[rA], formatInt(interp.Registers[rA]), formatInt(vX))
		return reason, instr.PC
	default:
		pvmLogger.Debugf("[%d]: pc: %d, %s %d, %s = %s, vX = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			newPC, RegName[rA], formatInt(interp.Registers[rA]), formatInt(vX))
		return reason, newPC
	}
}

// opcode 51
func instLoadImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	interp.Registers[instr.Dst] = instr.Imm[0]
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[instr.Dst], formatInt(interp.Registers[instr.Dst]))
	return ExitContinue, instr.PC
}

// opcode 52
func instLoadU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vX := instr.Imm[0]
	memVal, exitReason := loadFromMemory(interp, 1, uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}
	interp.Registers[instr.Dst] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[instr.Dst], formatInt(memVal))
	return ExitContinue, instr.PC
}

// opcode 53
func instLoadI8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vX := instr.Imm[0]
	memVal, exitReason := loadFromMemory(interp, 1, uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}
	extend, err := SignExtend(1, memVal)
	if err != nil {
		pvmLogger.Errorf("instLoadI8 SignExtend error: %v", err)
		return ExitPanic, instr.PC
	}
	interp.Registers[instr.Dst] = extend
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[instr.Dst], formatInt(memVal))
	return ExitContinue, instr.PC
}

// opcode 54
func instLoadU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vX := instr.Imm[0]
	memVal, exitReason := loadFromMemory(interp, 2, uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}
	interp.Registers[instr.Dst] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[instr.Dst], formatInt(memVal))
	return ExitContinue, instr.PC
}

// opcode 55
func instLoadI16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vX := instr.Imm[0]
	memVal, exitReason := loadFromMemory(interp, 2, uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}
	extend, err := SignExtend(2, memVal)
	if err != nil {
		pvmLogger.Errorf("instLoadI16 signExtend error: %v", err)
		return ExitPanic, instr.PC
	}
	interp.Registers[instr.Dst] = extend
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[instr.Dst], formatInt(extend))
	return ExitContinue, instr.PC
}

// opcode 56
func instLoadU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vX := instr.Imm[0]
	memVal, exitReason := loadFromMemory(interp, 4, uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}
	interp.Registers[instr.Dst] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[instr.Dst], formatInt(memVal))
	return ExitContinue, instr.PC
}

// opcode 57
func instLoadI32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vX := instr.Imm[0]
	memVal, exitReason := loadFromMemory(interp, 4, uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}
	extend, err := SignExtend(4, memVal)
	if err != nil {
		pvmLogger.Errorf("instLoadI32 signExtend error: %v", err)
		return ExitPanic, instr.PC
	}
	interp.Registers[instr.Dst] = extend
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[instr.Dst], formatInt(extend))
	return ExitContinue, instr.PC
}

// opcode 58
func instLoadU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vX := instr.Imm[0]
	memVal, exitReason := loadFromMemory(interp, 8, uint32(vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}
	interp.Registers[instr.Dst] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[instr.Dst], formatInt(memVal))
	return ExitContinue, instr.PC
}

// opcode 59
func instStoreU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX := instr.Dst, instr.Imm[0]
	exitReason := storeIntoMemory(interp, 1, uint32(vX), uint64(uint8(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vX), RegName[rA], formatInt(uint64(uint8(interp.Registers[rA]))))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vX))
	}
	return exitReason, instr.PC
}

// opcode 60
func instStoreU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX := instr.Dst, instr.Imm[0]
	exitReason := storeIntoMemory(interp, 2, uint32(vX), uint64(uint16(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vX), RegName[rA], formatInt(uint64(uint16(interp.Registers[rA]))))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vX))
	}
	return exitReason, instr.PC
}

// opcode 61
func instStoreU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX := instr.Dst, instr.Imm[0]
	exitReason := storeIntoMemory(interp, 4, uint32(vX), uint64(uint32(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vX), RegName[rA], formatInt(uint64(uint32(interp.Registers[rA]))))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vX))
	}
	return exitReason, instr.PC
}

// opcode 62
func instStoreU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX := instr.Dst, instr.Imm[0]
	exitReason := storeIntoMemory(interp, 8, uint32(vX), interp.Registers[rA])
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ 0x%x ] = %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vX), RegName[rA], formatInt(interp.Registers[rA]))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], uint32(vX))
	}
	return exitReason, instr.PC
}

// opcode 70
func instStoreImmIndU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX, vY := instr.Src[0], instr.Imm[0], uint64(uint8(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 1, uint32(interp.Registers[rA]+vX), vY)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX), formatInt(vY))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX))
	}
	return exitReason, instr.PC
}

// opcode 71
func instStoreImmIndU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX, vY := instr.Src[0], instr.Imm[0], uint64(uint16(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 2, uint32(interp.Registers[rA]+vX), vY)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX), formatInt(vY))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX))
	}
	return exitReason, instr.PC
}

// opcode 72
func instStoreImmIndU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX, vY := instr.Src[0], instr.Imm[0], uint64(uint32(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 4, uint32(interp.Registers[rA]+vX), vY)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX), formatInt(vY))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s= 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX))
	}
	return exitReason, instr.PC
}

// opcode 73
func instStoreImmIndU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX, vY := instr.Src[0], instr.Imm[0], instr.Imm[1]
	exitReason := storeIntoMemory(interp, 8, uint32(interp.Registers[rA]+vX), vY)
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX), formatInt(vY))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], formatInt(uint32(vX)), uint32(interp.Registers[rA]+vX))
	}
	return exitReason, instr.PC
}

// opcode in [80, 90]
func instImmediateBranchMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	vX := instr.Imm[0]
	vY := ProgramCounter(instr.Imm[1])
	branchCondition := false

	switch instr.Opcode {
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

	reason, newPC := branch(instr.PC, vY, branchCondition, interp.Program.Bitmasks, interp.Program.InstructionData)
	if reason != ExitContinue {
		pvmLogger.Debugf("[%d]: pc: %d, %s panic", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)])
		return reason, instr.PC
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s branch(%d, %s=%s, vX=%s) = %t",
		interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], vY, RegName[rA], formatInt(interp.Registers[rA]), formatInt(vX), branchCondition)
	return reason, newPC
}

// opcode 100
func instMoveRegMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	// mutation
	interp.Registers[rD] = interp.Registers[rA]
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], RegName[rA], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 101
func instSbrkMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]

	// this reivision is according to jam-test-vector traces: Note on SBRK
	if interp.Registers[rA] == 0 {
		interp.Registers[rD] = interp.Memory.heapPointer
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rD], formatInt(interp.Registers[rD]))
		return ExitContinue, instr.PC
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

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s + %s = %s + %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rD], RegName[rA], formatInt(interp.Registers[rD]), formatInt(interp.Registers[rA]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 102
func instCountSetBits64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
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
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 103
func instCountSetBits32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
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
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 104
func instLeadingZeroBits64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
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
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 105
func instLeadingZeroBits32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
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
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 106
func instTrailZeroBits64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
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
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 107
func instTrailZeroBits32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
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
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 108
func instSignExtend8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	// mutation
	regA := interp.Registers[rA]
	signedInt := int8(regA)
	unsignedInt := uint64(signedInt)

	interp.Registers[rD] = unsignedInt
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 109
func instSignExtend16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	// mutation
	regA := interp.Registers[rA]
	signedInt := int16(regA)
	unsignedInt := uint64(signedInt)

	interp.Registers[rD] = unsignedInt
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 110
func instZeroExtend16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	// mutation
	regA := interp.Registers[rA]
	interp.Registers[rD] = regA % (1 << 16)
	pvmLogger.Debugf("[%d]: pc: %d, %s , %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 111
func instReverseBytesMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	// mutation
	regA := types.U64(interp.Registers[rA])
	bytes := utils.SerializeFixedLength(regA, types.U64(8))
	var reversedBytes uint64 = 0
	for i := uint8(0); i < 8; i++ {
		reversedBytes = (reversedBytes << 8) | uint64(bytes[i])
	}
	interp.Registers[rD] = reversedBytes
	pvmLogger.Debugf("[%d]: pc: %d, %s , %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rD], formatInt(reversedBytes))
	return ExitContinue, instr.PC
}

// opcode 120
func instStoreIndU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 1
	exitReason := storeIntoMemory(interp, offset, uint32(interp.Registers[rB]+vX), uint64(uint8(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX), formatInt(uint64(uint8(interp.Registers[rA]))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+0x%x = 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rB], uint32(vX), uint32(interp.Registers[rB]+vX))
	}
	return exitReason, instr.PC
}

// opcode 121
func instStoreIndU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 2
	exitReason := storeIntoMemory(interp, offset, uint32(interp.Registers[rB]+vX), uint64(uint16(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s,[ %s+%s = 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rB], formatInt(uint32(vX)), formatInt(uint32(interp.Registers[rB]+vX)), formatInt(int64(uint16(interp.Registers[rA]))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s, page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rB], formatInt(uint32(vX)), formatInt(uint32(interp.Registers[rB]+vX)))
	}
	return exitReason, instr.PC
}

// opcode 122
func instStoreIndU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 4
	exitReason := storeIntoMemory(interp, offset, uint32(interp.Registers[rB]+vX), uint64(uint32(interp.Registers[rA])))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s ,[ %s+%s = 0x%x ] = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX), formatInt(uint64(uint32(interp.Registers[rA]))))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s , page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX))
	}
	return exitReason, instr.PC
}

// opcode 123
func instStoreIndU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 8
	exitReason := storeIntoMemory(interp, offset, uint32(interp.Registers[rB]+vX), uint64(interp.Registers[rA]))
	if exitReason.GetReasonType() == CONTINUE {
		pvmLogger.Debugf("[%d]: pc: %d, %s ,[ %s+%s = 0x%x...+%d ] = %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX), offset, RegName[rA], formatInt(uint64(interp.Registers[rA])))
	} else { // page fault error
		pvmLogger.Debugf("[%d]: pc: %d, %s , page fault error at mem[ %s+%s = 0x%x ]", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rB], formatInt(uint32(vX)), uint32(interp.Registers[rB]+vX))
	}

	return exitReason, instr.PC
}

// opcode 124
func instLoadIndU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 1
	memVal, exitReason := loadFromMemory(interp, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA],
		RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 125
func instLoadIndI8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 1
	memVal, exitReason := loadFromMemory(interp, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}

	interp.Registers[rA] = uint64(int8(memVal))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 126
func instLoadIndU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 2
	memVal, exitReason := loadFromMemory(interp, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 127
func instLoadIndI16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 2
	memVal, exitReason := loadFromMemory(interp, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}

	interp.Registers[rA] = uint64(int16(memVal))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 128
func instLoadIndU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 4
	memVal, exitReason := loadFromMemory(interp, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 129
func instLoadIndI32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 4
	memVal, exitReason := loadFromMemory(interp, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}

	interp.Registers[rA] = uint64(int32(memVal))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 130
func instLoadIndU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 8
	memVal, exitReason := loadFromMemory(interp, uint32(offset), uint32(interp.Registers[rB]+vX))
	if exitReason != ExitContinue {
		return exitReason, instr.PC
	}

	interp.Registers[rA] = memVal
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = mem[ %s+%s = 0x%x ] = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), uint32(interp.Registers[rB]+vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 131
func instAddImm32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	val, err := SignExtend(4, uint64(uint32(interp.Registers[rB]+vX)))
	if err != nil {
		pvmLogger.Errorf("instAddImm32 SignExtend error: %v", err)
	}
	interp.Registers[rA] = val
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 132
func instAndImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] & vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s & %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 133
func instXORImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] ^ vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s ^ %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 134
func instORImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] | vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s | %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 135
func instMulImm32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	val, err := SignExtend(4, uint64(uint32(interp.Registers[rB]*vX)))
	if err != nil {
		pvmLogger.Errorf("instMulImm32 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}
	interp.Registers[rA] = val
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 136
func instSetLtUImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	if interp.Registers[rB] < vX {
		interp.Registers[rA] = 1
	} else {
		interp.Registers[rA] = 0
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s < %s) = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 137
func instSetLtSImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	if int64(interp.Registers[rB]) < int64(vX) {
		interp.Registers[rA] = 1
	} else {
		interp.Registers[rA] = 0
	}

	return ExitContinue, instr.PC
}

// opcode 138
func instShloLImm32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	vX = vX & 31 // % 32
	imm, err := SignExtend(4, uint64(uint32(interp.Registers[rB]<<vX)))
	if err != nil {
		pvmLogger.Errorf("instShloLImm32 SignExtend error: %v", err)
		return ExitHalt, instr.PC
	}
	interp.Registers[rA] = imm

	return ExitContinue, instr.PC
}

// opcode 139
func instShloRImm32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	vX = vX & 31 // % 32
	imm, err := SignExtend(4, uint64(uint32(interp.Registers[rB])>>vX))
	if err != nil {
		pvmLogger.Errorf("instShloRImm32 signExtend error: %v", err)
		return ExitPanic, instr.PC
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 140
func instSharRImm32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	vX = vX & 31 // % 32
	interp.Registers[rA] = uint64(int32(interp.Registers[rB]) >> vX)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = 0x%x", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 141
func instNegAddImm32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	imm, err := SignExtend(4, uint64(uint32(vX+(1<<32)-interp.Registers[rB])))
	if err != nil {
		pvmLogger.Errorf("instNegAddImm32 signExtend: %v", err)
		return ExitHalt, instr.PC
	}
	interp.Registers[rA] = uint64(imm)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (0x%x + (1<<32) - %s) = (0x%x + (1<<32) - %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], vX, RegName[rB], vX, formatInt(interp.Registers[rB]), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 142
func instSetGtUImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	if interp.Registers[rB] > vX {
		interp.Registers[rA] = 1
	} else {
		interp.Registers[rA] = 0
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s > %s) = (%s > %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 143
func instSetGtSImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	if int64(interp.Registers[rB]) > int64(vX) {
		interp.Registers[rA] = 1
	} else {
		interp.Registers[rA] = 0
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s > %s) = (0x%x > %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(int64(vX)), formatInt(interp.Registers[rB]), formatInt(int64(vX)), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 144
func instShloLImmAlt32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	imm, err := SignExtend(4, uint64(uint32(vX<<(interp.Registers[rB]&31))))
	if err != nil {
		pvmLogger.Errorf("instShloLImmAlt32 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s) ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 145
func instShloRImmAlt32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	imm, err := SignExtend(4, uint64(uint32(vX)>>(interp.Registers[rB]&31)))
	if err != nil {
		pvmLogger.Errorf("instShloRImmAlt32 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]&31), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 146
func instSharRImmAlt32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	imm := uint64(int32(uint32(vX)) >> (interp.Registers[rB] & 31))
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> 0x%x) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(uint32(vX)), RegName[rB], formatInt(uint32(vX)), interp.Registers[rB], formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 147
func instCmovIzImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	if interp.Registers[rB] == 0 {
		interp.Registers[rA] = vX
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s (%s == 0)", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rA], formatInt(vX), RegName[rB])
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s (%s != 0)", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rA], formatInt(interp.Registers[rA]), RegName[rB])
	}

	return ExitContinue, instr.PC
}

// opcode 148
func instCmovNzImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	if interp.Registers[rB] != 0 {
		interp.Registers[rA] = vX
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s (%s != 0)", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rA], formatInt(vX), RegName[rB])
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s (%s == 0)", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rA], formatInt(vX), RegName[rB])
	}

	return ExitContinue, instr.PC
}

// opcode 149
func instAddImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] + vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s)  = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 150
func instMulImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] * vX
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = (%s • %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(vX), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 151
func instShloLImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	imm, err := SignExtend(8, interp.Registers[rB]<<(vX&63))
	if err != nil {
		pvmLogger.Errorf("instShloLImm64 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX&63), formatInt(interp.Registers[rB]), formatInt(vX&63), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 152
func instShloRImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	imm, err := SignExtend(8, interp.Registers[rB]>>(vX&63))
	if err != nil {
		pvmLogger.Errorf("instShloRImm64 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}
	interp.Registers[rA] = imm
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX&63), formatInt(interp.Registers[rB]), formatInt(vX&63), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 153
func instSharRImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = uint64(int64(interp.Registers[rB]) >> (vX & 63))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], RegName[rB], formatInt(vX&63), formatInt(interp.Registers[rB]), formatInt(vX&63), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 154
func instNegAddImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = vX - interp.Registers[rB]
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + (1<<64) - %s) = (%s + (1<<64) - %s) = %s ", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 155
func instShloLImmAlt64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = vX << (interp.Registers[rB] & 63)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]&63), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 156
func instShloRImmAlt64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = vX >> (interp.Registers[rB] & 63)
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s)", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(vX), RegName[rB], formatInt(vX), formatInt(interp.Registers[rB]&63), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 157
func instSharRImmAlt64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = uint64(int64(vX) >> (interp.Registers[rB] & 63))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s)", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(int64(vX)), RegName[rB], formatInt(int64(vX)), formatInt(interp.Registers[rB]&63), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 158
func instRotR64ImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	// rotate right
	interp.Registers[rA] = bits.RotateLeft64(interp.Registers[rB], -int(vX))
	// interp.Registers[rA] = (interp.Registers[rB] >> vX) | (interp.Registers[rB] << (64 - vX))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 159
func instRotR64ImmAltMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	// rotate right
	interp.Registers[rB] &= 63 // % 64
	interp.Registers[rA] = bits.RotateLeft64(vX, -int(interp.Registers[rB]))
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 160
func instRotR32ImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	// rotate right
	imm := bits.RotateLeft32(uint32(interp.Registers[rB]), -int(vX))

	val, err := SignExtend(4, uint64(imm))
	if err != nil {
		pvmLogger.Errorf("instRotR32Imm signExtend error: %v", err)
		return ExitPanic, instr.PC
	}
	interp.Registers[rA] = val
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 161
func instRotR32ImmAltMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	// rotate right
	imm := bits.RotateLeft32(uint32(vX), -int(interp.Registers[rB]))

	val, err := SignExtend(4, uint64(imm))
	if err != nil {
		pvmLogger.Errorf("instRotR32ImmAlt signExtend error: %v", err)
		return ExitPanic, instr.PC
	}
	interp.Registers[rA] = val
	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rA], formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode in [170, 175]
func instBranchMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	vX := ProgramCounter(instr.Imm[0])
	var op string
	branchCondition := false
	switch instr.Opcode {
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

	reason, newPC := branch(instr.PC, vX, branchCondition, interp.Program.Bitmasks, interp.Program.InstructionData)
	if reason != ExitContinue {
		pvmLogger.Errorf("[%d]: pc: %d, %s panic", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)])
		return ExitReason(reason), instr.PC
	}
	pvmLogger.Debugf("[%d]: pc: %d, %s branch(%d, %s=%s %s %s=%s) = %t",
		interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], vX, RegName[rA], formatInt(interp.Registers[rA]), op, RegName[rB], formatInt(interp.Registers[rB]), branchCondition)
	return reason, newPC
}

// opcode 180
func instLoadImmJumpIndMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Dst
	rB := instr.Src[0]
	vX := instr.Imm[0]
	vY := instr.Imm[1]
	// per https://github.com/koute/jamtestvectors/blob/master_pvm_initial/pvm/TESTCASES.md#inst_load_imm_and_jump_indirect_invalid_djump_to_zero_different_regs_without_offset_nok
	// the register update should take place even if the jump panics
	dest := uint32(interp.Registers[rB] + vY)
	reason, newPC := djump(instr.PC, dest, interp.Program.JumpTable, interp.Program.Bitmasks)

	interp.Registers[rA] = vX
	switch reason {
	case ExitPanic:
		pvmLogger.Debugf("[%d]: pc: %d PANIC, %s, %v", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], reason)
		return reason, instr.PC
	case ExitHalt:
		pvmLogger.Debugf("[%d]: pc: %d HALT, %s, %v", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)], reason)
		return reason, instr.PC
	default:
		pvmLogger.Debugf("[%d]: pc: %d, %s, (%s + %s) = (%s + %s) mod (1<<32) = %s)", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rB], formatInt(vY), formatInt(interp.Registers[rB]), formatInt(vY), formatInt(dest))
		return reason, newPC
	}
}

// opcode 190
func instAdd32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	var err error
	interp.Registers[rD], err = SignExtend(4, uint64(uint32(interp.Registers[rA]+interp.Registers[rB])))
	if err != nil {
		pvmLogger.Errorf("instAdd32 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s) = u32(%s + %s)  = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 191
func instSub32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	// bMod32 := uint32(interp.Registers[rB])
	var err error
	interp.Registers[rD], err = SignExtend(4, uint64(uint32(interp.Registers[rA])-uint32(interp.Registers[rB])))
	if err != nil {
		pvmLogger.Errorf("instSub32 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = u32(%s) - u32(%s) = u32(%s) - u32(%s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(uint32(interp.Registers[rA])), formatInt(uint32(interp.Registers[rB])), formatInt(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 192
func instMul32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	var err error
	interp.Registers[rD], err = SignExtend(4, uint64(uint32(interp.Registers[rA]*interp.Registers[rB])))
	if err != nil {
		pvmLogger.Errorf("instMul32 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = (%s • %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 193
func instDivU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	bMod32 := uint32(interp.Registers[rB])
	aMod32 := uint32(interp.Registers[rA])

	if bMod32 == 0 {
		interp.Registers[rD] = ^uint64(0) // 2^64 - 1
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rD], formatInt(interp.Registers[rD]))
	} else {
		var err error
		interp.Registers[rD], err = SignExtend(4, uint64(aMod32/bMod32))
		if err != nil {
			pvmLogger.Errorf("instDivU32 signExtend error: %v", err)
			return ExitHalt, instr.PC
		}
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s / %s) = (%s / %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	}

	return ExitContinue, instr.PC
}

// opcode 194
func instDivS32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	a := int64(int32(interp.Registers[rA]))
	b := int64(int32(interp.Registers[rB]))

	if b == 0 {
		interp.Registers[rD] = ^uint64(0) // 2^64 - 1
	} else if a == int64(-1<<31) && b == -1 {
		interp.Registers[rD] = uint64(a)
	} else {
		interp.Registers[rD] = uint64(a / b)
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = 0x%x", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 195
func instRemU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	bMod32 := uint32(interp.Registers[rB])
	aMod32 := uint32(interp.Registers[rA])

	var err error
	if bMod32 == 0 {
		interp.Registers[rD], err = SignExtend(4, uint64(aMod32))
	} else {
		interp.Registers[rD], err = SignExtend(4, uint64(aMod32%bMod32))
	}
	if err != nil {
		pvmLogger.Errorf("instRemU32 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 196
func instRemS32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst

	a := int64(int32(interp.Registers[rA]))
	b := int64(int32(interp.Registers[rB]))

	if a == int64(-1<<31) && b == -1 {
		interp.Registers[rD] = 0
	} else {
		interp.Registers[rD] = uint64((smod(a, b)))
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 197
func instShloL32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	shift := interp.Registers[rB] % 32
	var err error
	interp.Registers[rD], err = SignExtend(4, uint64(uint32(interp.Registers[rA]<<shift)))
	if err != nil {
		pvmLogger.Errorf("instShloL32 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = (%s << %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 198
func instShloR32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst

	modA := uint32(interp.Registers[rA])
	shift := interp.Registers[rB] % 32
	var err error
	interp.Registers[rD], err = SignExtend(4, uint64(modA>>shift))
	if err != nil {
		pvmLogger.Errorf("instShloR32 signExtend error: %v", err)
		return ExitHalt, instr.PC
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = (%s >> %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 199
func instSharR32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst

	signedA := int32(interp.Registers[rA])

	shift := interp.Registers[rB] % 32
	interp.Registers[rD] = uint64(signedA >> shift)

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(signedA), formatInt(shift), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 200
func instAdd64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] + interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s + %s) = (%s + %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 201
func instSub64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] + (^interp.Registers[rB] + 1)

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s - %s) = (%s - %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 202
func instMul64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] * interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s • %s) = (%s • %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 203
func instDivU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	if interp.Registers[rB] == 0 {
		interp.Registers[rD] = ^uint64(0) // 2^64 - 1
	} else {
		interp.Registers[rD] = interp.Registers[rA] / interp.Registers[rB]
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 204
func instDivS64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	if interp.Registers[rB] == 0 {
		interp.Registers[rD] = ^uint64(0) // 2^64 - 1
	} else if int64(interp.Registers[rA]) == -(1<<63) && int64(interp.Registers[rB]) == -1 {
		interp.Registers[rD] = interp.Registers[rA]
	} else {
		interp.Registers[rD] = uint64((int64(interp.Registers[rA]) / int64(interp.Registers[rB])))
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 205
func instRemU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	if interp.Registers[rB] == 0 {
		interp.Registers[rD] = interp.Registers[rA]
	} else {
		interp.Registers[rD] = interp.Registers[rA] % interp.Registers[rB]
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 206
func instRemS64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	if int64(interp.Registers[rA]) == -(1<<63) && int64(interp.Registers[rB]) == -1 {
		interp.Registers[rD] = 0
	} else {
		interp.Registers[rD] = uint64(smod(int64(interp.Registers[rA]), int64(interp.Registers[rB])))
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 207
func instShloL64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] << (interp.Registers[rB] % 64)

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s << %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]%64), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 208
func instShloR64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] >> (interp.Registers[rB] % 64)

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]%64), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 209
func instSharR64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = uint64(int64(interp.Registers[rA]) >> (interp.Registers[rB] % 64))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s >> %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(int64(interp.Registers[rA])), formatInt(interp.Registers[rB]%64), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 210
func instAndMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] & interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s & %s) = (%s & %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 211
func instXorMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] ^ interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s ^ %s) = (%s ^ %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 212
func instOrMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] | interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = (%s | %s) = (%s | %s) = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], RegName[rA], RegName[rB], formatInt(interp.Registers[rA]), formatInt(interp.Registers[rB]), formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 213
func instMulUpperSSMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	signedA := int64(interp.Registers[rA])
	signedB := int64(interp.Registers[rB])

	hi, _ := bits.Mul64(uint64(abs(signedA)), uint64(abs(signedB)))

	if (signedA < 0) == (signedB < 0) {
		interp.Registers[rD] = hi
	} else {
		interp.Registers[rD] = uint64(-int64(hi))
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 214
func instMulUpperUUMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	hi, _ := bits.Mul64(interp.Registers[rA], interp.Registers[rB])
	interp.Registers[rD] = hi

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 215
func instMulUpperSUMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
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

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 216
func instSetLtUMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	if interp.Registers[rA] < interp.Registers[rB] {
		interp.Registers[rD] = 1
	} else {
		interp.Registers[rD] = 0
	}

	return ExitContinue, instr.PC
}

// opcode 217
func instSetLtSSMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	if int64(interp.Registers[rA]) < int64(interp.Registers[rB]) {
		interp.Registers[rD] = 1
	} else {
		interp.Registers[rD] = 0
	}

	return ExitContinue, instr.PC
}

// opcode 218
func instCmovIzMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	if interp.Registers[rB] == 0 {
		interp.Registers[rD] = interp.Registers[rA]
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rD], RegName[rA], formatInt(interp.Registers[rD]))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rD], formatInt(interp.Registers[rD]))
	}

	return ExitContinue, instr.PC
}

// opcode 219
func instCmovNzMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	if interp.Registers[rB] != 0 {
		interp.Registers[rD] = interp.Registers[rA]
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rD], RegName[rA], formatInt(interp.Registers[rD]))
	} else {
		pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
			RegName[rD], formatInt(interp.Registers[rD]))
	}

	return ExitContinue, instr.PC
}

// opcode 220
func instRotL64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = bits.RotateLeft64(interp.Registers[rA], int(interp.Registers[rB]%64))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 221
func instRotL32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	rotated := uint64(bits.RotateLeft32(uint32(interp.Registers[rA]), int(interp.Registers[rB]%32)))
	extend, err := SignExtend(4, rotated)
	if err != nil {
		pvmLogger.Errorf("instRoTL32 signExtend error:%v", err)
		return ExitHalt, instr.PC
	}
	interp.Registers[rD] = extend

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 222
func instRotR64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = bits.RotateLeft64(interp.Registers[rA], -int(interp.Registers[rB]))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 223
func instRotR32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	rotated := uint64(bits.RotateLeft32(uint32(interp.Registers[rA]), -int(interp.Registers[rB])))
	extend, err := SignExtend(4, rotated)
	if err != nil {
		pvmLogger.Errorf("instRotR32 signExtend error:%v", err)
		return ExitHalt, instr.PC
	}
	interp.Registers[rD] = extend

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 224
func instAndInvMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] & ^interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 225
func instOrInvMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] | ^interp.Registers[rB]

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 226
func instXnorMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = ^(interp.Registers[rA] ^ interp.Registers[rB])

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 227
func instMaxMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst

	// mutation
	interp.Registers[rD] = uint64(max(int64(interp.Registers[rA]), int64(interp.Registers[rB])))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 228
func instMaxUMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst

	// mutation
	if interp.Registers[rA] > interp.Registers[rB] {
		interp.Registers[rD] = interp.Registers[rA]
	} else {
		interp.Registers[rD] = interp.Registers[rB]
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 229
func instMinMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst

	// mutation
	interp.Registers[rD] = uint64(min(int64(interp.Registers[rA]), int64(interp.Registers[rB])))

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}

// opcode 230
func instMinUMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst

	// mutation
	if interp.Registers[rA] < interp.Registers[rB] {
		interp.Registers[rD] = interp.Registers[rA]
	} else {
		interp.Registers[rD] = interp.Registers[rB]
	}

	pvmLogger.Debugf("[%d]: pc: %d, %s, %s = %s", interp.InstrCount, instr.PC, zeta[opcode(instr.Opcode)],
		RegName[rD], formatInt(interp.Registers[rD]))
	return ExitContinue, instr.PC
}
