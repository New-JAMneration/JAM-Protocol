package PVM

// Regressions for guest service ids that exceed U32: truncating to
// types.ServiceID must not alias an existing low-32-bit service.
// Mutating host-calls expect WHO; read-path host-calls expect NONE.

import (
	"math"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TestTransferRejectsHighDestAlias ensures a 64-bit destination that truncates
// to an existing ServiceID still returns WHO and does not mutate state.
//
// Without the d > MaxUint32 guard, types.ServiceID(d) would wrap and alias the
// low-32-bit service (e.g. d = 2^32+7 → service 7).
func TestTransferRejectsHighDestAlias(t *testing.T) {
	const (
		senderID = types.ServiceID(1)
		destID   = types.ServiceID(7)
		amount   = uint64(10)
		memoGas  = uint64(0)
		memoOff  = uint64(16 * ZP)
	)

	senderBal := types.U64(10_000)
	destBal := types.U64(1_000)

	sender := types.ServiceAccount{
		ServiceInfo: types.ServiceInfo{
			Balance:    senderBal,
			MinMemoGas: 0,
		},
	}
	dest := types.ServiceAccount{
		ServiceInfo: types.ServiceInfo{
			Balance:    destBal,
			MinMemoGas: 0,
		},
	}
	accounts := types.ServiceAccountState{
		senderID: sender,
		destID:   dest,
	}
	storage := types.StateKeyVals{}
	partial := types.PartialStateSet{ServiceAccounts: accounts}

	regs := Registers{}
	// d = 2^32 + 7 — truncates to ServiceID(7) if cast without a range check.
	regs[7] = uint64(math.MaxUint32) + 1 + uint64(destID)
	regs[8] = amount
	regs[9] = memoGas
	regs[10] = memoOff

	mem := &Memory{Pages: map[uint32]*Page{}}
	// Transfer memo is 128 bytes; one RW page is enough at 16*ZP.
	mem.Pages[16] = &Page{Value: make([]byte, ZP), Access: MemoryReadWrite}

	gas := Gas(HostGasTransfer + 1_000)
	acctCopy := sender
	stateCopy := accounts
	sid := senderID
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
	out := transfer(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceID:           &sid,
				ServiceAccount:      &acctCopy,
				ServiceAccountState: &stateCopy,
				StorageKeyVal:       &storage,
			},
			AccumulateArgs: AccumulateArgs{
				ResultContextX: ResultContext{
					ServiceID:         senderID,
					PartialState:      partial,
					DeferredTransfers: []types.DeferredTransfer{},
					StorageKeyVal:     &storage,
				},
			},
		},
	})

	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want Continue", out.ExitReason)
	}
	if regs[7] != WHO {
		t.Fatalf("reg7 = %d (%#x), want WHO (%#x)", regs[7], regs[7], WHO)
	}

	// State must be unchanged: no deferred transfer, balances intact.
	if n := len(out.Addition.ResultContextX.DeferredTransfers); n != 0 {
		t.Fatalf("deferred transfers = %d, want 0", n)
	}
	gotSender := out.Addition.ResultContextX.PartialState.ServiceAccounts[senderID]
	gotDest := out.Addition.ResultContextX.PartialState.ServiceAccounts[destID]
	if gotSender.ServiceInfo.Balance != senderBal {
		t.Fatalf("sender balance = %d, want unchanged %d", gotSender.ServiceInfo.Balance, senderBal)
	}
	if gotDest.ServiceInfo.Balance != destBal {
		t.Fatalf("dest balance = %d, want unchanged %d", gotDest.ServiceInfo.Balance, destBal)
	}

	// Sanity: the same call with d = 7 (in-range) should succeed — proves the
	// alias target is a real account that the truncated cast would have hit.
	regsOK := Registers{}
	regsOK[7] = uint64(destID)
	regsOK[8] = amount
	regsOK[9] = memoGas
	regsOK[10] = memoOff
	gasOK := Gas(HostGasTransfer + 1_000)
	acctOK := sender
	stateOK := types.ServiceAccountState{senderID: sender, destID: dest}
	partialOK := types.PartialStateSet{ServiceAccounts: stateOK}
	sidOK := senderID
	vmOK := &VMState{Registers: &regsOK, Gas: &gasOK, Mem: NewPagedGuestMemory(mem)}
	outOK := transfer(OmegaInput{
		VM: vmOK,
		Addition: HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceID:           &sidOK,
				ServiceAccount:      &acctOK,
				ServiceAccountState: &stateOK,
				StorageKeyVal:       &storage,
			},
			AccumulateArgs: AccumulateArgs{
				ResultContextX: ResultContext{
					ServiceID:         senderID,
					PartialState:      partialOK,
					DeferredTransfers: []types.DeferredTransfer{},
					StorageKeyVal:     &storage,
				},
			},
		},
	})
	if outOK.ExitReason != ExitContinue || regsOK[7] != OK {
		t.Fatalf("control d=%d: exit=%v reg7=%#x (want Continue/OK) — fixture may be wrong",
			destID, outOK.ExitReason, regsOK[7])
	}
	if len(outOK.Addition.ResultContextX.DeferredTransfers) != 1 {
		t.Fatalf("control: deferred transfers = %d, want 1", len(outOK.Addition.ResultContextX.DeferredTransfers))
	}
}

