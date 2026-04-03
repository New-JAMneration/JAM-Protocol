package ce

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/stretchr/testify/require"
)

// framedMsg wraps payload with a 4-byte LE length prefix (JAMNP message framing).
func framedMsg(payload []byte) []byte {
	var buf bytes.Buffer
	quic.WriteMessageFrame(&buf, payload)
	return buf.Bytes()
}

// makeGuaranteeBytes builds the raw wire bytes for a CE145 guarantee message.
// count must be 2 or 3; signatures are filled with a deterministic pattern.
func makeGuaranteeBytes(slot uint32, count int) []byte {
	var buf []byte
	buf = binary.LittleEndian.AppendUint32(buf, slot)
	buf = append(buf, byte(count)) // compact length
	for i := 0; i < count; i++ {
		buf = binary.LittleEndian.AppendUint16(buf, uint16(i+1)) // ValidatorIndex
		sig := createTestEd25519Signature([]byte{byte(i + 1)})
		buf = append(buf, sig[:]...)
	}
	return buf
}

// buildInvalidJudgmentStream returns two JAMNP-framed messages for a CE145 invalid
// judgment stream, ready for newMockStream.
func buildInvalidJudgmentStream(
	epochIndex types.U32,
	validatorIndex types.ValidatorIndex,
	workReportHash types.WorkReportHash,
	signature types.Ed25519Signature,
	guaranteeSlot uint32,
	guaranteeCount int,
) []byte {
	var header []byte
	header = binary.LittleEndian.AppendUint32(header, uint32(epochIndex))
	header = binary.LittleEndian.AppendUint16(header, uint16(validatorIndex))
	header = append(header, 0) // validity = 0 (invalid)
	header = append(header, workReportHash[:]...)
	header = append(header, signature[:]...)

	var out bytes.Buffer
	quic.WriteMessageFrame(&out, header)
	quic.WriteMessageFrame(&out, makeGuaranteeBytes(guaranteeSlot, guaranteeCount))
	return out.Bytes()
}

// --------------------------------------------------------------------------
// Existing happy-path test: valid judgment (no guarantee)
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncementValid(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	epochIndex := types.U32(12345)
	validatorIndex := types.ValidatorIndex(67)
	validity := uint8(1)
	workReportHash := createTestWorkReportHash([]byte("test-work-report-valid"))
	signature := createTestEd25519Signature([]byte("test-signature-valid"))

	announcementMsg, err := CreateJudgmentAnnouncement(epochIndex, validatorIndex, validity, workReportHash, signature)
	require.NoError(t, err, "CreateJudgmentAnnouncement")

	stream := newMockStream(framedMsg(announcementMsg))
	fakeBlockchain := SetupFakeBlockchain()

	require.NoError(t, HandleJudgmentAnnouncement(fakeBlockchain, stream), "HandleJudgmentAnnouncement")
	require.Empty(t, stream.w.Bytes(), "expected no response bytes")

	storedJudgment, err := GetJudgment(fakeBlockchain, workReportHash, epochIndex, validatorIndex)
	require.NoError(t, err, "GetJudgment")

	require.Equal(t, epochIndex, storedJudgment.EpochIndex, "epoch index mismatch")
	require.Equal(t, validatorIndex, storedJudgment.ValidatorIndex, "validator index mismatch")
	require.Equal(t, validity, storedJudgment.Validity, "validity mismatch")
	require.True(t, bytes.Equal(storedJudgment.WorkReportHash[:], workReportHash[:]), "stored work report hash doesn't match original")
	require.True(t, bytes.Equal(storedJudgment.Signature[:], signature[:]), "stored signature doesn't match original")
	require.True(t, storedJudgment.IsValid(), "judgment should be valid")
	require.False(t, storedJudgment.IsInvalid(), "judgment should not be invalid")
	require.Nil(t, storedJudgment.Guarantee, "valid judgment must not carry guarantee")
}

