package ce

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

// HandleECShardRequest handles an assurer's request for their erasure coded shards from a guarantor.
//
// [TODO-Validation]
// 1. Remove mock data and check work-report and erasure-coded bundle.
// 2. Use data retrieved from (1) then calculate Merkle proof.
func HandleECShardRequest(stream *quic.Stream, lookup func(erasureRoot []byte) (*CE137GuarantorPayload, bool)) error {
	// From Assurer: 32 bytes erasure-root + 4 bytes shard index + 'FIN'
	buf := make([]byte, 32+4+3)
	if err := stream.ReadFull(buf); err != nil {
		return err
	}
	erasureRoot := buf[:32]
	shardIndex := binary.LittleEndian.Uint32(buf[32:36])
	fin := buf[36:]
	if string(fin) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	// Guarantor
	// [TODO] Check real bundle shsard format
	bundleShard := []byte("BUNDLE_SHARD_MOCK")
	if _, err := stream.Write(bundleShard); err != nil {
		return err
	}
	// [TODO] Check real segment shard format (a list)
	segmentShard1 := []byte("SEGMENT_SHARD1_MOCK")
	segmentShard2 := []byte("SEGMENT_SHARD2_MOCK")
	if _, err := stream.Write(segmentShard1); err != nil {
		return err
	}
	if _, err := stream.Write(segmentShard2); err != nil {
		return err
	}

	justification, err := construct137Justification(erasureRoot, shardIndex)
	if err != nil {
		return fmt.Errorf("failed to construct justification: %w", err)
	}
	if _, err := stream.Write(justification); err != nil {
		return err
	}
	if _, err := stream.Write([]byte("FIN")); err != nil {
		return err
	}
	return nil
}

type CE137AssurerPayload struct {
	ErasureRoot []byte
	ShardIndex  uint16
}

type CE137GuarantorPayload struct {
	BundleShard   []byte
	SegmentShards [][]byte
	Justification []byte
}

func (h *DefaultCERequestHandler) encodeShardDistribution(message interface{}) ([]byte, error) {
	shardDist, ok := message.(*CE137GuarantorPayload)
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

// construct137Justification constructs justification using only T(s, i, H) function
func construct137Justification(erasureRoot []byte, shardIndex uint32) ([]byte, error) {
	// Get work package bundle from erasure root
	bundle, err := lookupWorkPackageBundle(erasureRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup work package bundle: %w", err)
	}

	// Get segment shard sequence (s) for the given shard index
	segmentShardSequence, err := getSegmentShardSequence(bundle, shardIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get segment shard sequence: %w", err)
	}

	segmentIndex := uint16(0)
	// Construct T(s, i, H) - the Merkle tree co-path for the segment
	merkleCoPath, err := constructMerkleCoPath(segmentShardSequence, segmentIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to construct Merkle co-path: %w", err)
	}

	return merkleCoPath, nil
}
