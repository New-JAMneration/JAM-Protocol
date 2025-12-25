package PVM

import (
	"errors"
	"fmt"
)

// per-instruction based of (A.1) ψ_1,
func SingleStepInvoke(program Program, pc ProgramCounter, gas Gas, reg Registers, mem Memory) (
	error, ProgramCounter, Gas, Registers, Memory,
) {
	var exitReason error

	exitReason, pcPrime, gasPrime, registersPrime, memoryPrime := SingleStepStateTransition(
		//program.InstructionData, program.Bitmasks, program.JumpTable, pc, gas, reg, mem)
		program, pc, gas, reg, mem)
	if exitReason == ErrNotImplemented {
		return exitReason, pcPrime, gasPrime, registersPrime, memoryPrime
	}

	switch exitReason.(*PVMExitReason).Reason {
	case CONTINUE:
		return SingleStepInvoke(program, pcPrime, gasPrime, registersPrime, memoryPrime)
	case HALT, PANIC:
		return exitReason, 0, gasPrime, registersPrime, memoryPrime
	default: // HOST_CALL, OUT_OF_GAS
		return exitReason, pcPrime, gasPrime, registersPrime, memoryPrime
	}
}

var ErrNotImplemented = errors.New("instruction not implemented")

// (v.0.7.1 A.6, A.7) SingleStepStateTransition
func SingleStepStateTransition(program Program, pc ProgramCounter, gas Gas, registers Registers, memory Memory) (
	error, ProgramCounter, Gas, Registers, Memory,
) {
	// check program-counter exceed blob length
	if int(pc) >= len(program.InstructionData) {
		return PVMExitTuple(PANIC, nil), pc, gas, registers, memory
	}

	var exitReason error

	// (v0.7.1  A.19) check opcode validity
	opcodeData := program.InstructionData.isOpcode(pc)
	// check gas
	if gas < 0 {
		// pvmLogger.Debugf("service out-of-gas: required %d, but only %d", instrCount, gas)
		return PVMExitTuple(OUT_OF_GAS, nil), pc, gas, registers, memory
	}
	gas -= 1

	target := execInstructions[opcodeData]
	if target == nil {
		return ErrNotImplemented, pc, gas, registers, memory
	}
	// (v0.7.1  A.20) l = skip(iota)
	skipLength := ProgramCounter(skip(int(pc), program.Bitmasks))

	exitReason, newPC, registersPrime, memoryPrime := execInstructions[opcodeData](program.InstructionData, pc, skipLength, registers, memory, program.JumpTable, program.Bitmasks, program.InstrCount)
	// update PVM states
	registers = registersPrime
	memory = memoryPrime
	program.InstrCount++
	// logger.Debug("gasPrime, regPrime: ", gas, registersPrime)
	var pvmExit *PVMExitReason
	if !errors.As(exitReason, &pvmExit) && exitReason != nil {
		return exitReason, pc, gas, registers, memory
	}

	reason := exitReason.(*PVMExitReason).Reason
	switch reason {
	case PANIC, HALT:
		// pvmLogger.Debugf("   gas: %d", gas)
		return exitReason, 0, gas, registers, memory
	case HOST_CALL:
		return exitReason, pc + skipLength + 1, gas, registers, memory
	}

	if pc != newPC {
		// execute branch instruction
		return exitReason, newPC, gas, registers, memory
	}

	// iota' = iota + 1 +skip(iota)
	newPC += skipLength + 1
	// detailed instruction print
	// logger.Debugf("instr:%s(%d) pc=%d operand=%v gas=%d registers=%x", zeta[opcode(opcodeData)], opcodeData, programCounter, instructionCode[programCounter:programCounter+skipLength+1], gas, registers)
	// logger.Debugf("       gas : %d -> %d", gas+gasDelta, gas)

	return exitReason, newPC, gas, registers, memory
}

