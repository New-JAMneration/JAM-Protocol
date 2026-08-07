package PVM

// memoryAccessCycles (𝔐) is the per-access cycle cost for load/store in A.10.
const memoryAccessCycles = 25

// InstrCost holds A.10 cost_cycles, cost_decodeslots, and cost_execunits.
// Table column order is (cycles, decode_slots, ALU, LOAD, STORE, MUL, DIV).
// 𝔓(a,b) / 𝔓_S(a,b) select decode_slots for ALU ops (cycles are fixed).
type InstrCost struct {
	Cycles int
	Decode int
	Units  ExecUnits
}

// selectOverlapCost implements 𝔓(a,b) for the decode-slots column:
// a when dst∩src ≠ ∅, else b.
func selectOverlapCost(a, b int, instr *InstrMeta) int {
	if regsOverlap(instSrcRegs(instr), instDstRegs(instr)) {
		return a
	}
	return b
}

// selectOverlapShiftCost implements 𝔓_S(a,b) for shift/rotate decode slots.
func selectOverlapShiftCost(a, b int, instr *InstrMeta) int {
	if instr.Src[0] != 0xFF && instr.Dst != 0xFF && instr.Src[0] == instr.Dst {
		return a
	}
	return b
}

func isTrapOrUnlikely(p *Program, pc int) bool {
	// eq:instructions — code is zero-padded; beyond |c| the opcode is trap.
	if pc < 0 {
		return false
	}
	if pc >= len(p.InstructionData) {
		return true
	}
	op := p.InstructionData[pc]
	return op == 0 || op == 2
}

// branchCycles is A.10 cost_cycles for branch / branch_imm (1 or 20).
func branchCycles(p *Program, instr *InstrMeta) int {
	pc := int(instr.PC)
	fallthroughPC := pc + 1 + int(instr.SkipLen)
	targetPC := int(instr.Imm[0])
	if opcodeInfoTable[instr.Opcode].Category == InstrCatOneRegImmOff {
		targetPC = int(instr.Imm[1])
	}
	if isTrapOrUnlikely(p, fallthroughPC) || isTrapOrUnlikely(p, targetPC) {
		return 1
	}
	return 20
}

func unitsALU() ExecUnits { return ExecUnits{A: 1} }

// Load/store/mul/div also take an ALU unit (GP A.10 cost_execunits table).
func unitsLoad() ExecUnits  { return ExecUnits{A: 1, L: 1} }
func unitsStore() ExecUnits { return ExecUnits{A: 1, S: 1} }
func unitsMul() ExecUnits   { return ExecUnits{A: 1, M: 1} }
func unitsDiv() ExecUnits   { return ExecUnits{A: 1, D: 1} }

// A.10 \simplealuthreeop: cycles=1, decode=𝔓(1,2)
func simpleAluThreeOp(instr *InstrMeta) InstrCost {
	return InstrCost{Cycles: 1, Decode: selectOverlapCost(1, 2, instr), Units: unitsALU()}
}

// A.10 \simplealuthreeopthirtytwo: cycles=2, decode=𝔓(2,3)
func simpleAluThreeOp32(instr *InstrMeta) InstrCost {
	return InstrCost{Cycles: 2, Decode: selectOverlapCost(2, 3, instr), Units: unitsALU()}
}

// A.10 \simplealutwoop: cycles=1, decode=𝔓(1,2)
func simpleAluTwoOp(instr *InstrMeta) InstrCost {
	return InstrCost{Cycles: 1, Decode: selectOverlapCost(1, 2, instr), Units: unitsALU()}
}

// A.10 \simplealutwoopthirtytwo: cycles=2, decode=𝔓(2,3)
func simpleAluTwoOp32(instr *InstrMeta) InstrCost {
	return InstrCost{Cycles: 2, Decode: selectOverlapCost(2, 3, instr), Units: unitsALU()}
}

// A.10 \trivialtwooponecycle
func trivialTwoOpOneCycle() InstrCost {
	return InstrCost{Cycles: 1, Decode: 1, Units: unitsALU()}
}

