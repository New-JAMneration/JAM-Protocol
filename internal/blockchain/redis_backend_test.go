package blockchain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/alicebob/miniredis/v2"
	"github.com/test-go/testify/require"
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

// TestRedisBackend_StoreBlockBySlot checks storing & retrieving a block by slot.
func TestRedisBackend_StoreBlockBySlot(t *testing.T) {
	ctx := context.Background()

	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	backend := NewRedisBackend(rdb)

	// We'll use a TimeSlot as the key
	ts := types.TimeSlot(123)
	extrinsicHash := types.OpaqueHash{}
	copy(extrinsicHash[:], []byte("extrinsic-hash-32-bytes"))

	// Create a sample block
	block := &types.Block{
		Header: types.Header{
			Slot:          123,
			Parent:        types.HeaderHash{},
			ExtrinsicHash: extrinsicHash,
		},
	}

	// Store the block by hash
	if err := backend.StoreBlockBySlot(ctx, block, ts); err != nil {
		t.Fatalf("StoreBlockByHash failed: %v", err)
	}

	// Retrieve it
	gotBlock, err := backend.GetBlockBySlot(ctx, ts)
	if err != nil {
		t.Fatalf("GetBlockBySlot failed: %v", err)
	}
	if gotBlock == nil {
		t.Fatalf("expected to retrieve a block, got nil")
	}

	// Check some fields
	if gotBlock.Header.Slot != block.Header.Slot {
		t.Errorf("Slot mismatch: want %d, got %d", block.Header.Slot, gotBlock.Header.Slot)
	}

	if gotBlock.Header.ExtrinsicHash != block.Header.ExtrinsicHash {
		t.Errorf("ExtrinsicHash mismatch: want %s, got %s", block.Header.ExtrinsicHash, gotBlock.Header.ExtrinsicHash)
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

func TestHashSegmentMap_LoadDict_Empty(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	backend := NewRedisBackend(rdb)

	result, err := backend.GetHashSegmentMap()
	require.NoError(t, err, "Should not error when loading from empty Redis key")
	require.Empty(t, result, "Expected empty map when no data is saved")
}

func TestHashSegmentMap_LoadDict(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	backend := NewRedisBackend(rdb)

	// Add 3 test data
	for i := 0; i < 3; i++ {
		var wpHash, segRoot types.OpaqueHash
		copy(wpHash[:], []byte(fmt.Sprintf("wp%02d", i)))
		copy(segRoot[:], []byte(fmt.Sprintf("seg%02d", i)))

		_, err := backend.SetHashSegmentMapWithLimit(wpHash, segRoot)
		require.NoError(t, err)
		time.Sleep(1 * time.Second)
	}

	// Get data from Redis
	result, err := backend.GetHashSegmentMap()
	require.NoError(t, err)
	require.Len(t, result, 3)

	for i := 0; i < 3; i++ {
		var wpHash, segRoot types.OpaqueHash
		copy(wpHash[:], []byte(fmt.Sprintf("wp%02d", i)))
		copy(segRoot[:], []byte(fmt.Sprintf("seg%02d", i)))

		val, ok := result[wpHash]
		require.True(t, ok, "Expected wpHash %v in result", wpHash)
		require.Equal(t, segRoot, val, "Mismatch for wpHash %v", wpHash)
	}
}

func TestHashSegmentMap_SaveWithLimit(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	backend := NewRedisBackend(rdb)

	for i := 0; i < 10; i++ {
		var wpHash, segRoot types.OpaqueHash
		copy(wpHash[:], []byte(fmt.Sprintf("wp%02d", i)))
		copy(segRoot[:], []byte(fmt.Sprintf("seg%02d", i)))

		_, err := backend.SetHashSegmentMapWithLimit(wpHash, segRoot)
		require.NoError(t, err)
		time.Sleep(1 * time.Second) // Simulate time difference
	}

	result, err := backend.GetHashSegmentMap()
	require.NoError(t, err)
	require.Len(t, result, 8)

	// Check that the oldest 2 data (wp0 and wp1) are evicted
	for i := 0; i < 2; i++ {
		var key types.OpaqueHash
		copy(key[:], []byte(fmt.Sprintf("wp%02d", i)))
		_, exists := result[key]
		require.False(t, exists)
	}

	// Check that the latest 8 data (wp2 to wp9) are present and correct
	for i := 2; i < 10; i++ {
		var key, expected types.OpaqueHash
		copy(key[:], []byte(fmt.Sprintf("wp%02d", i)))
		copy(expected[:], []byte(fmt.Sprintf("seg%02d", i)))

		val, exists := result[key]
		require.True(t, exists)
		require.Equal(t, expected, val)
	}
}

func TestHashSegmentMap_LoadDict_ReturnData(t *testing.T) {
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	backend := NewRedisBackend(rdb)

	// Add test data
	var wpHash1, segRoot1 types.OpaqueHash
	copy(wpHash1[:], []byte("wp01"))
	copy(segRoot1[:], []byte("seg01"))
	data, err := backend.SetHashSegmentMapWithLimit(wpHash1, segRoot1)
	require.NoError(t, err)
	require.Len(t, data, 1)
	require.Equal(t, segRoot1, data[wpHash1], "Expected wpHash %v to return segRoot %v", wpHash1, segRoot1)

	// Add another test data
	var wpHash2, segRoot2 types.OpaqueHash
	copy(wpHash2[:], []byte("wp02"))
	copy(segRoot2[:], []byte("seg02"))
	data, err = backend.SetHashSegmentMapWithLimit(wpHash2, segRoot2)
	require.NoError(t, err)
	require.Len(t, data, 2)
	require.Equal(t, segRoot1, data[wpHash1], "Expected wpHash %v to return segRoot %v", wpHash1, segRoot1)
	require.Equal(t, segRoot2, data[wpHash2], "Expected wpHash %v to return segRoot %v", wpHash2, segRoot2)
}

func TestSegmentErasureMap_SaveAndGet(t *testing.T) {
	// Setup
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	backend := NewRedisBackend(rdb)

	// Test data
	segmentRoot := types.OpaqueHash{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
	}
	erasureRoot := types.OpaqueHash{
		0x98, 0x76, 0x54, 0x32, 0x10, 0xfe, 0xdc, 0xba,
	}

	// Save
	err := backend.SetSegmentErasureMap(segmentRoot, erasureRoot)
	require.NoError(t, err)

	// Get
	got, err := backend.GetSegmentErasureMap(segmentRoot)
	require.NoError(t, err)
	require.Equal(t, erasureRoot, got)

	// Make sure getting a non-existent key returns empty OpaqueHash, not an error
	missingKey := types.OpaqueHash{}
	missingVal, err := backend.GetSegmentErasureMap(missingKey)
	require.NoError(t, err)
	require.Equal(t, types.OpaqueHash{}, missingVal)
}

func TestSegmentErasureMap_TTL(t *testing.T) {
	// Setup
	// Start a local, in-memory Redis server for testing
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to init miniredis %v:", err)
	}
	defer s.Close()

	// Initialize our RedisClient using the in-memory server's address
	rdb := NewRedisClient(s.Addr(), "", 0)

	backend := NewRedisBackend(rdb)

	segmentRoot := types.OpaqueHash{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
	}
	erasureRoot := types.OpaqueHash{
		0x98, 0x76, 0x54, 0x32, 0x10, 0xfe, 0xdc, 0xba,
	}

	err = backend.SetSegmentErasureMap(segmentRoot, erasureRoot)
	require.NoError(t, err)

	// Simulate 29 days passing
	s.FastForward(29 * 24 * time.Hour)

	// Check if the key has expired
	got, err := backend.GetSegmentErasureMap(segmentRoot)
	require.NoError(t, err)
	require.Equal(t, types.OpaqueHash{}, got)
}
