package ce

import (
	"errors"
	"io"
)

// HandleECShardRequest handles an assurer's request for their erasure coded shards from a guarantor.
//
// Request (from Assurer to Guarantor):
//
//	Erasure-Root (hash, []byte)
//	Shard Index (uint32 or int)
//	'FIN' (3 bytes)
//
// Response (from Guarantor to Assurer):
//
//	Bundle Shard ([]byte)
//	[Segment Shard] ([]byte, exported and proof segment shards with the given index)
//	Justification ([]byte, Merkle co-path proof)
//	'FIN' (3 bytes)
//
// The justification is a sequence of [0 ++ Hash OR 1 ++ Hash ++ Hash] as per the protocol.
//
// This function should use AssignShardIndex to determine the correct shard index.
func HandleECShardRequest(stream io.ReadWriter, lookup func(erasureRoot []byte) (*WorkReportBundle, bool)) error {
	// Read erasure-root (32 bytes) + shard index (4 bytes) + 'FIN' (3 bytes)
	buf := make([]byte, 32+4+3)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return err
	}
	// erasureRoot := buf[:32]
	// shardIndex := binary.LittleEndian.Uint32(buf[32:36])
	fin := buf[36:]
	if string(fin) != "FIN" {
		return errors.New("request does not end with FIN")
	}

	// bundle, ok := lookup(erasureRoot)
	// if !ok {
	// 	return errors.New("work report bundle not found")
	// }

	// --- MOCK RESPONSE ---
	// Write mock bundle shard
	bundleShard := []byte("BUNDLE_SHARD_MOCK")
	if _, err := stream.Write(bundleShard); err != nil {
		return err
	}
	// Write mock segment shards (simulate 2 segments)
	segmentShard1 := []byte("SEGMENT_SHARD1_MOCK")
	segmentShard2 := []byte("SEGMENT_SHARD2_MOCK")
	if _, err := stream.Write(segmentShard1); err != nil {
		return err
	}
	if _, err := stream.Write(segmentShard2); err != nil {
		return err
	}
	// Write mock justification (e.g., 0x00 + 32 bytes hash)
	justification := append([]byte{0x00}, make([]byte, 32)...) // 0 discriminator + 32 zero bytes
	if _, err := stream.Write(justification); err != nil {
		return err
	}
	// Write FIN
	if _, err := stream.Write([]byte("FIN")); err != nil {
		return err
	}
	return nil
}

// WorkReportBundle represents the data structure containing the erasure-coded bundle shard, segment shards, and justification for a work-report.
type WorkReportBundle struct {
	// BundleShard is the erasure-coded shard of the work-package bundle assigned to the validator.
	BundleShard []byte
	// SegmentShards are the exported and proof segment shards assigned to the validator (may be multiple per index).
	SegmentShards [][]byte
	// Justification is the Merkle co-path proof for the assigned shards.
	Justification []byte
}
