package auditing

import (
	"context"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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

// ---------------------------------------------------------------------------
// workReportsEqual
// ---------------------------------------------------------------------------

func TestWorkReportsEqual_IdenticalReports(t *testing.T) {
	r := makeWorkReport(0xAA, 1)
	assert.True(t, workReportsEqual(r, r), "identical reports should be equal")
}

func TestWorkReportsEqual_DifferentHash(t *testing.T) {
	a := makeWorkReport(0xAA, 1)
	b := makeWorkReport(0xBB, 1)
	assert.False(t, workReportsEqual(a, b), "reports with different package hash should not be equal")
}

func TestWorkReportsEqual_DifferentCoreIndex(t *testing.T) {
	a := makeWorkReport(0xAA, 1)
	b := makeWorkReport(0xAA, 2)
	assert.False(t, workReportsEqual(a, b), "reports with different core index should not be equal")
}

func TestWorkReportsEqual_DifferentAuthOutput(t *testing.T) {
	a := makeWorkReport(0xAA, 1)
	b := makeWorkReport(0xAA, 1)
	b.AuthOutput = types.ByteSequence{0x01}
	assert.False(t, workReportsEqual(a, b), "reports with different auth output should not be equal")
}

// ---------------------------------------------------------------------------
// GetJudgement — mock BundleFetcher
// ---------------------------------------------------------------------------

type mockBundleFetcher struct {
	bundle []byte
	err    error
}

func (m *mockBundleFetcher) FetchBundle(_ types.WorkReport) ([]byte, error) {
	return m.bundle, m.err
}

func TestGetJudgement_FetchError(t *testing.T) {
	original := DefaultBundleFetcher
	defer func() { DefaultBundleFetcher = original }()

	DefaultBundleFetcher = &mockBundleFetcher{
		bundle: nil,
		err:    assert.AnError,
	}

	ar := makeAuditReport(0x01, 0, 0, false)
	assert.False(t, GetJudgement(ar), "fetch error should return false")
}

// ---------------------------------------------------------------------------
// ClassifyJudgments
// ---------------------------------------------------------------------------

func TestClassifyJudgments(t *testing.T) {
	report := makeWorkReport(0xCC, 3)

	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: true},
		{Report: report, ValidatorID: 2, AuditResult: false},
		{Report: report, ValidatorID: 3, AuditResult: true},
		{Report: makeWorkReport(0xDD, 4), ValidatorID: 5, AuditResult: true},
	}

	pos, neg := ClassifyJudgments(report, judgments)
	assert.Len(t, pos, 2, "should have 2 positive judgers")
	assert.Len(t, neg, 1, "should have 1 negative judger")
	assert.True(t, pos[1])
	assert.True(t, pos[3])
	assert.True(t, neg[2])
}

// ---------------------------------------------------------------------------
// IsWorkReportAudited
// ---------------------------------------------------------------------------

func TestIsWorkReportAudited_AllAssignedPositive(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: true},
		{Report: report, ValidatorID: 2, AuditResult: true},
	}

	assert.True(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1, 2}))
}

func TestIsWorkReportAudited_HasNegative(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: true},
		{Report: report, ValidatorID: 2, AuditResult: false},
	}

	assert.False(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1, 2}))
}

func TestIsWorkReportAudited_MissingAssigned(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: true},
	}

	// validator 2 is assigned but has not judged → should not pass
	assert.False(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1, 2}))
}

// ---------------------------------------------------------------------------
// IsBlockAudited
// ---------------------------------------------------------------------------

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
// UpdateAssignmentMap
// ---------------------------------------------------------------------------

func TestUpdateAssignmentMap(t *testing.T) {
	a0 := []types.AuditReport{
		makeAuditReport(0x01, 0, 5, true),
		makeAuditReport(0x02, 1, 5, false),
	}
	am := make(types.AssignmentMap)
	am = UpdateAssignmentMap(a0, am)

	assert.Contains(t, am[a0[0].Report.PackageSpec.Hash], types.ValidatorIndex(5))
	assert.Contains(t, am[a0[1].Report.PackageSpec.Hash], types.ValidatorIndex(5))
}

// ---------------------------------------------------------------------------
// UpdatePositiveJudgersFromAudit
// ---------------------------------------------------------------------------

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

	assert.Len(t, pj[h1], 2, "two positive judgers for h1")
	assert.True(t, pj[h1][1])
	assert.True(t, pj[h1][3])
	_, exists := pj[h2]
	assert.False(t, exists, "h2 should have no positive judgers (only negative)")
}

