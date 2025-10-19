package ce

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandleAvailabilityAssuranceDistribution(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	types.CoresCount = 2

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	headerHash := types.HeaderHash{}
	for i := range headerHash {
		headerHash[i] = byte(i)
	}

	bitfield := []byte{0x03}

	assurance, err := CreateAvailabilityAssurance(headerHash, bitfield, privateKey)
	if err != nil {
		t.Fatalf("Failed to create availability assurance: %v", err)
	}

	stream := newMockStream(append(assurance, []byte("FIN")...))

	fakeBlockchain := SetupFakeBlockchain()

	err = HandleAvailabilityAssuranceDistribution_Validator(fakeBlockchain, quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("HandleAvailabilityAssuranceDistribution failed: %v", err)
	}

	response := stream.w.Bytes()
	if string(response) != "FIN" {
		t.Errorf("Expected FIN response, got: %s", string(response))
	}

	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		t.Fatalf("Failed to get Redis backend: %v", err)
	}

	headerHashHex := hex.EncodeToString(headerHash[:])
	key := "availability_assurance:" + headerHashHex

	client := redisBackend.GetClient()
	storedData, err := client.Get(key)
	if err != nil {
		t.Fatalf("Failed to get stored assurance from Redis: %v", err)
	}

	if storedData == nil {
		t.Fatal("Assurance was not stored in Redis")
	}

	decodedPayload := &CE141Payload{}
	err = decodedPayload.Decode(storedData)
	if err != nil {
		t.Fatalf("Failed to decode stored assurance: %v", err)
	}

	if !bytes.Equal(decodedPayload.HeaderHash[:], headerHash[:]) {
		t.Error("Stored header hash doesn't match original")
	}

	if !bytes.Equal(decodedPayload.Bitfield, bitfield) {
		t.Error("Stored bitfield doesn't match original")
	}

	setKey := "availability_assurances_set:" + headerHashHex
	isMember, err := client.SIsMember(setKey, storedData)
	if err != nil {
		t.Fatalf("Failed to check if assurance is in set: %v", err)
	}

	if !isMember {
		t.Error("Assurance was not added to the set")
	}
}

func TestCE141Payload(t *testing.T) {
	types.CoresCount = 2

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	payload := &CE141Payload{
		HeaderHash: types.HeaderHash{},
		Bitfield:   []byte{0x03},
		Signature:  types.Ed25519Signature{},
	}

	for i := range payload.HeaderHash {
		payload.HeaderHash[i] = byte(i)
	}

	message := make([]byte, 32+1)
	copy(message[:32], payload.HeaderHash[:])
	copy(message[32:], payload.Bitfield)
	signature := ed25519.Sign(privateKey, message)
	copy(payload.Signature[:], signature)

	if err := payload.Validate(); err != nil {
		t.Errorf("Payload validation failed: %v", err)
	}

	encoded, err := payload.Encode()
	if err != nil {
		t.Errorf("Payload encoding failed: %v", err)
	}

	decodedPayload := &CE141Payload{}
	if err := decodedPayload.Decode(encoded); err != nil {
		t.Errorf("Payload decoding failed: %v", err)
	}

	if !bytes.Equal(decodedPayload.HeaderHash[:], payload.HeaderHash[:]) {
		t.Error("Decoded header hash doesn't match original")
	}

	if !bytes.Equal(decodedPayload.Bitfield, payload.Bitfield) {
		t.Error("Decoded bitfield doesn't match original")
	}

	if !bytes.Equal(decodedPayload.Signature[:], payload.Signature[:]) {
		t.Error("Decoded signature doesn't match original")
	}
}
