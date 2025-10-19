package ce

import (
	"bytes"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandlePreimageRequest(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	testPreimage := []byte("This is a test preimage for CE143 protocol testing")

	hash := sha256.Sum256(testPreimage)
	opaqueHash := types.OpaqueHash{}
	copy(opaqueHash[:], hash[:])

	err := StorePreimage(opaqueHash, testPreimage)
	if err != nil {
		t.Fatalf("Failed to store preimage: %v", err)
	}

	request, err := CreatePreimageRequest(opaqueHash)
	if err != nil {
		t.Fatalf("Failed to create preimage request: %v", err)
	}

	stream := newMockStream(request)

	fakeBlockchain := SetupFakeBlockchain()

	err = HandlePreimageRequest(fakeBlockchain, &quic.Stream{Stream: stream})
	if err != nil {
		t.Fatalf("HandlePreimageRequest failed: %v", err)
	}

	response := stream.w.Bytes()
	expectedResponse := append(testPreimage, []byte("FIN")...)

	if !bytes.Equal(response, expectedResponse) {
		t.Errorf("Response mismatch.\nExpected: %s\nGot: %s",
			string(expectedResponse), string(response))
	}

	if len(response) < 3 {
		t.Fatal("Response too short")
	}

	responsePreimage := response[:len(response)-3] // Remove FIN
	responseFin := response[len(response)-3:]

	if !bytes.Equal(responsePreimage, testPreimage) {
		t.Errorf("Preimage content mismatch.\nExpected: %s\nGot: %s",
			string(testPreimage), string(responsePreimage))
	}

	if string(responseFin) != "FIN" {
		t.Errorf("Expected FIN suffix, got: %s", string(responseFin))
	}
}
