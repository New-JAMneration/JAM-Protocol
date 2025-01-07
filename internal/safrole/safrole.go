package safrole

import (
	"fmt"
	"slices"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// GetEpochIndex returns the epoch index of the most recent block't timeslot
// \tau : The most recent block't timeslot
// (6.2)
func GetEpochIndex(t types.TimeSlot) types.TimeSlot {
	return t / types.TimeSlot(types.EpochLength)
}

// GetSlotIndex returns the slot index of the most recent block't timeslot
// \tau : The most recent block't timeslot
// (6.2)
func GetSlotIndex(t types.TimeSlot) types.TimeSlot {
	return t % types.TimeSlot(types.EpochLength)
}

// R function return the epoch and slot index
// Equation (6.2)
func R(time types.TimeSlot) (epoch types.TimeSlot, slotIndex types.TimeSlot) {
	epoch = time / types.TimeSlot(types.EpochLength)
	slotIndex = time % types.TimeSlot(types.EpochLength)
	return epoch, slotIndex
}

// ValidatorIsOffender checks if the validator is an offender
// Equation (6.14)
func ValidatorIsOffender(validator types.Validator, offendersMark types.OffendersMark) bool {
	return slices.Contains(offendersMark, validator.Ed25519)
}

// UpdatePendingValidators updates the pending validators
// Equation (6.14)
func UpdatePendingValidators(validators types.ValidatorsData, offendersMark types.OffendersMark) types.ValidatorsData {
	for i, validator := range validators {
		if ValidatorIsOffender(validator, offendersMark) {
			// Replace the validator's Ed25519 key with a null key
			validators[i].Ed25519 = types.Ed25519Public{}
		}
	}

	return validators
}

// GetBandersnatchRingRootCommmitment returns the root commitment of the
// Bandersnatch ring.
// O function: The Bandersnatch ring root function.
// See section 3.8 and appendix G.
func GetBandersnatchRingRootCommmitment(bandersnatchKeys []types.BandersnatchPublic) (types.BandersnatchRingCommitment, error) {
	vrfHandler, handlerErr := CreateVRFHandler(bandersnatchKeys)
	if handlerErr != nil {
		return types.BandersnatchRingCommitment{}, handlerErr
	}
	defer vrfHandler.Free()

	commitment, commitmentErr := vrfHandler.GetCommitment()
	if commitmentErr != nil {
		return types.BandersnatchRingCommitment{}, commitmentErr
	}

	return types.BandersnatchRingCommitment(commitment), nil
}

// UpdateBandersnatchKeyRoot returns the root commitment of the Bandersnatch
// ring
// Equation (6.13)
func UpdateBandersnatchKeyRoot(validators types.ValidatorsData) (types.BandersnatchRingCommitment, error) {
	bandersnatchKeys := []types.BandersnatchPublic{}
	for _, validator := range validators {
		bandersnatchKeys = append(bandersnatchKeys, validator.Bandersnatch)
	}

	return GetBandersnatchRingRootCommmitment(bandersnatchKeys)
}

// GetNewSafroleState returns the new Safrole state
func GetNewSafroleState(t types.TimeSlot, tPrime types.TimeSlot, safroleState types.State, offendersMark types.OffendersMark) (newSafroleState types.State) {
	// Equation (6.13)
	e := GetEpochIndex(t)
	ePrime := GetEpochIndex(tPrime)

	if ePrime > e {
		// New epoch
		newSafroleState.Gamma.GammaK = UpdatePendingValidators(safroleState.Iota, offendersMark)
		newSafroleState.Kappa = safroleState.Gamma.GammaK
		newSafroleState.Lambda = safroleState.Kappa
		z, zErr := UpdateBandersnatchKeyRoot(safroleState.Gamma.GammaK)
		if zErr != nil {
			fmt.Printf("Error updating Bandersnatch key root: %v\n", zErr)
			return
		}
		newSafroleState.Gamma.GammaZ = z
		return newSafroleState
	} else {
		// Same epoch
		return safroleState
	}
}

// KeyRotate rotates the keys
// Update the state with the new Safrole state
// (6.13)
func KeyRotate() {
	s := store.GetInstance()

	// Get the most recent block
	block := s.GetLatestBlock()

	// Get prior state
	priorState := s.GetPriorState()

	// Get previous time slot index
	tau := priorState.Tau

	// Get current time slot
	now := time.Now().UTC()
	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))

	// Get offenders mark from header
	offendersMark := block.Header.OffendersMark

	// Execute key rotation
	newSafroleState := GetNewSafroleState(tau, tauPrime, priorState, offendersMark)

	// Update state to posterior state
	s.GetPosteriorStates().SetGammaK(newSafroleState.Gamma.GammaK)
	s.GetPosteriorStates().SetKappa(newSafroleState.Kappa)
	s.GetPosteriorStates().SetLambda(newSafroleState.Lambda)
	s.GetPosteriorStates().SetGammaZ(newSafroleState.Gamma.GammaZ)
}
