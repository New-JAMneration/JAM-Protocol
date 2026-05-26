package PVM

import "math/bits"

type instrMetaFn func(*Interpreter, *InstrMeta) (ExitReason, ProgramCounter)

func instrMetaExecForOpcode(op byte) instrMetaFn {
	switch op {
	case 0:
		return instTrapMeta
	case 1:
		return instFallthroughMeta
	case 10:
		return instEcalliMeta
	case 20:
		return instLoadImm64Meta
	case 30:
		return instStoreImmU8Meta
	case 31:
		return instStoreImmU16Meta
	case 32:
		return instStoreImmU32Meta
	case 33:
		return instStoreImmU64Meta
	case 40:
		return instJumpMeta
	case 50:
		return instJumpIndMeta
	case 51:
		return instLoadImmMeta
	case 52:
		return instLoadU8Meta
	case 53:
		return instLoadI8Meta
	case 54:
		return instLoadU16Meta
	case 55:
		return instLoadI16Meta
	case 56:
		return instLoadU32Meta
	case 57:
		return instLoadI32Meta
	case 58:
		return instLoadU64Meta
	case 59:
		return instStoreU8Meta
	case 60:
		return instStoreU16Meta
	case 61:
		return instStoreU32Meta
	case 62:
		return instStoreU64Meta
	case 70:
		return instStoreImmIndU8Meta
	case 71:
		return instStoreImmIndU16Meta
	case 72:
		return instStoreImmIndU32Meta
	case 73:
		return instStoreImmIndU64Meta
	case 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90:
		return instImmediateBranchMeta
	case 100:
		return instMoveRegMeta
	case 101:
		return instSbrkMeta
	case 102:
		return instCountSetBits64Meta
	case 103:
		return instCountSetBits32Meta
	case 104:
		return instLeadingZeroBits64Meta
	case 105:
		return instLeadingZeroBits32Meta
	case 106:
		return instTrailZeroBits64Meta
	case 107:
		return instTrailZeroBits32Meta
	case 108:
		return instSignExtend8Meta
	case 109:
		return instSignExtend16Meta
	case 110:
		return instZeroExtend16Meta
	case 111:
		return instReverseBytesMeta
	case 120:
		return instStoreIndU8Meta
	case 121:
		return instStoreIndU16Meta
	case 122:
		return instStoreIndU32Meta
	case 123:
		return instStoreIndU64Meta
	case 124:
		return instLoadIndU8Meta
	case 125:
		return instLoadIndI8Meta
	case 126:
		return instLoadIndU16Meta
	case 127:
		return instLoadIndI16Meta
	case 128:
		return instLoadIndU32Meta
	case 129:
		return instLoadIndI32Meta
	case 130:
		return instLoadIndU64Meta
	case 131:
		return instAddImm32Meta
	case 132:
		return instAndImmMeta
	case 133:
		return instXORImmMeta
	case 134:
		return instORImmMeta
	case 135:
		return instMulImm32Meta
	case 136:
		return instSetLtUImmMeta
	case 137:
		return instSetLtSImmMeta
	case 138:
		return instShloLImm32Meta
	case 139:
		return instShloRImm32Meta
	case 140:
		return instSharRImm32Meta
	case 141:
		return instNegAddImm32Meta
	case 142:
		return instSetGtUImmMeta
	case 143:
		return instSetGtSImmMeta
	case 144:
		return instShloLImmAlt32Meta
	case 145:
		return instShloRImmAlt32Meta
	case 146:
		return instSharRImmAlt32Meta
	case 147:
		return instCmovIzImmMeta
	case 148:
		return instCmovNzImmMeta
	case 149:
		return instAddImm64Meta
	case 150:
		return instMulImm64Meta
	case 151:
		return instShloLImm64Meta
	case 152:
		return instShloRImm64Meta
	case 153:
		return instSharRImm64Meta
	case 154:
		return instNegAddImm64Meta
	case 155:
		return instShloLImmAlt64Meta
	case 156:
		return instShloRImmAlt64Meta
	case 157:
		return instSharRImmAlt64Meta
	case 158:
		return instRotR64ImmMeta
	case 159:
		return instRotR64ImmAltMeta
	case 160:
		return instRotR32ImmMeta
	case 161:
		return instRotR32ImmAltMeta
	case 170, 171, 172, 173, 174, 175:
		return instBranchMeta
	case 180:
		return instLoadImmJumpIndMeta
	case 190:
		return instAdd32Meta
	case 191:
		return instSub32Meta
	case 192:
		return instMul32Meta
	case 193:
		return instDivU32Meta
	case 194:
		return instDivS32Meta
	case 195:
		return instRemU32Meta
	case 196:
		return instRemS32Meta
	case 197:
		return instShloL32Meta
	case 198:
		return instShloR32Meta
	case 199:
		return instSharR32Meta
	case 200:
		return instAdd64Meta
	case 201:
		return instSub64Meta
	case 202:
		return instMul64Meta
	case 203:
		return instDivU64Meta
	case 204:
		return instDivS64Meta
	case 205:
		return instRemU64Meta
	case 206:
		return instRemS64Meta
	case 207:
		return instShloL64Meta
	case 208:
		return instShloR64Meta
	case 209:
		return instSharR64Meta
	case 210:
		return instAndMeta
	case 211:
		return instXorMeta
	case 212:
		return instOrMeta
	case 213:
		return instMulUpperSSMeta
	case 214:
		return instMulUpperUUMeta
	case 215:
		return instMulUpperSUMeta
	case 216:
		return instSetLtUMeta
	case 217:
		return instSetLtSSMeta
	case 218:
		return instCmovIzMeta
	case 219:
		return instCmovNzMeta
	case 220:
		return instRotL64Meta
	case 221:
		return instRotL32Meta
	case 222:
		return instRotR64Meta
	case 223:
		return instRotR32Meta
	case 224:
		return instAndInvMeta
	case 225:
		return instOrInvMeta
	case 226:
		return instXnorMeta
	case 227:
		return instMaxMeta
	case 228:
		return instMaxUMeta
	case 229:
		return instMinMeta
	case 230:
		return instMinUMeta
	default:
		return instTrapMeta
	}
}

