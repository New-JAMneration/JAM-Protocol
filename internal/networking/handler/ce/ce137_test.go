package ce

import (
	"bytes"
	"encoding/binary"
	"testing"
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
	erasureRoot := []byte("fake-erasure-root-32bytes-long!!")
	// Test parameters for AssignShardIndex
	coreIndex := 1
	recoveryThreshold := 5
	validatorIndex := 3
	totalValidators := 10
	shardIndex := uint32(AssignShardIndex(coreIndex, recoveryThreshold, validatorIndex, totalValidators))

	computedIndex := AssignShardIndex(coreIndex, recoveryThreshold, validatorIndex, totalValidators)
	if int(shardIndex) != computedIndex {
		t.Fatalf("shardIndex (%d) does not match AssignShardIndex result (%d)", shardIndex, computedIndex)
	}

	// Prepare a real WorkReportBundle with mock data
	bundle := &CE137Payload{
		BundleShard: []byte("BUNDLE_SHARD_MOCK"),
		SegmentShards: [][]byte{
			[]byte("SEGMENT_SHARD1_MOCK"),
			[]byte("SEGMENT_SHARD2_MOCK"),
		},
		Justification: append([]byte{0x00}, make([]byte, 32)...), // 0 discriminator + 32 zero bytes
	}
	lookup := func(root []byte) (*CE137Payload, bool) {
		if bytes.Equal(root, erasureRoot) {
			return bundle, true
		}
		return nil, false
	}

	// Prepare request: erasureRoot (32 bytes) + shardIndex (2 bytes u16 LE); peer closes after
	req := make([]byte, 0, 32+2)
	req = append(req, erasureRoot...)
	req = append(req, byte(shardIndex), byte(shardIndex>>8))
	stream := newMockStream(req)

	err := HandleECShardRequest(stream, lookup)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	resp := stream.w.Bytes()
	offset := 0
	msg, offset, ok := readFramedMessage(resp, offset)
	if !ok || !bytes.Equal(msg, bundle.BundleShard) {
		t.Fatalf("response message 1 (bundle shard) mismatch")
	}
	msg, offset, ok = readFramedMessage(resp, offset)
	if !ok || !bytes.Equal(msg, bundle.SegmentShards[0]) {
		t.Fatalf("response message 2 (segment shard 1) mismatch")
	}
	msg, offset, ok = readFramedMessage(resp, offset)
	if !ok || !bytes.Equal(msg, bundle.SegmentShards[1]) {
		t.Fatalf("response message 3 (segment shard 2) mismatch")
	}
	msg, _, ok = readFramedMessage(resp, offset)
	if !ok || !bytes.Equal(msg, bundle.Justification) {
		t.Fatalf("response message 4 (justification) mismatch")
	}
}
