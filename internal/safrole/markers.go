// 6.6 The Markers (graypaper 0.5.4)
package safrole

import (
	"bytes"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/google/go-cmp/cmp"
)

// CreateEpochMarker creates the epoch marker
// (6.27)
func CreateEpochMarker(e types.TimeSlot, ePrime types.TimeSlot) {
	s := store.GetInstance()

	if ePrime > e {
		// New epoch, create epoch marker
		// Get eta_0, eta_1
		eta := s.GetPriorStates().GetEta()

		// Get gamma_k from posterior state
		gammaK := s.GetPosteriorStates().GetGammaK()

		// Get ed25519/bandersnatch keys from gamma_k
		var epochMarkValidatorKeys []types.EpochMarkValidatorKeys
		for _, validator := range gammaK {
			epochMarkValidatorKeys = append(epochMarkValidatorKeys, types.EpochMarkValidatorKeys{
				Bandersnatch: validator.Bandersnatch,
				Ed25519:      validator.Ed25519,
			})
		}

		epochMarker := &types.EpochMark{
			Entropy:        eta[0],
			TicketsEntropy: eta[1],
			Validators:     epochMarkValidatorKeys,
		}

		s.GetProcessingBlockPointer().SetEpochMark(epochMarker)
	} else {
		// The epoch is the same
		var epochMarker *types.EpochMark = nil
		s.GetProcessingBlockPointer().SetEpochMark(epochMarker)
	}
}

// CreateWinningTickets creates the winning tickets
// (6.28)
func CreateWinningTickets(e types.TimeSlot, ePrime types.TimeSlot, m types.TimeSlot, mPrime types.TimeSlot) {
	s := store.GetInstance()

	gammaA := s.GetPriorStates().GetGammaA()

	condition1 := ePrime == e
	condition2 := m < types.TimeSlot(types.SlotSubmissionEnd) && mPrime >= types.TimeSlot(types.SlotSubmissionEnd)
	condition3 := len(gammaA) == types.EpochLength

	if condition1 && condition2 && condition3 {
		// Z(gamma_a)
		ticketsMark := types.TicketsMark(OutsideInSequencer(&gammaA))
		s.GetProcessingBlockPointer().SetTicketsMark(&ticketsMark)
	} else {
		// The epoch is the same
		var ticketsMark *types.TicketsMark = nil
		s.GetProcessingBlockPointer().SetTicketsMark(ticketsMark)
	}
}

// ValidateHeaderEpochMark validates the epoch mark in the header
// Returns an error code if validation fails, nil otherwise
func ValidateHeaderEpochMark(header types.Header, priorState *types.State, posteriorState *types.State) *types.ErrorCode {
	e, _ := R(priorState.Tau)
	ePrime, _ := R(header.Slot)

	em := header.EpochMark
	shouldHave := ePrime > e

	// Early return: if epoch hasn't changed, epoch mark must be nil
	if !shouldHave {
		if em != nil {
			errCode := SafroleErrorCode.InvalidEpochMark
			return &errCode
		}
		return nil
	}

	// Early return: if epoch changed, epoch mark must be present
	if em == nil {
		errCode := SafroleErrorCode.InvalidEpochMark
		return &errCode
	}

	// Validate entropy fields
	if em.Entropy != priorState.Eta[0] || em.TicketsEntropy != priorState.Eta[1] {
		errCode := SafroleErrorCode.InvalidEpochMark
		return &errCode
	}

	// Validate validators count
	if types.ValidatorsCount > 0 && len(em.Validators) != types.ValidatorsCount {
		errCode := SafroleErrorCode.InvalidEpochMark
		return &errCode
	}

	// Validate validators match postState.Gamma.GammaK
	// Check bounds before accessing postState.Gamma.GammaK
	if len(posteriorState.Gamma.GammaK) != len(em.Validators) {
		errCode := SafroleErrorCode.InvalidEpochMark
		return &errCode
	}

	for i := range em.Validators {
		expectedValidator := posteriorState.Gamma.GammaK[i]
		actualValidator := em.Validators[i]

		// Compare both keys in a single check for better readability
		if !bytes.Equal(actualValidator.Bandersnatch[:], expectedValidator.Bandersnatch[:]) ||
			!bytes.Equal(actualValidator.Ed25519[:], expectedValidator.Ed25519[:]) {
			logger.Debugf("ValidateHeaderEpochMark, validator %d mismatch: %v", i, cmp.Diff(expectedValidator, actualValidator))
			errCode := SafroleErrorCode.InvalidEpochMark
			return &errCode
		}
	}

	return nil
}

func ValidateHeaderTicketsMark(header types.Header, state *types.State) *types.ErrorCode {
	tau := state.Tau
	e, m := R(tau)

	tauPrime := header.Slot
	ePrime, mPrime := R(tauPrime)

	tm := header.TicketsMark
	gammaA := state.Gamma.GammaA

	condition1 := ePrime == e
	condition2 := m < types.TimeSlot(types.SlotSubmissionEnd) && mPrime >= types.TimeSlot(types.SlotSubmissionEnd)
	condition3 := len(gammaA) == types.EpochLength
	shouldHave := condition1 && condition2 && condition3

	if !shouldHave {
		if tm != nil {
			errCode := SafroleErrorCode.InvalidTicketsMark
			return &errCode
		}
		return nil
	}

	if tm == nil {
		errCode := SafroleErrorCode.InvalidTicketsMark
		return &errCode
	}

	expectedTm := OutsideInSequencer(&gammaA)

	if len(*tm) != len(expectedTm) {
		errCode := SafroleErrorCode.InvalidTicketsMark
		return &errCode
	}

	for i := range *tm {
		if (*tm)[i] != expectedTm[i] {
			errCode := SafroleErrorCode.InvalidTicketsMark
			return &errCode
		}
	}

	return nil
}

func ValidateHeaderOffenderMarker(header types.Header, state *types.State) *types.ErrorCode {
	block := store.GetInstance().GetLatestBlock()
	if block.Header.Slot != header.Slot {
		// Not the latest block, skip validation
		return nil
	}
	disputes := block.Extrinsic.Disputes
	expected := make([]types.Ed25519Public, 0, len(disputes.Culprits)+len(disputes.Faults))

	for _, c := range disputes.Culprits {
		expected = append(expected, c.Key)
	}
	for _, f := range disputes.Faults {
		expected = append(expected, f.Key)
	}
	var got []types.Ed25519Public
	if header.OffendersMark != nil {
		got = header.OffendersMark
	}

	if len(expected) == 0 {
		if header.OffendersMark != nil && len(got) > 0 {
			err := SafroleErrorCode.InvalidOffenderMarker
			return &err
		}
		return nil
	}
	if header.OffendersMark == nil {
		err := SafroleErrorCode.InvalidOffenderMarker
		return &err
	}
	if len(got) != len(expected) {
		err := SafroleErrorCode.InvalidOffenderMarker
		return &err
	}
	for i := range expected {
		if !bytes.Equal(got[i][:], expected[i][:]) {
			err := SafroleErrorCode.InvalidOffenderMarker
			return &err
		}
	}

	return nil
}
