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
// helpers: build minimal but valid-looking fixtures
// ---------------------------------------------------------------------------

func makeWPHash(b byte) types.WorkPackageHash {
	var h types.WorkPackageHash
	h[0] = b
	return h
}

func makeWorkReport(hashByte byte, coreIndex types.CoreIndex) types.WorkReport {
	return types.WorkReport{
		PackageSpec: types.WorkPackageSpec{Hash: makeWPHash(hashByte)},
		CoreIndex:   coreIndex,
		Results:     []types.WorkResult{{}},
	}
}

func makeAuditReport(hashByte byte, coreIndex types.CoreIndex, validatorID types.ValidatorIndex, result bool) types.AuditReport {
	return types.AuditReport{
		CoreID:      coreIndex,
		Report:      makeWorkReport(hashByte, coreIndex),
		ValidatorID: validatorID,
		AuditResult: result,
	}
}

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

// GP 17.2: reports in ρ but absent from W must be excluded from Q.
func TestCollectAuditReportCandidatesFiltersUnavailableReports(t *testing.T) {
	setupDeterministicAuditChain(t)

	cs := blockchain.GetInstance()
	report0 := makeDetailedWorkReport(1, 0)
	report1 := makeDetailedWorkReport(2, 1)

	cs.GetPriorStates().SetRho(makeDetailedAvailabilityAssignments(report0, report1))
	cs.GetIntermediateStates().SetAvailableWorkReports([]types.WorkReport{report0})

	got := CollectAuditReportCandidates()

	require.Len(t, got, types.CoresCount)
	require.NotNil(t, got[0])
	assert.Equal(t, report0.PackageSpec.Hash, got[0].PackageSpec.Hash)
	assert.Nil(t, got[1])
}

// ---------------------------------------------------------------------------
// buildInitialAuditAssignmentFromCoreOrder (GP 17.5)
// ---------------------------------------------------------------------------

// Regression: derived AuditReport entries must carry the originating ValidatorID.
func TestBuildInitialAuditAssignmentFromCoreOrderSetsValidatorID(t *testing.T) {
	setupDeterministicAuditChain(t)

	report0 := makeDetailedWorkReport(1, 0)
	report1 := makeDetailedWorkReport(2, 1)
	Q := make([]*types.WorkReport, types.CoresCount)
	Q[0] = &report0
	Q[1] = &report1

	validatorIndex := types.ValidatorIndex(4)
	got := buildInitialAuditAssignmentFromCoreOrder(Q, validatorIndex, []types.U32{1, 0})

	require.Len(t, got, 2)
	for _, audit := range got {
		assert.Equal(t, validatorIndex, audit.ValidatorID)
		assert.False(t, audit.AuditResult)
	}
}

// ---------------------------------------------------------------------------
// UpdateAssignmentMap (GP 17.13)
// ---------------------------------------------------------------------------

// ValidatorID from computed assignments must propagate into the assignment map.
func TestUpdateAssignmentMapUsesValidatorIDFromComputedAssignments(t *testing.T) {
	setupDeterministicAuditChain(t)

	report0 := makeDetailedWorkReport(1, 0)
	report1 := makeDetailedWorkReport(2, 1)
	Q := make([]*types.WorkReport, types.CoresCount)
	Q[0] = &report0
	Q[1] = &report1

	validatorIndex := types.ValidatorIndex(4)
	assignments := buildInitialAuditAssignmentFromCoreOrder(Q, validatorIndex, []types.U32{0, 1})

	got := UpdateAssignmentMap(assignments, make(types.AssignmentMap))
	for _, audit := range assignments {
		assert.Contains(t, got[audit.Report.PackageSpec.Hash], validatorIndex)
	}
}

// Calling UpdateAssignmentMap twice accumulates validators, not overwrites.
func TestUpdateAssignmentMap_Accumulate(t *testing.T) {
	h := makeWPHash(0x01)
	r := types.WorkReport{PackageSpec: types.WorkPackageSpec{Hash: h}, Results: []types.WorkResult{{}}}

	am := make(types.AssignmentMap)
	am = UpdateAssignmentMap([]types.AuditReport{{Report: r, ValidatorID: 1}}, am)
	am = UpdateAssignmentMap([]types.AuditReport{{Report: r, ValidatorID: 2}}, am)

	assert.Equal(t, []types.ValidatorIndex{1, 2}, am[h])
}

