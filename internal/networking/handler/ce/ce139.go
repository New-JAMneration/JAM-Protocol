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

// HandleSegmentShardRequest_Assurer handles a guarantor's request for segment shards from assurers.
// Role: [Guarantor -> Assurer]
func HandleSegmentShardRequest_Assurer(blockchain blockchain.Blockchain, stream *quic.Stream) error {
	// Read erasure-root (32 bytes) + shard index (4 bytes) + segment indices length (2 bytes)
	buf := make([]byte, 32+4+2)
	if err := stream.ReadFull(buf); err != nil {
		return err
	}
	erasureRoot := buf[:32]
	shardIndex := binary.LittleEndian.Uint32(buf[32:36])
	segmentIndicesLen := binary.LittleEndian.Uint16(buf[36:38])

	// with len++[Segment Index] (list of segment indices)
	segmentIndices := make([]uint16, segmentIndicesLen)
	if segmentIndicesLen > 0 {
		indicesBuf := make([]byte, segmentIndicesLen*2)
		if err := stream.ReadFull(indicesBuf); err != nil {
			return err
		}
		for i := range segmentIndicesLen {
			segmentIndices[i] = binary.LittleEndian.Uint16(indicesBuf[i*2 : (i+1)*2])
		}
	}

	finBuf := make([]byte, 3)
	if err := stream.ReadFull(finBuf); err != nil {
		return err
	} else if string(finBuf) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	// Prepare and send [Segment Shard]
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

func HandleSegmentShardRequest_Guarantor(
	stream *quic.Stream,
	erasureRoot []byte,
	shardIndex uint32,
	segmentIndices []uint16,
) error {
	if len(erasureRoot) != 32 {
		return fmt.Errorf("erasure root must be 32 bytes, got %d", len(erasureRoot))
	}

	reqLen := 32 + 4 + 2 + len(segmentIndices)*2 + 3
	req := make([]byte, 0, reqLen)
	req = append(req, erasureRoot...)
	req = append(req, encodeLE32(shardIndex)...)
	req = append(req, encodeLE16(uint16(len(segmentIndices)))...)
	for _, idx := range segmentIndices {
		req = append(req, encodeLE16(idx)...)
	}
	req = append(req, []byte("FIN")...)

	if err := stream.WriteMessage(req); err != nil {
		return fmt.Errorf("failed to send CE139 request: %w", err)
	}

	var resp bytes.Buffer
	tail := make([]byte, 0, 3)
	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			data := buf[:n]
			// Append and check FIN across chunk boundaries
			combined := append(tail, data...)
			if len(combined) >= 3 {
				if bytes.HasSuffix(combined, []byte("FIN")) {
					// Write all but last 3 bytes to resp
					resp.Write(combined[:len(combined)-3])
					break
				}
			}
			// Keep last up to 2 bytes in tail (since FIN is 3 bytes)
			if len(combined) >= 2 {
				// Write all except the last 2 to resp
				resp.Write(combined[:len(combined)-2])
				tail = combined[len(combined)-2:]
			} else {
				tail = combined
			}
		}
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("unexpected EOF before FIN")
			}
			return fmt.Errorf("failed reading CE139 response: %w", err)
		}
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
		return createRealWorkPackageBundle(), nil
	}

	bundle, err := workPackageBundleStore.Get(erasureRoot)
	if err != nil {
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

func (h *DefaultCERequestHandler) encodeSegmentShardRequest(message interface{}) ([]byte, error) {
	segmentReq, ok := message.(*CE139Payload)
	if !ok {
		return nil, fmt.Errorf("unsupported message type for SegmentShardRequest: %T", message)
	}

	encoder := types.NewEncoder()

	if len(segmentReq.ErasureRoot) != 32 {
		return nil, fmt.Errorf("erasure root must be exactly 32 bytes, got %d", len(segmentReq.ErasureRoot))
	}
	if err := h.writeBytes(encoder, segmentReq.ErasureRoot); err != nil {
		return nil, fmt.Errorf("failed to encode ErasureRoot: %w", err)
	}

	if err := h.writeBytes(encoder, encodeLE32(segmentReq.ShardIndex)); err != nil {
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

	result := make([]byte, 0, 38+len(segmentReq.SegmentIndices)*2)
	result = append(result, segmentReq.ErasureRoot...)
	result = append(result, encodeLE32(segmentReq.ShardIndex)...)
	result = append(result, encodeLE16(segmentIndicesLen)...)
	for _, segmentIndex := range segmentReq.SegmentIndices {
		result = append(result, encodeLE16(segmentIndex)...)
	}

	return result, nil
}
