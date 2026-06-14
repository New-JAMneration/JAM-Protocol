// Followup tests for two coverage gaps reviewers flagged after stages 1-3
// landed: GP §17.15 ComputeAnForValidator threshold formula, and GP §17.5
// initial-assignment 10-core cap. See issues #931, #956.
package auditing

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// GP §17.15 isAssignedByThreshold
//
//	Y(sₙ(w))[0] · V / (256 · F)  <  mₙ   ⇒  validator is assigned
//
// V = ValidatorsCount, F = BiasFactor (= 2 per §17.16; Burdges/Cevallos 2024
// modeling: one no-show expects F = 2 validators to pick up the report next
// round, the cryptoeconomic recovery invariant for audit).
//
// Production code lives in auditing.go. The threshold check was extracted
// from inside ComputeAnForValidator into a pure helper so it can be tested
// without invoking the Bandersnatch VRF (Rust FFI).
// ===========================================================================

// Boundary cases: pick snByte0 values that straddle the integer-truncation
// threshold for several (V, F, mn) combinations and verify the assign / not
// decision flips at the right place. Catches: F = 1 instead of 2, < swapped
// to ≤, multiplication and division swapped, byte-vs-int order issues.
func TestIsAssignedByThreshold_BoundaryCases(t *testing.T) {
	tests := []struct {
		name            string
		snByte0         byte
		validatorsCount int
		biasFactor      int
		noShowCount     int
		wantAssigned    bool
	}{
		// Tiny mode V=6, F=2. Threshold for mn=1 is 256·2·1/6 ≈ 85.33,
		// so snByte0 ≤ 85 is assigned, ≥ 86 is not.
		{"tiny V=6 F=2 mn=1, X=0 (far below)", 0, 6, 2, 1, true},
		{"tiny V=6 F=2 mn=1, X=85 (just below boundary)", 85, 6, 2, 1, true},
		{"tiny V=6 F=2 mn=1, X=86 (just above boundary)", 86, 6, 2, 1, false},
		{"tiny V=6 F=2 mn=1, X=255 (far above)", 255, 6, 2, 1, false},

		// mn=2 doubles the threshold: snByte0 ≤ 170 assigned, ≥ 171 not.
		{"tiny V=6 F=2 mn=2, X=170 (just below)", 170, 6, 2, 2, true},
		{"tiny V=6 F=2 mn=2, X=171 (just above)", 171, 6, 2, 2, false},

		// Full mode V=341, F=2. Threshold for mn=1 is 512/341 ≈ 1.50,
		// so only X=0 or 1 is assigned. Many fewer hits per validator
		// because the full validator set is much larger.
		{"full V=341 F=2 mn=1, X=0", 0, 341, 2, 1, true},
		{"full V=341 F=2 mn=1, X=1", 1, 341, 2, 1, true},
		{"full V=341 F=2 mn=1, X=2", 2, 341, 2, 1, false},

		// Defensive: mn = 0 should always be false (LHS ≥ 0, 0 < 0 false).
		// Production code guards this with an early continue, but the
		// helper itself must not over-assign.
		{"defensive: mn=0 returns false even at X=0", 0, 6, 2, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isAssignedByThreshold(tc.snByte0, tc.validatorsCount, tc.biasFactor, tc.noShowCount)
			assert.Equal(t, tc.wantAssigned, got)
		})
	}
}

