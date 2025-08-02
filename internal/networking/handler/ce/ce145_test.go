package ce

import (
	"bytes"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandleJudgmentAnnouncementValid(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	epochIndex := types.U32(12345)
	validatorIndex := types.ValidatorIndex(67)
	validity := uint8(1) // Valid
	workReportHash := createTestWorkReportHash([]byte("test-work-report-valid"))
	signature := createTestEd25519Signature([]byte("test-signature-valid"))

	announcementMsg, err := CreateJudgmentAnnouncement(epochIndex, validatorIndex, validity, workReportHash, signature)
	if err != nil {
		t.Fatalf("Failed to create judgment announcement: %v", err)
	}

	fullMessage := append(announcementMsg, []byte("FIN")...)

	stream := newMockStream(fullMessage)

	fakeBlockchain := SetupFakeBlockchain()

	err = HandleJudgmentAnnouncement(fakeBlockchain, stream)
	if err != nil {
		t.Fatalf("HandleJudgmentAnnouncement failed: %v", err)
	}

	response := stream.w.Bytes()
	if string(response) != "FIN" {
		t.Errorf("Expected FIN response, got: %s", string(response))
	}

	storedJudgment, err := GetJudgment(workReportHash, epochIndex, validatorIndex)
	if err != nil {
		t.Fatalf("Failed to retrieve stored judgment: %v", err)
	}

	if storedJudgment.EpochIndex != epochIndex {
		t.Errorf("Stored epoch index doesn't match original: expected %d, got %d", epochIndex, storedJudgment.EpochIndex)
	}

	if storedJudgment.ValidatorIndex != validatorIndex {
		t.Errorf("Stored validator index doesn't match original: expected %d, got %d", validatorIndex, storedJudgment.ValidatorIndex)
	}

	if storedJudgment.Validity != validity {
		t.Errorf("Stored validity doesn't match original: expected %d, got %d", validity, storedJudgment.Validity)
	}

	if !bytes.Equal(storedJudgment.WorkReportHash[:], workReportHash[:]) {
		t.Error("Stored work report hash doesn't match original")
	}

	if !bytes.Equal(storedJudgment.Signature[:], signature[:]) {
		t.Error("Stored signature doesn't match original")
	}

	if !storedJudgment.IsValid() {
		t.Error("Judgment should be valid")
	}

	if storedJudgment.IsInvalid() {
		t.Error("Judgment should not be invalid")
	}
}

func TestGetAllJudgmentsForWorkReport(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	workReportHash := createTestWorkReportHash([]byte("test-work-report-multiple"))

	judgments := []struct {
		epochIndex     types.U32
		validatorIndex types.ValidatorIndex
		validity       uint8
	}{
		{types.U32(100), types.ValidatorIndex(1), 1},
		{types.U32(100), types.ValidatorIndex(2), 0},
		{types.U32(100), types.ValidatorIndex(3), 1},
	}

	for _, j := range judgments {
		signature := createTestEd25519Signature([]byte("test-signature"))
		err := storeJudgmentAnnouncement(j.epochIndex, j.validatorIndex, j.validity, workReportHash, signature)
		if err != nil {
			t.Fatalf("Failed to store judgment for validator %d: %v", j.validatorIndex, err)
		}
	}

	retrievedJudgments, err := GetAllJudgmentsForWorkReport(workReportHash)
	if err != nil {
		t.Fatalf("Failed to retrieve judgments: %v", err)
	}

	if len(retrievedJudgments) != 3 {
		t.Errorf("Expected 3 judgments, got %d", len(retrievedJudgments))
	}

	validCount := 0
	invalidCount := 0
	for _, judgment := range retrievedJudgments {
		if judgment.IsValid() {
			validCount++
		}
		if judgment.IsInvalid() {
			invalidCount++
		}
	}

	if validCount != 2 {
		t.Errorf("Expected 2 valid judgments, got %d", validCount)
	}

	if invalidCount != 1 {
		t.Errorf("Expected 1 invalid judgment, got %d", invalidCount)
	}
}

func TestGetAllJudgmentsForEpoch(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	epochIndex := types.U32(200)

	for i := 0; i < 3; i++ {
		workReportHash := createTestWorkReportHash([]byte("test-work-report-" + string(rune(i))))
		validatorIndex := types.ValidatorIndex(i + 10)
		validity := uint8(i % 2)
		signature := createTestEd25519Signature([]byte("test-signature"))

		err := storeJudgmentAnnouncement(epochIndex, validatorIndex, validity, workReportHash, signature)
		if err != nil {
			t.Fatalf("Failed to store judgment %d: %v", i, err)
		}
	}

	retrievedJudgments, err := GetAllJudgmentsForEpoch(epochIndex)
	if err != nil {
		t.Fatalf("Failed to retrieve judgments: %v", err)
	}

	if len(retrievedJudgments) != 3 {
		t.Errorf("Expected 3 judgments, got %d", len(retrievedJudgments))
	}

	for _, judgment := range retrievedJudgments {
		if judgment.EpochIndex != epochIndex {
			t.Errorf("Judgment has wrong epoch index: expected %d, got %d", epochIndex, judgment.EpochIndex)
		}
	}
}

func TestCE145PayloadEncodeDecode(t *testing.T) {
	epochIndex := types.U32(98765)
	validatorIndex := types.ValidatorIndex(123)
	validity := uint8(0)
	workReportHash := createTestWorkReportHash([]byte("test-work-report-encode"))
	signature := createTestEd25519Signature([]byte("test-signature-encode"))

	payload := &CE145Payload{
		EpochIndex:     epochIndex,
		ValidatorIndex: validatorIndex,
		Validity:       validity,
		WorkReportHash: workReportHash,
		Signature:      signature,
	}

	if err := payload.Validate(); err != nil {
		t.Errorf("Payload validation failed: %v", err)
	}
	encoded, err := payload.Encode()
	if err != nil {
		t.Errorf("Payload encoding failed: %v", err)
	}

	expectedSize := 4 + 2 + 1 + 32 + 64 // EpochIndex + ValidatorIndex + Validity + WorkReportHash + Signature
	if len(encoded) != expectedSize {
		t.Errorf("Encoded size mismatch: expected %d, got %d", expectedSize, len(encoded))
	}

	decodedPayload := &CE145Payload{}
	if err := decodedPayload.Decode(encoded); err != nil {
		t.Errorf("Payload decoding failed: %v", err)
	}

	if decodedPayload.EpochIndex != payload.EpochIndex {
		t.Errorf("Decoded epoch index doesn't match original: expected %d, got %d",
			payload.EpochIndex, decodedPayload.EpochIndex)
	}

	if decodedPayload.ValidatorIndex != payload.ValidatorIndex {
		t.Errorf("Decoded validator index doesn't match original: expected %d, got %d",
			payload.ValidatorIndex, decodedPayload.ValidatorIndex)
	}

	if decodedPayload.Validity != payload.Validity {
		t.Errorf("Decoded validity doesn't match original: expected %d, got %d",
			payload.Validity, decodedPayload.Validity)
	}

	if !bytes.Equal(decodedPayload.WorkReportHash[:], payload.WorkReportHash[:]) {
		t.Error("Decoded work report hash doesn't match original")
	}

	if !bytes.Equal(decodedPayload.Signature[:], payload.Signature[:]) {
		t.Error("Decoded signature doesn't match original")
	}

	if decodedPayload.IsValid() {
		t.Error("Decoded payload should not be valid (validity = 0)")
	}

	if !decodedPayload.IsInvalid() {
		t.Error("Decoded payload should be invalid (validity = 0)")
	}
}
