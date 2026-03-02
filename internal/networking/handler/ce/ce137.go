package ce

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
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
func HandleECShardRequest(bc blockchain.Blockchain, stream quic.MessageStream) error {
	// Read erasure-root (HashSize) + shard index (U16Size); FIN = peer closes send half (EOF)
	buf := make([]byte, CE137RequestSize)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return err
	}
	if err := expectRemoteFIN(stream); err != nil {
		return err
	}
	erasureRoot := buf[:HashSize]
	shardIndex := uint32(binary.LittleEndian.Uint16(buf[HashSize:CE137RequestSize]))

	bundle, err := lookupWorkPackageBundle(bc, erasureRoot)
	if err != nil {
		return fmt.Errorf("lookup work-package bundle: %w", err)
	}

	bundleShard, err := extractBundleShard(bundle, shardIndex)
	if err != nil {
		return fmt.Errorf("extract bundle shard: %w", err)
	}
	if len(bundleShard) == 0 {
		return fmt.Errorf("bundle shard is empty")
	}

	// Segment shards: all HashSize-byte segments within the bundle shard
	numSegments := (len(bundleShard) + HashSize - 1) / HashSize
	segmentIndices := make([]uint16, numSegments)
	for i := range segmentIndices {
		segmentIndices[i] = uint16(i)
	}
	segmentShards, err := extractSegmentShards(bundle, shardIndex, segmentIndices)
	if err != nil {
		return fmt.Errorf("extract segment shards: %w", err)
	}

	justification, err := constructAuditJustification(bc, erasureRoot, shardIndex, bundleShard)
	if err != nil {
		return fmt.Errorf("construct justification: %w", err)
	}

	if err := stream.WriteMessage(bundleShard); err != nil {
		return err
	}
	if err := stream.WriteMessage(segmentShards); err != nil {
		return err
	}
	if err := stream.WriteMessage(justification); err != nil {
		return err
	}
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
