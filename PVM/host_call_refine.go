package PVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// historical_lookup = 6
func historicalLookup(input OmegaInput) (output OmegaOutput) {
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   ExitOOG,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// first check v panic, then assign a
	h, o := input.Registers[8], input.Registers[9]

	offset := uint64(32)
	if !isReadable(h, offset, input.Memory) { // not readable, return panic
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   ExitPanic,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	codeHash := types.OpaqueHash(input.Memory.Read(h, offset))

	s := input.Addition.ServiceId
	// assign a
	var a *types.ServiceAccount
	var v *types.ByteSequence

	if account, accountExists := (*input.Addition.ServiceAccountState)[*s]; accountExists && input.Registers[7] == 0xffffffffffffffff {
		a = &account
	} else if account, accountExists := (*input.Addition.ServiceAccountState)[types.ServiceId(input.Registers[7])]; accountExists {
		a = &account
	}

	var f uint64
	var l uint64

	if a != nil {
		val := service_account.HistoricalLookup(*a, input.Addition.RefineArgs.TimeSlot, codeHash)
		v = &val
		f = min(input.Registers[10], uint64(len(*v)))
		l = min(input.Registers[11], uint64(len(*v))-f)
	}

	if !isWriteable(o, l, input.Memory) && l != 0 { // not writeable, return panic
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   ExitPanic,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	if v == nil {
		input.Registers[7] = NONE
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	input.Registers[7] = uint64(len(*v))
	input.Memory.Write(o, l, (*v)[f:f+l])

	return OmegaOutput{
		ExitReason:   ExitContinue,
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// export = 7
func export(input OmegaInput) (output OmegaOutput) {
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   ExitOOG,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	p := input.Registers[7]
	z := min(input.Registers[8], types.SegmentSize)

	if !isReadable(p, z, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   ExitPanic,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	segmentLength := input.Addition.ExportSegmentOffset + uint(len(input.Addition.ExportSegment))
	// otherwise if ζ + |e| >= W_M
	if segmentLength > types.MaxExportCount {
		input.Registers[7] = FULL
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// data = mu_p...+z
	data := input.Memory.Read(p, z)
	x := zeroPadding(data, types.SegmentSize)
	exportSegment := types.ExportSegment{}
	copy(exportSegment[:], x)

	input.Registers[7] = uint64(input.Addition.ExportSegmentOffset) + uint64(segmentLength)
	input.Addition.ExportSegment = append(input.Addition.ExportSegment, exportSegment)

	return OmegaOutput{
		ExitReason:   ExitContinue,
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// machine = 8
func machine(input OmegaInput) (output OmegaOutput) {
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   ExitOOG,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	po, pz, i := input.Registers[7], input.Registers[8], input.Registers[9]
	// pz = offset
	if !isReadable(po, pz, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   ExitPanic,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	p := input.Memory.Read(po, pz)

	// find first i not in K(m)
	n := uint64(0)
	for ; n <= ^uint64(0); n++ {
		if _, pvmTypeExists := input.Addition.IntegratedPVMMap[n]; !pvmTypeExists {
			break
		}
	}

	var u Memory
	_, exitReason := DeBlobProgramCode(p)
	// otherwise if deblob(p) = PANIC
	if exitReason == ExitPanic {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise
	input.Registers[7] = n
	input.Addition.IntegratedPVMMap[n] = IntegratedPVMType{
		ProgramCode: ProgramCode(p),
		Memory:      u,
		PC:          ProgramCounter(i),
	}

	return OmegaOutput{
		ExitReason:   ExitContinue,
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// peek = 9
func peek(input OmegaInput) (output OmegaOutput) {
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   ExitOOG,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	n, o, s, z := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]

	if z == 0 {
		input.Registers[7] = OK
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// z = offset
	if !isWriteable(o, z, input.Memory) { // not writeable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   ExitPanic,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if n not in K(m)
	if _, exists := input.Addition.IntegratedPVMMap[n]; !exists {
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if N_s...+z not subset of \mathbf{V}_m[n]_u
	// can be simplify to check readable, if not readable => Inaccessible
	if !isReadable(s, z, input.Addition.IntegratedPVMMap[n].Memory) {
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise
	// read data from m[n]_u first
	integratedPVMType := input.Addition.IntegratedPVMMap[n]
	data := integratedPVMType.Memory.Read(s, z)
	// write data into memory
	input.Memory.Write(o, z, data[s:s+z])

	input.Registers[7] = OK
	return OmegaOutput{
		ExitReason:   ExitContinue,
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// poke = 10
func poke(input OmegaInput) (output OmegaOutput) {
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   ExitOOG,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	n, s, o, z := input.Registers[7], input.Registers[8], input.Registers[9], input.Registers[10]

	if !isReadable(s, z, input.Memory) { // not readable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   ExitPanic,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if n not in K(m)
	if _, exists := input.Addition.IntegratedPVMMap[n]; !exists {
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if N_o...+z not subset of \mathbf{V}_m[n]_u
	if !isWriteable(o, z, input.Addition.IntegratedPVMMap[n].Memory) { // not writeable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   ExitPanic,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise
	// read data from memory first
	data := input.Memory.Read(s, z)
	// write data into m[n]_u
	integratedPVMType := input.Addition.IntegratedPVMMap[n]
	integratedPVMType.Memory.Write(o, z, data)
	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   ExitContinue,
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// pages = 11 , GP 0.6.7 void is renamed pages
func pages(input OmegaInput) (output OmegaOutput) {
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   ExitOOG,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	n, p, c := input.Registers[7], input.Registers[8], input.Registers[9]
	// u = panic
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		// u = panic
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if p < 16 or p + c >= 2^32 / ZP or i in N_p...+c : (u_A)_i = nil
	if p < 16 || p+c >= (1<<32)/ZP || !isReadable(p, c, input.Addition.IntegratedPVMMap[n].Memory) {
		input.Registers[7] = HUH
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise : ok
	for i := uint32(p); i < uint32(c); i++ {
		input.Addition.IntegratedPVMMap[n].Memory.Pages[i] = &Page{
			Value:  make([]byte, ZP),
			Access: MemoryInaccessible,
		}
	}

	input.Registers[7] = OK

	return OmegaOutput{
		ExitReason:   ExitContinue,
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// invoke = 12
func invoke(input OmegaInput) (output OmegaOutput) {
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   ExitOOG,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	n, o := input.Registers[7], input.Registers[8]

	offset := uint64(112)
	// g = panic
	if !isWriteable(o, offset, input.Addition.IntegratedPVMMap[n].Memory) { // not writeable, return
		input.Registers[7] = OOB
		return OmegaOutput{
			ExitReason:   ExitPanic,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// otherwise if n not in M
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		input.Registers[7] = WHO
		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	// assign g, w  |  g => gas , w => registers[13]   , 8(gas) + 8(uint64) * 13 = 112
	var gas uint64
	var w Registers

	// first read data from memory
	data := input.Memory.Read(o, offset)

	decoder := types.NewDecoder()
	// decode gas
	err := decoder.Decode(data[:8], &gas)
	if err != nil {
		pvmLogger.Errorf("host-call function \"invoke\" decode gas error : %v", err)
	}
	// decode registers
	for i := uint64(1); i < offset/8; i++ {
		err = decoder.Decode(data[8*i:8*(i+1)], &w[i-1])
		if err != nil {
			pvmLogger.Errorf("host-call function \"invoke\" decode register:%d error : %v", i-1, err)
		}
	}
	// psi
	input.Addition.Program.InstructionData = input.Addition.IntegratedPVMMap[n].ProgramCode

	var c ExitReason
	var pcPrime ProgramCounter
	var gasPrime Gas
	var wPrime Registers
	var uPrime Memory

	if GasChargingMode == "blockBased" {
		c, pcPrime, gasPrime, wPrime, uPrime = BlockBasedInvoke(input.Addition.Program, input.Addition.IntegratedPVMMap[n].PC, Gas(gas), w, input.Addition.IntegratedPVMMap[n].Memory)
	} else {
		c, pcPrime, gasPrime, wPrime, uPrime = SingleStepInvoke(input.Addition.Program, input.Addition.IntegratedPVMMap[n].PC, Gas(gas), w, input.Addition.IntegratedPVMMap[n].Memory)
	}

	// mu* = mu
	encoder := types.NewEncoder()
	data = types.ByteSequence(make([]byte, offset))
	encoded, _ := encoder.Encode(&gasPrime)
	copy(data, encoded)
	for i := uint64(1); i < offset/8; i++ {
		encoded, _ := encoder.Encode(&wPrime[i-1])
		copy(data[8*i:8*(i+1)], encoded)
	}
	// write data into memory (mu)
	input.Memory.Write(o, offset, data)

	// m* = m
	tmp := input.Addition.IntegratedPVMMap[n]
	tmp.Memory = uPrime
	if c.GetReasonType() == HOST_CALL {
		tmp.PC = pcPrime + 1 + ProgramCounter(skip(int(pcPrime), input.Addition.Program.Bitmasks))
	} else {
		tmp.PC = pcPrime
	}
	input.Addition.IntegratedPVMMap[n] = tmp

	switch c.GetReasonType() {
	case HOST_CALL:
		input.Registers[7] = INNERHOST
		input.Registers[8] = uint64(c.GetHostCallID())

	case PAGE_FAULT:
		input.Registers[7] = INNERFAULT
		input.Registers[8] = uint64(c.GetPageFaultAddress())

	case OUT_OF_GAS:
		input.Registers[7] = INNEROOG

	case PANIC:
		input.Registers[7] = INNERPANIC

	case HALT:
		input.Registers[7] = INNERHALT

	}

	return OmegaOutput{
		ExitReason:   ExitContinue,
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// expunge = 13
func expunge(input OmegaInput) (output OmegaOutput) {
	newGas := input.Gas - 10
	if newGas < 0 {
		return OmegaOutput{
			ExitReason:   ExitOOG,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	n := input.Registers[7]
	// n not in K(m)
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		input.Registers[7] = WHO

		return OmegaOutput{
			ExitReason:   ExitContinue,
			NewGas:       newGas,
			NewRegisters: input.Registers,
			NewMemory:    input.Memory,
			Addition:     input.Addition,
		}
	}

	input.Registers[7] = uint64(input.Addition.IntegratedPVMMap[n].PC)
	// m ˋ n
	delete(input.Addition.IntegratedPVMMap, n)

	return OmegaOutput{
		ExitReason:   ExitContinue,
		NewGas:       newGas,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}
