package ce

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	erasurecoding "github.com/New-JAMneration/JAM-Protocol/pkg/erasure_coding"
)

// HandleECShardRequest_Guarantor handles an assurer's request for their erasure coded shards from a guarantor.
// Role: [Assurer -> Guarantor]
//
// [TODO-Validation]
// 1. Remove mock data and check work-report and erasure-coded bundle.
// 2. Use data retrieved from (1) then calculate Merkle proof.
func HandleECShardRequest_Guarantor(stream *quic.Stream) error {
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

	// Look up the work package bundle shard for the given erasure root and shard index
	bundleShard, err := lookupWorkPackageBundleShard(erasureRoot, shardIndex)
	if err != nil {
		return fmt.Errorf("failed to lookup work package bundle shard: %w", err)
	}

	bundleShardLen := uint32(len(bundleShard))
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, bundleShardLen)
	if _, err := stream.Write(lenBuf); err != nil {
		return fmt.Errorf("failed to write bundle shard length: %w", err)
	}
	if _, err := stream.Write(bundleShard); err != nil {
		return fmt.Errorf("failed to write bundle shard: %w", err)
	}

	segmentShards, err := extractSegmentShardsForShardIndex(erasureRoot, shardIndex)
	if err != nil {
		return fmt.Errorf("failed to extract segment shards: %w", err)
	}

	segmentShardsCount := uint32(len(segmentShards))
	countBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(countBuf, segmentShardsCount)
	if _, err := stream.Write(countBuf); err != nil {
		return fmt.Errorf("failed to write segment shards count: %w", err)
	}

	for i, segmentShard := range segmentShards {
		shardLen := uint32(len(segmentShard))
		shardLenBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(shardLenBuf, shardLen)
		if _, err := stream.Write(shardLenBuf); err != nil {
			return fmt.Errorf("failed to write segment shard %d length: %w", i, err)
		}
		if _, err := stream.Write(segmentShard); err != nil {
			return fmt.Errorf("failed to write segment shard %d: %w", i, err)
		}
	}

	justification, err := construct137Justification(erasureRoot, shardIndex)
	if err != nil {
		return fmt.Errorf("failed to construct justification: %w", err)
	}
	if err := stream.WriteMessage(justification); err != nil {
		return fmt.Errorf("failed to write justification: %w", err)
	}

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return fmt.Errorf("failed to write FIN marker: %w", err)
	}
	return nil
}

// HandleECShardRequest_Assurer handles an assurer's request for erasure coded shards from a guarantor.
// Role: [Assurer -> Guarantor]
func HandleECShardRequest_Assurer(stream *quic.Stream, erasureRoot []byte, shardIndex uint32) error {
	// Send request: Erasure-Root (32 bytes) + Shard Index (4 bytes) + 'FIN' (3 bytes)
	request := make([]byte, 32+4+3)
	copy(request[:32], erasureRoot)
	binary.LittleEndian.PutUint32(request[32:36], shardIndex)
	copy(request[36:39], []byte("FIN"))

	if _, err := stream.Write(request); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Receive response from Guarantor
	lenBuf := make([]byte, 4)
	if err := stream.ReadFull(lenBuf); err != nil {
		return fmt.Errorf("failed to read bundle shard length: %w", err)
	}
	bundleShardLen := binary.LittleEndian.Uint32(lenBuf)

	bundleShard := make([]byte, bundleShardLen)
	if err := stream.ReadFull(bundleShard); err != nil {
		return fmt.Errorf("failed to read bundle shard: %w", err)
	}

	segmentShardsCountBuf := make([]byte, 4)
	if err := stream.ReadFull(segmentShardsCountBuf); err != nil {
		return fmt.Errorf("failed to read segment shards count: %w", err)
	}
	segmentShardsCount := binary.LittleEndian.Uint32(segmentShardsCountBuf)

	var segmentShards [][]byte
	for i := uint32(0); i < segmentShardsCount; i++ {
		// Read segment shard length
		shardLenBuf := make([]byte, 4)
		if err := stream.ReadFull(shardLenBuf); err != nil {
			return fmt.Errorf("failed to read segment shard %d length: %w", i, err)
		}
		shardLen := binary.LittleEndian.Uint32(shardLenBuf)

		// Read segment shard data
		segmentShard := make([]byte, shardLen)
		if err := stream.ReadFull(segmentShard); err != nil {
			return fmt.Errorf("failed to read segment shard %d: %w", i, err)
		}
		segmentShards = append(segmentShards, segmentShard)
	}

	justification, err := stream.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read justification: %w", err)
	}

	finBuf := make([]byte, 3)
	if err := stream.ReadFull(finBuf); err != nil {
		return fmt.Errorf("failed to read FIN marker: %w", err)
	}
	if string(finBuf) != "FIN" {
		return fmt.Errorf("expected FIN marker, got %q", string(finBuf))
	}

	// TODO: Process the received data
	// - Validate bundle shard
	// - Validate segment shards
	// - Verify justification
	_ = justification

	return stream.Close()
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

// extractSegmentShardsForShardIndex extracts all segment shards for a given shard index.
func extractSegmentShardsForShardIndex(erasureRoot []byte, shardIndex uint32) ([][]byte, error) {
	bundle, err := lookupWorkPackageBundle(erasureRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup work package bundle: %w", err)
	}

	encoder := types.NewEncoder()
	bundleBytes, err := encoder.Encode(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to encode bundle: %w", err)
	}

	k := (len(bundleBytes) + 683) / 684
	ec, err := erasurecoding.NewErasureCoding(types.DataShards, types.TotalShards, k)
	if err != nil {
		return nil, fmt.Errorf("failed to create erasure coding: %w", err)
	}

	shards, err := ec.Encode(bundleBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encode bundle into shards: %w", err)
	}

	if int(shardIndex) >= len(shards) {
		return nil, fmt.Errorf("shard index %d out of range (max: %d)", shardIndex, len(shards)-1)
	}

	requestedShard := shards[shardIndex]

	// Extract all segment shards (32 bytes each)
	var segmentShards [][]byte
	segmentSize := 32
	for startPos := 0; startPos < len(requestedShard); startPos += segmentSize {
		endPos := startPos + segmentSize
		if endPos > len(requestedShard) {
			endPos = len(requestedShard)
		}

		segmentShard := requestedShard[startPos:endPos]
		segmentShards = append(segmentShards, segmentShard)
	}

	return segmentShards, nil
}
