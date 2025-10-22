package store

import (
	"fmt"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/alicebob/miniredis/v2"
	"github.com/test-go/testify/require"
)

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
