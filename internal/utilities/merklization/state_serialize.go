package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// key 1: Alpha
func encodeAlphaKey() types.StateKey {
	alphaWrapper := StateWrapper{StateIndex: 1}
	return alphaWrapper.StateKeyConstruct()
}

func encodeAlpha(alpha types.AuthPools) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedAlpha, err := encoder.Encode(&alpha)
	if err != nil {
		return nil
	}
	return encodedAlpha
}

// key 2: Varphi
func encodeVarphiKey() types.StateKey {
	varphiWrapper := StateWrapper{StateIndex: 2}
	return varphiWrapper.StateKeyConstruct()
}

func encodeVarphi(varphi types.AuthQueues) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedVarphi, err := encoder.Encode(&varphi)
	if err != nil {
		return nil
	}
	return encodedVarphi
}

// key 3: Beta
func encodeBetaKey() types.StateKey {
	betaWrapper := StateWrapper{StateIndex: 3}
	return betaWrapper.StateKeyConstruct()
}

func encodeBeta(beta types.Beta) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedBeta, err := encoder.Encode(&beta)
	if err != nil {
		return nil
	}

	return encodedBeta
}

// key 4: gamma
func encodeGammaKey() types.StateKey {
	gammaWrapper := StateWrapper{StateIndex: 4}
	return gammaWrapper.StateKeyConstruct()
}

func encodeGamma(gamma types.Gamma) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedGamma, err := encoder.Encode(&gamma)
	if err != nil {
		return nil
	}
	return encodedGamma
}

// key 5: phi
func encodePsiKey() types.StateKey {
	psiWrapper := StateWrapper{StateIndex: 5}
	return psiWrapper.StateKeyConstruct()
}

func encodePsi(psi types.DisputesRecords) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedPsi, err := encoder.Encode(&psi)
	if err != nil {
		return nil
	}
	return encodedPsi
}

// key 6: eta
func encodeEtaKey() types.StateKey {
	epsilonWrapper := StateWrapper{StateIndex: 6}
	return epsilonWrapper.StateKeyConstruct()
}

func encodeEta(eta types.EntropyBuffer) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedEta, err := encoder.Encode(&eta)
	if err != nil {
		return nil
	}
	return encodedEta
}

// key 7: iota
func encodeIotaKey() types.StateKey {
	iotaWrapper := StateWrapper{StateIndex: 7}
	return iotaWrapper.StateKeyConstruct()
}

func encodeIota(iota types.ValidatorsData) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedIota, err := encoder.Encode(&iota)
	if err != nil {
		return nil
	}
	return encodedIota
}

// key 8: kappa
func encodeKappaKey() types.StateKey {
	kappaWrapper := StateWrapper{StateIndex: 8}
	return kappaWrapper.StateKeyConstruct()
}

func encodeKappa(kappa types.ValidatorsData) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedKappa, err := encoder.Encode(&kappa)
	if err != nil {
		return nil
	}
	return encodedKappa
}

// key 9: lambda
func encodeLambdaKey() types.StateKey {
	lambdaWrapper := StateWrapper{StateIndex: 9}
	return lambdaWrapper.StateKeyConstruct()
}

func encodeLambda(lambda types.ValidatorsData) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedLambda, err := encoder.Encode(&lambda)
	if err != nil {
		return nil
	}
	return encodedLambda
}

// key 10: rho
func encodeRhoKey() types.StateKey {
	rhoWrapper := StateWrapper{StateIndex: 10}
	return rhoWrapper.StateKeyConstruct()
}

func encodeRho(rho types.AvailabilityAssignments) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedRho, err := encoder.Encode(&rho)
	if err != nil {
		return nil
	}
	return encodedRho
}

// key 11: tau
func encodeTauKey() types.StateKey {
	tauWrapper := StateWrapper{StateIndex: 11}
	return tauWrapper.StateKeyConstruct()
}