// ---------------------------------------------------------------------------
// FilterJudgments
// ---------------------------------------------------------------------------

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
// WaitNextTranche
// ---------------------------------------------------------------------------

func TestWaitNextTranche_NormalExpiry(t *testing.T) {
	// slotStart = now, tranche 0 → deadline = now + 8s.
	// But we set slotStart to (now - 7.9s) so deadline is ~0.1s from now.
	slotStart := time.Now().Add(-7900 * time.Millisecond)
	ctx := context.Background()

	start := time.Now()
	err := WaitNextTranche(0, slotStart, ctx)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 500*time.Millisecond, "should complete quickly since deadline is near")
}

func TestWaitNextTranche_ContextCancel(t *testing.T) {
	slotStart := time.Now()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := WaitNextTranche(0, slotStart, ctx)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, elapsed, 1*time.Second, "should be cancelled quickly, not wait full 8s")
}

func TestWaitNextTranche_AlreadyPastDeadline(t *testing.T) {
	slotStart := time.Now().Add(-20 * time.Second)
	ctx := context.Background()

	start := time.Now()
	err := WaitNextTranche(0, slotStart, ctx)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 100*time.Millisecond, "deadline already past, should return immediately")
}

// ---------------------------------------------------------------------------
// AuditMessageBus — push + drain
// ---------------------------------------------------------------------------

func TestAuditMessageBus_CE144_PushAndDrain(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)

	h1 := makeWPHash(0x01)
	h2 := makeWPHash(0x02)

	bus.OnAuditAnnouncementReceived(CE144Announcement{
		ValidatorIndex: 10,
		WorkReports:    []types.WorkPackageHash{h1, h2},
	})
	bus.OnAuditAnnouncementReceived(CE144Announcement{
		ValidatorIndex: 20,
		WorkReports:    []types.WorkPackageHash{h1},
	})

	am := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	am = SyncAssignmentMapFromBus(bus, am)

	assert.ElementsMatch(t, am[h1], []types.ValidatorIndex{10, 20})
	assert.ElementsMatch(t, am[h2], []types.ValidatorIndex{10})
}

func TestAuditMessageBus_CE144_DeduplicateValidator(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)
	h := makeWPHash(0x01)

	// Same validator announces the same report twice.
	bus.OnAuditAnnouncementReceived(CE144Announcement{
		ValidatorIndex: 10,
		WorkReports:    []types.WorkPackageHash{h},
	})
	bus.OnAuditAnnouncementReceived(CE144Announcement{
		ValidatorIndex: 10,
		WorkReports:    []types.WorkPackageHash{h},
	})

	am := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	am = SyncAssignmentMapFromBus(bus, am)

	assert.Equal(t, []types.ValidatorIndex{10}, am[h],
		"duplicate validator should appear only once")
}

func TestAuditMessageBus_CE145_PushAndDrain(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)

	h1 := makeWPHash(0xAA)

	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h1, ValidatorIndex: 1, IsValid: true})
	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h1, ValidatorIndex: 2, IsValid: false})
	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h1, ValidatorIndex: 3, IsValid: true})

	pj := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)
	pj = SyncPositiveJudgersFromBus(bus, pj)

	assert.Len(t, pj[h1], 2, "only valid judgments merged")
	assert.True(t, pj[h1][1])
	assert.True(t, pj[h1][3])
}

func TestAuditMessageBus_DrainEmptyChannel(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)

	am := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	am = SyncAssignmentMapFromBus(bus, am)
	assert.Empty(t, am, "draining empty channel should return empty map")

	pj := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)
	pj = SyncPositiveJudgersFromBus(bus, pj)
	assert.Empty(t, pj, "draining empty channel should return empty map")
}

func TestAuditMessageBus_ChannelFullDrops(t *testing.T) {
	bus := NewAuditMessageBusWithSize(2) // very small buffer

	for i := 0; i < 5; i++ {
		bus.OnAuditAnnouncementReceived(CE144Announcement{ValidatorIndex: types.ValidatorIndex(i)})
	}

	// Only 2 should have been buffered, the other 3 are dropped.
	am := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	count := 0
	for {
		select {
		case <-bus.announcementCh:
			count++
		default:
			goto done
		}
	}
done:
	_ = am
	assert.Equal(t, 2, count, "channel of size 2 should only hold 2 messages")
}

