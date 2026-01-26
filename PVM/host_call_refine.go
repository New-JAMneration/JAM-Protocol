package PVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// historical_lookup = 6
func historicalLookup(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	// first check v panic, then assign a
	h, o := input.Interpreter.Registers[8], input.Interpreter.Registers[9]

	offset := uint64(32)
	if !isReadable(h, offset, *input.Interpreter.Memory) { // not readable, return panic
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	codeHash := types.OpaqueHash(input.Interpreter.Memory.Read(h, offset))

	s := input.Addition.ServiceId
	// assign a
	var a *types.ServiceAccount
	var v *types.ByteSequence

	if account, accountExists := (*input.Addition.ServiceAccountState)[*s]; accountExists && input.Interpreter.Registers[7] == 0xffffffffffffffff {
		a = &account
	} else if account, accountExists := (*input.Addition.ServiceAccountState)[types.ServiceId(input.Interpreter.Registers[7])]; accountExists {
		a = &account
	}

	var f uint64
	var l uint64

	if a != nil {
		val := service_account.HistoricalLookup(*a, input.Addition.RefineArgs.TimeSlot, codeHash)
		v = &val
		f = min(input.Interpreter.Registers[10], uint64(len(*v)))
		l = min(input.Interpreter.Registers[11], uint64(len(*v))-f)
	}

	if !isWriteable(o, l, *input.Interpreter.Memory) && l != 0 { // not writeable, return panic
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	if v == nil {
		input.Interpreter.Registers[7] = NONE
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	input.Interpreter.Registers[7] = uint64(len(*v))
	input.Interpreter.Memory.Write(o, l, (*v)[f:f+l])

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// export = 7
func export(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	p := input.Interpreter.Registers[7]
	z := min(input.Interpreter.Registers[8], types.SegmentSize)

	if !isReadable(p, z, *input.Interpreter.Memory) { // not readable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	segmentLength := input.Addition.ExportSegmentOffset + uint(len(input.Addition.ExportSegment))
	// otherwise if ζ + |e| >= W_M
	if segmentLength > types.MaxExportCount {
		input.Interpreter.Registers[7] = FULL
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// data = mu_p...+z
	data := input.Interpreter.Memory.Read(p, z)
	x := zeroPadding(data, types.SegmentSize)
	exportSegment := types.ExportSegment{}
	copy(exportSegment[:], x)

	input.Interpreter.Registers[7] = uint64(input.Addition.ExportSegmentOffset) + uint64(segmentLength)
	input.Addition.ExportSegment = append(input.Addition.ExportSegment, exportSegment)

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// machine = 8
func machine(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	po, pz, i := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9]
	// pz = offset
	if !isReadable(po, pz, *input.Interpreter.Memory) { // not readable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	p := input.Interpreter.Memory.Read(po, pz)

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
		input.Interpreter.Registers[7] = HUH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise
	input.Interpreter.Registers[7] = n
	input.Addition.IntegratedPVMMap[n] = IntegratedPVMType{
		ProgramCode: ProgramCode(p),
		Memory:      u,
		PC:          ProgramCounter(i),
	}

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// peek = 9
func peek(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	n, o, s, z := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9], input.Interpreter.Registers[10]

	if z == 0 {
		input.Interpreter.Registers[7] = OK
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// z = offset
	if !isWriteable(o, z, *input.Interpreter.Memory) { // not writeable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if n not in K(m)
	if _, exists := input.Addition.IntegratedPVMMap[n]; !exists {
		input.Interpreter.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise if N_s...+z not subset of \mathbf{V}_m[n]_u
	// can be simplify to check readable, if not readable => Inaccessible
	if !isReadable(s, z, input.Addition.IntegratedPVMMap[n].Memory) {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise
	// read data from m[n]_u first
	integratedPVMType := input.Addition.IntegratedPVMMap[n]
	data := (&integratedPVMType.Memory).Read(s, z)
	// write data into memory
	input.Interpreter.Memory.Write(o, z, data[s:s+z])

	input.Interpreter.Registers[7] = OK
	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// poke = 10
func poke(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	n, s, o, z := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9], input.Interpreter.Registers[10]

	if !isReadable(s, z, *input.Interpreter.Memory) { // not readable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if n not in K(m)
	if _, exists := input.Addition.IntegratedPVMMap[n]; !exists {
		input.Interpreter.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise if N_o...+z not subset of \mathbf{V}_m[n]_u
	if !isWriteable(o, z, input.Addition.IntegratedPVMMap[n].Memory) { // not writeable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise
	// read data from memory first
	data := input.Interpreter.Memory.Read(s, z)
	// write data into m[n]_u
	integratedPVMType := input.Addition.IntegratedPVMMap[n]
	(&integratedPVMType.Memory).Write(o, z, data)
	input.Interpreter.Registers[7] = OK

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// pages = 11 , GP 0.6.7 void is renamed pages
func pages(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	n, p, c := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9]
	// u = panic
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		// u = panic
		input.Interpreter.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise if p < 16 or p + c >= 2^32 / ZP or i in N_p...+c : (u_A)_i = nil
	if p < 16 || p+c >= (1<<32)/ZP || !isReadable(p, c, input.Addition.IntegratedPVMMap[n].Memory) {
		input.Interpreter.Registers[7] = HUH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise : ok
	for i := uint32(p); i < uint32(c); i++ {
		input.Addition.IntegratedPVMMap[n].Memory.Pages[i] = &Page{
			Value:  make([]byte, ZP),
			Access: MemoryInaccessible,
		}
	}

	input.Interpreter.Registers[7] = OK

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// invoke = 12
func invoke(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	n, o := input.Interpreter.Registers[7], input.Interpreter.Registers[8]

	offset := uint64(112)
	// g = panic
	if !isWriteable(o, offset, input.Addition.IntegratedPVMMap[n].Memory) { // not writeable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if n not in M
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		input.Interpreter.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// assign g, w  |  g => gas , w => registers[13]   , 8(gas) + 8(uint64) * 13 = 112
	var gas uint64
	var w Registers

	// first read data from memory
	data := input.Interpreter.Memory.Read(o, offset)

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

	tempMemory := input.Addition.IntegratedPVMMap[n].Memory
	tempAddition := input.Addition
	tempHost := NewHost(input.Addition.Program, w, &tempMemory, Gas(gas), tempAddition, input.HostCalls)

	var c ExitReason
	var pcPrime ProgramCounter

	if GasChargingMode == "blockBased" {
		c, pcPrime = tempHost.Interpreter.BlockBasedInvoke(input.Addition.IntegratedPVMMap[n].PC)
	} else {
		c, pcPrime = tempHost.Interpreter.SingleStepInvoke(input.Addition.IntegratedPVMMap[n].PC)
	}

	// mu* = mu
	encoder := types.NewEncoder()
	data = types.ByteSequence(make([]byte, offset))
	encoded, _ := encoder.Encode(&tempHost.Interpreter.Gas)
	copy(data, encoded)
	for i := uint64(1); i < offset/8; i++ {
		encoded, _ := encoder.Encode(&tempHost.Interpreter.Registers[i-1])
		copy(data[8*i:8*(i+1)], encoded)
	}
	// write data into memory (mu)
	input.Interpreter.Memory.Write(o, offset, data)

	// m* = m
	tmp := input.Addition.IntegratedPVMMap[n]
	tmp.Memory = *tempHost.Interpreter.Memory
	if c.GetReasonType() == HOST_CALL {
		tmp.PC = pcPrime + 1 + ProgramCounter(skip(int(pcPrime), input.Addition.Program.Bitmasks))
	} else {
		tmp.PC = pcPrime
	}
	input.Addition.IntegratedPVMMap[n] = tmp

	switch c.GetReasonType() {
	case HOST_CALL:
		input.Interpreter.Registers[7] = INNERHOST
		input.Interpreter.Registers[8] = uint64(c.GetHostCallID())

	case PAGE_FAULT:
		input.Interpreter.Registers[7] = INNERFAULT
		input.Interpreter.Registers[8] = uint64(c.GetPageFaultAddress())

	case OUT_OF_GAS:
		input.Interpreter.Registers[7] = INNEROOG

	case PANIC:
		input.Interpreter.Registers[7] = INNERPANIC

	case HALT:
		input.Interpreter.Registers[7] = INNERHALT

	}

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// expunge = 13
func expunge(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	n := input.Interpreter.Registers[7]
	// n not in K(m)
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		input.Interpreter.Registers[7] = WHO

		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	input.Interpreter.Registers[7] = uint64(input.Addition.IntegratedPVMMap[n].PC)
	// m ˋ n
	delete(input.Addition.IntegratedPVMMap, n)

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}