// opcode 0
func instTrapMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	return ExitPanic, instr.PC
}

// opcode 1
func instFallthroughMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	return ExitContinue, instr.PC
}

// opcode 10
func instEcalliMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	nuX := instr.Imm[0]
	return ExitHostCall | ExitReason(nuX), instr.PC
}

// opcode 20
func instLoadImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	interp.Registers[instr.Dst] = instr.Imm[0]
	return ExitContinue, instr.PC
}

// opcode 30
func instStoreImmU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vx, vy := instr.Imm[0], uint64(uint8(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 1, uint32(vx), vy)
	return exitReason, instr.PC
}

// opcode 31
func instStoreImmU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vx, vy := instr.Imm[0], uint64(uint16(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 2, uint32(vx), vy)
	return exitReason, instr.PC
}

// opcode 32
func instStoreImmU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vx, vy := instr.Imm[0], uint64(uint32(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 4, uint32(vx), vy)
	return exitReason, instr.PC
}

// opcode 33
func instStoreImmU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vx, vy := instr.Imm[0], instr.Imm[1]
	exitReason := storeIntoMemory(interp, 8, uint32(vx), vy)
	return exitReason, instr.PC
}

// opcode 40
func instJumpMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	vX := ProgramCounter(instr.Imm[0])
	reason, newPC := branch(instr.PC, vX, true, interp.Program.Bitmasks, interp.Program.InstructionData)
	if reason != ExitContinue {
		return reason, instr.PC
	}
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
		return reason, instr.PC
	case ExitHalt:
		return reason, instr.PC
	default:
		return reason, newPC
	}
}

// opcode 51
func instLoadImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	interp.Registers[instr.Dst] = instr.Imm[0]
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
	return ExitContinue, instr.PC
}

// opcode 59
func instStoreU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX := instr.Dst, instr.Imm[0]
	exitReason := storeIntoMemory(interp, 1, uint32(vX), uint64(uint8(interp.Registers[rA])))
	return exitReason, instr.PC
}

