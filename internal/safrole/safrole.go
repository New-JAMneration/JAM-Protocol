package safrole

import (
	"slices"

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
// !Warning : epoch datatype is undefined in jamtypes and is uncertain
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
func GetBandersnatchRingRootCommmitment(bandersnatchKeys []types.BandersnatchPublic) types.BandersnatchRingCommitment {
	vrfHandler, _ := CreateVRFHandler(bandersnatchKeys)
	commitment, _ := vrfHandler.GetCommitment()
	return types.BandersnatchRingCommitment(commitment)
}

// UpdateBandersnatchKeyRoot returns the root commitment of the Bandersnatch
// ring
// Equation (6.15)
func UpdateBandersnatchKeyRoot(validators types.ValidatorsData) types.BandersnatchRingCommitment {
	bandersnatchKeys := []types.BandersnatchPublic{}
	for _, validator := range validators {
		bandersnatchKeys = append(bandersnatchKeys, validator.Bandersnatch)
	}

	return GetBandersnatchRingRootCommmitment(bandersnatchKeys)
}

// GetNewSafroleState returns the new Safrole state
// Equation (6.13)
func GetNewSafroleState(t types.TimeSlot, tPrime types.TimeSlot, safroleState types.State, offendersMark types.OffendersMark) (newSafroleState types.State) {
	e := GetEpochIndex(t)
	ePrime := GetEpochIndex(tPrime)

	if ePrime > e {
		// New epoch
		newSafroleState.Gamma.GammaK = UpdatePendingValidators(safroleState.Iota, offendersMark)
		newSafroleState.Kappa = safroleState.Gamma.GammaK
		newSafroleState.Lambda = safroleState.Kappa
		newSafroleState.Gamma.GammaZ = UpdateBandersnatchKeyRoot(safroleState.Gamma.GammaK)
		return newSafroleState
	} else {
		// Same epoch
		return safroleState
	}
}
