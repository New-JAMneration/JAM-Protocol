package ce

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

func TestHandleAuditAnnouncementFirstTranche(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	headerHash := types.OpaqueHash{}
	for i := range headerHash {
		headerHash[i] = byte(i + 1)
	}
	tranche := uint8(0)

	workReports := []WorkReportEntry{
		{
			CoreIndex:      types.CoreIndex(1),
			WorkReportHash: createTestWorkReportHash([]byte("test-work-report-1")),
		},
		{
			CoreIndex:      types.CoreIndex(2),
			WorkReportHash: createTestWorkReportHash([]byte("test-work-report-2")),
		},
	}

	announcement := &CE144Announcement{
		WorkReports: workReports,
		Signature:   createTestEd25519Signature([]byte("test-signature")),
	}
	evidence := &CE144Evidence{
		IsFirstTranche:     true,
		BandersnatchSig:    createTestBandersnatchSignature([]byte("test-bandersnatch")),
		SubsequentEvidence: nil,
	}

	announcementMsg, err := CreateAuditAnnouncement(headerHash, tranche, announcement, evidence)
	require.NoError(t, err, "CreateAuditAnnouncement")

	stream := newMockStream(announcementMsg)
	fakeBlockchain := SetupFakeBlockchain()

	require.NoError(t, HandleAuditAnnouncement(fakeBlockchain, stream), "HandleAuditAnnouncement")

	require.Empty(t, stream.w.Bytes(), "expected no response bytes")

	storedAnnouncement, err := GetAuditAnnouncement(fakeBlockchain, headerHash, tranche)
	require.NoError(t, err, "GetAuditAnnouncement")

	require.True(t, bytes.Equal(storedAnnouncement.HeaderHash[:], headerHash[:]), "header hash mismatch")
	require.Equal(t, tranche, storedAnnouncement.Tranche, "tranche mismatch")
	require.Len(t, storedAnnouncement.Announcement.WorkReports, len(workReports), "work reports count mismatch")
	require.True(t, storedAnnouncement.Evidence.IsFirstTranche, "stored evidence should be first tranche")
}

func TestHandleAuditAnnouncementSubsequentTranche(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	headerHash := types.OpaqueHash{}
	for i := range headerHash {
		headerHash[i] = byte(i + 10)
	}
	tranche := uint8(1)

	workReports := []WorkReportEntry{
		{
			CoreIndex:      types.CoreIndex(3),
			WorkReportHash: createTestWorkReportHash([]byte("test-work-report-3")),
		},
	}

	announcement := &CE144Announcement{
		WorkReports: workReports,
		Signature:   createTestEd25519Signature([]byte("test-signature-2")),
	}
	noShows := []NoShow{
		{
			ValidatorIndex:       types.ValidatorIndex(5),
			PreviousAnnouncement: []byte("previous-announcement-data"),
		},
	}

	subsequentEvidence := []SubsequentTrancheEvidence{
		{
			BandersnatchSig: createTestBandersnatchSignature([]byte("test-bandersnatch-2")),
			NoShows:         noShows,
		},
	}

	evidence := &CE144Evidence{
		IsFirstTranche:     false,
		BandersnatchSig:    types.BandersnatchVrfSignature{},
		SubsequentEvidence: subsequentEvidence,
	}

	announcementMsg, err := CreateAuditAnnouncement(headerHash, tranche, announcement, evidence)
	require.NoError(t, err, "CreateAuditAnnouncement")

	stream := newMockStream(announcementMsg)
	fakeBlockchain := SetupFakeBlockchain()

	require.NoError(t, HandleAuditAnnouncement(fakeBlockchain, stream), "HandleAuditAnnouncement")
	require.Empty(t, stream.w.Bytes(), "expected no response bytes")

	storedAnnouncement, err := GetAuditAnnouncement(fakeBlockchain, headerHash, tranche)
	require.NoError(t, err, "GetAuditAnnouncement")

	require.False(t, storedAnnouncement.Evidence.IsFirstTranche, "stored evidence should not be first tranche")
	require.Len(t, storedAnnouncement.Evidence.SubsequentEvidence, 1, "expected 1 subsequent evidence entry")
	require.Len(t, storedAnnouncement.Evidence.SubsequentEvidence[0].NoShows, 1, "expected 1 no-show")
}

