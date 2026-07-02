package accumulation

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TestAccumulationPrefixLen_V080GasBudget covers the GP v0.8.0 eq:accseq
// prefix-selection budget: sum(digest gas) + sum(incoming transfer gas) +
// sum(free-accumulation gas) <= g. Under v0.7.x only the digest gas counted.
func TestAccumulationPrefixLen_V080GasBudget(t *testing.T) {
	report := func(gas types.Gas) types.WorkReport {
		return types.WorkReport{
			Results: []types.WorkResult{{AccumulateGas: gas}},
		}
	}
	reports := []types.WorkReport{report(100), report(100), report(100)}

	cases := []struct {
		name      string
		g         types.Gas
		transfers []types.DeferredTransfer
		free      types.AlwaysAccumulateMap
		want      int
	}{
		{
			name: "no reservations: all reports fit",
			g:    300,
			want: 3,
		},
		{
			name:      "incoming transfer gas is reserved (v0.8.0)",
			g:         300,
			transfers: []types.DeferredTransfer{{GasLimit: 150}},
			want:      1, // 150 + 100 <= 300, 150 + 200 > 300
		},
		{
			name: "free-accumulation gas is reserved (v0.8.0)",
			g:    300,
			free: types.AlwaysAccumulateMap{types.ServiceID(1): 250},
			want: 0, // 250 + 100 > 300
		},
		{
			name:      "reservations alone exhaust the budget",
			g:         100,
			transfers: []types.DeferredTransfer{{GasLimit: 60}},
			free:      types.AlwaysAccumulateMap{types.ServiceID(1): 60},
			want:      0,
		},
		{
			name: "exact fit is included",
			g:    250,
			free: types.AlwaysAccumulateMap{types.ServiceID(1): 50},
			want: 2, // 50 + 200 <= 250, 50 + 300 > 250
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := accumulationPrefixLen(tc.g, tc.transfers, reports, tc.free)
			if got != tc.want {
				t.Errorf("accumulationPrefixLen = %d, want %d", got, tc.want)
			}
		})
	}
}
