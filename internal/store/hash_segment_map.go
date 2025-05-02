package store

import (
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type HashSegmentMap struct {
	client *RedisClient
}

func NewHashSegmentMap(client *RedisClient) *HashSegmentMap {
	return &HashSegmentMap{client: client}
}

/*
example map:
{
	"1697123456_wpHashA": "segmentRootA",
	"1697123460_wpHashB": "segmentRootB",
}
*/

// dict length <= 8
func (h *HashSegmentMap) SaveWithLimit(wpHash, segmentRoot types.OpaqueHash) error {
	key := "segment_dict"
	existingBytes, err := h.client.Get(key)
	dict := make(map[string]string)

	if err == nil && existingBytes != nil {
		json.Unmarshal(existingBytes, &dict)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	dict[timestamp+"_"+hex.EncodeToString(wpHash[:])] = hex.EncodeToString(segmentRoot[:])

	if len(dict) > 8 {
		var keys []string
		for k := range dict {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		delete(dict, keys[0])
	}

	encoded, err := json.Marshal(dict)
	if err != nil {
		return err
	}
	return h.client.Put(key, encoded)
}

func (h *HashSegmentMap) LoadDict() (map[types.OpaqueHash]types.OpaqueHash, error) {
	key := "segment_dict"
	result := make(map[types.OpaqueHash]types.OpaqueHash)

	data, err := h.client.Get(key)
	if err != nil || data == nil {
		return result, err
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	for k, v := range raw {
		parts := strings.SplitN(k, "_", 2)
		if len(parts) != 2 {
			continue
		}
		var wpHash, segmentRoot types.OpaqueHash
		wpBytes, _ := hex.DecodeString(parts[1])
		rootBytes, _ := hex.DecodeString(v)
		copy(wpHash[:], wpBytes)
		copy(segmentRoot[:], rootBytes)
		result[wpHash] = segmentRoot
	}
	return result, nil
}