// ---------------------------------------------------------------------------
// UpdatePositiveJudgersFromAudit
// ---------------------------------------------------------------------------

// Only AuditResult=true entries are recorded; false entries must be ignored.
func TestUpdatePositiveJudgersFromAudit(t *testing.T) {
	audits := []types.AuditReport{
		makeAuditReport(0x01, 0, 1, true),
		makeAuditReport(0x02, 1, 2, false),
		makeAuditReport(0x01, 0, 3, true),
	}

	pj := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)
	pj = UpdatePositiveJudgersFromAudit(audits, pj)

	h1 := audits[0].Report.PackageSpec.Hash
	h2 := audits[1].Report.PackageSpec.Hash

	assert.Len(t, pj[h1], 2)
	assert.True(t, pj[h1][1])
	assert.True(t, pj[h1][3])
	_, exists := pj[h2]
	assert.False(t, exists, "negative-only report should not appear in positiveJudgers")
}

// ---------------------------------------------------------------------------
// ClassifyJudgments
// ---------------------------------------------------------------------------

// Mixed positive/negative judgments must be split correctly; unrelated reports excluded.
func TestClassifyJudgments(t *testing.T) {
	report := makeWorkReport(0xCC, 3)

	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: true},
		{Report: report, ValidatorID: 2, AuditResult: false},
		{Report: report, ValidatorID: 3, AuditResult: true},
		{Report: makeWorkReport(0xDD, 4), ValidatorID: 5, AuditResult: true},
	}

	pos, neg := ClassifyJudgments(report, judgments)
	assert.Len(t, pos, 2)
	assert.Len(t, neg, 1)
	assert.True(t, pos[1])
	assert.True(t, pos[3])
	assert.True(t, neg[2])
}

// ---------------------------------------------------------------------------
// FilterJudgments
// ---------------------------------------------------------------------------

// Only judgments matching the target work-report hash are returned.
func TestFilterJudgments(t *testing.T) {
	target := makeWorkReport(0xAA, 0)
	other := makeWorkReport(0xBB, 1)

	judgments := []types.AuditReport{
		{Report: target, ValidatorID: 1, AuditResult: true},
		{Report: other, ValidatorID: 2, AuditResult: true},
		{Report: target, ValidatorID: 3, AuditResult: false},
	}

	filtered := FilterJudgments(judgments, target.PackageSpec.Hash)
	assert.Len(t, filtered, 2)
}

// ---------------------------------------------------------------------------
// IsWorkReportAudited (GP 17.19)
// ---------------------------------------------------------------------------

// Rule 1: all assigned validators gave positive, no negatives → audited.
func TestIsWorkReportAudited_AllAssignedPositive(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: true},
		{Report: report, ValidatorID: 2, AuditResult: true},
	}
	assert.True(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1, 2}))
}

// Any negative judgment blocks Rule 1.
func TestIsWorkReportAudited_HasNegative(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: true},
		{Report: report, ValidatorID: 2, AuditResult: false},
	}
	assert.False(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1, 2}))
}

// Assigned validator missing from judgments blocks Rule 1.
func TestIsWorkReportAudited_MissingAssigned(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: true},
	}
	assert.False(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1, 2}))
}

// Rule 2: supermajority of positives overrides unconfirmed assigned validators.
func TestIsWorkReportAudited_Supermajority(t *testing.T) {
	report := makeWorkReport(0xAA, 0)

	var judgments []types.AuditReport
	for i := 0; i < types.ValidatorsSuperMajority; i++ {
		judgments = append(judgments, types.AuditReport{
			Report:      report,
			ValidatorID: types.ValidatorIndex(100 + i),
			AuditResult: true,
		})
	}

	assert.True(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1, 2}))
}

// One below supermajority with unconfirmed assigned → not audited.
func TestIsWorkReportAudited_BelowSupermajority(t *testing.T) {
	report := makeWorkReport(0xAA, 0)

	var judgments []types.AuditReport
	for i := 0; i < types.ValidatorsSuperMajority-1; i++ {
		judgments = append(judgments, types.AuditReport{
			Report:      report,
			ValidatorID: types.ValidatorIndex(100 + i),
			AuditResult: true,
		})
	}

	assert.False(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1}))
}

