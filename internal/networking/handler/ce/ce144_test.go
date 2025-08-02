package ce

import (
	"bytes"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandleAuditAnnouncementFirstTranche(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

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
	if err != nil {
		t.Fatalf("Failed to create audit announcement: %v", err)
	}

	fullMessage := append(announcementMsg, []byte("FIN")...)

	stream := newMockStream(fullMessage)

	fakeBlockchain := SetupFakeBlockchain()

	err = HandleAuditAnnouncement(fakeBlockchain, stream)
	if err != nil {
		t.Fatalf("HandleAuditAnnouncement failed: %v", err)
	}

	response := stream.w.Bytes()
	if string(response) != "FIN" {
		t.Errorf("Expected FIN response, got: %s", string(response))
	}
	storedAnnouncement, err := GetAuditAnnouncement(headerHash, tranche)
	if err != nil {
		t.Fatalf("Failed to retrieve stored announcement: %v", err)
	}

	if !bytes.Equal(storedAnnouncement.HeaderHash[:], headerHash[:]) {
		t.Error("Stored header hash doesn't match original")
	}

	if storedAnnouncement.Tranche != tranche {
		t.Errorf("Stored tranche doesn't match original: expected %d, got %d", tranche, storedAnnouncement.Tranche)
	}

	if len(storedAnnouncement.Announcement.WorkReports) != len(workReports) {
		t.Errorf("Stored work reports count doesn't match: expected %d, got %d",
			len(workReports), len(storedAnnouncement.Announcement.WorkReports))
	}

	if !storedAnnouncement.Evidence.IsFirstTranche {
		t.Error("Stored evidence should be first tranche")
	}
}

func TestHandleAuditAnnouncementSubsequentTranche(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

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
	if err != nil {
		t.Fatalf("Failed to create audit announcement: %v", err)
	}

	fullMessage := append(announcementMsg, []byte("FIN")...)

	stream := newMockStream(fullMessage)

	fakeBlockchain := SetupFakeBlockchain()
	err = HandleAuditAnnouncement(fakeBlockchain, stream)
	if err != nil {
		t.Fatalf("HandleAuditAnnouncement failed: %v", err)
	}

	response := stream.w.Bytes()
	if string(response) != "FIN" {
		t.Errorf("Expected FIN response, got: %s", string(response))
	}

	storedAnnouncement, err := GetAuditAnnouncement(headerHash, tranche)
	if err != nil {
		t.Fatalf("Failed to retrieve stored announcement: %v", err)
	}

	if storedAnnouncement.Evidence.IsFirstTranche {
		t.Error("Stored evidence should not be first tranche")
	}

	if len(storedAnnouncement.Evidence.SubsequentEvidence) != 1 {
		t.Errorf("Expected 1 subsequent evidence entry, got %d", len(storedAnnouncement.Evidence.SubsequentEvidence))
	}

	if len(storedAnnouncement.Evidence.SubsequentEvidence[0].NoShows) != 1 {
		t.Errorf("Expected 1 no-show, got %d", len(storedAnnouncement.Evidence.SubsequentEvidence[0].NoShows))
	}
}

func TestGetAllAuditAnnouncementsForHeader(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	headerHash := types.OpaqueHash{}
	for i := range headerHash {
		headerHash[i] = byte(i + 20)
	}

	// Store multiple announcements for the same header
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

		err := storeAuditAnnouncement(headerHash, tranche, announcement, evidence)
		if err != nil {
			t.Fatalf("Failed to store announcement for tranche %d: %v", tranche, err)
		}
	}

	// Retrieve all announcements
	announcements, err := GetAllAuditAnnouncementsForHeader(headerHash)
	if err != nil {
		t.Fatalf("Failed to retrieve announcements: %v", err)
	}

	if len(announcements) != 3 {
		t.Errorf("Expected 3 announcements, got %d", len(announcements))
	}
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
	// Fill the 96-byte signature with repeated hash data
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
