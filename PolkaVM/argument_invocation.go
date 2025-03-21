package PolkaVM

// (A.40) Ψ_M
func Psi_M(
	code StandardCodeFormat,
	counter ProgramCounter, // program counter
	gas Gas, // gas counter
	argument Argument, // argument
	omegas Omegas, // jump table
	addition any, // host-call context
) (
	psi_result Psi_M_ReturnType,
) {
	standardProgram := StandardProgramInit(code, argument)
	if standardProgram.ExitReason != nil {
		return Psi_M_ReturnType{
			Gas:           gas,
			ReasonOrBytes: standardProgram.ExitReason,
			Addition:      addition,
		}
	}

	// can be simplified during the refactor phase
	g, v, a := R(Psi_H(standardProgram, counter, gas, standardProgram.Registers, standardProgram.Memory, omegas, addition))
	return Psi_M_ReturnType{
		Gas:           g,
		ReasonOrBytes: v,
		Addition:      a,
	}
}

// (A.41) R
func R(Psi_H_Return Psi_H_ReturnType) (Gas, any, any) {
	switch Psi_H_Return.ExitReason.(*PVMExitReason).Reason {
	case OUT_OF_GAS:
		return Psi_H_Return.Gas, OUT_OF_GAS, Psi_H_Return.Addition
	case HALT:
		if isReadable(Psi_H_Return.Reg[7], Psi_H_Return.Reg[8], Psi_H_Return.Ram) {
			startPage := Psi_H_Return.Reg[7] / ZP
			endPage := (Psi_H_Return.Reg[7] + Psi_H_Return.Reg[8]) / ZP
			value := []byte{}
			for i := startPage; i <= endPage; i++ {
				value = append(value, Psi_H_Return.Ram.Pages[uint32(i)].Value...)
			}
			return Psi_H_Return.Gas, value, Psi_H_Return.Addition
		}
		return Psi_H_Return.Gas, []byte{}, Psi_H_Return.Addition
	default:
		return Psi_H_Return.Gas, PANIC, Psi_H_Return.Addition
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
	Gas           Gas
	ReasonOrBytes any
	Addition      any
}
