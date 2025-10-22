package store

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/alicebob/miniredis"
)

type RedisBackend struct {
	client  *RedisClient
	encoder *types.Encoder
	decoder *types.Decoder
}

var (
	redisInitOnce      sync.Once
	globalRedisBackend *RedisBackend
	globalMiniRedis    *miniredis.Miniredis
)

// NewRedisBackend initializes and returns a new RedisBackend.
func NewRedisBackend(client *RedisClient) *RedisBackend {
	return &RedisBackend{client: client, encoder: types.NewEncoder(), decoder: types.NewDecoder()}
}

func GetRedisBackend() (*RedisBackend, error) {
	redisInitOnce.Do(func() {
		redisConfig := config.Config.Redis
		client := NewRedisClient(redisConfig.Address, redisConfig.Password, redisConfig.Port)
		if err := client.Ping(); err != nil {
			// If Redis connection fails, check if we should use miniredis for testing
			useMini := os.Getenv("USE_MINI_REDIS")
			if useMini == "true" {
				log.Printf("USE_MINI_REDIS is set to true, starting miniredis for testing")
				mr, miniErr := miniredis.Run()
				if miniErr != nil {
					log.Printf("failed to start miniredis: %v", miniErr)
					return
				}
				globalMiniRedis = mr
				client = NewRedisClient(mr.Addr(), "", 0)
			} else {
				log.Printf("failed to connect to Redis: %v", err)
				return
			}
		}
		globalRedisBackend = NewRedisBackend(client)

		hashSegmentMap := genInitHashSegmentMap()
		err := globalRedisBackend.SetHashSegmentMap(context.Background(), hashSegmentMap)
		if err != nil {
			log.Printf("failed to set hash segment map in Redis: %v", err)
			return
		}
	})
	if globalRedisBackend == nil {
		return nil, errors.New("redis backend is not initialized")
	}
	return globalRedisBackend, nil
}

func genInitHashSegmentMap() map[string]string {
	return make(map[string]string)
}

/*
example HashSegmentMap:
{
	"1697123456_wpHashA": "segmentRootA",
	"1697123460_wpHashB": "segmentRootB",
}
*/

// dict length <= 8
func (r *RedisBackend) SetHashSegmentMapWithLimit(wpHash, segmentRoot types.OpaqueHash) (map[types.OpaqueHash]types.OpaqueHash, error) {
	key := "segment_dict"
	existingBytes, err := r.client.Get(key)
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
		return nil, err
	}
	if err := r.client.Put(key, encoded); err != nil {
		return nil, err
	}

	// Convert the map back to the original format
	final := make(map[types.OpaqueHash]types.OpaqueHash)
	for k, v := range dict {
		parts := strings.SplitN(k, "_", 2)
		if len(parts) != 2 {
			continue
		}
		var wph, sr types.OpaqueHash
		wpBytes, _ := hex.DecodeString(parts[1])
		segBytes, _ := hex.DecodeString(v)
		copy(wph[:], wpBytes)
		copy(sr[:], segBytes)
		final[wph] = sr
	}
	return final, nil
}

func (r *RedisBackend) SetHashSegmentMap(ctx context.Context, hashSegmentMap map[string]string) error {
	fmt.Println("Set Hash Segment Map")
	key := "segment_dict"
	encoded, err := json.Marshal(hashSegmentMap)
	if err != nil {
		return err
	}
	if err := r.client.Put(key, encoded); err != nil {
		return err
	}

	return nil
}

func (r *RedisBackend) GetHashSegmentMap() (map[types.OpaqueHash]types.OpaqueHash, error) {
	fmt.Println("Get Hash Segment Map")
	key := "segment_dict"
	result := make(map[types.OpaqueHash]types.OpaqueHash)

	data, err := r.client.Get(key)
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

/*
example SegmentErasureMap:

	{
		"segment_erasure:segmentRootA": "erasureRootA",
		"segment_erasure:segmentRootB": "erasureRootB",
	}
*/
func (r *RedisBackend) SetSegmentErasureMap(segmentRoot, erasureRoot types.OpaqueHash) error {
	key := "segment_erasure:" + hex.EncodeToString(segmentRoot[:])
	ttl := types.SegmentErasureTTL
	return r.client.PutWithTTL(key, erasureRoot[:], ttl)
}

func (r *RedisBackend) GetSegmentErasureMap(segmentRoot types.OpaqueHash) (types.OpaqueHash, error) {
	key := "segment_erasure:" + hex.EncodeToString(segmentRoot[:])
	val, err := r.client.Get(key)
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