// A.10 \trivialtwooptwocycles
func trivialTwoOpTwoCycles() InstrCost {
	return InstrCost{Cycles: 2, Decode: 1, Units: ExecUnits{A: 2}}
}

// A.10 \shiftsandrotates: cycles=1, decode=𝔓_S(2,3)
func shiftsAndRotates(instr *InstrMeta) InstrCost {
	return InstrCost{Cycles: 1, Decode: selectOverlapShiftCost(2, 3, instr), Units: unitsALU()}
}

// A.10 \shiftsandrotatesthirtytwo: cycles=2, decode=𝔓_S(3,4)
func shiftsAndRotates32(instr *InstrMeta) InstrCost {
	return InstrCost{Cycles: 2, Decode: selectOverlapShiftCost(3, 4, instr), Units: unitsALU()}
}

// A.10 \finish{n}
func finishCost(cycles int) InstrCost {
	return InstrCost{Cycles: cycles, Decode: 1, Units: ExecUnits{}}
}

// memLoadCost is A.10 \directload / \indirectload (identical cost rows).
func memLoadCost() InstrCost {
	return InstrCost{Cycles: memoryAccessCycles, Decode: 1, Units: unitsLoad()}
}

// memStoreCost is A.10 \directstore / \indirectstore / \storeimm / \indirectstoreimm.
func memStoreCost() InstrCost {
	return InstrCost{Cycles: memoryAccessCycles, Decode: 1, Units: unitsStore()}
}

