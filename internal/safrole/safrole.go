package safrole

import (
	"fmt"
	"slices"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
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
// Equation (6.14) Phi(k)
func ValidatorIsOffender(validator types.Validator, offendersMark types.OffendersMark) bool {
	return slices.Contains(offendersMark, validator.Ed25519)
}

// ReplaceOffenderKeys replaces the Ed25519 key of the validator with a null key
// Equation (6.14) Phi(k)
func ReplaceOffenderKeys(validators types.ValidatorsData) types.ValidatorsData {
	// Get offendersMark (Psi_O) from posterior state
	s := store.GetInstance()
	posteriorState := s.GetPosteriorStates()
	offendersMark := posteriorState.GetPsiO()

	for i, validator := range validators {
		if ValidatorIsOffender(validator, offendersMark) {
			// Replace the validator's keys with a null key
			validators[i].Bandersnatch = types.BandersnatchPublic{}
			validators[i].Ed25519 = types.Ed25519Public{}
			validators[i].Bls = types.BlsPublic{}
			validators[i].Metadata = types.ValidatorMetadata{}
		}
	}

	return validators
}

// GetBandersnatchRingRootCommmitment returns the root commitment of the
// Bandersnatch ring.
// O function: The Bandersnatch ring root function.
// See section 3.8 and appendix G.
func GetBandersnatchRingRootCommmitment(bandersnatchKeys []types.BandersnatchPublic) (types.BandersnatchRingCommitment, error) {
	ringBytes := []byte{}
	ringSize := uint(len(bandersnatchKeys))

	for _, bandersnatch := range bandersnatchKeys {
		ringBytes = append(ringBytes, bandersnatch[:]...)
	}

	verifier, err := vrf.NewVerifier(ringBytes, ringSize)
	if err != nil {
		fmt.Printf("Failed to create verifier: %v\n", err)
	}
	defer verifier.Free()

	commitment, commitmentErr := verifier.GetCommitment()
	if commitmentErr != nil {
		fmt.Printf("Failed to get commitment %v\n", commitmentErr)
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

// keyRotation rotates the keys, it updates the state with the new Safrole state
// Equation (6.13)
/*
func keyRotation(t types.TimeSlot, tPrime types.TimeSlot, safroleState types.State) (newSafroleState types.State) {
	e := GetEpochIndex(t)
	ePrime := GetEpochIndex(tPrime)

	if ePrime > e {
		// New epoch
		newSafroleState.Gamma.GammaK = ReplaceOffenderKeys(safroleState.Iota)
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
*/

// --- KeyRotate() is the outside usage of this safrole.go --- //
// KeyRotate rotates the keys
// Update the state with the new Safrole state
// (6.2, 6.13, 6.14)
func KeyRotate() {
	s := store.GetInstance()

	// Get prior state
	priorState := s.GetPriorStates()

	// Get previous time slot index
	tau := priorState.GetTau()

	// Get current time slot
	// tauPrime := s.GetIntermediateStates().GetTauInput()
	// var tauPrime types.TimeSlot
	// if s.GetIntermediateStates().GetTauInput() == 0 {
	// 	fmt.Println("tauPrime is 0")
	// 	now := time.Now().UTC()
	// 	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	// 	tauPrime = types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))
	// 	s.GetIntermediateStates().SetTauInput(tauPrime)
	// } else {
	// 	tauPrime = s.GetIntermediateStates().GetTauInput()
	// }
	tauPrime := s.GetIntermediateStates().GetTauInput()
	// Execute key rotation
	// newSafroleState := keyRotation(tau, tauPrime, priorState.GetState())
	e := GetEpochIndex(tau)
	ePrime := GetEpochIndex(tauPrime)
	fmt.Println("tau: ", tau)
	fmt.Println("tauPrime: ", tauPrime)
	fmt.Println("e: ", e)
	fmt.Println("ePrime: ", ePrime)
	// iota isn't change
	iota := priorState.GetIota()
	s.GetPosteriorStates().SetIota(iota)
	s.GetPosteriorStates().SetTau(tauPrime)
	if ePrime > e {
		// Update state to posterior state
		s.GetPosteriorStates().SetGammaK(ReplaceOffenderKeys(iota))
		s.GetPosteriorStates().SetKappa(priorState.GetGammaK())
		s.GetPosteriorStates().SetLambda(priorState.GetKappa())
		z, zErr := UpdateBandersnatchKeyRoot(s.GetPosteriorStates().GetGammaK())
		if zErr != nil {
			fmt.Printf("Error updating Bandersnatch key root: %v\n", zErr)
			return
		}
		s.GetPosteriorStates().SetGammaZ(z)
	} else {
		s.GetPosteriorStates().SetGammaK(priorState.GetGammaK())
		s.GetPosteriorStates().SetKappa(priorState.GetKappa())
		s.GetPosteriorStates().SetLambda(priorState.GetLambda())
		s.GetPosteriorStates().SetGammaZ(priorState.GetGammaZ())
	}
}
