package ce

import (
	"bytes"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestHandlePreimageRequest(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	testPreimage := []byte("This is a test preimage for CE143 protocol testing")

	opaqueHash := hash.Blake2bHash(types.ByteSequence(testPreimage))

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

	err = HandlePreimageRequest(fakeBlockchain, stream)
	if err != nil {
		t.Fatalf("HandlePreimageRequest failed: %v", err)
	}

	response := stream.w.Bytes()
	if !bytes.Equal(response, testPreimage) {
		t.Errorf("Response mismatch.\nExpected: %s\nGot: %s",
			string(testPreimage), string(response))
	}
}
