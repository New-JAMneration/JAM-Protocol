package extrinsic

import (
	"bytes"
	"sort"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

type PreimageController struct {
	Preimages []jamTypes.Preimage
}

// NewPreimageController creates a new PreimageController with the given
// Preimages slice.
func NewPreimageController() *PreimageController {
	return &PreimageController{
		Preimages: make([]jamTypes.Preimage, 0),
	}
}

// Set sets the Preimages slice to the given Preimages slice.
func (p *PreimageController) Set(preimages []jamTypes.Preimage) {
	p.Preimages = preimages
}

// Len returns the length of the Preimages slice.
func (p *PreimageController) Len() int {
	return len(p.Preimages)
}

// Less returns true if the Preimage at index i is less than the Preimage at
// index j.
func (p *PreimageController) Less(i, j int) bool {
	if p.Preimages[i].Requester == p.Preimages[j].Requester {
		return bytes.Compare(p.Preimages[i].Blob, p.Preimages[j].Blob) < 0
	}
	return p.Preimages[i].Requester < p.Preimages[j].Requester
}

// Swap swaps the Preimages at index i and j.
func (p *PreimageController) Swap(i, j int) {
	p.Preimages[i], p.Preimages[j] = p.Preimages[j], p.Preimages[i]
}

// Sort sorts the Preimages slice.
func (p *PreimageController) Sort() {
	sort.Sort(p)
}

// Add adds a new Preimage to the Preimages slice.
// It also removes the duplicates from the slice.
func (p *PreimageController) Add(newPreimage jamTypes.Preimage) {
	p.Preimages = append(p.Preimages, newPreimage)
	p.RemoveDuplicates()
}

// RemoveDuplicates removes the duplicates from the Preimages slice.
// It uses the Sort function to sort the slice first.
// And uses double pointers to remove the duplicates.
// See equation (12.29)
func (p *PreimageController) RemoveDuplicates() {
	if len(p.Preimages) == 0 {
		return
	}

	// First, sort the slice
	p.Sort()

	// Then, remove the duplicates
	j := 0
	for i := 1; i < len(p.Preimages); i++ {
		if p.Preimages[i].Requester != p.Preimages[j].Requester || !bytes.Equal(p.Preimages[i].Blob, p.Preimages[j].Blob) {
			j++
			p.Preimages[j] = p.Preimages[i]
		}
	}

	// Remove the duplicates
	p.Preimages = p.Preimages[:j+1]
}
