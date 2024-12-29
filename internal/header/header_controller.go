package header

import (
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// HeaderController is a controller for the header.
// This controller is used to manage the header.
// You can use this controller to create a header.
type HeaderController struct {
	Header types.Header
}

// NewHeaderController creates a new HeaderController.
func NewHeaderController() *HeaderController {
	return &HeaderController{
		Header: types.Header{},
	}
}

// Set sets the Header to the given Header.
// You can load the test data and generate a header from this function.
func (h *HeaderController) SetHeader(header types.Header) {
	h.Header = header
}

// Get returns the Header.
func (h *HeaderController) GetHeader() types.Header {
	return h.Header
}

// CreateParentHeaderHash creates the parent header hash of the header.
// (5.2)
// H_p: parent header hash
func (h *HeaderController) CreateParentHeaderHash(parentHeader types.Header) {
	// serialization
	serializedHeader := utilities.HeaderSerialization(parentHeader)

	// hash function (blake2b)
	parentHeaderHash := types.HeaderHash(hash.Blake2bHash(serializedHeader))

	h.Header.Parent = parentHeaderHash
}

// CreateExtrinsicHash creates the extrinsic hash of the header.
// (5.4), (5.5), (5.6)
// H_x: extrinsic hash
func (h *HeaderController) CreateExtrinsicHash(extrinsic types.Extrinsic) {
	ticketSerialized := hash.Blake2bHash(utilities.ExtrinsicTicketSerialization(extrinsic.Tickets))
	preimageSerialized := hash.Blake2bHash(utilities.ExtrinsicPreimageSerialization(extrinsic.Preimages))
	guaranteeSerialized := hash.Blake2bHash(utilities.ExtrinsicGuaranteeSerialization(extrinsic.Guarantees))
	AssureanceSerialized := hash.Blake2bHash(utilities.ExtrinsicAssuranceSerialization(extrinsic.Assurances))
	DisputeSerialized := hash.Blake2bHash(utilities.ExtrinsicDisputeSerialization(extrinsic.Disputes))

	// Serialize the hash of the extrinsic elements
	serializedElements := types.ByteSequence{}
	serializedElements = append(serializedElements, utilities.WrapByteArray32(types.ByteArray32(ticketSerialized)).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapByteArray32(types.ByteArray32(preimageSerialized)).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapByteArray32(types.ByteArray32(guaranteeSerialized)).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapByteArray32(types.ByteArray32(AssureanceSerialized)).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapByteArray32(types.ByteArray32(DisputeSerialized)).Serialize()...)

	// Hash the serialized elements
	extrinsicHash := types.OpaqueHash(hash.Blake2bHash(serializedElements))

	h.Header.ExtrinsicHash = extrinsicHash
}

// FIXME: Use the real Jam Common Era
func getCurrentTimeInSecond() uint64 {
	// The Jam Common Era is 2025-01-01 12:00:00 UTC defined in the graypaper.
	// jamCommonEra := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	jamCommonEra := time.Now().UTC().Add(-600 * time.Second)

	now := time.Now().UTC()
	secondsSinceJam := uint64(now.Sub(jamCommonEra).Seconds())

	return secondsSinceJam
}

func (h *HeaderController) ValidateTimeSlot(parentHeader types.Header, timeslot types.TimeSlot) error {
	// INFO: Validate the time slot
	// 1. The time slot of the header is always larger than the parent header's
	// time slot.
	// 2. The time slot of the header is always smaller than the current time.

	if timeslot <= parentHeader.Slot {
		return fmt.Errorf("The time slot of the header is always larger than the parent header's time slot.")
	}

	// Get the current time in seconds.
	currentTimeInSecond := getCurrentTimeInSecond()
	const slotPeriod uint8 = 6
	timeslotInSecond := uint64(timeslot) * uint64(slotPeriod)

	if timeslotInSecond > currentTimeInSecond {
		return fmt.Errorf("The time slot of the header is always smaller than the current time.")
	}

	return nil
}

// CreateHeaderSlot creates a time slot for the header.
// (5.7) H_t: time slot
func (h *HeaderController) CreateHeaderSlot(parentHeader types.Header, timeslot types.TimeSlot) error {
	err := h.ValidateTimeSlot(parentHeader, timeslot)
	if err != nil {
		return err
	}

	h.Header.Slot = timeslot
	return nil
}

// (5.8) H_r: state root hash
func (h *HeaderController) CreateStateRootHash(state types.State) {
	// State merklization
	// FIXME: call function to merklize the state
	// We're waiting for the merklization function merge to the main branch.
}

// H_i: a Bandersnatch block author index
func (h *HeaderController) CreateBlockAuthorIndex(authorIndex types.ValidatorIndex) {
	h.Header.AuthorIndex = authorIndex
}

// H_a = k'[H_i]
// k': posterior current validator set
func (h *HeaderController) GetAutherBendersnatchKey(header types.Header) types.BandersnatchPublic {
	authorIndex := header.AuthorIndex

	// Get the posterior current validator set
	s := store.GetInstance()

	// Get the validator by index
	validator := s.GetPosteriorValidatorByIndex(authorIndex)

	return validator.Bandersnatch
}

// H_e: epoch
// (5.10)
func (h *HeaderController) CreateEpochMark(epochMark *types.EpochMark) {
	h.Header.EpochMark = epochMark
}

// H_w: winning tickets
// (5.10)
func (h *HeaderController) CreateWinningTickets(ticketsMark *types.TicketsMark) {
	h.Header.TicketsMark = ticketsMark
}

// H_o: offenders markers
// (5.10)
func (h *HeaderController) CreateOffendersMarkers(offendersMark types.OffendersMark) {
	h.Header.OffendersMark = offendersMark
}

// H_v: the entropy-yielding VRF signature
// EntropySource
func (h *HeaderController) CreateEntropySource(vrfSignature types.BandersnatchVrfSignature) {
	h.Header.EntropySource = vrfSignature
}

// H_s: a block seal
func (h *HeaderController) CreateBlockSeal(blockSeal types.BandersnatchVrfSignature) {
	h.Header.Seal = blockSeal
}

// GetParentHeader returns all ancestor headers.
// (5.3) A
func (h *HeaderController) GetAncestorHeaders() []types.Header {
	s := store.GetInstance()
	return s.GetAncestorHeaders()
}

// AddAncestorHeader adds the header to the ancestor headers.
func (h *HeaderController) AddAncestorHeader(header types.Header) {
	s := store.GetInstance()
	s.AddAncestorHeader(header)
}
