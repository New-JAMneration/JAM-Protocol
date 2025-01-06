package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func SerializeAlpha(alpha types.AuthPools) (output types.ByteSequence) {
	for _, authPool := range alpha {
		output = append(output, utilities.SerializeU64(types.U64(len(authPool)))...)
		for _, authorizerHash := range authPool {
			output = append(output, utilities.SerializeByteSequence(types.ByteSequence(authorizerHash[:]))...)
		}
	}
	return output
}

func SerializeVarphi(varphi types.AuthQueues) (output types.ByteSequence) {
	for _, v := range varphi {
		for _, authQueue := range v {
			output = append(output, utilities.SerializeByteSequence(types.ByteSequence(authQueue[:]))...)
		}
	}
	return output
}

func SerializeBeta(beta types.BlocksHistory) (output types.ByteSequence) {
	output = append(output, utilities.SerializeU64(types.U64(len(beta)))...)
	for _, blockInfo := range beta {
		// h
		output = append(output, utilities.SerializeU64(types.U64(len(blockInfo.HeaderHash)))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(blockInfo.HeaderHash[:]))...)

		// b
		mmrResult := SerializeFromMmrPeaks(blockInfo.Mmr)
		output = append(output, utilities.SerializeU64(types.U64(len(mmrResult)))...)
		output = append(output, utilities.SerializeByteSequence(mmrResult)...)

		// s
		output = append(output, utilities.SerializeU64(types.U64(len(blockInfo.StateRoot)))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(blockInfo.StateRoot[:]))...)

		// p  (?) BlockInfo.Reported is different from GP, list or dict
		output = append(output, utilities.SerializeU64(types.U64(len(blockInfo.Reported)))...)
		for _, report := range blockInfo.Reported {
			output = append(output, utilities.SerializeByteSequence(types.ByteSequence(report.Hash[:]))...)
			output = append(output, utilities.SerializeByteSequence(types.ByteSequence(report.ExportsRoot[:]))...)
		}
	}
	return output
}

func SerializeGamma(gamma types.Gamma) (output types.ByteSequence) {
	// gamma_k
	for _, v := range gamma.GammaK {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Bandersnatch[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Ed25519[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Bls[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Metadata[:]))...)
	}

	// gamma_z
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(gamma.GammaZ[:]))...)

	// gamma_s (6.5)
	if gamma.GammaS.Tickets != nil {
		output = append(output, utilities.SerializeU64(types.U64(0))...)
		for _, ticket := range gamma.GammaS.Tickets {
			output = append(output, utilities.SerializeByteSequence(types.ByteSequence(ticket.Id[:]))...)
			output = append(output, utilities.SerializeU64(types.U64(ticket.Attempt))...)
		}
	} else if gamma.GammaS.Keys != nil {
		output = append(output, utilities.SerializeU64(types.U64(1))...)
		for _, key := range gamma.GammaS.Keys {
			output = append(output, utilities.SerializeByteSequence(types.ByteSequence(key[:]))...)
		}
	}

	// gamma_a
	output = append(output, utilities.SerializeU64(types.U64(len(gamma.GammaA)))...)
	for _, ticket := range gamma.GammaA {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(ticket.Id[:]))...)
		output = append(output, utilities.SerializeU64(types.U64(ticket.Attempt))...)
	}
	return output
}

func SerializePsi(psi types.DisputesRecords) (output types.ByteSequence) {
	// psi_g
	output = append(output, utilities.SerializeU64(types.U64(len(psi.Good)))...)
	for _, workReportHash := range psi.Good {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReportHash[:]))...)
	}

	// psi_b
	output = append(output, utilities.SerializeU64(types.U64(len(psi.Bad)))...)
	for _, workReportHash := range psi.Bad {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReportHash[:]))...)
	}

	// psi_w
	output = append(output, utilities.SerializeU64(types.U64(len(psi.Wonky)))...)
	for _, workReportHash := range psi.Wonky {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReportHash[:]))...)
	}

	// psi_o
	output = append(output, utilities.SerializeU64(types.U64(len(psi.Offenders)))...)
	for _, ed25519public := range psi.Offenders {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(ed25519public[:]))...)
	}
	return output
}

