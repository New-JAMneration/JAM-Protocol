package store

import (
	"encoding/hex"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type SegmentErasureMap struct {
	client *RedisClient
}

func NewSegmentErasureMap(client *RedisClient) *SegmentErasureMap {
	return &SegmentErasureMap{client: client}
}

func (s *SegmentErasureMap) Save(segmentRoot, erasureRoot types.OpaqueHash) error {
	key := "segment_erasure:" + hex.EncodeToString(segmentRoot[:])
	ttl := types.SegmentErasureTTL
	return s.client.PutWithTTL(key, erasureRoot[:], ttl)
}

func (s *SegmentErasureMap) Get(segmentRoot types.OpaqueHash) (types.OpaqueHash, error) {
	key := "segment_erasure:" + hex.EncodeToString(segmentRoot[:])
	val, err := s.client.Get(key)
	if err != nil {
		return types.OpaqueHash{}, err
	}
	if val == nil {
		// No value found for the given key
		return types.OpaqueHash{}, nil
	}

	var erasureRoot types.OpaqueHash
	copy(erasureRoot[:], val)
	return erasureRoot, nil
}
