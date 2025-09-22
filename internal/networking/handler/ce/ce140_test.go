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

	shardIndexBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(shardIndexBytes, shardIndex)
	buf.Write(shardIndexBytes)

	segmentIndicesLenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(segmentIndicesLenBytes, segmentIndicesLen)
	buf.Write(segmentIndicesLenBytes)

	for _, index := range segmentIndices {
		indexBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(indexBytes, index)
		buf.Write(indexBytes)
	}

	buf.Write([]byte("FIN"))

	mockStream := newMockStream(buf.Bytes())

	err := HandleSegmentShardRequestWithJustification_Assurer(nil, &quic.Stream{Stream: mockStream})
	if err != nil {
		t.Fatalf("HandleSegmentShardRequestWithJustification failed: %v", err)
	}

	response := mockStream.w.Bytes()

	// The response should contain the concatenated segment shards (2 segments * 32 bytes = 64 bytes minimum)
	if len(response) < 64 {
		t.Errorf("Expected response to contain segment shards, got length %d", len(response))
	}

	if !bytes.HasSuffix(response, []byte("FIN")) {
		t.Error("Response does not end with FIN")
	}

	// Each justification should contain the combined justifications from CE137, bundle shard hash, and Merkle co-path
	// For 3 segment indices, we should have 3 justifications
	// The justifications should be after the segment shards and before FIN
	finIndex := bytes.LastIndex(response, []byte("FIN"))
	if finIndex == -1 {
		t.Error("Response does not end with FIN")
	}

	if finIndex <= 64 {
		t.Error("Expected justifications between segment shards and FIN")
	}
}
