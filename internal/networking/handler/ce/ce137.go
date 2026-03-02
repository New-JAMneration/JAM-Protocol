package ce

import (
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

// HandleECShardRequest handles an assurer's request for their erasure coded shards from a guarantor.
//
// Request (from Assurer to Guarantor):
//
//	Erasure-Root (hash, []byte)
//	Shard Index (u16)
//	FIN (stream close)
//
// Response (from Guarantor to Assurer):
//
//	Bundle Shard ([]byte)
//	[Segment Shard] ([]byte, exported and proof segment shards with the given index)
//	Justification ([]byte, Merkle co-path proof)
//	FIN (stream close)
//
// The justification is a sequence of [0 ++ Hash OR 1 ++ Hash ++ Hash] as per the protocol.
//
// This function should use AssignShardIndex to determine the correct shard index.
func HandleECShardRequest(stream quic.MessageStream, lookup func(erasureRoot []byte) (*CE137Payload, bool)) error {
	// Read erasure-root (32 bytes) + shard index (2 bytes, u16); FIN = peer closes send half (EOF)
	buf := make([]byte, 32+2)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return err
	}
	if err := expectRemoteFIN(stream); err != nil {
		return err
	}

	// Response: each part sent as a framed message (4-byte LE length + payload)
	bundleShard := []byte("BUNDLE_SHARD_MOCK")
	if err := stream.WriteMessage(bundleShard); err != nil {
		return err
	}
	// Write mock segment shards (simulate 2 segments)
	segmentShard1 := []byte("SEGMENT_SHARD1_MOCK")
	segmentShard2 := []byte("SEGMENT_SHARD2_MOCK")
	if err := stream.WriteMessage(segmentShard1); err != nil {
		return err
	}
	if err := stream.WriteMessage(segmentShard2); err != nil {
		return err
	}
	justification := append([]byte{0x00}, make([]byte, 32)...) // 0 discriminator + 32 zero bytes
	if err := stream.WriteMessage(justification); err != nil {
		return err
	}
	// Send FIN by closing write half
	return stream.Close()
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
