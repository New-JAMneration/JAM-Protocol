package ce

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merkle_tree"
	erasurecoding "github.com/New-JAMneration/JAM-Protocol/pkg/erasure_coding"
)

// HandleSegmentShardRequestWithJustification handles a guarantor's request for segment shards from assurers.
// This is protocol variant 140 where justification is provided for each returned segment shard.
//
// Request (from Guarantor to Assurer):
//
//	Erasure-Root (hash, []byte)
//	Shard Index (uint32)
//	Segment Indices Length (uint16)
//	Segment Indices ([]uint16)
//	'FIN' (3 bytes)
//
// Response (from Assurer to Guarantor):
//
//	Segment Shards ([]byte, concatenated)
//	Justifications ([]byte, concatenated justifications for each segment shard)
//	'FIN' (3 bytes)
//
// The justification for a segment shard should be the co-path from the erasure-root to the shard,
// given by: j ^ [b] ^ T(s, i, H)
// where:
// - j is the relevant justification provided to the assurer via CE 137
// - b is the corresponding work-package bundle shard hash
// - s is the full sequence of segment shards with the given shard index
// - i is the segment index
// - H is the Blake 2b hash function
// - T is as defined in the General Merklization appendix of the GP
//
// The number of segment shards requested should not exceed 2W_M (W_M=3072).
func HandleSegmentShardRequestWithJustification(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// Read erasure-root (32 bytes) + shard index (4 bytes) + segment indices length (2 bytes)
	buf := make([]byte, 32+4+2)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return err
	}
	erasureRoot := buf[:32]
	shardIndex := binary.LittleEndian.Uint32(buf[32:36])
	segmentIndicesLen := binary.LittleEndian.Uint16(buf[36:38])

	// Validate segment indices length (should not exceed 2W_M = 6144)
	if segmentIndicesLen > 6144 {
		return errors.New("segment indices length exceeds maximum allowed (2W_M)")
	}

	segmentIndices := make([]uint16, segmentIndicesLen)
	if segmentIndicesLen > 0 {
		indicesBuf := make([]byte, segmentIndicesLen*2)
		if _, err := io.ReadFull(stream, indicesBuf); err != nil {
			return err
		}
		for i := uint16(0); i < segmentIndicesLen; i++ {
			segmentIndices[i] = binary.LittleEndian.Uint16(indicesBuf[i*2 : (i+1)*2])
		}
	}

	finBuf := make([]byte, 3)
	if _, err := io.ReadFull(stream, finBuf); err != nil {
		return err
	}
	if string(finBuf) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	bundle, err := lookupWorkPackageBundle(erasureRoot)
	if err != nil {
		return fmt.Errorf("failed to lookup work package bundle: %w", err)
	}

	segmentShards, err := extractSegmentShards(bundle, shardIndex, segmentIndices)
	if err != nil {
		return fmt.Errorf("failed to extract segment shards: %w", err)
	}

	if _, err := stream.Write(segmentShards); err != nil {
		return err
	}

	for _, segmentIndex := range segmentIndices {
		justification, err := constructJustification(bundle, erasureRoot, shardIndex, segmentIndex)
		if err != nil {
			return fmt.Errorf("failed to construct justification for segment %d: %w", segmentIndex, err)
		}
		if _, err := stream.Write(justification); err != nil {
			return err
		}
	}

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return err
	}

	return stream.Close()
}

// constructJustification constructs the justification for a segment shard using the formula:
// j ^ [b] ^ T(s, i, H)
func constructJustification(bundle *types.WorkPackageBundle, erasureRoot []byte, shardIndex uint32, segmentIndex uint16) ([]byte, error) {
	// Get the CE137 justification (j) - this would come from the assurer's CE137 response
	ce137Justification, err := getCE137Justification(erasureRoot, shardIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get CE137 justification: %w", err)
	}

	// Get the bundle shard hash (b)
	bundleShardHash, err := getBundleShardHash(bundle, shardIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle shard hash: %w", err)
	}

	// Get the segment shard sequence (s) for the given shard index
	segmentShardSequence, err := getSegmentShardSequence(bundle, shardIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get segment shard sequence: %w", err)
	}

	// Construct T(s, i, H) - the Merkle tree co-path for the segment
	merkleCoPath, err := constructMerkleCoPath(segmentShardSequence, segmentIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to construct Merkle co-path: %w", err)
	}

	// Combine the justifications: j ^ [b] ^ T(s, i, H)
	combinedJustification := combineJustifications(ce137Justification, bundleShardHash, merkleCoPath)

	return combinedJustification, nil
}