func TestGetAllAuditAnnouncementsForHeader(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	fakeBlockchain := SetupFakeBlockchain()

	headerHash := types.OpaqueHash{}
	for i := range headerHash {
		headerHash[i] = byte(i + 20)
	}

	for tranche := uint8(0); tranche < 3; tranche++ {
		workReports := []WorkReportEntry{
			{
				CoreIndex:      types.CoreIndex(tranche + 1),
				WorkReportHash: createTestWorkReportHash([]byte("test-work-report")),
			},
		}

		announcement := &CE144Announcement{
			WorkReports: workReports,
			Signature:   createTestEd25519Signature([]byte("test-signature")),
		}

		var evidence *CE144Evidence
		if tranche == 0 {
			evidence = &CE144Evidence{
				IsFirstTranche:     true,
				BandersnatchSig:    createTestBandersnatchSignature([]byte("test-bandersnatch")),
				SubsequentEvidence: nil,
			}
		} else {
			evidence = &CE144Evidence{
				IsFirstTranche: false,
				SubsequentEvidence: []SubsequentTrancheEvidence{
					{
						BandersnatchSig: createTestBandersnatchSignature([]byte("test-bandersnatch")),
						NoShows:         []NoShow{},
					},
				},
			}
		}

		err := storeAuditAnnouncement(fakeBlockchain, headerHash, tranche, announcement, evidence)
		require.NoError(t, err, "storeAuditAnnouncement tranche %d", tranche)
	}

	announcements, err := GetAllAuditAnnouncementsForHeader(fakeBlockchain, headerHash)
	require.NoError(t, err, "GetAllAuditAnnouncementsForHeader")
	require.Len(t, announcements, 3, "expected 3 announcements")
}

// --------------------------------------------------------------------------
// Error: tranche 0 with subsequent-tranche evidence must fail
// --------------------------------------------------------------------------

func TestHandleAuditAnnouncementTranche0WithSubsequentEvidence(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	headerHash := types.OpaqueHash{}
	for i := range headerHash {
		headerHash[i] = byte(i + 5)
	}

	workReports := []WorkReportEntry{
		{CoreIndex: 1, WorkReportHash: createTestWorkReportHash([]byte("wr-tranche0-err"))},
	}
	announcement := &CE144Announcement{
		WorkReports: workReports,
		Signature:   createTestEd25519Signature([]byte("sig")),
	}
	evidence := &CE144Evidence{
		IsFirstTranche: false,
		SubsequentEvidence: []SubsequentTrancheEvidence{
			{
				BandersnatchSig: createTestBandersnatchSignature([]byte("bs")),
				NoShows:         []NoShow{},
			},
		},
	}

	err := validateAuditAnnouncement(headerHash, 0, announcement, evidence)
	require.Error(t, err, "expected error for tranche 0 with subsequent evidence")
}

// --------------------------------------------------------------------------
// Error: tranche > 0 with first-tranche evidence must fail
// --------------------------------------------------------------------------

func TestHandleAuditAnnouncementTranche1WithFirstEvidence(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	headerHash := types.OpaqueHash{}
	for i := range headerHash {
		headerHash[i] = byte(i + 6)
	}

	workReports := []WorkReportEntry{
		{CoreIndex: 2, WorkReportHash: createTestWorkReportHash([]byte("wr-tranche1-err"))},
	}
	announcement := &CE144Announcement{
		WorkReports: workReports,
		Signature:   createTestEd25519Signature([]byte("sig")),
	}
	evidence := &CE144Evidence{
		IsFirstTranche:  true,
		BandersnatchSig: createTestBandersnatchSignature([]byte("bs")),
	}

	err := validateAuditAnnouncement(headerHash, 1, announcement, evidence)
	require.Error(t, err, "expected error for tranche 1 with first-tranche evidence")
}

// --------------------------------------------------------------------------
// Error: subsequent evidence count != work reports count must fail
// --------------------------------------------------------------------------

func TestHandleAuditAnnouncementEvidenceCountMismatch(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	headerHash := types.OpaqueHash{}

	workReports := []WorkReportEntry{
		{CoreIndex: 3, WorkReportHash: createTestWorkReportHash([]byte("wr-a"))},
		{CoreIndex: 4, WorkReportHash: createTestWorkReportHash([]byte("wr-b"))},
	}
	announcement := &CE144Announcement{
		WorkReports: workReports,
		Signature:   createTestEd25519Signature([]byte("sig")),
	}
	evidence := &CE144Evidence{
		IsFirstTranche: false,
		SubsequentEvidence: []SubsequentTrancheEvidence{
			{
				BandersnatchSig: createTestBandersnatchSignature([]byte("bs1")),
				NoShows:         []NoShow{{ValidatorIndex: 1, PreviousAnnouncement: []byte("prev")}},
			},
		},
	}

	err := validateAuditAnnouncement(headerHash, 1, announcement, evidence)
	require.Error(t, err, "expected error for evidence/work-report count mismatch")
}

