package header

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// FIXME: We have to manage the global variable in a better way.
type Blocks []jamTypes.Block

var blocks = Blocks{
	// Genesis block
	{
		Header:    jamTypes.Header{},
		Extrinsic: jamTypes.Extrinsic{},
	},
}

type HeaderController struct {
	Header jamTypes.Header
}

func NewHeaderController() *HeaderController {
	return &HeaderController{
		Header: jamTypes.Header{},
	}
}

// Set sets the Header to the given Header.
// You can load the test data and generate a header from this function.
func (h *HeaderController) Set(header jamTypes.Header) {
	h.Header = header
}

// Get returns the Header.
func (h *HeaderController) Get() jamTypes.Header {
	return h.Header
}

// GetParentHeader returns the parent header of the header.
// H^- = P(H)
func (h *HeaderController) GetParentHeader(header jamTypes.Header) jamTypes.Header {
	// Use the time slot of the header to get the parent header.
	slotIndex := header.Slot - 1
	if slotIndex < 0 {
		// throw error (The input is the genesis block.)
		err := "The input is the genesis block."
		panic(err)
	}

	return blocks[slotIndex].Header
}

// GenerateParentHash generates the parent hash of the header.
// (5.2) H_p
func (h *HeaderController) GenerateParentHash(header jamTypes.Header) (parentHeaderHash jamTypes.HeaderHash) {
	// P function => get parent header
	// parentHeader := h.GetParentHeader(header)

	// serialization
	// FIXME: 呼叫 (C.19)

	// hash function (blake2b)

	return parentHeaderHash
}

// FIXME: We have to manage the global variable in a better way.
// We only require implmentations to store headers of ancestors which were
// authored in the previous L = 24 hours of any block B they wish to validate.
var ancestorHeaders = []jamTypes.Header{}

func (h *HeaderController) AppendAncestorHeader(header jamTypes.Header) {
	ancestorHeaders = append(ancestorHeaders, header)
}

// FIXME: We have to manage the global variable in a better way.
var posteriorCurrentValidators = jamTypes.ValidatorsData{}

// GetAuthor returns the author of the block.
// (5.9) H_a
func (h *HeaderController) GetAuthor(authorIndex jamTypes.ValidatorIndex) jamTypes.Validator {
	return posteriorCurrentValidators[authorIndex]
}
