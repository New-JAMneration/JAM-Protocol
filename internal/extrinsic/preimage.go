package extrinsic

import (
	"bytes"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// PreimagesExtrinsic is a local type that embeds jam_types.PreimagesExtrinsic
type PreimagesExtrinsic struct {
	jam_types.PreimagesExtrinsic
}

// Len returns the number of elements in the Preimages slice.
func (p *PreimagesExtrinsic) Len() int {
	return len(p.PreimagesExtrinsic)
}

// Less reports whether the element with index i should sort before the element with index j.
func (p *PreimagesExtrinsic) Less(i, j int) bool {
	if p.PreimagesExtrinsic[i].Requester == p.PreimagesExtrinsic[j].Requester {
		return bytes.Compare(p.PreimagesExtrinsic[i].Blob, p.PreimagesExtrinsic[j].Blob) < 0
	}
	return p.PreimagesExtrinsic[i].Requester < p.PreimagesExtrinsic[j].Requester
}

// Swap swaps the elements with indexes i and j.
func (p *PreimagesExtrinsic) Swap(i, j int) {
	p.PreimagesExtrinsic[i], p.PreimagesExtrinsic[j] = p.PreimagesExtrinsic[j], p.PreimagesExtrinsic[i]
}

// Sort sorts the Preimages in place.
func (p *PreimagesExtrinsic) Sort() {
	sort.Sort(p)
}

// Add adds a new Preimage to the Preimages slice.
func (p *PreimagesExtrinsic) Add(newPreimage jam_types.Preimage) {
	p.PreimagesExtrinsic = append(p.PreimagesExtrinsic, newPreimage)
	p.RemoveDuplicates()
}

// RemoveDuplicates removes the duplicates from the Preimages slice.
func (p *PreimagesExtrinsic) RemoveDuplicates() {
	if len(p.PreimagesExtrinsic) == 0 {
		return
	}

	// First, sort the slice
	p.Sort()

	// Then, remove the duplicates
	j := 0
	for i := 1; i < len(p.PreimagesExtrinsic); i++ {
		if p.PreimagesExtrinsic[i].Requester != p.PreimagesExtrinsic[j].Requester || !bytes.Equal(p.PreimagesExtrinsic[i].Blob, p.PreimagesExtrinsic[j].Blob) {
			j++
			p.PreimagesExtrinsic[j] = p.PreimagesExtrinsic[i]
		}
	}

	// Remove the duplicates
	p.PreimagesExtrinsic = p.PreimagesExtrinsic[:j+1]
}
