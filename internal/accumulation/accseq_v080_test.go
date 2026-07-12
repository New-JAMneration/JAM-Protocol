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

// TestOuterAccumulation_V080RecursionBudgetAndProcessedTransfers stubs the
// PVM-backed ∆* (via the parallelize indirection) to exercise the two
// remaining eq:accseq changes end to end: the recursion budget
// g* = g + Σ(t*_gas) − Σ(u*) sums the transfers PRODUCED by the round (not
// the incoming ones), and the processed-transfers output preserves the
// t ⌢ t† order across recursion rounds.
func TestOuterAccumulation_V080RecursionBudgetAndProcessedTransfers(t *testing.T) {
	report := func(gas types.Gas) types.WorkReport {
		return types.WorkReport{Results: []types.WorkResult{{AccumulateGas: gas}}}
	}
	incoming := types.DeferredTransfer{ReceiverID: 7, GasLimit: 50}
	produced := types.DeferredTransfer{ReceiverID: 8, GasLimit: 5}

	run := func(t *testing.T, roundOneGasUsed types.Gas) OuterAccumulationOutput {
		t.Helper()

		calls := 0
		orig := parallelize
		parallelize = func(input ParallelizedAccumulationInput) (ParallelizedAccumulationOutput, error) {
			calls++
			out := ParallelizedAccumulationOutput{
				PartialStateSet:          input.PartialStateSet,
				AccumulatedServiceOutput: types.AccumulatedServiceOutput{},
			}
			if calls == 1 {
				out.DeferredTransfers = []types.DeferredTransfer{produced}
				out.ServiceGasUsedList = types.ServiceGasUsedList{{ServiceID: 1, Gas: roundOneGasUsed}}
			}
			return out, nil
		}
		t.Cleanup(func() { parallelize = orig })

		output, err := OuterAccumulation(OuterAccumulationInput{
			GasLimit:          100,
			DeferredTransfers: []types.DeferredTransfer{incoming},
			WorkReports:       []types.WorkReport{report(30), report(30)},
		})
		if err != nil {
			t.Fatalf("OuterAccumulation: %v", err)
		}
		return output
	}

	t.Run("g* sums produced transfers, not incoming", func(t *testing.T) {
		// Round 1: reserved 50 (incoming) + 30 <= 100 selects exactly one
		// report. Recursion budget g* = 100 + 5 (produced t*) − 100 (u*) = 5,
		// too small for the second report (5 reserved + 30 digest > 5). Under
		// the v0.7.x rule (incoming t: 100 + 50 − 100 = 50) it would still
		// fit — the accumulated count tells the two apart.
		output := run(t, 100)
		if got := output.NumberOfWorkResultsAccumulated; got != 1 {
			t.Errorf("accumulated = %d, want 1 (g* must sum produced t*, not incoming t)", got)
		}
	})

	t.Run("larger g* admits the second report", func(t *testing.T) {
		// Same setup with zero gas used: g* = 100 + 5 − 0 = 105 fits the
		// second report — proving the count above is budget-driven.
		output := run(t, 0)
		if got := output.NumberOfWorkResultsAccumulated; got != 2 {
			t.Errorf("accumulated = %d, want 2", got)
		}
	})

	t.Run("processed transfers preserve t concat t-dagger order", func(t *testing.T) {
		output := run(t, 100)
		want := []types.ServiceID{7, 8} // incoming first, then the round's produced
		if len(output.ProcessedTransfers) != len(want) {
			t.Fatalf("processed transfers = %d, want %d", len(output.ProcessedTransfers), len(want))
		}
		for i, receiver := range want {
			if output.ProcessedTransfers[i].ReceiverID != receiver {
				t.Errorf("processed[%d].ReceiverID = %d, want %d", i, output.ProcessedTransfers[i].ReceiverID, receiver)
			}
		}
	})
}
