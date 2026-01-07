package accumulation

import (
	"bytes"
	"slices"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// (12.20)
func sumPrivilegesGas(alwaysAccumulateMap types.AlwaysAccumulateMap) types.Gas {
	sum := types.Gas(0)
	for _, value := range alwaysAccumulateMap {
		sum += value
	}

	return sum
}

// Calculate max gas used v0.6.4 (12.20)
func calculateMaxGasUsed(alwaysAccumulateMap types.AlwaysAccumulateMap) types.Gas {
	GT := types.Gas(types.TotalGas)
	GA := types.Gas(types.MaxAccumulateGas)
	C := types.Gas(types.CoresCount)

	sum := sumPrivilegesGas(alwaysAccumulateMap)

	return max(GT, GA*C+sum)
}

func updatePartialStateSetToPosteriorState(cs *blockchain.ChainState, o types.PartialStateSet) {
	// (12.22)
	deltaDagger := o.ServiceAccounts
	postIota := o.ValidatorKeys
	postVarphi := o.Authorizers

	// Update the posterior state
	cs.GetPosteriorStates().SetChi(types.Privileges{
		Bless:       o.Bless,
		Assign:      o.Assign,
		Designate:   o.Designate,
		AlwaysAccum: o.AlwaysAccum,
	})
	cs.GetIntermediateStates().SetDeltaDagger(deltaDagger)
	cs.GetPosteriorStates().SetIota(postIota)
	cs.GetPosteriorStates().SetVarphi(postVarphi)
}

// (12.25)
// N(s)
// W^*: accumulatableWorkReports
// w_r: work result
// r_s: the index of the service whose state is to be altered and thus whose refine code was already executed
// s: serviceId
// NOTE: While it is possible to refactor the function to use a map where the key is the service ID and the value is the number of work results,
// this approach would differ from the graypaper and is not being implemented at this time.
func getWorkResultByService(s types.ServiceId, n types.U64) []types.WorkResult {
	// Get W^*
	cs := blockchain.GetInstance()
	accumulatableWorkReports := cs.GetIntermediateStates().GetAccumulatableWorkReports()

	// First pass: count matching results to determine exact capacity
	count := 0
	for _, workReport := range accumulatableWorkReports[:n] {
		for _, result := range workReport.Results {
			if result.ServiceId == s {
				count++
			}
		}
	}

	// Second pass: append results to a pre-allocated slice
	output := make([]types.WorkResult, 0, count)
	for _, workReport := range accumulatableWorkReports[:n] {
		for _, result := range workReport.Results {
			if result.ServiceId == s {
				output = append(output, result)
			}
		}
	}

	return output
}

// (12.23) (12.24) (12.25)
// u from outer accuulation function
// INFO: Acutally, The I(accumulation statistics) used in chapter 13 (pi_S)
// We save the accumulation statistics in the cs
func calculateAccumulationStatistics(serviceGasUsedList types.ServiceGasUsedList, n types.U64) types.AccumulationStatistics {
	// (12.28–12.29)
	// S ≡ {(s ↦ (G(s), N(s))) | G(s)+N(s) ≠ 0}
	// where:
	//   G(s) ≡ Σ₍ₛ,ᵤ₎∈ᵤ(u)
	//   N(s) ≡ [[d | r ∈ R*...n, d ∈ r_d, dₛ = s]]
	G := map[types.ServiceId]types.Gas{} // G(s)
	for _, serviceGasUsed := range serviceGasUsedList {
		G[serviceGasUsed.ServiceId] += serviceGasUsed.Gas
	}

	// calcualte the number of work reports accumulated
	S := types.AccumulationStatistics{}
	for s, Gs := range G {
		Ns := types.U64(len(getWorkResultByService(s, n)))

		if types.U64(Gs)+Ns == 0 {
			continue // skip, N(S) = []
		}

		S[s] = types.GasAndNumAccumulatedReports{
			Gas:                   Gs,
			NumAccumulatedReports: Ns,
		}
	}
	return S
}

// (12.28) (12.29)
// Build delta double dagger (second intermediate state)
// NOTE: v0.7.1 has removed deferred transfers & Ψ_T
func updateDeltaDoubleDagger(cs *blockchain.ChainState, accumulationStatistics types.AccumulationStatistics) {
	deltaDagger := cs.GetIntermediateStates().GetDeltaDagger()
	tauPrime := cs.GetPosteriorStates().GetTau()

	deltaDoubleDagger := types.ServiceAccountState{}

	for serviceId, acc := range deltaDagger {
		// If this service was actually accumulated this round
		if _, ok := accumulationStatistics[serviceId]; ok {
			acc.ServiceInfo.LastAccumulationSlot = tauPrime
		}
		deltaDoubleDagger[serviceId] = acc
	}
	cs.GetIntermediateStates().SetDeltaDoubleDagger(deltaDoubleDagger)
}

// (12.31) (12.32)
// Update the AccumulatedQueue(AccumulatedQueue)
func updateXi(cs *blockchain.ChainState, n types.U64) {
	// Get W^* (accumulatable work-reports in this block)
	accumulatableWorkReports := cs.GetIntermediateStates().GetAccumulatableWorkReports()

	priorXi := cs.GetPriorStates().GetXi()
	posteriorXi := cs.GetPosteriorStates().GetXi()

	// (12.31) Update the last element
	posteriorXi[types.EpochLength-1] = ExtractWorkReportHashes(accumulatableWorkReports[:n])

	// (12.32)
	// Update the rest of the elements
	for i := 0; i < types.EpochLength-1; i++ {
		posteriorXi[i] = priorXi[i+1]
	}
	// WONDER: this sort is not mentioned in the graypaper
	for _, x := range posteriorXi {
		slices.SortStableFunc(x, func(a, b types.WorkPackageHash) int {
			// Put empty slices at the end
			if len(a) == 0 && len(b) == 0 {
				return 0
			}
			if len(a) == 0 {
				return 1
			}
			if len(b) == 0 {
				return -1
			}
			return bytes.Compare(a[:], b[:])
		})
	}

	// Update posteriorXi to cs
	cs.GetPosteriorStates().SetXi(posteriorXi)
}

// (12.33)
// Update ReadyQueue(Theta)
func updateTheta(cs *blockchain.ChainState) {
	// (12.10) let m = H_t mode E
	headerSlot := cs.GetLatestBlock().Header.Slot
	m := int(headerSlot) % types.EpochLength

	// (6.2) tau and tau prime
	// Get previous time slot index
	tau := cs.GetPriorStates().GetTau()

	// Get current time slot index
	tauPrime := cs.GetPosteriorStates().GetTau()

	tauOffset := tauPrime - tau

	// Get queued work reports
	queueWorkReports := cs.GetIntermediateStates().GetQueuedWorkReports()

	// Get prior theta and posterior theta (ReadyQueue)
	priorTheta := cs.GetPriorStates().GetTheta()
	posteriorTheta := cs.GetPosteriorStates().GetTheta()

	// Get posterior xi
	posteriorXi := cs.GetPosteriorStates().GetXi()

	for i := 0; i < types.EpochLength; i++ {
		// s[i]↺ ≡ s[ i % ∣s∣ ]
		index := (m - i + types.EpochLength) % types.EpochLength
		index = index % len(posteriorTheta)

		firstCondition := i == 0
		secondCondition := (1 <= i) && (i < int(tauOffset))
		thirdCondition := i >= int(tauOffset)

		if firstCondition {
			posteriorTheta[index] = QueueEditingFunction(queueWorkReports, posteriorXi[types.EpochLength-1])
		} else if secondCondition {
			posteriorTheta[index] = types.ReadyQueueItem{}
		} else if thirdCondition {
			posteriorTheta[index] = QueueEditingFunction(priorTheta[index], posteriorXi[types.EpochLength-1])
		}
	}

	// Update posterior theta
	cs.GetPosteriorStates().SetTheta(posteriorTheta)
}

// (12.20) (12.21)
func executeOuterAccumulation(cs *blockchain.ChainState) (OuterAccumulationOutput, error) {
	// Get W^* (accumulatable work-reports in this block)
	accumulatableWorkReports := cs.GetIntermediateStates().GetAccumulatableWorkReports()

	// (12.13) PartialStateSet
	priorState := cs.GetPriorStates()
	chi := priorState.GetChi()
	delta := priorState.GetDelta()
	iota := priorState.GetIota()
	varphi := priorState.GetVarphi()

	partialStateSet := types.PartialStateSet{
		ServiceAccounts: delta,
		ValidatorKeys:   iota,
		Authorizers:     varphi,
		Bless:           chi.Bless,
		Assign:          chi.Assign,
		Designate:       chi.Designate,
		CreateAcct:      chi.CreateAcct,
		AlwaysAccum:     chi.AlwaysAccum,
	}

	// \chi_g
	chi_g := chi.AlwaysAccum

	// (12.20)
	// Get g (max gas used)
	g := calculateMaxGasUsed(chi_g)

	// (12.21)
	// Execute outer accumulation
	outerAccumulationInput := OuterAccumulationInput{
		GasLimit:                     g,
		DeferredTransfers:            []types.DeferredTransfer{}, // empty
		WorkReports:                  accumulatableWorkReports,
		InitPartialStateSet:          partialStateSet,
		ServicesWithFreeAccumulation: chi_g,
	}
	output, err := OuterAccumulation(outerAccumulationInput)
	if err != nil {
		return OuterAccumulationOutput{}, err
	}
	b := output.AccumulatedServiceOutput
	ePrime := output.PartialStateSet
	// (12.22)
	// Update the partial state set to posterior state
	updatePartialStateSetToPosteriorState(cs, ePrime)

	// (12.26) θ′ ≡ [[(s, h) ∈ b]]
	// Convert accumulated service output (b) to the accumulation output log (θ′)
	var thetaPrime types.LastAccOut
	for accumulatedServiceHash := range b {
		// append accumulatedServiceHash to lastAccOut
		thetaPrime = append(thetaPrime, accumulatedServiceHash)
	}

	sort.Slice(thetaPrime, func(i, j int) bool {
		if thetaPrime[i].ServiceId != thetaPrime[j].ServiceId {
			return thetaPrime[i].ServiceId < thetaPrime[j].ServiceId
		}
		return bytes.Compare(thetaPrime[i].Hash[:], thetaPrime[j].Hash[:]) < 0
	})

	cs.GetPosteriorStates().SetLastAccOut(thetaPrime)

	return output, nil
}

// (v0.6.4) 12.3 Deferred Transfers And State Integration.
func DeferredTransfers() error {
	// Get parameters from the cs
	cs := blockchain.GetInstance()

	// (12.20) (12.21) (12.22)
	output, err := executeOuterAccumulation(cs)
	if err != nil {
		return err
	}
	n := output.NumberOfWorkResultsAccumulated // n
	u := output.ServiceGasUsedList             // u
	// (12.23) (12.24) (12.25)
	// Calculate the accumulation statistics I
	// (12.28–12.29) S ≡ {(s ↦ (G(s), N(s))) | G(s)+N(s) ≠ 0}
	S := calculateAccumulationStatistics(u, n)
	cs.GetIntermediateStates().SetAccumulationStatistics(S)

	// (12.27) (12.28) (12.29) (12.30)
	updateDeltaDoubleDagger(cs, S)

	// (12.31) (12.32)
	// Update the AccumulatedQueue(AccumulatedQueue)
	updateXi(cs, n)

	// (12.33)
	// Update ReadyQueue(Theta)
	updateTheta(cs)

	return nil
}
