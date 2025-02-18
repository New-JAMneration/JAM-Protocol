package store

import (
	"context"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TestRedisBackend_Heads checks AddHead, RemoveHead, IsHead, and GetHeads.
func TestRedisBackend_Heads(t *testing.T) {
	ctx := context.Background()

	// 1) Start miniredis + create RedisClient
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	// 2) Create the RedisBackend
	backend := NewRedisBackend(rdb)

	// 3) Create a sample OpaqueHash
	var testHash types.OpaqueHash
	copy(testHash[:], []byte("test-hash-32-bytes-long-abcdefg-123456")) // must be 32 bytes

	// Ensure we start with an empty set
	isHead, err := backend.IsHead(ctx, testHash)
	if err != nil {
		t.Fatalf("IsHead failed: %v", err)
	}
	if isHead {
		t.Fatalf("expected not a head yet, but got true")
	}

	// Add head
	if err := backend.AddHead(ctx, testHash); err != nil {
		t.Fatalf("AddHead failed: %v", err)
	}

	// Now it should be a head
	isHead, err = backend.IsHead(ctx, testHash)
	if err != nil {
		t.Fatalf("IsHead failed after AddHead: %v", err)
	}
	if !isHead {
		t.Fatalf("expected to be a head, but got false")
	}

	// Check GetHeads
	heads, err := backend.GetHeads(ctx)
	if err != nil {
		t.Fatalf("GetHeads failed: %v", err)
	}
	if len(heads) != 1 {
		t.Fatalf("expected 1 head, got %d", len(heads))
	}

	// Remove head
	if err := backend.RemoveHead(ctx, testHash); err != nil {
		t.Fatalf("RemoveHead failed: %v", err)
	}

	// Confirm it's no longer a head
	isHead, err = backend.IsHead(ctx, testHash)
	if err != nil {
		t.Fatalf("IsHead failed after RemoveHead: %v", err)
	}
	if isHead {
		t.Fatalf("expected not a head after removal, but got true")
	}
}

// TestRedisBackend_StoreBlockByHash checks storing & retrieving a block by hash.
func TestRedisBackend_StoreBlockByHash(t *testing.T) {
	ctx := context.Background()

	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	backend := NewRedisBackend(rdb)

	// Create a sample block
	block := &types.Block{
		Header: types.Header{
			Slot:   123,
			Parent: types.HeaderHash{},
		},
	}

	// We'll make up a block hash
	var blockHash types.OpaqueHash
	copy(blockHash[:], []byte("block-hash-32-bytes"))

	// Store the block by hash
	if err := backend.StoreBlockByHash(ctx, block, blockHash); err != nil {
		t.Fatalf("StoreBlockByHash failed: %v", err)
	}

	// Retrieve it
	gotBlock, err := backend.GetBlockByHash(ctx, blockHash)
	if err != nil {
		t.Fatalf("GetBlockByHash failed: %v", err)
	}
	if gotBlock == nil {
		t.Fatalf("expected to retrieve a block, got nil")
	}

	// Check some fields
	if gotBlock.Header.Slot != block.Header.Slot {
		t.Errorf("Slot mismatch: want %d, got %d", block.Header.Slot, gotBlock.Header.Slot)
	}
}

func TestRedisBackend_UpdateHead(t *testing.T) {
	ctx := context.Background()

	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	backend := NewRedisBackend(rdb)

	// We'll define a "parent" hash & a "newHead"
	var parentHash types.OpaqueHash
	copy(parentHash[:], []byte("parent-32-bytes"))
	var newHeadHash types.OpaqueHash
	copy(newHeadHash[:], []byte("new-head-32-bytes"))

	// Add the head
	if err := backend.AddHead(ctx, parentHash); err != nil {
		t.Fatalf("AddHead parent failed: %v", err)
	}
	backend.StoreBlockByHash(ctx, &types.Block{}, parentHash)

	// Update head
	if err := backend.UpdateHead(ctx, newHeadHash, parentHash); err != nil {
		t.Fatalf("UpdateHead failed: %v", err)
	}

	// Now the parent should no longer be a head
	isHead, err := backend.IsHead(ctx, parentHash)
	if err != nil {
		t.Fatalf("IsHead check failed: %v", err)
	}
	if isHead {
		t.Fatalf("parent should not be a head after update")
	}

	// newHead should be a head
	isHead, err = backend.IsHead(ctx, newHeadHash)
	if err != nil {
		t.Fatalf("IsHead(newHead) failed: %v", err)
	}
	if !isHead {
		t.Fatalf("expected newHead to be a head")
	}
}

func TestRedisBackend_RemoveBlock(t *testing.T) {
	ctx := context.Background()

	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	backend := NewRedisBackend(rdb)

	var blockHash types.OpaqueHash
	copy(blockHash[:], []byte("block-32-bytes-abcdefg-1234567890xxxx"))

	// Create a block
	block := &types.Block{
		Header: types.Header{
			Slot:   123,
			Parent: types.HeaderHash{},
		},
	}

	// Store by hash
	if err := backend.StoreBlockByHash(ctx, block, blockHash); err != nil {
		t.Fatalf("StoreBlockByHash failed: %v", err)
	}

	// Also test "HasBlockHash" -> should be true
	hasBlock, err := backend.HasBlockHash(ctx, blockHash)
	if err != nil {
		t.Fatalf("HasBlockHash failed: %v", err)
	}
	if !hasBlock {
		t.Fatalf("expected block to exist, but doesn't")
	}

	// Remove
	if err := backend.RemoveBlock(ctx, blockHash); err != nil {
		t.Fatalf("RemoveBlock failed: %v", err)
	}

	// HasBlockHash -> should be false
	hasBlock, err = backend.HasBlockHash(ctx, blockHash)
	if err != nil {
		t.Fatalf("HasBlockHash after remove failed: %v", err)
	}
	if hasBlock {
		t.Fatalf("block still exists after removal, expected false")
	}
}
