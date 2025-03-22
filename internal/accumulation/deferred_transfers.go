package accumulation

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/PolkaVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func sumPrivilegesGas(privileges types.Privileges) types.Gas {
	sum := types.Gas(0)
	for _, value := range privileges.AlwaysAccum {
		sum += value
	}

	return sum
}

// Calculate max gas used v0.6.4 (12.20)
func calculateMaxGasUsed() types.Gas {
	GT := types.Gas(types.TotalGas)
	GA := types.Gas(types.MaxAccumulateGas)
	C := types.Gas(types.CoresCount)

	store := store.GetInstance()
	priorPrivileges := store.GetPriorStates().GetChi()

	sum := sumPrivilegesGas(priorPrivileges)

	return max(GT, GA*C+sum)
}

func updatePartialStateSet(o types.PartialStateSet) {
	// (12.22)
	postChi := o.Privileges
	deltaDagger := o.ServiceAccounts
	postIota := o.ValidatorKeys
	postVarphi := o.Authorizers

	store := store.GetInstance()

	// Update the posterior state
	store.GetPosteriorStates().SetChi(postChi)
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
func getWorkResultByService(s types.ServiceId) []types.WorkResult {
	// Get W^*
	store := store.GetInstance()
	accumulatableWorkReports := store.GetAccumulatableWorkReports()

	output := []types.WorkResult{}

	for _, workReport := range accumulatableWorkReports {
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
func calculateAccumulationStatistics(serviceGasUsedList types.ServiceGasUsedList) {
	// Sum of gas used of the service
	// service id map to sum of gas used
	sumOfGasUsedMap := map[types.ServiceId]types.Gas{}
	for _, serviceGasUsed := range serviceGasUsedList {
		sumOfGasUsedMap[serviceGasUsed.ServiceId] += serviceGasUsed.Gas
	}

	// calcualte the number of work reports accumulated
	accumulationStatistics := types.AccumulationStatistics{}
	for serviceId, sumOfGasUsed := range sumOfGasUsedMap {
		numOfWorkReportsAccumulated := types.U64(len(getWorkResultByService(serviceId)))

		accumulationStatistics[serviceId] = types.GasAndNumAccumulatedReports{
			Gas:                   sumOfGasUsed,
			NumAccumulatedReports: numOfWorkReportsAccumulated,
		}
	}

	store := store.GetInstance()
	store.GetAccumulationStatistics().SetAccumulationStatistics(accumulationStatistics)
}

// (12.26)
// R: Selection function
func SelectionFunction(transfers types.DeferredTransfers, destinationServiceId types.ServiceId) types.DeferredTransfers {
	// ordered primarily according to the source service index and secondarily their order within t.
	// FIXME: When should we order the delta map with key?

	// FIXME: Where can I get all the services? delta keys?
	// intermediate delta?

	store := store.GetInstance()
	delta := store.GetIntermediateStates().GetDeltaDagger()

	// Get all the services from delta keys
	services := []types.ServiceId{}
	for serviceId := range delta {
		services = append(services, serviceId)
	}

	output := []types.DeferredTransfer{}

	for _, service := range services {
		for _, transfer := range transfers {
			condition_1 := transfer.SenderID == service
			condition_2 := transfer.ReceiverID == destinationServiceId

			if condition_1 && condition_2 {
				output = append(output, transfer)
			}
		}
	}

	return output
}

// (12.27) (12.28)
// delta double dagger: Second intermediate state
// On-Transfer service-account invocation function as ΨT
// INFO: t from the outer accumulation function
func updateDeltaDoubleDagger(t types.DeferredTransfers) {
	// Get delta dagger
	store := store.GetInstance()
	deltaDagger := store.GetIntermediateStates().GetDeltaDagger()
	tauPrime := store.GetPosteriorStates().GetTau()

	// Call OnTransferInvoke

	tempDeltaDagger := types.ServiceAccountState{}
	deferredTransfersStatisics := types.DeferredTransfersStatistics{}

	for serviceId := range deltaDagger {
		onTransferInput := PolkaVM.OnTransferInput{
			ServiceAccounts:   deltaDagger,
			Timeslot:          tauPrime,
			ServiceID:         serviceId,
			DeferredTransfers: SelectionFunction(t, serviceId),
		}

		// (12.27) x
		serviceAccount, gas := PolkaVM.OnTransferInvoke(onTransferInput)

		// (12.28)
		tempDeltaDagger[serviceId] = serviceAccount

		// Calculate transfers statistics (X)
		selectionFunctionOutput := SelectionFunction(t, serviceId)

		if len(selectionFunctionOutput) != 0 {
			deferredTransfersStatisics[serviceId] = types.NumDeferredTransfersAndTotalGasUsed{
				NumDeferredTransfers: types.U64(len(selectionFunctionOutput)),
				TotalGasUsed:         gas,
			}
		}
	}

	// Update delta double dagger
	store.GetIntermediateStates().SetDeltaDoubleDagger(tempDeltaDagger)

	// Save the deferred transfers statistics
	store.GetDeferredTransfersStatistics().SetDeferredTransfersStatistics(deferredTransfersStatisics)
}

func DeferredTransfers() {
	// (12.20)
	// Get g (max gas used)
	g := calculateMaxGasUsed()

	// (12.21)
	// Execute outer accumulation
	store := store.GetInstance()

	// Get W^* (accumulatable work-reports in this block)
	accumulatableWorkReports := store.GetAccumulatableWorkReports()

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
		Privileges:      chi,
	}

	// \chi_g
	chi_g := chi.AlwaysAccum

	// (12.21)
	// n, o, t, C, u
	outerAccumulationInput := OuterAccumulationInput{
		GasLimit:                     g,
		WorkReports:                  accumulatableWorkReports,
		InitPartialStateSet:          partialStateSet,
		ServicesWithFreeAccumulation: chi_g,
	}
	output, err := OuterAccumulation(outerAccumulationInput)
	if err != nil {
		fmt.Errorf("OuterAccumulation failed: %v", err)
	}

	// (12.22)
	// Update the partial state set
	updatePartialStateSet(output.PartialStateSet)

	// (12.23) (12.24) (12.25)
	calculateAccumulationStatistics(output.ServiceGasUsedList)

	// (12.27) (12.28)
	updateDeltaDoubleDagger(output.DeferredTransfers)

	// (12.31)
	// Update the AccumulatedQueue(AccumulatedHistories)
	priorXi := store.GetPriorStates().GetXi()
	posteriorXi := store.GetPosteriorStates().GetXi()

	// Update the last element
	posteriorXi[types.EpochLength-1] = MappingFunction(accumulatableWorkReports)

	// (12.32)
	// Update the rest of the elements
	for i := 0; i < types.EpochLength-1; i++ {
		posteriorXi[i] = priorXi[i+1]
	}

	// Update posteriorXi to store
	store.GetPosteriorStates().SetXi(posteriorXi)

	// (12.33)
	// Update ReadyQueue(Theta)

	// (12.10) let m = H_t mode E
	headerSlot := store.GetIntermediateHeaderPointer().GetSlot()
	m := int(headerSlot) % types.EpochLength

	// (6.2) tau and tau prime
	// Get previous time slot index
	tau := store.GetPriorStates().GetTau()

	// Get current time slot index
	tauPrime := store.GetPosteriorStates().GetTau()

	tauOffset := tauPrime - tau

	// Get queued work reports
	queueWorkReports := store.GetQueuedWorkReports()

	// Get prior theta and posterior theta (ReadyQueue)
	priorTheta := store.GetPriorStates().GetTheta()
	posteriorTheta := store.GetPosteriorStates().GetTheta()

	for i := 0; i < types.EpochLength; i++ {
		// s[i]↺ ≡ s[ i % ∣s∣ ]
		index := m - i
		index = index % len(posteriorTheta)

		firstCondition := i == 0
		secondCondition := (1 <= i) && (i < int(tauOffset))
		thirdCondition := i >= int(tauOffset)

		if firstCondition {
			posteriorTheta[index] = QueueEditingFunction(queueWorkReports, posteriorXi[types.EpochLength-1])
		}

		if secondCondition {
			posteriorTheta[index] = types.ReadyQueueItem{}
		}

		if thirdCondition {
			posteriorTheta[index] = QueueEditingFunction(priorTheta[index], posteriorXi[types.EpochLength-1])
		}
	}

	// Update posterior theta
	store.GetPosteriorStates().SetTheta(posteriorTheta)
}
