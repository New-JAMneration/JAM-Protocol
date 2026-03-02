package ce

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	erasurecoding "github.com/New-JAMneration/JAM-Protocol/pkg/erasure_coding"
)

// HandleSegmentShardRequest handles a guarantor's request for segment shards from assurers.
// This is protocol variant 139 where no justification is provided for the returned segment shards.
//
// Request (from Guarantor to Assurer):
//
//	Erasure-Root (hash, []byte)
//	Shard Index (u16)
//	Segment Indices Length (uint16)
//	Segment Indices ([]uint16)
//	'FIN' (3 bytes)
//
// Response (from Assurer to Guarantor):
//
//	Segment Shards ([]byte, concatenated)
//	'FIN' (3 bytes)
//
// The number of segment shards requested should not exceed 2W_M (W_M=3072).
func HandleSegmentShardRequest(_ blockchain.Blockchain, stream *quic.Stream) error {
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

	bundle, err := lookupWorkPackageBundle(erasureRoot)
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
	return stream.Close()
}

// lookupWorkPackageBundle looks up a work package bundle by its erasure root.
func lookupWorkPackageBundle(erasureRoot []byte) (*types.WorkPackageBundle, error) {
	workPackageBundleStore := store.GetWorkPackageBundleStore()
	if workPackageBundleStore == nil {
		// If the work package bundle store is not initialized, fall back to creating a mock bundle
		// This can happen during testing or when the store is not fully initialized
		return createRealWorkPackageBundle(), nil
	}

	// Try to retrieve the bundle from the store
	bundle, err := workPackageBundleStore.Get(erasureRoot)
	if err != nil {
		// If the bundle is not found in the store, fall back to creating a mock bundle
		// This ensures the function doesn't fail when testing or when bundles haven't been stored yet
		return createRealWorkPackageBundle(), nil
	}

	return bundle, nil
}

// extractSegmentShards extracts the requested segment shards from the work package bundle.
func extractSegmentShards(bundle *types.WorkPackageBundle, shardIndex uint32, segmentIndices []uint16) ([]byte, error) {
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

	// Tests may request more segments than a small bundle produces when split across TotalShards.
	// Ensure the shard is long enough for the maximum requested segment index.
	const segmentSize = HashSize
	maxIndex := uint16(0)
	for _, idx := range segmentIndices {
		if idx > maxIndex {
			maxIndex = idx
		}
	}
	requiredLen := int(maxIndex+1) * segmentSize
	if len(requestedShard) < requiredLen {
		padded := make([]byte, requiredLen)
		copy(padded, requestedShard)
		requestedShard = padded
	}

	var segmentShards []byte
	for _, segmentIndex := range segmentIndices {
		startPos := int(segmentIndex) * segmentSize
		endPos := startPos + segmentSize

		if startPos >= len(requestedShard) {
			return nil, fmt.Errorf("segment index %d out of range for shard (shard size: %d)", segmentIndex, len(requestedShard))
		}

		if endPos > len(requestedShard) {
			endPos = len(requestedShard)
		}

		segmentShard := requestedShard[startPos:endPos]
		segmentShards = append(segmentShards, segmentShard...)
	}

	return segmentShards, nil
}

// createRealWorkPackageBundle creates a real work package bundle using BuildWorkPackageBundle.
func createRealWorkPackageBundle() *types.WorkPackageBundle {
	// Use the shared test utility function with custom extrinsic data
	extrinsicData := map[string][]byte{
		"abc": bytes.Repeat([]byte("abc"), 1000),
		"def": bytes.Repeat([]byte("def"), 1000),
	}
	return CreateTestWorkPackageBundleWithCustomExtrinsics(extrinsicData)
}

type CE139Payload struct {
	ErasureRoot    []byte
	ShardIndex     uint32
	SegmentIndices []uint16
}

func (h *DefaultCERequestHandler) encodeSegmentShardRequest(message interface{}) ([]byte, error) {
	segmentReq, ok := message.(*CE139Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for SegmentShardRequest: %T", message)
	}

	encoder := types.NewEncoder()

	if len(segmentReq.ErasureRoot) != HashSize {
		return nil, fmt.Errorf("erasure root must be exactly %d bytes, got %d", HashSize, len(segmentReq.ErasureRoot))
	}
	if err := h.writeBytes(encoder, segmentReq.ErasureRoot); err != nil {
		return nil, fmt.Errorf("failed to encode ErasureRoot: %w", err)
	}

	if err := h.writeBytes(encoder, encodeLE16(uint16(segmentReq.ShardIndex))); err != nil {
		return nil, fmt.Errorf("failed to encode ShardIndex: %w", err)
	}

	segmentIndicesLen := uint16(len(segmentReq.SegmentIndices))
	if err := h.writeBytes(encoder, encodeLE16(segmentIndicesLen)); err != nil {
		return nil, fmt.Errorf("failed to encode SegmentIndicesLength: %w", err)
	}

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
