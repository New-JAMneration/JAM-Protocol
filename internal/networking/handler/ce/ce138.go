package ce

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	erasurecoding "github.com/New-JAMneration/JAM-Protocol/pkg/erasure_coding"
)

// Role: [Auditor -> Assurer]
func HandleAuditShardRequest_Assurer(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// Read erasure-root (32 bytes) + shard index (4 bytes) + 'FIN' (3 bytes)
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

	// Look up the work-package bundle shard for the given erasure root and shard index
	bundleShard, err := lookupWorkPackageBundleShard(erasureRoot, shardIndex)
	if err != nil {
		return fmt.Errorf("failed to lookup work package bundle shard: %w", err)
	}

	if err := stream.WriteMessage(bundleShard); err != nil {
		return fmt.Errorf("failed to write bundle shard: %w", err)
	}

	justification, err := constructAuditJustification(erasureRoot, shardIndex, bundleShard)
	if err != nil {
		return fmt.Errorf("failed to construct justification: %w", err)
	}

	if err := stream.WriteMessage(justification); err != nil {
		return fmt.Errorf("failed to write justification: %w", err)
	}

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return err
	}

	return stream.Close()
}

func HandleAuditShardRequest_Auditor(blockchain blockchain.Blockchain, stream *quic.Stream, erasureRoot []byte, shardIndex uint32) error {
	// Send request: Erasure-Root (32 bytes) + Shard Index (4 bytes) + 'FIN' (3 bytes)
	if len(erasureRoot) != 32 {
		return fmt.Errorf("erasure root must be 32 bytes, got %d", len(erasureRoot))
	}

	request := make([]byte, 32+4+3)
	copy(request[:32], erasureRoot)
	binary.LittleEndian.PutUint32(request[32:36], shardIndex)
	copy(request[36:39], []byte("FIN"))

	if _, err := stream.Write(request); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Read response: length-prefixed bundle shard, justification (ReadMessage), then FIN
	lenBuf := make([]byte, 4)
	if err := stream.ReadFull(lenBuf); err != nil {
		return fmt.Errorf("failed to read bundle shard length: %w", err)
	}
	bundleShardLen := binary.LittleEndian.Uint32(lenBuf)
	bundleShard := make([]byte, bundleShardLen)
	if err := stream.ReadFull(bundleShard); err != nil {
		return fmt.Errorf("failed to read bundle shard: %w", err)
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

	_ = bundleShard
	_ = justification

	return stream.Close()
}

// lookupWorkPackageBundleShard looks up a work package bundle by its erasure root and extracts the specific shard.
func lookupWorkPackageBundleShard(erasureRoot []byte, shardIndex uint32) ([]byte, error) {
	bundle, err := lookupWorkPackageBundle(erasureRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup work package bundle: %w", err)
	}

	bundleShard, err := extractBundleShard(bundle, shardIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to extract bundle shard: %w", err)
	}

	return bundleShard, nil
}

// extractBundleShard extracts the specific shard from the work package bundle using erasure coding.
func extractBundleShard(bundle *types.WorkPackageBundle, shardIndex uint32) ([]byte, error) {
	encoder := types.NewEncoder()
	bundleBytes, err := encoder.Encode(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to encode bundle: %w", err)
	}

	ec, err := erasurecoding.NewErasureCoding(types.DataShards, types.TotalShards, (len(bundleBytes)+683)/684)
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

	return shards[shardIndex], nil
}

// constructAuditJustification constructs the justification for the audit shard request.
// The justification should be constructed by appending segment shard roots to the CE137 justification.
func constructAuditJustification(erasureRoot []byte, shardIndex uint32, bundleShard []byte) ([]byte, error) {
	// Get the CE137 justification (j) - this would come from the assurer's CE137 response
	ce137Justification, err := getCE137JustificationForAudit(erasureRoot, shardIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get CE137 justification: %w", err)
	}

	bundleShardHash := hash.Blake2bHash(types.ByteSequence(bundleShard))

	segmentShardRoots, err := getSegmentShardRoots(bundleShard)
	if err != nil {
		return nil, fmt.Errorf("failed to get segment shard roots: %w", err)
	}

	// Construct the combined justification: j ^ [b] ^ segment_shard_roots
	combinedJustification := combineAuditJustifications(ce137Justification, bundleShardHash[:], segmentShardRoots)

	return combinedJustification, nil
}

// getCE137JustificationForAudit retrieves the CE137 justification for the given erasure root and shard index.
func getCE137JustificationForAudit(erasureRoot []byte, shardIndex uint32) ([]byte, error) {
	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis backend: %w", err)
	}

	justification, err := redisBackend.GetJustification(context.Background(), erasureRoot, shardIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get CE137 justification from storage: %w", err)
	}

	if justification != nil {
		return justification, nil
	}

	mockJustification := make([]byte, 33) // 1 byte discriminator + 32 bytes hash
	mockJustification[0] = 0x00           // Type 0: single hash

	hashInput := append(erasureRoot, byte(shardIndex), byte(shardIndex>>8), byte(shardIndex>>16), byte(shardIndex>>24))
	mockHash := hash.Blake2bHash(types.ByteSequence(hashInput))
	copy(mockJustification[1:], mockHash[:])

	err = redisBackend.PutJustification(context.Background(), erasureRoot, shardIndex, mockJustification)
	if err != nil {
		return nil, fmt.Errorf("failed to store CE137 justification: %w", err)
	}

	return mockJustification, nil
}

