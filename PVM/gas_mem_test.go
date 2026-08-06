package PVM

import (
	"math"
	"testing"
)

func TestFetchCost(t *testing.T) {
	tests := []struct {
		name          string
		discriminator uint64
		length        uint64
		want          Gas
	}{
		{name: "constant only", discriminator: 0, length: 1024, want: 390},
		{name: "rounded linear term", discriminator: 2, length: 1, want: 81},
		{name: "full linear unit", discriminator: 2, length: 1024, want: 176},
		{name: "unknown discriminator", discriminator: 16, length: 1024, want: 80},
		{name: "saturates", discriminator: 2, length: math.MaxUint64, want: math.MaxInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fetchCost(tt.discriminator, tt.length); got != tt.want {
				t.Fatalf("fetchCost(%d, %d) = %d, want %d", tt.discriminator, tt.length, got, tt.want)
			}
		})
	}
}
