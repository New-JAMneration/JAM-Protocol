package PVM

// SingleStepInvoke is kept for GP 0.7.2; deprecated for v0.8.0 and later.
// Used by refine host-call → invoke host-call transition (per-instruction gas).
func (interp *Interpreter) SingleStepInvoke(pc ProgramCounter) (ExitReason, ProgramCounter) {
	for {
		exitReason, pcPrime := interp.SingleStepStateTransition(pc)
		switch exitReason.GetReasonType() {
		case CONTINUE:
			pc = pcPrime
			continue
		case HALT, PANIC:
			return exitReason, 0
		default: // HOST_CALL, OUT_OF_GAS
			return exitReason, pcPrime
		}
	}
}

// SingleStepStateTransition is kept for GP 0.7.2; deprecated for v0.8.0 and later.
// Per-instruction gas charging (1 gas per instruction).
func (interp *Interpreter) SingleStepStateTransition(pc ProgramCounter) (ExitReason, ProgramCounter) {
	// check program-counter exceed blob length
	if int(pc) >= len(interp.Program.InstructionData) {
		return ExitPanic, pc
	}

	var exitReason ExitReason

	// (v0.7.1  A.19) check opcode validity
	opcodeData := interp.Program.InstructionData.isOpcode(pc)
	// (GP A.6) OOG when ρ < 1 (gas insufficient for next instruction)
	if interp.Gas < 1 {
		return ExitOOG, pc
	}
	interp.Gas -= 1

	target := execInstructions[opcodeData]
	if target == nil {
		pvmLogger.Errorf("instruction not implemented")
		return ExitPanic, pc
	}
	// (v0.7.1  A.20) l = skip(iota)
	skipLength := ProgramCounter(skip(int(pc), interp.Program.Bitmasks))

	exitReason, newPC := execInstructions[opcodeData](interp, pc, skipLength) // update PVM states

	reason := exitReason.GetReasonType()
	switch reason {
	case PANIC, HALT:
		// pvmLogger.Debugf("   gas: %d", interp.Gas)
		return exitReason, 0
	case HOST_CALL: // host-call: newPC = pc
		return exitReason, newPC
	}

	if pc != newPC {
		// execute branch instruction
		return exitReason, newPC
	}

	// iota' = iota + 1 +skip(iota)
	newPC += skipLength + 1
	return exitReason, newPC
}

// SingleStepInvokeDecodedBlocks is kept for GP 0.7.2; deprecated for v0.8.0 and later.
// Pre-decoded block execution with per-instruction gas charging.
func (interp *Interpreter) SingleStepInvokeDecodedBlocks(pc ProgramCounter) (ExitReason, ProgramCounter) {
	prog := interp.Program
	instrSlice := prog.Instrs
	n := len(prog.BlockAt)

	for {
		if int(pc) >= n {
			return ExitPanic, 0
		}

		var startIdx, endIdx int
		if block := prog.BlockAt[pc]; block != nil {
			startIdx = block.InstrStart
			endIdx = block.InstrEnd
		} else if idx := prog.InstrIdxAt[pc]; idx >= 0 {
			startIdx = int(idx)
			foundTerminator := false
			for endIdx = startIdx; endIdx < len(instrSlice); endIdx++ {
				if IsBlockTerminator(instrSlice[endIdx].Opcode) {
					endIdx++
					foundTerminator = true
					break
				}
			}
			if !foundTerminator {
				return ExitPanic, 0
			}
		} else {
			return ExitPanic, 0
		}

		instrs := instrSlice[startIdx:endIdx]
		branchTaken := false
		for i := range instrs {
			instr := &instrs[i]

			// (GP A.6) OOG when ρ < 1 (gas insufficient for next instruction)
			if interp.Gas < 1 {
				return ExitOOG, instr.PC
			}
			interp.Gas -= 1

			var src1Val, src2Val uint64
			if instr.Src[0] != 0xff {
				src1Val = interp.Registers[instr.Src[0]]
			}
			if instr.Src[1] != 0xff {
				src2Val = interp.Registers[instr.Src[1]]
			}

			exitReason, newPC := instr.Exec(interp, instr)
			interp.recordInstrTraceStepAfterMeta(instr, src1Val, src2Val)

			switch exitReason.GetReasonType() {
			case HALT, PANIC:
				return exitReason, 0
			case PAGE_FAULT, OUT_OF_GAS:
				return exitReason, instr.PC
			case HOST_CALL:
				return exitReason, instr.PC + ProgramCounter(instr.SkipLen) + 1
			}

			if instr.PC != newPC {
				pc = newPC
				branchTaken = true
				break
			}
		}
		if !branchTaken {
			last := &instrs[len(instrs)-1]
			pc = last.PC + ProgramCounter(last.SkipLen) + 1
		}
	}
}

