package ce

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandleBundleRequest_RoundTrip(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	erasureRoot := make([]byte, HashSize)
	for i := range erasureRoot {
		erasureRoot[i] = byte(i + 1)
	}

	testBundle := CreateTestWorkPackageBundle()
	enc := types.NewEncoder()
	enc.SetHashSegmentMap(map[types.OpaqueHash]types.OpaqueHash{})
	bundleBytes, err := enc.Encode(testBundle)
	if err != nil {
		t.Fatalf("encode bundle: %v", err)
	}
	if err := PutKV(db, wpBundleKey(erasureRoot), bundleBytes); err != nil {
		t.Fatalf("PutKV: %v", err)
	}

	stream := newMockStream(framePayload(erasureRoot))
	fakeBC := SetupFakeBlockchain()
	if err := HandleBundleRequest(fakeBC, &quic.Stream{Stream: stream}); err != nil {
		t.Fatalf("HandleBundleRequest: %v", err)
	}

	resp := stream.w.Bytes()
	if len(resp) < 4 {
		t.Fatalf("response too short")
	}
	msgLen := binary.LittleEndian.Uint32(resp[:4])
	if int(msgLen)+4 != len(resp) {
		t.Fatalf("expected single framed message, got total len %d msg len %d", len(resp), msgLen)
	}
	got := resp[4 : 4+msgLen]
	if !bytes.Equal(got, bundleBytes) {
		t.Fatalf("response bundle bytes mismatch")
	}

	var decoded types.WorkPackageBundle
	dec := types.NewDecoder()
	dec.SetHashSegmentMap(map[types.OpaqueHash]types.OpaqueHash{})
	if err := dec.Decode(got, &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func TestHandleBundleRequest_RequestTooShort(t *testing.T) {
	short := make([]byte, HashSize-1)
	stream := newMockStream(framePayload(short))
	err := HandleBundleRequest(nil, &quic.Stream{Stream: stream})
	if err == nil {
		t.Fatal("expected error for short erasure root")
	}
}

func TestEncodeBundleRequest(t *testing.T) {
	h := NewDefaultCERequestHandler()
	root := make([]byte, HashSize)
	for i := range root {
		root[i] = byte(i)
	}
	out, err := h.Encode(BundleRequest, &CE147Payload{ErasureRoot: root})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// Same pattern as CE128/CE129: Encode returns body only; callers send protocol ID separately then frame this payload.
	if len(out) != CE147RequestSize {
		t.Fatalf("encoded len: got %d want %d", len(out), CE147RequestSize)
	}
	if !bytes.Equal(out, root) {
		t.Fatalf("erasure root mismatch in encoded payload")
	}
}
