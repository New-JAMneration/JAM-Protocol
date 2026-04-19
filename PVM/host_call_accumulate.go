package PVM

import (
	"bytes"

	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// bless = 14
func bless(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	m, a, v := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9]
	r, o, n := input.Interpreter.Registers[10], input.Interpreter.Registers[11], input.Interpreter.Registers[12]

	// if N_{a...+4C} not readable
	offset := uint64(4 * types.CoresCount)
	if !isReadable(a, offset, *input.Interpreter.Memory) {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// \mathbb{a}
	rawData := input.Interpreter.Memory.Read(a, offset)
	var assignData types.ServiceIDList
	decoder := types.NewDecoder()
	assignErr := decoder.Decode(rawData, &assignData)
	if assignErr != nil {
		pvmLogger.Errorf("host-call function \"bless\" decode assignData error : %v", assignErr)
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	offset = uint64(12 * n)
	if !isReadable(o, offset, *input.Interpreter.Memory) { // not readable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// read data from memory, might cross many pages
	rawData = input.Interpreter.Memory.Read(o, offset)

	// s -> g this will update into (x_u)_x => partialState.Chi_g, decode rawData
	alwaysAccum := make(types.AlwaysAccumulateMap)
	var accumErr error
	for len(rawData) > 0 {
		var alwaysAccumServiceID types.ServiceID
		var alwaysAccumServiceGas types.Gas
		alwaysAccumRawData := rawData[:12]
		accumErr = decoder.Decode(alwaysAccumRawData[:4], &alwaysAccumServiceID)
		if accumErr != nil {
			pvmLogger.Errorf("host-call function \"bless\" decode alwaysAccum error : %v", accumErr)
			return OmegaOutput{
				ExitReason: ExitPanic,
				Addition:   input.Addition,
			}
		}
		accumErr = decoder.Decode(alwaysAccumRawData[4:], &alwaysAccumServiceGas)
		if accumErr != nil {
			pvmLogger.Errorf("host-call function \"bless\" decode alwaysAccum error : %v", accumErr)
			return OmegaOutput{
				ExitReason: ExitPanic,
				Addition:   input.Addition,
			}
		}
		rawData = rawData[12:]
		alwaysAccum[alwaysAccumServiceID] = alwaysAccumServiceGas
	}

	if accumErr != nil {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// (m, v, r) \not in N_s
	limit := uint64(1 << 32)

	if m >= limit || v >= limit || r >= limit {
		input.Interpreter.Registers[7] = WHO

		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}
	// otherwise
	input.Interpreter.Registers[7] = OK
	input.Addition.ResultContextX.PartialState.Bless = types.ServiceID(m)
	input.Addition.ResultContextX.PartialState.Assign = assignData
	input.Addition.ResultContextX.PartialState.Designate = types.ServiceID(v)
	input.Addition.ResultContextX.PartialState.CreateAcct = types.ServiceID(r)
	input.Addition.ResultContextX.PartialState.AlwaysAccum = alwaysAccum

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// assign = 15
func assign(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	c, o, a := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9]

	offset := uint64(32 * types.AuthQueueSize)
	if !isReadable(o, offset, *input.Interpreter.Memory) { // not readable, panic
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if c >= C
	if c >= uint64(types.CoresCount) {
		input.Interpreter.Registers[7] = CORE
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise if x_s ≠ (x_u)_a[c]
	if input.Addition.ResultContextX.ServiceID != input.Addition.ResultContextX.PartialState.Assign[c] {
		input.Interpreter.Registers[7] = HUH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	if a >= (1 << 32) {
		input.Interpreter.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	rawData := input.Interpreter.Memory.Read(o, offset)

	// decode rawData , authQueue = mathbb{q}
	authQueue := types.AuthQueue{}
	decoder := types.NewDecoder()
	err := decoder.Decode(rawData, &authQueue)
	if err != nil {
		pvmLogger.Errorf("host-call function \"assign\" decode error : %v", err)
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	input.Addition.ResultContextX.PartialState.Authorizers[c] = authQueue
	input.Addition.ResultContextX.PartialState.Assign[c] = types.ServiceID(a)
	input.Interpreter.Registers[7] = OK

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// designate = 16
func designate(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	o := input.Interpreter.Registers[7]

	offset := uint64(336 * types.ValidatorsCount)
	if !isReadable(o, offset, *input.Interpreter.Memory) { // not readable, panic
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if x_s ≠ (x_u)_v
	if input.Addition.ResultContextX.ServiceID != input.Addition.ResultContextX.PartialState.Designate {
		input.Interpreter.Registers[7] = HUH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// 336 * types.ValidatorsCount might cross many pages
	rawData := input.Interpreter.Memory.Read(o, offset) // bold{v}

	validatorsData := types.ValidatorsData{}
	decoder := types.NewDecoder()
	err := decoder.Decode(rawData, &validatorsData)
	if err != nil {
		pvmLogger.Errorf("host-call function \"designate\" decode validatorsData error : %v", err)
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	input.Addition.ResultContextX.PartialState.ValidatorKeys = validatorsData
	input.Interpreter.Registers[7] = OK

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// checkpoint = 17
func checkpoint(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	input.Addition.ResultContextY = input.Addition.ResultContextX.DeepCopy()

	input.Interpreter.Registers[7] = uint64(input.Interpreter.Gas)

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// new = 18
func new(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	o, l, g, m, f, i := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9], input.Interpreter.Registers[10], input.Interpreter.Registers[11], input.Interpreter.Registers[12]
	offset := uint64(32)
	// if c = ∇
	if !(isReadable(o, offset, *input.Interpreter.Memory) && l < (1<<32)) { // not readable, return
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	// otherwise if f ≠ 0 and x_s ≠ (x_u)_m
	if f != 0 && input.Addition.ResultContextX.ServiceID != input.Addition.ResultContextY.PartialState.Bless {
		input.Interpreter.Registers[7] = HUH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	c := input.Interpreter.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceID
	s := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]

	// new an account
	a := types.ServiceAccount{
		ServiceInfo: types.ServiceInfo{
			CodeHash:             types.OpaqueHash(c),                     // c
			Balance:              0,                                       // b, will be updated later
			MinItemGas:           types.Gas(g),                            // g
			MinMemoGas:           types.Gas(m),                            // m
			CreationSlot:         input.Addition.AccumulateArgs.Timeslot,  // r
			DepositOffset:        types.U64(0),                            // f
			LastAccumulationSlot: types.TimeSlot(0),                       // a
			ParentService:        input.Addition.ResultContextX.ServiceID, // p
		},
		PreimageLookup: types.PreimagesMapEntry{}, // p
		LookupDict: types.LookupMetaMapEntry{ // l
			types.LookupMetaMapkey{
				Hash:   types.OpaqueHash(c),
				Length: types.U32(l),
			}: types.TimeSlotSet{},
		},
		StorageDict: types.Storage{}, // s
	}

	derive := service_account.GetServiceAccountDerivatives(a)
	at := derive.Minbalance
	a.ServiceInfo.Items = derive.Items
	a.ServiceInfo.Bytes = derive.Bytes
	a.ServiceInfo.Balance = at
	// s_b = (x_s)_b - at
	newBalance := s.ServiceInfo.Balance - at
	// otherwise if s_b < (x_s)_t, transfer a_t tokens to new service, so need to check balance(b) > minBalance()
	minBalance := service_account.CalcThresholdBalance(s.ServiceInfo.Items, s.ServiceInfo.Bytes, s.ServiceInfo.DepositOffset)
	if newBalance < minBalance {
		input.Interpreter.Registers[7] = CASH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise if x_s = (x_e)r and i < S and i \in K((x_e)_d)
	_, exists := input.Addition.ResultContextX.PartialState.ServiceAccounts[types.ServiceID(i)]
	if serviceID == input.Addition.ResultContextX.PartialState.CreateAcct && i < types.MinimumServiceIndex && exists {
		input.Interpreter.Registers[7] = FULL

		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// the remaining condition will new a service, so pre-update service info
	s.ServiceInfo.Balance = newBalance

	// otherwise if x_s = (x_e)_r and i < S
	if serviceID == input.Addition.ResultContextX.PartialState.CreateAcct && i < types.MinimumServiceIndex {
		// reg[7] = i
		input.Interpreter.Registers[7] = i
		// d = { (i -> a) }
		input.Addition.ResultContextX.PartialState.ServiceAccounts[types.ServiceID(i)] = a
		// d = { (x_s -> s) }
		input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = s
		if serviceID == *input.Addition.GeneralArgs.ServiceID { // update general args
			(*input.Addition.GeneralArgs.ServiceAccountState)[serviceID] = s
			*input.Addition.GeneralArgs.ServiceAccount = s
		}
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise
	importServiceID := input.Addition.ResultContextX.ImportServiceID

	// reg[7] = x_i
	input.Interpreter.Registers[7] = uint64(importServiceID)
	// i* = check(i)
	iStar := check(types.MinimumServiceIndex+(importServiceID-types.MinimumServiceIndex+42)%(1<<32-types.MinimumServiceIndex-(1<<8)), input.Addition.ResultContextX.PartialState.ServiceAccounts)
	input.Addition.ResultContextX.ImportServiceID = iStar
	// mathbb{d} : x_i -> a
	input.Addition.ResultContextX.PartialState.ServiceAccounts[importServiceID] = a
	// mathbb{d} : x_s -> s
	input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = s
	if serviceID == *input.Addition.GeneralArgs.ServiceID { // update general args
		(*input.Addition.GeneralArgs.ServiceAccountState)[serviceID] = s
		*input.Addition.GeneralArgs.ServiceAccount = s
	}
	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// upgrade = 19
func upgrade(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	o, g, m := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9]

	offset := uint64(32)
	if !isReadable(o, offset, *input.Interpreter.Memory) { // not readable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	c := input.Interpreter.Memory.Read(o, offset)

	input.Interpreter.Registers[7] = OK

	serviceID := input.Addition.ResultContextX.ServiceID
	// x_bold{s} = (x_u)_d[x_s]
	if serviceAccount, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountExists {
		serviceAccount.ServiceInfo.CodeHash = types.OpaqueHash(c)
		serviceAccount.ServiceInfo.MinItemGas = types.Gas(g)
		serviceAccount.ServiceInfo.MinMemoGas = types.Gas(m)
		input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = serviceAccount
		(*input.Addition.GeneralArgs.ServiceAccountState)[serviceID] = serviceAccount // update general args
		*input.Addition.GeneralArgs.ServiceAccount = serviceAccount
	} else {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		pvmLogger.Debugf("host-call function \"upgrade\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// transfer = 20
func transfer(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	d, a, l, o := input.Interpreter.Registers[7], input.Interpreter.Registers[8], input.Interpreter.Registers[9], input.Interpreter.Registers[10]
	if !isReadable(o, uint64(types.TransferMemoSize), *input.Interpreter.Memory) { // not readable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}
	// m
	rawData := input.Interpreter.Memory.Read(o, types.TransferMemoSize)
	if accountD, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[types.ServiceID(d)]; !accountExists {
		// not exist
		input.Interpreter.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	} else if l < uint64(accountD.ServiceInfo.MinMemoGas) {
		input.Interpreter.Registers[7] = LOW
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}
	serviceID := input.Addition.ResultContextX.ServiceID
	if accountS, accountSExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountSExists {
		b := accountS.ServiceInfo.Balance - types.U64(a) // b = (x_s)_b - a
		minBalance := service_account.CalcThresholdBalance(accountS.ServiceInfo.Items, accountS.ServiceInfo.Bytes, accountS.ServiceInfo.DepositOffset)
		if b < types.U64(minBalance) || accountS.ServiceInfo.Balance < types.U64(a) { //  check b underflow
			input.Interpreter.Registers[7] = CASH
			return OmegaOutput{
				ExitReason: ExitContinue,
				Addition:   input.Addition,
			}
		}

		t := types.DeferredTransfer{
			SenderID:   serviceID,
			ReceiverID: types.ServiceID(d),
			Balance:    types.U64(a),
			Memo:       [128]byte(rawData),
			GasLimit:   types.Gas(l),
		}

		accountS.ServiceInfo.Balance = b
		input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = accountS
		(*input.Addition.GeneralArgs.ServiceAccountState)[serviceID] = accountS // update general
		*input.Addition.GeneralArgs.ServiceAccount = accountS
		input.Addition.ResultContextX.DeferredTransfers = append(input.Addition.ResultContextX.DeferredTransfers, t)
	} else {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		pvmLogger.Debugf("host-call function \"transfer\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	// l = reg[9]
	if uint64(input.Interpreter.Gas) < l {
		input.Interpreter.Gas = 0
		return OmegaOutput{
			ExitReason: ExitOOG,
			Addition:   input.Addition,
		}
	}
	input.Interpreter.Gas -= Gas(l)
	input.Interpreter.Registers[7] = OK
	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// eject = 21
func eject(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	d, o := input.Interpreter.Registers[7], input.Interpreter.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, *input.Interpreter.Memory) { // not readable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	h := input.Interpreter.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceID

	accountD, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[types.ServiceID(d)]
	if !(types.ServiceID(d) != serviceID && accountExists) {
		// bold{d} = panic => CONTINUE, WHO
		input.Interpreter.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// else : d = account
	serviceIDSerialized := utils.SerializeFixedLength(types.U32(serviceID), types.U32(32))
	if !bytes.Equal(accountD.ServiceInfo.CodeHash[:], serviceIDSerialized) {
		// d_c not equal E_32(x_s)
		input.Interpreter.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	l := max(81, accountD.ServiceInfo.Bytes) - 81 // a_o

	lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(l)} // x_bold{s}_l
	lookupData, lookupDataExists := accountD.LookupDict[lookupKey]

	if accountD.ServiceInfo.Items != 2 || !lookupDataExists {
		input.Interpreter.Registers[7] = HUH

		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	timeslot := input.Addition.Timeslot
	lookupDataLength := len(lookupData)

	if lookupDataLength == 2 {
		if int(lookupData[1]) < int(timeslot)-int(types.TimeSlot(types.UnreferencedPreimageTimeslots)) {
			if accountS, accountSExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountSExists {

				accountS.ServiceInfo.Balance += accountD.ServiceInfo.Balance // s'_b
				input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = accountS
				(*input.Addition.GeneralArgs.ServiceAccountState)[serviceID] = accountS // update general
				*input.Addition.GeneralArgs.ServiceAccount = accountS
				delete(input.Addition.ResultContextX.PartialState.ServiceAccounts, types.ServiceID(d))
				input.Interpreter.Registers[7] = OK

				return OmegaOutput{
					ExitReason: ExitContinue,
					Addition:   input.Addition,
				}
			}
			// according GP, no need to check the service exists => it should in ServiceAccountState
			pvmLogger.Debugf("host-call function \"eject\" serviceID : %d not in ServiceAccount state", serviceID)
		}
	}

	input.Interpreter.Registers[7] = HUH

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// query = 22
func query(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	o, z := input.Interpreter.Registers[7], input.Interpreter.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, *input.Interpreter.Memory) { // not readable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	h := input.Interpreter.Memory.Read(o, offset)

	serviceID := input.Addition.ResultContextX.ServiceID
	// x_bold{s} = (x_u)_d[x_s]
	account, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]
	if !accountExists {
		// according GP, no need to check the service exists => it should in ServiceAccountState
		pvmLogger.Debugf("host-call function \"query\" serviceID : %d not in ServiceAccount state", serviceID)
	}
	lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)} // x_bold{s}_l
	var timeSlotSet types.TimeSlotSet
	lookupTimeSlotSet := getLookupItemFromKeyVal(input.Addition.ResultContextX.StorageKeyVal, serviceID, lookupKey)
	if lookupTimeSlotSet != nil {
		decoder := types.NewDecoder()
		err := decoder.Decode(lookupTimeSlotSet, &timeSlotSet)
		if err != nil {
			return OmegaOutput{
				ExitReason: ExitPanic,
				Addition:   input.Addition,
			}
		}
		account.LookupDict[lookupKey] = timeSlotSet
	}

	lookupData, lookupDataExists := account.LookupDict[lookupKey]
	if lookupDataExists {
		// a = lookupData[h,z]
		switch len(lookupData) {
		case 0:
			input.Interpreter.Registers[7], input.Interpreter.Registers[8] = 0, 0
		case 1:
			input.Interpreter.Registers[7] = 1 + uint64(1<<32)*uint64(lookupData[0])
			input.Interpreter.Registers[8] = 0
		case 2:
			input.Interpreter.Registers[7] = 2 + uint64(1<<32)*uint64(lookupData[0])
			input.Interpreter.Registers[8] = uint64(lookupData[1])
		case 3:
			input.Interpreter.Registers[7] = 3 + uint64(1<<32)*uint64(lookupData[0])
			input.Interpreter.Registers[8] = uint64(lookupData[1]) + uint64(1<<32)*uint64(lookupData[2])
		}
	} else {
		// a = panic
		input.Interpreter.Registers[7] = NONE
		input.Interpreter.Registers[8] = 0

		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

func loadLookupTimeSlotSet(account *types.ServiceAccount, storageKeyVal *types.StateKeyVals, serviceID types.ServiceID, lookupKey types.LookupMetaMapkey) *OmegaOutput {
	lookupTimeSlotSet := getLookupItemFromKeyVal(storageKeyVal, serviceID, lookupKey)
	if lookupTimeSlotSet == nil {
		return nil
	}
	var timeSlotSet types.TimeSlotSet
	decoder := types.NewDecoder()
	if err := decoder.Decode(lookupTimeSlotSet, &timeSlotSet); err != nil {
		return &OmegaOutput{ExitReason: ExitPanic}
	}
	account.LookupDict[lookupKey] = timeSlotSet
	return nil
}

func handleSolicitNewLookup(account *types.ServiceAccount, lookupKey types.LookupMetaMapkey, itemFootprintItems types.U32, itemFootprintOctets types.U64, registers *Registers) *OmegaOutput {
	newFootprintItems := account.ServiceInfo.Items + itemFootprintItems
	newFootprintOctets := account.ServiceInfo.Bytes + itemFootprintOctets
	newMinBalance := service_account.CalcThresholdBalance(newFootprintItems, newFootprintOctets, account.ServiceInfo.DepositOffset)
	if account.ServiceInfo.Balance < newMinBalance {
		registers[7] = FULL
		return &OmegaOutput{ExitReason: ExitContinue}
	}
	account.LookupDict[lookupKey] = make(types.TimeSlotSet, 0)
	account.ServiceInfo.Items = newFootprintItems
	account.ServiceInfo.Bytes = newFootprintOctets
	return nil
}

func handleSolicitExistingLookup(account *types.ServiceAccount, lookupKey types.LookupMetaMapkey, lookupData types.TimeSlotSet, itemFootprintItems types.U32, itemFootprintOctets types.U64, timeslot types.TimeSlot) *OmegaOutput {
	newFootprintItems := account.ServiceInfo.Items - itemFootprintItems
	newFootprintOctets := account.ServiceInfo.Bytes - itemFootprintOctets
	lookupData = append(lookupData, timeslot)
	newItemFootprintItems, newItemFootprintOctets := service_account.CalcLookupItemfootprint(lookupKey)
	account.LookupDict[lookupKey] = lookupData
	account.ServiceInfo.Items = newFootprintItems + newItemFootprintItems
	account.ServiceInfo.Bytes = newFootprintOctets + newItemFootprintOctets
	return nil
}

func processSolicitLookupData(account *types.ServiceAccount, lookupKey types.LookupMetaMapkey, lookupData types.TimeSlotSet, lookupDataExists bool, timeslot types.TimeSlot, registers *Registers) *OmegaOutput {
	itemFootprintItems, itemFootprintOctets := service_account.CalcLookupItemfootprint(lookupKey)

	if !lookupDataExists {
		return handleSolicitNewLookup(account, lookupKey, itemFootprintItems, itemFootprintOctets, registers)
	}
	if len(lookupData) == 2 {
		return handleSolicitExistingLookup(account, lookupKey, lookupData, itemFootprintItems, itemFootprintOctets, timeslot)
	}
	registers[7] = HUH
	return &OmegaOutput{ExitReason: ExitContinue}
}

// solicit = 23
func solicit(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	o, z := input.Interpreter.Registers[7], input.Interpreter.Registers[8]
	offset := uint64(32)
	if !isReadable(o, offset, *input.Interpreter.Memory) {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	h := input.Interpreter.Memory.Read(o, offset)
	serviceID := input.Addition.ResultContextX.ServiceID
	timeslot := input.Addition.Timeslot

	a, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]
	if !accountExists {
		pvmLogger.Debugf("host-call function \"solicit\" serviceID : %d not in ServiceAccount state", serviceID)
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)}
	if result := loadLookupTimeSlotSet(&a, input.Addition.ResultContextX.StorageKeyVal, serviceID, lookupKey); result != nil {
		result.Addition = input.Addition
		return *result
	}

	lookupData, lookupDataExists := a.LookupDict[lookupKey]
	if result := processSolicitLookupData(&a, lookupKey, lookupData, lookupDataExists, timeslot, &input.Interpreter.Registers); result != nil {
		result.Addition = input.Addition
		return *result
	}

	input.Interpreter.Registers[7] = OK
	input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
	(*input.Addition.GeneralArgs.ServiceAccountState)[serviceID] = a
	*input.Addition.GeneralArgs.ServiceAccount = a

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// forget = 24
func forget(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	o, z := input.Interpreter.Registers[7], input.Interpreter.Registers[8]

	offset := uint64(32)
	if !isReadable(o, offset, *input.Interpreter.Memory) { // not readable, return
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	h := input.Interpreter.Memory.Read(o, offset)
	serviceID := input.Addition.ResultContextX.ServiceID
	timeslot := input.Addition.Timeslot
	// x_bold{s} = (x_u)_d[x_s] check service exists
	if a, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID]; accountExists {
		lookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)} // x_bold{s}_l
		// check lookupItem from key-val
		var timeSlotSet types.TimeSlotSet
		lookupTimeSlotSet := getLookupItemFromKeyVal(input.Addition.ResultContextX.StorageKeyVal, serviceID, lookupKey)
		if lookupTimeSlotSet != nil {
			decoder := types.NewDecoder()
			err := decoder.Decode(lookupTimeSlotSet, &timeSlotSet)
			if err != nil {
				return OmegaOutput{
					ExitReason: ExitPanic,
					Addition:   input.Addition,
				}
			}
			a.LookupDict[lookupKey] = timeSlotSet
		}

		if lookupData, lookupDataExists := a.LookupDict[lookupKey]; lookupDataExists {
			lookupDataLength := len(lookupData)
			itemFootprintItems, itemFootprintOctets := service_account.CalcLookupItemfootprint(lookupKey)

			newFootprintItems := a.ServiceInfo.Items
			newFootprintOctets := a.ServiceInfo.Bytes
			if lookupDataLength == 0 || (lookupDataLength == 2 && int(lookupData[1]) < int(timeslot)-int(types.UnreferencedPreimageTimeslots)) {
				// delete (h,z) from a_l
				expectedRemoveLookupKey := types.LookupMetaMapkey{Hash: types.OpaqueHash(h), Length: types.U32(z)}
				delete(a.LookupDict, expectedRemoveLookupKey) // if key not exist, delete do nothing
				// delete (h) from a_p
				delete(a.PreimageLookup, types.OpaqueHash(h))
				newFootprintItems -= itemFootprintItems
				newFootprintOctets -= itemFootprintOctets
			} else if lookupDataLength == 1 {
				newFootprintItems -= itemFootprintItems
				newFootprintOctets -= itemFootprintOctets
				// a_l[h,z] = [x,t]
				lookupData = append(lookupData, timeslot)
				a.LookupDict[lookupKey] = lookupData
				itemFootprintItems, itemFootprintOctets = service_account.CalcLookupItemfootprint(lookupKey)
				newFootprintItems += itemFootprintItems
				newFootprintOctets += itemFootprintOctets

			} else if lookupDataLength == 3 && int(lookupData[1]) < int(timeslot)-int(types.UnreferencedPreimageTimeslots) {
				newFootprintItems -= itemFootprintItems
				newFootprintOctets -= itemFootprintOctets
				// a_l[h,z] = [w,t]
				lookupData[0] = lookupData[2]
				lookupData[1] = timeslot
				lookupData = lookupData[:2]
				a.LookupDict[lookupKey] = lookupData
				itemFootprintItems, itemFootprintOctets = service_account.CalcLookupItemfootprint(lookupKey)
				newFootprintItems += itemFootprintItems
				newFootprintOctets += itemFootprintOctets
			} else { // otherwise, panic
				input.Interpreter.Registers[7] = HUH
				return OmegaOutput{
					ExitReason: ExitContinue,
					Addition:   input.Addition,
				}
			}
			// x'_s = a
			a.ServiceInfo.Items = newFootprintItems
			a.ServiceInfo.Bytes = newFootprintOctets
			input.Addition.ResultContextX.PartialState.ServiceAccounts[serviceID] = a
			(*input.Addition.GeneralArgs.ServiceAccountState)[serviceID] = a
			*input.Addition.GeneralArgs.ServiceAccount = a

			input.Interpreter.Registers[7] = OK
		} else { // otherwise : lookupData (x_s)_l[h,z] not exist
			input.Interpreter.Registers[7] = HUH
		}
	} else {
		pvmLogger.Debugf("host-call function \"forget\" serviceID : %d not in ServiceAccount state", serviceID)
	}

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// yield = 25
func yield(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	o := input.Interpreter.Registers[7]

	offset := uint64(32)
	if !isReadable(o, offset, *input.Interpreter.Memory) {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}

	h := input.Interpreter.Memory.Read(o, offset)
	opaqueHash := types.OpaqueHash(h)
	input.Addition.ResultContextX.Exception = &opaqueHash
	input.Interpreter.Registers[7] = OK

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}

// provide = 26
func provide(input OmegaInput) (output OmegaOutput) {
	if result := chargeGasAndCheck(&input); result != nil {
		return *result
	}

	o, z := input.Interpreter.Registers[8], input.Interpreter.Registers[9]
	// i = panic
	offset := uint64(z)
	if !isReadable(o, offset, *input.Interpreter.Memory) {
		input.Interpreter.Registers[7] = OOB
		return OmegaOutput{
			ExitReason: ExitPanic,
			Addition:   input.Addition,
		}
	}
	// i = mu_o...+z
	i := input.Interpreter.Memory.Read(o, z)

	// s = x_s or s = omega_7
	var s types.ServiceID
	if input.Interpreter.Registers[7] == 0xffffffffffffffff {
		s = input.Addition.ResultContextX.ServiceID
	} else {
		s = types.ServiceID(input.Interpreter.Registers[7])
	}

	// a = d[s*] or nil,  d = (x_u)_d
	account, accountExists := input.Addition.ResultContextX.PartialState.ServiceAccounts[s]
	if !accountExists {
		// otherwise if a = nil
		input.Interpreter.Registers[7] = WHO
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	lookupKey := types.LookupMetaMapkey{
		Hash:   hash.Blake2bHash(i),
		Length: types.U32(z),
	}

	// check lookupItem from key-val
	var timeSlotSet types.TimeSlotSet
	lookupTimeSlotSet := getLookupItemFromKeyVal(input.Addition.ResultContextX.StorageKeyVal, s, lookupKey)
	if lookupTimeSlotSet != nil {
		decoder := types.NewDecoder()
		err := decoder.Decode(lookupTimeSlotSet, &timeSlotSet)
		if err != nil {
			return OmegaOutput{
				ExitReason: ExitPanic,
				Addition:   input.Addition,
			}
		}
		account.LookupDict[lookupKey] = timeSlotSet
	}

	// otherwise if a_l[H(i), z] not in []
	if lookupData, lookupDataExists := account.LookupDict[lookupKey]; (lookupDataExists && len(lookupData) != 0) || !lookupDataExists {
		input.Interpreter.Registers[7] = HUH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	serviceBlob := types.ServiceBlob{
		ServiceID: s,
		Blob:      i,
	}

	encoder := types.NewEncoder()
	serialized, _ := encoder.Encode(&s)
	encoded, _ := encoder.Encode(&i)
	serialized = append(serialized, encoded...)
	hashKey := hash.Blake2bHash(serialized)

	// golang can not have slice in map key, so use hash instead
	if _, hashExists := input.Addition.ResultContextX.ServiceBlobs[hashKey]; hashExists {
		input.Interpreter.Registers[7] = HUH
		return OmegaOutput{
			ExitReason: ExitContinue,
			Addition:   input.Addition,
		}
	}

	// otherwise OK
	input.Addition.ResultContextX.ServiceBlobs[hashKey] = serviceBlob
	input.Interpreter.Registers[7] = OK

	return OmegaOutput{
		ExitReason: ExitContinue,
		Addition:   input.Addition,
	}
}
