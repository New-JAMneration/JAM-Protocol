package extrinsic

import "sort"

// Preimage is a structure that contains the requester and the blob.
// It is used to store the preimages in the Extrinsic.
// You can find more information about the preimage in the gray paper:
// 12.4 Preimage integration, formula (12.28), (12.29)
type Preimage struct {
	Requester uint32 `json:"requester"` // service index
	Blob      string `json:"blob"`      // octet string
}

// Preimages is a slice of Preimage.
type Preimages []Preimage

// Len is the number of elements in the collection.
func (p *Preimages) Len() int {
	return len(*p)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p *Preimages) Less(i, j int) bool {
	if (*p)[i].Requester == (*p)[j].Requester {
		return (*p)[i].Blob < (*p)[j].Blob
	}
	return (*p)[i].Requester < (*p)[j].Requester
}

// Swap swaps the elements with indexes i and j.
func (p *Preimages) Swap(i, j int) {
	(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
}

// Sort sorts the Preimages in place.
func (p *Preimages) Sort() {
	sort.Sort(p)
}

// Add adds a new Preimage to the Preimages slice.
func (p *Preimages) Add(newPreimage Preimage) {
	if len(*p) == 0 {
		// Add the new Preimage
		*p = append(*p, newPreimage)
		return
	}

	// Add the new Preimage
	*p = append(*p, newPreimage)

	// Remove the duplicates (includes sorting)
	p.RemoveDuplicates()
}

// RemoveDuplicates removes the duplicates from the Preimages slice.
// It uses the two-pointer technique to ensure that each preimage is unique.
// The time complexity is O(n log n) because of the sorting.
func (p *Preimages) RemoveDuplicates() {
	if len(*p) == 0 {
		return
	}

	// First, sort the slice
	p.Sort()

	// Then, remove the duplicates
	j := 0
	for i := 1; i < len(*p); i++ {
		if (*p)[i].Requester != (*p)[j].Requester || (*p)[i].Blob != (*p)[j].Blob {
			j++
			(*p)[j] = (*p)[i]
		}
	}

	// Remove the duplicates
	*p = (*p)[:j+1]
}