// TestEjectRejectsHighDestAlias: d = 2^32+7 must not delete service 7.
func TestEjectRejectsHighDestAlias(t *testing.T) {
	const (
		callerID = types.ServiceID(1)
		destID   = types.ServiceID(7)
		hashOff  = uint64(16 * ZP)
	)

	caller := types.ServiceAccount{ServiceInfo: types.ServiceInfo{Balance: 10_000}}
	dest := types.ServiceAccount{ServiceInfo: types.ServiceInfo{Balance: 1_000, Items: 2}}
	accounts := types.ServiceAccountState{callerID: caller, destID: dest}
	storage := types.StateKeyVals{}
	partial := types.PartialStateSet{ServiceAccounts: accounts}

	regs := Registers{}
	regs[7] = uint64(math.MaxUint32) + 1 + uint64(destID)
	regs[8] = hashOff

	mem := &Memory{Pages: map[uint32]*Page{}}
	mem.Pages[16] = &Page{Value: make([]byte, ZP), Access: MemoryReadWrite}

	gas := Gas(HostGasEject + 100)
	sid := callerID
	acct := caller
	state := accounts
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
	out := eject(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceID:           &sid,
				ServiceAccount:      &acct,
				ServiceAccountState: &state,
				StorageKeyVal:       &storage,
			},
			AccumulateArgs: AccumulateArgs{
				ResultContextX: ResultContext{
					ServiceID:     callerID,
					PartialState:  partial,
					StorageKeyVal: &storage,
				},
				Timeslot: 0,
			},
		},
	})

	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want Continue", out.ExitReason)
	}
	if regs[7] != WHO {
		t.Fatalf("reg7 = %#x, want WHO (%#x)", regs[7], WHO)
	}
	if _, ok := out.Addition.ResultContextX.PartialState.ServiceAccounts[destID]; !ok {
		t.Fatal("dest service was deleted; high-bit d must not alias")
	}
}

