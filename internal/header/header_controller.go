package header

import (
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
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
// INFO: This is different between Appendix C (C.16) and (5.4), (5.5), (5.6).
func (h *HeaderController) CreateExtrinsicHash(extrinsic types.Extrinsic) {
	ticketSerializedHash := hash.Blake2bHash(utilities.ExtrinsicTicketSerialization(extrinsic.Tickets))
	preimageSerializedHash := hash.Blake2bHash(utilities.ExtrinsicPreimageSerialization(extrinsic.Preimages))
	AssureanceSerializedHash := hash.Blake2bHash(utilities.ExtrinsicAssuranceSerialization(extrinsic.Assurances))
	DisputeSerializedHash := hash.Blake2bHash(utilities.ExtrinsicDisputeSerialization(extrinsic.Disputes))

	// g (5.6)
	g := types.ByteSequence{}
	g = append(g, utilities.SerializeU64(types.U64(len(extrinsic.Guarantees)))...)
	for _, guarantee := range extrinsic.Guarantees {
		// w, WorkReport
		w := guarantee.Report
		wHash := hash.Blake2bHash(utilities.WorkReportSerialization(w))

		// t, Slot
		t := guarantee.Slot
		tSerialized := utilities.SerializeFixedLength(types.U32(t), 4)

		// a, Signatures (credential)
		signaturesLength, Signatures := utilities.LensElementPair(guarantee.Signatures)

		elementSerialized := types.ByteSequence{}
		elementSerialized = append(elementSerialized, utilities.SerializeByteSequence(wHash[:])...)
		elementSerialized = append(elementSerialized, utilities.SerializeByteSequence(tSerialized)...)
		elementSerialized = append(elementSerialized, utilities.SerializeU64(types.U64(signaturesLength))...)
		for _, signature := range Signatures {
			elementSerialized = append(elementSerialized, utilities.SerializeU64(types.U64(signature.ValidatorIndex))...)
			elementSerialized = append(elementSerialized, utilities.SerializeByteSequence(signature.Signature[:])...)
		}

		// If the input type of serialization is octet sequence, we can directly
		// append it because it is already serialized.
		g = append(g, elementSerialized...)
	}

	gHash := hash.Blake2bHash(g)

	// Serialize the hash of the extrinsic elements
	serializedElements := types.ByteSequence{}
	serializedElements = append(serializedElements, utilities.WrapOpaqueHash(ticketSerializedHash).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapOpaqueHash(preimageSerializedHash).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapOpaqueHash(gHash).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapOpaqueHash(AssureanceSerializedHash).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapOpaqueHash(DisputeSerializedHash).Serialize()...)

	// Hash the serialized elements
	extrinsicHash := hash.Blake2bHash(serializedElements)

	h.Header.ExtrinsicHash = extrinsicHash
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

	h.Header.Slot = currentTimeslot
	return nil
}

// (5.8) H_r: state root hash
func (h *HeaderController) CreateStateRootHash(parentState types.State) {
	// State merklization
	parentStateRoot := merklization.MerklizationState(parentState)
	h.Header.ParentStateRoot = types.StateRoot(parentStateRoot)
}

// H_i: a Bandersnatch block author index
func (h *HeaderController) CreateBlockAuthorIndex(authorIndex types.ValidatorIndex) {
	h.Header.AuthorIndex = authorIndex
}

// H_a = k'[H_i]
// k': posterior current validator set
func (h *HeaderController) GetAuthorBandersnatchKey(header types.Header) types.BandersnatchPublic {
	authorIndex := header.AuthorIndex

	// Get the posterior current validator set
	s := store.GetInstance()

	// Get the validator by index
	validator := s.GetPosteriorCurrentValidatorByIndex(authorIndex)

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
