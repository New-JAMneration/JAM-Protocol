// Stage 1 tests — pure audit judgment logic that does not depend on
// networking, VRF, or PVM execution. See docs/ce-audit-test-plan.md
// and issue #931 for the full staged test plan.
package auditing

import (
	"crypto/ed25519"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Stage 1 helpers
// ---------------------------------------------------------------------------

// setupDeterministicAuditChain sets tiny mode, resets blockchain singleton,
// and returns a fixed Ed25519 private key for signing tests.
func setupDeterministicAuditChain(t *testing.T) ed25519.PrivateKey {
	t.Helper()

	types.SetTinyMode()
	blockchain.ResetInstance()
	t.Cleanup(blockchain.ResetInstance)

	cs := blockchain.GetInstance()
	var entropy types.BandersnatchVrfSignature
	entropy[0] = 1

	cs.GetProcessingBlockPointer().SetHeader(types.Header{
		Slot:          100,
		AuthorIndex:   0,
		EntropySource: entropy,
	})

	seed := make([]byte, ed25519.SeedSize)
	seed[0] = 42
	return ed25519.NewKeyFromSeed(seed)
}

// makeDetailedWorkReport builds a WorkReport with all fields populated using
// unique values derived from id, so map lookups stay easy to trace in tests.
func makeDetailedWorkReport(id byte, coreID types.CoreIndex) types.WorkReport {
	var hashValue types.WorkPackageHash
	hashValue[0] = id

	var erasureRoot types.ErasureRoot
	erasureRoot[0] = id + 10

	var exportsRoot types.ExportsRoot
	exportsRoot[0] = id + 20

	var codeHash types.OpaqueHash
	codeHash[0] = id + 30

	var payloadHash types.OpaqueHash
	payloadHash[0] = id + 40

	return types.WorkReport{
		PackageSpec: types.WorkPackageSpec{
			Hash:         hashValue,
			Length:       types.U32(128 + id),
			ErasureRoot:  erasureRoot,
			ExportsRoot:  exportsRoot,
			ExportsCount: 1,
		},
		Context:   types.RefineContext{},
		CoreIndex: coreID,
		Results: []types.WorkResult{
			{
				ServiceID:     types.ServiceID(id),
				CodeHash:      codeHash,
				PayloadHash:   payloadHash,
				AccumulateGas: types.Gas(10 + id),
				Result:        types.GetWorkExecResult(types.WorkExecResultOk, []byte{id}),
				RefineLoad: types.RefineLoad{
					GasUsed:        types.Gas(1),
					Imports:        1,
					ExtrinsicCount: 1,
					ExtrinsicSize:  1,
					Exports:        1,
				},
			},
		},
	}
}

// makeDetailedAvailabilityAssignments wraps WorkReports into ρ format,
// placing each report at its CoreIndex position.
func makeDetailedAvailabilityAssignments(reports ...types.WorkReport) types.AvailabilityAssignments {
	assignments := make(types.AvailabilityAssignments, types.CoresCount)
	for _, report := range reports {
		assignments[report.CoreIndex] = &types.AvailabilityAssignment{
			Report:       report,
			AssignedSlot: 1,
		}
	}
	return assignments
}

// ---------------------------------------------------------------------------
// CollectAuditReportCandidates (GP 17.1-17.2)
// ---------------------------------------------------------------------------

// GP 17.2: Q keeps reports that are both assigned (ρ) and available (W), filters out the rest.
func TestCollectCandidates(t *testing.T) {
	setupDeterministicAuditChain(t)

	// Use 5 cores so we can test a mix of assigned/available/filtered.
	types.CoresCount = 5
	t.Cleanup(func() { types.SetTinyMode() })

	cs := blockchain.GetInstance()
	r0 := makeDetailedWorkReport(1, 0)
	r1 := makeDetailedWorkReport(2, 1)
	r2 := makeDetailedWorkReport(3, 2)
	r3 := makeDetailedWorkReport(4, 3)
	// core 4: not assigned (nil in ρ)

	cs.GetPriorStates().SetRho(makeDetailedAvailabilityAssignments(r0, r1, r2, r3))
	// W only has r0 and r1 — r2 and r3 are assigned but not available
	cs.GetIntermediateStates().SetAvailableWorkReports([]types.WorkReport{r0, r1})

	got := CollectAuditReportCandidates()

	require.Len(t, got, 5)
	require.NotNil(t, got[0], "r0: assigned + available → kept")
	require.NotNil(t, got[1], "r1: assigned + available → kept")
	assert.Nil(t, got[2], "r2: assigned but not available → filtered")
	assert.Nil(t, got[3], "r3: assigned but not available → filtered")
	assert.Nil(t, got[4], "core 4: not assigned → nil")
	assert.Equal(t, r0.PackageSpec.Hash, got[0].PackageSpec.Hash)
	assert.Equal(t, r1.PackageSpec.Hash, got[1].PackageSpec.Hash)
}

// ---------------------------------------------------------------------------
// buildInitialAuditAssignmentFromCoreOrder (GP 17.5)
// ---------------------------------------------------------------------------

// Regression (#935): each AuditReport must carry the originating ValidatorID.
// Also verify core traversal order and nil skipping.
func TestInitialAssignment_SetsValidatorID(t *testing.T) {
	setupDeterministicAuditChain(t)

	types.CoresCount = 5
	t.Cleanup(func() { types.SetTinyMode() })

	r0 := makeDetailedWorkReport(1, 0)
	r1 := makeDetailedWorkReport(2, 1)
	r2 := makeDetailedWorkReport(3, 2)
	// core 3, 4: nil (no report)

	Q := make([]*types.WorkReport, types.CoresCount)
	Q[0] = &r0
	Q[1] = &r1
	Q[2] = &r2

	validatorIndex := types.ValidatorIndex(4)
	// shuffle order: 3(nil) → 1(r1) → 4(nil) → 0(r0) → 2(r2)
	got := buildInitialAuditAssignmentFromCoreOrder(Q, validatorIndex, []types.U32{3, 1, 4, 0, 2})

	// Only non-nil cores picked: r1, r0, r2 (in shuffle order)
	require.Len(t, got, 3)
	assert.Equal(t, types.CoreIndex(1), got[0].CoreID, "first non-nil in shuffle order")
	assert.Equal(t, types.CoreIndex(0), got[1].CoreID)
	assert.Equal(t, types.CoreIndex(2), got[2].CoreID)

	for _, audit := range got {
		assert.Equal(t, validatorIndex, audit.ValidatorID, "ValidatorID must be set, not default 0")
		assert.False(t, audit.AuditResult)
	}
}

// ---------------------------------------------------------------------------
// BuildAnnouncement (GP 17.9-17.11)
// ---------------------------------------------------------------------------

// Reconstructed signing context (XI ⌢ n ⌢ xn ⌢ H(H)) must verify against produced Ed25519 signature.
func TestAnnouncement_SignContext(t *testing.T) {
	privKey := setupDeterministicAuditChain(t)
	pubKey := privKey.Public().(ed25519.PublicKey)

	report0 := makeDetailedWorkReport(12, 0)
	report1 := makeDetailedWorkReport(13, 1)
	assignments := []types.AuditReport{
		{CoreID: 0, Report: report0, ValidatorID: 2},
		{CoreID: 1, Report: report1, ValidatorID: 2},
	}

	signature, err := BuildAnnouncement(0, assignments, hash.Blake2bHash, 2, privKey)
	require.NoError(t, err)

	var xnPayload types.ByteSequence
	for _, pair := range assignments {
		xnPayload = append(xnPayload, utilities.SerializeFixedLength(types.U64(pair.CoreID), 2)...)
		reportHash := hash.Blake2bHash(utilities.WorkReportSerialization(pair.Report))
		xnPayload = append(xnPayload, reportHash[:]...)
	}

	headerBytes, err := utilities.HeaderSerialization(blockchain.GetInstance().GetProcessingBlockPointer().GetHeader())
	require.NoError(t, err)
	headerHash := hash.Blake2bHash(headerBytes)

	xi := []byte(types.JamAnnounce)
	expectedContext := make(types.ByteSequence, 0, len(xi)+1+len(xnPayload)+len(headerHash))
	expectedContext = append(expectedContext, xi...)
	expectedContext = append(expectedContext, byte(0))
	expectedContext = append(expectedContext, utilities.SerializeByteSequence(xnPayload)...)
	expectedContext = append(expectedContext, headerHash[:]...)

	assert.True(t, ed25519.Verify(pubKey, expectedContext, signature[:]))
}