func SerializeEta(eta types.EntropyBuffer) (output types.ByteSequence) {
	for _, entropy := range eta {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(entropy[:]))...)
	}
	return output
}

func SerializeIota(iota types.ValidatorsData) (output types.ByteSequence) {
	for _, v := range iota {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Bandersnatch[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Ed25519[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Bls[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Metadata[:]))...)
	}
	return output
}

func SerializeKappa(kappa types.ValidatorsData) (output types.ByteSequence) {
	for _, v := range kappa {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Bandersnatch[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Ed25519[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Bls[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Metadata[:]))...)
	}
	return output
}

func SerializeLambda(lambda types.ValidatorsData) (output types.ByteSequence) {
	for _, v := range lambda {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Bandersnatch[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Ed25519[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Bls[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.Metadata[:]))...)
	}
	return output
}

func SerializeRho(rho types.AvailabilityAssignments) (output types.ByteSequence) {
	for _, v := range rho {
		if v != nil {
			output = append(output, utilities.SerializeU64(types.U64(1))...)
			output = append(output, SerializeWorkReport(v.Report)...)
			output = append(output, utilities.SerializeFixedLength(types.U32(v.Timeout), 4)...)
		} else {
			output = append(output, utilities.SerializeU64(types.U64(0))...)
		}
	}
	return output
}

func SerializeTau(tau types.TimeSlot) (output types.ByteSequence) {
	output = utilities.SerializeFixedLength(types.U32(tau), 4)
	return output
}

func SerializeChi(chi types.PrivilegedServices) (output types.ByteSequence) {
	output = append(output, utilities.SerializeFixedLength(chi.ManagerServiceIndex, 4)...)
	output = append(output, utilities.SerializeFixedLength(chi.AlterPhiServiceIndex, 4)...)
	output = append(output, utilities.SerializeFixedLength(chi.AlterIotaServiceIndex, 4)...)

	for k, v := range chi.AutoAccumulateGasLimits {
		output = append(output, utilities.SerializeU64(types.U64(k))...)
		output = append(output, utilities.SerializeU64(v)...)
	}
	return output
}

func SerializePi(pi types.Statistics) (output types.ByteSequence) {
	for _, activityRecord := range pi.Current {
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.Blocks, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.Tickets, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.PreImages, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.PreImagesSize, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.Guarantees, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.Assurances, 4))...)
	}
	for _, activityRecord := range pi.Last {
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.Blocks, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.Tickets, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.PreImages, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.PreImagesSize, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.Guarantees, 4))...)
		output = append(output, utilities.SerializeByteSequence(utilities.SerializeFixedLength(activityRecord.Assurances, 4))...)
	}
	return output
}

func SerializeTheta(theta types.UnaccumulateWorkReports) (output types.ByteSequence) {
	output = utilities.SerializeU64(types.U64(len(theta)))
	for _, v := range theta {
		output = append(output, SerializeWorkReport(v.WorkReport)...)
		for _, dep := range v.UnaccumulatedDeps {
			output = append(output, utilities.SerializeByteSequence(types.ByteSequence(dep[:]))...)
		}
	}
	return output
}

