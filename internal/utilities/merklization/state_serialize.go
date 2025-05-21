package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// key 1: Alpha
func encodeAlphaKey() types.StateKey {
	alphaWrapper := StateWrapper{StateIndex: 1}
	return alphaWrapper.StateKeyConstruct()
}

// value 1: Alpha
func serializeAlpha(alpha types.AuthPools) (output types.ByteSequence) {
	for _, authPool := range alpha {
		output = append(output, utilities.SerializeU64(types.U64(len(authPool)))...)
		for _, authorizerHash := range authPool {
			output = append(output, utilities.SerializeByteSequence(authorizerHash[:])...)
		}
	}
	return output
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

// value 2: Varphi
func serializeVarphi(varphi types.AuthQueues) (output types.ByteSequence) {
	for _, v := range varphi {
		for _, authQueue := range v {
			output = append(output, utilities.SerializeByteSequence(authQueue[:])...)
		}
	}
	return output
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

// value 3: Beta
func serializeBeta(beta types.BlocksHistory) (output types.ByteSequence) {
	output = append(output, utilities.SerializeU64(types.U64(len(beta)))...)
	for _, blockInfo := range beta {
		// h
		output = append(output, utilities.SerializeByteSequence(blockInfo.HeaderHash[:])...)

		// b
		mmrResult := serializeFromMmrPeaks(blockInfo.Mmr)
		output = append(output, utilities.SerializeByteSequence(mmrResult)...)

		// s
		output = append(output, utilities.SerializeByteSequence(blockInfo.StateRoot[:])...)

		// p  (?) BlockInfo.Reported is different from GP, list or dict
		output = append(output, utilities.SerializeU64(types.U64(len(blockInfo.Reported)))...)
		for _, report := range blockInfo.Reported {
			output = append(output, utilities.SerializeByteSequence(report.Hash[:])...)
			output = append(output, utilities.SerializeByteSequence(report.ExportsRoot[:])...)
		}
	}
	return output
}

func encodeBeta(beta types.BlocksHistory) types.ByteSequence {
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

// value 4: gamma
func serializeGamma(gamma types.Gamma) (output types.ByteSequence) {
	// gamma_k
	for _, v := range gamma.GammaK {
		output = append(output, utilities.SerializeByteSequence(v.Bandersnatch[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Ed25519[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Bls[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Metadata[:])...)
	}

	// gamma_z
	output = append(output, utilities.SerializeByteSequence(gamma.GammaZ[:])...)
	// gamma_s (6.5)
	if gamma.GammaS.Tickets != nil {
		output = append(output, utilities.SerializeFixedLength(types.U64(0), 1)...)
		for _, ticket := range gamma.GammaS.Tickets {
			output = append(output, utilities.SerializeByteSequence(ticket.Id[:])...)
			output = append(output, utilities.SerializeFixedLength(types.U64(ticket.Attempt), 1)...)
		}
	} else if gamma.GammaS.Keys != nil {
		output = append(output, utilities.SerializeFixedLength(types.U64(1), 1)...)
		for _, key := range gamma.GammaS.Keys {
			output = append(output, utilities.SerializeByteSequence(key[:])...)
		}
	}

	// gamma_a
	output = append(output, utilities.SerializeU64(types.U64(len(gamma.GammaA)))...)
	for _, ticket := range gamma.GammaA {
		output = append(output, utilities.SerializeByteSequence(ticket.Id[:])...)
		output = append(output, utilities.SerializeFixedLength(types.U64(ticket.Attempt), 1)...)
	}
	return output
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

// value 5: phi
func serializePsi(psi types.DisputesRecords) (output types.ByteSequence) {
	// psi_g
	output = append(output, utilities.SerializeU64(types.U64(len(psi.Good)))...)
	for _, workReportHash := range psi.Good {
		output = append(output, utilities.SerializeByteSequence(workReportHash[:])...)
	}

	// psi_b
	output = append(output, utilities.SerializeU64(types.U64(len(psi.Bad)))...)
	for _, workReportHash := range psi.Bad {
		output = append(output, utilities.SerializeByteSequence(workReportHash[:])...)
	}

	// psi_w
	output = append(output, utilities.SerializeU64(types.U64(len(psi.Wonky)))...)
	for _, workReportHash := range psi.Wonky {
		output = append(output, utilities.SerializeByteSequence(workReportHash[:])...)
	}

	// psi_o
	output = append(output, utilities.SerializeU64(types.U64(len(psi.Offenders)))...)
	for _, ed25519public := range psi.Offenders {
		output = append(output, utilities.SerializeByteSequence(ed25519public[:])...)
	}
	return output
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

// value 6: eta
func serializeEta(eta types.EntropyBuffer) (output types.ByteSequence) {
	for _, entropy := range eta {
		output = append(output, utilities.SerializeByteSequence(entropy[:])...)
	}
	return output
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

// value 7: iota
func serializeIota(iota types.ValidatorsData) (output types.ByteSequence) {
	for _, v := range iota {
		output = append(output, utilities.SerializeByteSequence(v.Bandersnatch[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Ed25519[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Bls[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Metadata[:])...)
	}
	return output
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

// value 8: kappa
func serializeKappa(kappa types.ValidatorsData) (output types.ByteSequence) {
	for _, v := range kappa {
		output = append(output, utilities.SerializeByteSequence(v.Bandersnatch[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Ed25519[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Bls[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Metadata[:])...)
	}
	return output
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

// value 9: lambda
func serializeLambda(lambda types.ValidatorsData) (output types.ByteSequence) {
	for _, v := range lambda {
		output = append(output, utilities.SerializeByteSequence(v.Bandersnatch[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Ed25519[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Bls[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Metadata[:])...)
	}
	return output
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

// value 10: rho
func serializeRho(rho types.AvailabilityAssignments) (output types.ByteSequence) {
	for _, v := range rho {
		if v != nil {
			output = append(output, utilities.SerializeFixedLength(types.U64(1), 1)...)
			output = append(output, utilities.WorkReportSerialization(v.Report)...)
			output = append(output, utilities.SerializeFixedLength(types.U32(v.Timeout), 4)...)
		} else {
			output = append(output, utilities.SerializeFixedLength(types.U64(0), 1)...)
		}
	}
	return output
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

// value 11: tau
func serializeTau(tau types.TimeSlot) (output types.ByteSequence) {
	output = utilities.SerializeFixedLength(types.U32(tau), 4)
	return output
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

// value 12: chi
func serializeChi(chi types.Privileges) (output types.ByteSequence) {
	output = append(output, utilities.SerializeFixedLength(types.U32(chi.Bless), 4)...)
	output = append(output, utilities.SerializeFixedLength(types.U32(chi.Assign), 4)...)
	output = append(output, utilities.SerializeFixedLength(types.U32(chi.Designate), 4)...)
	output = append(output, utilities.SerializeU64(types.U64(len(chi.AlwaysAccum)))...)
	for k, v := range chi.AlwaysAccum {
		output = append(output, utilities.SerializeFixedLength(types.U32(k), 4)...)
		output = append(output, utilities.SerializeFixedLength(types.U64(v), 8)...)
	}
	return output
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

// value 13: pi
func serializePi(pi types.Statistics) (output types.ByteSequence) {
	for _, activityRecord := range pi.ValsCurr {
		output = append(output, utilities.SerializeFixedLength(activityRecord.Blocks, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.Tickets, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.PreImages, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.PreImagesSize, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.Guarantees, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.Assurances, 4)...)
	}
	for _, activityRecord := range pi.ValsLast {
		output = append(output, utilities.SerializeFixedLength(activityRecord.Blocks, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.Tickets, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.PreImages, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.PreImagesSize, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.Guarantees, 4)...)
		output = append(output, utilities.SerializeFixedLength(activityRecord.Assurances, 4)...)
	}
	for _, core := range pi.Cores {
		output = append(output, utilities.SerializeU64(types.U64(core.DALoad))...)
		output = append(output, utilities.SerializeU64(types.U64(core.Popularity))...)
		output = append(output, utilities.SerializeU64(types.U64(core.Imports))...)
		output = append(output, utilities.SerializeU64(types.U64(core.Exports))...)
		output = append(output, utilities.SerializeU64(types.U64(core.ExtrinsicSize))...)
		output = append(output, utilities.SerializeU64(types.U64(core.ExtrinsicCount))...)
		output = append(output, utilities.SerializeU64(types.U64(core.BundleSize))...)
		output = append(output, utilities.SerializeU64(types.U64(core.GasUsed))...)
	}
	serPiS := utilities.WrapStatisticsServiceMap(pi.Services)
	output = append(output, serPiS.Serialize()...)
	return output
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

// value 14: theta
func serializeTheta(theta types.ReadyQueue) (output types.ByteSequence) {
	for _, unaccumulateWorkReports := range theta {
		output = append(output, utilities.SerializeU64(types.U64(len(unaccumulateWorkReports)))...)
		for _, v := range unaccumulateWorkReports {
			output = append(output, utilities.WorkReportSerialization(v.Report)...)
			output = append(output, utilities.SerializeU64(types.U64(len(v.Dependencies)))...)
			for _, dep := range v.Dependencies {
				output = append(output, utilities.SerializeByteSequence(dep[:])...)
			}
		}
	}
	return output
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

// value 15: xi
func serializeXi(xi types.AccumulatedQueue) (output types.ByteSequence) {
	for _, v := range xi {
		output = append(output, utilities.SerializeU64(types.U64(len(v)))...)
		for _, history := range v {
			output = append(output, utilities.SerializeByteSequence(history[:])...)
		}
	}
	return output
}

func encodeXi(xi types.AccumulatedQueue) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedXi, err := encoder.Encode(&xi)
	if err != nil {
		return nil
	}
	return encodedXi
}

// value 16: delta1
func serializeDelta1(s types.ServiceAccount) (output types.ByteSequence) {
	// a_c
	delta1Output := types.ByteSequence(s.ServiceInfo.CodeHash[:])

	// a_b, a_g, a_m
	delta1Output = append(delta1Output, utilities.SerializeFixedLength(types.U64(s.ServiceInfo.Balance), 8)...)
	delta1Output = append(delta1Output, utilities.SerializeFixedLength(types.U64(s.ServiceInfo.MinItemGas), 8)...)
	delta1Output = append(delta1Output, utilities.SerializeFixedLength(types.U64(s.ServiceInfo.MinMemoGas), 8)...)

	// a_l (U64)
	delta1Output = append(delta1Output, utilities.SerializeFixedLength(calculateItemNumbers(s), 8)...)

	// a_i
	delta1Output = append(delta1Output, utilities.SerializeFixedLength(calculateOctectsNumbers(s), 4)...)

	return delta1Output
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
	a_o := service_account.CalcOctets(serviceAccount)
	encodedOctets, err := encoder.Encode(&a_o)
	if err != nil {
		return nil
	}
	output = append(output, encodedOctets...)

	// a_i
	a_i := service_account.CalcKeys(serviceAccount)
	encodedKeys, err := encoder.Encode(&a_i)
	if err != nil {
		return nil
	}
	output = append(output, encodedKeys...)
	return output
}

// key 16: delta1
func serializeDelta1KeyVal(id types.ServiceId, delta types.ServiceAccount) (key16 types.StateKey, delta1Output types.ByteSequence) {
	serviceWrapper := StateServiceWrapper{StateIndex: 255, ServiceIndex: types.ServiceId(id)}
	key16 = serviceWrapper.StateKeyConstruct()
	delta1Output = serializeDelta1(delta)
	return key16, delta1Output
}

func encodeDelta1KeyVal(id types.ServiceId, delta types.ServiceAccount) (stateKeyVal types.StateKeyVal) {
	serviceWrapper := StateServiceWrapper{StateIndex: 255, ServiceIndex: types.ServiceId(id)}
	stateKeyVal = types.StateKeyVal{
		Key:   serviceWrapper.StateKeyConstruct(),
		Value: encodeDelta1(delta),
	}
	return stateKeyVal
}

// key 17: delta2
func serializeDelta2KeyVal(id types.ServiceId, key types.OpaqueHash, value types.ByteSequence) (key17 types.StateKey, delta2Output types.ByteSequence) {
	encoder := types.NewEncoder()

	encodeLength := 4
	part_1, _ := encoder.EncodeUintWithLength((1<<32 - 1), encodeLength)
	part_2 := key[:23]

	h := [27]byte{}
	copy(h[:], part_1)
	copy(h[encodeLength:], part_2)

	serviceWrapper := ServiceWrapper{ServiceIndex: types.ServiceId(id), h: h}
	key17 = serviceWrapper.StateKeyConstruct()
	delta2Output = value

	return key17, delta2Output
}

func encodeDelta2KeyVal(id types.ServiceId, key types.OpaqueHash, value types.ByteSequence) (stateKeyVal types.StateKeyVal) {
	encoder := types.NewEncoder()

	encodeLength := 4
	part_1, _ := encoder.EncodeUintWithLength((1<<32 - 1), encodeLength)
	part_2 := key[:23]

	h := [27]byte{}
	copy(h[:], part_1)
	copy(h[encodeLength:], part_2)

	serviceWrapper := ServiceWrapper{ServiceIndex: types.ServiceId(id), h: h}
	stateKeyVal = types.StateKeyVal{
		Key:   serviceWrapper.StateKeyConstruct(),
		Value: value,
	}

	return stateKeyVal
}

// key 18: delta3
func serializeDelta3KeyVal(id types.ServiceId, key types.OpaqueHash, value types.ByteSequence) (key18 types.StateKey, delta3Output types.ByteSequence) {
	encoder := types.NewEncoder()

	encodeLength := 4
	part_1, _ := encoder.EncodeUintWithLength((1<<32 - 2), encodeLength)
	part_2 := key[1:24]

	h := [27]byte{}
	copy(h[:], part_1)
	copy(h[encodeLength:], part_2)

	serviceWrapper := ServiceWrapper{ServiceIndex: types.ServiceId(id), h: h}
	key18 = serviceWrapper.StateKeyConstruct()
	delta3Output = value

	return key18, delta3Output
}

func encodeDelta3KeyVal(id types.ServiceId, key types.OpaqueHash, value types.ByteSequence) (stateKeyVal types.StateKeyVal) {
	encoder := types.NewEncoder()

	encodeLength := 4
	part_1, _ := encoder.EncodeUintWithLength((1<<32 - 2), encodeLength)
	part_2 := key[1:24]

	h := [27]byte{}
	copy(h[:], part_1)
	copy(h[encodeLength:], part_2)

	serviceWrapper := ServiceWrapper{ServiceIndex: types.ServiceId(id), h: h}
	stateKeyVal = types.StateKeyVal{
		Key:   serviceWrapper.StateKeyConstruct(),
		Value: value,
	}

	return stateKeyVal
}

// key 19: delta4
func serializeDelta4KeyVal(id types.ServiceId, key types.LookupMetaMapkey, value types.TimeSlotSet) (key19 types.StateKey, delta4Output types.ByteSequence) {
	encoder := types.NewEncoder()

	encodeLength := 4
	part_1, _ := encoder.EncodeUintWithLength(uint64(key.Length), encodeLength)
	hash := hash.Blake2bHash(types.ByteSequence(key.Hash[:]))
	part_2 := hash[2:25]

	h := [27]byte{}
	copy(h[:], part_1)
	copy(h[encodeLength:], part_2)

	serviceWrapper := ServiceWrapper{ServiceIndex: types.ServiceId(id), h: h}
	key19 = serviceWrapper.StateKeyConstruct()

	for _, timeSlot := range value {
		delta4Output = append(delta4Output, utilities.SerializeFixedLength(types.U64(timeSlot), 4)...)
	}

	delta4Output = append(utilities.SerializeU64(types.U64(len(value))), delta4Output...)

	return key19, delta4Output
}

func encodeDelta4KeyVal(id types.ServiceId, key types.LookupMetaMapkey, value types.TimeSlotSet) (stateKeyVal types.StateKeyVal) {
	encoder := types.NewEncoder()

	encodeLength := 4
	part_1, _ := encoder.EncodeUintWithLength(uint64(key.Length), encodeLength)
	hash := hash.Blake2bHash(types.ByteSequence(key.Hash[:]))
	part_2 := hash[2:25]

	h := [27]byte{}
	copy(h[:], part_1)
	copy(h[encodeLength:], part_2)

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

func StateSerialize(state types.State) (map[types.StateKey]types.ByteSequence, error) {
	serialized := make(map[types.StateKey]types.ByteSequence)

	// key 1: Alpha
	alphaWrapper := StateWrapper{StateIndex: 1}
	key1 := alphaWrapper.StateKeyConstruct()
	serialized[key1] = serializeAlpha(state.Alpha)

	// key 2: Varphi
	varphiWrapper := StateWrapper{StateIndex: 2}
	key2 := varphiWrapper.StateKeyConstruct()
	serialized[key2] = serializeVarphi(state.Varphi)

	// key 3: Beta
	betaWrapper := StateWrapper{StateIndex: 3}
	key3 := betaWrapper.StateKeyConstruct()
	serialized[key3] = serializeBeta(state.Beta)

	// key 4: gamma
	gammaWrapper := StateWrapper{StateIndex: 4}
	key4 := gammaWrapper.StateKeyConstruct()
	serialized[key4] = serializeGamma(state.Gamma)

	// key 5: phi
	psiWrapper := StateWrapper{StateIndex: 5}
	key5 := psiWrapper.StateKeyConstruct()
	serialized[key5] = serializePsi(state.Psi)

	// key 6: eta
	etaWrapper := StateWrapper{StateIndex: 6}
	key6 := etaWrapper.StateKeyConstruct()
	serialized[key6] = serializeEta(state.Eta)

	// key 7: iota
	iotaWrapper := StateWrapper{StateIndex: 7}
	key7 := iotaWrapper.StateKeyConstruct()
	serialized[key7] = serializeIota(state.Iota)

	// key 8: kappa
	kappaWrapper := StateWrapper{StateIndex: 8}
	key8 := kappaWrapper.StateKeyConstruct()
	serialized[key8] = serializeKappa(state.Kappa)

	// key 9: lambda
	lambdaWrapper := StateWrapper{StateIndex: 9}
	key9 := lambdaWrapper.StateKeyConstruct()
	serialized[key9] = serializeLambda(state.Lambda)

	// key 10: rho
	rhoWrapper := StateWrapper{StateIndex: 10}
	key10 := rhoWrapper.StateKeyConstruct()
	serialized[key10] = serializeRho(state.Rho)

	// key 11: tau
	tauWrapper := StateWrapper{StateIndex: 11}
	key11 := tauWrapper.StateKeyConstruct()
	serialized[key11] = serializeTau(state.Tau)

	// key 12: chi
	chiWrapper := StateWrapper{StateIndex: 12}
	key12 := chiWrapper.StateKeyConstruct()
	serialized[key12] = serializeChi(state.Chi)

	// key 13: pi
	piWrapper := StateWrapper{StateIndex: 13}
	key13 := piWrapper.StateKeyConstruct()
	serialized[key13] = serializePi(state.Pi)

	// key 14: theta
	thetaWrapper := StateWrapper{StateIndex: 14}
	key14 := thetaWrapper.StateKeyConstruct()
	serialized[key14] = serializeTheta(state.Theta)

	// key 15: xi
	xiWrapper := StateWrapper{StateIndex: 15}
	key15 := xiWrapper.StateKeyConstruct()
	serialized[key15] = serializeXi(state.Xi)

	// delta 1
	for k, v := range state.Delta {
		key16, delta1Output := serializeDelta1KeyVal(types.ServiceId(k), v)
		serialized[key16] = delta1Output
	}

	// delta 2
	for id, account := range state.Delta {
		for key, value := range account.StorageDict {
			key17, delta2Output := serializeDelta2KeyVal(types.ServiceId(id), key, value)
			serialized[key17] = delta2Output
		}
	}

	// delta 3
	for id, account := range state.Delta {
		for key, value := range account.PreimageLookup {
			key18, delta3Output := serializeDelta3KeyVal(types.ServiceId(id), key, value)
			serialized[key18] = delta3Output
		}
	}

	// delta 4
	for id, account := range state.Delta {
		for key, value := range account.LookupDict {
			key19, delta4Output := serializeDelta4KeyVal(types.ServiceId(id), key, value)
			serialized[key19] = delta4Output
		}
	}

	return serialized, nil
}

// (D.2) T(σ)
func StateEncoder(state types.State) (types.StateKeyVals, error) {
	serialized := types.StateKeyVals{}

	// key 1: alpha
	keyval1 := types.StateKeyVal{
		Key:   encodeAlphaKey(),
		Value: encodeAlpha(state.Alpha),
	}
	serialized = append(serialized, keyval1)

	// key 2: varphi
	keyval2 := types.StateKeyVal{
		Key:   encodeVarphiKey(),
		Value: encodeVarphi(state.Varphi),
	}
	serialized = append(serialized, keyval2)

	// key 3: beta
	keyval3 := types.StateKeyVal{
		Key:   encodeBetaKey(),
		Value: encodeBeta(state.Beta),
	}
	serialized = append(serialized, keyval3)

	// key 4: gamma
	keyval4 := types.StateKeyVal{
		Key:   encodeGammaKey(),
		Value: encodeGamma(state.Gamma),
	}
	serialized = append(serialized, keyval4)

	// key 5: psi
	keyval5 := types.StateKeyVal{
		Key:   encodePsiKey(),
		Value: encodePsi(state.Psi),
	}
	serialized = append(serialized, keyval5)

	// key 6: eta
	keyval6 := types.StateKeyVal{
		Key:   encodeEtaKey(),
		Value: encodeEta(state.Eta),
	}
	serialized = append(serialized, keyval6)

	// key 7: iota
	keyval7 := types.StateKeyVal{
		Key:   encodeIotaKey(),
		Value: encodeIota(state.Iota),
	}
	serialized = append(serialized, keyval7)

	// key 8: kappa
	keyval8 := types.StateKeyVal{
		Key:   encodeKappaKey(),
		Value: encodeKappa(state.Kappa),
	}
	serialized = append(serialized, keyval8)

	// key 9: lambda
	keyval9 := types.StateKeyVal{
		Key:   encodeLambdaKey(),
		Value: encodeLambda(state.Lambda),
	}
	serialized = append(serialized, keyval9)

	// key 10: rho
	keyval10 := types.StateKeyVal{
		Key:   encodeRhoKey(),
		Value: encodeRho(state.Rho),
	}
	serialized = append(serialized, keyval10)

	// key 11: tau
	keyval11 := types.StateKeyVal{
		Key:   encodeTauKey(),
		Value: encodeTau(state.Tau),
	}
	serialized = append(serialized, keyval11)

	// key 12: chi
	keyval12 := types.StateKeyVal{
		Key:   encodeChiKey(),
		Value: encodeChi(state.Chi),
	}
	serialized = append(serialized, keyval12)

	// key 13: pi
	keyval13 := types.StateKeyVal{
		Key:   encodePiKey(),
		Value: encodePi(state.Pi),
	}
	serialized = append(serialized, keyval13)

	// key 14: theta
	keyval14 := types.StateKeyVal{
		Key:   encodeThetaKey(),
		Value: encodeTheta(state.Theta),
	}
	serialized = append(serialized, keyval14)

	// key 15: xi
	keyval15 := types.StateKeyVal{
		Key:   encodeXiKey(),
		Value: encodeXi(state.Xi),
	}
	serialized = append(serialized, keyval15)

	// delta 1
	for id, account := range state.Delta {
		stateKeyVal := encodeDelta1KeyVal(id, account)
		serialized = append(serialized, stateKeyVal)
	}

	// delta 2
	for id, account := range state.Delta {
		for key, val := range account.StorageDict {
			stateKeyVal := encodeDelta2KeyVal(id, key, val)
			serialized = append(serialized, stateKeyVal)
		}
	}

	// delta 3
	for id, account := range state.Delta {
		for key, val := range account.PreimageLookup {
			stateKeyVal := encodeDelta3KeyVal(id, key, val)
			serialized = append(serialized, stateKeyVal)
		}
	}

	// delta 4
	for id, account := range state.Delta {
		for key, val := range account.LookupDict {
			stateKeyVal := encodeDelta4KeyVal(id, key, val)
			serialized = append(serialized, stateKeyVal)
		}
	}

	return serialized, nil
}

func serializeFromMmrPeaks(m types.Mmr) (output types.ByteSequence) {
	output = append(output, utilities.SerializeU64(types.U64(len(m.Peaks)))...)
	for _, peak := range m.Peaks {
		if peak == nil || len(*peak) == 0 {
			output = append(output, utilities.SerializeU64(types.U64(0))...)
		} else {
			output = append(output, utilities.SerializeU64(types.U64(1))...)
			output = append(output, utilities.SerializeByteSequence(peak[:])...)
		}
	}
	return output
}

func calculateOctectsNumbers(serviceAccount types.ServiceAccount) (output types.U32) {
	output = 2*types.U32(len(serviceAccount.LookupDict)) + types.U32(len(serviceAccount.StorageDict))
	return output
}

func calculateItemNumbers(serviceAccount types.ServiceAccount) (output types.U64) {
	var sum1 types.U64
	for key := range serviceAccount.LookupDict {
		sum1 += 81 + types.U64(key.Length)
	}
	var sum2 types.U64
	for _, value := range serviceAccount.StorageDict {
		sum2 += 32 + types.U64(len(value))
	}
	output = sum1 + sum2
	return output
}
