package ce

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandlePreimageAnnouncement(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	serviceID := types.ServiceID(12345)
	hash := types.OpaqueHash{}
	for i := range hash {
		hash[i] = byte(i + 1)
	}
	preimageLength := types.U32(1024)

	announcement, err := CreatePreimageAnnouncement(serviceID, hash, preimageLength)
	if err != nil {
		t.Fatalf("Failed to create preimage announcement: %v", err)
	}

	stream := newMockStream(announcement)

	fakeBlockchain := SetupFakeBlockchain()

	err = HandlePreimageAnnouncement(fakeBlockchain, stream)
	if err != nil {
		t.Fatalf("HandlePreimageAnnouncement failed: %v", err)
	}

	response := stream.w.Bytes()
	if len(response) != 0 {
		t.Errorf("Expected no response bytes, got: %x", response)
	}

	storedData, err := GetKV(db, cePreimageAnnKey(hash))
	if err != nil {
		t.Fatalf("Failed to get stored announcement: %v", err)
	}

	if storedData == nil {
		t.Fatal("Announcement was not stored")
	}

	// Verify decoded data matches original
	decodedPayload := &CE142Payload{}
	err = decodedPayload.Decode(storedData)
	if err != nil {
		t.Fatalf("Failed to decode stored announcement: %v", err)
	}

	if decodedPayload.ServiceID != serviceID {
		t.Errorf("Stored service ID doesn't match original: expected %d, got %d", serviceID, decodedPayload.ServiceID)
	}

	if !bytes.Equal(decodedPayload.Hash[:], hash[:]) {
		t.Error("Stored hash doesn't match original")
	}

	if decodedPayload.PreimageLength != preimageLength {
		t.Errorf("Stored preimage length doesn't match original: expected %d, got %d", preimageLength, decodedPayload.PreimageLength)
	}

	isMember, err := SIsMember(db, cePreimageAnnServiceSetKey(serviceID), storedData)
	if err != nil {
		t.Fatalf("Failed to check if announcement is in service set: %v", err)
	}

	if !isMember {
		t.Error("Announcement was not added to the service set")
	}
}

func TestCE142Payload(t *testing.T) {
	payload := &CE142Payload{
		ServiceID:      types.ServiceID(54321),
		Hash:           types.OpaqueHash{},
		PreimageLength: types.U32(2048),
	}

	for i := range payload.Hash {
		payload.Hash[i] = byte(i * 2)
	}

	if err := payload.Validate(); err != nil {
		t.Errorf("Payload validation failed: %v", err)
	}

	encoded, err := payload.Encode()
	if err != nil {
		t.Errorf("Payload encoding failed: %v", err)
	}

	expectedSize := 4 + 32 + 4 // ServiceID + Hash + PreimageLength
	if len(encoded) != expectedSize {
		t.Errorf("Encoded size mismatch: expected %d, got %d", expectedSize, len(encoded))
	}

	decodedPayload := &CE142Payload{}
	if err := decodedPayload.Decode(encoded); err != nil {
		t.Errorf("Payload decoding failed: %v", err)
	}

	if decodedPayload.ServiceID != payload.ServiceID {
		t.Errorf("Decoded service ID doesn't match original: expected %d, got %d", payload.ServiceID, decodedPayload.ServiceID)
	}

	if !bytes.Equal(decodedPayload.Hash[:], payload.Hash[:]) {
		t.Error("Decoded hash doesn't match original")
	}

	if decodedPayload.PreimageLength != payload.PreimageLength {
		t.Errorf("Decoded preimage length doesn't match original: expected %d, got %d", payload.PreimageLength, decodedPayload.PreimageLength)
	}
}
