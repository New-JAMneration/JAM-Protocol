package PVM

import (
	"encoding/binary"
	"math"

	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// gasFromUint64 clamps a guest-supplied gas budget into signed Gas.
func gasFromUint64(v uint64) Gas {
	if v > uint64(math.MaxInt64) {
		return math.MaxInt64
	}
	return Gas(v)
}

// historical_lookup = 7
func historicalLookup(input OmegaInput) (output OmegaOutput) {
	// g = M_{H,c} + 𝒢(M_{H,ℓ}, z); z = ω₁₁
	cost := addGas(HostGasHistoricalLookupConst, MemGas(HostGasHistoricalLookupOctets, input.VM.Registers[11]))
	if result := chargeGasAndCheck(&input, cost); result != nil {
		return *result
	}

	// first check v panic, then assign a
	h, o := input.VM.Registers[8], input.VM.Registers[9]

	offset := uint64(32)
	if !input.VM.Mem.IsReadable(h, offset) { // not readable, return panic
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	codeHash := types.OpaqueHash(input.VM.Mem.Read(h, offset))

	s := input.Addition.ServiceID
	// assign a
	var a *types.ServiceAccount
	var v types.ByteSequence

	if account, accountExists := (*input.Addition.ServiceAccountState)[*s]; accountExists && input.VM.Registers[7] == 0xffffffffffffffff {
		a = &account
	} else if account, accountExists := (*input.Addition.ServiceAccountState)[types.ServiceID(input.VM.Registers[7])]; accountExists {
		a = &account
	}

	var f uint64
	var l uint64

	if a != nil {
		v = service_account.HistoricalLookup(*a, input.Addition.RefineArgs.TimeSlot, codeHash)
		f = min(input.VM.Registers[10], uint64(len(v)))
		l = min(input.VM.Registers[11], uint64(len(v))-f)
	}

	if !input.VM.Mem.IsWriteable(o, l) && l != 0 { // not writeable, return panic
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	if v == nil {
		input.VM.Registers[7] = NONE
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	input.VM.Registers[7] = uint64(len(v))
	input.VM.Mem.Write(o, (v)[f:f+l])

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// export = 8
func export(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input, HostGasExport); result != nil { // M_E
		return *result
	}

	p := input.VM.Registers[7]
	z := min(input.VM.Registers[8], types.SegmentSize)

	if !input.VM.Mem.IsReadable(p, z) { // not readable, return
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	segmentLength := uint64(input.Addition.ExportSegmentOffset) + uint64(len(input.Addition.ExportSegment))
	// otherwise if ζ + |e| >= W_X
	if segmentLength > types.MaxExportCount {
		input.VM.Registers[7] = FULL
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// data = mu_p...+z
	data := input.VM.Mem.Read(p, z)
	x := zeroPadding(data, types.SegmentSize)
	exportSegment := types.ExportSegment{}
	copy(exportSegment[:], x)

	input.VM.Registers[7] = segmentLength
	input.Addition.ExportSegment = append(input.Addition.ExportSegment, exportSegment)

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// machine = 9
func machine(input OmegaInput) (output OmegaOutput) {
	po, pz, i := input.VM.Registers[7], input.VM.Registers[8], input.VM.Registers[9]
	// g = M_{M,c} + 𝒢(M_{M,ℓ}, p_Z)
	cost := addGas(HostGasMachineConst, MemGas(HostGasMachineOctets, pz))
	if result := chargeGasAndCheck(&input, cost); result != nil {
		return *result
	}
	if !input.VM.Mem.IsReadable(po, pz) {
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	if uint64(len(input.Addition.IntegratedPVMMap)) >= 63 {
		input.VM.Registers[7] = HUH
		return OmegaOutput{ExitReason: ExitContinue, Addition: input.Addition}
	}

	p := input.VM.Mem.Read(po, pz)

	// find first n not in K(m)
	n := uint64(0)
	for ; n <= ^uint64(0); n++ {
		if _, pvmTypeExists := input.Addition.IntegratedPVMMap[n]; !pvmTypeExists {
			break
		}
	}

	prog, exitReason := DeBlobProgramCode(p, i)
	if exitReason != ExitContinue {
		input.VM.Registers[7] = HUH
		return OmegaOutput{ExitReason: ExitContinue, Addition: input.Addition}
	}

	// otherwise
	input.VM.Registers[7] = n
	input.Addition.IntegratedPVMMap[n] = IntegratedPVMType{
		ProgramCode: ProgramCode(p),
		Program:     &prog,
		Memory:      Memory{},
		PC:          ProgramCounter(i),
		GasCharged:  false, // GP B.6 Ω_M: gaschargedflag = ⊥ at creation
	}

	return OmegaOutput{ExitReason: ExitContinue, Addition: input.Addition}
}

// peek = 10
func peek(input OmegaInput) (output OmegaOutput) {
	n, o, s, z := input.VM.Registers[7], input.VM.Registers[8], input.VM.Registers[9], input.VM.Registers[10]
	// g = M_{P,c} + 𝒢(M_{P,ℓ}, z)
	cost := addGas(HostGasPeekConst, MemGas(HostGasPeekOctets, z))
	if result := chargeGasAndCheck(&input, cost); result != nil {
		return *result
	}

	if z == 0 {
		input.VM.Registers[7] = OK
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// z = offset
	if !input.VM.Mem.IsWriteable(o, z) { // not writeable, return
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if n not in K(m)
	if _, exists := input.Addition.IntegratedPVMMap[n]; !exists {
		input.VM.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise if N_s...+z not subset of \mathbf{V}_m[n]_u
	if !isReadable(s, z, input.Addition.IntegratedPVMMap[n].Memory) {
		input.VM.Registers[7] = OOB
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
	input.VM.Mem.Write(o, data)

	input.VM.Registers[7] = OK
	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// poke = 11
func poke(input OmegaInput) (output OmegaOutput) {
	n, s, o, z := input.VM.Registers[7], input.VM.Registers[8], input.VM.Registers[9], input.VM.Registers[10]
	// g = M_{O,c} + 𝒢(M_{O,ℓ}, z)
	cost := addGas(HostGasPokeConst, MemGas(HostGasPokeOctets, z))
	if result := chargeGasAndCheck(&input, cost); result != nil {
		return *result
	}

	if !input.VM.Mem.IsReadable(s, z) { // not readable, return
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if n not in K(m)
	if _, exists := input.Addition.IntegratedPVMMap[n]; !exists {
		input.VM.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise if N_o...+z not subset of \mathbf{V}_m[n]_u
	if !isWriteable(o, z, input.Addition.IntegratedPVMMap[n].Memory) { // not writeable, return
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise
	// read data from memory first
	data := input.VM.Mem.Read(s, z)
	// write data into m[n]_u
	integratedPVMType := input.Addition.IntegratedPVMMap[n]
	(&integratedPVMType.Memory).Write(o, data)
	input.VM.Registers[7] = OK

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// pagesGasCost is Ω_Z gas from B.6: free (r=0), alloc (r∈{1,2}),
// setmode (r∈{3,4}), else invalid. Linear term is per page, not MemGas.
func pagesGasCost(r, c uint64) Gas {
	switch r {
	case 0:
		return unitGasCost(HostGasPagesFreeConst, HostGasPagesFreePage, c)
	case 1, 2:
		return unitGasCost(HostGasPagesAllocConst, HostGasPagesAllocPage, c)
	case 3, 4:
		return unitGasCost(HostGasPagesSetModeConst, HostGasPagesSetModePage, c)
	default:
		return HostGasPagesInvalid
	}
}

// pages = 12
func pages(input OmegaInput) (output OmegaOutput) {
	n, p, c, r := input.VM.Registers[7], input.VM.Registers[8], input.VM.Registers[9], input.VM.Registers[10]
	if result := chargeGasAndCheck(&input, pagesGasCost(r, c)); result != nil {
		return *result
	}
	// u = panic
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		// u = panic
		input.VM.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise if p < 16 or p + c >= 2^32 / ZP or i in N_p...+c : (u_A)_i = nil
	if r > 4 || p < 16 || p+c >= (1<<32)/ZP {
		input.VM.Registers[7] = HUH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	if r > 2 && !isReadable(p, c, input.Addition.IntegratedPVMMap[n].Memory) {
		input.VM.Registers[7] = HUH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise : ok
	// u_v
	if r >= 3 {
		for i := uint32(p); i < uint32(p+c); i++ {
			input.Addition.IntegratedPVMMap[n].Memory.Pages[i] = &Page{
				Value:  make([]byte, ZP),
				Access: MemoryInaccessible,
			}
		}
	}

	// u_a
	if r == 1 || r == 3 {
		for i := uint32(p); i < uint32(p+c); i++ {
			input.Addition.IntegratedPVMMap[n].Memory.Pages[i] = &Page{
				Value:  make([]byte, ZP),
				Access: MemoryReadOnly,
			}
		}
	}

	if r == 2 || r == 4 {
		for i := uint32(p); i < uint32(p+c); i++ {
			input.Addition.IntegratedPVMMap[n].Memory.Pages[i] = &Page{
				Value:  make([]byte, ZP),
				Access: MemoryReadWrite,
			}
		}
	}

	input.VM.Registers[7] = OK

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// invoke = 13 | B.6 Ω_K: g = M_K + g_R; success path refunds g_R' to outer gas.
func invoke(input OmegaInput) (output OmegaOutput) {
	n, o := input.VM.Registers[7], input.VM.Registers[8]

	offset := uint64(112)
	// w = error ⇒ g_R = 0, g = M_K; charge then panic.
	if !input.VM.Mem.IsWriteable(o, offset) {
		if result := chargeGasAndCheck(&input, HostGasInvoke); result != nil {
			return *result
		}
		input.VM.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// assign g_R, w  |  g_R => gas , w => registers[13]   , 8(gas) + 8(uint64) * 13 = 112
	var gR uint64
	var w Registers

	// read data from memory
	data := input.VM.Mem.Read(o, offset)
	decoder := types.NewDecoder()
	 // decode gas
	err := decoder.Decode(data[:8], &gR)
	if err != nil {
		pvmLogger.Errorf("host-call function \"invoke\" decode gas error : %v", err)
	}
	// decode registers
	for i := uint64(1); i < offset/8; i++ { // start after gas used(8)
		err = decoder.Decode(data[8*i:8*(i+1)], &w[i-1])
		if err != nil {
			pvmLogger.Errorf("host-call function \"invoke\" decode register:%d error : %v", i-1, err)
		}
	}

	innerBudget := gasFromUint64(gR)
	// Charge M_K + g_R up front; remaining g_R' is refunded after Ψ returns.
	if result := chargeGasAndCheck(&input, addGas(HostGasInvoke, innerBudget)); result != nil {
		return *result
	}

	// otherwise if n not in M — no refund
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		input.VM.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	integrated := input.Addition.IntegratedPVMMap[n]
	tmpProgram, decodeReason := IntegratedProgramForInvoke(integrated)
	if decodeReason != ExitContinue {
		// Inner never ran; treat as panic host-result and refund full g_R
		// (g_R' = g_R) so outer only paid M_K.
		*input.VM.Gas += innerBudget
		input.VM.Registers[7] = INNERPANIC
		return OmegaOutput{ExitReason: ExitContinue, Addition: input.Addition}
	}

	tempMemory := integrated.Memory
	tempInterp := NewInterpreter(tmpProgram, w, &tempMemory, innerBudget)
	tempInterp.GasCharged = gasChargedForIntegratedResume(tmpProgram, integrated.PC, integrated.GasCharged)

	c, pcPrime := tempInterp.BlockBasedInvokeDecodedBlocks(integrated.PC)

	// gascounter' = gascounter − g + g_R'  (refund remaining inner gas)
	*input.VM.Gas += tempInterp.Gas

	// mu* = mu — same fixed 8-byte little-endian layout as the read path above.
	data = types.ByteSequence(make([]byte, offset))
	binary.LittleEndian.PutUint64(data[0:8], uint64(tempInterp.Gas))
	for i := uint64(1); i < offset/8; i++ {
		binary.LittleEndian.PutUint64(data[8*i:8*(i+1)], tempInterp.Registers[i-1])
	}
	// write data into memory (mu)
	input.VM.Mem.Write(o, data)

	// m* = m
	tmp := input.Addition.IntegratedPVMMap[n]
	tmp.Memory = *tempInterp.Memory
	tmp.GasCharged = tempInterp.GasCharged
	tmp.PC = pcPrime
	input.Addition.IntegratedPVMMap[n] = tmp

	switch c.GetReasonType() {
	case HOST_CALL:
		input.VM.Registers[7] = INNERHOST
		input.VM.Registers[8] = uint64(c.GetHostCallID())

	case PAGE_FAULT:
		input.VM.Registers[7] = INNERFAULT
		input.VM.Registers[8] = uint64(c.GetPageFaultAddress())

	case OUT_OF_GAS:
		input.VM.Registers[7] = INNEROOG

	case PANIC:
		input.VM.Registers[7] = INNERPANIC

	case HALT:
		input.VM.Registers[7] = INNERHALT

	}

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// expunge = 14
func expunge(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input, HostGasExpunge); result != nil { // M_X
		return *result
	}

	n := input.VM.Registers[7]
	// n not in K(m)
	if _, nExists := input.Addition.IntegratedPVMMap[n]; !nExists {
		input.VM.Registers[7] = WHO

		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	input.VM.Registers[7] = uint64(input.Addition.IntegratedPVMMap[n].PC)
	// m ˋ n
	delete(input.Addition.IntegratedPVMMap, n)

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}