// BlockBasedInvoke: runtime DecodeInstructionBlock + A.9 block gas via pre-decoded
// BlockMeta (tests / legacy). Requires preDecodeBlocks on prog.
// Production: outer Ψ_M → BlockBasedInvokeDecodedBlocks; refine Ω_K invoke → same.
func (interp *Interpreter) BlockBasedInvoke(pc ProgramCounter) (ExitReason, ProgramCounter) {
	prog := interp.Program
	if prog == nil {
		return ExitPanic, 0
	}

	for {
		blockStart, ok := containingBlockStart(pc, prog.Bitmasks)
		if !ok {
			return ExitPanic, 0
		}

		pcPrime, _, exitReason := DecodeInstructionBlock(prog.InstructionData, blockStart, prog.Bitmasks)
		if exitReason.GetReasonType() != CONTINUE {
			pvmLogger.Errorf("DecodeInstructionBlock error : %v", exitReason)
			return exitReason, 0
		}

		block := prog.BlockContaining(pc)
		if block == nil {
			return ExitPanic, 0
		}

		if !interp.GasCharged {
			cost := blockGasAtPC(prog, pc, block)
			if interp.Gas < cost {
				return ExitOOG, pc
			}
			interp.Gas -= cost
			interp.GasCharged = true
		}

		pc, exitReason = interp.ExecuteInstructions(pc, pcPrime)
		switch exitReason.GetReasonType() {
		case PANIC, HALT:
			return exitReason, 0
		case HOST_CALL, OUT_OF_GAS, PAGE_FAULT:
			return exitReason, pc
		case CONTINUE:
			// fallthrough to next block
		default:
			return ExitPanic, 0
		}
	}
}

// containingBlockStart is 𝔏(pc) via the bitmask: nearest basic-block start
// at or before pc. pc itself must be an instruction start.
func containingBlockStart(pc ProgramCounter, bitmask Bitmask) (ProgramCounter, bool) {
	if !bitmask.IsStartOfInstruction(int(pc)) {
		return 0, false
	}
	for b := pc; ; b-- {
		if bitmask.IsStartOfBasicBlock(b) {
			return b, true
		}
		if b == 0 {
			return 0, false
		}
	}
}

// gasChargedForIntegratedResume restores integrated gaschargedflag for inner Ψ.
// A.4 keeps ⊤ across fault/halt/panic; only CONTINUE/HOST_CALL terminators clear it.
// A stored ⊤ at a block-entry PC is therefore valid (e.g. fault on the first instr).
func gasChargedForIntegratedResume(prog *Program, pc ProgramCounter, stored bool) bool {
	return stored && prog != nil && prog.ValidInstructionAt(uint64(pc))
}

// blockGasAtPC returns A.4/A.9 gas for entering at pc within block.
// Always the full containing-block cost (gascostforblock(c, k, L(ι))),
// including fresh mid-block entry with gaschargedflag = ⊥.
func blockGasAtPC(_ *Program, _ ProgramCounter, block *BlockMeta) Gas {
	return block.GasCost
}

// BlockBasedInvokeDecodedBlocks: pre-decoded blocks + A.7 gas. Production path for
// outer Ψ_M (MachineInvoke) and refine Ω_K invoke (host_call_refine).
func (interp *Interpreter) BlockBasedInvokeDecodedBlocks(pc ProgramCounter) (ExitReason, ProgramCounter) {
	prog := interp.Program
	if prog == nil {
		return ExitPanic, 0
	}

	for {
		if int(pc) >= len(prog.InstrIdxAt) {
			return ExitPanic, 0
		}

		instrIdx := prog.InstrIdxAt[pc]
		block := prog.BlockContaining(pc)
		if instrIdx < 0 || block == nil {
			return ExitPanic, 0
		}

		startIdx := int(instrIdx)
		if startIdx < block.InstrStart || startIdx >= block.InstrEnd {
			return ExitPanic, 0
		}

		if !interp.GasCharged {
			blockGas := blockGasAtPC(prog, pc, block)
			if interp.Gas < blockGas {
				return ExitOOG, pc
			}
			interp.Gas -= blockGas
			interp.GasCharged = true
		}

		instrs := prog.Instrs[startIdx:block.InstrEnd]
		branchTaken := false
		for i := range instrs {
			instr := &instrs[i]

			var src1Val, src2Val uint64
			if instr.Src[0] != 0xff {
				src1Val = interp.Registers[instr.Src[0]]
			}
			if instr.Src[1] != 0xff {
				src2Val = interp.Registers[instr.Src[1]]
			}

			exitReason, newPC := instr.Exec(interp, instr)
			interp.recordInstrTraceStepAfterMeta(instr, src1Val, src2Val)

			reason := exitReason.GetReasonType()
			if IsBlockTerminator(instr.Opcode) && (reason == CONTINUE || reason == HOST_CALL) {
				interp.GasCharged = false
			}

			switch reason {
			case PANIC, HALT:
				return exitReason, 0
			case PAGE_FAULT, OUT_OF_GAS:
				return exitReason, instr.PC
			case HOST_CALL:
				return exitReason, instr.PC + ProgramCounter(instr.SkipLen) + 1
			}

			// Terminator defines control flow even when newPC == instr.PC (self-loop).
			if IsBlockTerminator(instr.Opcode) {
				pc = newPC
				branchTaken = true
				break
			}
		}

		if !branchTaken {
			last := &instrs[len(instrs)-1]
			pc = last.PC + ProgramCounter(last.SkipLen) + 1
		}
	}
}

