package auditing

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// Stage 4: AuditMessageBus — CE144/CE145 channel push/drain
// ===========================================================================

// Multiple announcements from different validators merge into assignment map.
func TestStage4_CE144_PushAndDrain(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)
	h1 := makeWPHash(0x01)
	h2 := makeWPHash(0x02)

	bus.OnAuditAnnouncementReceived(CE144Announcement{ValidatorIndex: 10, WorkReports: []types.WorkPackageHash{h1, h2}})
	bus.OnAuditAnnouncementReceived(CE144Announcement{ValidatorIndex: 20, WorkReports: []types.WorkPackageHash{h1}})

	am := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	am = SyncAssignmentMapFromBus(bus, am)

	assert.ElementsMatch(t, am[h1], []types.ValidatorIndex{10, 20})
	assert.ElementsMatch(t, am[h2], []types.ValidatorIndex{10})
}

// Same validator announcing twice → deduplicated to single entry.
func TestStage4_CE144_DeduplicateValidator(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)
	h := makeWPHash(0x01)

	bus.OnAuditAnnouncementReceived(CE144Announcement{ValidatorIndex: 10, WorkReports: []types.WorkPackageHash{h}})
	bus.OnAuditAnnouncementReceived(CE144Announcement{ValidatorIndex: 10, WorkReports: []types.WorkPackageHash{h}})

	am := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	am = SyncAssignmentMapFromBus(bus, am)

	assert.Equal(t, []types.ValidatorIndex{10}, am[h])
}

// Draining across multiple calls accumulates, not overwrites.
func TestStage4_CE144_MultiDrain(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)
	h := makeWPHash(0x01)

	bus.OnAuditAnnouncementReceived(CE144Announcement{ValidatorIndex: 1, WorkReports: []types.WorkPackageHash{h}})
	am := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	am = SyncAssignmentMapFromBus(bus, am)
	assert.Equal(t, []types.ValidatorIndex{1}, am[h])

	bus.OnAuditAnnouncementReceived(CE144Announcement{ValidatorIndex: 2, WorkReports: []types.WorkPackageHash{h}})
	am = SyncAssignmentMapFromBus(bus, am)
	assert.Equal(t, []types.ValidatorIndex{1, 2}, am[h])
}

// Only valid judgments are merged into positiveJudgers; invalid ones skipped.
func TestStage4_CE145_PushAndDrain(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)
	h := makeWPHash(0xAA)

	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h, ValidatorIndex: 1, IsValid: true})
	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h, ValidatorIndex: 2, IsValid: false})
	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h, ValidatorIndex: 3, IsValid: true})

	pj := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)
	pj = SyncPositiveJudgersFromBus(bus, pj)

	assert.Len(t, pj[h], 2)
	assert.True(t, pj[h][1])
	assert.True(t, pj[h][3])
}

// Draining across multiple calls accumulates positive judgers.
func TestStage4_CE145_MultiDrain(t *testing.T) {
	bus := NewAuditMessageBusWithSize(16)
	h := makeWPHash(0xBB)

	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h, ValidatorIndex: 1, IsValid: true})
	pj := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)
	pj = SyncPositiveJudgersFromBus(bus, pj)
	assert.Len(t, pj[h], 1)

	bus.OnJudgmentReceived(CE145Judgment{WorkReportHash: h, ValidatorIndex: 2, IsValid: true})
	pj = SyncPositiveJudgersFromBus(bus, pj)
	assert.Len(t, pj[h], 2)
}

// Concurrent goroutines pushing to the bus must not lose messages.
func TestStage4_ConcurrentPush(t *testing.T) {
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
	assert.Equal(t, pushers*msgsPerPusher, len(pj[h]))
}

// ===========================================================================
// Stage 4 TODO: CE144/CE145 stream round-trip (send → receive → store → retrieve)
// ===========================================================================

func TestStage4_CE144_StreamRoundTrip(t *testing.T) {
	t.Skip("TODO: requires #925 merge — build CE144Payload → HandleAuditAnnouncement_Send → HandleAuditAnnouncement → GetAuditAnnouncement → compare")
}

func TestStage4_CE145_StreamRoundTrip(t *testing.T) {
	t.Skip("TODO: requires #925 merge — build CE145Payload → HandleJudgmentAnnouncement_Send → HandleJudgmentAnnouncement → GetJudgment → compare")
}

// ===========================================================================
// Stage 4 TODO: SingleNodeAuditingAndPublish end-to-end
// ===========================================================================

func TestStage4_SingleNodeAuditingAndPublish_HappyPath(t *testing.T) {
	t.Skip("TODO: mock blockchain singleton + ρ + BundleFetcher + bus → run full loop → verify IsBlockAudited")
}

func TestStage4_SingleNodeAuditingAndPublish_MultiTranche(t *testing.T) {
	t.Skip("TODO: tranche 0 insufficient judgers → tranche 1 re-assignment → eventually audited")
}

// ===========================================================================
// Stage 4 TODO: CE144 Send round-trip (requires stream mock with framing)
// ===========================================================================

func TestStage4_CE144_SendReceive(t *testing.T) {
	t.Skip("TODO: HandleAuditAnnouncement_Send writes to mockStream → parse output bytes → verify payload fields")
}

func TestStage4_CE145_SendReceive(t *testing.T) {
	t.Skip("TODO: HandleJudgmentAnnouncement_Send writes to mockStream → parse output bytes → verify payload fields")
}

// Helpers (makeWPHash, makeAuditReport, etc.) defined in auditing_test.go — same package.