// --------------------------------------------------------------------------
// Happy-path: invalid judgment + 2 guarantee signatures
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncementInvalidWith2Sigs(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	epochIndex := types.U32(111)
	validatorIndex := types.ValidatorIndex(5)
	workReportHash := createTestWorkReportHash([]byte("test-invalid-2sigs"))
	signature := createTestEd25519Signature([]byte("test-sig-invalid-2sigs"))

	buf := buildInvalidJudgmentStream(epochIndex, validatorIndex, workReportHash, signature, 999, 2)
	stream := newMockStream(buf)
	fakeBlockchain := SetupFakeBlockchain()

	require.NoError(t, HandleJudgmentAnnouncement(fakeBlockchain, stream), "HandleJudgmentAnnouncement")

	stored, err := GetJudgment(fakeBlockchain, workReportHash, epochIndex, validatorIndex)
	require.NoError(t, err, "GetJudgment")
	require.Equal(t, uint8(0), stored.Validity, "expected validity 0")
	require.True(t, stored.IsInvalid(), "judgment should be invalid")
	require.NotNil(t, stored.Guarantee, "stored judgment must carry guarantee for invalid judgment")
	require.Len(t, stored.Guarantee.Signatures, 2, "expected 2 guarantee signatures")
	require.Equal(t, types.TimeSlot(999), stored.Guarantee.Slot, "expected guarantee slot 999")
}

// --------------------------------------------------------------------------
// Happy-path: invalid judgment + 3 guarantee signatures
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncementInvalidWith3Sigs(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	epochIndex := types.U32(222)
	validatorIndex := types.ValidatorIndex(10)
	workReportHash := createTestWorkReportHash([]byte("test-invalid-3sigs"))
	signature := createTestEd25519Signature([]byte("test-sig-invalid-3sigs"))

	buf := buildInvalidJudgmentStream(epochIndex, validatorIndex, workReportHash, signature, 1234, 3)
	stream := newMockStream(buf)
	fakeBlockchain := SetupFakeBlockchain()

	require.NoError(t, HandleJudgmentAnnouncement(fakeBlockchain, stream), "HandleJudgmentAnnouncement")

	stored, err := GetJudgment(fakeBlockchain, workReportHash, epochIndex, validatorIndex)
	require.NoError(t, err, "GetJudgment")
	require.NotNil(t, stored.Guarantee, "stored judgment must carry guarantee")
	require.Len(t, stored.Guarantee.Signatures, 3, "expected 3 guarantee signatures")
}

// --------------------------------------------------------------------------
// Error: invalid judgment with guarantee count = 0
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncementInvalidGuaranteeCount0(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	epochIndex := types.U32(300)
	validatorIndex := types.ValidatorIndex(1)
	workReportHash := createTestWorkReportHash([]byte("test-count0"))
	signature := createTestEd25519Signature([]byte("test-sig-count0"))

	var header []byte
	header = binary.LittleEndian.AppendUint32(header, uint32(epochIndex))
	header = binary.LittleEndian.AppendUint16(header, uint16(validatorIndex))
	header = append(header, 0) // validity = 0
	header = append(header, workReportHash[:]...)
	header = append(header, signature[:]...)
	guarantee := binary.LittleEndian.AppendUint32(nil, 888)
	guarantee = append(guarantee, 0) // count = 0

	var buf bytes.Buffer
	quic.WriteMessageFrame(&buf, header)
	quic.WriteMessageFrame(&buf, guarantee)
	stream := newMockStream(buf.Bytes())
	fakeBlockchain := SetupFakeBlockchain()

	err := HandleJudgmentAnnouncement(fakeBlockchain, stream)
	require.Error(t, err, "expected error for guarantee count 0")
	require.True(t, strings.Contains(err.Error(), "out of range"), "unexpected error message: %v", err)
}

// --------------------------------------------------------------------------
// Error: invalid judgment with guarantee count = 1
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncementInvalidGuaranteeCount1(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	epochIndex := types.U32(301)
	validatorIndex := types.ValidatorIndex(2)
	workReportHash := createTestWorkReportHash([]byte("test-count1"))
	signature := createTestEd25519Signature([]byte("test-sig-count1"))

	var header []byte
	header = binary.LittleEndian.AppendUint32(header, uint32(epochIndex))
	header = binary.LittleEndian.AppendUint16(header, uint16(validatorIndex))
	header = append(header, 0) // validity = 0
	header = append(header, workReportHash[:]...)
	header = append(header, signature[:]...)
	onlySig := createTestEd25519Signature([]byte("only-sig"))
	guarantee := binary.LittleEndian.AppendUint32(nil, 777)
	guarantee = append(guarantee, 1) // count = 1
	guarantee = binary.LittleEndian.AppendUint16(guarantee, 0)
	guarantee = append(guarantee, onlySig[:]...)

	var buf bytes.Buffer
	quic.WriteMessageFrame(&buf, header)
	quic.WriteMessageFrame(&buf, guarantee)
	stream := newMockStream(buf.Bytes())
	fakeBlockchain := SetupFakeBlockchain()

	err := HandleJudgmentAnnouncement(fakeBlockchain, stream)
	require.Error(t, err, "expected error for guarantee count 1")
	require.True(t, strings.Contains(err.Error(), "out of range"), "unexpected error message: %v", err)
}

