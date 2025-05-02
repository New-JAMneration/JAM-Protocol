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
func (h *HeaderController) CreateParentHeaderHash(parentHeader types.Header) error {
	encoder := types.NewEncoder()
	encoded_parent_header, err := encoder.Encode(&parentHeader)
	if err != nil {
		return err
	}

	// hash function (blake2b)
	parentHeaderHash := types.HeaderHash(hash.Blake2bHash(encoded_parent_header))

	// TODO: We have to save the parent hash in the store.
	h.Header.Parent = parentHeaderHash

	return nil
}

// CreateExtrinsicHash creates the extrinsic hash of the header.
// (5.4), (5.5), (5.6)
// H_x: extrinsic hash
<<<<<<< HEAD
func (h *HeaderController) CreateExtrinsicHash(extrinsic types.Extrinsic) error {
	// Encode the extrinsic elements
	encodedTicketsExtrinsic, err := utilities.EncodeExtrinsicTickets(extrinsic.Tickets)
=======
// INFO: This is different between Appendix C (C.16) and (5.4), (5.5), (5.6).
func (h *HeaderController) CreateExtrinsicHash(extrinsic types.Extrinsic) error {
	ticketSerializedHash := hash.Blake2bHash(utilities.ExtrinsicTicketSerialization(extrinsic.Tickets))
	preimageSerializedHash := hash.Blake2bHash(utilities.ExtrinsicPreimageSerialization(extrinsic.Preimages))
	AssureanceSerializedHash := hash.Blake2bHash(utilities.ExtrinsicAssuranceSerialization(extrinsic.Assurances))
	DisputeSerializedHash := hash.Blake2bHash(utilities.ExtrinsicDisputeSerialization(extrinsic.Disputes))

	// g (5.6)
	g := types.ByteSequence{}

	// Encode the length of the guarantees
	guaranteesLength := uint64(len(extrinsic.Guarantees))
	encoder := types.NewEncoder()

	encoded, err := encoder.EncodeUint(guaranteesLength)
>>>>>>> 198805e (refactor guarantee serialization using encode)
	if err != nil {
		return err
	}

<<<<<<< HEAD
	encodedPreimagesExtrinsic, err := utilities.EncodeExtrinsicPreimages(extrinsic.Preimages)
	if err != nil {
		return err
	}

	encodedGuaranteesExtrinsic, err := utilities.EncodeExtrinsicGuarantees(extrinsic.Guarantees)
	if err != nil {
		return err
	}
=======
	g = append(g, encoded...)

	for _, guarantee := range extrinsic.Guarantees {
		// encode the w
		encoded, err := encoder.Encode(&guarantee.Report)
		if err != nil {
			return err
		}

		// hash the encoded data
		hash := hash.Blake2bHash(types.ByteSequence(encoded))

		// append the encoded data to the output
		g = append(g, hash[:]...)

		// encode the t
		encoded, err = encoder.Encode(&(guarantee.Slot))
		if err != nil {
			return err
		}

		g = append(g, encoded...)

		// encode the length of the guarantee.a
		signatureLength := uint64(len(guarantee.Signatures))
		encoded, err = encoder.EncodeUint(signatureLength)
		if err != nil {
			return err
		}
		g = append(g, encoded...)

		// encode the guarantee.a
		for _, signature := range guarantee.Signatures {
			encoded, err = encoder.Encode(&signature)
			if err != nil {
				return err
			}
			g = append(g, encoded...)
		}
	}

	// Serialize the hash of the extrinsic elements
	serializedElements := types.ByteSequence{}
	serializedElements = append(serializedElements, utilities.WrapOpaqueHash(ticketSerializedHash).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapOpaqueHash(preimageSerializedHash).Serialize()...)
	serializedElements = append(serializedElements, g...)
	serializedElements = append(serializedElements, utilities.WrapOpaqueHash(AssureanceSerializedHash).Serialize()...)
	serializedElements = append(serializedElements, utilities.WrapOpaqueHash(DisputeSerializedHash).Serialize()...)
>>>>>>> 198805e (refactor guarantee serialization using encode)

	encodedAssurancesExtrinsic, err := utilities.EncodeExtrinsicAssurances(extrinsic.Assurances)
	if err != nil {
		return err
	}

	encodedDisputesExtrinsic, err := utilities.EncodeExtrinsicDisputes(extrinsic.Disputes)
	if err != nil {
		return err
	}

	// Hash encoded elements
	encodedTicketsHash := hash.Blake2bHash(encodedTicketsExtrinsic)
	encodedPreimagesHash := hash.Blake2bHash(encodedPreimagesExtrinsic)
	encodedGuaranteesHash := hash.Blake2bHash(encodedGuaranteesExtrinsic)
	encodedAssurancesHash := hash.Blake2bHash(encodedAssurancesExtrinsic)
	encodedDisputesHash := hash.Blake2bHash(encodedDisputesExtrinsic)

	// Concatenate the encoded elements
	encodedHash := types.ByteSequence{}
	encodedHash = append(encodedHash, types.ByteSequence(encodedTicketsHash[:])...)
	encodedHash = append(encodedHash, types.ByteSequence(encodedPreimagesHash[:])...)
	encodedHash = append(encodedHash, types.ByteSequence(encodedGuaranteesHash[:])...)
	encodedHash = append(encodedHash, types.ByteSequence(encodedAssurancesHash[:])...)
	encodedHash = append(encodedHash, types.ByteSequence(encodedDisputesHash[:])...)

	// Hash the encoded elements
	extrinsicHash := hash.Blake2bHash(encodedHash)

	// TODO: We have to save the extrinsic hash in the store.
	h.Header.ExtrinsicHash = extrinsicHash

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
