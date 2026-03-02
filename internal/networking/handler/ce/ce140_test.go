package ce

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

func TestHandleSegmentShardRequestWithJustification(t *testing.T) {
	// Erasure root (32 bytes) + shard index (4 bytes) + segment indices length (2 bytes) + segment indices + FIN
	erasureRoot := make([]byte, 32)
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

	mockStream := newMockStream(buf.Bytes())

	err := HandleSegmentShardRequestWithJustification(nil, &quic.Stream{Stream: mockStream})
	if err != nil {
		t.Fatalf("HandleSegmentShardRequestWithJustification failed: %v", err)
	}

	response := mockStream.w.Bytes()
	if len(response) < 64 {
		t.Errorf("Expected response to contain segment shards, got length %d", len(response))
	}
}