// DebugSingleStepInvoke runs one pre-decoded instruction per iteration with
// block-level gas pre-charge (A.7). Used by pvmtrace to emit per-instruction
// streams aligned with recompiler DebugSingleStepInvoke.
func (interp *Interpreter) DebugSingleStepInvoke(pc ProgramCounter) (ExitReason, ProgramCounter) {
	prog := interp.Program
	if prog == nil {
		return ExitPanic, 0
	}

	for {
		if int(pc) >= len(prog.InstrIdxAt) {
			return ExitPanic, 0
		}

		instrIdx := prog.InstrIdxAt[pc]
		block := prog.BlockContaining(pc)
		if instrIdx < 0 || block == nil {
			return ExitPanic, 0
		}

		startIdx := int(instrIdx)
		if startIdx < block.InstrStart || startIdx >= block.InstrEnd {
			return ExitPanic, 0
		}

		if !interp.GasCharged {
			blockGas := blockGasAtPC(prog, pc, block)
			if interp.Gas < blockGas {
				return ExitOOG, pc
			}
			interp.Gas -= blockGas
			interp.GasCharged = true
		}

		instr := &prog.Instrs[startIdx]

		var src1Val, src2Val uint64
		if instr.Src[0] != 0xff {
			src1Val = interp.Registers[instr.Src[0]]
		}
		if instr.Src[1] != 0xff {
			src2Val = interp.Registers[instr.Src[1]]
		}

		exitReason, newPC := instr.Exec(interp, instr)
		interp.recordInstrTraceStepAfterMeta(instr, src1Val, src2Val)

		reason := exitReason.GetReasonType()
		if IsBlockTerminator(instr.Opcode) && (reason == CONTINUE || reason == HOST_CALL) {
			interp.GasCharged = false
		}

		switch reason {
		case PANIC, HALT:
			return exitReason, 0
		case PAGE_FAULT, OUT_OF_GAS:
			return exitReason, instr.PC
		case HOST_CALL:
			return exitReason, instr.PC + ProgramCounter(instr.SkipLen) + 1
		}

		if IsBlockTerminator(instr.Opcode) {
			pc = newPC
			continue
		}
		pc = instr.PC + ProgramCounter(instr.SkipLen) + 1
	}
}

func DecodeInstructionBlock(instructionData ProgramCode, pc ProgramCounter, bitmask Bitmask) (ProgramCounter, int64, ExitReason) {
	pcPrime := pc
	count := int64(1)

	for {
		// check pc is not out of range and avoid infinit-loop
		if pcPrime >= ProgramCounter(len(instructionData)) {
			// pvmLogger.Debugf("PVM panic: program counter out of range, pcPrime = %d > program-length = %d", pcPrime, len(instructionData))
			return pc, 0, ExitPanic
		}

		if !IsValidOpcode(instructionData[pcPrime]) {
			return pc, 0, ExitPanic
		}

		// reach instruction block end
		if IsBlockTerminator(instructionData[pcPrime]) {
			return pcPrime, count, ExitContinue
		}
		count++
		skipLength := skip(int(pcPrime), bitmask)
		pcPrime += ProgramCounter(skipLength) + 1
	}
}

// ExecuteInstructions executes block[pc:pcPrime]. Gas is pre-charged at block entry.
func (interp *Interpreter) ExecuteInstructions(pc ProgramCounter, pcPrime ProgramCounter) (ProgramCounter, ExitReason) {
	for pc <= pcPrime {
		opcodeData := interp.Program.InstructionData[pc]
		skipLength := ProgramCounter(skip(int(pc), interp.Program.Bitmasks))

		target := execInstructions[opcodeData]
		if target == nil {
			return pc, ExitPanic
		}
		exitReason, newPC := target(interp, pc, skipLength)

		reason := exitReason.GetReasonType()
		if IsBlockTerminator(opcodeData) && (reason == CONTINUE || reason == HOST_CALL) {
			interp.GasCharged = false
		}

		switch reason {
		case PANIC, HALT:
			return 0, exitReason
		case PAGE_FAULT, OUT_OF_GAS:
			return pc, exitReason
		case HOST_CALL:
			return pc + skipLength + 1, exitReason
		}

		if IsBlockTerminator(opcodeData) {
			return newPC, exitReason
		}

		pc += skipLength + 1
	}

	return pc, ExitContinue
}