func SerializeXi(xi types.AccumulatedHistories) (output types.ByteSequence) {
	output = utilities.SerializeU64(types.U64(len(xi)))
	for _, v := range xi {
		for _, history := range v {
			output = append(output, utilities.SerializeByteSequence(types.ByteSequence(history[:]))...)
		}
	}
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

		// a_c
		delta1Output := types.ByteSequence(v.CodeHash[:])

		// a_b, a_g, a_m
		delta1Output = append(delta1Output, utilities.SerializeFixedLength(types.U64(v.Balance), 8)...)
		delta1Output = append(delta1Output, utilities.SerializeFixedLength(types.U64(v.MinItemGas), 8)...)
		delta1Output = append(delta1Output, utilities.SerializeFixedLength(types.U64(v.MinMemoGas), 8)...)

		// a_l (U64)
		delta1Output = append(delta1Output, utilities.SerializeFixedLength(CalculateItemNumbers(v), 8)...)

		// a_i
		delta1Output = append(delta1Output, utilities.SerializeFixedLength(CalculateOctectsNumbers(v), 4)...)
		serialized[key] = delta1Output
	}

	// delta 2
	for k, v := range state.Delta {
		for _, value := range v.StorageDict {
			hash := append(utilities.SerializeFixedLength(types.U32(1<<32-1), 4), value[0:29]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: types.U32(k), Hash: types.OpaqueHash(hash)}
			key := serviceWrapper.StateKeyConstruct()
			serialized[key] = value
		}
	}

	// delta 3
	for k, v := range state.Delta {
		for key, value := range v.PreimageLookup {
			hash := append(utilities.SerializeFixedLength(types.U32(1<<32-2), 4), key[1:30]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: types.U32(k), Hash: types.OpaqueHash(hash)}
			key := serviceWrapper.StateKeyConstruct()
			serialized[key] = value
		}
	}

	// delta 4
	for k, v := range state.Delta {
		for key, value := range v.LookupDict {
			h := hash.Blake2bHash(types.ByteSequence(key.Hash[:]))
			hashKey := append(utilities.SerializeFixedLength(types.U32(key.Length), 4), h[2:31]...)
			serviceWrapper := ServiceWrapper{ServiceIndex: types.U32(k), Hash: types.OpaqueHash(hashKey)}
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

func SerializeFromMmrPeaks(m types.Mmr) (output types.ByteSequence) {
	serialItems := []utilities.Serializable{}
	for _, peak := range m.Peaks {
		// empty
		if peak == nil || len(*peak) == 0 {
			serialItems = append(serialItems, utilities.U64Wrapper{})
		} else {
			serialItems = append(serialItems, utilities.SerializableSequence{utilities.U64Wrapper{Value: 1}, utilities.ByteArray32Wrapper{Value: types.ByteArray32(*peak)}})
		}
	}
	disc := utilities.Discriminator{
		Value: serialItems,
	}
	disc.Serialize()
	return output
}

func SerializeWorkReport(workReport types.WorkReport) (output types.ByteSequence) {
	// packagespec
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReport.PackageSpec.Hash[:]))...)
	output = append(output, utilities.SerializeU64(types.U64(workReport.PackageSpec.Length))...)
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReport.PackageSpec.ErasureRoot[:]))...)
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReport.PackageSpec.ExportsRoot[:]))...)
	output = append(output, utilities.SerializeU64(types.U64(workReport.PackageSpec.ExportsCount))...)

	// refine context
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReport.Context.Anchor[:]))...)
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReport.Context.StateRoot[:]))...)
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReport.Context.BeefyRoot[:]))...)
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReport.Context.LookupAnchor[:]))...)
	output = append(output, utilities.SerializeU64(types.U64(workReport.Context.LookupAnchorSlot))...)
	for _, v := range workReport.Context.Prerequisites {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v[:]))...)
	}

	// core index
	output = append(output, utilities.SerializeU64(types.U64(workReport.CoreIndex))...)

	// authorization hash
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReport.AuthorizerHash[:]))...)

	// authoutput
	output = append(output, utilities.SerializeByteSequence(types.ByteSequence(workReport.AuthOutput[:]))...)

	// segment root lookup
	for _, v := range workReport.SegmentRootLookup {
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.WorkPackageHash[:]))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.SegmentTreeRoot[:]))...)
	}

	// results
	for _, v := range workReport.Results {
		output = append(output, utilities.SerializeU64(types.U64(v.ServiceId))...)
		output = append(output, utilities.SerializeByteSequence(types.ByteSequence(v.CodeHash[:]))...)
	}

	return output
}

func CalculateOctectsNumbers(serviceAccount types.ServiceAccount) (output types.U32) {
	output = 2*types.U32(len(serviceAccount.LookupDict)) + types.U32(len(serviceAccount.StorageDict))
	return output
}

func CalculateItemNumbers(serviceAccount types.ServiceAccount) (output types.U64) {
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