// opcode 60
func instStoreU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX := instr.Dst, instr.Imm[0]
	exitReason := storeIntoMemory(interp, 2, uint32(vX), uint64(uint16(interp.Registers[rA])))
	return exitReason, instr.PC
}

// opcode 61
func instStoreU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX := instr.Dst, instr.Imm[0]
	exitReason := storeIntoMemory(interp, 4, uint32(vX), uint64(uint32(interp.Registers[rA])))
	return exitReason, instr.PC
}

// opcode 62
func instStoreU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX := instr.Dst, instr.Imm[0]
	exitReason := storeIntoMemory(interp, 8, uint32(vX), interp.Registers[rA])
	return exitReason, instr.PC
}

// opcode 70
func instStoreImmIndU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX, vY := instr.Src[0], instr.Imm[0], uint64(uint8(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 1, uint32(interp.Registers[rA]+vX), vY)
	return exitReason, instr.PC
}

// opcode 71
func instStoreImmIndU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX, vY := instr.Src[0], instr.Imm[0], uint64(uint16(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 2, uint32(interp.Registers[rA]+vX), vY)
	return exitReason, instr.PC
}

// opcode 72
func instStoreImmIndU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX, vY := instr.Src[0], instr.Imm[0], uint64(uint32(instr.Imm[1]))
	exitReason := storeIntoMemory(interp, 4, uint32(interp.Registers[rA]+vX), vY)
	return exitReason, instr.PC
}

// opcode 73
func instStoreImmIndU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, vX, vY := instr.Src[0], instr.Imm[0], instr.Imm[1]
	exitReason := storeIntoMemory(interp, 8, uint32(interp.Registers[rA]+vX), vY)
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
		pvmLogger.Errorf("instImmediateBranchMeta: unexpected opcode %d, expected [80, 90]", instr.Opcode)
		return ExitPanic, instr.PC
	}

	reason, newPC := branch(instr.PC, vY, branchCondition, interp.Program.Bitmasks, interp.Program.InstructionData)
	if reason != ExitContinue {
		return reason, instr.PC
	}

	return reason, newPC
}

// opcode 100
func instMoveRegMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	// mutation
	interp.Registers[rD] = interp.Registers[rA]
	return ExitContinue, instr.PC
}

// opcode 101
func instSbrkMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]

	// this reivision is according to jam-test-vector traces: Note on SBRK
	if interp.Registers[rA] == 0 {
		interp.Registers[rD] = interp.Memory.heapPointer
		return ExitContinue, instr.PC
	}

	mem := interp.Memory
	newHeapPointer := mem.heapPointer + interp.Registers[rA]
	if newHeapPointer < mem.heapPointer || newHeapPointer > mem.heapLimit {
		interp.Registers[rD] = 0
		return ExitContinue, instr.PC
	}

	nextPageBoundary := P(int(mem.heapPointer))
	if newHeapPointer > uint64(nextPageBoundary) {
		finalBoundary := P(int(newHeapPointer))
		allocateMemorySegment(mem, uint32(mem.heapPointer), uint32(finalBoundary), nil, MemoryReadWrite)
	}

	mem.heapPointer = newHeapPointer
	interp.Registers[rD] = newHeapPointer
	return ExitContinue, instr.PC
}

// opcode 102
func instCountSetBits64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	interp.Registers[rD] = uint64(bits.OnesCount64(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 103
func instCountSetBits32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	interp.Registers[rD] = uint64(bits.OnesCount32(uint32(interp.Registers[rA])))
	return ExitContinue, instr.PC
}

// opcode 104
func instLeadingZeroBits64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	interp.Registers[rD] = uint64(bits.LeadingZeros64(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 105
func instLeadingZeroBits32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	interp.Registers[rD] = uint64(bits.LeadingZeros32(uint32(interp.Registers[rA])))
	return ExitContinue, instr.PC
}

// opcode 106
func instTrailZeroBits64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	interp.Registers[rD] = uint64(bits.TrailingZeros64(interp.Registers[rA]))
	return ExitContinue, instr.PC
}

// opcode 107
func instTrailZeroBits32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	interp.Registers[rD] = uint64(bits.TrailingZeros32(uint32(interp.Registers[rA])))
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
	return ExitContinue, instr.PC
}

