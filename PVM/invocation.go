package PVM

import (
	"fmt"
)

// per-instruction based of (A.1) ψ_1,
func (interp *Interpreter) SingleStepInvoke(pc ProgramCounter) (ExitReason, ProgramCounter) {
	var exitReason ExitReason

	exitReason, pcPrime := interp.SingleStepStateTransition(pc)

	switch exitReason.GetReasonType() {
	case CONTINUE:
		return interp.SingleStepInvoke(pcPrime)
	case HALT, PANIC:
		return exitReason, 0
	default: // HOST_CALL, OUT_OF_GAS
		return exitReason, pcPrime
	}
}

// (v.0.7.1 A.6, A.7) SingleStepStateTransition
func (interp *Interpreter) SingleStepStateTransition(pc ProgramCounter) (ExitReason, ProgramCounter) {
	// check program-counter exceed blob length
	if int(pc) >= len(interp.Program.InstructionData) {
		return ExitPanic, pc
	}

	var exitReason ExitReason

	// (v0.7.1  A.19) check opcode validity
	opcodeData := interp.Program.InstructionData.isOpcode(pc)
	// check gas
	if interp.Gas < 0 {
		// pvmLogger.Debugf("service out-of-gas: required %d, but only %d", interp.InstrCount, interp.Gas)
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
	interp.Program.InstrCount++

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
	// detailed instruction print
	// logger.Debugf("instr:%s(%d) pc=%d operand=%v gas=%d registers=%x", zeta[opcode(opcodeData)], opcodeData, programCounter, instructionCode[programCounter:programCounter+skipLength+1], interp.Gas, interp.Registers)
	// logger.Debugf("       gas : %d -> %d", interp.Gas+gasDelta, interp.Gas)

	return exitReason, newPC
}

// block based version of (A.1) ψ_1
func (interp *Interpreter) BlockBasedInvoke(pc ProgramCounter) (ExitReason, ProgramCounter) {
	// decode instructions in a block
	pcPrime, _, exitReason := DecodeInstructionBlock(interp.Program.InstructionData, pc, interp.Program.Bitmasks)
	if exitReason.GetReasonType() != CONTINUE {
		pvmLogger.Errorf("DecodeInstructionBlock error : %v", exitReason)
		return exitReason, 0
	}

	// execute instructions in the block
	pc, exitReason = interp.ExecuteInstructions(pc, pcPrime)
	reason := exitReason.GetReasonType()
	switch reason {
	case PANIC, HALT:
		return exitReason, 0
	case HOST_CALL, OUT_OF_GAS:
		return exitReason, pc
	}

	// reason == CONTINUE
	return interp.BlockBasedInvoke(pc)
}

func DecodeInstructionBlock(instructionData ProgramCode, pc ProgramCounter, bitmask Bitmask) (ProgramCounter, int64, ExitReason) {
	pcPrime := pc
	count := int64(1)

	for {
		// check pc is not out of range and avoid infinit-loop
		if pc > ProgramCounter(len(instructionData)) {
			// pvmLogger.Debugf("PVM panic: program counter out of range, pcPrime = %d > program-length = %d", pcPrime, len(instructionData))
			return pc, 0, ExitPanic
		}

		// check opcode is valid after computing with skip
		if !instructionData.isOpcodeValid(pc) {
			// pvmLogger.Debugf("PVM panic: decode program failed: opcode invalid")
			return pc, 0, ExitPanic
		}

		// reach instruction block end
		if isBasicBlockTerminationInstruction(instructionData[pcPrime]) {
			return pcPrime, count, ExitContinue
		}
		count++
		skipLength := skip(int(pcPrime), bitmask)
		pcPrime += ProgramCounter(skipLength) + 1
	}
}

// execute each instruction in block[pc:pcPrime] , pcPrime is computed by DecodeInstructionBlock
func (interp *Interpreter) ExecuteInstructions(pc ProgramCounter, pcPrime ProgramCounter) (ProgramCounter, ExitReason) { // no need to worry about gas, opcode valid here, it's checked in HostCall and DecodeInstructionBlock respectively
	for pc <= pcPrime {
		if interp.Gas < 1 {
			return pc, ExitOOG
		}
		opcodeData := interp.Program.InstructionData[pc]
		skipLength := ProgramCounter(skip(int(pc), interp.Program.Bitmasks))

		exitReason, newPC := execInstructions[opcodeData](interp, pc, skipLength)
		interp.Program.InstrCount++
		interp.Gas -= 1
		// logger.Debug("gasPrime: ", interp.Gas)
		reason := exitReason.GetReasonType()
		switch reason {
		case PANIC, HALT:
			return 0, exitReason
		case HOST_CALL:
			return pc + skipLength + 1, exitReason
		}

		if pc != newPC {
			// check branch
			return newPC, exitReason
		}

		pc += skipLength + 1
	}

	return pc, ExitContinue
}

type Int interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int8 | ~int16 | ~int32 | ~int64 | ~int | ~uint
}

func formatInt[T Int](num T) string {
	if instrLogFormat == "hex" {
		return fmt.Sprintf("0x%x", num)
	}
	return fmt.Sprintf("%d", num)
}
