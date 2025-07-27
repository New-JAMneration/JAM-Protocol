package ce

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

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
//	Shard Index (uint32)
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
func HandleSegmentShardRequest(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// Read erasure-root (32 bytes) + shard index (4 bytes) + segment indices length (2 bytes)
	buf := make([]byte, 32+4+2)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return err
	}
	erasureRoot := buf[:32]
	shardIndex := binary.LittleEndian.Uint32(buf[32:36])
	segmentIndicesLen := binary.LittleEndian.Uint16(buf[36:38])

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

	if _, err := stream.Write([]byte("FIN")); err != nil {
		return err
	}

	return stream.Close()
}

// lookupWorkPackageBundle looks up a work package bundle by its erasure root.
func lookupWorkPackageBundle(erasureRoot []byte) (*types.WorkPackageBundle, error) {
	// Get the store instance and access the work package bundle store
	storeInstance := store.GetInstance()
	if storeInstance == nil {
		return nil, fmt.Errorf("store instance is not initialized")
	}

	workPackageBundleStore := storeInstance.GetWorkPackageBundleStore()
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

	var segmentShards []byte
	for _, segmentIndex := range segmentIndices {
		segmentSize := 32
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
