package ce

import (
	"bytes"
	"encoding/binary"
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
	if len(response) < 4 {
		t.Fatalf("response too short for message frame")
	}
	n := binary.LittleEndian.Uint32(response[:4])
	payload := response[4:]
	if uint32(len(payload)) < n {
		t.Fatalf("response truncated: want %d payload bytes, got %d", n, len(payload))
	}
	payload = payload[:n]
	if !bytes.Equal(payload, testPreimage) {
		t.Errorf("Response mismatch.\nExpected: %s\nGot: %s",
			string(testPreimage), string(payload))
	}
}