// TestProvideRejectsHighServiceAlias: ω₇ = 2^32+7 must not target service 7.
func TestProvideRejectsHighServiceAlias(t *testing.T) {
	const (
		callerID = types.ServiceID(1)
		targetID = types.ServiceID(7)
		blobOff  = uint64(16 * ZP)
		blobLen  = uint64(32)
	)

	caller := types.ServiceAccount{ServiceInfo: types.ServiceInfo{Balance: 10_000}}
	target := types.ServiceAccount{
		ServiceInfo:    types.ServiceInfo{Balance: 1_000},
		LookupDict:     types.LookupMetaMapEntry{},
		PreimageLookup: types.PreimagesMapEntry{},
	}
	accounts := types.ServiceAccountState{callerID: caller, targetID: target}
	storage := types.StateKeyVals{}
	partial := types.PartialStateSet{ServiceAccounts: accounts}

	regs := Registers{}
	regs[7] = uint64(math.MaxUint32) + 1 + uint64(targetID)
	regs[8] = blobOff
	regs[9] = blobLen

	mem := &Memory{Pages: map[uint32]*Page{}}
	mem.Pages[16] = &Page{Value: make([]byte, ZP), Access: MemoryReadWrite}

	gas := Gas(HostGasProvideConst + HostGasProvideOctets + 100)
	sid := callerID
	acct := caller
	state := accounts
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
	out := provide(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceID:           &sid,
				ServiceAccount:      &acct,
				ServiceAccountState: &state,
				StorageKeyVal:       &storage,
			},
			AccumulateArgs: AccumulateArgs{
				ResultContextX: ResultContext{
					ServiceID:     callerID,
					PartialState:  partial,
					StorageKeyVal: &storage,
				},
			},
		},
	})

	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want Continue", out.ExitReason)
	}
	if regs[7] != WHO {
		t.Fatalf("reg7 = %#x, want WHO (%#x)", regs[7], WHO)
	}
	got := out.Addition.ResultContextX.PartialState.ServiceAccounts[targetID]
	if len(got.LookupDict) != 0 || len(got.PreimageLookup) != 0 {
		t.Fatal("target service was mutated; high-bit s must not alias")
	}
}

func highServiceAliasReg(target types.ServiceID) uint64 {
	return uint64(math.MaxUint32) + 1 + uint64(target)
}

// Read-path host-calls return NONE (not WHO) when the service is missing; a
// truncated high-bit id must not resolve to an existing low-32 service.

func TestInfoRejectsHighServiceAlias(t *testing.T) {
	const (
		callerID = types.ServiceID(1)
		targetID = types.ServiceID(7)
	)
	caller := types.ServiceAccount{ServiceInfo: types.ServiceInfo{Balance: 10_000, CodeHash: types.OpaqueHash{1}}}
	target := types.ServiceAccount{ServiceInfo: types.ServiceInfo{Balance: 1_000, CodeHash: types.OpaqueHash{7}}}
	state := types.ServiceAccountState{callerID: caller, targetID: target}

	regs := Registers{}
	regs[7] = highServiceAliasReg(targetID)
	gas := Gas(HostGasInfo + 100)
	sid := callerID
	acct := caller
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(&Memory{Pages: map[uint32]*Page{}})}
	out := info(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceID:           &sid,
				ServiceAccount:      &acct,
				ServiceAccountState: &state,
			},
		},
	})
	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want Continue", out.ExitReason)
	}
	if regs[7] != NONE {
		t.Fatalf("reg7 = %#x, want NONE (%#x) — high id must not alias service %d", regs[7], NONE, targetID)
	}
}

func TestLookupRejectsHighServiceAlias(t *testing.T) {
	const (
		callerID = types.ServiceID(1)
		targetID = types.ServiceID(7)
		hashOff  = uint64(16 * ZP)
	)
	preimageKey := types.OpaqueHash{0xab}
	target := types.ServiceAccount{
		ServiceInfo:    types.ServiceInfo{Balance: 1_000},
		PreimageLookup: types.PreimagesMapEntry{preimageKey: types.ByteSequence{0xde, 0xad}},
	}
	caller := types.ServiceAccount{ServiceInfo: types.ServiceInfo{Balance: 10_000}}
	state := types.ServiceAccountState{callerID: caller, targetID: target}

	regs := Registers{}
	regs[7] = highServiceAliasReg(targetID)
	regs[8] = hashOff
	regs[9] = hashOff + 64
	regs[10] = 0
	regs[11] = 2

	mem := &Memory{Pages: map[uint32]*Page{}}
	page := &Page{Value: make([]byte, ZP), Access: MemoryReadWrite}
	copy(page.Value[:32], preimageKey[:])
	mem.Pages[16] = page

	gas := Gas(HostGasLookupConst + HostGasLookupOctets + 100)
	sid := callerID
	acct := caller
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
	out := lookup(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceID:           &sid,
				ServiceAccount:      &acct,
				ServiceAccountState: &state,
			},
		},
	})
	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want Continue", out.ExitReason)
	}
	if regs[7] != NONE {
		t.Fatalf("reg7 = %#x, want NONE (%#x)", regs[7], NONE)
	}
}

