package store

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type RedisBackend struct {
	client  *RedisClient
	encoder *types.Encoder
	decoder *types.Decoder
}

// NewRedisBackend initializes and returns a new RedisBackend.
func NewRedisBackend(client *RedisClient) *RedisBackend {
	return &RedisBackend{client: client, encoder: types.NewEncoder(), decoder: types.NewDecoder()}
}

func hexOf(h types.OpaqueHash) string {
	return hex.EncodeToString(h[:])
}

// AddHead => SADD "heads" <OpaqueHash-bytes>
func (r *RedisBackend) AddHead(ctx context.Context, head types.OpaqueHash) error {
	key := "heads"
	return r.client.SAdd(key, head[:]) // head is [32]byte, convert to []byte
}

// RemoveHead => SREM "heads" <OpaqueHash-bytes>
func (r *RedisBackend) RemoveHead(ctx context.Context, head types.OpaqueHash) error {
	key := "heads"
	return r.client.SRem(key, head[:])
}

// IsHead => SISMEMBER "heads" <OpaqueHash-bytes>
func (r *RedisBackend) IsHead(ctx context.Context, head types.OpaqueHash) (bool, error) {
	key := "heads"
	return r.client.SIsMember(key, head[:])
}

// GetHeads => SMEMBERS "heads"
func (r *RedisBackend) GetHeads(ctx context.Context) ([]types.OpaqueHash, error) {
	key := "heads"
	members, err := r.client.SMembers(key)
	if err != nil {
		return nil, err
	}

	heads := make([]types.OpaqueHash, 0, len(members))
	for _, m := range members {
		if len(m) != 32 {
			continue
		}
		var h types.OpaqueHash
		copy(h[:], m)
		heads = append(heads, h)
	}
	return heads, nil
}

// Blocks By Hash
func (r *RedisBackend) StoreBlockByHash(ctx context.Context, block *types.Block, blockHash types.OpaqueHash) error {
	key := fmt.Sprintf("block:%s", hexOf(blockHash))
	data, err := r.encoder.Encode(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}
	return r.client.Put(key, data)
}

func (r *RedisBackend) GetBlockByHash(ctx context.Context, blockHash types.OpaqueHash) (*types.Block, error) {
	key := fmt.Sprintf("block:%s", hexOf(blockHash))

	data, err := r.client.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	block := &types.Block{}
	if err := r.decoder.Decode(data, block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}
	return block, nil
}

