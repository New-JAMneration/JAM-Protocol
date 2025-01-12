package safrole

import (
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestCreateEpochMarkerSameEpoch(t *testing.T) {
	s := store.GetInstance()

	now := time.Now().UTC()
	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))

	// Simulate previous time slot to trigger create epoch marker
	priorTau := tauPrime - types.TimeSlot(types.EpochLength)

	s.GetPriorStates().SetTau(priorTau)

	// Set gamma_k into posterior state
	// Load fake validators
	fakeValidators := LoadFakeValidators()

	// Create validators data
	validatorsData := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		validatorsData = append(validatorsData, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	s.GetPosteriorStates().SetGammaK(validatorsData)

	// Prepare eta_0, eta_1
	eta0 := types.Entropy(hash.Blake2bHash([]byte("eta0")))
	eta1 := types.Entropy(hash.Blake2bHash([]byte("eta1")))

	// Get Eta from prior state
	priorEta := s.GetPriorStates().GetEta()

	// Update Eta
	priorEta[0] = eta0
	priorEta[1] = eta1

	// Set Eta
	s.GetPriorStates().SetEta(priorEta)

	CreateEpochMarker()

	// Check if epoch marker is created
	epochMarker := s.GetIntermediateHeaderPointer().GetEpochMark()

	if epochMarker == nil {
		t.Errorf("Epoch marker should not be nil")
		return
	}

	// Check if epoch marker is correct
	if epochMarker.Entropy != eta0 {
		t.Errorf("Epoch marker entropy is incorrect")
	}

	if epochMarker.TicketsEntropy != eta1 {
		t.Errorf("Epoch marker tickets entropy is incorrect")
	}

	if len(epochMarker.Validators) != len(fakeValidators) {
		t.Errorf("Epoch marker validators length is incorrect")
	}

	// Compare each validator
	for i, validator := range epochMarker.Validators {
		if validator != fakeValidators[i].Bandersnatch {
			t.Errorf("Epoch marker validator is incorrect")
		}
	}
}

func TestCreateEpochMarkerNewEpoch(t *testing.T) {
	s := store.GetInstance()

	now := time.Now().UTC()
	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))

	// Set the prior tau to the current tau, so that the epoch index is the same
	priorTau := tauPrime

	s.GetPriorStates().SetTau(priorTau)

	// Set gamma_k into posterior state
	// Load fake validators
	fakeValidators := LoadFakeValidators()

	// Create validators data
	validatorsData := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		validatorsData = append(validatorsData, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	s.GetPosteriorStates().SetGammaK(validatorsData)

	CreateEpochMarker()

	// Check if epoch marker is nil
	if s.GetIntermediateHeaderPointer().GetEpochMark() != nil {
		t.Errorf("Epoch marker should be nil")
	}
}
