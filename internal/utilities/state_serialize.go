package utilities

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func SerializeAlpha(alpha jamTypes.AuthPools) (output jamTypes.ByteSequence) {
	for _, authPool := range alpha {
		output = append(output, SerializeU64(jamTypes.U64(len(authPool)))...)
		for _, authorizerHash := range authPool {
			output = append(output, SerializeByteArray(jamTypes.ByteSequence(authorizerHash[:]))...)
		}
	}
	return output
}

func SerializeVarphi(varphi jamTypes.AuthQueues) (output jamTypes.ByteSequence) {
	for _, v := range varphi {
		for _, authQueue := range v {
			output = append(output, SerializeByteArray(jamTypes.ByteSequence(authQueue[:]))...)
		}
	}
	return output
}

func SerializeBeta(beta jamTypes.BlocksHistory) (output jamTypes.ByteSequence) {
	output = append(output, SerializeU64(jamTypes.U64(len(beta)))...)
	for _, blockInfo := range beta {
		// h
		output = append(output, SerializeU64(jamTypes.U64(len(blockInfo.HeaderHash)))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(blockInfo.HeaderHash[:]))...)

		// b
		mmrResult := SerializeFromMmrPeaks(blockInfo.Mmr)
		output = append(output, SerializeU64(jamTypes.U64(len(mmrResult)))...)
		output = append(output, SerializeByteArray(mmrResult)...)

		// s
		output = append(output, SerializeU64(jamTypes.U64(len(blockInfo.StateRoot)))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(blockInfo.StateRoot[:]))...)

		// p  (?) BlockInfo.Reported is different from GP, list or dict
		output = append(output, SerializeU64(jamTypes.U64(len(blockInfo.Reported)))...)
		for _, report := range blockInfo.Reported {
			output = append(output, SerializeByteArray(jamTypes.ByteSequence(report.Hash[:]))...)
			output = append(output, SerializeByteArray(jamTypes.ByteSequence(report.ExportsRoot[:]))...)
		}
	}
	return output
}

func SerializeGamma(gamma jamTypes.Gamma) (output jamTypes.ByteSequence) {
	// gamma_k
	for _, v := range gamma.GammaK {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Bandersnatch[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Ed25519[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Bls[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Metadata[:]))...)
	}

	// gamma_z
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(gamma.GammaZ[:]))...)

	// gamma_s (6.5)
	if gamma.GammaS.Tickets != nil {
		output = append(output, SerializeU64(jamTypes.U64(0))...)
		for _, ticket := range gamma.GammaS.Tickets {
			output = append(output, SerializeByteArray(jamTypes.ByteSequence(ticket.Id[:]))...)
			output = append(output, SerializeU64(jamTypes.U64(ticket.Attempt))...)
		}
	} else if gamma.GammaS.Keys != nil {
		output = append(output, SerializeU64(jamTypes.U64(1))...)
		for _, key := range gamma.GammaS.Keys {
			output = append(output, SerializeByteArray(jamTypes.ByteSequence(key[:]))...)
		}
	}

	// gamma_a
	output = append(output, SerializeU64(jamTypes.U64(len(gamma.GammaA)))...)
	for _, ticket := range gamma.GammaA {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(ticket.Id[:]))...)
		output = append(output, SerializeU64(jamTypes.U64(ticket.Attempt))...)
	}
	return output
}

func SerializePsi(psi jamTypes.DisputesRecords) (output jamTypes.ByteSequence) {
	// psi_g
	output = append(output, SerializeU64(jamTypes.U64(len(psi.Good)))...)
	for _, workReportHash := range psi.Good {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReportHash[:]))...)
	}

	// psi_b
	output = append(output, SerializeU64(jamTypes.U64(len(psi.Bad)))...)
	for _, workReportHash := range psi.Bad {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReportHash[:]))...)
	}

	// psi_w
	output = append(output, SerializeU64(jamTypes.U64(len(psi.Wonky)))...)
	for _, workReportHash := range psi.Wonky {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReportHash[:]))...)
	}

	// psi_o
	output = append(output, SerializeU64(jamTypes.U64(len(psi.Offenders)))...)
	for _, ed25519public := range psi.Offenders {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(ed25519public[:]))...)
	}
	return output
}

func SerializeEta(eta jamTypes.EntropyBuffer) (output jamTypes.ByteSequence) {
	for _, entropy := range eta {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(entropy[:]))...)
	}
	return output
}

func SerializeIota(iota jamTypes.ValidatorsData) (output jamTypes.ByteSequence) {
	for _, v := range iota {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Bandersnatch[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Ed25519[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Bls[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Metadata[:]))...)
	}
	return output
}