// --------------------------------------------------------------------------
// Error: empty announcement (0 work reports) must fail
// --------------------------------------------------------------------------

func TestHandleAuditAnnouncementEmptyWorkReports(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	headerHash := types.OpaqueHash{}

	announcement := &CE144Announcement{
		WorkReports: []WorkReportEntry{},
		Signature:   createTestEd25519Signature([]byte("sig")),
	}
	evidence := &CE144Evidence{
		IsFirstTranche:  true,
		BandersnatchSig: createTestBandersnatchSignature([]byte("bs")),
	}

	err := validateAuditAnnouncement(headerHash, 0, announcement, evidence)
	require.Error(t, err, "expected error for empty work reports")
}

// --------------------------------------------------------------------------
// Error: no-show with empty PreviousAnnouncement must fail
// --------------------------------------------------------------------------

func TestHandleAuditAnnouncementNoShowEmptyPreviousAnnouncement(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	headerHash := types.OpaqueHash{}

	workReports := []WorkReportEntry{
		{CoreIndex: 5, WorkReportHash: createTestWorkReportHash([]byte("wr-noshow"))},
	}
	announcement := &CE144Announcement{
		WorkReports: workReports,
		Signature:   createTestEd25519Signature([]byte("sig")),
	}
	evidence := &CE144Evidence{
		IsFirstTranche: false,
		SubsequentEvidence: []SubsequentTrancheEvidence{
			{
				BandersnatchSig: createTestBandersnatchSignature([]byte("bs")),
				NoShows: []NoShow{
					{ValidatorIndex: 1, PreviousAnnouncement: []byte{}},
				},
			},
		},
	}

	err := validateAuditAnnouncement(headerHash, 1, announcement, evidence)
	require.Error(t, err, "expected error for no-show with empty PreviousAnnouncement")
}

// --------------------------------------------------------------------------
// CE144Payload Encode/Decode round-trip: first tranche
// --------------------------------------------------------------------------

func TestCE144PayloadEncodeDecodeFirstTranche(t *testing.T) {
	headerHash := types.OpaqueHash{}
	for i := range headerHash {
		headerHash[i] = byte(i + 1)
	}

	payload := &CE144Payload{
		HeaderHash: headerHash,
		Tranche:    0,
		Announcement: CE144Announcement{
			WorkReports: []WorkReportEntry{
				{CoreIndex: 10, WorkReportHash: createTestWorkReportHash([]byte("rt1"))},
			},
			Signature: createTestEd25519Signature([]byte("rt-sig")),
		},
		Evidence: CE144Evidence{
			IsFirstTranche:  true,
			BandersnatchSig: createTestBandersnatchSignature([]byte("rt-bs")),
		},
	}

	encoded, err := payload.Encode()
	require.NoError(t, err, "Encode")

	decoded := &CE144Payload{}
	require.NoError(t, decoded.Decode(encoded), "Decode")

	require.True(t, bytes.Equal(decoded.HeaderHash[:], headerHash[:]), "header hash mismatch after round-trip")
	require.Equal(t, uint8(0), decoded.Tranche, "tranche mismatch")
	require.Len(t, decoded.Announcement.WorkReports, 1, "work report count mismatch")
	require.True(t, decoded.Evidence.IsFirstTranche, "evidence should be first-tranche after round-trip")
}

// --------------------------------------------------------------------------
// CE144Payload Encode/Decode round-trip: subsequent tranche
// --------------------------------------------------------------------------

