package store

import (
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/alicebob/miniredis/v2"
	"github.com/test-go/testify/require"
)

func TestSegmentErasureMap_SaveAndGet(t *testing.T) {
	// Setup
	rdb, cleanup := setupTestRedis(t)
	defer cleanup()

	store := NewSegmentErasureMap(rdb)

	// Test data
	segmentRoot := types.OpaqueHash{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
	}
	erasureRoot := types.OpaqueHash{
		0x98, 0x76, 0x54, 0x32, 0x10, 0xfe, 0xdc, 0xba,
	}

	// Save
	err := store.Save(segmentRoot, erasureRoot)
	require.NoError(t, err)

	// Get
	got, err := store.Get(segmentRoot)
	require.NoError(t, err)
	require.Equal(t, erasureRoot, got)

	// Make sure getting a non-existent key returns empty OpaqueHash, not an error
	missingKey := types.OpaqueHash{}
	missingVal, err := store.Get(missingKey)
	require.NoError(t, err)
	require.Equal(t, types.OpaqueHash{}, missingVal)
}

func TestSegmentErasureMap_TTL(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := NewRedisClient(s.Addr(), "", 0)
	store := NewSegmentErasureMap(rdb)

	segmentRoot := types.OpaqueHash{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
	}
	erasureRoot := types.OpaqueHash{
		0x98, 0x76, 0x54, 0x32, 0x10, 0xfe, 0xdc, 0xba,
	}

	err = store.Save(segmentRoot, erasureRoot)
	require.NoError(t, err)

	// Simulate 29 days passing
	s.FastForward(29 * 24 * time.Hour)

	// Check if the key has expired
	got, err := store.Get(segmentRoot)
	require.NoError(t, err)
	require.Equal(t, types.OpaqueHash{}, got)
}