// opcode 110
func instZeroExtend16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	// mutation
	regA := interp.Registers[rA]
	interp.Registers[rD] = regA % (1 << 16)
	return ExitContinue, instr.PC
}

// opcode 111
func instReverseBytesMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rD, rA := instr.Dst, instr.Src[0]
	interp.Registers[rD] = bits.ReverseBytes64(interp.Registers[rA])
	return ExitContinue, instr.PC
}

// opcode 120
func instStoreIndU8Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 1
	exitReason := storeIntoMemory(interp, offset, uint32(interp.Registers[rB]+vX), uint64(uint8(interp.Registers[rA])))
	return exitReason, instr.PC
}

// opcode 121
func instStoreIndU16Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 2
	exitReason := storeIntoMemory(interp, offset, uint32(interp.Registers[rB]+vX), uint64(uint16(interp.Registers[rA])))
	return exitReason, instr.PC
}

// opcode 122
func instStoreIndU32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 4
	exitReason := storeIntoMemory(interp, offset, uint32(interp.Registers[rB]+vX), uint64(uint32(interp.Registers[rA])))
	return exitReason, instr.PC
}

// opcode 123
func instStoreIndU64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	offset := 8
	exitReason := storeIntoMemory(interp, offset, uint32(interp.Registers[rB]+vX), uint64(interp.Registers[rA]))

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
	return ExitContinue, instr.PC
}

// opcode 132
func instAndImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] & vX
	return ExitContinue, instr.PC
}

// opcode 133
func instXORImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] ^ vX
	return ExitContinue, instr.PC
}

// opcode 134
func instORImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] | vX
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
	return ExitContinue, instr.PC
}

// opcode 140
func instSharRImm32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	vX = vX & 31 // % 32
	interp.Registers[rA] = uint64(int32(interp.Registers[rB]) >> vX)
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
	return ExitContinue, instr.PC
}

// opcode 146
func instSharRImmAlt32Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	imm := uint64(int32(uint32(vX)) >> (interp.Registers[rB] & 31))
	interp.Registers[rA] = imm
	return ExitContinue, instr.PC
}

// opcode 147
func instCmovIzImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	if interp.Registers[rB] == 0 {
		interp.Registers[rA] = vX
	}

	return ExitContinue, instr.PC
}

// opcode 148
func instCmovNzImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	if interp.Registers[rB] != 0 {
		interp.Registers[rA] = vX
	}

	return ExitContinue, instr.PC
}

// opcode 149
func instAddImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] + vX
	return ExitContinue, instr.PC
}

// opcode 150
func instMulImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = interp.Registers[rB] * vX
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
	return ExitContinue, instr.PC
}

// opcode 153
func instSharRImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = uint64(int64(interp.Registers[rB]) >> (vX & 63))
	return ExitContinue, instr.PC
}

// opcode 154
func instNegAddImm64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = vX - interp.Registers[rB]
	return ExitContinue, instr.PC
}

// opcode 155
func instShloLImmAlt64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = vX << (interp.Registers[rB] & 63)
	return ExitContinue, instr.PC
}

// opcode 156
func instShloRImmAlt64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = vX >> (interp.Registers[rB] & 63)
	return ExitContinue, instr.PC
}

// opcode 157
func instSharRImmAlt64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	interp.Registers[rA] = uint64(int64(vX) >> (interp.Registers[rB] & 63))
	return ExitContinue, instr.PC
}

// opcode 158
func instRotR64ImmMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	// rotate right
	interp.Registers[rA] = bits.RotateLeft64(interp.Registers[rB], -int(vX))
	// interp.Registers[rA] = (interp.Registers[rB] >> vX) | (interp.Registers[rB] << (64 - vX))
	return ExitContinue, instr.PC
}

// opcode 159
func instRotR64ImmAltMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA, rB, vX := instr.Dst, instr.Src[0], instr.Imm[0]

	// rotate right
	interp.Registers[rA] = bits.RotateLeft64(vX, -int(interp.Registers[rB]&63))
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
	return ExitContinue, instr.PC
}

