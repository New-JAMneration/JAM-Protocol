package extrinsic

import "sort"

// 12.4 Preimage integration
// (12.28), (12.29)
type Preimage struct {
	Requester uint32 // service index
	Blob      string // octet string
}

type Preimages []Preimage

// Len is the number of elements in the collection.
func (p Preimages) Len() int {
	return len(p)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p Preimages) Less(i, j int) bool {
	if p[i].Requester == p[j].Requester {
		return p[i].Blob < p[j].Blob
	}
	return p[i].Requester < p[j].Requester
}

// Swap swaps the elements with indexes i and j.
func (p Preimages) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Sort sorts the Preimages in place.
func (p Preimages) Sort() {
	sort.Sort(p)
}

// The add function will check the newPreimage does not already exist in the
// slice and sort the slice.
func (p *Preimages) Add(newPreimage Preimage) {
	// Check if the newPreimage already exists
	for _, preimage := range *p {
		if preimage.Requester == newPreimage.Requester && preimage.Blob == newPreimage.Blob {
			// If it already exists, do not add it
			return
		}
	}

	// Add the new Preimage
	*p = append(*p, newPreimage)

	// Sort the slice
	p.Sort()
}