func encodeTau(tau types.TimeSlot) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedTau, err := encoder.Encode(&tau)
	if err != nil {
		return nil
	}
	return encodedTau
}

// key 12: chi
func encodeChiKey() types.StateKey {
	chiWrapper := StateWrapper{StateIndex: 12}
	return chiWrapper.StateKeyConstruct()
}

func encodeChi(chi types.Privileges) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedChi, err := encoder.Encode(&chi)
	if err != nil {
		return nil
	}
	return encodedChi
}

// key 13: pi
func encodePiKey() types.StateKey {
	piWrapper := StateWrapper{StateIndex: 13}
	return piWrapper.StateKeyConstruct()
}

func encodePi(pi types.Statistics) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedPi, err := encoder.Encode(&pi)
	if err != nil {
		return nil
	}
	return encodedPi
}

// key 14: theta
func encodeThetaKey() types.StateKey {
	thetaWrapper := StateWrapper{StateIndex: 14}
	return thetaWrapper.StateKeyConstruct()
}

func encodeTheta(theta types.ReadyQueue) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedTheta, err := encoder.Encode(&theta)
	if err != nil {
		return nil
	}
	return encodedTheta
}

// key 15: xi
func encodeXiKey() types.StateKey {
	xiWrapper := StateWrapper{StateIndex: 15}
	return xiWrapper.StateKeyConstruct()
}

func encodeXi(xi types.AccumulatedQueue) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedXi, err := encoder.Encode(&xi)
	if err != nil {
		return nil
	}
	return encodedXi
}

// key 16: lastaccount (theta)
func encodeLastAccKey() types.StateKey {
	lastAccountWrapper := StateWrapper{StateIndex: 16}
	return lastAccountWrapper.StateKeyConstruct()
}

// value 16: lastaccount (theta)
func encodeLastAccOut(lastAccOut types.AccumulatedServiceOutput) (output types.ByteSequence) {
	encoder := types.NewEncoder()
	encodedLastAccount, err := encoder.Encode(&lastAccOut)
	if err != nil {
		return nil
	}

	return encodedLastAccount
}

func encodeDelta1(serviceAccount types.ServiceAccount) (output types.ByteSequence) {
	encoder := types.NewEncoder()
	// a_c
	encodedCodeHash, err := encoder.Encode(&serviceAccount.ServiceInfo.CodeHash)
	if err != nil {
		return nil
	}
	output = append(output, encodedCodeHash...)

	// a_b
	encodedBalance, err := encoder.Encode(&serviceAccount.ServiceInfo.Balance)
	if err != nil {
		return nil
	}
	output = append(output, encodedBalance...)

	// a_g
	encodedMinItemGas, err := encoder.Encode(&serviceAccount.ServiceInfo.MinItemGas)
	if err != nil {
		return nil
	}
	output = append(output, encodedMinItemGas...)

	// a_m
	encodedMinMemoGas, err := encoder.Encode(&serviceAccount.ServiceInfo.MinMemoGas)
	if err != nil {
		return nil
	}
	output = append(output, encodedMinMemoGas...)

	// a_o
	encodedBytes, err := encoder.Encode(&serviceAccount.ServiceInfo.Bytes)
	if err != nil {
		return nil
	}
	output = append(output, encodedBytes...)

	// a_f
	encodedGratisStorageOffset, err := encoder.Encode(&serviceAccount.ServiceInfo.GratisStorageOffset)
	if err != nil {
		return nil
	}
	output = append(output, encodedGratisStorageOffset...)

	// a_i
	encodedItems, err := encoder.Encode(&serviceAccount.ServiceInfo.Items)
	if err != nil {
		return nil
	}
	output = append(output, encodedItems...)

	// a_r
	encodedCreateTime, err := encoder.Encode(&serviceAccount.ServiceInfo.CreateTime)
	if err != nil {
		return nil
	}
	output = append(output, encodedCreateTime...)

	// a_a
	encodedRecentAccumulateTime, err := encoder.Encode(&serviceAccount.ServiceInfo.RecentAccumulateTime)
	if err != nil {
		return nil
	}
	output = append(output, encodedRecentAccumulateTime...)

	// a_p
	encodedParentService, err := encoder.Encode(&serviceAccount.ServiceInfo.ParentService)
	if err != nil {
		return nil
	}
	output = append(output, encodedParentService...)

	return output
}

