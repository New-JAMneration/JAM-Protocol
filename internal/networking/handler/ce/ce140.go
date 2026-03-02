package ce

import (
	"encoding/binary"
	"errors"
	"fmt"

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
func HandleSegmentShardRequestWithJustification(bc blockchain.Blockchain, stream *quic.Stream) error {
	// Request: single message = erasure-root (HashSize) + shard index (U16Size) + segment indices length (U16Size) + segment indices
	payload, err := stream.ReadMessage()
	if err != nil {
		return err
	}
	if len(payload) < CE139140MinRequestSize {
		return errors.New("segment shard request too short")
	}
	erasureRoot := payload[:HashSize]
	shardIndex := uint32(binary.LittleEndian.Uint16(payload[HashSize : HashSize+U16Size]))
	segmentIndicesLen := binary.LittleEndian.Uint16(payload[HashSize+U16Size : CE139140MinRequestSize])

	if segmentIndicesLen > MaxSegmentIndicesCount {
		return errors.New("segment indices length exceeds maximum allowed (2W_M)")
	}

	rest := payload[CE139140MinRequestSize:]
	if len(rest) < int(segmentIndicesLen)*SegmentIndexSize {
		return errors.New("segment shard request truncated")
	}
	segmentIndices := make([]uint16, segmentIndicesLen)
	for i := uint16(0); i < segmentIndicesLen; i++ {
		segmentIndices[i] = binary.LittleEndian.Uint16(rest[i*SegmentIndexSize : (i+1)*SegmentIndexSize])
	}

	bundle, err := lookupWorkPackageBundle(bc, erasureRoot)
	if err != nil {
		return fmt.Errorf("failed to lookup work package bundle: %w", err)
	}

	segmentShards, err := extractSegmentShards(bundle, shardIndex, segmentIndices)
	if err != nil {
		return fmt.Errorf("failed to extract segment shards: %w", err)
	}

	if err := stream.WriteMessage(segmentShards); err != nil {
		return err
	}

	for _, segmentIndex := range segmentIndices {
		justification, err := constructJustification(bundle, erasureRoot, shardIndex, segmentIndex)
		if err != nil {
			return fmt.Errorf("failed to construct justification for segment %d: %w", segmentIndex, err)
		}
		if err := stream.WriteMessage(justification); err != nil {
			return err
		}
	}
	return stream.Close()
}

// constructJustification constructs the justification for a segment shard using the formula:
// j ++ [b] ++ T(s, i, H)
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
	segmentShardSequence, err := getSegmentShardSequence(bundle, shardIndex, segmentIndex)
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
	mockJustification := make([]byte, JustificationHashEntrySize) // 1 byte discriminator + HashSize bytes hash
	mockJustification[0] = 0x00                                   // Type 0: single hash

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
	requestedShard := shards[shardIndex]

	shardHash := hash.Blake2bHash(types.ByteSequence(requestedShard))
	return shardHash[:], nil
}

// getSegmentShardSequence gets the full sequence of segment shards for the given shard index.
func getSegmentShardSequence(bundle *types.WorkPackageBundle, shardIndex uint32, minSegments uint16) ([][]byte, error) {
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
	requestedShard := shards[shardIndex]

	segmentSize := HashSize
	requiredLen := int(minSegments+1) * segmentSize
	if len(requestedShard) < requiredLen {
		padded := make([]byte, requiredLen)
		copy(padded, requestedShard)
		requestedShard = padded
	}
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

	// In this codebase, merkle_tree.T returns a sequence of sibling hashes as ByteSequence.
	// Encode as repeated (0x00 ++ HashSize-byte-hash).
	result := make([]byte, 0, len(coPath)*JustificationHashEntrySize)
	for _, h := range coPath {
		result = append(result, 0x00)
		result = append(result, h...)
	}
	return result, nil
}

// combineJustifications concatenates justifications per JAMNP: j ++ [b] ++ T(s, i, H).
// ++ is concatenation; [b] is the bundle shard hash entry (discriminator 0x00 + HashSize-byte hash).
func combineJustifications(ce137Justification, bundleShardHash, merkleCoPath []byte) []byte {
	result := make([]byte, 0, len(ce137Justification)+1+len(bundleShardHash)+len(merkleCoPath))
	result = append(result, ce137Justification...)
	result = append(result, 0x00) // discriminator for single hash
	result = append(result, bundleShardHash...)
	result = append(result, merkleCoPath...)
	return result
}

type CE140Payload struct {
	ErasureRoot    []byte
	ShardIndex     uint32
	SegmentIndices []uint16
}

func (h *DefaultCERequestHandler) encodeSegmentShardRequestWithJustification(message interface{}) ([]byte, error) {
	segmentReq, ok := message.(*CE140Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for SegmentShardRequestWithJustification: %T", message)
	}

	encoder := types.NewEncoder()

	// Encode ErasureRoot (HashSize bytes)
	if len(segmentReq.ErasureRoot) != HashSize {
		return nil, fmt.Errorf("erasure root must be exactly %d bytes, got %d", HashSize, len(segmentReq.ErasureRoot))
	}
	if err := h.writeBytes(encoder, segmentReq.ErasureRoot); err != nil {
		return nil, fmt.Errorf("failed to encode ErasureRoot: %w", err)
	}

	// Encode ShardIndex (2 bytes, u16)
	if err := h.writeBytes(encoder, encodeLE16(uint16(segmentReq.ShardIndex))); err != nil {
		return nil, fmt.Errorf("failed to encode ShardIndex: %w", err)
	}

	// Encode Segment Indices Length (2 bytes little-endian)
	segmentIndicesLen := uint16(len(segmentReq.SegmentIndices))
	if err := h.writeBytes(encoder, encodeLE16(segmentIndicesLen)); err != nil {
		return nil, fmt.Errorf("failed to encode SegmentIndicesLength: %w", err)
	}

	// Encode Segment Indices (2 bytes each, little-endian)
	for _, segmentIndex := range segmentReq.SegmentIndices {
		if err := h.writeBytes(encoder, encodeLE16(segmentIndex)); err != nil {
			return nil, fmt.Errorf("failed to encode SegmentIndex: %w", err)
		}
	}

	result := make([]byte, 0, CE139140MinRequestSize+len(segmentReq.SegmentIndices)*SegmentIndexSize)
	result = append(result, segmentReq.ErasureRoot...)
	result = append(result, encodeLE16(uint16(segmentReq.ShardIndex))...)
	result = append(result, encodeLE16(segmentIndicesLen)...)
	for _, segmentIndex := range segmentReq.SegmentIndices {
		result = append(result, encodeLE16(segmentIndex)...)
	}

	return result, nil
}