// getSegmentShardRoots computes the hashes of all segment shards within the bundle shard.
func getSegmentShardRoots(bundleShard []byte) ([]byte, error) {
	segmentSize := 32
	var segmentRoots []byte

	// Split the bundle shard into segments and hash each segment
	for i := 0; i < len(bundleShard); i += segmentSize {
		end := i + segmentSize
		if end > len(bundleShard) {
			end = len(bundleShard)
		}

		segment := bundleShard[i:end]
		segmentHash := hash.Blake2bHash(types.ByteSequence(segment))
		segmentRoots = append(segmentRoots, segmentHash[:]...)
	}

	return segmentRoots, nil
}

// combineAuditJustifications constructs the proper justification format.
// The justification should be a sequence of [0 ++ Hash OR 1 ++ Hash ++ Hash] as per the protocol.
// We append the bundle shard hash and segment shard roots to the CE137 justification.
func combineAuditJustifications(ce137Justification, bundleShardHash, segmentShardRoots []byte) []byte {
	result := make([]byte, len(ce137Justification))
	copy(result, ce137Justification)

	// Append the bundle shard hash with discriminator 0x00
	result = append(result, 0x00) // discriminator for single hash
	result = append(result, bundleShardHash...)

	// Append segment shard roots
	if len(segmentShardRoots) > 0 {
		result = append(result, 0x00) // discriminator for single hash
		result = append(result, segmentShardRoots...)
	}

	return result
}

type CE138Payload struct {
	ErasureRoot []byte
	ShardIndex  uint32
}

func (h *DefaultCERequestHandler) encodeAuditShardRequest(message interface{}) ([]byte, error) {
	auditReq, ok := message.(*CE138Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for AuditShardRequest: %T", message)
	}

	encoder := types.NewEncoder()

	writeRaw := func(b []byte) error {
		for _, v := range b {
			if err := encoder.WriteByte(v); err != nil {
				return err
			}
		}
		return nil
	}

	if len(auditReq.ErasureRoot) != 32 {
		return nil, fmt.Errorf("erasure root must be exactly 32 bytes, got %d", len(auditReq.ErasureRoot))
	}
	if err := writeRaw(auditReq.ErasureRoot); err != nil {
		return nil, fmt.Errorf("failed to encode ErasureRoot: %w", err)
	}

	shardIndexBytes := []byte{
		byte(auditReq.ShardIndex),
		byte(auditReq.ShardIndex >> 8),
		byte(auditReq.ShardIndex >> 16),
		byte(auditReq.ShardIndex >> 24),
	}
	if err := writeRaw(shardIndexBytes); err != nil {
		return nil, fmt.Errorf("failed to encode ShardIndex: %w", err)
	}

	result := make([]byte, 0, 36)
	result = append(result, auditReq.ErasureRoot...)
	result = append(result, shardIndexBytes...)

	return result, nil
}
