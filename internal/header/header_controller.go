package header

import (
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

// HeaderController is a controller for the header.
// This controller is used to manage the header.
// You can use this controller to create a header.
type HeaderController struct {
	Store *blockchain.ChainState
}

// NewHeaderController creates a new HeaderController.
func NewHeaderController() *HeaderController {
	return &HeaderController{
		Store: blockchain.GetInstance(),
	}
}

// Set sets the Header to the given Header.
// You can load the test data and generate a header from this function.
func (h *HeaderController) SetHeader(header types.Header) {
	h.Store.GetProcessingBlockPointer().SetHeader(header)
}

// Get returns the Header.
func (h *HeaderController) GetHeader() types.Header {
	return h.Store.GetProcessingBlockPointer().GetHeader()
}

// CreateParentHeaderHash creates the parent header hash of the header.
// (5.2)
// H_p: parent header hash
func (h *HeaderController) CreateParentHeaderHash(parentHeader types.Header) error {
	encoder := types.NewEncoder()
	encoded_parent_header, err := encoder.Encode(&parentHeader)
	if err != nil {
		return err
	}

	// hash function (blake2b)
	parentHeaderHash := types.HeaderHash(hash.Blake2bHash(encoded_parent_header))
	h.Store.GetProcessingBlockPointer().SetParent(parentHeaderHash)

	return nil
}

// CreateExtrinsicHash creates the extrinsic hash of the header.
// (5.4), (5.5), (5.6)
// H_x: extrinsic hash
func (h *HeaderController) CreateExtrinsicHash(extrinsic types.Extrinsic) error {
	extrinsicHash, err := utilities.CreateExtrinsicHash(extrinsic)
	if err != nil {
		return err
	}

	h.Store.GetProcessingBlockPointer().SetExtrinsicHash(extrinsicHash)

	return nil
}

func GetCurrentTimeInSecond() uint64 {
	// The Jam Common Era is 2025-01-01 12:00:00 UTC defined in the graypaper.
	now := time.Now().UTC()
	secondsSinceJam := uint64(now.Sub(types.JamCommonEra).Seconds())

	return secondsSinceJam
}

// ValidateTimeSlot validates the time slot of the header.
// 1. The time slot of the header is always larger than the parent header's
// time slot.
// 2. The time slot of the header is always smaller than the current time.
func (h *HeaderController) ValidateTimeSlot(parentHeader types.Header, timeslot types.TimeSlot) error {
	if timeslot <= parentHeader.Slot {
		return fmt.Errorf("The time slot of the header is always larger than the parent header's time slot.")
	}

	// Get the current time in seconds.
	currentTimeInSecond := GetCurrentTimeInSecond()
	timeslotInSecond := uint64(timeslot) * uint64(types.SlotPeriod)

	if timeslotInSecond > currentTimeInSecond {
		return fmt.Errorf("The time slot of the header is always smaller than the current time.")
	}

	return nil
}

// CreateHeaderSlot creates the time slot of the header.
// (5.7) H_t: time slot
// Users can use the function to set the timeslot of this header (block).
// It means the block is built in this timeslot.
func (h *HeaderController) CreateHeaderSlot(parentHeader types.Header, currentTimeslot types.TimeSlot) error {
	err := h.ValidateTimeSlot(parentHeader, currentTimeslot)
	if err != nil {
		return err
	}

	h.Store.GetProcessingBlockPointer().SetSlot(currentTimeslot)
	return nil
}

// (5.8) H_r: state root hash
func (h *HeaderController) CreateStateRootHash(parentState types.State) {
	// State merklization
	parentStateRoot := merklization.MerklizationState(parentState)
	h.Store.GetProcessingBlockPointer().SetParentStateRoot(types.StateRoot(parentStateRoot))
}

/*
func CreateStateRootHash(parentState types.State) {
	parentStateRoot := merklization.MerklizationState(parentState)
	h.Store.GetProcessingBlockPointer().SetParentStateRoot(types.StateRoot(parentStateRoot))
}
*/

// H_i: a Bandersnatch block author index
func (h *HeaderController) CreateBlockAuthorIndex(authorIndex types.ValidatorIndex) {
	h.Store.GetProcessingBlockPointer().SetAuthorIndex(authorIndex)
}

// H_a = k'[H_i]
// k': posterior current validator set
func (h *HeaderController) GetAuthorBandersnatchKey(header types.Header) types.BandersnatchPublic {
	authorIndex := header.AuthorIndex

	// Get the posterior current validator set
	s := blockchain.GetInstance()

	// Get the validator by index
	validator := s.GetPosteriorCurrentValidatorByIndex(authorIndex)

	return validator.Bandersnatch
}

// H_e: epoch
// (5.10)
func (h *HeaderController) CreateEpochMark(epochMark *types.EpochMark) {
	h.Store.GetProcessingBlockPointer().SetEpochMark(epochMark)
}

// H_w: winning tickets
// (5.10)
func (h *HeaderController) CreateWinningTickets(ticketsMark *types.TicketsMark) {
	h.Store.GetProcessingBlockPointer().SetTicketsMark(ticketsMark)
}

// H_o: offenders markers
// (5.10)
func (h *HeaderController) CreateOffendersMarkers(offendersMark types.OffendersMark) {
	h.Store.GetProcessingBlockPointer().SetOffendersMark(offendersMark)
}

// H_v: the entropy-yielding VRF signature
// EntropySource
func (h *HeaderController) CreateEntropySource(vrfSignature types.BandersnatchVrfSignature) {
	h.Store.GetProcessingBlockPointer().SetEntropySource(vrfSignature)
}

// H_s: a block seal
func (h *HeaderController) CreateBlockSeal(blockSeal types.BandersnatchVrfSignature) {
	h.Store.GetProcessingBlockPointer().SetSeal(blockSeal)
}

// GetAncestorHeaders returns all ancestor headers as Ancestry type.
// (5.3) A
func (h *HeaderController) GetAncestorHeaders() types.Ancestry {
	s := blockchain.GetInstance()
	return s.GetAncestry()
}

// AddAncestorHeader adds the header to the ancestor headers.
// It converts Header to AncestryItem internally.
func (h *HeaderController) AddAncestorHeader(header types.Header) {
	s := blockchain.GetInstance()
	s.AddAncestorHeader(header)
}
