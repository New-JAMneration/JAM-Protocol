package ce

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// readFramedMessage reads one JAMNP message (4-byte LE length + payload) from b at offset; returns payload and new offset.
func readFramedMessage(b []byte, offset int) (payload []byte, next int, ok bool) {
	if offset+4 > len(b) {
		return nil, offset, false
	}
	n := binary.LittleEndian.Uint32(b[offset : offset+4])
	offset += 4
	if offset+int(n) > len(b) {
		return nil, offset, false
	}
	return b[offset : offset+int(n)], offset + int(n), true
}

func TestHandleECShardRequest_Basic(t *testing.T) {
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

	shardIndex := uint16(0)
	req := make([]byte, 0, CE137RequestSize)
	req = append(req, erasureRoot...)
	req = append(req, byte(shardIndex), byte(shardIndex>>8))
	stream := newMockStream(req)

	fakeBlockchain := SetupFakeBlockchain()

	err = HandleECShardRequest(fakeBlockchain, stream)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	resp := stream.w.Bytes()
	offset := 0

	bundleShard, offset, ok := readFramedMessage(resp, offset)
	if !ok || len(bundleShard) == 0 {
		t.Fatalf("response message 1 (bundle shard): missing or empty")
	}

	segmentShards, offset, ok := readFramedMessage(resp, offset)
	if !ok || len(segmentShards) == 0 {
		t.Fatalf("response message 2 (segment shards): missing or empty")
	}

	justification, _, ok := readFramedMessage(resp, offset)
	if !ok || len(justification) == 0 {
		t.Fatalf("response message 3 (justification): missing or empty")
	}

	// Verify bundle shard matches expected (extractBundleShard for shard 0)
	expectedShard, err := extractBundleShard(testBundle, uint32(shardIndex))
	if err != nil {
		t.Fatalf("extractBundleShard: %v", err)
	}
	if !bytes.Equal(bundleShard, expectedShard) {
		t.Errorf("bundle shard mismatch: expected %d bytes, got %d", len(expectedShard), len(bundleShard))
	}

	// Justification should have 0x00 discriminator or valid structure
	if justification[0] != 0x00 && justification[0] != 0x01 {
		t.Errorf("justification first byte should be 0x00 or 0x01, got %x", justification[0])
	}
}
