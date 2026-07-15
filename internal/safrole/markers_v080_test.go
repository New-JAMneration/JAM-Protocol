package safrole

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
)

func TestValidateHeaderEpochMarkPresenceAndContent(t *testing.T) {
	validators := make(types.ValidatorsData, types.ValidatorsCount)
	keys := make([]types.EpochMarkValidatorKeys, len(validators))
	for i := range validators {
		validators[i].Bandersnatch[0] = byte(i + 1)
		validators[i].Ed25519[0] = byte(i + 1)
		keys[i] = types.EpochMarkValidatorKeys{
			Bandersnatch: validators[i].Bandersnatch,
			Ed25519:      validators[i].Ed25519,
		}
	}
	prior := types.State{Tau: types.TimeSlot(types.EpochLength - 1)}
	prior.Eta[0] = types.Entropy{0x11}
	prior.Eta[1] = types.Entropy{0x22}
	posterior := types.State{}
	posterior.Gamma.GammaK = validators
	validMark := &types.EpochMark{
		Entropy:        prior.Eta[0],
		TicketsEntropy: prior.Eta[1],
		Validators:     keys,
	}

	tests := []struct {
		name    string
		header  types.Header
		wantErr bool
	}{
		{
			name:   "transition requires matching mark",
			header: types.Header{Slot: types.TimeSlot(types.EpochLength), EpochMark: validMark},
		},
		{
			name:    "transition rejects missing mark",
			header:  types.Header{Slot: types.TimeSlot(types.EpochLength)},
			wantErr: true,
		},
		{
			name: "same epoch rejects unexpected mark",
			header: types.Header{
				Slot:      types.TimeSlot(types.EpochLength - 1),
				EpochMark: validMark,
			},
			wantErr: true,
		},
		{
			name:   "same epoch accepts no mark",
			header: types.Header{Slot: types.TimeSlot(types.EpochLength - 1)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateHeaderEpochMark(tc.header, &prior, &posterior)
			if tc.wantErr {
				if err == nil || *err != SafroleErrorCode.InvalidEpochMark {
					t.Fatalf("error = %v, want InvalidEpochMark", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateHeaderTicketsMarkPresence(t *testing.T) {
	state := types.State{Tau: types.TimeSlot(types.SlotSubmissionEnd - 1)}
	state.Gamma.GammaA = make(types.TicketsAccumulator, types.EpochLength)
	for i := range state.Gamma.GammaA {
		state.Gamma.GammaA[i].ID[0] = byte(i + 1)
	}
	expected := types.TicketsMark(OutsideInSequencer(&state.Gamma.GammaA))

	tests := []struct {
		name    string
		header  types.Header
		wantErr bool
	}{
		{
			name: "boundary accepts expected mark",
			header: types.Header{
				Slot:        types.TimeSlot(types.SlotSubmissionEnd),
				TicketsMark: &expected,
			},
		},
		{
			name:    "boundary rejects missing mark",
			header:  types.Header{Slot: types.TimeSlot(types.SlotSubmissionEnd)},
			wantErr: true,
		},
		{
			name: "outside boundary rejects unexpected mark",
			header: types.Header{
				Slot:        types.TimeSlot(types.SlotSubmissionEnd - 1),
				TicketsMark: &expected,
			},
			wantErr: true,
		},
		{
			name:   "outside boundary accepts no mark",
			header: types.Header{Slot: types.TimeSlot(types.SlotSubmissionEnd - 1)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateHeaderTicketsMark(tc.header, &state)
			if tc.wantErr {
				if err == nil || *err != SafroleErrorCode.InvalidTicketsMark {
					t.Fatalf("error = %v, want InvalidTicketsMark", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
