package extrinsic

import (
	"bytes"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func Len(p []jam_types.Preimage) int {
	return len(p)
}

func Less(p []jam_types.Preimage, i, j int) bool {
	if p[i].Requester == p[j].Requester {
		return bytes.Compare(p[i].Blob, p[j].Blob) < 0
	}
	return p[i].Requester < p[j].Requester
}

func Swap(p []jam_types.Preimage, i, j int) {
	p[i], p[j] = p[j], p[i]
}

func Sort(p []jam_types.Preimage) {
	sort.Slice(p, func(i, j int) bool {
		return Less(p, i, j)
	})
}

// Add adds a new Preimage to the Preimages slice.
// It also removes the duplicates from the slice.
func Add(p []jam_types.Preimage, newPreimage jam_types.Preimage) []jam_types.Preimage {
	p = append(p, newPreimage)
	return RemoveDuplicates(p)
}

// RemoveDuplicates removes the duplicates from the Preimages slice.
// It uses the Sort function to sort the slice first.
// And uses double pointers to remove the duplicates.
func RemoveDuplicates(p []jam_types.Preimage) []jam_types.Preimage {
	if len(p) == 0 {
		return p
	}

	// First, sort the slice
	Sort(p)

	// Then, remove the duplicates
	j := 0
	for i := 1; i < len(p); i++ {
		if p[i].Requester != p[j].Requester || !bytes.Equal(p[i].Blob, p[j].Blob) {
			j++
			p[j] = p[i]
		}
	}

	// Remove the duplicates
	return p[:j+1]
}
