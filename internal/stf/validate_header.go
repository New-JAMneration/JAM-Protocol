package stf

import (
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

// TODO: Align the official errorCode
func ValidateHeader(header types.Header, state *types.State) error {
	if err := safrole.ValidateHeaderSeal(header, state); err != nil {
		return err
	}

	if err := safrole.ValidateHeaderEntropy(header, state); err != nil {
		return err
	}

	if err := safrole.ValidateHeaderEpochMark(header, state); err != nil {
		return err
	}

	if err := safrole.ValidateHeaderTicketsMark(header, state); err != nil {
		return err
	}

	if err := safrole.ValidateHeaderOffenderMarker(header, state); err != nil {
		return err
	}

	if err := validateHeaderSlot(header); err != nil {
		return err
	}

	if err := validateStateRootHash(header, state); err != nil {
		return err
	}

	return nil
}

func validateHeaderSlot(header types.Header) error {
	// Get the parent header from the store
	s := store.GetInstance()

	parentHeaderHash := header.Parent
	parentBlock, err := s.GetBlockByHash(parentHeaderHash)
	if err != nil {
		return err
	}

	parentHeader := parentBlock.Header

	if header.Slot <= parentHeader.Slot {
		errCode := SafroleErrorCode.BadSlot
		return &errCode
	}

	// Get the current time in seconds.
	currentTimeInSecond := getCurrentTimeInSecond()
	timeslotInSecond := uint64(header.Slot) * uint64(types.SlotPeriod)

	if timeslotInSecond > currentTimeInSecond {
		errCode := SafroleErrorCode.BadSlot
		return &errCode
	}

	return nil
}

func getCurrentTimeInSecond() uint64 {
	// The Jam Common Era is 2025-01-01 12:00:00 UTC defined in the graypaper.
	now := time.Now().UTC()
	secondsSinceJam := uint64(now.Sub(types.JamCommonEra).Seconds())

	return secondsSinceJam
}

func validateStateRootHash(header types.Header, state *types.State) error {
	// State merklization
	parentStateRoot := merklization.MerklizationState(*state)
	if types.StateRoot(parentStateRoot) != header.ParentStateRoot {
		error := fmt.Errorf("invalid state root hash")
		return error
	}
	return nil
}