// ---------------------------------------------------------------------------
// GetJudgement — execution failure (invalid bundle bytes)
// ---------------------------------------------------------------------------

func TestGetJudgement_ProcessError(t *testing.T) {
	original := DefaultBundleFetcher
	defer func() { DefaultBundleFetcher = original }()

	// FetchBundle succeeds but returns garbage — Process() should fail.
	DefaultBundleFetcher = &mockBundleFetcher{
		bundle: []byte{0xDE, 0xAD},
		err:    nil,
	}

	ar := makeAuditReport(0x01, 0, 0, false)
	assert.False(t, GetJudgement(ar), "invalid bundle should cause Process() to fail → false")
}

// ---------------------------------------------------------------------------
// ClassifyJudgments — edge cases
// ---------------------------------------------------------------------------

func TestClassifyJudgments_Empty(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	pos, neg := ClassifyJudgments(report, nil)
	assert.Empty(t, pos)
	assert.Empty(t, neg)
}

func TestClassifyJudgments_AllPositive(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: true},
		{Report: report, ValidatorID: 2, AuditResult: true},
	}
	pos, neg := ClassifyJudgments(report, judgments)
	assert.Len(t, pos, 2)
	assert.Empty(t, neg)
}

func TestClassifyJudgments_AllNegative(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	judgments := []types.AuditReport{
		{Report: report, ValidatorID: 1, AuditResult: false},
		{Report: report, ValidatorID: 2, AuditResult: false},
	}
	pos, neg := ClassifyJudgments(report, judgments)
	assert.Empty(t, pos)
	assert.Len(t, neg, 2)
}

// ---------------------------------------------------------------------------
// IsWorkReportAudited — supermajority path
// ---------------------------------------------------------------------------

func TestIsWorkReportAudited_Supermajority(t *testing.T) {
	report := makeWorkReport(0xAA, 0)

	// Generate ValidatorsSuperMajority positive judgments from non-assigned validators.
	// Even if not all assigned confirmed, supermajority should pass.
	var judgments []types.AuditReport
	for i := 0; i < types.ValidatorsSuperMajority; i++ {
		judgments = append(judgments, types.AuditReport{
			Report:      report,
			ValidatorID: types.ValidatorIndex(100 + i),
			AuditResult: true,
		})
	}

	// assigned = [1, 2] but they haven't judged; supermajority still passes
	assert.True(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1, 2}))
}

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

	// Not enough for supermajority, and assigned [1] hasn't judged
	assert.False(t, IsWorkReportAudited(report, judgments, []types.ValidatorIndex{1}))
}

// ---------------------------------------------------------------------------
// IsBlockAudited — edge cases
// ---------------------------------------------------------------------------

func TestIsBlockAudited_EmptyReports(t *testing.T) {
	am := map[types.WorkPackageHash][]types.ValidatorIndex{}
	assert.True(t, IsBlockAudited(nil, nil, am), "no reports → trivially audited")
}

// ---------------------------------------------------------------------------
// GetAssignedValidators
// ---------------------------------------------------------------------------

func TestGetAssignedValidators_Found(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	am := types.AssignmentMap{
		report.PackageSpec.Hash: {1, 2, 3},
	}
	result := GetAssignedValidators(report, am)
	assert.Equal(t, []types.ValidatorIndex{1, 2, 3}, result)
}

func TestGetAssignedValidators_NotFound(t *testing.T) {
	report := makeWorkReport(0xAA, 0)
	am := types.AssignmentMap{}
	result := GetAssignedValidators(report, am)
	assert.Empty(t, result)
}

// ---------------------------------------------------------------------------
// UpdateAssignmentMap — accumulation
// ---------------------------------------------------------------------------

func TestUpdateAssignmentMap_Accumulate(t *testing.T) {
	h := makeWPHash(0x01)
	r := types.WorkReport{PackageSpec: types.WorkPackageSpec{Hash: h}, Results: []types.WorkResult{{}}}

	am := make(types.AssignmentMap)
	am = UpdateAssignmentMap([]types.AuditReport{
		{Report: r, ValidatorID: 1},
	}, am)
	am = UpdateAssignmentMap([]types.AuditReport{
		{Report: r, ValidatorID: 2},
	}, am)

	assert.Equal(t, []types.ValidatorIndex{1, 2}, am[h])
}

