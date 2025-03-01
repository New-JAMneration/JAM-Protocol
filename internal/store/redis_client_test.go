package store

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

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