// Exhaustive: walk all 256 byte values for fixed (V, F, mn) and assert the
// total number of "assigned" outcomes matches min(256, ⌈256·F·mn/V⌉). The
// 256 cap matters once the threshold reaches the byte domain: e.g. tiny
// mn=3 gives ⌈256·2·3/6⌉ = 256, so every X qualifies and the count
// saturates rather than overflowing the 256-value range. This pins the
// formula's behaviour across the full input domain in one go and catches
// off-by-one errors at the threshold edge that a sparse boundary test
// might miss.
func TestIsAssignedByThreshold_ExhaustiveByteRangeCount(t *testing.T) {
	tests := []struct {
		name             string
		validatorsCount  int
		biasFactor       int
		noShowCount      int
		wantAssignedHits int
	}{
		// tiny mn=1: threshold 256·2·1/6 = 85.33 → 86 values (X=0..85)
		{"tiny V=6 F=2 mn=1", 6, 2, 1, 86},
		// tiny mn=2: threshold 170.67 → 171 values
		{"tiny V=6 F=2 mn=2", 6, 2, 2, 171},
		// tiny mn=3: threshold 256, capped → all 256 X values qualify
		{"tiny V=6 F=2 mn=3", 6, 2, 3, 256},
		// full V=341 mn=1: only X=0,1 assigned
		{"full V=341 F=2 mn=1", 341, 2, 1, 2},
		// full V=341 mn=2: X=0..3 assigned (LHS goes 0,0,1,1,2 at X=0..4)
		{"full V=341 F=2 mn=2", 341, 2, 2, 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hits := 0
			for x := 0; x < 256; x++ {
				if isAssignedByThreshold(byte(x), tc.validatorsCount, tc.biasFactor, tc.noShowCount) {
					hits++
				}
			}
			assert.Equal(t, tc.wantAssignedHits, hits,
				"V=%d F=%d mn=%d: expected %d hits, got %d",
				tc.validatorsCount, tc.biasFactor, tc.noShowCount, tc.wantAssignedHits, hits)
		})
	}
}

// ===========================================================================
// GP §17.5 initial-assignment 10-core cap
//
//	a0 = top 10 of (c, Q[c]) where Q[c] ≠ ∅, in shuffle order.
//
// Stage 1's TestInitialAssignment_SetsValidatorID used 3 reports and never
// hit the 10-cap branch. This test feeds 12 non-nil reports in shuffle order
// and asserts the cap holds.
// ===========================================================================

func TestBuildInitialAuditAssignmentFromCoreOrder_TopTenCap(t *testing.T) {
	// Helper is pure (no chain-state lookups); no setupChainState needed.
	// Build 12 work reports across 12 cores (above the 10-cap threshold).
	const coreCount = 12
	Q := make([]*types.WorkReport, coreCount)
	for i := 0; i < coreCount; i++ {
		r := makeDetailedWorkReport(byte(i+1), types.CoreIndex(i))
		Q[i] = &r
	}

	// Shuffle order: identity 0..11.
	shuffled := make([]types.U32, coreCount)
	for i := range shuffled {
		shuffled[i] = types.U32(i)
	}

	got := buildInitialAuditAssignmentFromCoreOrder(Q, types.ValidatorIndex(7), shuffled)

	require.Len(t, got, 10, "must cap at top 10 per GP §17.5")
	// First 10 in shuffle order are picked; cores 10 and 11 are dropped.
	for i := 0; i < 10; i++ {
		assert.Equal(t, types.CoreIndex(i), got[i].CoreID,
			"position %d should hold core %d", i, i)
		assert.Equal(t, types.ValidatorIndex(7), got[i].ValidatorID,
			"position %d ValidatorID must propagate from caller", i)
	}
}

// Variant: nils interspersed in the shuffle order are skipped, the 10-cap
// counts only non-nil entries.
//
// Uses a different ValidatorIndex from TopTenCap (7 vs 3) to confirm the
// helper picks up the caller's value rather than a fixed one and to keep
// the propagation check varied across the two tests.
func TestBuildInitialAuditAssignmentFromCoreOrder_TopTenCap_SkipsNils(t *testing.T) {
	const coreCount = 14
	Q := make([]*types.WorkReport, coreCount)
	// Fill cores 0..13 with reports, then nil out cores 3, 7, 11.
	for i := 0; i < coreCount; i++ {
		r := makeDetailedWorkReport(byte(i+1), types.CoreIndex(i))
		Q[i] = &r
	}
	Q[3] = nil
	Q[7] = nil
	Q[11] = nil

	shuffled := make([]types.U32, coreCount)
	for i := range shuffled {
		shuffled[i] = types.U32(i)
	}

	got := buildInitialAuditAssignmentFromCoreOrder(Q, types.ValidatorIndex(3), shuffled)

	// 14 total - 3 nils = 11 non-nil, cap at 10.
	require.Len(t, got, 10)
	// Picked cores in order: 0, 1, 2, 4, 5, 6, 8, 9, 10, 12 (nils skipped).
	expectedCores := []types.CoreIndex{0, 1, 2, 4, 5, 6, 8, 9, 10, 12}
	for i, want := range expectedCores {
		assert.Equal(t, want, got[i].CoreID, "position %d", i)
		assert.Equal(t, types.ValidatorIndex(3), got[i].ValidatorID,
			"position %d ValidatorID must propagate", i)
	}
}