func encodeDelta1KeyVal(id types.ServiceId, delta types.ServiceAccount) (stateKeyVal types.StateKeyVal) {
	serviceWrapper := StateServiceWrapper{StateIndex: 255, ServiceIndex: types.ServiceId(id)}
	stateKeyVal = types.StateKeyVal{
		Key:   serviceWrapper.StateKeyConstruct(),
		Value: encodeDelta1(delta),
	}
	return stateKeyVal
}

func encodeDelta2KeyVal(id types.ServiceId, key types.ByteSequence, value types.ByteSequence) (stateKeyVal types.StateKeyVal) {
	encoder := types.NewEncoder()

	encodeLength := 4
	part_1, _ := encoder.EncodeUintWithLength((1<<32 - 1), encodeLength)
	part_2 := key

	h := types.ByteSequence{}
	copy(h, part_1)
	copy(h[encodeLength:], part_2[:])

	serviceWrapper := ServiceWrapper{ServiceIndex: types.ServiceId(id), h: h}
	stateKeyVal = types.StateKeyVal{
		Key:   serviceWrapper.StateKeyConstruct(),
		Value: value,
	}

	return stateKeyVal
}

func encodeDelta3KeyVal(id types.ServiceId, key types.OpaqueHash, value types.ByteSequence) (stateKeyVal types.StateKeyVal) {
	encoder := types.NewEncoder()

	encodeLength := 4
	part_1, _ := encoder.EncodeUintWithLength((1<<32 - 2), encodeLength)
	part_2 := key

	h := types.ByteSequence{}
	copy(h, part_1)
	copy(h[encodeLength:], part_2[:])

	serviceWrapper := ServiceWrapper{ServiceIndex: types.ServiceId(id), h: h}
	stateKeyVal = types.StateKeyVal{
		Key:   serviceWrapper.StateKeyConstruct(),
		Value: value,
	}

	return stateKeyVal
}

func encodeDelta4KeyVal(id types.ServiceId, key types.LookupMetaMapkey, value types.TimeSlotSet) (stateKeyVal types.StateKeyVal) {
	encoder := types.NewEncoder()

	encodeLength := 4
	part_1, _ := encoder.EncodeUintWithLength(uint64(key.Length), encodeLength)
	hash := hash.Blake2bHash(types.ByteSequence(key.Hash[:]))
	part_2 := hash[2:25]

	h := types.ByteSequence{}
	copy(h, part_1)
	copy(h[encodeLength:], part_2[:])

	serviceWrapper := ServiceWrapper{ServiceIndex: types.ServiceId(id), h: h}

	stateValue := types.ByteSequence{}
	for _, timeSlot := range value {
		stateValue = append(stateValue, utilities.SerializeFixedLength(types.U64(timeSlot), 4)...)
	}
	stateValue = append(utilities.SerializeU64(types.U64(len(value))), stateValue...)

	stateKeyVal = types.StateKeyVal{
		Key:   serviceWrapper.StateKeyConstruct(),
		Value: stateValue,
	}

	return stateKeyVal
}