// --------------------------------------------------------------------------
// Error: invalid judgment with guarantee count = 4
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncementInvalidGuaranteeCount4(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	epochIndex := types.U32(302)
	validatorIndex := types.ValidatorIndex(3)
	workReportHash := createTestWorkReportHash([]byte("test-count4"))
	signature := createTestEd25519Signature([]byte("test-sig-count4"))

	var header []byte
	header = binary.LittleEndian.AppendUint32(header, uint32(epochIndex))
	header = binary.LittleEndian.AppendUint16(header, uint16(validatorIndex))
	header = append(header, 0) // validity = 0
	header = append(header, workReportHash[:]...)
	header = append(header, signature[:]...)
	guarantee := binary.LittleEndian.AppendUint32(nil, 666)
	guarantee = append(guarantee, 4) // count = 4
	for i := 0; i < 4; i++ {
		s := createTestEd25519Signature([]byte{byte(i)})
		guarantee = binary.LittleEndian.AppendUint16(guarantee, uint16(i))
		guarantee = append(guarantee, s[:]...)
	}

	var buf bytes.Buffer
	quic.WriteMessageFrame(&buf, header)
	quic.WriteMessageFrame(&buf, guarantee)
	stream := newMockStream(buf.Bytes())
	fakeBlockchain := SetupFakeBlockchain()

	err := HandleJudgmentAnnouncement(fakeBlockchain, stream)
	require.Error(t, err, "expected error for guarantee count 4")
	require.True(t, strings.Contains(err.Error(), "out of range"), "unexpected error message: %v", err)
}

// --------------------------------------------------------------------------
// Error: invalid judgment missing guarantee entirely
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncementInvalidMissingGuarantee(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	epochIndex := types.U32(303)
	validatorIndex := types.ValidatorIndex(4)
	workReportHash := createTestWorkReportHash([]byte("test-missing-guarantee"))
	signature := createTestEd25519Signature([]byte("test-sig-missing"))

	// Manually build the raw header bytes (Validity=0) without going through
	// CreateJudgmentAnnouncement, which now requires a guarantee for invalid judgments.
	var header []byte
	header = binary.LittleEndian.AppendUint32(header, uint32(epochIndex))
	header = binary.LittleEndian.AppendUint16(header, uint16(validatorIndex))
	header = append(header, 0) // validity = 0 (invalid)
	header = append(header, workReportHash[:]...)
	header = append(header, signature[:]...)

	// Only stream msg1; no guarantee message follows — handler must return an error.
	stream := newMockStream(framedMsg(header))
	fakeBlockchain := SetupFakeBlockchain()

	err := HandleJudgmentAnnouncement(fakeBlockchain, stream)
	require.Error(t, err, "expected error when guarantee is missing for invalid judgment")
}

// --------------------------------------------------------------------------
// Sender: valid judgment does not write guarantee bytes
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncement_Send_Valid(t *testing.T) {
	stream := newMockStream(nil)

	payload := &CE145Payload{
		EpochIndex:     types.U32(500),
		ValidatorIndex: types.ValidatorIndex(10),
		Validity:       1,
		WorkReportHash: createTestWorkReportHash([]byte("send-valid")),
		Signature:      createTestEd25519Signature([]byte("send-valid-sig")),
		Guarantee:      nil,
	}

	require.NoError(t, HandleJudgmentAnnouncement_Send(stream, payload), "Send")

	written := stream.w.Bytes()
	// Expect: protocol byte (1) + framed header (4+103) = 108 bytes
	require.Len(t, written, 1+4+103, "expected 108 bytes for valid judgment")
	require.Equal(t, byte(145), written[0], "expected protocol ID 145")
	// Validity byte: written[0]=protocol, written[1:5]=length, written[5:]=header.
	// Validity is at header offset ce145OffValidity (6), so written[5+6] = written[11].
	require.Equal(t, byte(1), written[11], "expected validity byte 1")
}