// getCE137Justification retrieves the CE137 justification for the given erasure root and shard index.
func getCE137Justification(erasureRoot []byte, shardIndex uint32) ([]byte, error) {
	// TODO: Implement actual CE137 justification lookup
	// 1. Look up the CE137 justification from the assurer's local storage
	// 2. The justification would have been received when the assurer requested their shard via CE137
	// 3. Return the stored justification or an error if not found

	// For now, create a mock justification that simulates a real CE137 response
	// TODO:  the justification received from the guarantor via CE137
	mockJustification := make([]byte, 33) // 1 byte discriminator + 32 bytes hash
	mockJustification[0] = 0x00           // Type 0: single hash

	// Create a deterministic hash based on erasure root and shard index
	// TODO: actual hash from the CE137 response
	hashInput := append(erasureRoot, byte(shardIndex), byte(shardIndex>>8), byte(shardIndex>>16), byte(shardIndex>>24))
	mockHash := hash.Blake2bHash(types.ByteSequence(hashInput))
	copy(mockJustification[1:], mockHash[:])

	return mockJustification, nil
}

// getBundleShardHash computes the hash of the bundle shard at the given index.
func getBundleShardHash(bundle *types.WorkPackageBundle, shardIndex uint32) ([]byte, error) {
	// Encode the bundle to get the raw bytes
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
	requestedShard := shards[shardIndex]

	shardHash := hash.Blake2bHash(types.ByteSequence(requestedShard))
	return shardHash[:], nil
}

// getSegmentShardSequence gets the full sequence of segment shards for the given shard index.
func getSegmentShardSequence(bundle *types.WorkPackageBundle, shardIndex uint32) ([][]byte, error) {
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
	requestedShard := shards[shardIndex]

	segmentSize := 32
	var segments [][]byte
	for i := 0; i < len(requestedShard); i += segmentSize {
		end := i + segmentSize
		if end > len(requestedShard) {
			end = len(requestedShard)
		}
		segments = append(segments, requestedShard[i:end])
	}

	return segments, nil
}

// constructMerkleCoPath constructs the Merkle tree co-path for the given segment index.
func constructMerkleCoPath(segmentShardSequence [][]byte, segmentIndex uint16) ([]byte, error) {
	if int(segmentIndex) >= len(segmentShardSequence) {
		return nil, fmt.Errorf("segment index %d out of range (max: %d)", segmentIndex, len(segmentShardSequence)-1)
	}

	var byteSequences []types.ByteSequence
	for _, segment := range segmentShardSequence {
		byteSequences = append(byteSequences, types.ByteSequence(segment))
	}

	// Use the T function from merkle_tree package to get the co-path
	// T(s, i, H) where s is the segment sequence, i is the segment index, H is Blake2b
	coPath := merkle_tree.T(byteSequences, types.U32(segmentIndex), hash.Blake2bHash)

	// The co-path should be encoded as a sequence of [0 ++ Hash OR 1 ++ Hash ++ Hash OR 2 ++ Segment Shard]
	var result []byte

	for _, item := range coPath {
		if len(item.ByteSequence) > 0 {
			// This is a segment shard (type 2)
			result = append(result, 0x02)
			result = append(result, item.ByteSequence...)
		} else {
			// This is a hash (type 0)
			result = append(result, 0x00) // discriminator for hash
			result = append(result, item.Hash[:]...)
		}
	}

	return result, nil
}

// combineJustifications combines multiple justifications using XOR operation.
// The formula is: j ^ [b] ^ T(s, i, H)
func combineJustifications(ce137Justification, bundleShardHash, merkleCoPath []byte) []byte {
	// Find the maximum length among all justifications
	maxLen := len(ce137Justification)
	if len(bundleShardHash) > maxLen {
		maxLen = len(bundleShardHash)
	}
	if len(merkleCoPath) > maxLen {
		maxLen = len(merkleCoPath)
	}

	paddedCE137 := make([]byte, maxLen)
	copy(paddedCE137, ce137Justification)

	paddedBundleHash := make([]byte, maxLen)
	copy(paddedBundleHash, bundleShardHash)

	paddedMerkleCoPath := make([]byte, maxLen)
	copy(paddedMerkleCoPath, merkleCoPath)

	combined := make([]byte, maxLen)
	for i := 0; i < maxLen; i++ {
		combined[i] = paddedCE137[i] ^ paddedBundleHash[i] ^ paddedMerkleCoPath[i]
	}

	return combined
}

type CE140Payload struct {
	ErasureRoot    []byte
	ShardIndex     uint32
	SegmentIndices []uint16
}