func TestCE144PayloadEncodeDecodeSubsequentTranche(t *testing.T) {
	headerHash := types.OpaqueHash{}

	payload := &CE144Payload{
		HeaderHash: headerHash,
		Tranche:    2,
		Announcement: CE144Announcement{
			WorkReports: []WorkReportEntry{
				{CoreIndex: 7, WorkReportHash: createTestWorkReportHash([]byte("rt2"))},
			},
			Signature: createTestEd25519Signature([]byte("rt2-sig")),
		},
		Evidence: CE144Evidence{
			IsFirstTranche: false,
			SubsequentEvidence: []SubsequentTrancheEvidence{
				{
					BandersnatchSig: createTestBandersnatchSignature([]byte("rt2-bs")),
					NoShows: []NoShow{
						{ValidatorIndex: 3, PreviousAnnouncement: []byte("prev-ann")},
					},
				},
			},
		},
	}

	encoded, err := payload.Encode()
	require.NoError(t, err, "Encode")

	decoded := &CE144Payload{}
	require.NoError(t, decoded.Decode(encoded), "Decode")

	require.Equal(t, uint8(2), decoded.Tranche, "tranche mismatch")
	require.False(t, decoded.Evidence.IsFirstTranche, "evidence should be subsequent-tranche after round-trip")
	require.Len(t, decoded.Evidence.SubsequentEvidence, 1, "subsequent evidence count mismatch")
	require.Len(t, decoded.Evidence.SubsequentEvidence[0].NoShows, 1, "no-shows count mismatch")

	noShow := decoded.Evidence.SubsequentEvidence[0].NoShows[0]
	require.Equal(t, types.ValidatorIndex(3), noShow.ValidatorIndex, "no-show validator index mismatch")
	require.True(t, bytes.Equal(noShow.PreviousAnnouncement, []byte("prev-ann")), "no-show previous announcement mismatch")
}

// --------------------------------------------------------------------------
// Stage 3: Truncated bytes must fail with descriptive error
// --------------------------------------------------------------------------

func TestCE144Payload_Decode_TruncatedHeader(t *testing.T) {
	// Less than minimum msg1 size (32 header + 1 tranche + 1 count + 64 sig = 98)
	err := (&CE144Payload{}).Decode(make([]byte, 10))
	require.Error(t, err)
}

func TestCE144Payload_Decode_TruncatedEvidence(t *testing.T) {
	// Valid msg1 for tranche 0 but evidence truncated (need 96 bytes Bandersnatch sig)
	payload := &CE144Payload{
		HeaderHash: types.OpaqueHash{},
		Tranche:    0,
		Announcement: CE144Announcement{
			WorkReports: []WorkReportEntry{{CoreIndex: 1, WorkReportHash: createTestWorkReportHash([]byte("trunc"))}},
			Signature:   createTestEd25519Signature([]byte("sig")),
		},
		Evidence: CE144Evidence{IsFirstTranche: true, BandersnatchSig: createTestBandersnatchSignature([]byte("bs"))},
	}
	encoded, err := payload.Encode()
	require.NoError(t, err)

	// Chop off last 50 bytes from evidence
	err = (&CE144Payload{}).Decode(encoded[:len(encoded)-50])
	require.Error(t, err)
}

// --------------------------------------------------------------------------
// Stage 3: Tiny mode cores full — roundtrip with CoresCount work reports
// --------------------------------------------------------------------------

func TestCE144Payload_TinyCoresFull(t *testing.T) {
	var workReports []WorkReportEntry
	for i := 0; i < types.CoresCount; i++ {
		workReports = append(workReports, WorkReportEntry{
			CoreIndex:      types.CoreIndex(i),
			WorkReportHash: createTestWorkReportHash([]byte{byte(i)}),
		})
	}

	payload := &CE144Payload{
		HeaderHash: types.OpaqueHash{1},
		Tranche:    0,
		Announcement: CE144Announcement{
			WorkReports: workReports,
			Signature:   createTestEd25519Signature([]byte("full")),
		},
		Evidence: CE144Evidence{IsFirstTranche: true, BandersnatchSig: createTestBandersnatchSignature([]byte("full-bs"))},
	}

	encoded, err := payload.Encode()
	require.NoError(t, err)

	decoded := &CE144Payload{}
	require.NoError(t, decoded.Decode(encoded))
	require.Len(t, decoded.Announcement.WorkReports, types.CoresCount)
}

// Helper functions for creating test data

func createTestWorkReportHash(data []byte) types.WorkReportHash {
	hash := sha256.Sum256(data)
	var workReportHash types.WorkReportHash
	copy(workReportHash[:], hash[:])
	return workReportHash
}

func createTestEd25519Signature(data []byte) types.Ed25519Signature {
	var signature types.Ed25519Signature
	hash := sha256.Sum256(data)
	copy(signature[:32], hash[:])
	copy(signature[32:], hash[:])
	return signature
}

func createTestBandersnatchSignature(data []byte) types.BandersnatchVrfSignature {
	var signature types.BandersnatchVrfSignature
	hash := sha256.Sum256(data)
	for i := 0; i < 96; i += 32 {
		remaining := 96 - i
		if remaining >= 32 {
			copy(signature[i:i+32], hash[:])
		} else {
			copy(signature[i:i+remaining], hash[:remaining])
		}
	}
	return signature
}