// block based version of (A.1) ψ_1
func BlockBasedInvoke(program Program, pc ProgramCounter, gas Gas, reg Registers, mem Memory) (error, ProgramCounter, Gas, Registers, Memory) {
	gasPrime := Gas(gas)
	// decode instructions in a block
	pcPrime, _, err := DecodeInstructionBlock(program.InstructionData, pc, program.Bitmasks)
	if err != nil {
		pvmLogger.Errorf("DecodeInstructionBlock error : %v", err)
		return err, 0, Gas(gas), reg, mem
	}
	/*
		//check block gas then execute
		// check gas, currently each instruction gas = 1, so only check instrCount
		if Gas(gas) < Gas(blockInstrCount) {
			pvmLogger.Debugf("service out-of-gas: required %d, but only %d", blockInstrCount, gas)
			return PVMExitTuple(OUT_OF_GAS, nil), pc, Gas(gas), reg, mem
		}
	*/
	// To avoid duplicate charging
	// ecalli will back to host-call level, then continue to execute remaining instructions
	/*
		if program.Bitmasks.IsStartOfBasicBlock(pc) {
			// charge gas
			// gasPrime -= Gas(blockInstrCount)
			pvmLogger.Debugf("    charge gas : %d -> %d", gas, gasPrime)
		}
	*/
	// execute instructions in the block
	pc, regPrime, memPrime, gasPrime, exitReason := ExecuteInstructions(program, pc, pcPrime, reg, mem, gas)

	var pvmExit *PVMExitReason
	if !errors.As(exitReason, &pvmExit) {
		return exitReason, 0, 0, Registers{}, Memory{}
	}

	reason := exitReason.(*PVMExitReason).Reason
	switch reason {
	case PANIC, HALT:
		return exitReason, 0, Gas(gasPrime), regPrime, memPrime
	case HOST_CALL, OUT_OF_GAS:
		return exitReason, pc, Gas(gasPrime), regPrime, memPrime
	}

	// reason == CONTINUE
	return BlockBasedInvoke(program, pc, gasPrime, regPrime, memPrime)
}

func DecodeInstructionBlock(instructionData ProgramCode, pc ProgramCounter, bitmask Bitmask) (ProgramCounter, int64, error) {
	pcPrime := pc
	count := int64(1)

	for {
		// check pc is not out of range and avoid infinit-loop
		if pc > ProgramCounter(len(instructionData)) {
			// pvmLogger.Debugf("PVM panic: program counter out of range, pcPrime = %d > program-length = %d", pcPrime, len(instructionData))
			return pc, 0, PVMExitTuple(PANIC, nil)
		}

		// check opcode is valid after computing with skip
		if !instructionData.isOpcodeValid(pc) {
			// pvmLogger.Debugf("PVM panic: decode program failed: opcode invalid")
			return pc, 0, PVMExitTuple(PANIC, nil)
		}

		// reach instruction block end
		if isBasicBlockTerminationInstruction(instructionData[pcPrime]) {
			return pcPrime, count, nil
		}
		count++
		skipLength := skip(int(pcPrime), bitmask)
		pcPrime += ProgramCounter(skipLength) + 1
	}
}

// execute each instruction in block[pc:pcPrime] , pcPrime is computed by DecodeInstructionBlock
func ExecuteInstructions(program Program, pc ProgramCounter, pcPrime ProgramCounter, registers Registers, memory Memory, gas Gas) (ProgramCounter, Registers, Memory, Gas, error) {
	// no need to worry about gas, opcode valid here, it's checked in HostCall and DecodeInstructionBlock respectively
	for pc <= pcPrime {
		if gas < 1 {
			return pc, registers, memory, gas, PVMExitTuple(OUT_OF_GAS, nil)
		}
		opcodeData := program.InstructionData[pc]
		skipLength := ProgramCounter(skip(int(pc), program.Bitmasks))

		exitReason, newPC, registersPrime, memoryPrime := execInstructions[opcodeData](program.InstructionData, pc, skipLength, registers, memory, program.JumpTable, program.Bitmasks, program.InstrCount)

		registers = registersPrime
		memory = memoryPrime
		program.InstrCount++
		gas -= 1
		// logger.Debug("gasPrime: ", gas)
		var pvmExit *PVMExitReason
		if !errors.As(exitReason, &pvmExit) && exitReason != nil {
			return pc, registers, memory, gas, exitReason
		}
		reason := exitReason.(*PVMExitReason).Reason
		switch reason {
		case PANIC, HALT:
			return 0, registers, memory, gas, exitReason
		case HOST_CALL:
			return pc + skipLength + 1, registers, memory, gas, exitReason
		}

		if pc != newPC {
			// check branch
			return newPC, registers, memory, gas, exitReason
		}

		pc += skipLength + 1
	}

	return pc, registers, memory, gas, PVMExitTuple(CONTINUE, nil)
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
