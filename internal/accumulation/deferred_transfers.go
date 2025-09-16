package accumulation

import (
	"bytes"
	"slices"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
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

func updatePartialStateSetToPosteriorState(store *store.Store, o types.PartialStateSet) {
	// (12.22)
	deltaDagger := o.ServiceAccounts
	postIota := o.ValidatorKeys
	postVarphi := o.Authorizers

	// Update the posterior state
	store.GetPosteriorStates().SetChi(types.Privileges{
		Bless:       o.Bless,
		Assign:      o.Assign,
		Designate:   o.Designate,
		AlwaysAccum: o.AlwaysAccum,
	})
	store.GetIntermediateStates().SetDeltaDagger(deltaDagger)
	store.GetPosteriorStates().SetIota(postIota)
	store.GetPosteriorStates().SetVarphi(postVarphi)
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
	store := store.GetInstance()
	accumulatableWorkReports := store.GetIntermediateStates().GetAccumulatableWorkReports()

	output := []types.WorkResult{}

	for _, workReport := range accumulatableWorkReports[:n] {
		workResult := workReport.Results
		for _, result := range workResult {
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
// We save the accumulation statistics in the store
func calculateAccumulationStatistics(serviceGasUsedList types.ServiceGasUsedList, n types.U64) types.AccumulationStatistics {
	// Sum of gas used of the service
	// service id map to sum of gas used
	sumOfGasUsedMap := map[types.ServiceId]types.Gas{}
	for _, serviceGasUsed := range serviceGasUsedList {
		sumOfGasUsedMap[serviceGasUsed.ServiceId] += serviceGasUsed.Gas
	}

	// calcualte the number of work reports accumulated
	accumulationStatistics := types.AccumulationStatistics{}
	for serviceId, sumOfGasUsed := range sumOfGasUsedMap {
		numOfWorkReportsAccumulated := types.U64(len(getWorkResultByService(serviceId, n)))
		if numOfWorkReportsAccumulated == 0 {
			continue // skip, N(S) = []
		}
		accumulationStatistics[serviceId] = types.GasAndNumAccumulatedReports{
			Gas:                   sumOfGasUsed,
			NumAccumulatedReports: numOfWorkReportsAccumulated,
		}
	}
	return accumulationStatistics
}

// (12.26)
// R: Selection function
// ordered primarily according to the source service index and secondarily their order within t.
func selectionFunction(transfers types.DeferredTransfers, destinationServiceId types.ServiceId) types.DeferredTransfers {
	// Filter the destination service id
	outputTransfers := []types.DeferredTransfer{}
	for _, transfer := range transfers {
		if transfer.ReceiverID == destinationServiceId {
			outputTransfers = append(outputTransfers, transfer)
		}
	}

	// https://pkg.go.dev/sort#SliceStable
	sort.SliceStable(outputTransfers, func(i, j int) bool {
		return outputTransfers[i].SenderID < outputTransfers[j].SenderID
	})

	return outputTransfers
}

// (12.27) (12.28) (12.29) (12.30)
// delta double dagger: Second intermediate state
// On-Transfer service-account invocation function as ΨT
// INFO: t from the outer accumulation function
func updateDeltaDoubleDagger(store *store.Store, t types.DeferredTransfers, accumulationStatistics types.AccumulationStatistics) {
	// Get delta dagger
	deltaDagger := store.GetIntermediateStates().GetDeltaDagger()
	tauPrime := store.GetPosteriorStates().GetTau()

	// Call OnTransferInvoke
	deltaDoubleDagger := types.ServiceAccountState{}
	deferredTransfersStatisics := types.DeferredTransfersStatistics{}

	for serviceId := range deltaDagger {
		selectionFunctionOutput := selectionFunction(t, serviceId)
		onTransferInput := PVM.OnTransferInput{
			ServiceAccounts:   deltaDagger,
			Timeslot:          tauPrime,
			ServiceID:         serviceId,
			DeferredTransfers: selectionFunctionOutput,
		}

		// (12.27) x
		serviceAccount, gas := PVM.OnTransferInvoke(onTransferInput)

		// === (12.32) apply a'_a = τ′ for s ∈ K(S) ===
		if _, ok := accumulationStatistics[serviceId]; ok {
			serviceAccount.ServiceInfo.LastAccumulationSlot = tauPrime
		}
		deltaDoubleDagger[serviceId] = serviceAccount

		// Calculate transfers statistics (X)
		// (12.29) (12.30) Calculate the deferred transfers statistics
		if len(selectionFunctionOutput) != 0 {
			deferredTransfersStatisics[serviceId] = types.NumDeferredTransfersAndTotalGasUsed{
				NumDeferredTransfers: types.U64(len(selectionFunctionOutput)),
				TotalGasUsed:         gas,
			}
		}

	}

	// Update delta double dagger
	store.GetIntermediateStates().SetDeltaDoubleDagger(deltaDoubleDagger)

	// Save the deferred transfers statistics
	store.GetIntermediateStates().SetDeferredTransfersStatistics(deferredTransfersStatisics)
}

// (12.31) (12.32)
// Update the AccumulatedQueue(AccumulatedQueue)
func updateXi(store *store.Store, n types.U64) {
	// Get W^* (accumulatable work-reports in this block)
	accumulatableWorkReports := store.GetIntermediateStates().GetAccumulatableWorkReports()

	priorXi := store.GetPriorStates().GetXi()
	posteriorXi := store.GetPosteriorStates().GetXi()

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

	// Update posteriorXi to store
	store.GetPosteriorStates().SetXi(posteriorXi)
}

// (12.33)
// Update ReadyQueue(Theta)
func updateTheta(store *store.Store) {
	// (12.10) let m = H_t mode E
	headerSlot := store.GetLatestBlock().Header.Slot
	m := int(headerSlot) % types.EpochLength

	// (6.2) tau and tau prime
	// Get previous time slot index
	tau := store.GetPriorStates().GetTau()

	// Get current time slot index
	tauPrime := store.GetPosteriorStates().GetTau()

	tauOffset := tauPrime - tau

	// Get queued work reports
	queueWorkReports := store.GetIntermediateStates().GetQueuedWorkReports()

	// Get prior theta and posterior theta (ReadyQueue)
	priorTheta := store.GetPriorStates().GetTheta()
	posteriorTheta := store.GetPosteriorStates().GetTheta()

	// Get posterior xi
	posteriorXi := store.GetPosteriorStates().GetXi()

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
	store.GetPosteriorStates().SetTheta(posteriorTheta)
}

// (12.20) (12.21)
func executeOuterAccumulation(store *store.Store) (OuterAccumulationOutput, error) {
	// Get W^* (accumulatable work-reports in this block)
	accumulatableWorkReports := store.GetIntermediateStates().GetAccumulatableWorkReports()

	// (12.13) PartialStateSet
	priorState := store.GetPriorStates()
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
		WorkReports:                  accumulatableWorkReports,
		InitPartialStateSet:          partialStateSet,
		ServicesWithFreeAccumulation: chi_g,
	}
	output, err := OuterAccumulation(outerAccumulationInput)
	if err != nil {
		return OuterAccumulationOutput{}, err
	}

	// (12.22)
	// Update the partial state set to posterior state
	updatePartialStateSetToPosteriorState(store, output.PartialStateSet)

	// Convert AccumulatedServiceOutput to LastAccOut and assign it to the store
	var lastAccOut types.LastAccOut
	for accumulatedServiceHash := range output.AccumulatedServiceOutput {
		// append accumulatedServiceHash to lastAccOut
		lastAccOut = append(lastAccOut, accumulatedServiceHash)
	}
	store.GetPosteriorStates().SetLastAccOut(lastAccOut)

	return output, nil
}

// (v0.6.4) 12.3 Deferred Transfers And State Integration.
func DeferredTransfers() error {
	// Get parameters from the store
	store := store.GetInstance()

	// (12.20) (12.21) (12.22)
	output, err := executeOuterAccumulation(store)
	if err != nil {
		return err
	}

	// (12.23) (12.24) (12.25)
	accumulationStatistics := calculateAccumulationStatistics(output.ServiceGasUsedList, output.NumberOfWorkResultsAccumulated)
	store.GetIntermediateStates().SetAccumulationStatistics(accumulationStatistics)

	// (12.27) (12.28) (12.29) (12.30)
	updateDeltaDoubleDagger(store, output.DeferredTransfers, accumulationStatistics)

	// (12.31) (12.32)
	// Update the AccumulatedQueue(AccumulatedQueue)
	updateXi(store, output.NumberOfWorkResultsAccumulated)

	// (12.33)
	// Update ReadyQueue(Theta)
	updateTheta(store)

	return nil
}
