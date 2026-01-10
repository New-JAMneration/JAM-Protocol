package PVM

// A.52
func (b *BlockState) SimulatePipeline(program *Program, pc ProgramCounter) Gas {
	for program.ValidateOpcode(pc); ; {
		// Ξ'(n+1)
		blockEnd := (b.Step != 0 && program.Bitmasks.IsStartOfInstruction(int(b.Iota)))
		opcodeIdx := program.InstructionData[b.Iota]
		opcodeResource := getOpcodeResource(opcode(opcodeIdx))
		if b.Step == 0 || (!blockEnd && opcodeResource.decodeSlotsNeeded <= b.DecodeSlots && len(b.S) < 32) {
			b.decodeInstr(program, pc)
		}

		// Ξ''(n+1)
		ready, _ := b.checkROBReady()
		if ready && b.ExecutionSlots > 0 {
			b.retire()
		}

		// ∅
		if blockEnd && len(b.S) == 0 {
			return max(b.Cyc-3, 1) // A.53
		}

		// Ξ'''(n+1)
		b.Advance()
	}
}

// A.54
func (b *BlockState) decodeInstr(program *Program, pc ProgramCounter) {
	opcodeIdx := program.InstructionData[b.Iota]
	if opcodeIdx == 100 { // move_reg
		b.decodeNoDispatch(program, pc)
	} else {
		b.decodeDispatchToROB(program, pc)
	}

}

// A.55 move_reg
func (b *BlockState) decodeNoDispatch(program *Program, pc ProgramCounter) {
	targetPC := b.Iota
	b.Iota += ProgramCounter(skip(int(pc), program.Bitmasks)) + 1
	b.DecodeSlots -= 1
	operands := getOperands(program, targetPC)

	for j := 0; j < len(b.R); j++ {
		for k := 0; k < len(b.R[j]); k++ {
			if b.R[j][k] == Reg(operands.srcReg[0]) { // move_reg only one srcRg
				b.R.remove2D(j, k)
				return
			}
		}
	}
	// no dependency found
	b.R[len(b.R)] = make([]Reg, 1)
	b.R[len(b.R)][0] = Reg(operands.srcReg[0])
}

// A.56 decode
func (b *BlockState) decodeDispatchToROB(program *Program, pc ProgramCounter) {
	targetPC := b.Iota
	opcodeResource := getOpcodeResource(opcode(program.InstructionData[targetPC]))
	// iota
	b.Iota += ProgramCounter(skip(int(pc), program.Bitmasks)) + 1
	// d
	b.DecodeSlots -= opcodeResource.decodeSlotsNeeded
	// n
	b.Next += 1
	// s->
	// the ROB length is limited to MaxROB, so no need to check using append
	b.S = append(b.S, 1) // decoded
	// c->
	if len(b.C) >= b.Next {
		b.C[b.Next-1] = opcodeResource.cyclesNeeded
	}
	// x->
	if len(b.X) >= b.Next {
		b.X[b.Next-1] = opcodeResource.execUnitsNeeded
	}
	// p->
	operands := getOperands(program, targetPC)

	b.P = append(b.P, operands.dstReg)

	dstRegMap := make(map[Reg]bool, len(operands.dstReg))
	for _, v := range operands.dstReg {
		dstRegMap[v] = true
	}
	// r->
	for j := 0; j < len(b.P); j++ {
		if j == b.Step {
			b.R = append(b.R, operands.dstReg)
		} else {
			for k := 0; k < len(b.P[j]); k++ {
				if _, ok := dstRegMap[b.R[j][k]]; ok {
					b.R.remove2D(j, k)
				}
			}
		}
	}
}

// A.57 retire instruction in ROB
func (b *BlockState) retire() {
	_, robIndex := b.checkROBReady()
	b.S[robIndex] = 3 // executing

	b.UnitsAvail.A -= b.X[robIndex].A
	b.UnitsAvail.L -= b.X[robIndex].L
	b.UnitsAvail.S -= b.X[robIndex].S
	b.UnitsAvail.M -= b.X[robIndex].M
	b.UnitsAvail.D -= b.X[robIndex].D

	b.ExecutionSlots -= 1

}

// A.59
func (b *BlockState) Advance() {
	// most of the state depend on S->
	// only need to iterate k+1 to j, since s->(0), ... s->(k) == 4 (finish)
	// update S comes last, so that will not affect the order of S
	var k int
	for j := len(b.S) - 1; j >= 0; j-- {
		if b.S[j] == 4 {
			k = j
			break
		}
	}

	// update C-> , R->, X
	for j := k + 1; j < len(b.C); j++ {
		if b.S[j] != 3 {
			continue
		}
		b.C[j] -= 1      // update C->
		if b.C[j] == 1 { // about to finish
			// update R, remove from ROB
			b.R.remove1D(j)

			// update X
			b.UnitsAvail.A += b.X[j].A
			b.UnitsAvail.D += b.X[j].D
			b.UnitsAvail.L += b.X[j].L
			b.UnitsAvail.M += b.X[j].M
			b.UnitsAvail.S += b.X[j].S
		}
	}

	b.Cyc += 1
	b.DecodeSlots = 4
	b.ExecutionSlots = 5

}

