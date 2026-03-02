package ce

import (
	"bytes"
	"testing"
)

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
	if !bytes.HasPrefix(resp[offset:], bundle.BundleShard) {
		t.Fatalf("response does not start with bundle shard")
	}
	offset += len(bundle.BundleShard)
	if !bytes.HasPrefix(resp[offset:], bundle.SegmentShards[0]) {
		t.Fatalf("response does not contain segment shard 1 at expected position")
	}
	offset += len(bundle.SegmentShards[0])
	if !bytes.HasPrefix(resp[offset:], bundle.SegmentShards[1]) {
		t.Fatalf("response does not contain segment shard 2 at expected position")
	}
	offset += len(bundle.SegmentShards[1])
	if !bytes.Equal(resp[offset:], bundle.Justification) {
		t.Fatalf("response does not contain justification at expected position")
	}
}
