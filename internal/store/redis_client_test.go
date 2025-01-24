package store

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/alicebob/miniredis/v2"
	"github.com/test-go/testify/require"
)

func HexToBytes(hexString string) []byte {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		fmt.Printf("failed to decode hex string: %v", err)
	}
	return bytes
}

func setupTestRedis(t *testing.T) (*RedisClient, func()) {
	// Start a local, in-memory Redis server for testing
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to init miniredis %v:", err)
	}

	// Create a cleanup function to close the server after tests
	cleanup := func() {
		s.Close()
	}

	// Initialize our RedisClient using the in-memory server's address
	rdb := NewRedisClient(s.Addr(), "", 0)
	return rdb, cleanup
}

func TestRedisClient_Ping(t *testing.T) {
	// Setup
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	// Exercise + Verify
	err := rdb.Ping()
	if err != nil {
		t.Errorf("ping error:%v", err)
	}
}

func TestRedisClient_PutAndGet(t *testing.T) {
	// Setup
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	// Put
	key := "test-key"
	value := []byte("hello world")
	err := rdb.Put(key, value)
	require.NoError(t, err)

	// Get
	gotVal, err := rdb.Get(key)
	require.NoError(t, err)
	require.Equal(t, value, gotVal)

	// Make sure getting a non-existent key returns nil, not an error
	missingVal, err := rdb.Get("missing-key")
	require.NoError(t, err)
	require.Nil(t, missingVal)
}

func TestRedisClient_Delete(t *testing.T) {
	// Setup
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	// Put a key first
	key := "test-delete-key"
	value := []byte("to-be-deleted")
	err := rdb.Put(key, value)
	require.NoError(t, err)

	// Delete
	err = rdb.Delete(key)
	require.NoError(t, err)

	// Ensure it's gone
	gotVal, err := rdb.Get(key)
	require.NoError(t, err)
	require.Nil(t, gotVal)
}

func TestRedisClient_Batch(t *testing.T) {
	// Setup
	ctx := context.Background()
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	ops := []BatchOperation{
		{
			Type:  OpPut,
			Key:   "batch-key-1",
			Value: []byte("value-1"),
		},
		{
			Type:  OpPut,
			Key:   "batch-key-2",
			Value: []byte("value-2"),
		},
		{
			Type: OpDelete,
			Key:  "batch-key-3",
		},
	}

	// Pre-insert batch-key-3 so that we can test the Delete operation
	require.NoError(t, rdb.Put("batch-key-3", []byte("value-3")))

	// Perform batch operation
	err := rdb.Batch(ctx, ops)
	require.NoError(t, err)

	// Verify results
	val1, err := rdb.Get("batch-key-1")
	require.NoError(t, err)
	require.Equal(t, []byte("value-1"), val1)

	val2, err := rdb.Get("batch-key-2")
	require.NoError(t, err)
	require.Equal(t, []byte("value-2"), val2)

	val3, err := rdb.Get("batch-key-3")
	require.NoError(t, err)
	require.Nil(t, val3, "Expected batch-key-3 to be deleted")
}

func TestRedisClient_StoreBlockAndGetBlock(t *testing.T) {
	// Setup
	ctx := context.Background()
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	var header types.Header
	var empty_mark []types.Ed25519Public

	header.Slot = 999
	header.Parent = types.HeaderHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000"))
	header.ParentStateRoot = types.StateRoot(HexToBytes("0x14aee91ef5e8e22daf2946eab3d688190b84edd7dececbecf5007fcbd0ecd7eb"))
	header.ExtrinsicHash = types.OpaqueHash(HexToBytes("0x189d15af832dfe4f67744008b62c334b569fcbb4c261e0f065655697306ca252"))
	header.OffendersMark = empty_mark
	header.AuthorIndex = 4
	header.EntropySource = types.BandersnatchVrfSignature(HexToBytes("0x9f9f647b5fe173545f735cfca7432b3edfb757f258e4b66980f672d2066b513863b8fcbab8533327586ae3adc6ed6ddbd5a5454f4bc3afc53e61d48a3fba15072f35e3ab005fcf3cb43471036d80f506f0410a65021738d4ca46e9d94afe2610"))

	// Construct a dummy block
	block := types.Block{
		Header: header,
		Extrinsic: types.Extrinsic{
			Tickets: types.TicketsExtrinsic{types.TicketEnvelope{
				Attempt: types.TicketAttempt(1),
			}},
		},
	}

	// Store the block
	err := rdb.StoreBlock(block)
	require.NoError(t, err)

	// Retrieve the block
	gotBlock, err := rdb.GetBlock(ctx, block.Header.Slot)
	require.NoError(t, err)
	require.NotNil(t, gotBlock)
	require.Equal(t, block.Header.Slot, gotBlock.Header.Slot)
	require.Equal(t, block.Header.Parent, gotBlock.Header.Parent)
	require.Equal(t, block.Header.ExtrinsicHash, gotBlock.Header.ExtrinsicHash)
	require.Equal(t, block.Header.OffendersMark, gotBlock.Header.OffendersMark)
	require.Equal(t, block.Header.AuthorIndex, gotBlock.Header.AuthorIndex)
	require.Equal(t, block.Header.EntropySource, gotBlock.Header.EntropySource)
	require.Equal(t, block.Extrinsic.Tickets[0].Attempt, gotBlock.Extrinsic.Tickets[0].Attempt)
	require.Equal(t, block.Extrinsic.Tickets[0].Signature, gotBlock.Extrinsic.Tickets[0].Signature)

	// Ensure the stored data is valid JSON in Redis
	redisKey := "block:999"
	raw, err := rdb.Get(redisKey)
	require.NoError(t, err)
	require.NotNil(t, raw)

	var check types.Block
	require.NoError(t, json.Unmarshal(raw, &check))
	require.Equal(t, block.Header.Slot, check.Header.Slot)
	require.Equal(t, block.Header.Parent, check.Header.Parent)
	require.Equal(t, block.Header.ExtrinsicHash, check.Header.ExtrinsicHash)
	require.Equal(t, block.Header.OffendersMark, check.Header.OffendersMark)
	require.Equal(t, block.Header.AuthorIndex, check.Header.AuthorIndex)
	require.Equal(t, block.Header.EntropySource, check.Header.EntropySource)
}

func TestRedisClient_DeleteBlock(t *testing.T) {
	// Setup
	ctx := context.Background()
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	// Construct and store a block
	block := types.Block{
		Header: types.Header{
			Slot: 1234,
		},
	}
	err := rdb.StoreBlock(block)
	require.NoError(t, err)

	// Delete it
	err = rdb.DeleteBlock(ctx, block.Header.Slot)
	require.NoError(t, err)

	// Make sure it's gone
	gotBlock, err := rdb.GetBlock(ctx, block.Header.Slot)
	require.NoError(t, err)
	require.Nil(t, gotBlock)
}