// ---------------------------------------------------------------------------
// IsBlockAudited (GP 17.20)
// ---------------------------------------------------------------------------

// All reports have sufficient positive judgments → block audited.
func TestIsBlockAudited_AllPassed(t *testing.T) {
	r1 := makeWorkReport(0x01, 0)
	r2 := makeWorkReport(0x02, 1)

	judgments := []types.AuditReport{
		{Report: r1, ValidatorID: 10, AuditResult: true},
		{Report: r2, ValidatorID: 20, AuditResult: true},
	}
	assignmentMap := map[types.WorkPackageHash][]types.ValidatorIndex{
		r1.PackageSpec.Hash: {10},
		r2.PackageSpec.Hash: {20},
	}

	assert.True(t, IsBlockAudited([]types.WorkReport{r1, r2}, judgments, assignmentMap))
}

// One report missing judgment → block NOT audited.
func TestIsBlockAudited_OneFailed(t *testing.T) {
	r1 := makeWorkReport(0x01, 0)
	r2 := makeWorkReport(0x02, 1)

	judgments := []types.AuditReport{
		{Report: r1, ValidatorID: 10, AuditResult: true},
	}
	assignmentMap := map[types.WorkPackageHash][]types.ValidatorIndex{
		r1.PackageSpec.Hash: {10},
		r2.PackageSpec.Hash: {20},
	}

	assert.False(t, IsBlockAudited([]types.WorkReport{r1, r2}, judgments, assignmentMap))
}

// ---------------------------------------------------------------------------
// workReportsEqual
// ---------------------------------------------------------------------------

// Same report compared to itself → equal.
func TestWorkReportsEqual_IdenticalReports(t *testing.T) {
	r := makeWorkReport(0xAA, 1)
	assert.True(t, workReportsEqual(r, r))
}

// Different PackageSpec.Hash → not equal.
func TestWorkReportsEqual_DifferentHash(t *testing.T) {
	a := makeWorkReport(0xAA, 1)
	b := makeWorkReport(0xBB, 1)
	assert.False(t, workReportsEqual(a, b))
}

// Different number of WorkResults → not equal.
func TestWorkReportsEqual_DifferentResultCount(t *testing.T) {
	a := makeWorkReport(0xAA, 1)
	b := makeWorkReport(0xAA, 1)
	b.Results = append(b.Results, types.WorkResult{})
	assert.False(t, workReportsEqual(a, b))
}

// Same result count but different content → not equal.
func TestWorkReportsEqual_DifferentResultContent(t *testing.T) {
	a := makeWorkReport(0xAA, 1)
	b := makeWorkReport(0xAA, 1)
	b.Results[0].ServiceID = 42
	assert.False(t, workReportsEqual(a, b))
}

// ---------------------------------------------------------------------------
// BuildAnnouncement (GP 17.9-17.11)
// ---------------------------------------------------------------------------

// Verify Ed25519 signature matches the reconstructed GP signing context.
func TestBuildAnnouncementSignsExpectedContext(t *testing.T) {
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

	// Reconstruct the expected signing context: XI ⌢ n ⌢ xn ⌢ H(H)
	var xnPayload types.ByteSequence
	for _, pair := range assignments {
		xnPayload = append(xnPayload, utilities.SerializeFixedLength(types.U64(pair.CoreID), 2)...)
		reportHash := hash.Blake2bHash(utilities.WorkReportSerialization(pair.Report))
		xnPayload = append(xnPayload, reportHash[:]...)
	}

	headerBytes, err := utilities.HeaderSerialization(blockchain.GetInstance().GetProcessingBlockPointer().GetHeader())
	require.NoError(t, err)
	headerHash := hash.Blake2bHash(headerBytes)

	// Build context without mutating the JamAnnounce constant.
	xi := []byte(types.JamAnnounce)
	expectedContext := make(types.ByteSequence, 0, len(xi)+1+len(xnPayload)+len(headerHash))
	expectedContext = append(expectedContext, xi...)
	expectedContext = append(expectedContext, byte(0))
	expectedContext = append(expectedContext, utilities.SerializeByteSequence(xnPayload)...)
	expectedContext = append(expectedContext, headerHash[:]...)

	assert.True(t, ed25519.Verify(pubKey, expectedContext, signature[:]))
}