// A.58
func (b *BlockState) checkROBReady() (bool, int) {
	for j := 0; j < len(b.S); j++ {
		if b.S[j] != 2 { //  b.S[i] == 2 => pending
			continue
		}
		// check execute units availability
		if b.X[j].A > b.UnitsAvail.A {
			continue
		}
		if b.X[j].L > b.UnitsAvail.L {
			continue
		}
		if b.X[j].S > b.UnitsAvail.S {
			continue
		}
		if b.X[j].M > b.UnitsAvail.M {
			continue
		}
		if b.X[j].D > b.UnitsAvail.D {
			continue
		}

		// check dependencies
		execCycleSet := make(map[int]int, len(b.C))
		for k := 0; k < len(b.C); k++ {
			execCycleSet[k] = b.C[k]
		}

		for k := 0; k < len(b.P[j]); k++ {
			if _, ok := execCycleSet[int(b.P[j][k])]; ok {
				return false, -1
			}
		}
		return true, j
	}
	return false, -1
}

type opcodeResource struct {
	cyclesNeeded      int
	decodeSlotsNeeded int
	execUnitsNeeded   ExecUnits
}

type operand struct {
	srcReg []Reg
	dstReg []Reg
}

func getOperands(program *Program, pc ProgramCounter) operand {
	opcodeIdx := program.InstructionData[pc]
	// According to A.5
	switch {
	// no reg needed
	case (opcodeIdx >= 0 && opcodeIdx <= 2), // A.5.1
		opcodeIdx == 10,                      // A.5.2
		(opcodeIdx >= 30 && opcodeIdx <= 33), // A.5.4
		opcodeIdx == 40:                      // A.5.5
		return operand{nil, nil}

	// one dstReg
	case opcodeIdx == 20, // A.5.3
		(opcodeIdx >= 51 && opcodeIdx <= 58), // A.5.6
		(opcodeIdx >= 70 && opcodeIdx <= 73), // A.5.7
		opcodeIdx == 80:                      // A.5.8
		rA := min(12, (program.InstructionData[pc+1])%16)
		return operand{nil, []Reg{Reg(rA)}}

	// one srcReg
	case opcodeIdx == 50, (opcodeIdx >= 59 && opcodeIdx <= 62), // A.5.6
		(opcodeIdx >= 81 && opcodeIdx <= 90): // A.5.8
		rA := min(12, (program.InstructionData[pc+1])%16)
		return operand{[]Reg{Reg(rA)}, nil}

	// one srcReg, one dstReg,
	case (opcodeIdx >= 100 && opcodeIdx <= 110), // A.5.9
		(opcodeIdx >= 124 && opcodeIdx <= 161), // A.5.10
		opcodeIdx == 180:                       // A.5.12   rA => rD (dstReg), rB => rA (srcReg)
		rD := min(12, (program.InstructionData[pc+1])%16)
		rA := min(12, (program.InstructionData[pc+1])>>4)
		return operand{[]Reg{Reg(rA)}, []Reg{Reg(rD)}}

	// one srcReg, one dstRg (different srcReg, dstReg computation)
	case (opcodeIdx >= 120 && opcodeIdx <= 123): // A.5.10
		rA := min(12, (program.InstructionData[pc+1])%16)
		rB := min(12, (program.InstructionData[pc+1])>>4)
		return operand{[]Reg{Reg(rA)}, []Reg{Reg(rB)}}

	// Two srcReg
	case (opcodeIdx >= 170 && opcodeIdx <= 175): // A.5.11
		rA := min(12, (program.InstructionData[pc+1])%16)
		rB := min(12, (program.InstructionData[pc+1])>>4)
		return operand{[]Reg{Reg(rA), Reg(rB)}, nil}

		// Three regs: one dstReg, two srcReg
	case opcodeIdx >= 190 && opcodeIdx <= 230:
		rA := min(12, (program.InstructionData[pc+1])%16)
		rB := min(12, (program.InstructionData[pc+1])>>4)
		rD := min(12, program.InstructionData[pc+2])
		return operand{[]Reg{Reg(rA), Reg(rB)}, []Reg{Reg(rD)}}

	default:
		// this will never be reached in current implementation, opcode is checked before
		return operand{nil, nil}
	}
}

// A.57
// TODO: implement according to actual gas model table
func getOpcodeResource(opcodeData opcode) opcodeResource {
	switch opcodeData {
	case 0:
		return opcodeResource{2, 2, ExecUnits{A: 1, L: 0, S: 0, M: 0, D: 0}}
	case 1:
		return opcodeResource{2, 2, ExecUnits{A: 1, L: 0, S: 0, M: 0, D: 0}}
	default:
		return opcodeResource{1, 1, ExecUnits{A: 1, L: 0, S: 0, M: 0, D: 0}}
	}
}
