package ce

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestHandlePreimageRequest(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	testPreimage := []byte("This is a test preimage for CE143 protocol testing")

	opaqueHash := hash.Blake2bHash(types.ByteSequence(testPreimage))

	err := StorePreimage(nil, opaqueHash, testPreimage)
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
