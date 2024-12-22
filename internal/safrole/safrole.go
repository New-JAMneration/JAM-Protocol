package safrole

import (
	"slices"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// R function return the epoch and slot index
// Equation (6.2)
// !Warning : epoch datatype is undefined in jamtypes and is uncertain
func R(time jamTypes.TimeSlot) (epoch jamTypes.TimeSlot, slotIndex jamTypes.TimeSlot) {
	epoch = time / jamTypes.TimeSlot(jamTypes.EpochLength)
	slotIndex = time % jamTypes.TimeSlot(jamTypes.EpochLength)
	return epoch, slotIndex
}

// GetEpochIndex returns the epoch index of the most recent block't timeslot
// \tau : The most recent block't timeslot
func GetEpochIndex(t jamTypes.TimeSlot) jamTypes.TimeSlot {
	return t / jamTypes.TimeSlot(jamTypes.EpochLength)
}

// GetSlotIndex returns the slot index of the most recent block't timeslot
// \tau : The most recent block't timeslot
func GetSlotIndex(t jamTypes.TimeSlot) jamTypes.TimeSlot {
	return t % jamTypes.TimeSlot(jamTypes.EpochLength)
}

// ValidatorIsOffender checks if the validator is an offender
// Equation (6.14)
func ValidatorIsOffender(validator ValidatorData, offendersMark jamTypes.OffendersMark) bool {
	// FIXME: 目前因為 jamTypes.Ed25519Public 與 safrole_types.Ed25519Key 不同,
	// 需要轉換
	convertedEd25519 := jamTypes.Ed25519Public{}
	for i, v := range validator.Ed25519 {
		convertedEd25519[i] = byte(v)
	}

	return slices.Contains(offendersMark, convertedEd25519)
}

// UpdatePendingValidators updates the pending validators
// Equation (6.14)
func UpdatePendingValidators(validators ValidatorsData, offendersMark jamTypes.OffendersMark) ValidatorsData {
	for i, validator := range validators {
		if ValidatorIsOffender(validator, offendersMark) {
			// Replace the validator's Ed25519 key with a null key
			validators[i].Ed25519 = Ed25519Key{}
		}
	}

	return validators
}

// GetBandersnatchRingRootCommmitment returns the root commitment of the
// Bandersnatch ring.
// O function: The Bandersnatch ring root function.
// See section 3.8 and appendix G.
func GetBandersnatchRingRootCommmitment(bandersnatchKeys []BandersnatchKey) [144]U8 {
	// FIXME: Call rust function
	return [144]U8{}
}

// UpdateBandersnatchKeyRoot returns the root commitment of the Bandersnatch
// ring
// Equation (6.15)
func UpdateBandersnatchKeyRoot(validators ValidatorsData) [144]U8 {
	bandersnatchKeys := []BandersnatchKey{}
	for _, validator := range validators {
		bandersnatchKeys = append(bandersnatchKeys, validator.Bandersnatch)
	}

	return GetBandersnatchRingRootCommmitment(bandersnatchKeys)
}

// GetNewSafroleState returns the new Safrole state
// Equation (6.13)
func GetNewSafroleState(t jamTypes.TimeSlot, tPrime jamTypes.TimeSlot, safroleState State, offendersMark jamTypes.OffendersMark) (newSafroleState State) {
	e := GetEpochIndex(t)
	ePrime := GetEpochIndex(tPrime)

	if ePrime > e {
		// New epoch
		newSafroleState.GammaK = UpdatePendingValidators(safroleState.Iota, offendersMark)
		newSafroleState.Kappa = safroleState.GammaK
		newSafroleState.Lambda = safroleState.Kappa
		newSafroleState.GammaZ = UpdateBandersnatchKeyRoot(safroleState.GammaK)
		return newSafroleState
	} else {
		// Same epoch
		return safroleState
	}
}