func SerializeKappa(kappa jamTypes.ValidatorsData) (output jamTypes.ByteSequence) {
	for _, v := range kappa {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Bandersnatch[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Ed25519[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Bls[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Metadata[:]))...)
	}
	return output
}

func SerializeLambda(lambda jamTypes.ValidatorsData) (output jamTypes.ByteSequence) {
	for _, v := range lambda {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Bandersnatch[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Ed25519[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Bls[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.Metadata[:]))...)
	}
	return output
}

func SerializeRho(rho jamTypes.AvailabilityAssignments) (output jamTypes.ByteSequence) {
	for _, v := range rho {
		if v != nil {
			output = append(output, SerializeU64(jamTypes.U64(1))...)
			output = append(output, SerializeWorkReport(v.Report)...)
			output = append(output, SerializeFixedLength(v.Timeout, 4)...)
		} else {
			output = append(output, SerializeU64(jamTypes.U64(0))...)
		}
	}
	return output
}

func SerializeTau(tau jamTypes.TimeSlot) (output jamTypes.ByteSequence) {
	output = SerializeFixedLength(jamTypes.U32(tau), 4)
	return output
}

func SerializeChi(chi jamTypes.PrivilegedServices) (output jamTypes.ByteSequence) {
	output = append(output, SerializeFixedLength(chi.ManagerServiceIndex, 4)...)
	output = append(output, SerializeFixedLength(chi.AlterPhiServiceIndex, 4)...)
	output = append(output, SerializeFixedLength(chi.AlterIotaServiceIndex, 4)...)

	for k, v := range chi.AutoAccumulateGasLimits {
		output = append(output, SerializeU64(jamTypes.U64(k))...)
		output = append(output, SerializeU64(v)...)
	}
	return output
}

func SerializePi(pi jamTypes.Statistics) (output jamTypes.ByteSequence) {
	for _, activityRecord := range pi.Current {
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.Blocks, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.Tickets, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.PreImages, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.PreImagesSize, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.Guarantees, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.Assurances, 4))...)
	}
	for _, activityRecord := range pi.Last {
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.Blocks, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.Tickets, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.PreImages, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.PreImagesSize, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.Guarantees, 4))...)
		output = append(output, SerializeByteArray(SerializeFixedLength(activityRecord.Assurances, 4))...)
	}
	return output
}

func SerializeTheta(theta jamTypes.UnaccumulateWorkReports) (output jamTypes.ByteSequence) {
	output = SerializeU64(jamTypes.U64(len(theta)))
	for _, v := range theta {
		output = append(output, SerializeWorkReport(v.WorkReport)...)
		for _, dep := range v.UnaccumulatedDeps {
			output = append(output, SerializeByteArray(jamTypes.ByteSequence(dep[:]))...)
		}
	}
	return output
}

func SerializeXi(xi jamTypes.AccumulatedHistories) (output jamTypes.ByteSequence) {
	output = SerializeU64(jamTypes.U64(len(xi)))
	for _, v := range xi {
		for _, history := range v {
			output = append(output, SerializeByteArray(jamTypes.ByteSequence(history[:]))...)
		}
	}
	return output
}

