package validator

import (
	"math"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type GridMapper struct {
	Previous types.ValidatorsData
	Current  types.ValidatorsData
	Next     types.ValidatorsData
}

func ComputeWidth(n int) int {
	if n <= 0 {
		return 1
	}
	w := int(math.Sqrt(float64(n)))
	if w < 1 {
		return 1
	}
	return w
}

func (g *GridMapper) NeighborIndicesInEpoch(index int) []int {
	n := len(g.Current)
	if n == 0 || index < 0 || index >= n {
		return nil
	}

	width := ComputeWidth(n)
	row := index / width
	col := index % width

	neighbors := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if i == index {
			continue
		}
		if (i/width) == row || (i%width) == col {
			neighbors = append(neighbors, i)
		}
	}

	return neighbors
}

func (g *GridMapper) AllNeighborValidators(index int) []types.Validator {
	if index < 0 {
		return nil
	}

	result := make([]types.Validator, 0)
	for _, i := range g.NeighborIndicesInEpoch(index) {
		result = append(result, g.Current[i])
	}

	if index < len(g.Previous) {
		result = append(result, g.Previous[index])
	}
	if index < len(g.Next) {
		result = append(result, g.Next[index])
	}

	return result
}

// IsNeighborInEpoch reports whether indices a and b are grid neighbours within the current epoch
// (same row or same column in the sqrt(V) grid).
func (g *GridMapper) IsNeighborInEpoch(a, b int) bool {
	n := len(g.Current)
	if n == 0 || a < 0 || b < 0 || a >= n || b >= n || a == b {
		return false
	}

	width := ComputeWidth(n)
	return (a/width) == (b/width) || (a%width) == (b%width)
}

// IsSameIndexCrossEpoch reports whether key is the Ed25519 public key of the validator at the
// same index in Previous or Next
func (g *GridMapper) IsSameIndexCrossEpoch(index int, key types.Ed25519Public) bool {
	if index < 0 {
		return false
	}
	if index < len(g.Previous) && g.Previous[index].Ed25519 == key {
		return true
	}
	if index < len(g.Next) && g.Next[index].Ed25519 == key {
		return true
	}
	return false
}

func (g *GridMapper) FindIndex(key types.Ed25519Public) (int, bool) {
	for i, v := range g.Current {
		if v.Ed25519 == key {
			return i, true
		}
	}
	return -1, false
}
