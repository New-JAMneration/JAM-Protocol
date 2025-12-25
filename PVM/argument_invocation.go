package PVM

import (
	"errors"

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
	instrCount = 0
	programCode, registers, memory, err := SingleInitializer(code, argument)
	// Y(p) = nil
	if err != nil {
		return Psi_M_ReturnType{
			Gas:           0,
			ReasonOrBytes: PVMExitTuple(PANIC, nil),
			Addition:      addition,
		}
	}

	program, err := DeBlobProgramCode(programCode)
	var pvmExit *PVMExitReason
	if !errors.As(err, &pvmExit) {
		return
	}
	reason := err.(*PVMExitReason).Reason
	if reason == PANIC {
		return Psi_M_ReturnType{
			Gas:           0,
			ReasonOrBytes: PVMExitTuple(PANIC, nil),
			Addition:      addition,
		}
	}

	addition.Program = program

	g, v, a := R(gas, HostCall(program, counter, gas, registers, memory, omegas, addition))
	return Psi_M_ReturnType{
		Gas:           types.Gas(g),
		ReasonOrBytes: v,
		Addition:      a,
	}
}

// (A.41) R
func R(priorGas types.Gas, Psi_H_Return Psi_H_ReturnType) (Gas, any, HostCallArgs) {
	u := priorGas - types.Gas(max(Psi_H_Return.Gas, 0))

	var pvmExit *PVMExitReason
	if !errors.As(Psi_H_Return.ExitReason, &pvmExit) {
		return Gas(u), Psi_H_Return.ExitReason, Psi_H_Return.Addition
	}
	switch Psi_H_Return.ExitReason.(*PVMExitReason).Reason {
	case OUT_OF_GAS:
		return Gas(u), OUT_OF_GAS, Psi_H_Return.Addition
	case HALT:
		if isReadable(Psi_H_Return.Reg[7], Psi_H_Return.Reg[8], Psi_H_Return.Ram) {
			start := uint64(Psi_H_Return.Reg[7])
			length := uint64(Psi_H_Return.Reg[8])
			if length == 0 {
				return Gas(u), nil, Psi_H_Return.Addition
			}

			value, ok := readRAM(start, length, Psi_H_Return.Ram)
			if !ok {
				return Gas(u), []byte{}, Psi_H_Return.Addition
			}
			return Gas(u), value, Psi_H_Return.Addition
		}
		return Gas(u), []byte{}, Psi_H_Return.Addition
	default:
		return Gas(u), PANIC, Psi_H_Return.Addition
	}
}

func isReadable(start, offset uint64, m Memory) bool {
	if offset == 0 {
		return true
	}
	startPage := uint32(start / ZP)
	endPage := uint32((start + offset - 1) / ZP)
	for p := startPage; p <= endPage; p++ {
		if m.GetPageAccess(p) == MemoryInaccessible {
			return false
		}
	}
	return true
}

func isWriteable(start, offset uint64, m Memory) bool {
	startPage := uint32(start / ZP)
	endPage := uint32((start + offset) / ZP)

	return m.GetPageAccess(startPage) == MemoryReadWrite &&
		m.GetPageAccess(endPage) == MemoryReadWrite
}

func readRAM(start, length uint64, m Memory) ([]byte, bool) {
	if length == 0 {
		return []byte{}, true
	}
	end := start + length // [start, end)
	startPage := uint32(start / ZP)
	endPage := uint32((end - 1) / ZP)

	out := make([]byte, 0, length)

	for p := startPage; p <= endPage; p++ {
		page, ok := m.Pages[p]
		if !ok || page.Value == nil || len(page.Value) == 0 {
			return nil, false
		}

		pageStartAddr := uint64(p) * ZP
		pageEndAddr := pageStartAddr + ZP

		s := max(start, pageStartAddr)
		e := min(end, pageEndAddr)
		if e <= s {
			continue
		}

		offS := s - pageStartAddr
		offE := e - pageStartAddr

		if offE > uint64(len(page.Value)) {
			return nil, false
		}

		out = append(out, page.Value[offS:offE]...)
	}

	if uint64(len(out)) != length {
		return nil, false
	}
	return out, true
}

type Psi_M_ReturnType struct {
	Gas           types.Gas
	ReasonOrBytes any
	Addition      HostCallArgs
}