func StateSerialize(state jamTypes.State) (map[jamTypes.OpaqueHash]jamTypes.ByteSequence, error) {
	serialized := make(map[jamTypes.OpaqueHash]jamTypes.ByteSequence)

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
		serviceWrapper := StateServiceWrapper{StateIndex: 255, ServiceIndex: jamTypes.U32(k)}
		key := serviceWrapper.StateKeyConstruct()

		// a_c
		delta1Output := jamTypes.ByteSequence(v.CodeHash[:])

		// a_b, a_g, a_m
		delta1Output = append(delta1Output, SerializeFixedLength(jamTypes.U64(v.Balance), 8)...)
		delta1Output = append(delta1Output, SerializeFixedLength(jamTypes.U64(v.MinItemGas), 8)...)
		delta1Output = append(delta1Output, SerializeFixedLength(jamTypes.U64(v.MinMemoGas), 8)...)

		// a_l (U64)
		delta1Output = append(delta1Output, SerializeFixedLength(CalculateItemNumbers(v), 8)...)

		// a_i
		delta1Output = append(delta1Output, SerializeFixedLength(CalculateOctectsNumbers(v), 4)...)
		serialized[key] = delta1Output
	}

	// delta 2
	for k, v := range state.Delta {
		for _, value := range v.StorageDict {
			hash := append(SerializeFixedLength(jamTypes.U32(1<<32-1), 4), value[0:29]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: jamTypes.U32(k), Hash: jamTypes.OpaqueHash(hash)}
			key := serviceWrapper.StateKeyConstruct()
			serialized[key] = value
		}
	}

	// delta 3
	for k, v := range state.Delta {
		for key, value := range v.PreimageLookup {
			hash := append(SerializeFixedLength(jamTypes.U32(1<<32-2), 4), key[1:30]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: jamTypes.U32(k), Hash: jamTypes.OpaqueHash(hash)}
			key := serviceWrapper.StateKeyConstruct()
			serialized[key] = value
		}
	}

	// delta 4
	for k, v := range state.Delta {
		for key, value := range v.LookupDict {
			h := hash.Blake2bHash(jamTypes.ByteSequence(key.Hash[:]))
			hashKey := append(SerializeFixedLength(jamTypes.U32(key.Length), 4), h[2:31]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: jamTypes.U32(k), Hash: jamTypes.OpaqueHash(hashKey)}
			key := serviceWrapper.StateKeyConstruct()

			var delta4Output jamTypes.ByteSequence
			for _, timeSlot := range value {
				delta4Output = append(delta4Output, SerializeFixedLength(jamTypes.U64(timeSlot), 4)...)
			}
			delta4Output = append(SerializeU64(jamTypes.U64(len(delta4Output))), SerializeByteArray(delta4Output)...)
			serialized[key] = delta4Output
		}
	}

	return serialized, nil
}

func SerializeFromMmrPeaks(m jamTypes.Mmr) (output jamTypes.ByteSequence) {
	serialItems := []Serializable{}
	for _, peak := range m.Peaks {
		// empty
		if peak == nil || len(*peak) == 0 {
			serialItems = append(serialItems, U64Wrapper{})
		} else {
			serialItems = append(serialItems, SerializableSequence{U64Wrapper{Value: 1}, ByteArray32Wrapper{Value: jamTypes.ByteArray32(*peak)}})
		}
	}
	disc := Discriminator{
		Value: serialItems,
	}
	disc.Serialize()
	return output
}

func SerializeWorkReport(workReport jamTypes.WorkReport) (output jamTypes.ByteSequence) {
	// packagespec
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReport.PackageSpec.Hash[:]))...)
	output = append(output, SerializeU64(jamTypes.U64(workReport.PackageSpec.Length))...)
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReport.PackageSpec.ErasureRoot[:]))...)
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReport.PackageSpec.ExportsRoot[:]))...)
	output = append(output, SerializeU64(jamTypes.U64(workReport.PackageSpec.ExportsCount))...)

	// refine context
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReport.Context.Anchor[:]))...)
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReport.Context.StateRoot[:]))...)
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReport.Context.BeefyRoot[:]))...)
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReport.Context.LookupAnchor[:]))...)
	output = append(output, SerializeU64(jamTypes.U64(workReport.Context.LookupAnchorSlot))...)
	for _, v := range workReport.Context.Prerequisites {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v[:]))...)
	}

	// core index
	output = append(output, SerializeU64(jamTypes.U64(workReport.CoreIndex))...)

	// authorization hash
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReport.AuthorizerHash[:]))...)

	// authoutput
	output = append(output, SerializeByteArray(jamTypes.ByteSequence(workReport.AuthOutput[:]))...)

	// segment root lookup
	for _, v := range workReport.SegmentRootLookup {
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.WorkPackageHash[:]))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.SegmentTreeRoot[:]))...)
	}

	// results
	for _, v := range workReport.Results {
		output = append(output, SerializeU64(jamTypes.U64(v.ServiceId))...)
		output = append(output, SerializeByteArray(jamTypes.ByteSequence(v.CodeHash[:]))...)
	}

	return output
}

func CalculateOctectsNumbers(serviceAccount jamTypes.ServiceAccount) (output jamTypes.U32) {
	output = 2*jamTypes.U32(len(serviceAccount.LookupDict)) + jamTypes.U32(len(serviceAccount.StorageDict))
	return output
}

func CalculateItemNumbers(serviceAccount jamTypes.ServiceAccount) (output jamTypes.U64) {
	var sum1 jamTypes.U64
	for key := range serviceAccount.LookupDict {
		sum1 += 81 + jamTypes.U64(key.Length)
	}
	var sum2 jamTypes.U64
	for _, value := range serviceAccount.StorageDict {
		sum2 += 32 + jamTypes.U64(len(value))
	}
	output = sum1 + sum2
	return output
}
