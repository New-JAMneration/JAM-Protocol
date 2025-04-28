// 13. Validator Activity Statistics

// pi
// b: The number of blocks produced by the validator.
// t: The number of tickets introduced by the validator.
// p: The number of preimages introduced by the validator.
// d: The total number of octets across all preimages introduced by the
// validator.
// g: The number of reports guaranteed by the validator.
// a: The number of availability assurances made by the validator.

// 這個 pi 會統計整個 epoch 時間範圍中的所有 validator 的活動情況。
// 一個 epoch 可能有多個 block author (validator),
// 因此，可以持續的統計每個 validator 的活動情況。

package statistics

import (
	"math"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// (13.3)
func GetEpochIndex(t types.TimeSlot) types.TimeSlot {
	return types.TimeSlot(math.Floor(float64(t) / float64(types.EpochLength)))
}

// b: The number of blocks produced by the validator.
func UpdateBlockStatistics(statistics *types.Statistics, authorIndex types.ValidatorIndex) {
	statistics.ValsCurrent[authorIndex].Blocks++
}

// t: The number of tickets introduced by the validator.
func UpdateTicketStatistics(statistics *types.Statistics, authorIndex types.ValidatorIndex, tickets types.TicketsExtrinsic) {
	// Only update the number of tickets for the author of the block.
	statistics.ValsCurrent[authorIndex].Tickets += types.U32(len(tickets))
}

// p: The number of preimages introduced by the validator.
func UpdatePreimageStatistics(statistics *types.Statistics, authorIndex types.ValidatorIndex, preimages types.PreimagesExtrinsic) {
	// Only update the number of preimages for the author of the block.
	statistics.ValsCurrent[authorIndex].PreImages += types.U32(len(preimages))
}

// d: The total number of octets across all preimages introduced by the
// validator.
func UpdatePreimageOctetStatistics(statistics *types.Statistics, authorIndex types.ValidatorIndex, preimages types.PreimagesExtrinsic) {
	// Only update the number of preimage size for the author of the block.
	for _, preimage := range preimages {
		statistics.ValsCurrent[authorIndex].PreImagesSize += types.U32(len(preimage.Blob))
	}
}

// g: The number of reports guaranteed by the validator.
// We note that the Ed25519 key of each validator whose
// signature is in a credential is placed in the reporters set R.
func UpdateReportStatistics(statistics *types.Statistics, authorIndex types.ValidatorIndex, reports types.GuaranteesExtrinsic) {
	// Check if the author is in the reporters set R.
	// If the author is in the reporters set R, then update the statistics.
	for _, report := range reports {
		for _, signature := range report.Signatures {
			statistics.ValsCurrent[signature.ValidatorIndex].Guarantees++
		}
	}
}

// a: The number of availability assurances made by the validator.
func UpdateAvailabilityStatistics(statistics *types.Statistics, authorIndex types.ValidatorIndex, assurances types.AssurancesExtrinsic) {
	for _, assurance := range assurances {
		statistics.ValsCurrent[assurance.ValidatorIndex].Assurances++
	}
}

func UpdateCurrentStatistics(extrinsic types.Extrinsic) {
	// Get current slot
	s := store.GetInstance()

	// Get author index
	authorIndex := s.GetProcessingBlockPointer().GetAuthorIndex()

	// Get statistics
	statistics := s.GetPosteriorStates().GetPi()

	UpdateBlockStatistics(&statistics, authorIndex)
	UpdateTicketStatistics(&statistics, authorIndex, extrinsic.Tickets)
	UpdatePreimageStatistics(&statistics, authorIndex, extrinsic.Preimages)
	UpdatePreimageOctetStatistics(&statistics, authorIndex, extrinsic.Preimages)
	UpdateReportStatistics(&statistics, authorIndex, extrinsic.Guarantees)
	UpdateAvailabilityStatistics(&statistics, authorIndex, extrinsic.Assurances)

	// Update current statistics
	s.GetPosteriorStates().SetPiCurrent(statistics.ValsCurrent)
}

type (
	CoreWorkReportMap map[types.CoreIndex]types.WorkReport
)

// INFO: The type of varaibles are different between RefineLoad, ServiceActivityRecord and CoreActivityRecord
// Now, we convert them to the same type in this file.
type Pi_C_R_Output struct {
	Imports        types.U16 // i
	ExtrinsicCount types.U16 // x
	ExtrinsicSize  types.U32 // z
	Exports        types.U16 // e
	GasUsed        types.U64 // u
	BundleSize     types.U32 // b
}

type Pi_S_R_Output struct {
	n              types.U32 // n
	GasUsed        types.U64 // u
	Imports        types.U32 // i
	ExtrinsicCount types.U32 // x
	ExtrinsicSize  types.U32 // z
	Exports        types.U32 // e
}

// (13.8) p
func CalculatePopularity(coreIndex types.CoreIndex, assurancesExtrinsic types.AssurancesExtrinsic) types.U16 {
	output := types.U16(0)
	for _, assurance := range assurancesExtrinsic {
		output += types.U16(assurance.Bitfield[coreIndex])
	}

	return output
}

// (13.10) D
func CalculateDALoad(coreIndex types.CoreIndex, WMap CoreWorkReportMap) types.U32 {
	workReport, ok := WMap[coreIndex]
	if !ok {
		return 0
	}

	ceilValue := types.U32(math.Ceil(float64(workReport.PackageSpec.ExportsCount) * 65 / 64))
	output := workReport.PackageSpec.Length + types.SegmentSize*ceilValue

	return output
}

// (13.9) R
func CalculateWorkResults(coreIndex types.CoreIndex, wMap CoreWorkReportMap) Pi_C_R_Output {
	workReport, ok := wMap[coreIndex]
	if !ok {
		return Pi_C_R_Output{}
	}

	output := Pi_C_R_Output{}

	for _, workResult := range workReport.Results {
		output.Imports += workResult.RefineLoad.Imports
		output.ExtrinsicCount += workResult.RefineLoad.ExtrinsicCount
		output.ExtrinsicSize += workResult.RefineLoad.ExtrinsicSize
		output.Exports += workResult.RefineLoad.Exports
		output.GasUsed += workResult.RefineLoad.GasUsed
		output.BundleSize = workReport.PackageSpec.Length
	}

	return output
}

func createWorkReportMap(workReports []types.WorkReport) CoreWorkReportMap {
	workReportMap := make(CoreWorkReportMap)
	for _, workReport := range workReports {
		workReportMap[workReport.CoreIndex] = workReport
	}
	return workReportMap
}

// (13.8)
func UpdateCoreActivityStatistics(extrinsic types.Extrinsic) {
	store := store.GetInstance()

	// **w**: the incoming work-reports (11.28)
	// **W**: the newly available work-reports (11.16)
	w := store.GetIntermediateStates().GetPresentWorkReports()
	W := store.GetIntermediateStates().GetAvailableWorkReports()

	// Create a map for the work reports
	wMap := createWorkReportMap(w)
	WMap := createWorkReportMap(W)

	// Initialize the cores statistics (13.6)
	coreActivityStatisitics := make([]types.CoreActivityRecord, types.CoresCount)

	for i := 0; i < types.CoresCount; i++ {
		coreIndex := types.CoreIndex(i)
		R := CalculateWorkResults(coreIndex, wMap)
		D := CalculateDALoad(coreIndex, WMap)
		p := CalculatePopularity(coreIndex, extrinsic.Assurances)

		coreActivityStatisitics[coreIndex] = types.CoreActivityRecord{
			Imports:        R.Imports,        // i
			ExtrinsicCount: R.ExtrinsicCount, // x
			ExtrinsicSize:  R.ExtrinsicSize,  // z
			Exports:        R.Exports,        // e
			GasUsed:        R.GasUsed,        // u
			BundleSize:     R.BundleSize,     // b
			DALoad:         D,                // D
			Popularity:     p,                // p
		}
	}

	// Set the core activity statistics to the store
	store.GetPosteriorStates().SetCoresStatistics(coreActivityStatisitics)
}

type ServiceWorkResultsMap map[types.ServiceId][]types.WorkResult

// Create a map (service Id -> []work result)
func CreateServiceWorkResultsMap() ServiceWorkResultsMap {
	store := store.GetInstance()
	w := store.GetIntermediateStates().GetPresentWorkReports()

	// Create a map to store the service id map to work results
	serviceWorkResultsMap := make(ServiceWorkResultsMap)

	// Get all work reports from work results
	for _, workReport := range w {
		for _, result := range workReport.Results {
			serviceWorkResultsMap[result.ServiceId] = append(serviceWorkResultsMap[result.ServiceId], result)
		}
	}

	return serviceWorkResultsMap
}

// (13.13) r
// Get services from the coming work reports
func GetServicesFromPresentWorkReport() []types.ServiceId {
	store := store.GetInstance()
	w := store.GetIntermediateStates().GetPresentWorkReports()

	services := []types.ServiceId{}

	for _, workReport := range w {
		for _, result := range workReport.Results {
			services = append(services, result.ServiceId)
		}
	}

	return services
}

// (13.14) p
func GetServicesFromPreimagesExtrinsic(preimagesExtrinsic types.PreimagesExtrinsic) []types.ServiceId {
	servicesMap := make(map[types.ServiceId]bool)

	for _, preimage := range preimagesExtrinsic {
		serviceId := preimage.Requester
		if _, exists := servicesMap[serviceId]; !exists {
			servicesMap[serviceId] = true
		}
	}

	services := make([]types.ServiceId, 0, len(servicesMap))
	for key := range servicesMap {
		services = append(services, key)
	}

	return services
}

func GetServicesFromAccumulationStatistics() []types.ServiceId {
	store := store.GetInstance()

	// Get the accumulation statistics (I)
	accumulationStatistics := store.GetAccumulationStatisticsPointer().GetAccumulationStatistics()

	services := make([]types.ServiceId, 0, len(accumulationStatistics))
	for key := range accumulationStatistics {
		services = append(services, key)
	}

	return services
}

func GetServicesFromDeferredTransfersStatistics() []types.ServiceId {
	store := store.GetInstance()

	// Get the deferred transfers statistics (X)
	deferredTransfersStatistics := store.GetDeferredTransfersStatisticsPointer().GetDeferredTransfersStatistics()

	services := make([]types.ServiceId, 0, len(deferredTransfersStatistics))
	for key := range deferredTransfersStatistics {
		services = append(services, key)
	}

	return services
}

// s (13.12)
func GetAllServices(preimagesExtrinsic types.PreimagesExtrinsic) []types.ServiceId {
	// r: services from the incoming work-reports (13.13)
	r := GetServicesFromPresentWorkReport()
	// p: services from the preimages extrinsic (13.14)
	p := GetServicesFromPreimagesExtrinsic(preimagesExtrinsic)
	// I: services from the accumulation statistics (12.23)
	I := GetServicesFromAccumulationStatistics()
	// X: services from the deferred transfers statistics (12.29)
	X := GetServicesFromDeferredTransfersStatistics()

	// Merge all services (without duplicates)
	servicesMap := make(map[types.ServiceId]bool)
	for _, serviceId := range r {
		servicesMap[serviceId] = true
	}

	for _, serviceId := range p {
		servicesMap[serviceId] = true
	}

	for _, serviceId := range I {
		servicesMap[serviceId] = true
	}

	for _, serviceId := range X {
		servicesMap[serviceId] = true
	}

	services := make([]types.ServiceId, 0, len(servicesMap))
	for key := range servicesMap {
		services = append(services, key)
	}

	return services
}

// (13.15)
func CalculateServiceResults(serviceId types.ServiceId, serviceWorkResultsMap ServiceWorkResultsMap) Pi_S_R_Output {
	workResults, ok := serviceWorkResultsMap[serviceId]
	if !ok {
		return Pi_S_R_Output{}
	}

	output := Pi_S_R_Output{}

	for _, workResult := range workResults {
		output.n += 1
		output.GasUsed += workResult.RefineLoad.GasUsed
		output.Imports += types.U32(workResult.RefineLoad.Imports)
		output.ExtrinsicCount += types.U32(workResult.RefineLoad.ExtrinsicCount)
		output.ExtrinsicSize += workResult.RefineLoad.ExtrinsicSize
		output.Exports += types.U32(workResult.RefineLoad.Exports)
	}

	return output
}

// p
func CalculateProvidedStatistics(serviceId types.ServiceId, preimagesExtrinsic types.PreimagesExtrinsic) (providedCount types.U16, providedSize types.U32) {
	providedCount = 0
	providedSize = 0

	for _, preimage := range preimagesExtrinsic {
		if preimage.Requester != serviceId {
			continue
		}
		providedCount += 1
		providedSize += types.U32(len(preimage.Blob))
	}

	return providedCount, providedSize
}

// a
// AccumulateCount, AccumulateGasUsed
func CalculateAccumulationStatistics(serviceId types.ServiceId, accumulationStatistics types.AccumulationStatistics) (accumulateCount types.U32, accumulateGasUsed types.U64) {
	accumulateCount = 0
	accumulateGasUsed = 0

	value, ok := accumulationStatistics[serviceId]
	if ok {
		accumulateCount = types.U32(value.NumAccumulatedReports)
		accumulateGasUsed = types.U64(value.Gas)
	} else {
		// If the service id is not found, return 0
		accumulateCount = 0
		accumulateGasUsed = 0
	}

	return accumulateCount, accumulateGasUsed
}

// t
// OnTransfersCount, OnTransfersGasUsed
func CalculateTransfersStatistics(serviceId types.ServiceId, deferredTransfersStatistics types.DeferredTransfersStatistics) (onTransfersCount types.U32, onTransfersGasUsed types.U64) {
	onTransfersCount = 0
	onTransfersGasUsed = 0

	value, ok := deferredTransfersStatistics[serviceId]
	if ok {
		onTransfersCount = types.U32(value.NumDeferredTransfers)
		onTransfersGasUsed = types.U64(value.TotalGasUsed)
	} else {
		// If the service id is not found, return 0
		onTransfersCount = 0
		onTransfersGasUsed = 0
	}

	return onTransfersCount, onTransfersGasUsed
}

// (13.11)
func UpdateServiceActivityStatistics(extrinsic types.Extrinsic) {
	store := store.GetInstance()
	accumulationStatisitcs := store.GetAccumulationStatisticsPointer().GetAccumulationStatistics()
	transfersStatistics := store.GetDeferredTransfersStatisticsPointer().GetDeferredTransfersStatistics()

	services := GetAllServices(extrinsic.Preimages)
	serviceWorkResultsMap := CreateServiceWorkResultsMap()

	// Initialize the services statistics (13.7)
	servicesStatistics := make(types.ServicesStatistics)

	for _, serviceId := range services {
		// Calculate the service results (R)
		R := CalculateServiceResults(serviceId, serviceWorkResultsMap)

		// p
		providedCount, providedSize := CalculateProvidedStatistics(serviceId, extrinsic.Preimages)

		// a
		accumulateCount, accumulateGasUsed := CalculateAccumulationStatistics(serviceId, accumulationStatisitcs)

		// t
		onTransfersCount, onTransfersGasUsed := CalculateTransfersStatistics(serviceId, transfersStatistics)

		servicesStatistics[serviceId] = types.ServiceActivityRecord{
			ProvidedCount:      providedCount,
			ProvidedSize:       providedSize,
			RefinementCount:    R.n,
			RefinementGasUsed:  R.GasUsed,
			Imports:            R.Imports,
			Exports:            R.Exports,
			ExtrinsicSize:      R.ExtrinsicSize,
			ExtrinsicCount:     R.ExtrinsicCount,
			AccumulateCount:    accumulateCount,
			AccumulateGasUsed:  accumulateGasUsed,
			OnTransfersCount:   onTransfersCount,
			OnTransfersGasUsed: onTransfersGasUsed,
		}
	}

	// Set the service activity statistics to the store
	if len(servicesStatistics) == 0 {
		servicesStatistics = nil
	}
	store.GetPosteriorStates().SetServicesStatistics(servicesStatistics)
}

// (13.3)
// π ≡ (πV , πL, πC , πS)
// (πV, πL) => (current, last)
func UpdateValidatorActivityStatistics(extrinsic types.Extrinsic) {
	s := store.GetInstance()

	preTau := s.GetPriorStates().GetTau()
	postTau := s.GetPosteriorStates().GetTau()

	preEpochIndex := GetEpochIndex(preTau)
	postEpochIndex := GetEpochIndex(postTau)

	preStatistics := s.GetPriorStates().GetPi()

	if preEpochIndex == postEpochIndex {
		// If the epoch index is the same, we will keep using the same statistics.
		valsCurrent := preStatistics.ValsCurrent
		valsLast := preStatistics.ValsLast
		s.GetPosteriorStates().SetPiCurrent(valsCurrent)
		s.GetPosteriorStates().SetPiLast(valsLast)
	} else {
		// If the epoch index is different, we will reset the statistics.
		valsCurrent := make(types.ActivityRecords, types.ValidatorsCount)
		valsLast := preStatistics.ValsCurrent
		s.GetPosteriorStates().SetPiCurrent(valsCurrent)
		s.GetPosteriorStates().SetPiLast(valsLast)
	}

	// Update current statistics.
	UpdateCurrentStatistics(extrinsic)

	// Update the cores statistics.
	UpdateCoreActivityStatistics(extrinsic)

	// Update the services statistics.
	UpdateServiceActivityStatistics(extrinsic)
}