// opcode in [170, 175]
func instBranchMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	vX := ProgramCounter(instr.Imm[0])
	branchCondition := false
	switch instr.Opcode {
	case 170:
		branchCondition = interp.Registers[rA] == interp.Registers[rB]
	case 171:
		branchCondition = interp.Registers[rA] != interp.Registers[rB]
	case 172:
		branchCondition = interp.Registers[rA] < interp.Registers[rB]
	case 173:
		branchCondition = int64(interp.Registers[rA]) < int64(interp.Registers[rB])
	case 174:
		branchCondition = interp.Registers[rA] >= interp.Registers[rB]
	case 175:
		branchCondition = int64(interp.Registers[rA]) >= int64(interp.Registers[rB])
	default:
		pvmLogger.Errorf("instBranchMeta: unexpected opcode %d, expected [170, 175]", instr.Opcode)
		return ExitPanic, instr.PC
	}

	reason, newPC := branch(instr.PC, vX, branchCondition, interp.Program.Bitmasks, interp.Program.InstructionData)
	if reason != ExitContinue {
		pvmLogger.Errorf("instBranchMeta branch error at pc: %d, opcode: %s", instr.PC, zeta[opcode(instr.Opcode)])
		return ExitReason(reason), instr.PC
	}
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
		return reason, instr.PC
	case ExitHalt:
		return reason, instr.PC
	default:
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
	} else {
		var err error
		interp.Registers[rD], err = SignExtend(4, uint64(aMod32/bMod32))
		if err != nil {
			pvmLogger.Errorf("instDivU32 signExtend error: %v", err)
			return ExitHalt, instr.PC
		}
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

	return ExitContinue, instr.PC
}

// opcode 200
func instAdd64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] + interp.Registers[rB]

	return ExitContinue, instr.PC
}

// opcode 201
func instSub64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] - interp.Registers[rB]

	return ExitContinue, instr.PC
}

// opcode 202
func instMul64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] * interp.Registers[rB]

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

	return ExitContinue, instr.PC
}

// opcode 207
func instShloL64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] << (interp.Registers[rB] % 64)

	return ExitContinue, instr.PC
}

// opcode 208
func instShloR64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] >> (interp.Registers[rB] % 64)

	return ExitContinue, instr.PC
}

// opcode 209
func instSharR64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = uint64(int64(interp.Registers[rA]) >> (interp.Registers[rB] % 64))

	return ExitContinue, instr.PC
}

// opcode 210
func instAndMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] & interp.Registers[rB]

	return ExitContinue, instr.PC
}

// opcode 211
func instXorMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] ^ interp.Registers[rB]

	return ExitContinue, instr.PC
}

// opcode 212
func instOrMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] | interp.Registers[rB]

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

	return ExitContinue, instr.PC
}

// opcode 222
func instRotR64Meta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = bits.RotateLeft64(interp.Registers[rA], -int(interp.Registers[rB]))

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

	return ExitContinue, instr.PC
}

// opcode 224
func instAndInvMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] & ^interp.Registers[rB]

	return ExitContinue, instr.PC
}

// opcode 225
func instOrInvMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = interp.Registers[rA] | ^interp.Registers[rB]

	return ExitContinue, instr.PC
}

// opcode 226
func instXnorMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst
	// mutation
	interp.Registers[rD] = ^(interp.Registers[rA] ^ interp.Registers[rB])

	return ExitContinue, instr.PC
}

// opcode 227
func instMaxMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst

	// mutation
	interp.Registers[rD] = uint64(max(int64(interp.Registers[rA]), int64(interp.Registers[rB])))

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

	return ExitContinue, instr.PC
}

// opcode 229
func instMinMeta(interp *Interpreter, instr *InstrMeta) (ExitReason, ProgramCounter) {
	rA := instr.Src[0]
	rB := instr.Src[1]
	rD := instr.Dst

	// mutation
	interp.Registers[rD] = uint64(min(int64(interp.Registers[rA]), int64(interp.Registers[rB])))

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

	return ExitContinue, instr.PC
}
