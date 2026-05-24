package store

import (
	"bytes"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// DirtyEntry represents a single state change between two consecutive blocks.
type DirtyEntry struct {
	Key      types.StateKey
	NewValue []byte // nil for deletes
	IsDelete bool
}

// DiffSortedKeyVals performs a merge-scan on two sorted StateKeyVals slices
// and returns the set of dirty entries (inserted, deleted, modified).
// Both inputs MUST be sorted by Key in ascending order.
// Cost: O(n) byte comparisons where n = max(len(prior), len(current)).
func DiffSortedKeyVals(prior, current types.StateKeyVals) []DirtyEntry {
	var result []DirtyEntry
	i, j := 0, 0

	for i < len(prior) && j < len(current) {
		cmp := bytes.Compare(prior[i].Key[:], current[j].Key[:])
		switch {
		case cmp < 0:
			// Key in prior but not in current → deleted
			result = append(result, DirtyEntry{
				Key:      prior[i].Key,
				IsDelete: true,
			})
			i++
		case cmp > 0:
			// Key in current but not in prior → inserted
			result = append(result, DirtyEntry{
				Key:      current[j].Key,
				NewValue: current[j].Value,
			})
			j++
		default:
			// Same key — check if value changed
			if !bytes.Equal(prior[i].Value, current[j].Value) {
				result = append(result, DirtyEntry{
					Key:      current[j].Key,
					NewValue: current[j].Value,
				})
			}
			i++
			j++
		}
	}

	// Remaining in prior → all deleted
	for ; i < len(prior); i++ {
		result = append(result, DirtyEntry{
			Key:      prior[i].Key,
			IsDelete: true,
		})
	}

	// Remaining in current → all inserted
	for ; j < len(current); j++ {
		result = append(result, DirtyEntry{
			Key:      current[j].Key,
			NewValue: current[j].Value,
		})
	}

	return result
}
