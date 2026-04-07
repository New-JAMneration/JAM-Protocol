package auditing

import (
	"context"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// Stage 2: WaitNextTranche — timer + context cancellation
// ===========================================================================

// Deadline nearly reached → returns immediately.
func TestStage2_WaitNextTranche_NormalExpiry(t *testing.T) {
	slotStart := time.Now().Add(-7900 * time.Millisecond)

	start := time.Now()
	err := WaitNextTranche(0, slotStart, context.Background())
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 500*time.Millisecond)
}

// Context cancelled mid-wait → returns context.Canceled.
func TestStage2_WaitNextTranche_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := WaitNextTranche(0, time.Now(), ctx)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, elapsed, 1*time.Second)
}

// Deadline already in the past → returns immediately.
func TestStage2_WaitNextTranche_AlreadyPast(t *testing.T) {
	start := time.Now()
	err := WaitNextTranche(0, time.Now().Add(-20*time.Second), context.Background())
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 100*time.Millisecond)
}

// ===========================================================================
// Stage 2: GetJudgement — mock BundleFetcher error paths
// ===========================================================================

type mockBundleFetcherS2 struct {
	bundle []byte
	err    error
}

func (m *mockBundleFetcherS2) FetchBundle(_ types.WorkReport) ([]byte, error) {
	return m.bundle, m.err
}

// Bundle fetch failure → false.
func TestStage2_GetJudgement_FetchError(t *testing.T) {
	original := DefaultBundleFetcher
	defer func() { DefaultBundleFetcher = original }()

	DefaultBundleFetcher = &mockBundleFetcherS2{bundle: nil, err: assert.AnError}

	ar := makeAuditReport(0x01, 0, 0, false)
	assert.False(t, GetJudgement(ar))
}

// Garbage bundle → Process fails → false.
func TestStage2_GetJudgement_ProcessError(t *testing.T) {
	original := DefaultBundleFetcher
	defer func() { DefaultBundleFetcher = original }()

	DefaultBundleFetcher = &mockBundleFetcherS2{bundle: []byte{0xDE, 0xAD}, err: nil}

	ar := makeAuditReport(0x01, 0, 0, false)
	assert.False(t, GetJudgement(ar))
}

// ===========================================================================
// Stage 2 TODO stubs
// ===========================================================================

func TestStage2_GetJudgement_SuccessPath(t *testing.T) {
	t.Skip("TODO: mock PVM executor → Process returns matching report → GetJudgement returns true (ref: TestPrepareInputs_Shared)")
}

func TestStage2_BuildJudgements_SignatureVerification(t *testing.T) {
	t.Skip("TODO: needs code fix first — BuildJudgements uses Kappa[v].Ed25519 (public key 32B) to call ed25519.Sign (needs private key 64B)")
}

func TestStage2_GetTranchIndex(t *testing.T) {
	t.Skip("TODO: verify formula n = (T - P*Ht) / A with known time/slot values")
}

func TestStage2_ComputeAnForValidator(t *testing.T) {
	t.Skip("TODO: stochastic assignment with fixed Bandersnatch key from pkg/Rust-VRF test fixtures")
}
