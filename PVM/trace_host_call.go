package PVM

// HostCallInstrPC returns the PC of the ecalli instruction given its fallthrough PC.
func HostCallInstrPC(program *Program, fallthroughPC ProgramCounter) ProgramCounter {
	if program == nil || fallthroughPC == 0 {
		return 0
	}
	for i := range program.Instrs {
		instr := &program.Instrs[i]
		if instr.PC+ProgramCounter(instr.SkipLen)+1 == fallthroughPC {
			return instr.PC
		}
	}
	return 0
}
