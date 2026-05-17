// Stage 2 tests cover execution-layer audit logic: judgment signing and
// tranche-index derivation. See issue #931 for the full plan.
//
// `WaitNextTranche` is intentionally not retested here — main's
// auditing_test.go already has TestWaitNextTranche_{NormalExpiry,
// ContextCancel, AlreadyPastDeadline, Tranche2, ContextDeadlineExceeded}
// from #910, which fully covers this stage's planned cases.
//
// These tests share package-global state with stage 1 (types.* constants,
// blockchain singleton, header wall-clock) and MUST NOT call t.Parallel().
package auditing

import (
	"crypto/ed25519"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/header"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// BuildJudgements: signature verification (GP §17.17, v0.7.2)
//
//	jₙ = { S_κ[v]ᵉ ⟨Xe(w) ⌢ H(w)⟩ | (c, w) ∈ aₙ }
//
// Xe = "jam_valid" / "jam_invalid" depending on AuditResult.
// ===========================================================================

func TestStage2_BuildJudgements_SignatureVerification(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	seed[0] = 7
	privKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privKey.Public().(ed25519.PublicKey)

	report0 := makeDetailedWorkReport(0xA0, 0)
	report1 := makeDetailedWorkReport(0xB1, 1)

	auditReports := []types.AuditReport{
		{CoreID: 0, Report: report0, ValidatorID: 3, AuditResult: true},  // → JamValid
		{CoreID: 1, Report: report1, ValidatorID: 3, AuditResult: false}, // → JamInvalid
	}

	signed := BuildJudgements(0, auditReports, hash.Blake2bHash, 3, privKey)
	require.Len(t, signed, 2)

	// Reconstruct each ⟨Xe(w) ⌢ H(w)⟩ message and verify with the public key.
	for _, audit := range signed {
		var xe []byte
		if audit.AuditResult {
			xe = []byte(types.JamValid)
		} else {
			xe = []byte(types.JamInvalid)
		}
		hashW := hash.Blake2bHash(utilities.WorkReportSerialization(audit.Report))
		msg := append(append([]byte{}, xe...), hashW[:]...)

		assert.True(t, ed25519.Verify(pubKey, msg, audit.Signature[:]),
			"signature for core %d (result=%v) must verify against ⟨Xe ⌢ H(w)⟩",
			audit.CoreID, audit.AuditResult)
	}
}

// Empty input must not panic; zero-iteration loop returns the empty slice.
func TestStage2_BuildJudgements_EmptyInput(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	seed[0] = 7
	privKey := ed25519.NewKeyFromSeed(seed)

	signed := BuildJudgements(0, []types.AuditReport{}, hash.Blake2bHash, 3, privKey)
	assert.Empty(t, signed)
}

// ===========================================================================
// GetTranchIndex: formula n = (T − P·Ht) / A (GP §17.8)
// ===========================================================================

func TestStage2_GetTranchIndex(t *testing.T) {
	setupChainState(t)

	// Place the block's slot ~120 seconds before "now" so n ≈ 120/A is a
	// stable, non-trivial answer (away from 0).
	const targetSecondsBack = 120
	nowSec := header.GetCurrentTimeInSecond()
	slot := types.TimeSlot((nowSec - targetSecondsBack) / uint64(types.SlotPeriod))

	blockchain.GetInstance().GetProcessingBlockPointer().SetHeader(types.Header{
		Slot: slot,
	})

	// Wrap the call between two wall-clock reads. The window is microseconds
	// in practice, so expectedMin == expectedMax almost always; the wrap covers
	// the case where T crosses a tranche boundary during the call.
	P := types.U64(types.SlotPeriod)
	A := types.U64(types.TranchePeriod)

	tBefore := types.U64(header.GetCurrentTimeInSecond())
	got := GetTranchIndex()
	tAfter := types.U64(header.GetCurrentTimeInSecond())

	expectedMin := (tBefore - P*types.U64(slot)) / A
	expectedMax := (tAfter - P*types.U64(slot)) / A

	assert.GreaterOrEqual(t, uint64(got), uint64(expectedMin),
		"n must be ≥ floor((T_before − P·Ht) / A)")
	assert.LessOrEqual(t, uint64(got), uint64(expectedMax),
		"n must be ≤ floor((T_after − P·Ht) / A)")
}

// Block slot in the future (clock skew, malformed header, future-slot lookahead)
// must not underflow u64 arithmetic into a huge bogus tranche index.
// Production code clamps T < P·Ht to n = 0.
func TestStage2_GetTranchIndex_FutureSlot(t *testing.T) {
	setupChainState(t)

	// Slot derived from a wall-clock value 600s ahead of now → P·Ht > T.
	nowSec := header.GetCurrentTimeInSecond()
	futureSlot := types.TimeSlot((nowSec + 600) / uint64(types.SlotPeriod))
	blockchain.GetInstance().GetProcessingBlockPointer().SetHeader(types.Header{
		Slot: futureSlot,
	})

	got := GetTranchIndex()
	assert.Equal(t, types.U64(0), got,
		"future slot must clamp to 0, not wrap u64")
}
