package safrole

import (
	"fmt"
	"log"
	"slices"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
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
// KeyRotate rotates the keys
// Update the state with the new Safrole state
// (6.13)
func KeyRotate(e types.TimeSlot, ePrime types.TimeSlot) error {
	s := store.GetInstance()

	// Get prior state
	priorState := s.GetPriorStates()

	if ePrime > e {
		// Update state to posterior state
		s.GetPosteriorStates().SetGammaK(ReplaceOffenderKeys(priorState.GetIota()))
		s.GetPosteriorStates().SetKappa(priorState.GetGammaK())
		s.GetPosteriorStates().SetLambda(priorState.GetKappa())
		z, zErr := UpdateBandersnatchKeyRoot(s.GetPosteriorStates().GetGammaK())
		if zErr != nil {
			return fmt.Errorf("error updating Bandersnatch key root: %v", zErr)
		}
		s.GetPosteriorStates().SetGammaZ(z)
	} else {
		s.GetPosteriorStates().SetGammaK(priorState.GetGammaK())
		s.GetPosteriorStates().SetKappa(priorState.GetKappa())
		s.GetPosteriorStates().SetLambda(priorState.GetLambda())
		s.GetPosteriorStates().SetGammaZ(priorState.GetGammaZ())
	}
	return nil
}

// Outer Safrole function
// I made this function return ErrorCode only
func OuterUsedSafrole() *types.ErrorCode {
	s := store.GetInstance()

	// --- STEP 1 Get Epoch and Slot for safrole --- //
	var (
		tau            = s.GetPriorStates().GetTau()
		tauPrime       = s.GetPosteriorStates().GetTau()
		e, m           = R(tau)
		ePrime, mPrime = R(tauPrime)
	)
	log.Println(e, ePrime, m, mPrime)
	// prior time slot must be less than posterior time slot
	if tau >= tauPrime {
		errCode := SafroleErrorCode.BadSlot
		return &errCode
	}
	// --- STEP 2 Update Entropy123 --- //
	// (GP 6.23)
	UpdateEntropy(e, ePrime)

	// --- STEP 3 safrole.go (GP 6.2, 6.13, 6.14) --- //
	// (6.2, 6.13, 6.14)
	// This function will update GammaK, GammaZ, Lambda, Kappa
	err := KeyRotate(e, ePrime)
	if err != nil {
		log.Println("keyRotateErr:", err)
	}

	// (GP 6.22)
	err = UpdateEtaPrime0()
	if err != nil {
		log.Println("UpdateEtaPrime0Err:", err)
	}

	// (GP 6.17) // This will be used to write H_v to new header
	// UpdateHeaderEntropy()

	// --- slot_key_sequence.go (GP 6.25, 6.26) --- //
	UpdateSlotKeySequence(e, ePrime, m)

	// --- STEP 4 Check TicketExtrinsic --- //
	// --- extrinsic_tickets.go (GP 6.30~6.34) --- //
	EtErrCode := CreateNewTicketAccumulator()
	if EtErrCode != nil {
		return EtErrCode
	}
	// (GP 6.28)
	CreateWinningTickets(e, ePrime, m, mPrime)

	// --- sealing.go (GP 6.15~6.24) --- //
	err = SealingHeader()
	if err != nil {
		log.Println("SealingHeaderErr:", err)
	}

	// --- markers.go (GP 6.27, 6.28) --- //
	// (GP 6.27)
	CreateEpochMarker(e, ePrime)

	return nil
}