// (D.2) T(σ)
func StateEncoder(state types.State) (types.StateKeyVals, error) {
	encoded := types.StateKeyVals{}

	// key 1: alpha
	keyval1 := types.StateKeyVal{
		Key:   encodeAlphaKey(),
		Value: encodeAlpha(state.Alpha),
	}
	encoded = append(encoded, keyval1)

	// key 2: varphi
	keyval2 := types.StateKeyVal{
		Key:   encodeVarphiKey(),
		Value: encodeVarphi(state.Varphi),
	}
	encoded = append(encoded, keyval2)

	// key 3: beta
	keyval3 := types.StateKeyVal{
		Key:   encodeBetaKey(),
		Value: encodeBeta(state.Beta),
	}
	encoded = append(encoded, keyval3)

	// key 4: gamma
	keyval4 := types.StateKeyVal{
		Key:   encodeGammaKey(),
		Value: encodeGamma(state.Gamma),
	}
	encoded = append(encoded, keyval4)

	// key 5: psi
	keyval5 := types.StateKeyVal{
		Key:   encodePsiKey(),
		Value: encodePsi(state.Psi),
	}
	encoded = append(encoded, keyval5)

	// key 6: eta
	keyval6 := types.StateKeyVal{
		Key:   encodeEtaKey(),
		Value: encodeEta(state.Eta),
	}
	encoded = append(encoded, keyval6)

	// key 7: iota
	keyval7 := types.StateKeyVal{
		Key:   encodeIotaKey(),
		Value: encodeIota(state.Iota),
	}
	encoded = append(encoded, keyval7)

	// key 8: kappa
	keyval8 := types.StateKeyVal{
		Key:   encodeKappaKey(),
		Value: encodeKappa(state.Kappa),
	}
	encoded = append(encoded, keyval8)

	// key 9: lambda
	keyval9 := types.StateKeyVal{
		Key:   encodeLambdaKey(),
		Value: encodeLambda(state.Lambda),
	}
	encoded = append(encoded, keyval9)

	// key 10: rho
	keyval10 := types.StateKeyVal{
		Key:   encodeRhoKey(),
		Value: encodeRho(state.Rho),
	}
	encoded = append(encoded, keyval10)

	// key 11: tau
	keyval11 := types.StateKeyVal{
		Key:   encodeTauKey(),
		Value: encodeTau(state.Tau),
	}
	encoded = append(encoded, keyval11)

	// key 12: chi
	keyval12 := types.StateKeyVal{
		Key:   encodeChiKey(),
		Value: encodeChi(state.Chi),
	}
	encoded = append(encoded, keyval12)

	// key 13: pi
	keyval13 := types.StateKeyVal{
		Key:   encodePiKey(),
		Value: encodePi(state.Pi),
	}
	encoded = append(encoded, keyval13)

	// key 14: theta
	keyval14 := types.StateKeyVal{
		Key:   encodeThetaKey(),
		Value: encodeTheta(state.Theta),
	}
	encoded = append(encoded, keyval14)

	// key 15: xi
	keyval15 := types.StateKeyVal{
		Key:   encodeXiKey(),
		Value: encodeXi(state.Xi),
	}
	encoded = append(encoded, keyval15)

	// key 16 theta lastaccout
	// TODO: rename LastAccOut to Theta, and Theta to Vartheta
	keyval16 := types.StateKeyVal{
		Key:   encodeLastAccKey(),
		Value: encodeLastAccOut(state.LastAccOut),
	}
	encoded = append(encoded, keyval16)

	// delta 1
	for id, account := range state.Delta {
		stateKeyVal := encodeDelta1KeyVal(id, account)
		encoded = append(encoded, stateKeyVal)
	}

	// delta 2
	for id, account := range state.Delta {
		for key, val := range account.StorageDict {
			stateKeyVal := encodeDelta2KeyVal(id, types.ByteSequence(key), val)
			encoded = append(encoded, stateKeyVal)
		}
	}

	// delta 3
	for id, account := range state.Delta {
		for key, val := range account.PreimageLookup {
			stateKeyVal := encodeDelta3KeyVal(id, key, val)
			encoded = append(encoded, stateKeyVal)
		}
	}

	// delta 4
	for id, account := range state.Delta {
		for key, val := range account.LookupDict {
			stateKeyVal := encodeDelta4KeyVal(id, key, val)
			encoded = append(encoded, stateKeyVal)
		}
	}

	return encoded, nil
}
