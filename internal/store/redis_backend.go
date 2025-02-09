package store

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type RedisBackend struct {
	client *RedisClient
}

// NewRedisBackend initializes and returns a new RedisBackend.
func NewRedisBackend(client *RedisClient) *RedisBackend {
	return &RedisBackend{client: client}
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
	data, err := json.Marshal(block)
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
		return nil, nil // not found
	}
	var block types.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}
	return &block, nil
}

// Blocks By Slot
func (r *RedisBackend) StoreBlockBySlot(ctx context.Context, block types.Block) error {
	return r.client.StoreBlock(block)
}

func (r *RedisBackend) FetchBlockBySlot(ctx context.Context, slot types.TimeSlot) (*types.Block, error) {
	return r.client.GetBlock(ctx, slot)
}

func (r *RedisBackend) DeleteBlockBySlot(ctx context.Context, slot types.TimeSlot) error {
	return r.client.DeleteBlock(ctx, slot)
}

func (r *RedisBackend) SetFinalizedHead(ctx context.Context, hash types.OpaqueHash) error {
	key := "meta:finalizedHead"
	// store the raw 32 bytes
	return r.client.Put(key, hash[:])
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
