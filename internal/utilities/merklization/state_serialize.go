package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func SerializeAlpha(alpha types.AuthPools) (output types.ByteSequence) {
	for _, authPool := range alpha {
		output = append(output, utilities.SerializeU64(types.U64(len(authPool)))...)
		for _, authorizerHash := range authPool {
			output = append(output, utilities.SerializeByteSequence(authorizerHash[:])...)
		}
	}
	return output
}

func EncodeAlpha(alpha types.AuthPools) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedAlpha, err := encoder.Encode(&alpha)
	if err != nil {
		return nil
	}
	return encodedAlpha
}

func SerializeVarphi(varphi types.AuthQueues) (output types.ByteSequence) {
	for _, v := range varphi {
		for _, authQueue := range v {
			output = append(output, utilities.SerializeByteSequence(authQueue[:])...)
		}
	}
	return output
}

func EncodeVarphi(varphi types.AuthQueues) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedVarphi, err := encoder.Encode(&varphi)
	if err != nil {
		return nil
	}
	return encodedVarphi
}

func SerializeBeta(beta types.BlocksHistory) (output types.ByteSequence) {
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

func EncodeBeta(beta types.BlocksHistory) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedBeta, err := encoder.Encode(&beta)
	if err != nil {
		return nil
	}
	return encodedBeta
}

func SerializeGamma(gamma types.Gamma) (output types.ByteSequence) {
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

func EncodeGamma(gamma types.Gamma) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedGamma, err := encoder.Encode(&gamma)
	if err != nil {
		return nil
	}
	return encodedGamma
}

func SerializePsi(psi types.DisputesRecords) (output types.ByteSequence) {
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

func EncodePsi(psi types.DisputesRecords) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedPsi, err := encoder.Encode(&psi)
	if err != nil {
		return nil
	}
	return encodedPsi
}

func SerializeEta(eta types.EntropyBuffer) (output types.ByteSequence) {
	for _, entropy := range eta {
		output = append(output, utilities.SerializeByteSequence(entropy[:])...)
	}
	return output
}

func EncodeEta(eta types.EntropyBuffer) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedEta, err := encoder.Encode(&eta)
	if err != nil {
		return nil
	}
	return encodedEta
}