// --------------------------------------------------------------------------
// Sender: invalid judgment writes both judgment header and guarantee
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncement_Send_Invalid(t *testing.T) {
	stream := newMockStream(nil)

	payload := &CE145Payload{
		EpochIndex:     types.U32(600),
		ValidatorIndex: types.ValidatorIndex(20),
		Validity:       0,
		WorkReportHash: createTestWorkReportHash([]byte("send-invalid")),
		Signature:      createTestEd25519Signature([]byte("send-invalid-sig")),
		Guarantee: &CE145Guarantee{
			Slot: 42,
			Signatures: []types.ValidatorSignature{
				{ValidatorIndex: 1, Signature: createTestEd25519Signature([]byte("g1"))},
				{ValidatorIndex: 2, Signature: createTestEd25519Signature([]byte("g2"))},
			},
		},
	}

	require.NoError(t, HandleJudgmentAnnouncement_Send(stream, payload), "Send")

	written := stream.w.Bytes()
	// Expect: protocol byte (1) + framed header (4+103) + framed guarantee (4 + Slot:4 + count:1 + 2*66)
	expectedGuaranteePayload := 4 + 1 + 2*66 // 137
	expected := 1 + (4 + 103) + (4 + expectedGuaranteePayload)
	require.Len(t, written, expected, "expected %d bytes for invalid judgment with 2 guarantee sigs", expected)
}

// --------------------------------------------------------------------------
// Sender: invalid judgment without guarantee field must fail
// --------------------------------------------------------------------------

func TestHandleJudgmentAnnouncement_Send_InvalidMissingGuarantee(t *testing.T) {
	stream := newMockStream(nil)

	payload := &CE145Payload{
		EpochIndex:     types.U32(700),
		ValidatorIndex: types.ValidatorIndex(30),
		Validity:       0,
		WorkReportHash: createTestWorkReportHash([]byte("send-invalid-no-guarantee")),
		Signature:      createTestEd25519Signature([]byte("send-sig")),
		Guarantee:      nil,
	}

	err := HandleJudgmentAnnouncement_Send(stream, payload)
	require.Error(t, err, "expected error when Guarantee is nil for invalid judgment")
}

// --------------------------------------------------------------------------
// Encode/Decode round-trip: valid judgment (Validity==1, no guarantee)
// --------------------------------------------------------------------------

func TestCE145PayloadEncodeDecode(t *testing.T) {
	epochIndex := types.U32(98765)
	validatorIndex := types.ValidatorIndex(123)
	workReportHash := createTestWorkReportHash([]byte("test-work-report-encode"))
	signature := createTestEd25519Signature([]byte("test-signature-encode"))

	payload := &CE145Payload{
		EpochIndex:     epochIndex,
		ValidatorIndex: validatorIndex,
		Validity:       1,
		WorkReportHash: workReportHash,
		Signature:      signature,
	}

	require.NoError(t, payload.Validate(), "Validate")

	encoded, err := payload.Encode()
	require.NoError(t, err, "Encode")

	// Valid judgment has no guarantee → only header bytes
	require.Len(t, encoded, 4+2+1+32+64, "encoded size mismatch")

	decoded := &CE145Payload{}
	require.NoError(t, decoded.Decode(encoded), "Decode")

	require.Equal(t, epochIndex, decoded.EpochIndex, "epoch index mismatch")
	require.Equal(t, validatorIndex, decoded.ValidatorIndex, "validator index mismatch")
	require.Equal(t, uint8(1), decoded.Validity, "validity mismatch")
	require.True(t, bytes.Equal(decoded.WorkReportHash[:], workReportHash[:]), "decoded work report hash doesn't match original")
	require.True(t, bytes.Equal(decoded.Signature[:], signature[:]), "decoded signature doesn't match original")
	require.True(t, decoded.IsValid(), "decoded payload should be valid (validity = 1)")
	require.False(t, decoded.IsInvalid(), "decoded payload should not be invalid (validity = 1)")
	require.Nil(t, decoded.Guarantee, "valid judgment must not carry guarantee")
}

// --------------------------------------------------------------------------
// Encode/Decode round-trip: invalid judgment with guarantee
// --------------------------------------------------------------------------

func TestCE145PayloadEncodeDecodeWithGuarantee(t *testing.T) {
	workReportHash := createTestWorkReportHash([]byte("encode-decode-guarantee"))
	guarantee := &CE145Guarantee{
		Slot: 555,
		Signatures: []types.ValidatorSignature{
			{ValidatorIndex: 0, Signature: createTestEd25519Signature([]byte("gs1"))},
			{ValidatorIndex: 1, Signature: createTestEd25519Signature([]byte("gs2"))},
			{ValidatorIndex: 2, Signature: createTestEd25519Signature([]byte("gs3"))},
		},
	}
	payload := &CE145Payload{
		EpochIndex:     types.U32(400),
		ValidatorIndex: types.ValidatorIndex(50),
		Validity:       0,
		WorkReportHash: workReportHash,
		Signature:      createTestEd25519Signature([]byte("encode-sig")),
		Guarantee:      guarantee,
	}

	encoded, err := payload.Encode()
	require.NoError(t, err, "Encode")

	// 103 (header) + 4 (slot) + 1 (count) + 3*66 (sigs) = 306
	require.Len(t, encoded, 103+4+1+3*66, "encoded size mismatch")

	decoded := &CE145Payload{}
	require.NoError(t, decoded.Decode(encoded), "Decode")

	require.NotNil(t, decoded.Guarantee, "decoded Guarantee should not be nil")
	require.Len(t, decoded.Guarantee.Signatures, 3, "expected 3 guarantee sigs")
	require.Equal(t, types.TimeSlot(555), decoded.Guarantee.Slot, "expected slot 555")
	require.Equal(t, types.ValidatorIndex(0), decoded.Guarantee.Signatures[0].ValidatorIndex, "expected validator index 0")
	require.True(t, bytes.Equal(decoded.Guarantee.Signatures[2].Signature[:], guarantee.Signatures[2].Signature[:]), "guarantee signature[2] mismatch")
}

