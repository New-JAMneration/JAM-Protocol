package safrole

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

func measureTimeSafrole(operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	if err != nil {
		log.Printf("⏱️  %-35s took: %10v (ERROR: %v)", operation, duration, err)
	} else {
		log.Printf("⏱️  %-35s took: %10v", operation, duration)
	}
	return err
}

// measureTimeNoErrSafrole measures execution time of a function without error return
func measureTimeNoErrSafrole(operation string, fn func()) {
	start := time.Now()
	fn()
	duration := time.Since(start)
	log.Printf("⏱️  %-35s took: %10v", operation, duration)
}

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
		// z, zErr := UpdateBandersnatchKeyRoot(s.GetPosteriorStates().GetGammaK())
		// if zErr != nil {
		// 	return fmt.Errorf("error updating Bandersnatch key root: %v", zErr)
		// }
		// s.GetPosteriorStates().SetGammaZ(z)
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
	totalStart := time.Now()
	defer func() {
		fmt.Printf("\n⏱️  %-40s Total Safrole took: %12v\n", "OuterUsedSafrole(TOTAL)", time.Since(totalStart))
	}()

	// --- STEP 1 Get Epoch and Slot for safrole --- //
	var (
		err               error
		s                 = store.GetInstance()
		header            = s.GetLatestBlock().Header
		headerAuthorIndex = header.AuthorIndex
		tau               = s.GetPriorStates().GetTau()
		tauPrime          = s.GetPosteriorStates().GetTau()
		e, m              = R(tau)
		ePrime, _         = R(tauPrime)
	)

	// prior time slot must be less than posterior time slot
	if tau >= tauPrime {
		errCode := SafroleErrorCode.BadSlot
		return &errCode
	}
	// --- STEP 2 Update Entropy123 --- //
	// (GP 6.23)
	measureTimeNoErrSafrole("UpdateEntropy", func() {
		UpdateEntropy(e, ePrime)
	})
	// --- STEP 3 safrole.go (GP 6.2, 6.13, 6.14) --- //
	// (6.2, 6.13, 6.14)
	// This function will update GammaK, GammaZ, Lambda, Kappa
	err = measureTimeSafrole("KeyRotate", func() error {
		return KeyRotate(e, ePrime)
	})
	if err != nil {
		log.Println("keyRotateErr:", err)
	}

	// After KeyRotate, gammaK and kappa are updated
	var (
		postGammaK = s.GetPosteriorStates().GetGammaK()
		postKappa  = s.GetPosteriorStates().GetKappa()
	)

	// Create ring verifier (gammaK)
	ring := []byte{}
	for _, validator := range postGammaK {
		ring = append(ring, []byte(validator.Bandersnatch[:])...)
	}
	var ringVerifier *vrf.Verifier
	err = measureTimeSafrole("NewVerifier(GammaK ring)", func() error {
		var e error
		ringVerifier, e = vrf.NewVerifier(ring, uint(len(postGammaK)))
		return e
	})
	if err != nil {
		log.Println("Failed to create verifier:", err)
		return nil
	}
	defer ringVerifier.Free()

	// Update GammaZ commitment (gammaZ)
	if ePrime > e {
		var commitment []byte
		err = measureTimeSafrole("GetCommitment(GammaZ)", func() error {
			var e error
			commitment, e = ringVerifier.GetCommitment()
			return e
		})
		if err != nil {
			log.Println("Failed to get commitment:", err)
		} else {
			s.GetPosteriorStates().SetGammaZ(types.BandersnatchRingCommitment(commitment))
		}
	}

	// IETF verifier for author Kappa[authorIndex]
	var verifier *vrf.Verifier
	err = measureTimeSafrole("NewVerifier(author key)", func() error {
		var e error
		verifier, e = vrf.NewVerifier(postKappa[headerAuthorIndex].Bandersnatch[:], 0)
		return e
	})
	if err != nil {
		log.Println("Failed to create verifier:", err)
		return nil
	}
	if verifier == nil {
		log.Println("Failed to create verifier for author key")
		return nil
	}
	defer verifier.Free()

	// (GP 6.17) // This will be used to write H_v to new header
	// UpdateHeaderEntropy()

	// --- slot_key_sequence.go (GP 6.25, 6.26) --- //
	measureTimeNoErrSafrole("UpdateSlotKeySequence", func() {
		UpdateSlotKeySequence(e, ePrime, m)
	})

	// (GP 6.22)
	err = measureTimeSafrole("UpdateEtaPrime0", func() error {
		return UpdateEtaPrime0(verifier)
	})
	if err != nil {
		log.Println("UpdateEtaPrime0Err:", err)
	}

	postState := s.GetPosteriorStates().GetState()
	HSErrCode := func() *types.ErrorCode {
		var ec *types.ErrorCode
		_ = measureTimeSafrole("ValidateHeaderSeal", func() error {
			ec = ValidateHeaderSeal(verifier, header, &postState)
			if ec != nil {
				// use non-nil error to mark failure in timing helper
				return fmt.Errorf("ValidateHeaderSeal error code: %v", *ec)
			}
			return nil
		})
		return ec
	}()
	if HSErrCode != nil {
		return HSErrCode
	}

	HEErrCode := func() *types.ErrorCode {
		var ec *types.ErrorCode
		_ = measureTimeSafrole("ValidateHeaderEntropy", func() error {
			ec = ValidateHeaderEntropy(verifier, header)
			if ec != nil {
				// use non-nil error to mark failure in timing helper
				return fmt.Errorf("ValidateHeaderEntropy error code: %v", *ec)
			}
			return nil
		})
		return ec
	}()
	if HEErrCode != nil {
		return HEErrCode
	}

	// --- STEP 4 Check TicketExtrinsic --- //
	// --- extrinsic_tickets.go (GP 6.30~6.34) --- //
	EtErrCode := func() *types.ErrorCode {
		var ec *types.ErrorCode
		_ = measureTimeSafrole("CreateNewTicketAccumulator", func() error {
			ec = CreateNewTicketAccumulator(ringVerifier)
			if ec != nil {
				// use non-nil error to mark failure in timing helper
				return fmt.Errorf("CreateNewTicketAccumulator error code: %v", *ec)
			}
			return nil
		})
		return ec
	}()
	if EtErrCode != nil {
		return EtErrCode
	}
	// (GP 6.28)
	// CreateWinningTickets(e, ePrime, m, mPrime)

	// --- sealing.go (GP 6.15~6.24) --- //
	// err = SealingHeader()
	// if err != nil {
	// 	log.Println("SealingHeaderErr:", err)
	// }

	// --- markers.go (GP 6.27, 6.28) --- //
	// (GP 6.27)
	// CreateEpochMarker(e, ePrime)

	return nil
}
