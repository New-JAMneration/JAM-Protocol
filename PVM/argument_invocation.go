package PVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// (A.40) Ψ_M
func Psi_M(
	code StandardCodeFormat,
	counter ProgramCounter, // program counter
	gas types.Gas, // gas counter
	argument Argument, // argument
	omegas Omegas, // jump table
	addition HostCallArgs, // host-call context
) (
	psi_result Psi_M_ReturnType,
) {
	programCode, registers, memory, err := SingleInitializer(code, argument)
	// Y(p) = nil
	if err != nil {
		return Psi_M_ReturnType{
			Gas:           0,
			ReasonOrBytes: PVMExitTuple(PANIC, nil),
			Addition:      addition,
		}
	}

	// addition
	programBlob, exitReason := DeBlobProgramCode(programCode)
	standardProgram := StandardProgram{
		Memory:      memory,
		Registers:   registers,
		ProgramBlob: programBlob,
		ExitReason:  exitReason,
	}

	g, v, a := R(gas, Psi_H(standardProgram, counter, gas, standardProgram.Registers, standardProgram.Memory, omegas, addition))
	return Psi_M_ReturnType{
		Gas:           types.Gas(g),
		ReasonOrBytes: v,
		Addition:      a,
	}
}

// (A.41) R
func R(priorGas types.Gas, Psi_H_Return Psi_H_ReturnType) (Gas, any, HostCallArgs) {
	u := priorGas - types.Gas(max(Psi_H_Return.Gas, 0))
	switch Psi_H_Return.ExitReason.(*PVMExitReason).Reason {
	case OUT_OF_GAS:
		return Gas(u), OUT_OF_GAS, Psi_H_Return.Addition
	case HALT:
		if isReadable(Psi_H_Return.Reg[7], Psi_H_Return.Reg[8], Psi_H_Return.Ram) {
			startPage := Psi_H_Return.Reg[7] / ZP
			endPage := (Psi_H_Return.Reg[7] + Psi_H_Return.Reg[8]) / ZP
			value := []byte{}
			for i := startPage; i <= endPage; i++ {
				value = append(value, Psi_H_Return.Ram.Pages[uint32(i)].Value...)
			}
			return Gas(u), value, Psi_H_Return.Addition
		}
		return Gas(u), []byte{}, Psi_H_Return.Addition
	default:
		return Gas(u), PANIC, Psi_H_Return.Addition
	}
}

func isReadable(start, offset uint64, m Memory) bool {
	startPage := uint32(start / ZP)
	endPage := uint32((start + offset) / ZP)

	return !(m.GetPageAccess(startPage) == MemoryInaccessible ||
		m.GetPageAccess(endPage) == MemoryInaccessible)
}

func isWriteable(start, offset uint64, m Memory) bool {
	startPage := uint32(start / ZP)
	endPage := uint32((start + offset) / ZP)

	return m.GetPageAccess(startPage) == MemoryReadWrite &&
		m.GetPageAccess(endPage) == MemoryReadWrite
}

type Psi_M_ReturnType struct {
	Gas           types.Gas
	ReasonOrBytes any
	Addition      HostCallArgs
}
