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
func HandleECShardRequest(stream io.ReadWriter, lookup func(erasureRoot []byte) (*CE137Payload, bool)) error {
	// Read erasure-root (32 bytes) + shard index (4 bytes) + 'FIN' (3 bytes)
	buf := make([]byte, 32+4+3)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return err
	}
	fin := buf[36:]
	if string(fin) != "FIN" {
		return errors.New("request does not end with FIN")
	}

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
	justification := append([]byte{0x00}, make([]byte, 32)...) // 0 discriminator + 32 zero bytes
	if _, err := stream.Write(justification); err != nil {
		return err
	}
	if _, err := stream.Write([]byte("FIN")); err != nil {
		return err
	}
	return nil
}

type CE137Payload struct {
	BundleShard   []byte
	SegmentShards [][]byte
	Justification []byte
}
