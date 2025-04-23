package store

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type SegmentRootDict struct {
	client *RedisClient
}

// NewSegmentRootDict creates a new Redis-backed segment root dictionary.
func NewSegmentRootDict(client *RedisClient) *SegmentRootDict {
	return &SegmentRootDict{client: client}
}

// Put saves a mapping from segment root → work package hash
func (s *SegmentRootDict) Put(segmentRoot types.OpaqueHash, wpHash types.OpaqueHash) error {
	key := "segment_root_dict:" + hex.EncodeToString(segmentRoot[:])
	err := s.client.Put(key, wpHash[:])
	if err != nil {
		return fmt.Errorf("SegmentRootDict.Put failed: %w", err)
	}
	return nil
}

// Get retrieves the work package hash for a segment root
func (s *SegmentRootDict) Get(segmentRoot types.OpaqueHash) (*types.OpaqueHash, error) {
	key := "segment_root_dict:" + hex.EncodeToString(segmentRoot[:])
	data, err := s.client.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var wpHash types.OpaqueHash
	copy(wpHash[:], data)
	return &wpHash, nil
}