func SerializeIota(iota types.ValidatorsData) (output types.ByteSequence) {
	for _, v := range iota {
		output = append(output, utilities.SerializeByteSequence(v.Bandersnatch[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Ed25519[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Bls[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Metadata[:])...)
	}
	return output
}

func EncodeIota(iota types.ValidatorsData) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedIota, err := encoder.Encode(&iota)
	if err != nil {
		return nil
	}
	return encodedIota
}

func SerializeKappa(kappa types.ValidatorsData) (output types.ByteSequence) {
	for _, v := range kappa {
		output = append(output, utilities.SerializeByteSequence(v.Bandersnatch[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Ed25519[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Bls[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Metadata[:])...)
	}
	return output
}

func EncodeKappa(kappa types.ValidatorsData) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedKappa, err := encoder.Encode(&kappa)
	if err != nil {
		return nil
	}
	return encodedKappa
}

func SerializeLambda(lambda types.ValidatorsData) (output types.ByteSequence) {
	for _, v := range lambda {
		output = append(output, utilities.SerializeByteSequence(v.Bandersnatch[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Ed25519[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Bls[:])...)
		output = append(output, utilities.SerializeByteSequence(v.Metadata[:])...)
	}
	return output
}

func EncodeLambda(lambda types.ValidatorsData) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedLambda, err := encoder.Encode(&lambda)
	if err != nil {
		return nil
	}
	return encodedLambda
}

func SerializeRho(rho types.AvailabilityAssignments) (output types.ByteSequence) {
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

func EncodeRho(rho types.AvailabilityAssignments) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedRho, err := encoder.Encode(&rho)
	if err != nil {
		return nil
	}
	return encodedRho
}

func SerializeTau(tau types.TimeSlot) (output types.ByteSequence) {
	output = utilities.SerializeFixedLength(types.U32(tau), 4)
	return output
}

func EncodeTau(tau types.TimeSlot) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedTau, err := encoder.Encode(&tau)
	if err != nil {
		return nil
	}
	return encodedTau
}

func SerializeChi(chi types.Privileges) (output types.ByteSequence) {
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

func EncodeChi(chi types.Privileges) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedChi, err := encoder.Encode(&chi)
	if err != nil {
		return nil
	}
	return encodedChi
}

func SerializePi(pi types.Statistics) (output types.ByteSequence) {
	for _, activityRecord := range pi.ValsCurrent {
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
		// output = append(output, byte(0xdd), byte(n), byte(0xdd))
		output = append(output, utilities.SerializeU64(types.U64(core.DALoad))...)
		output = append(output, utilities.SerializeU64(types.U64(core.Popularity))...)
		output = append(output, utilities.SerializeU64(types.U64(core.Imports))...)
		output = append(output, utilities.SerializeU64(types.U64(core.Exports))...)
		output = append(output, utilities.SerializeU64(types.U64(core.ExtrinsicSize))...)
		output = append(output, utilities.SerializeU64(types.U64(core.ExtrinsicCount))...)
		// output = append(output, byte(0xed))
		output = append(output, utilities.SerializeU64(types.U64(core.BundleSize))...)
		// output = append(output, byte(0xee))
		output = append(output, utilities.SerializeU64(types.U64(core.GasUsed))...)
		// output = append(output, byte(0xef))
	}
	// output = append(output, byte(0xff))
	// output = append(output, utilities.SerializeU64(types.U64(len(pi.Services)))...)
	serPiS := utilities.WrapStatisticsServiceMap(pi.Services)
	output = append(output, serPiS.Serialize()...)
	return output
}

func EncodePi(pi types.Statistics) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedPi, err := encoder.Encode(&pi)
	if err != nil {
		return nil
	}
	return encodedPi
}

func SerializeTheta(theta types.ReadyQueue) (output types.ByteSequence) {
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

func EncodeTheta(theta types.ReadyQueue) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedTheta, err := encoder.Encode(&theta)
	if err != nil {
		return nil
	}
	return encodedTheta
}

func SerializeXi(xi types.AccumulatedQueue) (output types.ByteSequence) {
	for _, v := range xi {
		output = append(output, utilities.SerializeU64(types.U64(len(v)))...)
		for _, history := range v {
			output = append(output, utilities.SerializeByteSequence(history[:])...)
		}
	}
	return output
}

func EncodeXi(xi types.AccumulatedQueue) types.ByteSequence {
	encoder := types.NewEncoder()
	encodedXi, err := encoder.Encode(&xi)
	if err != nil {
		return nil
	}
	return encodedXi
}

func SerializeDelta1(s types.ServiceAccount) (output types.ByteSequence) {
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

func EncodeDelta1(serviceAccount types.ServiceAccount) (output types.ByteSequence) {
	encoder := types.NewEncoder()
	encodedCodeHash, err := encoder.Encode(&serviceAccount.ServiceInfo.CodeHash)
	if err != nil {
		return nil
	}
	output = append(output, encodedCodeHash...)

	encodedBalance, err := encoder.Encode(&serviceAccount.ServiceInfo.Balance)
	if err != nil {
		return nil
	}
	output = append(output, encodedBalance...)

	encodedMinItemGas, err := encoder.Encode(&serviceAccount.ServiceInfo.MinItemGas)
	if err != nil {
		return nil
	}
	output = append(output, encodedMinItemGas...)

	encodedMinMemoGas, err := encoder.Encode(&serviceAccount.ServiceInfo.MinMemoGas)
	if err != nil {
		return nil
	}
	output = append(output, encodedMinMemoGas...)

	a_o := service_account.CalcOctets(serviceAccount)
	encodedOctets, err := encoder.Encode(&a_o)
	if err != nil {
		return nil
	}
	output = append(output, encodedOctets...)

	a_i := service_account.CalcKeys(serviceAccount)
	encodedKeys, err := encoder.Encode(&a_i)
	if err != nil {
		return nil
	}
	output = append(output, encodedKeys...)
	return output
}

func StateSerialize(state types.State) (map[types.OpaqueHash]types.ByteSequence, error) {
	serialized := make(map[types.OpaqueHash]types.ByteSequence)

	// key 1: Alpha
	alphaWrapper := StateWrapper{StateIndex: 1}
	key1 := alphaWrapper.StateKeyConstruct()
	serialized[key1] = SerializeAlpha(state.Alpha)

	// key 2: Varphi
	varphiWrapper := StateWrapper{StateIndex: 2}
	key2 := varphiWrapper.StateKeyConstruct()
	serialized[key2] = SerializeVarphi(state.Varphi)

	// key 3: Beta
	betaWrapper := StateWrapper{StateIndex: 3}
	key3 := betaWrapper.StateKeyConstruct()
	serialized[key3] = SerializeBeta(state.Beta)

	// key 4: gamma
	gammaWrapper := StateWrapper{StateIndex: 4}
	key4 := gammaWrapper.StateKeyConstruct()
	serialized[key4] = SerializeGamma(state.Gamma)

	// key 5: phi
	psiWrapper := StateWrapper{StateIndex: 5}
	key5 := psiWrapper.StateKeyConstruct()
	serialized[key5] = SerializePsi(state.Psi)

	// key 6: eta
	etaWrapper := StateWrapper{StateIndex: 6}
	key6 := etaWrapper.StateKeyConstruct()
	serialized[key6] = SerializeEta(state.Eta)

	// key 7: iota
	iotaWrapper := StateWrapper{StateIndex: 7}
	key7 := iotaWrapper.StateKeyConstruct()
	serialized[key7] = SerializeIota(state.Iota)

	// key 8: kappa
	kappaWrapper := StateWrapper{StateIndex: 8}
	key8 := kappaWrapper.StateKeyConstruct()
	serialized[key8] = SerializeKappa(state.Kappa)

	// key 9: lambda
	lambdaWrapper := StateWrapper{StateIndex: 9}
	key9 := lambdaWrapper.StateKeyConstruct()
	serialized[key9] = SerializeLambda(state.Lambda)

	// key 10: rho
	rhoWrapper := StateWrapper{StateIndex: 10}
	key10 := rhoWrapper.StateKeyConstruct()
	serialized[key10] = SerializeRho(state.Rho)

	// key 11: tau
	tauWrapper := StateWrapper{StateIndex: 11}
	key11 := tauWrapper.StateKeyConstruct()
	serialized[key11] = SerializeTau(state.Tau)

	// key 12: chi
	chiWrapper := StateWrapper{StateIndex: 12}
	key12 := chiWrapper.StateKeyConstruct()
	serialized[key12] = SerializeChi(state.Chi)

	// key 13: pi
	piWrapper := StateWrapper{StateIndex: 13}
	key13 := piWrapper.StateKeyConstruct()
	serialized[key13] = SerializePi(state.Pi)

	// key 14: theta
	thetaWrapper := StateWrapper{StateIndex: 14}
	key14 := thetaWrapper.StateKeyConstruct()
	serialized[key14] = SerializeTheta(state.Theta)

	// key 15: xi
	xiWrapper := StateWrapper{StateIndex: 15}
	key15 := xiWrapper.StateKeyConstruct()
	serialized[key15] = SerializeXi(state.Xi)

	// delta 1
	for k, v := range state.Delta {
		serviceWrapper := StateServiceWrapper{StateIndex: 255, ServiceIndex: types.U32(k)}
		key := serviceWrapper.StateKeyConstruct()

		delta1Output := SerializeDelta1(v)

		serialized[key] = delta1Output
	}

	// delta 2
	for id, account := range state.Delta {
		for key, value := range account.StorageDict {
			hash := append(utilities.SerializeFixedLength(types.U32(1<<32-1), 4), key[0:29]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: types.U32(id), Hash: types.OpaqueHash(hash)}
			key := serviceWrapper.StateKeyConstruct()
			serialized[key] = value
		}
	}

	// delta 3
	for id, account := range state.Delta {
		for key, value := range account.PreimageLookup {
			hash := append(utilities.SerializeFixedLength(types.U32(1<<32-2), 4), key[1:30]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: types.U32(id), Hash: types.OpaqueHash(hash)}
			key := serviceWrapper.StateKeyConstruct()
			serialized[key] = value
		}
	}

	// delta 4
	for id, account := range state.Delta {
		for key, value := range account.LookupDict {
			h := hash.Blake2bHash(types.ByteSequence(key.Hash[:]))
			hashKey := append(utilities.SerializeFixedLength(types.U32(key.Length), 4), h[2:31]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: types.U32(id), Hash: types.OpaqueHash(hashKey)}
			key := serviceWrapper.StateKeyConstruct()

			var delta4Output types.ByteSequence
			for _, timeSlot := range value {
				delta4Output = append(delta4Output, utilities.SerializeFixedLength(types.U64(timeSlot), 4)...)
			}
			delta4Output = append(utilities.SerializeU64(types.U64(len(delta4Output))), utilities.SerializeByteSequence(delta4Output)...)
			serialized[key] = delta4Output
		}
	}

	return serialized, nil
}

func StateEncoder(state types.State) (map[types.OpaqueHash]types.ByteSequence, error) {
	serialized := make(map[types.OpaqueHash]types.ByteSequence)

	// key 1: Alpha
	alphaWrapper := StateWrapper{StateIndex: 1}
	key1 := alphaWrapper.StateKeyConstruct()
	serialized[key1] = EncodeAlpha(state.Alpha)

	// key 2: Varphi
	varphiWrapper := StateWrapper{StateIndex: 2}
	key2 := varphiWrapper.StateKeyConstruct()
	serialized[key2] = EncodeVarphi(state.Varphi)

	// key 3: Beta
	betaWrapper := StateWrapper{StateIndex: 3}
	key3 := betaWrapper.StateKeyConstruct()
	serialized[key3] = EncodeBeta(state.Beta)

	// key 4: gamma
	gammaWrapper := StateWrapper{StateIndex: 4}
	key4 := gammaWrapper.StateKeyConstruct()
	serialized[key4] = EncodeGamma(state.Gamma)

	// key 5: phi
	psiWrapper := StateWrapper{StateIndex: 5}
	key5 := psiWrapper.StateKeyConstruct()
	serialized[key5] = EncodePsi(state.Psi)

	// key 6: eta
	etaWrapper := StateWrapper{StateIndex: 6}
	key6 := etaWrapper.StateKeyConstruct()
	serialized[key6] = EncodeEta(state.Eta)

	// key 7: iota
	iotaWrapper := StateWrapper{StateIndex: 7}
	key7 := iotaWrapper.StateKeyConstruct()
	serialized[key7] = EncodeIota(state.Iota)

	// key 8: kappa
	kappaWrapper := StateWrapper{StateIndex: 8}
	key8 := kappaWrapper.StateKeyConstruct()
	serialized[key8] = EncodeKappa(state.Kappa)

	// key 9: lambda
	lambdaWrapper := StateWrapper{StateIndex: 9}
	key9 := lambdaWrapper.StateKeyConstruct()
	serialized[key9] = EncodeLambda(state.Lambda)

	// key 10: rho
	rhoWrapper := StateWrapper{StateIndex: 10}
	key10 := rhoWrapper.StateKeyConstruct()
	serialized[key10] = EncodeRho(state.Rho)

	// key 11: tau
	tauWrapper := StateWrapper{StateIndex: 11}
	key11 := tauWrapper.StateKeyConstruct()
	serialized[key11] = EncodeTau(state.Tau)

	// key 12: chi
	chiWrapper := StateWrapper{StateIndex: 12}
	key12 := chiWrapper.StateKeyConstruct()
	serialized[key12] = EncodeChi(state.Chi)

	// key 13: pi
	piWrapper := StateWrapper{StateIndex: 13}
	key13 := piWrapper.StateKeyConstruct()
	serialized[key13] = EncodePi(state.Pi)

	// key 14: theta
	thetaWrapper := StateWrapper{StateIndex: 14}
	key14 := thetaWrapper.StateKeyConstruct()
	serialized[key14] = EncodeTheta(state.Theta)

	// key 15: xi
	xiWrapper := StateWrapper{StateIndex: 15}
	key15 := xiWrapper.StateKeyConstruct()
	serialized[key15] = EncodeXi(state.Xi)

	// delta 1
	for k, v := range state.Delta {
		serviceWrapper := StateServiceWrapper{StateIndex: 255, ServiceIndex: types.U32(k)}
		key := serviceWrapper.StateKeyConstruct()

		delta1Output := EncodeDelta1(v)
		serialized[key] = delta1Output
	}

	// delta 2
	for id, account := range state.Delta {
		for key, value := range account.StorageDict {
			hash := append(utilities.SerializeFixedLength(types.U32(1<<32-1), 4), key[0:29]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: types.U32(id), Hash: types.OpaqueHash(hash)}
			key := serviceWrapper.StateKeyConstruct()
			serialized[key] = value
		}
	}

	// delta 3
	for id, account := range state.Delta {
		for key, value := range account.PreimageLookup {
			hash := append(utilities.SerializeFixedLength(types.U32(1<<32-2), 4), key[1:30]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: types.U32(id), Hash: types.OpaqueHash(hash)}
			key := serviceWrapper.StateKeyConstruct()
			serialized[key] = value
		}
	}

	// delta 4
	for id, account := range state.Delta {
		for key, value := range account.LookupDict {
			h := hash.Blake2bHash(types.ByteSequence(key.Hash[:]))
			hashKey := append(utilities.SerializeFixedLength(types.U32(key.Length), 4), h[2:31]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: types.U32(id), Hash: types.OpaqueHash(hashKey)}
			key := serviceWrapper.StateKeyConstruct()

			var delta4Output types.ByteSequence
			for _, timeSlot := range value {
				delta4Output = append(delta4Output, utilities.SerializeFixedLength(types.U64(timeSlot), 4)...)
			}
			delta4Output = append(utilities.SerializeU64(types.U64(len(delta4Output))), utilities.SerializeByteSequence(delta4Output)...)
			serialized[key] = delta4Output
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
