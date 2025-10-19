package ce

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandlePreimageAnnouncement(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	serviceID := types.ServiceId(12345)
	hash := types.OpaqueHash{}
	for i := range hash {
		hash[i] = byte(i + 1)
	}
	preimageLength := types.U32(1024)

	announcement, err := CreatePreimageAnnouncement(serviceID, hash, preimageLength)
	if err != nil {
		t.Fatalf("Failed to create preimage announcement: %v", err)
	}

	stream := newMockStream(append(announcement, []byte("FIN")...))

	fakeBlockchain := SetupFakeBlockchain()

	err = HandlePreimageAnnouncement_Validator(fakeBlockchain, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("HandlePreimageAnnouncement failed: %v", err)
	}

	response := stream.w.Bytes()
	if string(response) != "FIN" {
		t.Errorf("Expected FIN response, got: %s", string(response))
	}

	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		t.Fatalf("Failed to get Redis backend: %v", err)
	}

	hashHex := hex.EncodeToString(hash[:])
	key := "preimage_announcement:" + hashHex

	client := redisBackend.GetClient()
	storedData, err := client.Get(key)
	if err != nil {
		t.Fatalf("Failed to get stored announcement from Redis: %v", err)
	}

	if storedData == nil {
		t.Fatal("Announcement was not stored in Redis")
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

	serviceKey := fmt.Sprintf("service_preimage_announcements:%d", serviceID)
	isMember, err := client.SIsMember(serviceKey, storedData)
	if err != nil {
		t.Fatalf("Failed to check if announcement is in service set: %v", err)
	}

	if !isMember {
		t.Error("Announcement was not added to the service set")
	}
}

func TestCE142Payload(t *testing.T) {
	payload := &CE142Payload{
		ServiceID:      types.ServiceId(54321),
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