// Blocks By Slot
func (r *RedisBackend) StoreBlockBySlot(ctx context.Context, block *types.Block, ts types.TimeSlot) error {
	// Create key based on slot
	key := fmt.Sprintf("block:%d", ts)

	data, err := r.encoder.Encode(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	return r.client.Put(key, data)
}

func (r *RedisBackend) GetBlockBySlot(ctx context.Context, slot types.TimeSlot) (*types.Block, error) {
	key := fmt.Sprintf("block:%d", slot)

	data, err := r.client.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	block := &types.Block{}
	if err := r.decoder.Decode(data, block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return block, nil
}

func (r *RedisBackend) DeleteBlockBySlot(ctx context.Context, slot types.TimeSlot) error {
	key := fmt.Sprintf("block:%d", slot)
	return r.client.Delete(key)
}

func (r *RedisBackend) SetFinalizedHead(ctx context.Context, hash types.OpaqueHash) error {
	key := "meta:finalizedHead"
	// store the raw 32 bytes
	return r.client.Put(key, hash[:])
}

func (r *RedisBackend) SetGenesisBlock(ctx context.Context, block *types.Block) error {
	key := "meta:genesisBlock"
	data, err := r.encoder.Encode(block)
	if err != nil {
		return fmt.Errorf("failed to marshal genesis block: %w", err)
	}
	if err := r.client.Put(key, data); err != nil {
		return fmt.Errorf("failed to store genesis block: %w", err)
	}
	log.Printf("Genesis block stored with key %s", key)
	return nil
}

func (r *RedisBackend) GetGenesisBlock(ctx context.Context) (*types.Block, error) {
	key := "meta:genesisBlock"
	data, err := r.client.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	block := &types.Block{}
	if err := r.decoder.Decode(data, block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis block: %w", err)
	}
	return block, nil
}

func (r *RedisBackend) GetFinalizedHead(ctx context.Context) (*types.OpaqueHash, error) {
	key := "meta:finalizedHead"
	data, err := r.client.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	if len(data) != 32 {
		return nil, fmt.Errorf("invalid length for finalizedHead: got %d, want 32", len(data))
	}
	var out types.OpaqueHash
	copy(out[:], data)
	return &out, nil
}

// UpdateHead ensures `parent` is either an existing head or a known block, then
// removes `parent` from the heads set and adds `hash`.
func (r *RedisBackend) UpdateHead(ctx context.Context, newHead, parent types.OpaqueHash) error {
	// Try removing `parent` from heads
	removedErr := r.client.SRem("heads", parent[:])
	if removedErr != nil {
		return fmt.Errorf("failed removing parent from heads: %w", removedErr)
	}
	// If SREM didn't remove anything, maybe parent wasn't a head => check if parent's a known block
	isParentBlock, err := r.HasBlockHash(ctx, parent)
	if err != nil {
		return err
	}
	if !isParentBlock {
		// parent not in heads, not in blocks => error
		return fmt.Errorf("no data for parent %s", hexOf(parent))
	}
	// Now add the new head
	addErr := r.client.SAdd("heads", newHead[:])
	if addErr != nil {
		return fmt.Errorf("failed adding new head: %w", addErr)
	}
	return nil
}

// HasBlockHash is a quick check if "block:<hash>" exists:
func (r *RedisBackend) HasBlockHash(ctx context.Context, h types.OpaqueHash) (bool, error) {
	key := fmt.Sprintf("block:%s", hexOf(h))
	return r.client.Exists(key)
}

func (r *RedisBackend) RemoveBlock(ctx context.Context, hash types.OpaqueHash) error {
	// 1) remove the block
	key := fmt.Sprintf("block:%s", hexOf(hash))
	if err := r.client.Delete(key); err != nil {
		return fmt.Errorf("failed to delete block: %w", err)
	}

	// 2) if you want to remove it from timeslot or blockNumber sets, you'd do that here,
	//    but you'd need to read the block first to know which timeslot it had:
	block, err := r.GetBlockByHash(ctx, hash)
	if err != nil {
		return err
	}
	if block != nil {
		timeslotKey := fmt.Sprintf("blockHashByTimeslot:%d", block.Header.Slot)
		if err := r.client.SRem(timeslotKey, hash[:]); err != nil {
			return fmt.Errorf("failed to remove from blockHashByTimeslot: %w", err)
		}
		// etc. for block number sets
	}

	// 3) remove from heads if it was a head
	if err := r.client.SRem("heads", hash[:]); err != nil {
		return fmt.Errorf("failed to remove from heads: %w", err)
	}

	return nil
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

// State Storage Functions

func (r *RedisBackend) StoreStateRootByBlockHash(
	ctx context.Context,
	blockHash types.HeaderHash,
	stateRoot types.StateRoot,
) error {
	key := fmt.Sprintf("state_root:%s", hex.EncodeToString(blockHash[:]))
	return r.client.Put(key, stateRoot[:])
}

func (r *RedisBackend) GetStateRootByBlockHash(
	ctx context.Context,
	blockHash types.HeaderHash,
) (*types.StateRoot, error) {
	key := fmt.Sprintf("state_root:%s", hex.EncodeToString(blockHash[:]))
	data, err := r.client.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("state root not found for block %x", blockHash)
	}

	var stateRoot types.StateRoot
	copy(stateRoot[:], data)
	return &stateRoot, nil
}

func (r *RedisBackend) StoreStateData(
	ctx context.Context,
	stateRoot types.StateRoot,
	stateKeyVals types.StateKeyVals,
) error {
	key := fmt.Sprintf("state_data:%s", hex.EncodeToString(stateRoot[:]))

	data, err := r.encoder.Encode(&stateKeyVals)
	if err != nil {
		return fmt.Errorf("failed to encode state data: %w", err)
	}

	return r.client.Put(key, data)
}

func (r *RedisBackend) GetStateData(
	ctx context.Context,
	stateRoot types.StateRoot,
) (types.StateKeyVals, error) {
	key := fmt.Sprintf("state_data:%s", hex.EncodeToString(stateRoot[:]))
	data, err := r.client.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("state data not found for state root %x", stateRoot)
	}

	var stateKeyVals types.StateKeyVals
	err = r.decoder.Decode(data, &stateKeyVals)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state data: %w", err)
	}

	return stateKeyVals, nil
}

func (r *RedisBackend) GetStateByBlockHash(
	ctx context.Context,
	blockHash types.HeaderHash,
) (types.StateKeyVals, error) {
	stateRoot, err := r.GetStateRootByBlockHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}

	return r.GetStateData(ctx, *stateRoot)
}

// get raw bytes for debugging purposes
func (r *RedisBackend) DebugGetBytes(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(key)
}
