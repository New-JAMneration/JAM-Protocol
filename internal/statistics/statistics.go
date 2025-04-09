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

// (13.2)
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

	// Update statistics
	s.GetPosteriorStates().SetPi(statistics)
}

// (13.3)
// (pi_0, pi_1) => (current, last)
func UpdateValidatorActivityStatistics(extrinsic types.Extrinsic) {
	s := store.GetInstance()

	preTau := s.GetPriorStates().GetTau()
	postTau := s.GetPosteriorStates().GetTau()

	preEpochIndex := GetEpochIndex(preTau)
	postEpochIndex := GetEpochIndex(postTau)

	preStatistics := s.GetPriorStates().GetPi()

	// Update current and last statistics if the epoch index is different.
	if preEpochIndex == postEpochIndex {
		// If the epoch index is the same, we will keep using the same statistics.
		s.GetPosteriorStates().SetPi(preStatistics)
	} else {
		// If the epoch index is different, we will reset the statistics.
		pi_0 := make(types.ActivityRecords, types.ValidatorsCount)
		pi_1 := preStatistics.ValsCurrent
		s.GetPosteriorStates().SetPi(types.Statistics{
			ValsCurrent: pi_0,
			ValsLast:    pi_1,
		})
	}

	// Update current statistics.
	UpdateCurrentStatistics(extrinsic)
}
