package ce

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandleJudgmentAnnouncementValid(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	epochIndex := types.U32(12345)
	validatorIndex := types.ValidatorIndex(67)
	validity := uint8(1) // Valid
	workReportHash := createTestWorkReportHash([]byte("test-work-report-valid"))
	signature := createTestEd25519Signature([]byte("test-signature-valid"))

	announcementMsg, err := CreateJudgmentAnnouncement(epochIndex, validatorIndex, validity, workReportHash, signature)
	if err != nil {
		t.Fatalf("Failed to create judgment announcement: %v", err)
	}

	fullMessage := announcementMsg
	stream := newMockStream(fullMessage)

	fakeBlockchain := SetupFakeBlockchain()

	err = HandleJudgmentAnnouncement(fakeBlockchain, stream)
	if err != nil {
		t.Fatalf("HandleJudgmentAnnouncement failed: %v", err)
	}

	response := stream.w.Bytes()
	if len(response) != 0 {
		t.Errorf("Expected no response bytes, got: %x", response)
	}

	storedJudgment, err := GetJudgment(fakeBlockchain, workReportHash, epochIndex, validatorIndex)
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

func TestHandleJudgmentAnnouncementInvalidWithGuarantee(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	epochIndex := types.U32(111)
	validatorIndex := types.ValidatorIndex(5)
	validity := uint8(0) // Invalid - requires Guarantee message before WorkReportHash
	workReportHash := createTestWorkReportHash([]byte("test-invalid"))
	signature := createTestEd25519Signature([]byte("test-sig-invalid"))

	// Build message: EpochIndex + ValidatorIndex + Validity + Guarantee(Slot u32 ++ len++[ValidatorIndex++Sig]) + WorkReportHash + Signature
	buf := make([]byte, 0, 200)
	buf = append(buf, byte(epochIndex), byte(epochIndex>>8), byte(epochIndex>>16), byte(epochIndex>>24))
	buf = append(buf, byte(validatorIndex), byte(validatorIndex>>8))
	buf = append(buf, validity)

	// Guarantee: Slot (4) + len++ (0 = 1 byte 0x00)
	slot := uint32(999)
	buf = append(buf, byte(slot), byte(slot>>8), byte(slot>>16), byte(slot>>24))
	buf = append(buf, 0) // len++ = 0 (no guarantors)
	buf = append(buf, workReportHash[:]...)
	buf = append(buf, signature[:]...)

	stream := newMockStream(buf)
	fakeBlockchain := SetupFakeBlockchain()

	err := HandleJudgmentAnnouncement(fakeBlockchain, stream)
	if err != nil {
		t.Fatalf("HandleJudgmentAnnouncement (invalid+guarantee) failed: %v", err)
	}

	storedJudgment, err := GetJudgment(fakeBlockchain, workReportHash, epochIndex, validatorIndex)
	if err != nil {
		t.Fatalf("Failed to retrieve stored judgment: %v", err)
	}
	if storedJudgment.Validity != 0 {
		t.Errorf("Expected validity 0, got %d", storedJudgment.Validity)
	}
	if !storedJudgment.IsInvalid() {
		t.Error("Judgment should be invalid")
	}
}

func TestGetAllJudgmentsForWorkReport(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	fakeBlockchain := SetupFakeBlockchain()

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
		err := storeJudgmentAnnouncement(fakeBlockchain, j.epochIndex, j.validatorIndex, j.validity, workReportHash, signature)
		if err != nil {
			t.Fatalf("Failed to store judgment for validator %d: %v", j.validatorIndex, err)
		}
	}

	retrievedJudgments, err := GetAllJudgmentsForWorkReport(fakeBlockchain, workReportHash)
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
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	fakeBlockchain := SetupFakeBlockchain()

	epochIndex := types.U32(200)

	for i := 0; i < 3; i++ {
		workReportHash := createTestWorkReportHash([]byte("test-work-report-" + string(rune(i))))
		validatorIndex := types.ValidatorIndex(i + 10)
		validity := uint8(i % 2)
		signature := createTestEd25519Signature([]byte("test-signature"))

		err := storeJudgmentAnnouncement(fakeBlockchain, epochIndex, validatorIndex, validity, workReportHash, signature)
		if err != nil {
			t.Fatalf("Failed to store judgment %d: %v", i, err)
		}
	}

	retrievedJudgments, err := GetAllJudgmentsForEpoch(fakeBlockchain, epochIndex)
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
