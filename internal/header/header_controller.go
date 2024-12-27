package header

import (
	"errors"
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

type HeaderController struct {
	Header types.Header
}

func NewHeaderController() *HeaderController {
	return &HeaderController{
		Header: types.Header{},
	}
}

// Set sets the Header to the given Header.
// You can load the test data and generate a header from this function.
func (h *HeaderController) Set(header types.Header) {
	h.Header = header
}

// Get returns the Header.
func (h *HeaderController) Get() types.Header {
	return h.Header
}

// GetParentHeader returns the parent header of the header.
// H^- = P(H)
// (5.2) (5.3)
func (h *HeaderController) GetParentHeader(header types.Header) (types.Header, error) {
	s := store.GetInstance()
	headers := s.GetAncestorHeaders()

	// Use the time slot of the header to get the parent header.
	if header.Slot == 0 {
		return types.Header{}, errors.New("This is the genesis header.")
	}

	// TODO: 僅透過 Slot 來取得 parent header 是否合理?
	parentSlot := header.Slot - 1

	return headers[parentSlot], nil
}

// GenerateParentHash generates the parent hash of the header.
// (5.2) H_p
func (h *HeaderController) GenerateParentHash(header types.Header) (types.HeaderHash, error) {
	// P function => get parent header
	parentHeader, err := h.GetParentHeader(header)
	if err != nil {
		fmt.Println(err)
		return types.HeaderHash{}, err
	}

	// serialization
	serializedHeader := utilities.HeaderSerialization(parentHeader)

	// hash function (blake2b)
	parentHeaderHash := types.HeaderHash(hash.Blake2bHash(serializedHeader))

	return parentHeaderHash, nil
}

func GetCurrentTimeInSecond() int64 {
	// The Jam Common Era is 2025-01-01 12:00:00 UTC defined in the graypaper.
	// jamCommonEra := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// FIXME: Use the real Jam Common Era
	jamCommonEra := time.Now().UTC().Add(-600 * time.Second)

	now := time.Now().UTC()
	secondsSinceJam := int64(now.Sub(jamCommonEra).Seconds())

	return secondsSinceJam
}

// (5.4), (5.5), (5.6)
// Implement the equation in Appendix C.2 Block Serialization

// (5.7)
// 1. 現在 header 的 time slot 一定在過去的時間點 (H_t * P <= T)
// 2. 現在 header 的 time slot 一定大於 parent header 的 time slot (P(H)_t <
// H_t)
func (h *HeaderController) ValidateHeaderTimeslot(header types.Header) (bool, error) {
	parentHeader, err := h.GetParentHeader(header)
	if err != nil {
		// If the header is the genesis header, continue.
		return false, err
	}

	headerTimeslotGreaterThanParent := header.Slot > parentHeader.Slot

	// TODO: Define the const in types.go or const.go
	const slotPeriod uint8 = 6
	headerTimeSlotInSecond := header.Slot * types.TimeSlot(slotPeriod)
	headerTimeslotLessThanCurrentTime := int64(headerTimeSlotInSecond) <= GetCurrentTimeInSecond()

	if headerTimeslotGreaterThanParent && headerTimeslotLessThanCurrentTime {
		return true, nil
	}
	return false, nil
}

// (5.8)
// Implement the equation in Appendix D.2 Merklization
// TODO: The merklization function is under development.

// GetAuthor returns the author of the block.
// (5.9) H_a
func (h *HeaderController) GetAuthor(authorIndex types.ValidatorIndex) types.Validator {
	s := store.GetInstance()
	return s.GetPosteriorValidatorByIndex(authorIndex)
}
