package ce

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestHandleSegmentShardRequest(t *testing.T) {
	// Erasure root (32 bytes) + shard index (4 bytes) + segment indices length (2 bytes) + segment indices + FIN
	erasureRoot := make([]byte, HashSize)
	for i := range erasureRoot {
		erasureRoot[i] = byte(i)
	}

	shardIndex := uint32(0)
	segmentIndices := []uint16{0, 1}
	segmentIndicesLen := uint16(len(segmentIndices))

	var buf bytes.Buffer

	buf.Write(erasureRoot)

	shardIndexBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(shardIndexBytes, uint16(shardIndex))
	buf.Write(shardIndexBytes)

	segmentIndicesLenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(segmentIndicesLenBytes, segmentIndicesLen)
	buf.Write(segmentIndicesLenBytes)

	for _, index := range segmentIndices {
		indexBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(indexBytes, index)
		buf.Write(indexBytes)
	}

	// Request is one length-prefixed message
	mockStream := newMockStream(framePayload(buf.Bytes()))

	err := HandleSegmentShardRequest(nil, &quic.Stream{Stream: mockStream})
	if err != nil {
		t.Fatalf("HandleSegmentShardRequest failed: %v", err)
	}

	response := mockStream.w.Bytes()

	// Response is one length-prefixed message: [U32Size][segmentShards]
	if len(response) < U32Size+64 {
		t.Errorf("Expected response to contain length prefix + segment shards (at least 64 bytes), got length %d", len(response))
	}
}

// TestLookupWorkPackageBundleWithCEStorage verifies bundle lookup via CE persistent db (no Redis).
func TestLookupWorkPackageBundleWithCEStorage(t *testing.T) {
	db := memory.NewDatabase()
	defer db.Close()
	SetDatabase(db)
	defer SetDatabase(nil)

	erasureRoot := make([]byte, HashSize)
	for i := range erasureRoot {
		erasureRoot[i] = byte(i)
	}

	testBundle := CreateTestWorkPackageBundle()
	encoder := types.NewEncoder()
	encoder.SetHashSegmentMap(map[types.OpaqueHash]types.OpaqueHash{})
	bundleBytes, err := encoder.Encode(testBundle)
	if err != nil {
		t.Fatalf("encode bundle: %v", err)
	}
	if err := PutKV(db, wpBundleKey(erasureRoot), bundleBytes); err != nil {
		t.Fatalf("PutKV: %v", err)
	}

	// DB(nil) returns GetDatabase() which we set above
	retrieved, err := lookupWorkPackageBundle(nil, erasureRoot)
	if err != nil {
		t.Fatalf("lookupWorkPackageBundle: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected bundle, got nil")
	}
	if len(retrieved.Package.Items) != len(testBundle.Package.Items) {
		t.Errorf("expected %d items, got %d", len(testBundle.Package.Items), len(retrieved.Package.Items))
	}
}