// ---------------------------------------------------------------------------
// WaitNextTranche — higher tranche
// ---------------------------------------------------------------------------

func TestWaitNextTranche_Tranche2(t *testing.T) {
	// tranche=2, deadline = slotStart + 3*8s = slotStart + 24s.
	// Set slotStart to now-23.9s so deadline is ~100ms from now.
	slotStart := time.Now().Add(-23900 * time.Millisecond)
	ctx := context.Background()

	start := time.Now()
	err := WaitNextTranche(2, slotStart, ctx)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 500*time.Millisecond)
}

func TestWaitNextTranche_ContextDeadlineExceeded(t *testing.T) {
	slotStart := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := WaitNextTranche(0, slotStart, ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// ---------------------------------------------------------------------------
// AuditMessageBus — multi-drain accumulation
// ---------------------------------------------------------------------------

func TestAuditMessageBus_CE144_MultiDrain(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)
	h := makeWPHash(0x01)

	// First batch
	bus.OnAuditAnnouncementReceived(CE144Announcement{ValidatorIndex: 1, WorkReports: []types.WorkPackageHash{h}})
	am := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	am = SyncAssignmentMapFromBus(bus, am)
	assert.Equal(t, []types.ValidatorIndex{1}, am[h])

	// Second batch — should accumulate, not overwrite
	bus.OnAuditAnnouncementReceived(CE144Announcement{ValidatorIndex: 2, WorkReports: []types.WorkPackageHash{h}})
	am = SyncAssignmentMapFromBus(bus, am)
	assert.Equal(t, []types.ValidatorIndex{1, 2}, am[h])
}

func TestAuditMessageBus_CE145_MultiDrain(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)
	h := makeWPHash(0xBB)

	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h, ValidatorIndex: 1, IsValid: true})
	pj := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)
	pj = SyncPositiveJudgersFromBus(bus, pj)
	assert.Len(t, pj[h], 1)

	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h, ValidatorIndex: 2, IsValid: true})
	pj = SyncPositiveJudgersFromBus(bus, pj)
	assert.Len(t, pj[h], 2, "should accumulate across drains")
}

func TestAuditMessageBus_CE145_IgnoreInvalid(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)
	h := makeWPHash(0xCC)

	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h, ValidatorIndex: 1, IsValid: false})
	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h, ValidatorIndex: 2, IsValid: false})

	pj := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)
	pj = SyncPositiveJudgersFromBus(bus, pj)
	assert.Empty(t, pj, "all-invalid judgments should not create entries")
}

// ---------------------------------------------------------------------------
// AuditMessageBus — concurrent push safety
// ---------------------------------------------------------------------------

func TestAuditMessageBus_ConcurrentPush(t *testing.T) {
	bus := NewAuditMessageBusWithSize(256)
	h := makeWPHash(0x01)

	done := make(chan struct{})
	const pushers = 10
	const msgsPerPusher = 20

	for i := 0; i < pushers; i++ {
		go func(vid int) {
			defer func() { done <- struct{}{} }()
			for j := 0; j < msgsPerPusher; j++ {
				bus.OnJudgmentReceived(CE145Judgment{
					WorkReportHash: h,
					ValidatorIndex: types.ValidatorIndex(vid*100 + j),
					IsValid:        true,
				})
			}
		}(i)
	}

	for i := 0; i < pushers; i++ {
		<-done
	}

	pj := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)
	pj = SyncPositiveJudgersFromBus(bus, pj)
	// With buffer=256, all 200 messages should fit.
	assert.Equal(t, pushers*msgsPerPusher, len(pj[h]), "all concurrent messages should be drained")
}

// ---------------------------------------------------------------------------
// workReportsEqual — result list differences
// ---------------------------------------------------------------------------

func TestWorkReportsEqual_DifferentResultCount(t *testing.T) {
	a := makeWorkReport(0xAA, 1)
	b := makeWorkReport(0xAA, 1)
	b.Results = append(b.Results, types.WorkResult{})
	assert.False(t, workReportsEqual(a, b), "different number of results should not be equal")
}

func TestWorkReportsEqual_DifferentResultContent(t *testing.T) {
	a := makeWorkReport(0xAA, 1)
	b := makeWorkReport(0xAA, 1)
	b.Results[0].ServiceId = 42
	assert.False(t, workReportsEqual(a, b), "different result service id should not be equal")
}
