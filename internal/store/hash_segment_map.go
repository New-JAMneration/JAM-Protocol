package store

import (
	"encoding/hex"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type HashSegmentMap struct {
	client *RedisClient
}

func NewHashSegmentMap(client *RedisClient) *HashSegmentMap {
	return &HashSegmentMap{client: client}
}

// 14.12 L(r)
func (d *HashSegmentMap) Lookup(possiblyMarkedRoot types.OpaqueHash) types.OpaqueHash {
	key := "segment_lookup:" + hex.EncodeToString(possiblyMarkedRoot[:])
	val, err := d.client.Get(key)
	if err != nil || val == nil {
		return possiblyMarkedRoot
	}
	var segmentRoot types.OpaqueHash
	copy(segmentRoot[:], val)
	return segmentRoot
}

func (d *HashSegmentMap) Set(workPackageHash, segmentRoot types.OpaqueHash) error {
	key := "segment_lookup:" + hex.EncodeToString(workPackageHash[:])
	return d.client.Put(key, segmentRoot[:])
}

func (d *HashSegmentMap) SetWithLimit(workPackageHash, segmentRoot types.OpaqueHash) error {
	key := "segment_lookup:" + hex.EncodeToString(workPackageHash[:])
	if err := d.client.Put(key, segmentRoot[:]); err != nil {
		return err
	}

	indexKey := "segment_lookup:index"
	hashHex := hex.EncodeToString(workPackageHash[:])
	if err := d.client.SAdd(indexKey, []byte(hashHex)); err != nil {
		return err
	}

	members, err := d.client.SMembers(indexKey)
	if err != nil {
		return err
	}
	if len(members) > 8 {
		oldest := members[0]
		if err := d.client.SRem(indexKey, oldest); err != nil {
			return err
		}
		deleteKey := "segment_lookup:" + string(oldest)
		if err := d.client.Delete(deleteKey); err != nil {
			return err
		}
	}
	return nil
}