// InstructionCost returns A.10 cost for a pre-decoded instruction.
// Opcode numbers follow PVM/instructions.go name table.
func InstructionCost(p *Program, instr *InstrMeta) InstrCost {
	switch instr.Opcode {
	case 0, 1: // trap, fallthrough
		return finishCost(2)
	case 2: // unlikely
		return InstrCost{Cycles: 40, Decode: 1, Units: ExecUnits{}}
	case 10: // ecalli
		return InstrCost{Cycles: 100, Decode: 4, Units: unitsALU()}
	case 20: // load_imm_64
		return InstrCost{Cycles: 1, Decode: 2, Units: ExecUnits{}}

	case 30, 31, 32, 33: // store_imm_u*
		return memStoreCost()
	case 40: // jump
		return finishCost(15)
	case 50: // jump_ind
		return InstrCost{Cycles: 22, Decode: 1, Units: ExecUnits{}}
	case 51: // load_imm
		return InstrCost{Cycles: 1, Decode: 1, Units: ExecUnits{}}
	case 52, 53, 54, 55, 56, 57, 58: // load_u*/i*
		return memLoadCost()
	case 59, 60, 61, 62: // store_u*
		return memStoreCost()
	case 70, 71, 72, 73: // store_imm_ind_u*
		return memStoreCost()
	case 80: // load_imm_jump
		return finishCost(15)
	case 81, 82, 83, 84, 85, 86, 87, 88, 89, 90: // branch_*_imm
		return InstrCost{Cycles: branchCycles(p, instr), Decode: 1, Units: unitsALU()}
	case 100: // move_reg
		return InstrCost{Cycles: 0, Decode: 1, Units: ExecUnits{}}

	case 101, 102, 103, 104, 107, 108, 109: // count/leading/sign/zero extend
		return trivialTwoOpOneCycle()
	case 105, 106: // trailing_zero_bits_*
		return trivialTwoOpTwoCycles()
	case 110: // reverse_bytes
		return simpleAluTwoOp(instr)

	case 120, 121, 122, 123: // store_ind_u*
		return memStoreCost()
	case 124, 125, 126, 127, 128, 129, 130: // load_ind_*
		return memLoadCost()

	case 131: // add_imm_32
		return simpleAluTwoOp32(instr)
	case 132, 133, 134: // and/xor/or_imm
		return simpleAluTwoOp(instr)
	case 135: // mul_imm_32
		return InstrCost{Cycles: 4, Decode: selectOverlapCost(2, 3, instr), Units: unitsMul()}
	case 136, 137, 142, 143: // set_lt/gt_*_imm
		return InstrCost{Cycles: 3, Decode: 3, Units: unitsALU()}
	case 138, 139, 140: // sh*_imm_32
		return simpleAluTwoOp32(instr)
	case 141: // neg_add_imm_32
		return InstrCost{Cycles: 3, Decode: 4, Units: unitsALU()}
	case 144, 145, 146: // sh*_imm_alt_32
		return InstrCost{Cycles: 2, Decode: 4, Units: unitsALU()}
	case 147, 148: // cmov_*_imm
		return InstrCost{Cycles: 2, Decode: 3, Units: unitsALU()}
	case 149: // add_imm_64
		return simpleAluTwoOp(instr)
	case 150: // mul_imm_64
		return InstrCost{Cycles: 3, Decode: selectOverlapCost(1, 2, instr), Units: unitsMul()}
	case 151, 152, 153, 158: // sh*_imm_64, rot_r_64_imm
		return simpleAluTwoOp(instr)
	case 154: // neg_add_imm_64
		return InstrCost{Cycles: 2, Decode: 3, Units: unitsALU()}
	case 155, 156, 157, 159: // sh*_imm_alt_64, rot_r_64_imm_alt
		return InstrCost{Cycles: 1, Decode: 3, Units: unitsALU()}
	case 160: // rot_r_32_imm
		return simpleAluTwoOp32(instr)
	case 161: // rot_r_32_imm_alt
		return InstrCost{Cycles: 2, Decode: 4, Units: unitsALU()}

	case 170, 171, 172, 173, 174, 175: // branch_*
		return InstrCost{Cycles: branchCycles(p, instr), Decode: 1, Units: unitsALU()}
	case 180: // load_imm_jump_ind
		return InstrCost{Cycles: 22, Decode: 1, Units: ExecUnits{}}

	case 190, 191: // add/sub_32
		return simpleAluThreeOp32(instr)
	case 192: // mul_32
		return InstrCost{Cycles: 4, Decode: selectOverlapCost(2, 3, instr), Units: unitsMul()}
	case 193, 194, 195, 196: // div/rem_32
		return InstrCost{Cycles: 60, Decode: 4, Units: unitsDiv()}
	case 197, 198, 199: // sh*_32
		return shiftsAndRotates32(instr)
	case 200, 201: // add/sub_64
		return simpleAluThreeOp(instr)
	case 202: // mul_64
		return InstrCost{Cycles: 3, Decode: selectOverlapCost(1, 2, instr), Units: unitsMul()}
	case 203, 204, 205, 206: // div/rem_64
		return InstrCost{Cycles: 60, Decode: 4, Units: unitsDiv()}
	case 207, 208, 209: // sh*_64
		return shiftsAndRotates(instr)
	case 210, 211, 212: // and/xor/or
		return simpleAluThreeOp(instr)
	case 213, 214: // mul_upper_s_s / u_u
		return InstrCost{Cycles: 4, Decode: 4, Units: unitsMul()}
	case 215: // mul_upper_s_u
		return InstrCost{Cycles: 6, Decode: 4, Units: unitsMul()}
	case 216, 217: // set_lt_*
		return InstrCost{Cycles: 3, Decode: 3, Units: unitsALU()}
	case 218, 219: // cmov_*
		return InstrCost{Cycles: 2, Decode: 2, Units: unitsALU()}
	case 220, 222: // rot_*_64
		return shiftsAndRotates(instr)
	case 221, 223: // rot_*_32
		return shiftsAndRotates32(instr)
	case 224, 225: // and_inv / or_inv
		return InstrCost{Cycles: 2, Decode: 3, Units: unitsALU()}
	case 226: // xnor
		return InstrCost{Cycles: 2, Decode: selectOverlapCost(2, 3, instr), Units: unitsALU()}
	case 227, 228, 229, 230: // min/max
		return InstrCost{Cycles: 3, Decode: selectOverlapCost(2, 3, instr), Units: unitsALU()}

	default:
		return InstrCost{Cycles: 1, Decode: 1, Units: unitsALU()}
	}
}
