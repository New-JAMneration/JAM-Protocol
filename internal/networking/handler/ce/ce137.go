package ce

import (
	"errors"
	"fmt"
	"io"
)

// HandleECShardRequest handles an assurer's request for their erasure coded shards from a guarantor.
//
// [TODO-Validation]
// 1. Remove mock data and check work-report and erasure-coded bundle.
// 2. Use data retrieved from (1) then calculate Merkle proof.
func HandleECShardRequest(stream io.ReadWriter, lookup func(erasureRoot []byte) (*CE137Payload, bool)) error {
	// Read erasure-root (32 bytes) + shard index (4 bytes) + 'FIN' (3 bytes)
	buf := make([]byte, 32+4+3)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return err
	}
	fin := buf[36:]
	if string(fin) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	// Write mock bundle shard
	bundleShard := []byte("BUNDLE_SHARD_MOCK")
	if _, err := stream.Write(bundleShard); err != nil {
		return err
	}
	// Write mock segment shards (simulate 2 segments)
	segmentShard1 := []byte("SEGMENT_SHARD1_MOCK")
	segmentShard2 := []byte("SEGMENT_SHARD2_MOCK")
	if _, err := stream.Write(segmentShard1); err != nil {
		return err
	}
	if _, err := stream.Write(segmentShard2); err != nil {
		return err
	}
	justification := append([]byte{0x00}, make([]byte, 32)...) // 0 discriminator + 32 zero bytes
	if _, err := stream.Write(justification); err != nil {
		return err
	}
	if _, err := stream.Write([]byte("FIN")); err != nil {
		return err
	}
	return nil
}

type CE137Payload struct {
	BundleShard   []byte
	SegmentShards [][]byte
	Justification []byte
}

func (h *DefaultCERequestHandler) encodeShardDistribution(message interface{}) ([]byte, error) {
	shardDist, ok := message.(*CE137Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for ShardDistribution: %T", message)
	}

	if shardDist == nil {
		return nil, fmt.Errorf("nil payload for ShardDistribution")
	}

	requestType := byte(ShardDistribution)

	segmentShardsLen := 0
	for _, shard := range shardDist.SegmentShards {
		segmentShardsLen += len(shard)
	}

	totalLen := 1 + // request type
		4 + // length of bundle shard
		len(shardDist.BundleShard) + // bundle shard data
		4 + // number of segment shards
		segmentShardsLen + // segment shards data
		len(shardDist.Justification) // justification data

	result := make([]byte, 0, totalLen)

	result = append(result, requestType)
	result = append(result, encodeLE32(uint32(len(shardDist.BundleShard)))...)
	result = append(result, shardDist.BundleShard...)
	result = append(result, encodeLE32(uint32(len(shardDist.SegmentShards)))...)

	for _, shard := range shardDist.SegmentShards {
		result = append(result, encodeLE32(uint32(len(shard)))...)
		result = append(result, shard...)
	}

	result = append(result, shardDist.Justification...)

	return result, nil
}