func TestReadRejectsHighServiceAlias(t *testing.T) {
	const (
		callerID = types.ServiceID(1)
		targetID = types.ServiceID(7)
		keyOff   = uint64(16 * ZP)
		keyLen   = uint64(4)
	)
	target := types.ServiceAccount{
		ServiceInfo: types.ServiceInfo{Balance: 1_000},
		StorageDict: types.Storage{"key!": []byte("value-from-7")},
	}
	caller := types.ServiceAccount{ServiceInfo: types.ServiceInfo{Balance: 10_000}}
	state := types.ServiceAccountState{callerID: caller, targetID: target}
	storageKV := types.StateKeyVals{}

	regs := Registers{}
	regs[7] = highServiceAliasReg(targetID)
	regs[8] = keyOff
	regs[9] = keyLen
	regs[10] = keyOff + 64
	regs[12] = 32

	mem := &Memory{Pages: map[uint32]*Page{}}
	page := &Page{Value: make([]byte, ZP), Access: MemoryReadWrite}
	copy(page.Value[:keyLen], []byte("key!"))
	mem.Pages[16] = page

	gas := Gas(HostGasReadConst + HostGasReadKeyOctets + HostGasReadValueOctets + 100)
	sid := callerID
	acct := caller
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
	out := read(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceID:           &sid,
				ServiceAccount:      &acct,
				ServiceAccountState: &state,
				StorageKeyVal:       &storageKV,
			},
		},
	})
	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want Continue", out.ExitReason)
	}
	if regs[7] != NONE {
		t.Fatalf("reg7 = %#x, want NONE (%#x)", regs[7], NONE)
	}
}

func TestHistoricalLookupRejectsHighServiceAlias(t *testing.T) {
	const (
		callerID = types.ServiceID(1)
		targetID = types.ServiceID(7)
		hashOff  = uint64(16 * ZP)
	)
	codeHash := types.OpaqueHash{0x11}
	target := types.ServiceAccount{
		ServiceInfo: types.ServiceInfo{Balance: 1_000},
		// Preimage present so a successful alias would return a length, not NONE.
		PreimageLookup: types.PreimagesMapEntry{codeHash: types.ByteSequence{1, 2, 3, 4}},
		LookupDict: types.LookupMetaMapEntry{
			types.LookupMetaMapkey{Hash: codeHash, Length: 4}: types.TimeSlotSet{0},
		},
	}
	caller := types.ServiceAccount{ServiceInfo: types.ServiceInfo{Balance: 10_000}}
	state := types.ServiceAccountState{callerID: caller, targetID: target}

	regs := Registers{}
	regs[7] = highServiceAliasReg(targetID)
	regs[8] = hashOff
	regs[9] = hashOff + 64
	regs[10] = 0
	regs[11] = 4

	mem := &Memory{Pages: map[uint32]*Page{}}
	page := &Page{Value: make([]byte, ZP), Access: MemoryReadWrite}
	copy(page.Value[:32], codeHash[:])
	mem.Pages[16] = page

	gas := Gas(HostGasHistoricalLookupConst + HostGasHistoricalLookupOctets + 100)
	sid := callerID
	acct := caller
	vm := &VMState{Registers: &regs, Gas: &gas, Mem: NewPagedGuestMemory(mem)}
	out := historicalLookup(OmegaInput{
		VM: vm,
		Addition: HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceID:           &sid,
				ServiceAccount:      &acct,
				ServiceAccountState: &state,
			},
			RefineArgs: RefineArgs{TimeSlot: 0},
		},
	})
	if out.ExitReason != ExitContinue {
		t.Fatalf("exit = %v, want Continue", out.ExitReason)
	}
	if regs[7] != NONE {
		t.Fatalf("reg7 = %#x, want NONE (%#x)", regs[7], NONE)
	}
}
