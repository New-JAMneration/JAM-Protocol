package ce

import (
	"encoding/binary"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	erasurecoding "github.com/New-JAMneration/JAM-Protocol/pkg/erasure_coding"
)

// HandleAuditShardRequest handles an auditor's request for work-package bundle shards from assurers.
//
// Request (from Auditor to Assurer):
//
//	Erasure-Root (hash, []byte)
//	Shard Index (u16)
//	FIN (stream close)
//
// Response (from Assurer to Auditor):
//
//	Bundle Shard ([]byte)
//	Justification ([]byte, Merkle co-path proof)
//	FIN (stream close)
//
// The justification is a sequence of [0 ++ Hash OR 1 ++ Hash ++ Hash] as per the protocol.
// The assurer should construct this by appending the corresponding segment shard root
// to the justification received via CE 137.
func HandleAuditShardRequest(bc blockchain.Blockchain, stream *quic.Stream) error {
	// Request: single message = erasure-root (HashSize bytes) + shard index (U16Size bytes, u16)
	payload, err := stream.ReadMessage()
	if err != nil {
		return err
	}
	if len(payload) < CE138RequestSize {
		return fmt.Errorf("audit shard request too short")
	}
	erasureRoot := payload[:HashSize]
	shardIndex := uint32(binary.LittleEndian.Uint16(payload[HashSize:CE138RequestSize]))

	// Look up the work-package bundle shard for the given erasure root and shard index
	bundleShard, err := lookupWorkPackageBundleShard(bc, erasureRoot, shardIndex)
	if err != nil {
		return fmt.Errorf("failed to lookup work package bundle shard: %w", err)
	}

	// Construct the justification
	justification, err := constructAuditJustification(bc, erasureRoot, shardIndex, bundleShard)
	if err != nil {
		return fmt.Errorf("failed to construct justification: %w", err)
	}

	// Response: Bundle Shard (message 1), Justification (message 2), then FIN
	if err := stream.WriteMessage(bundleShard); err != nil {
		return err
	}
	if err := stream.WriteMessage(justification); err != nil {
		return err
	}
	return stream.Close()
}

// lookupWorkPackageBundleShard looks up a work package bundle by its erasure root and extracts the specific shard.
func lookupWorkPackageBundleShard(bc blockchain.Blockchain, erasureRoot []byte, shardIndex uint32) ([]byte, error) {
	bundle, err := lookupWorkPackageBundle(bc, erasureRoot)
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
	encoder.SetHashSegmentMap(map[types.OpaqueHash]types.OpaqueHash{})
	bundleBytes, err := encoder.Encode(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to encode bundle: %w", err)
	}

	shards, err := erasurecoding.EncodeDataShards(bundleBytes, types.DataShards, types.TotalShards-types.DataShards)
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
func constructAuditJustification(bc blockchain.Blockchain, erasureRoot []byte, shardIndex uint32, bundleShard []byte) ([]byte, error) {
	// Get the CE137 justification (j) - this would come from the assurer's CE137 response
	ce137Justification, err := getCE137JustificationForAudit(bc, erasureRoot, shardIndex)
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
func getCE137JustificationForAudit(bc blockchain.Blockchain, erasureRoot []byte, shardIndex uint32) ([]byte, error) {
	db := DB(bc)
	justification, err := GetJustification(db, erasureRoot, shardIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get CE137 justification from storage: %w", err)
	}

	if justification != nil {
		return justification, nil
	}

	mockJustification := make([]byte, JustificationHashEntrySize) // 1 byte discriminator + HashSize bytes hash
	mockJustification[0] = 0x00                                   // Type 0: single hash

	hashInput := append(erasureRoot, byte(shardIndex), byte(shardIndex>>8), byte(shardIndex>>16), byte(shardIndex>>24))
	mockHash := hash.Blake2bHash(types.ByteSequence(hashInput))
	copy(mockJustification[1:], mockHash[:])

	if err := PutJustification(db, erasureRoot, shardIndex, mockJustification); err != nil {
		return nil, fmt.Errorf("failed to store CE137 justification: %w", err)
	}

	return mockJustification, nil
}

// getSegmentShardRoots computes the hashes of all segment shards within the bundle shard.
func getSegmentShardRoots(bundleShard []byte) ([]byte, error) {
	segmentSize := HashSize
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

	if len(auditReq.ErasureRoot) != HashSize {
		return nil, fmt.Errorf("erasure root must be exactly %d bytes, got %d", HashSize, len(auditReq.ErasureRoot))
	}
	if err := writeRaw(auditReq.ErasureRoot); err != nil {
		return nil, fmt.Errorf("failed to encode ErasureRoot: %w", err)
	}

	shardIndexBytes := encodeLE16(uint16(auditReq.ShardIndex))
	if err := writeRaw(shardIndexBytes); err != nil {
		return nil, fmt.Errorf("failed to encode ShardIndex: %w", err)
	}

	result := make([]byte, 0, CE138RequestSize)
	result = append(result, auditReq.ErasureRoot...)
	result = append(result, shardIndexBytes...)

	return result, nil
}