// --------------------------------------------------------------------------
// GetAllJudgmentsForWorkReport: mixed valid/invalid
// --------------------------------------------------------------------------

func TestGetAllJudgmentsForWorkReport(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	fakeBlockchain := SetupFakeBlockchain()

	workReportHash := createTestWorkReportHash([]byte("test-work-report-multiple"))

	validSig := createTestEd25519Signature([]byte("valid-sig"))
	invalidSig := createTestEd25519Signature([]byte("invalid-sig"))
	guarantee := &CE145Guarantee{
		Slot: 100,
		Signatures: []types.ValidatorSignature{
			{ValidatorIndex: 0, Signature: createTestEd25519Signature([]byte("g0"))},
			{ValidatorIndex: 1, Signature: createTestEd25519Signature([]byte("g1"))},
		},
	}

	// Store 2 valid + 1 invalid judgment
	_ = storeJudgmentAnnouncement(fakeBlockchain, types.U32(100), types.ValidatorIndex(1), 1, workReportHash, validSig, nil)
	_ = storeJudgmentAnnouncement(fakeBlockchain, types.U32(100), types.ValidatorIndex(2), 0, workReportHash, invalidSig, guarantee)
	_ = storeJudgmentAnnouncement(fakeBlockchain, types.U32(100), types.ValidatorIndex(3), 1, workReportHash, validSig, nil)

	retrieved, err := GetAllJudgmentsForWorkReport(fakeBlockchain, workReportHash)
	require.NoError(t, err, "GetAllJudgmentsForWorkReport")
	require.Len(t, retrieved, 3, "expected 3 judgments")

	validCount, invalidCount := 0, 0
	for _, j := range retrieved {
		if j.IsValid() {
			validCount++
		}
		if j.IsInvalid() {
			invalidCount++
			require.NotNil(t, j.Guarantee, "invalid judgment retrieved from store must have Guarantee")
		}
	}
	require.Equal(t, 2, validCount, "expected 2 valid judgments")
	require.Equal(t, 1, invalidCount, "expected 1 invalid judgment")
}

// --------------------------------------------------------------------------
// GetAllJudgmentsForEpoch
// --------------------------------------------------------------------------

func TestGetAllJudgmentsForEpoch(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	fakeBlockchain := SetupFakeBlockchain()

	epochIndex := types.U32(200)

	for i := 0; i < 3; i++ {
		wrHash := createTestWorkReportHash([]byte("test-work-report-" + string(rune(i))))
		vi := types.ValidatorIndex(i + 10)
		validity := uint8(i % 2)
		sig := createTestEd25519Signature([]byte("test-signature"))

		var g *CE145Guarantee
		if validity == 0 {
			g = &CE145Guarantee{
				Slot: types.TimeSlot(i),
				Signatures: []types.ValidatorSignature{
					{ValidatorIndex: types.ValidatorIndex(i), Signature: createTestEd25519Signature([]byte("gx"))},
					{ValidatorIndex: types.ValidatorIndex(i + 1), Signature: createTestEd25519Signature([]byte("gy"))},
				},
			}
		}

		err := storeJudgmentAnnouncement(fakeBlockchain, epochIndex, vi, validity, wrHash, sig, g)
		require.NoError(t, err, "storeJudgmentAnnouncement %d", i)
	}

	retrieved, err := GetAllJudgmentsForEpoch(fakeBlockchain, epochIndex)
	require.NoError(t, err, "GetAllJudgmentsForEpoch")
	require.Len(t, retrieved, 3, "expected 3 judgments")
	for _, j := range retrieved {
		require.Equal(t, epochIndex, j.EpochIndex, "wrong epoch index")
	}
}
